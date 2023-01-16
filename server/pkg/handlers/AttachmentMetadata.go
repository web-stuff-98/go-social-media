package handlers

import (
	"encoding/json"
	"io/ioutil"
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
