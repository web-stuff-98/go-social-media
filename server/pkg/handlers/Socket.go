package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/web-stuff-98/go-social-media/pkg/attachmentserver"
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
	Socket event handling.

	Voting and commenting and room invites are done in the API handlers, I could have put that in here but I didn't

	Todo:
	 - rewrite everything so that there aren't tonnes of if statements, indentation and generic error handling
*/

func reader(conn *websocket.Conn, socketServer *socketserver.SocketServer, attachmentServer *attachmentserver.AttachmentServer, uid *primitive.ObjectID, colls *db.Collections) {
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
			// DUPLICATED
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
						// If opening a post page, remove the notifications for replies on the users comments
						if strings.Contains(inMsg.Name, "post_page=") {
							colls.NotificationsCollection.UpdateByID(context.Background(), *uid, bson.M{
								"$pull": bson.M{
									"notifications": bson.M{"type": "REPLY:" + strings.ReplaceAll(inMsg.Name, "post_page=", "")},
								},
							})
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
				// DUPLICATED
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
						// If opening a post page, remove the notifications for replies on the users comments
						if strings.Contains(name, "post_page=") {
							colls.NotificationsCollection.UpdateByID(context.Background(), *uid, bson.M{
								"$pull": bson.M{
									"notifications": bson.M{"type": "REPLY:" + strings.ReplaceAll(name, "post_page=", "")},
								},
							})
						}
					}
				}
			} else if eventType == "OPEN_CONV" {
				var inMsg socketmodels.OpenExitConv
				if err := json.Unmarshal(p, &inMsg); err != nil {
					sendErrorMessageThroughSocket(conn)
				} else {
					if convUid, err := primitive.ObjectIDFromHex(inMsg.Uid); err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						if _, ok := socketServer.OpenConversations[*uid]; ok {
							socketServer.OpenConversations[*uid][convUid] = struct{}{}
						} else {
							convs := make(map[primitive.ObjectID]struct{})
							convs[convUid] = struct{}{}
							socketServer.OpenConversations[*uid] = convs
						}
						// Conversation was opened, remove notifications
						colls.NotificationsCollection.UpdateByID(context.Background(), *uid, bson.M{
							"$pull": bson.M{
								"notifications": bson.M{"type": "MSG:" + inMsg.Uid},
							},
						})
					}
				}
			} else if eventType == "EXIT_CONV" {
				var inMsg socketmodels.OpenExitConv
				if err := json.Unmarshal(p, &inMsg); err != nil {
					sendErrorMessageThroughSocket(conn)
				} else {
					if convUid, err := primitive.ObjectIDFromHex(inMsg.Uid); err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						if _, ok := socketServer.OpenConversations[*uid]; ok {
							delete(socketServer.OpenConversations[*uid], convUid)
						}
					}
				}
			} else if eventType == "PRIVATE_MESSAGE" {
				var inMsg socketmodels.PrivateMessage
				if *uid == primitive.NilObjectID {
					sendErrorMessageThroughSocket(conn)
				} else {
					if err := json.Unmarshal(p, &inMsg); err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						inMsg.Type = "PRIVATE_MESSAGE"
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
								ID:                   primitive.NewObjectIDFromTimestamp(time.Now()),
								Content:              inMsg.Content,
								HasAttachment:        false,
								Uid:                  *uid,
								CreatedAt:            primitive.NewDateTimeFromTime(time.Now()),
								UpdatedAt:            primitive.NewDateTimeFromTime(time.Now()),
								RecipientId:          recipientId,
								IsInvitation:         false,
								IsAcceptedInvitation: false,
								IsDeclinedInvitation: false,
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
								addNotification := true
								if openConvs, ok := socketServer.OpenConversations[recipientId]; ok {
									for oi := range openConvs {
										if oi == *uid {
											// Recipient has conversation open. Don't create the notification
											addNotification = false
											break
										}
									}
								}
								if _, err := colls.InboxCollection.UpdateByID(context.TODO(), recipientId, bson.M{
									"$push": bson.M{
										"messages": msg,
									},
								}); err != nil {
									sendErrorMessageThroughSocket(conn)
								} else {
									if addNotification {
										if _, err := colls.NotificationsCollection.UpdateByID(context.TODO(), recipientId, bson.M{
											"$push": bson.M{
												"notifications": models.Notification{
													Type: "MSG:" + uid.Hex(),
												},
											},
										}); err != nil {
											sendErrorMessageThroughSocket(conn)
										}
									}
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
											socketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
												Names: []string{"inbox=" + recipientId.Hex(), "inbox=" + uid.Hex()},
												Data:  outBytes,
											}
										}
									}
								}
							}
						}
					}
				}
			} else if eventType == "PRIVATE_MESSAGE_DELETE" {
				var inMsg socketmodels.PrivateMessageDelete
				if *uid == primitive.NilObjectID {
					sendErrorMessageThroughSocket(conn)
				} else {
					if err := json.Unmarshal(p, &inMsg); err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						if recipientId, err := primitive.ObjectIDFromHex(inMsg.RecipientId); err != nil {
							sendErrorMessageThroughSocket(conn)
						} else if msgId, err := primitive.ObjectIDFromHex(inMsg.MsgId); err != nil {
							sendErrorMessageThroughSocket(conn)
						} else {
							if _, err := colls.InboxCollection.UpdateByID(context.TODO(), recipientId, bson.M{"$pull": bson.M{"messages": bson.M{"_id": msgId, "author_id": *uid}}}); err != nil {
								sendErrorMessageThroughSocket(conn)
							} else {
								attachmentServer.DeleteChunksChan <- msgId
								data := make(map[string]interface{})
								data["ID"] = msgId.Hex()
								data["recipient_id"] = recipientId.Hex()
								dataBytes, err := json.Marshal(data)
								if err != nil {
									sendErrorMessageThroughSocket(conn)
								} else {
									outBytes, err := json.Marshal(socketmodels.OutMessage{
										Type: "PRIVATE_MESSAGE_DELETE",
										Data: string(dataBytes),
									})
									if err != nil {
										sendErrorMessageThroughSocket(conn)
									} else {
										socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
											Name: "inbox=" + recipientId.Hex(),
											Data: outBytes,
										}
										// Also send the message to the sender
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
			} else if eventType == "PRIVATE_MESSAGE_UPDATE" {
				var inMsg socketmodels.PrivateMessageUpdate
				if *uid == primitive.NilObjectID {
					sendErrorMessageThroughSocket(conn)
				} else {
					if err := json.Unmarshal(p, &inMsg); err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						if recipientId, err := primitive.ObjectIDFromHex(inMsg.RecipientId); err != nil {
							sendErrorMessageThroughSocket(conn)
						} else if msgId, err := primitive.ObjectIDFromHex(inMsg.MsgId); err != nil {
							sendErrorMessageThroughSocket(conn)
						} else {
							if res := colls.InboxCollection.FindOneAndUpdate(context.TODO(), bson.M{
								"_id":          recipientId,
								"messages._id": msgId,
							}, bson.M{
								"$set": bson.M{"messages.$.content": inMsg.Content, "messages.$.invitation": false},
							}); res.Err() != nil {
								sendErrorMessageThroughSocket(conn)
							} else {
								data := make(map[string]interface{})
								data["ID"] = msgId.Hex()
								data["content"] = inMsg.Content
								data["recipient_id"] = recipientId.Hex()
								dataBytes, err := json.Marshal(data)
								if err != nil {
									sendErrorMessageThroughSocket(conn)
								} else {
									outBytes, err := json.Marshal(socketmodels.OutMessage{
										Type: "PRIVATE_MESSAGE_UPDATE",
										Data: string(dataBytes),
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
			} else if eventType == "ACCEPT_INVITATION" {
				var inMsg socketmodels.AcceptDeclineInvitation
				if err := json.Unmarshal(p, &inMsg); err != nil {
					log.Println("A ERR :", err)
					sendErrorMessageThroughSocket(conn)
				} else {
					if senderId, err := primitive.ObjectIDFromHex(inMsg.SenderId); err != nil {
						log.Println("B ERR :", err)
						sendErrorMessageThroughSocket(conn)
					} else {
						if msgId, err := primitive.ObjectIDFromHex(inMsg.MsgId); err != nil {
							log.Println("C ERR :", err)
							sendErrorMessageThroughSocket(conn)
						} else {
							inbox := &models.Inbox{}
							msg := &models.PrivateMessage{}
							foundMsg := false
							if err := colls.InboxCollection.FindOne(context.Background(), bson.M{"_id": uid}).Decode(&inbox); err != nil {
								log.Println("D ERR :", err)
								sendErrorMessageThroughSocket(conn)
							} else {
								for _, pm := range inbox.Messages {
									if pm.ID == msgId {
										msg = &pm
										foundMsg = true
										break
									}
								}
								if !foundMsg {
									log.Println("E ERR :", err)
									sendErrorMessageThroughSocket(conn)
								} else {
									if !msg.IsInvitation || msg.Uid == *uid || msg.Uid != senderId {
										log.Println("F ERR :", err)
										sendErrorMessageThroughSocket(conn)
									} else {
										if res := colls.InboxCollection.FindOneAndUpdate(context.Background(), bson.M{
											"_id":          uid,
											"messages._id": msgId,
										}, bson.M{
											"$set": bson.M{"messages.$.invitation_accepted": true},
										}); res.Err() != nil {
											log.Println("G ERR :", err)
											sendErrorMessageThroughSocket(conn)
										} else {
											if roomId, err := primitive.ObjectIDFromHex(msg.Content); err != nil {
												log.Println("H ERR :", err)
												sendErrorMessageThroughSocket(conn)
											} else {
												if res := colls.RoomPrivateDataCollection.FindOneAndUpdate(context.Background(), bson.M{
													"_id": roomId,
												}, bson.M{
													"$addToSet": bson.M{"members": uid},
													"$pull":     bson.M{"banned": uid},
												}); res.Err() != nil {
													log.Println("I ERR :", err)
													sendErrorMessageThroughSocket(conn)
												} else {
													data := make(map[string]interface{})
													data["ID"] = msgId.Hex()
													data["invitation_accepted"] = true
													data["recipient_id"] = uid.Hex()
													dataBytes, _ := json.Marshal(data)
													outBytes, _ := json.Marshal(socketmodels.OutMessage{
														Type: "PRIVATE_MESSAGE_UPDATE",
														Data: string(dataBytes),
													})
													Names := []string{"inbox=" + uid.Hex(), "inbox=" + msg.Uid.Hex()}
													log.Println(Names)
													socketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
														Names: Names,
														Data:  outBytes,
													}
													outChangeBytes, _ := json.Marshal(socketmodels.OutChangeMessage{
														Type:   "CHANGE",
														Method: "INSERT",
														Entity: "MEMBER",
														Data:   `{"ID":"` + uid.Hex() + `"}`,
													})
													socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
														Name: "room_private_data=" + roomId.Hex(),
														Data: outChangeBytes,
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			} else if eventType == "DECLINE_INVITATION" {
				var inMsg socketmodels.AcceptDeclineInvitation
				if err := json.Unmarshal(p, &inMsg); err != nil {
					sendErrorMessageThroughSocket(conn)
				} else {
					if senderId, err := primitive.ObjectIDFromHex(inMsg.SenderId); err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						if msgId, err := primitive.ObjectIDFromHex(inMsg.MsgId); err != nil {
							sendErrorMessageThroughSocket(conn)
						} else {
							inbox := &models.Inbox{}
							msg := &models.PrivateMessage{}
							foundMsg := false
							if err := colls.InboxCollection.FindOne(context.Background(), bson.M{"_id": uid}).Decode(&inbox); err != nil {
								sendErrorMessageThroughSocket(conn)
							} else {
								for _, pm := range inbox.Messages {
									if pm.ID == msgId {
										msg = &pm
										foundMsg = true
										break
									}
								}
								if !foundMsg {
									sendErrorMessageThroughSocket(conn)
								} else {
									if !msg.IsInvitation || msg.Uid == *uid || msg.Uid != senderId {
										sendErrorMessageThroughSocket(conn)
									} else {
										if res := colls.InboxCollection.FindOneAndUpdate(context.Background(), bson.M{
											"_id":          uid,
											"messages._id": msgId,
										}, bson.M{
											"$set": bson.M{"messages.$.invitation_declined": true},
										}); res.Err() != nil {
											sendErrorMessageThroughSocket(conn)
										} else {
											data := make(map[string]interface{})
											data["ID"] = msgId.Hex()
											data["invitation_declined"] = true
											data["recipient_id"] = uid.Hex()
											dataBytes, _ := json.Marshal(data)
											outBytes, _ := json.Marshal(socketmodels.OutMessage{
												Type: "PRIVATE_MESSAGE_UPDATE",
												Data: string(dataBytes),
											})
											socketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
												Names: []string{"inbox=" + uid.Hex(), "inbox=" + msg.Uid.Hex()},
												Data:  outBytes,
											}
											log.Println("SENT")
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
			} else if eventType == "ROOM_MESSAGE_DELETE" {
				var inMsg socketmodels.RoomMessageDelete
				if *uid == primitive.NilObjectID {
					sendErrorMessageThroughSocket(conn)
				} else {
					if err := json.Unmarshal(p, &inMsg); err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						if roomId, err := primitive.ObjectIDFromHex(inMsg.RoomId); err != nil {
							sendErrorMessageThroughSocket(conn)
						} else if msgId, err := primitive.ObjectIDFromHex(inMsg.MsgId); err != nil {
							sendErrorMessageThroughSocket(conn)
						} else {
							if _, err := colls.RoomMessagesCollection.UpdateByID(context.TODO(), roomId, bson.M{"$pull": bson.M{"messages": bson.M{"_id": msgId, "author_id": *uid}}}); err != nil {
								sendErrorMessageThroughSocket(conn)
							} else {
								attachmentServer.DeleteChunksChan <- msgId
								data := make(map[string]interface{})
								data["ID"] = msgId.Hex()
								dataBytes, err := json.Marshal(data)
								if err != nil {
									sendErrorMessageThroughSocket(conn)
								} else {
									outBytes, err := json.Marshal(socketmodels.OutMessage{
										Type: "ROOM_MESSAGE_DELETE",
										Data: string(dataBytes),
									})
									if err != nil {
										sendErrorMessageThroughSocket(conn)
									} else {
										socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
											Name: "room=" + roomId.Hex(),
											Data: outBytes,
										}
									}
								}
							}
						}
					}
				}
			} else if eventType == "ROOM_MESSAGE_UPDATE" {
				var inMsg socketmodels.RoomMessageUpdate
				if *uid == primitive.NilObjectID {
					sendErrorMessageThroughSocket(conn)
				} else {
					if err := json.Unmarshal(p, &inMsg); err != nil {
						sendErrorMessageThroughSocket(conn)
					} else {
						if roomId, err := primitive.ObjectIDFromHex(inMsg.RoomId); err != nil {
							sendErrorMessageThroughSocket(conn)
						} else if msgId, err := primitive.ObjectIDFromHex(inMsg.MsgId); err != nil {
							sendErrorMessageThroughSocket(conn)
						} else {
							if res := colls.RoomMessagesCollection.FindOneAndUpdate(context.TODO(), bson.M{
								"_id":          roomId,
								"messages._id": msgId,
							}, bson.M{
								"$set": bson.M{"messages.$.content": inMsg.Content},
							}); res.Err() != nil {
								sendErrorMessageThroughSocket(conn)
							} else {
								data := make(map[string]interface{})
								data["ID"] = msgId.Hex()
								data["content"] = inMsg.Content
								dataBytes, err := json.Marshal(data)
								if err != nil {
									sendErrorMessageThroughSocket(conn)
								} else {
									outBytes, err := json.Marshal(socketmodels.OutMessage{
										Type: "ROOM_MESSAGE_UPDATE",
										Data: string(dataBytes),
									})
									if err != nil {
										sendErrorMessageThroughSocket(conn)
									} else {
										socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
											Name: "room=" + roomId.Hex(),
											Data: outBytes,
										}
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
	reader(ws, h.SocketServer, h.AttachmentServer, &uid, h.Collections)
}
