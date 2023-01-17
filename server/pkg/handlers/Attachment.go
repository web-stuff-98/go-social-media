package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/filesocketserver"
	"github.com/web-stuff-98/go-social-media/pkg/helpers"
	"github.com/web-stuff-98/go-social-media/pkg/validation"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (h handler) HandleAttachmentMetadata(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)

	rawMsgId := mux.Vars(r)["msgId"]
	msgId, err := primitive.ObjectIDFromHex(rawMsgId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	rawRecipientId := mux.Vars(r)["recipientId"]
	// Recipient ID can be either a user for private messages, or a room
	recipientId, err := primitive.ObjectIDFromHex(rawRecipientId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var metadataInput validation.AttachmentMetadata
	if json.Unmarshal(body, &metadataInput); err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	validate := validator.New()
	if err := validate.Struct(metadataInput); err != nil {
		responseMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	//First validate the message exists by finding it in the recipient inbox / chatroom
	isPrivateMsg := false
	found := false
	var inbox models.Inbox
	if err := h.Collections.InboxCollection.FindOne(r.Context(), bson.M{"_id": recipientId}).Decode(&inbox); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	} else {
		isPrivateMsg = true
	}
	if isPrivateMsg {
		for _, pm := range inbox.Messages {
			if pm.ID == msgId && pm.Uid == user.ID {
				found = true
				break
			}
		}
	} else {
		var roomMsgs models.RoomMessages
		if err := h.Collections.RoomMessagesCollection.FindOne(r.Context(), bson.M{"_id": recipientId}).Decode(&roomMsgs); err != nil {
			if err != mongo.ErrNoDocuments {
				responseMessage(w, http.StatusInternalServerError, "Internal error")
			} else {
				responseMessage(w, http.StatusNotFound, "Not found")
			}
			return
		} else {
			for _, rm := range roomMsgs.Messages {
				if rm.ID == msgId && rm.Uid == user.ID {
					found = true
					break
				}
			}
		}
	}
	if !found {
		responseMessage(w, http.StatusNotFound, "Not found")
		return
	}

	if _, err := h.Collections.AttachmentMetadataCollection.InsertOne(r.Context(), models.AttachmentMetadata{
		ID:       msgId,
		MimeType: metadataInput.MimeType,
		Name:     metadataInput.Name,
		Size:     metadataInput.Size,
		Pending:  true,
		Failed:   false,
	}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	if isPrivateMsg {
		h.FileSocketServer.AttachmentSubscriptionNames[msgId] = []string{"inbox=" + recipientId.Hex(), "inbox=" + user.ID.Hex()}
	} else {
		h.FileSocketServer.AttachmentSubscriptionNames[msgId] = []string{"room=" + recipientId.Hex()}
	}

	h.FileSocketServer.AttachmentBytesProcessed[msgId] = filesocketserver.BytesInfo{
		TotalBytes: metadataInput.Size,
		BytesDone:  0,
	}

	responseMessage(w, http.StatusCreated, "Created attachment metadata")
}

// Download attachment using octet stream
func (h handler) DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	rawAttachmentId := mux.Vars(r)["id"]
	attachmentId, err := primitive.ObjectIDFromHex(rawAttachmentId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var metaData models.AttachmentMetadata
	if h.Collections.AttachmentMetadataCollection.FindOne(r.Context(), bson.M{"_id": attachmentId}).Decode(&metaData); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	var firstChunk models.AttachmentChunk
	if err := h.Collections.AttachmentChunksCollection.FindOne(r.Context(), bson.M{"_id": attachmentId}).Decode(&firstChunk); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", `attachment; filename="`+metaData.Name+`"`)

	w.Write(firstChunk.Bytes.Data)
	log.Println("Wrote first chunk")

	if firstChunk.NextChunk != primitive.NilObjectID {
		recursivelyWriteAttachmentChunksToResponse(w, firstChunk.NextChunk, h.Collections.AttachmentChunksCollection, r.Context())
	}
}

func recursivelyWriteAttachmentChunksToResponse(w http.ResponseWriter, NextChunkID primitive.ObjectID, chunkColl *mongo.Collection, ctx context.Context) error {
	var chunk models.AttachmentChunk
	if err := chunkColl.FindOne(ctx, bson.M{"_id": NextChunkID}).Decode(&chunk); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		} else {
			return err
		}
	} else {
		w.Write(chunk.Bytes.Data)
		log.Println("Wrote chunk")
		if chunk.NextChunk != primitive.NilObjectID {
			return recursivelyWriteAttachmentChunksToResponse(w, chunk.NextChunk, chunkColl, ctx)
		} else {
			return nil
		}
	}
}
