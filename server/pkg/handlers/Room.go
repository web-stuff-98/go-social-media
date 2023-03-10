package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/helpers"
	"github.com/web-stuff-98/go-social-media/pkg/socketmodels"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"
	"github.com/web-stuff-98/go-social-media/pkg/validation"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h handler) GetRoomPage(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pageNumberString := mux.Vars(r)["page"]
	pageNumber, err := strconv.Atoi(pageNumberString)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid page")
		return
	}
	pageSize := 20

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

	filter := bson.M{}
	if r.URL.Query().Has("term") {
		if r.URL.Query().Get("term") != " " {
			filter = bson.M{
				"$text": bson.M{
					"$search":        r.URL.Query().Get("term"),
					"$caseSensitive": false,
				},
			}
		}
	}
	if r.URL.Query().Has("OWN_ROOMS") {
		filter = bson.M{
			"author_id": user.ID,
		}
		if r.URL.Query().Has("term") {
			if r.URL.Query().Get("term") != " " {
				filter = bson.M{
					"$text": bson.M{
						"$search":        r.URL.Query().Get("term"),
						"$caseSensitive": false,
					},
					"author_id": user.ID,
				}
			}
		}
	}
	if r.URL.Query().Has("INVITED_ROOMS") {
		matchingIds := []primitive.ObjectID{}
		if cursor, err := h.Collections.RoomPrivateDataCollection.Find(r.Context(), bson.M{"members": bson.M{"$all": []primitive.ObjectID{user.ID}}}); err != nil {
			cursor.Close(r.Context())
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		} else {
			for cursor.Next(r.Context()) {
				roomPrivateData := &models.RoomPrivateData{}
				if err := cursor.Decode(&roomPrivateData); err != nil {
					cursor.Close(r.Context())
					responseMessage(w, http.StatusInternalServerError, "Internal error")
					return
				} else {
					matchingIds = append(matchingIds, roomPrivateData.ID)
				}
			}
			cursor.Close(r.Context())
		}
		filter = bson.M{
			"_id": bson.M{"$in": matchingIds},
		}
		if r.URL.Query().Has("term") {
			if r.URL.Query().Get("term") != " " {
				filter = bson.M{
					"$text": bson.M{
						"$search":        r.URL.Query().Get("term"),
						"$caseSensitive": false,
					},
					"_id": bson.M{"$in": matchingIds},
				}
			}
		}
	}

	// Because countdocuments is expensive, O(n) it says in docs, look for the value stored in cache first
	// It's fine if the count is slightly out of date. Probably a better way to do this
	var count int64
	filterKey := "FIND-ROOMS-FILTER-COUNT=" + fmt.Sprint(filter)
	getFilterCmd := h.RedisClient.Get(r.Context(), filterKey)
	if getFilterCmd.Err() == nil {
		if cachedCount, err := strconv.Atoi(getFilterCmd.Val()); err == nil {
			count = int64(cachedCount)
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	} else {
		exactCount, err := h.Collections.RoomCollection.CountDocuments(r.Context(), filter)
		if err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
		count = exactCount
		setCmd := h.RedisClient.Set(r.Context(), filterKey, fmt.Sprint(exactCount), time.Second*15)
		if setCmd.Err() != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	}

	findOptions.SetLimit(int64(pageSize))
	findOptions.SetSkip(int64(pageSize) * (int64(pageNumber) - 1))

	cursor, err := h.Collections.RoomCollection.Find(r.Context(), filter, findOptions)
	defer cursor.Close(r.Context())
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	var rooms []models.Room
	if err = cursor.All(r.Context(), &rooms); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	// Need to fill out the can_access property for frontend display, so get the RoomPrivateData documents for each room
	for i, room := range rooms {
		privateData := &models.RoomPrivateData{}
		if err := h.Collections.RoomPrivateDataCollection.FindOne(r.Context(), bson.M{"_id": room.ID}).Decode(&privateData); err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
		rooms[i].CanAccess = true
		if room.Author != user.ID {
			if room.Private {
				isMember := false
				for _, oi := range privateData.Members {
					if oi == user.ID {
						isMember = true
						break
					}
				}
				rooms[i].CanAccess = isMember
			}
			for _, oi := range privateData.Banned {
				isBanned := false
				if oi == user.ID {
					isBanned = true
					break
				}
				if isBanned {
					rooms[i].CanAccess = false
				}
			}
		}
	}

	roomBytes, err := json.Marshal(rooms)

	out := map[string]string{
		"count": fmt.Sprint(count),
		"rooms": string(roomBytes),
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(out)
}

func (h handler) GetOwnRooms(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rooms := []models.Room{}
	cursor, err := h.Collections.RoomCollection.Find(r.Context(), bson.M{"author_id": user.ID})
	defer cursor.Close(r.Context())
	for cursor.Next(r.Context()) {
		room := &models.Room{}
		cursor.Decode(&room)
		rooms = append(rooms, *room)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rooms)
}

func (h handler) InviteToRoom(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawId := mux.Vars(r)["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	room := &models.Room{}
	if err := h.Collections.RoomCollection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&room); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	if room.Author != user.ID {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var rawUid string
	if r.URL.Query().Has("uid") {
		rawUid = r.URL.Query().Get("uid")
	} else {
		responseMessage(w, http.StatusBadRequest, "No UID provided")
		return
	}

	recipientId, err := primitive.ObjectIDFromHex(rawUid)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid UID")
		return
	}

	if recipientId == user.ID {
		responseMessage(w, http.StatusBadRequest, "You cannot invite yourself")
		return
	}

	msg := &models.PrivateMessage{
		ID:                   primitive.NewObjectID(),
		Content:              id.Hex(),
		IsInvitation:         true,
		IsAcceptedInvitation: false,
		IsDeclinedInvitation: false,
		Uid:                  user.ID,
		CreatedAt:            primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:            primitive.NewDateTimeFromTime(time.Now()),
		RecipientId:          recipientId,
		HasAttachment:        false,
	}

	if _, err := h.Collections.InboxCollection.UpdateByID(r.Context(), recipientId, bson.M{"$push": bson.M{"messages": msg}}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	if _, err := h.Collections.InboxCollection.UpdateByID(r.Context(), user.ID, bson.M{"$addToSet": bson.M{"messages_sent_to": recipientId}}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	hasConvsOpenWithRecv := make(chan bool)
	h.SocketServer.GetUserConversationsOpenWith <- socketserver.GetUserConversationsOpenWith{
		RecvChan: hasConvsOpenWithRecv,
		Uid:      recipientId,
		UidB:     user.ID,
	}
	hasConvsOpenWith := <-hasConvsOpenWithRecv
	addNotification := !hasConvsOpenWith

	if addNotification {
		if _, err := h.Collections.NotificationsCollection.UpdateByID(context.TODO(), recipientId, bson.M{
			"$push": bson.M{
				"notifications": bson.M{"type": "MSG:" + user.ID.Hex()},
			},
		}); err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	}

	data, err := json.Marshal(msg)
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	} else {
		outBytes, err := json.Marshal(socketmodels.OutMessage{
			Type: "PRIVATE_MESSAGE",
			Data: string(data),
		})
		if err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		} else {
			h.SocketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
				Names: []string{"inbox=" + recipientId.Hex(), "inbox=" + user.ID.Hex()},
				Data:  outBytes,
			}
		}
	}

	responseMessage(w, http.StatusCreated, "Invitation sent")
}

func (h handler) AcceptRoomInvite(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawId := mux.Vars(r)["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	rawMsgId := mux.Vars(r)["msgId"]
	msgId, err := primitive.ObjectIDFromHex(rawMsgId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	room := &models.Room{}
	if err := h.Collections.RoomCollection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&room); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	var rawUid string
	if r.URL.Query().Has("uid") {
		rawUid = r.URL.Query().Get("uid")
	} else {
		responseMessage(w, http.StatusBadRequest, "No UID provided")
		return
	}

	uid, err := primitive.ObjectIDFromHex(rawUid)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid UID")
		return
	}

	if room.Author != uid {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	inbox := &models.Inbox{}
	h.Collections.InboxCollection.FindOne(context.TODO(), bson.M{"_id": user.ID}).Decode(&inbox)
	found := false
	for _, pm := range inbox.Messages {
		if pm.ID == msgId {
			found = true
			if pm.IsDeclinedInvitation {
				responseMessage(w, http.StatusBadRequest, "You have already declined this invitation")
				return
			}
			break
		}
	}

	if !found {
		responseMessage(w, http.StatusNotFound, "Invitation not found")
	}

	if _, err := h.Collections.InboxCollection.UpdateOne(context.TODO(), bson.M{
		"_id":          user.ID,
		"messages._id": msgId,
	}, bson.M{
		"$set": bson.M{"messages.$.invitation_accepted": true},
	}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	if res := h.Collections.RoomPrivateDataCollection.FindOneAndUpdate(context.TODO(), bson.M{
		"_id": room.ID,
	}, bson.M{
		"$addToSet": bson.M{"members": user.ID},
		"$pull":     bson.M{"banned": user.ID},
	}); res.Err() != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	outData := make(map[string]interface{})
	outData["ID"] = msgId.Hex()
	outData["accepted"] = true
	outData["recipient_id"] = user.ID.Hex()
	dataBytes, err := json.Marshal(outData)
	outBytes, err := json.Marshal(socketmodels.OutMessage{
		Type: "PRIVATE_MESSAGE_INVITE_RESPONDED",
		Data: string(dataBytes),
	})
	h.SocketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
		Names: []string{"inbox=" + uid.Hex(), "inbox=" + user.ID.Hex()},
		Data:  outBytes,
	}

	if outPrivateDataChangeBytes, err := json.Marshal(socketmodels.OutChangeMessage{
		Type:   "CHANGE",
		Method: "INSERT",
		Entity: "MEMBER",
		Data:   `{"ID":"` + user.ID.Hex() + `"}`,
	}); err != nil {
		h.SocketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "room_private_data=" + room.ID.Hex(),
			Data: outPrivateDataChangeBytes,
		}
	}

	h.SocketServer.SendDataToUser <- socketserver.UserDataMessage{
		Uid: user.ID,
		Data: socketmodels.OutChangeMessage{
			Method: "UPDATE",
			Entity: "ROOM",
			Data:   `{"ID":"` + room.ID.Hex() + `","can_access":true}`,
		},
		Type: "CHANGE",
	}

	responseMessage(w, http.StatusOK, "Invitation accepted")
}

func (h handler) DeclineRoomInvite(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawId := mux.Vars(r)["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	rawMsgId := mux.Vars(r)["msgId"]
	msgId, err := primitive.ObjectIDFromHex(rawMsgId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	room := &models.Room{}
	if err := h.Collections.RoomCollection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&room); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	var rawUid string
	if r.URL.Query().Has("uid") {
		rawUid = r.URL.Query().Get("uid")
	} else {
		responseMessage(w, http.StatusBadRequest, "No UID provided")
		return
	}

	uid, err := primitive.ObjectIDFromHex(rawUid)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid UID")
		return
	}

	if room.Author != uid {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	inbox := &models.Inbox{}
	h.Collections.InboxCollection.FindOne(context.TODO(), bson.M{"_id": user.ID}).Decode(&inbox)
	found := false
	for _, pm := range inbox.Messages {
		if pm.ID == msgId {
			found = true
			if pm.IsAcceptedInvitation {
				responseMessage(w, http.StatusBadRequest, "You have already accepted this invitation")
				return
			}
			break
		}
	}

	if !found {
		responseMessage(w, http.StatusNotFound, "Invitation not found")
	}

	if _, err := h.Collections.InboxCollection.UpdateOne(context.TODO(), bson.M{
		"_id":          user.ID,
		"messages._id": msgId,
	}, bson.M{
		"$set": bson.M{"messages.$.invitation_declined": true},
	}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	outData := make(map[string]interface{})
	outData["ID"] = msgId.Hex()
	outData["accepted"] = false
	outData["recipient_id"] = user.ID.Hex()
	dataBytes, err := json.Marshal(outData)
	outBytes, err := json.Marshal(socketmodels.OutMessage{
		Type: "PRIVATE_MESSAGE_INVITE_RESPONDED",
		Data: string(dataBytes),
	})
	h.SocketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
		Names: []string{"inbox=" + uid.Hex(), "inbox=" + user.ID.Hex()},
		Data:  outBytes,
	}

	responseMessage(w, http.StatusOK, "Invitation declined")
}

func (h handler) BanUserFromRoom(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawId := mux.Vars(r)["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	room := &models.Room{}
	if err := h.Collections.RoomCollection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&room); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	var rawUid string
	if r.URL.Query().Has("uid") {
		rawUid = r.URL.Query().Get("uid")
	} else {
		responseMessage(w, http.StatusBadRequest, "No UID provided")
		return
	}

	uid, err := primitive.ObjectIDFromHex(rawUid)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid UID")
		return
	}

	if uid == user.ID {
		responseMessage(w, http.StatusBadRequest, "You cannot ban yourself")
		return
	}

	if room.Author != user.ID {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if _, err := h.Collections.RoomPrivateDataCollection.UpdateByID(r.Context(), id, bson.M{"$addToSet": bson.M{"banned": uid}, "$pull": bson.M{"members": uid}}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	h.SocketServer.RemoveUserFromSubscription <- socketserver.RemoveUserFromSubscription{
		Name: "room=" + room.ID.Hex(),
		Uid:  uid,
	}

	outChangeBytes, err := json.Marshal(socketmodels.OutChangeMessage{
		Type:   "CHANGE",
		Method: "INSERT",
		Entity: "BANNED",
		Data:   `{"ID":"` + uid.Hex() + `"}`,
	})
	h.SocketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
		Name: "room_private_data=" + room.ID.Hex(),
		Data: outChangeBytes,
	}

	h.SocketServer.SendDataToUser <- socketserver.UserDataMessage{
		Uid: uid,
		Data: socketmodels.OutChangeMessage{
			Method: "UPDATE",
			Entity: "ROOM",
			Data:   `{"ID":"` + room.ID.Hex() + `","can_access":false}`,
		},
		Type: "CHANGE",
	}

	responseMessage(w, http.StatusOK, "User banned")
}

func (h handler) UnBanUserFromRoom(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawId := mux.Vars(r)["id"]
	id, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	room := &models.Room{}
	if err := h.Collections.RoomCollection.FindOne(r.Context(), bson.M{"_id": id}).Decode(&room); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	var rawUid string
	if r.URL.Query().Has("uid") {
		rawUid = r.URL.Query().Get("uid")
	} else {
		responseMessage(w, http.StatusBadRequest, "No UID provided")
		return
	}

	uid, err := primitive.ObjectIDFromHex(rawUid)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid UID")
		return
	}

	if room.Author != user.ID {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if _, err := h.Collections.RoomPrivateDataCollection.UpdateByID(r.Context(), id, bson.M{"$pull": bson.M{"banned": uid}}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	outChangeBytes, err := json.Marshal(socketmodels.OutChangeMessage{
		Type:   "CHANGE",
		Method: "DELETE",
		Entity: "BANNED",
		Data:   `{"ID":"` + uid.Hex() + `"}`,
	})
	h.SocketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
		Name: "room_private_data=" + room.ID.Hex(),
		Data: outChangeBytes,
	}

	if !room.Private {
		h.SocketServer.SendDataToUser <- socketserver.UserDataMessage{
			Uid: uid,
			Data: socketmodels.OutChangeMessage{
				Method: "UPDATE",
				Entity: "ROOM",
				Data:   `{"ID":"` + room.ID.Hex() + `","can_access":true}`,
			},
			Type: "CHANGE",
		}
	}

	responseMessage(w, http.StatusOK, "User unbanned")
}

func (h handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	var roomInput validation.Room
	if json.Unmarshal(body, &roomInput); err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	validate := validator.New()
	if err := validate.Struct(roomInput); err != nil {
		responseMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	numRooms := 0
	cur, err := h.Collections.RoomCollection.Find(r.Context(), bson.M{
		"author_id": user.ID,
	})
	defer cur.Close(r.Context())
	if err != nil {
		for cur.Next(r.Context()) {
			numRooms++
			room := &models.Room{}
			cur.Decode(&room)
			if strings.ToLower(room.Name) == strings.ToLower(roomInput.Name) {
				responseMessage(w, http.StatusBadRequest, "You already have a room by that name")
				cur.Close(r.Context())
				return
			}
		}
		if numRooms == 4 {
			responseMessage(w, http.StatusBadRequest, "You can create a maximum of 8 rooms")
			return
		}
	} else {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	var room = &models.Room{
		ID:           primitive.NewObjectIDFromTimestamp(time.Now()),
		Name:         roomInput.Name,
		Author:       user.ID,
		CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		ImgBlur:      "",
		ImagePending: true,
		Private:      roomInput.Private,
	}

	inserted, err := h.Collections.RoomCollection.InsertOne(r.Context(), room)
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	var roomMessages = &models.RoomMessages{
		ID:       inserted.InsertedID.(primitive.ObjectID),
		Messages: []models.RoomMessage{},
	}

	if _, err := h.Collections.RoomMessagesCollection.InsertOne(r.Context(), roomMessages); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	var roomPrivateData = &models.RoomPrivateData{
		ID:      inserted.InsertedID.(primitive.ObjectID),
		Members: []primitive.ObjectID{},
		Banned:  []primitive.ObjectID{},
	}

	if _, err := h.Collections.RoomPrivateDataCollection.InsertOne(r.Context(), roomPrivateData); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inserted.InsertedID.(primitive.ObjectID).Hex())
}

func (h handler) GetRoomPrivateData(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawId := mux.Vars(r)["id"]
	roomId, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	roomPrivateData := &models.RoomPrivateData{}
	if err := h.Collections.RoomPrivateDataCollection.FindOne(r.Context(), bson.M{"_id": roomId}).Decode(&roomPrivateData); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	room := &models.Room{}
	if err := h.Collections.RoomCollection.FindOne(r.Context(), bson.M{"_id": roomId}).Decode(&room); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	userIsMember := false
	for _, oi := range roomPrivateData.Members {
		if oi == user.ID {
			userIsMember = true
			break
		}
	}

	if room.Author != user.ID && !userIsMember {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(roomPrivateData)
}

func (h handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawId := mux.Vars(r)["id"]
	roomId, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var room models.Room
	if err := h.Collections.RoomCollection.FindOne(r.Context(), bson.M{"_id": roomId}).Decode(&room); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Room not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	var roomPrivateData models.RoomPrivateData
	if err := h.Collections.RoomPrivateDataCollection.FindOne(r.Context(), bson.M{"_id": roomId}).Decode(&roomPrivateData); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	for _, oi := range roomPrivateData.Banned {
		if oi == user.ID {
			responseMessage(w, http.StatusUnauthorized, "You are banned from this room")
			break
		}
	}
	if room.Private == true {
		isMember := false
		for _, oi := range roomPrivateData.Members {
			if oi == user.ID {
				isMember = true
				break
			}
		}
		if user.ID != room.Author && !isMember {
			responseMessage(w, http.StatusUnauthorized, "This room is private")
			return
		}
	}

	var roomMessages models.RoomMessages
	if err := h.Collections.RoomMessagesCollection.FindOne(r.Context(), bson.M{"_id": roomId}).Decode(&roomMessages); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Room messages not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	room.Messages = roomMessages.Messages

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(room)
}

func (h handler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawId := mux.Vars(r)["id"]
	roomId, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var room models.Room
	if h.Collections.RoomCollection.FindOne(r.Context(), bson.M{"_id": roomId}).Decode(&room); err != nil {
		responseMessage(w, http.StatusNotFound, "Room not found")
		return
	}

	if _, isProtected := h.ProtectedIDs.Rids[roomId]; isProtected {
		responseMessage(w, http.StatusUnauthorized, "You cannot modify example rooms")
		return
	}

	if room.Author != user.ID {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}

	var roomInput validation.Room
	if json.Unmarshal(body, &roomInput); err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	validate := validator.New()
	if err := validate.Struct(roomInput); err != nil {
		responseMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := h.Collections.RoomCollection.Find(r.Context(), bson.M{
		"name": bson.M{
			"$regex":   roomInput.Name,
			"$options": "i",
		},
		"author_id": user.ID,
	})
	defer cursor.Close(r.Context())
	if err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var croom models.Room
		err := cursor.Decode(&croom)
		if err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
		if croom.ID != roomId {
			responseMessage(w, http.StatusBadRequest, "You already have a room by that name")
			return
		}
	}

	result, err := h.Collections.RoomCollection.UpdateByID(r.Context(), roomId, bson.M{
		"$set": bson.M{
			"name":    roomInput.Name,
			"private": roomInput.Private,
		},
	})

	if result.MatchedCount == 0 {
		responseMessage(w, http.StatusBadRequest, "Room not matched")
		return
	}

	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	responseMessage(w, http.StatusOK, "Room updated")
}

func (h handler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rawId := mux.Vars(r)["id"]
	roomId, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var room models.Room
	if err := h.Collections.RoomCollection.FindOne(r.Context(), bson.M{"_id": roomId}).Decode(&room); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Room not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	if _, isProtected := h.ProtectedIDs.Rids[roomId]; isProtected {
		responseMessage(w, http.StatusUnauthorized, "You cannot delete example rooms")
		return
	}

	if room.Author != user.ID {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	res, err := h.Collections.RoomCollection.DeleteOne(r.Context(), bson.M{"_id": roomId})
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if res.DeletedCount == 0 {
		responseMessage(w, http.StatusNotFound, "Not found")
		return
	}

	responseMessage(w, http.StatusOK, "Room deleted")
}

func (h handler) GetRoomImage(w http.ResponseWriter, r *http.Request) {
	rawId := mux.Vars(r)["id"]
	roomId, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var roomImage models.RoomImage
	if err := h.Collections.RoomImageCollection.FindOne(r.Context(), bson.M{"_id": roomId}).Decode(&roomImage); err != nil {
		if err == mongo.ErrNoDocuments {
			responseMessage(w, http.StatusNotFound, "Not found")
		} else {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		}
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(roomImage.Binary.Data)))
	if _, err := w.Write(roomImage.Binary.Data); err != nil {
		log.Println("Unable to write image to response")
	}
}

func (h handler) UploadRoomImage(w http.ResponseWriter, r *http.Request) {
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	if err != nil {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	rawId := mux.Vars(r)["id"]
	roomId, err := primitive.ObjectIDFromHex(rawId)
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var room models.Room
	if h.Collections.RoomCollection.FindOne(r.Context(), bson.M{"_id": roomId}).Decode(&room); err != nil {
		if err != mongo.ErrNoDocuments {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
		} else {
			responseMessage(w, http.StatusNotFound, "Not found")
		}
		return
	}

	if _, isProtected := h.ProtectedIDs.Rids[roomId]; isProtected {
		responseMessage(w, http.StatusUnauthorized, "You cannot modify example rooms")
		return
	}

	if room.Author != user.ID {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	r.ParseMultipartForm(32 << 40)

	file, handler, err := r.FormFile("file")
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	defer file.Close()

	if handler.Size > 20*1024*1024 {
		responseMessage(w, http.StatusRequestEntityTooLarge, "File too large, max 20mb.")
		return
	}

	src, err := handler.Open()
	if err != nil {
		responseMessage(w, http.StatusBadRequest, "Bad request")
		return
	}
	var isJPEG, isPNG bool
	isJPEG = handler.Header.Get("Content-Type") == "image/jpeg"
	isPNG = handler.Header.Get("Content-Type") == "image/png"
	if !isJPEG && !isPNG {
		responseMessage(w, http.StatusBadRequest, "Only JPEG and PNG are supported")
		return
	}
	var img image.Image
	var blurImg image.Image
	var decodeErr error
	if isJPEG {
		img, decodeErr = jpeg.Decode(src)
	} else {
		img, decodeErr = png.Decode(src)
	}
	if decodeErr != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	buf := &bytes.Buffer{}
	blurBuf := &bytes.Buffer{}
	width := img.Bounds().Dx()
	if width > 400 {
		img = resize.Resize(400, 0, img, resize.Lanczos2)
	} else {
		img = resize.Resize(uint(width), 0, img, resize.Lanczos2)
	}
	blurImg = resize.Resize(16, 0, img, resize.Lanczos2)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if err := jpeg.Encode(blurBuf, blurImg, nil); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	imgRes, err := h.Collections.RoomImageCollection.UpdateByID(r.Context(), room.ID, bson.M{"$set": bson.M{"binary": primitive.Binary{Data: buf.Bytes()}}})
	if err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}
	if imgRes.MatchedCount == 0 {
		_, err := h.Collections.RoomImageCollection.InsertOne(r.Context(), models.RoomImage{
			ID:     room.ID,
			Binary: primitive.Binary{Data: buf.Bytes()},
		})
		if err != nil {
			responseMessage(w, http.StatusInternalServerError, "Internal error")
			return
		}
	}

	if h.Collections.RoomCollection.UpdateByID(r.Context(), room.ID, bson.M{
		"$set": bson.M{
			"img_blur":      "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(blurBuf.Bytes()),
			"image_pending": false,
		},
	}); err != nil {
		responseMessage(w, http.StatusInternalServerError, "Internal error")
		return
	}

	responseMessage(w, http.StatusCreated, "Image uploaded")
}
