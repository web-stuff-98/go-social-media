package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/helpers"
	"github.com/web-stuff-98/go-social-media/pkg/socketmodels"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

/*
	This is where private message and room messages socket event are triggered from, some are triggered from API routes,
	like voting & commenting
*/

func reader(conn *websocket.Conn, socketServer *socketserver.SocketServer, uid *primitive.ObjectID, colls *db.Collections) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		var data map[string]interface{}
		json.Unmarshal(p, &data)

		defer func() {
			r := recover()
			if r != nil {
				log.Println("Recovered from panic in WS reader loop : ", r)
			}
		}()

		eventType, eventTypeOk := data["event_type"]

		if eventTypeOk {
			if eventType == "OPEN_SUBSCRIPTION" || eventType == "CLOSE_SUBSCRIPTION" {
				// Authorization check for private subscriptions is done inside socketServer
				var inMsg socketmodels.OpenCloseSubscription
				if err := json.Unmarshal(p, &inMsg); err != nil {
					sendErrorMessageThroughSocket(conn)
				} else {
					if eventType == "OPEN_SUBSCRIPTION" {
						socketServer.RegisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
							Name: inMsg.Name,
							Uid:  *uid,
							Conn: conn,
						}
					}
					if eventType == "CLOSE_SUBSCRIPTION" {
						socketServer.UnregisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
							Name: inMsg.Name,
							Uid:  *uid,
							Conn: conn,
						}
					}
				}
			} else if eventType == "OPEN_SUBSCRIPTIONS" {
				var inMsg socketmodels.OpenCloseSubscriptions
				if err := json.Unmarshal(p, &inMsg); err != nil {
					sendErrorMessageThroughSocket(conn)
				} else {
					for _, name := range inMsg.Names {
						socketServer.RegisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
							Name: name,
							Uid:  *uid,
							Conn: conn,
						}
					}
				}
			} else if eventType == "PRIVATE_MESSAGE" {
				var inMsg socketmodels.PrivateMessage
				if err := json.Unmarshal(p, &inMsg); err != nil {
					sendErrorMessageThroughSocket(conn)
				} else {
					inBytes, err := json.Marshal(inMsg)
					socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
						Name: "inbox=" + inMsg.RecipientId,
						Data: inBytes,
					}
					recipientId, err := primitive.ObjectIDFromHex(inMsg.RecipientId)
					if err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						msg := &models.PrivateMessage{
							ID:            primitive.NewObjectIDFromTimestamp(time.Now()),
							Content:       inMsg.Content,
							HasAttachment: false,
							Uid:           *uid,
							CreatedAt:     primitive.NewDateTimeFromTime(time.Now()),
							UpdatedAt:     primitive.NewDateTimeFromTime(time.Now()),
							RecipientId:   recipientId,
						}
						if inMsg.HasAttachment {
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
							sendErrorMessageThroughSocket(conn)
						} else {
							if _, err := colls.InboxCollection.UpdateByID(context.TODO(), recipientId, bson.M{
								"$push": bson.M{
									"messages": msg,
								},
							}); err != nil {
								sendErrorMessageThroughSocket(conn)
							} else {
								data, err := json.Marshal(msg)
								if err != nil {
									sendErrorMessageThroughSocket(conn)
								} else {
									outBytes, err := json.Marshal(socketmodels.OutMessage{
										Type: "PRIVATE_MESSAGE",
										Data: string(data),
									})
									if err != nil {
										sendErrorMessageThroughSocket(conn)
									} else {
										socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
											Name: "inbox=" + recipientId.Hex(),
											Data: outBytes,
										}
										// Also send the message to the sender because they need to be able to see their own message
										socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
											Name: "inbox=" + uid.Hex(),
											Data: outBytes,
										}
									}
								}
							}
						}
					}
				}
			} else if eventType == "ROOM_MESSAGE" {
				if *uid == primitive.NilObjectID {
					sendErrorMessageThroughSocket(conn)
				} else {
					var inMsg socketmodels.RoomMessage
					if err := json.Unmarshal(p, &inMsg); err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						roomId, err := primitive.ObjectIDFromHex(inMsg.RoomId)
						if err != nil {
							sendErrorMessageThroughSocket(conn)
						} else {
							msg := &models.RoomMessage{
								ID:            primitive.NewObjectID(),
								Content:       inMsg.Content,
								HasAttachment: inMsg.HasAttachment,
								Uid:           *uid,
								CreatedAt:     primitive.NewDateTimeFromTime(time.Now()),
								UpdatedAt:     primitive.NewDateTimeFromTime(time.Now()),
							}
							if inMsg.HasAttachment {
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
								sendErrorMessageThroughSocket(conn)
							} else {
								data, err := json.Marshal(msg)
								if err != nil {
									sendErrorMessageThroughSocket(conn)
								} else {
									outBytes, err := json.Marshal(socketmodels.OutMessage{
										Type: "ROOM_MESSAGE",
										Data: string(data),
									})
									if err != nil {
										sendErrorMessageThroughSocket(conn)
									} else {
										socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
											Name: "room=" + roomId.Hex(),
											Data: outBytes,
										}
										colls.UserCollection.UpdateByID(context.Background(), uid, bson.M{"$addToSet": bson.M{"rooms_in": roomId}})
									}
								}
							}
						}
					}
				}
			} else if eventType == "VID_SENDING_SIGNAL_IN" {
				var inMsg socketmodels.InVidChatSendingSignal
				if err := json.Unmarshal(p, &inMsg); err != nil {
					sendErrorMessageThroughSocket(conn)
				} else {
					userToSignalID, err := primitive.ObjectIDFromHex(inMsg.UserToSignal)
					if err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						socketServer.SendDataToUser <- socketserver.UserDataMessage{
							Type: "VID_USER_JOINED",
							Uid:  userToSignalID,
							Data: socketmodels.OutVidChatUserJoined{
								CallerUID:  uid.Hex(),
								SignalJSON: inMsg.SignalJSON,
							},
						}
					}
				}
			} else if eventType == "VID_RETURNING_SIGNAL_IN" {
				var inMsg socketmodels.InVidChatReturningSignal
				if err := json.Unmarshal(p, &inMsg); err != nil {
					sendErrorMessageThroughSocket(conn)
				} else {
					callerUID, err := primitive.ObjectIDFromHex(inMsg.CallerUID)
					if err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						socketServer.SendDataToUser <- socketserver.UserDataMessage{
							Type: "VID_RECEIVING_RETURNED_SIGNAL",
							Uid:  callerUID,
							Data: socketmodels.OutVidChatReceivingReturnedSignal{
								SignalJSON: inMsg.SignalJSON,
								UID:        uid.Hex(),
							},
						}
					}
				}
			} else if eventType == "VID_JOIN" {
				var inMsg socketmodels.InVidChatJoin
				if err := json.Unmarshal(p, &inMsg); err != nil {
					sendErrorMessageThroughSocket(conn)
				} else {
					joinID, err := primitive.ObjectIDFromHex(inMsg.JoinID)
					if err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						var allUsers []string
						if inMsg.IsRoom {
							room := &models.Room{}
							if err := colls.RoomCollection.FindOne(context.Background(), bson.M{"_id": joinID}).Decode(&room); err != nil {
								sendErrorMessageThroughSocket(conn)
							} else {
								// Find all the users connected to the room, check if they have video chat
								// open in the room, if they do add to allUsers
								for k, v := range socketServer.Subscriptions {
									if strings.ReplaceAll(k, "room=", "") == inMsg.JoinID {
										for _, oi := range v {
											if oi != *uid {
												for c, oi2 := range socketServer.Connections {
													if oi2 == oi {
														if status, ok := socketServer.VidChatStatus[c]; ok {
															if status.Id == joinID {
																allUsers = append(allUsers, oi.Hex())
															}
														}
														break
													}
												}
											}
										}
										break
									}
								}
							}
						} else {
							// The only other user is the user receiving the direct video.
							// First check if the other user has video chat open in the conversation before
							// forming the WebRTC connection.
							hasOpen := false
							for c, oi := range socketServer.Connections {
								if oi == joinID {
									if status, ok := socketServer.VidChatStatus[c]; ok {
										if status.Id == *uid {
											hasOpen = true
										}
									}
									break
								}
							}
							if hasOpen {
								allUsers = []string{inMsg.JoinID}
							}
						}
						socketServer.VidChatOpenChan <- socketserver.VidChatOpenData{
							Id:   joinID,
							Conn: conn,
						}
						// Send all uids back to conn
						socketServer.SendDataToUser <- socketserver.UserDataMessage{
							Type: "VID_ALL_USERS",
							Uid:  *uid,
							Data: socketmodels.OutVidChatAllUsers{
								UIDs: allUsers,
							},
						}
					}
				}
			} else if eventType == "VID_LEAVE" {
				var inMsg socketmodels.InVidChatLeave
				if err := json.Unmarshal(p, &inMsg); err != nil {
					sendErrorMessageThroughSocket(conn)
				} else {
					id, err := primitive.ObjectIDFromHex(inMsg.ID)
					if err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						if inMsg.IsRoom {
							room := &models.Room{}
							if err := colls.RoomCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&room); err != nil {
								sendErrorMessageThroughSocket(conn)
							} else {
								// Find all the users connected to the room
								for k, v := range socketServer.Subscriptions {
									if strings.ReplaceAll(k, "room=", "") == inMsg.ID {
										for _, oi := range v {
											// Tell all the other users the user has left
											socketServer.SendDataToUser <- socketserver.UserDataMessage{
												Type: "VID_USER_LEFT",
												Uid:  oi,
												Data: socketmodels.OutVidChatUserLeft{
													UID: uid.Hex(),
												},
											}
										}
										break
									}
								}
							}
						} else {
							socketServer.VidChatCloseChan <- conn
							// Tell the other user the user has left
							socketServer.SendDataToUser <- socketserver.UserDataMessage{
								Type: "VID_USER_LEFT",
								Uid:  id,
								Data: socketmodels.OutVidChatUserLeft{
									UID: uid.Hex(),
								},
							}
						}
					}
				}
			} else {
				// eventType is not recognized, send error
				if reflect.TypeOf(eventType).String() == "string" {
					sendErrorMessageThroughSocket(conn)
				}
			}
		} else {
			// eventType was not received. Send error.
			sendErrorMessageThroughSocket(conn)
		}
	}
}

func sendErrorMessageThroughSocket(conn *websocket.Conn) {
	err := conn.WriteJSON(map[string]string{
		"TYPE": "RESPONSE_MESSAGE",
		"DATA": `{"msg":"Socket error","err":true}`,
	})
	if err != nil {
		log.Println(err)
	}
}

func (h handler) WebSocketEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	uid := primitive.NilObjectID
	if user != nil {
		uid = user.ID
	}
	h.SocketServer.RegisterConn <- socketserver.ConnectionInfo{
		Conn:        ws,
		Uid:         uid,
		VidChatOpen: false,
	}
	defer func() {
		h.SocketServer.UnregisterConn <- socketserver.ConnectionInfo{
			Conn:        ws,
			Uid:         uid,
			VidChatOpen: false,
		}
	}()
	reader(ws, h.SocketServer, &uid, h.Collections)
}
