package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/web-stuff-98/go-social-media/pkg/attachmentserver"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/helpers"
	"github.com/web-stuff-98/go-social-media/pkg/socketmodels"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"
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

	numChunks := math.Ceil(float64(metadataInput.Size) / float64(1048576)) // +1
	chunkIDs := []primitive.ObjectID{msgId}
	for i := 1; i < int(numChunks); i++ {
		chunkIDs = append(chunkIDs, primitive.NewObjectID())
	}
	chunkIDs = append(chunkIDs, primitive.NilObjectID)

	totalChunks := int(math.Ceil(float64(metadataInput.Size) / 1048576))
	if _, ok := h.AttachmentServer.Uploaders[user.ID]; !ok {
		h.AttachmentServer.Uploaders[user.ID] = make(map[primitive.ObjectID]attachmentserver.Upload)
	}
	h.AttachmentServer.Uploaders[user.ID][msgId] = attachmentserver.Upload{
		ChunksDone:        0,
		TotalChunks:       totalChunks,
		ChunkIDs:          chunkIDs,
		SubscriptionNames: metadataInput.SubscriptionNames,
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
	log.Println("Chunk incoming")

	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawMsgId := mux.Vars(r)["msgId"]
	msgId, err := primitive.ObjectIDFromHex(rawMsgId)
	if err != nil {
		log.Println(err)
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)

	if r.ContentLength == -1 || r.ContentLength == 0 {
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

	if _, err := h.Collections.AttachmentChunksCollection.InsertOne(r.Context(), models.AttachmentChunk{
		ID:        chunkID,
		Bytes:     primitive.Binary{Data: body},
		NextChunk: upload.ChunkIDs[upload.ChunksDone+1],
	}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	log.Println(upload.ChunksDone)
	log.Println(upload.TotalChunks)

	if upload.ChunksDone == upload.TotalChunks-1 {
		outBytes, _ := json.Marshal(socketmodels.OutMessage{
			Type: "ATTACHMENT_PROGRESS",
			Data: `{"ID":"` + msgId.Hex() + `","failed":false,"pending":false,"ratio":1}`,
		})
		h.Collections.AttachmentMetadataCollection.UpdateByID(r.Context(), msgId, bson.M{"$set": bson.M{"pending": false}})
		h.SocketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
			Names: upload.SubscriptionNames,
			Data:  outBytes,
		}
		log.Println("DONE LAST CHUNK")
	} else {
		outBytes, _ := json.Marshal(socketmodels.OutMessage{
			Type: "ATTACHMENT_PROGRESS",
			Data: `{"ID":"` + msgId.Hex() + `","failed":false,"pending":true,"ratio":` + getProgressString(upload) + `}`,
		})
		h.SocketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
			Names: upload.SubscriptionNames,
			Data:  outBytes,
		}
	}

	upload.ChunksDone++
	h.AttachmentServer.Uploaders[user.ID][msgId] = upload

	responseMessage(w, http.StatusCreated, "Chunk created")
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

// BROKEN - CLIENT CANT PLAY BACK VIDEO FOR SOME REASON, I GIVE UP AFTER WASTING 5 DAYS
/*func (h handler) GetVideoPartialContent(w http.ResponseWriter, r *http.Request) {
	rawId := mux.Vars(r)["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {s
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

	// Process the range header
	var maxLength int64
	var start, end int64
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		_, err := fmt.Sscanf(rangeHeader, "bytes=%d-", &start)
		if err != nil {
			responseMessage(w, http.StatusBadRequest, "Invalid range header")
			return
		}
		maxLength = 1 * 1024 * 1024
		if start+maxLength > int64(metaData.Size) {
			maxLength = int64(metaData.Size) - start
		}
		// check if end is present in the range header
		if i := strings.Index(rangeHeader, "-"); i != -1 {
			end, err = strconv.ParseInt(rangeHeader[i+1:], 10, 64)
			if err != nil {
				// if end is absent, set it
				end = start + maxLength
			}
		} else {
			// if end is absent, set it
			end = start + maxLength
		}
	}

	// Calculate the start and end chunk indexes
	startChunkIndex := int(start / (1 * 1024 * 1024))
	endChunkIndex := startChunkIndex + 1

	log.Println("START CHUNK INDEX:", startChunkIndex)
	log.Println("END CHUNK INDEX:", endChunkIndex)

	// Retrieve the chunks
	chunkBytes := []byte{}
	cursor, err := h.Collections.AttachmentChunksCollection.Find(r.Context(), bson.M{"_id": bson.M{"$in": metaData.ChunkIDs[startChunkIndex:endChunkIndex]}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}
	i := 0
	for cursor.Next(r.Context()) {
		var chunk models.AttachmentChunk
		if err := cursor.Decode(&chunk); err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
		if i == 0 {
			chunkBytes = append(chunkBytes, chunk.Bytes.Data[start-(int64(startChunkIndex)*1*1024*1024):]...)
		} else {
			chunkBytes = append(chunkBytes, chunk.Bytes.Data...)
		}
		i++
	}

	log.Println(len(chunkBytes))

	w.Header().Add("Accept-Ranges", "bytes")
	w.Header().Add("Content-Length", fmt.Sprint(maxLength))
	w.Header().Add("Content-Range", fmt.Sprint(start)+"-"+fmt.Sprint(end)+"/"+fmt.Sprint(metaData.Size))
	w.Header().Add("Content-Type", "video/mp4")

	w.Write(chunkBytes[:maxLength])
}*/

func getProgressString(upload attachmentserver.Upload) string {
	return fmt.Sprintf("%v", float32(upload.ChunksDone+1)/float32(upload.TotalChunks))
}
