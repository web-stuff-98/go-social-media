package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/web-stuff-98/go-social-media/pkg/attachmentserver"
	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/socketmodels"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
	Moved from Socket.go, because the code was using tonnes of if/else statements for error
	handling and was getting messy looking.
*/

func HandleSocketEvent(eventType string, data []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	switch eventType {
	case "OPEN_SUBSCRIPTION":
		err := openSubscription(data, conn, uid, ss, as, colls)
		return err
	case "CLOSE_SUBSCRIPTION":
		err := closeSubscription(data, conn, uid, ss, as, colls)
		return err
	case "OPEN_SUBSCRIPTIONS":
		err := openSubscriptions(data, conn, uid, ss, as, colls)
		return err
	case "OPEN_CONV":
		err := openConv(data, conn, uid, ss, as, colls)
		return err
	case "EXIT_CONV":
		err := exitConv(data, conn, uid, ss, as, colls)
		return err
	case "PRIVATE_MESSAGE":
		err := privateMessage(data, conn, uid, ss, as, colls)
		return err
	case "PRIVATE_MESSAGE_DELETE":
		err := privateMessageDelete(data, conn, uid, ss, as, colls)
		return err
	case "PRIVATE_MESSAGE_UPDATE":
		err := privateMessageUpdate(data, conn, uid, ss, as, colls)
		return err
	case "ROOM_MESSAGE":
		err := roomMessage(data, conn, uid, ss, as, colls)
		return err
	case "ROOM_MESSAGE_DELETE":
		err := roomMessageDelete(data, conn, uid, ss, as, colls)
		return err
	case "ROOM_MESSAGE_UPDATE":
		err := roomMessageUpdate(data, conn, uid, ss, as, colls)
		return err
	case "VID_SENDING_SIGNAL_IN":
		err := vidSendingSignalIn(data, conn, uid, ss, as, colls)
		return err
	case "VID_RETURNING_SIGNAL_IN":
		err := vidReturningSignalIn(data, conn, uid, ss, as, colls)
		return err
	case "VID_JOIN":
		err := vidJoin(data, conn, uid, ss, as, colls)
		return err
	case "VID_LEAVE":
		err := vidExit(data, conn, uid, ss, as, colls)
		return err
	}
	return fmt.Errorf("Unrecognized event type")
}

func openSubscription(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.OpenCloseSubscription
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	ss.RegisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
		Name: data.Name,
		Uid:  uid,
		Conn: conn,
	}
	return nil
}

func closeSubscription(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.OpenCloseSubscription
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	ss.UnregisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
		Name: data.Name,
		Uid:  uid,
		Conn: conn,
	}
	return nil
}

func openSubscriptions(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.OpenCloseSubscriptions
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	for _, name := range data.Names {
		ss.RegisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
			Name: name,
			Uid:  uid,
			Conn: conn,
		}
	}
	return nil
}

func openConv(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.OpenExitConv
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if convUid, err := primitive.ObjectIDFromHex(data.Uid); err != nil {
		return err
	} else {
		ss.UserOpenConversationWith <- socketserver.UserOpenCloseConversationWith{
			Uid:     uid,
			ConvUid: convUid,
		}
		// Conversation was opened, remove notifications
		colls.NotificationsCollection.UpdateByID(context.Background(), uid, bson.M{
			"$pull": bson.M{
				"notifications": bson.M{"type": "MSG:" + data.Uid},
			},
		})
	}
	return nil
}

func exitConv(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.OpenExitConv
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	if convUid, err := primitive.ObjectIDFromHex(data.Uid); err != nil {
		return err
	} else {
		ss.UserCloseConversationWith <- socketserver.UserOpenCloseConversationWith{
			Uid:     uid,
			ConvUid: convUid,
		}
	}
	return nil
}

func privateMessage(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.PrivateMessage
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	recipientId, err := primitive.ObjectIDFromHex(data.RecipientId)
	if err != nil {
		return err
	}
	msg := &models.PrivateMessage{
		ID:                   primitive.NewObjectIDFromTimestamp(time.Now()),
		Content:              data.Content,
		HasAttachment:        false,
		Uid:                  uid,
		CreatedAt:            primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:            primitive.NewDateTimeFromTime(time.Now()),
		RecipientId:          recipientId,
		IsInvitation:         false,
		IsAcceptedInvitation: false,
		IsDeclinedInvitation: false,
	}
	if data.HasAttachment {
		msg.AttachmentProgress = models.AttachmentProgress{
			Failed:  false,
			Pending: true,
			Ratio:   0,
		}
		msg.HasAttachment = true
	}
	if _, err := colls.InboxCollection.UpdateByID(context.TODO(), uid, bson.M{
		"$addToSet": bson.M{
			"messages_sent_to": recipientId,
		},
	}); err != nil {
		return err
	}

	hasConvsOpenWithRecv := make(chan bool)
	ss.GetUserConversationsOpenWith <- socketserver.GetUserConversationsOpenWith{
		RecvChan: hasConvsOpenWithRecv,
		Uid:      recipientId,
		UidB:     uid,
	}
	hasConvsOpenWith := <-hasConvsOpenWithRecv
	addNotification := !hasConvsOpenWith

	if _, err := colls.InboxCollection.UpdateByID(context.TODO(), recipientId, bson.M{
		"$push": bson.M{
			"messages": msg,
		},
	}); err != nil {
		return err
	} else {
		if addNotification {
			if _, err := colls.NotificationsCollection.UpdateByID(context.TODO(), recipientId, bson.M{
				"$push": bson.M{
					"notifications": models.Notification{
						Type: "MSG:" + uid.Hex(),
					},
				},
			}); err != nil {
				return err
			}
		}
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		outBytes, err := json.Marshal(socketmodels.OutMessage{
			Type: "PRIVATE_MESSAGE",
			Data: string(data),
		})
		if err != nil {
			return err
		}
		ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
			Names: []string{"inbox=" + recipientId.Hex(), "inbox=" + uid.Hex()},
			Data:  outBytes,
		}
	}
	return nil
}

func privateMessageDelete(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.PrivateMessageDelete
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	recipientId, err := primitive.ObjectIDFromHex(data.RecipientId)
	if err != nil {
		return err
	}
	msgId, err := primitive.ObjectIDFromHex(data.MsgId)
	if err != nil {
		return err
	}
	if _, err := colls.InboxCollection.UpdateByID(context.TODO(), recipientId, bson.M{
		"$pull": bson.M{
			"messages": bson.M{
				"_id": msgId,
			},
		},
	}); err != nil {
		return err
	}
	as.DeleteChunksChan <- msgId
	outData := make(map[string]interface{})
	outData["ID"] = msgId.Hex()
	outData["recipient_id"] = recipientId.Hex()
	dataBytes, err := json.Marshal(outData)
	if err != nil {
		return err
	}
	outBytes, err := json.Marshal(socketmodels.OutMessage{
		Type: "PRIVATE_MESSAGE_DELETE",
		Data: string(dataBytes),
	})
	if err != nil {
		return err
	}
	ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
		Names: []string{"inbox=" + recipientId.Hex(), "inbox=" + uid.Hex()},
		Data:  outBytes,
	}
	return nil
}

func privateMessageUpdate(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.PrivateMessageUpdate
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	recipientId, err := primitive.ObjectIDFromHex(data.RecipientId)
	if err != nil {
		return err
	}
	msgId, err := primitive.ObjectIDFromHex(data.MsgId)
	if err != nil {
		return err
	}
	if res := colls.InboxCollection.FindOneAndUpdate(context.TODO(), bson.M{
		"_id":          recipientId,
		"messages._id": msgId,
	}, bson.M{
		"$set": bson.M{"messages.$.content": data.Content, "messages.$.invitation": false},
	}); res.Err() != nil {
		return err
	}
	outData := make(map[string]interface{})
	outData["ID"] = msgId.Hex()
	outData["content"] = data.Content
	outData["recipient_id"] = recipientId.Hex()
	dataBytes, err := json.Marshal(outData)
	if err != nil {
		return err
	}
	outBytes, err := json.Marshal(socketmodels.OutMessage{
		Type: "PRIVATE_MESSAGE_UPDATE",
		Data: string(dataBytes),
	})
	if err != nil {
		return err
	}
	ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
		Names: []string{"inbox=" + recipientId.Hex(), "inbox=" + uid.Hex()},
		Data:  outBytes,
	}
	return nil
}

func roomMessage(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.RoomMessage
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	roomId, err := primitive.ObjectIDFromHex(data.RoomId)
	if err != nil {
		return err
	}
	msg := &models.RoomMessage{
		ID:            primitive.NewObjectID(),
		Content:       data.Content,
		HasAttachment: data.HasAttachment,
		Uid:           uid,
		CreatedAt:     primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:     primitive.NewDateTimeFromTime(time.Now()),
	}
	if data.HasAttachment {
		msg.HasAttachment = true
		msg.AttachmentProgress = models.AttachmentProgress{
			Failed:  false,
			Pending: true,
			Ratio:   0,
		}
	}
	if _, err := colls.RoomMessagesCollection.UpdateByID(context.TODO(), roomId, bson.M{
		"$push": bson.M{
			"messages": msg,
		},
	}); err != nil {
		return err
	}
	outData, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	outBytes, err := json.Marshal(socketmodels.OutMessage{
		Type: "ROOM_MESSAGE",
		Data: string(outData),
	})
	if err != nil {
		return err
	}
	ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
		Name: "room=" + roomId.Hex(),
		Data: outBytes,
	}
	colls.UserCollection.UpdateByID(context.Background(), uid, bson.M{"$addToSet": bson.M{"rooms_messages_in": roomId}})
	return nil
}

func roomMessageDelete(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.RoomMessageDelete
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	roomId, err := primitive.ObjectIDFromHex(data.RoomId)
	if err != nil {
		return err
	}
	msgId, err := primitive.ObjectIDFromHex(data.MsgId)
	if err != nil {
		return err
	}
	if _, err := colls.RoomMessagesCollection.UpdateByID(context.TODO(), roomId, bson.M{
		"$pull": bson.M{
			"messages": bson.M{
				"_id": msgId,
			},
		},
	}); err != nil {
		return err
	}
	as.DeleteChunksChan <- msgId
	outData := make(map[string]interface{})
	outData["ID"] = msgId.Hex()
	dataBytes, err := json.Marshal(outData)
	if err != nil {
		return err
	}
	outBytes, err := json.Marshal(socketmodels.OutMessage{
		Type: "ROOM_MESSAGE_DELETE",
		Data: string(dataBytes),
	})
	if err != nil {
		return err
	}
	ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
		Name: "room=" + roomId.Hex(),
		Data: outBytes,
	}
	return nil
}

func roomMessageUpdate(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.RoomMessageUpdate
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	roomId, err := primitive.ObjectIDFromHex(data.RoomId)
	if err != nil {
		return err
	}
	msgId, err := primitive.ObjectIDFromHex(data.MsgId)
	if err != nil {
		return err
	}
	if res := colls.RoomMessagesCollection.FindOneAndUpdate(context.TODO(), bson.M{
		"_id":          roomId,
		"messages._id": msgId,
	}, bson.M{
		"$set": bson.M{"messages.$.content": data.Content},
	}); res.Err() != nil {
		return err
	}
	outData := make(map[string]interface{})
	outData["ID"] = msgId.Hex()
	outData["content"] = data.Content
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	outBytes, err := json.Marshal(socketmodels.OutMessage{
		Type: "ROOM_MESSAGE_UPDATE",
		Data: string(dataBytes),
	})
	if err != nil {
		return err
	}
	ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
		Name: "room=" + roomId.Hex(),
		Data: outBytes,
	}
	return nil
}

func vidSendingSignalIn(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.InVidChatSendingSignal
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	userToSignalId, err := primitive.ObjectIDFromHex(data.UserToSignal)
	if err != nil {
		return err
	}
	ss.SendDataToUser <- socketserver.UserDataMessage{
		Type: "VID_USER_JOINED",
		Uid:  userToSignalId,
		Data: socketmodels.OutVidChatUserJoined{
			CallerUID:  uid.Hex(),
			SignalJSON: data.SignalJSON,
		},
	}
	return nil
}

func vidReturningSignalIn(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.InVidChatReturningSignal
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	callerUID, err := primitive.ObjectIDFromHex(data.CallerUID)
	if err != nil {
		return err
	}
	ss.SendDataToUser <- socketserver.UserDataMessage{
		Type: "VID_RECEIVING_RETURNED_SIGNAL",
		Uid:  callerUID,
		Data: socketmodels.OutVidChatReceivingReturnedSignal{
			SignalJSON: data.SignalJSON,
			UID:        uid.Hex(),
		},
	}
	return nil
}

func vidJoin(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.InVidChatJoin
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	joinID, err := primitive.ObjectIDFromHex(data.JoinID)
	if err != nil {
		return err
	}
	var allUsers []string
	if data.IsRoom {
		room := &models.Room{}
		if err := colls.RoomCollection.FindOne(context.Background(), bson.M{"_id": joinID}).Decode(&room); err != nil {
			return err
		}
		// Find all the users connected to the room, check if they have video chat
		// open in the room, if they do add to allUsers
		allUsersRecv := make(chan []string)
		ss.VidChatGetAllUsersInRoom <- socketserver.VidChatGetAllUsersInRoom{
			Uid:       uid,
			RoomIdHex: data.JoinID,
			RecvChan:  allUsersRecv,
		}
		received := <-allUsersRecv
		allUsers = received
	} else {
		// The only other user is the user receiving the direct video.
		// First check if the other user has video chat open in the conversation before
		// forming the WebRTC connection.
		allUsersRecv := make(chan []string)
		ss.VidChatGetOtherUserVidOpen <- socketserver.VidChatGetOtherUserVidOpen{
			Uid:      uid,
			UidB:     joinID,
			RecvChan: allUsersRecv,
		}
		received := <-allUsersRecv
		allUsers = received
	}
	ss.VidChatOpenChan <- socketserver.VidChatOpenData{
		Id:   joinID,
		Conn: conn,
	}
	// Send all uids back to conn
	ss.SendDataToUser <- socketserver.UserDataMessage{
		Type: "VID_ALL_USERS",
		Uid:  uid,
		Data: socketmodels.OutVidChatAllUsers{
			UIDs: allUsers,
		},
	}
	return nil
}

func vidExit(b []byte, conn *websocket.Conn, uid primitive.ObjectID, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer, colls *db.Collections) error {
	var data socketmodels.InVidChatLeave
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	id, err := primitive.ObjectIDFromHex(data.ID)
	if err != nil {
		return err
	}
	if data.IsRoom {
		room := &models.Room{}
		if err := colls.RoomCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&room); err != nil {
			return err
		}
		// Find all the users connected to the room and tell them the user left
		var allUsers []string
		allUsersRecv := make(chan []string)
		ss.VidChatGetAllUsersInRoom <- socketserver.VidChatGetAllUsersInRoom{
			Uid:       uid,
			RoomIdHex: id.Hex(),
			RecvChan:  allUsersRecv,
		}
		received := <-allUsersRecv
		allUsers = received
		for _, v := range allUsers {
			if oid, err := primitive.ObjectIDFromHex(v); err != nil {
				return err
			} else {
				ss.SendDataToUser <- socketserver.UserDataMessage{
					Type: "VID_USER_LEFT",
					Uid:  oid,
					Data: socketmodels.OutVidChatUserLeft{
						UID: uid.Hex(),
					},
				}
			}
		}
	} else {
		ss.VidChatCloseChan <- conn
		// Tell the other user the user has left
		ss.SendDataToUser <- socketserver.UserDataMessage{
			Type: "VID_USER_LEFT",
			Uid:  id,
			Data: socketmodels.OutVidChatUserLeft{
				UID: uid.Hex(),
			},
		}
	}
	return nil
}
