package handlers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/web-stuff-98/go-social-media/pkg/attachmentserver"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
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

	numChunks := math.Ceil(float64(metadataInput.Size) / float64(1024*1024*2))
	chunkIDs := []primitive.ObjectID{msgId}
	for i := 0; i < int(numChunks)+1; i++ { //add an extra object ID at the end that will not be used to avoid out of range error
		chunkIDs = append(chunkIDs, primitive.NewObjectID())
	}

	totalChunks := int(math.Ceil(float64(metadataInput.Size) / 1048576))
	if _, ok := h.AttachmentServer.Uploaders[user.ID]; !ok {
		h.AttachmentServer.Uploaders[user.ID] = make(map[primitive.ObjectID]attachmentserver.Upload)
	}
	h.AttachmentServer.Uploaders[user.ID][msgId] = attachmentserver.Upload{
		ChunksDone:  0,
		TotalChunks: totalChunks,
		ChunkIDs:    chunkIDs,
	}

	if _, err := h.Collections.AttachmentMetadataCollection.InsertOne(r.Context(), models.AttachmentMetadata{
		ID:          msgId,
		MimeType:    metadataInput.MimeType,
		Name:        metadataInput.Name,
		Size:        metadataInput.Size,
		VideoLength: metadataInput.Length,
		Pending:     true,
		Failed:      false,
		ChunkIDs:    chunkIDs[:len(chunkIDs)-1], //remove the last ID because its just there to stop the out of range error
	}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	responseMessage(w, http.StatusCreated, "Created attachment metadata")
}

func (h handler) UploadAttachmentChunk(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawMsgId := mux.Vars(r)["msgId"]
	msgId, err := primitive.ObjectIDFromHex(rawMsgId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)

	if r.ContentLength == -1 {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	if r.ContentLength > 1048576 {
		responseMessage(w, http.StatusRequestEntityTooLarge, "Bad request")
		return
	}

	uploads, ok := h.AttachmentServer.Uploaders[user.ID]
	if !ok {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	upload, ok := uploads[msgId]
	if !ok {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	chunkID := upload.ChunkIDs[upload.ChunksDone]

	h.Collections.AttachmentChunksCollection.InsertOne(r.Context(), models.AttachmentChunk{
		ID:        chunkID,
		Bytes:     primitive.Binary{Data: body},
		NextChunk: chunkID,
	})
}

// Download attachment as a file using octet stream
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
	w.Header().Add("Content-Length", strconv.Itoa(metaData.Size))
	w.Header().Add("Content-Disposition", `attachment; filename="`+metaData.Name+`"`)

	w.Write(firstChunk.Bytes.Data)

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
		if chunk.NextChunk != primitive.NilObjectID {
			return recursivelyWriteAttachmentChunksToResponse(w, chunk.NextChunk, chunkColl, ctx)
		} else {
			return nil
		}
	}
}

// Get partial content from attachment for video player (this doesn't work on attachments that are larger than 2 chunks)
func (h handler) GetVideoPartialContent(w http.ResponseWriter, r *http.Request) {
	rangeString := r.Header.Get("Range")
	log.Println("range header: ", rangeString)
	if rangeString == "" {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	bytesPortion := strings.ReplaceAll(rangeString, "bytes=", "") // Returns the start-end portion of the header
	if bytesPortion == "" {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	bytesHeaderPortionArr := strings.Split(bytesPortion, "-")
	if len(bytesHeaderPortionArr) == 0 {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	rangeStart, err := strconv.Atoi(bytesHeaderPortionArr[0])
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	rawId := mux.Vars(r)["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var metaData models.AttachmentMetadata
	if err := h.Collections.AttachmentMetadataCollection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&metaData); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}
	if metaData.MimeType != "video/mp4" {
		responseMessage(w, http.StatusBadRequest, "File is not a video")
		return
	}
	maxLength := int(math.Min(float64(metaData.Size-rangeStart), 1048576))
	rangeEnd := rangeStart + maxLength

	log.Println("Getting bytes", rangeStart, "to", rangeEnd)

	// Determine which chunks are needed
	startChunk := float64(rangeStart) / 1048576
	endChunk := math.Ceil(float64(rangeEnd) / 1048576)

	log.Println("From", int32(startChunk), "to", int32(endChunk))

	chunkIDs := metaData.ChunkIDs[int32(startChunk):int32(endChunk)]
	log.Println("Retrieving chunk ids", chunkIDs)
	// Get the bytes from the chunks
	vidBytes := []byte{}
	cursor, err := h.Collections.AttachmentChunksCollection.Find(r.Context(), bson.M{"_id": bson.M{"$in": chunkIDs}})
	defer cursor.Close(r.Context())
	for cursor.Next(r.Context()) {
		var chunk models.AttachmentChunk
		if err := cursor.Decode(&chunk); err != nil {
			log.Println("DECODE ERROR : ", err)
		}
		vidBytes = append(vidBytes, chunk.Bytes.Data...)
		log.Println("Retrieved :", len(vidBytes))
	}

	bytesStartString := strconv.Itoa(rangeStart)
	bytesEndString := strconv.Itoa(rangeEnd)
	bytesSizeString := strconv.Itoa(metaData.Size)

	w.Header().Add("Accept-Ranges", "bytes")
	w.Header().Add("Content-Length", strconv.Itoa(maxLength))
	w.Header().Add("Content-Range", bytesStartString+"-"+bytesEndString+"/"+bytesSizeString)
	w.Header().Add("Content-Type", "video/mp4")

	w.Write(vidBytes)
}
