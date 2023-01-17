package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
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
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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
				// Authorization check for private subscriptions is done inside socketserver
				var inMsg socketmodels.OpenCloseSubscription
				if err := json.Unmarshal(p, &inMsg); err != nil {
					err := conn.WriteJSON(map[string]string{
						"TYPE": "RESPONSE_MESSAGE",
						"DATA": `{"msg":"Bad request","err":true}`,
					})
					if err != nil {
						log.Println(err)
					}
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
			}
			if eventType == "OPEN_SUBSCRIPTIONS" {
				var inMsg socketmodels.OpenCloseSubscriptions
				if err := json.Unmarshal(p, &inMsg); err != nil {
					err := conn.WriteJSON(map[string]string{
						"TYPE": "RESPONSE_MESSAGE",
						"DATA": `{"msg":"Bad request","err":true}`,
					})
					if err != nil {
						log.Println(err)
					}
				} else {
					for _, name := range inMsg.Names {
						socketServer.RegisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
							Name: name,
							Uid:  *uid,
							Conn: conn,
						}
					}
				}
			}
			if eventType == "PRIVATE_MESSAGE" {
				var inMsg socketmodels.PrivateMessage
				if err := json.Unmarshal(p, &inMsg); err != nil {
					err := conn.WriteJSON(map[string]string{
						"TYPE": "RESPONSE_MESSAGE",
						"DATA": `{"msg":"Bad request","err":true}`,
					})
					if err != nil {
						log.Println(err)
					}
				} else {
					inBytes, err := json.Marshal(inMsg)
					socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
						Name: "inbox=" + inMsg.RecipientId,
						Data: inBytes,
					}
					recipientId, err := primitive.ObjectIDFromHex(inMsg.RecipientId)
					if err != nil {
						err := conn.WriteJSON(map[string]string{
							"TYPE": "RESPONSE_MESSAGE",
							"DATA": `{"msg":"Bad request","err":true}`,
						})
						if err != nil {
							log.Println(err)
						}
					} else {
						var msg models.PrivateMessage
						if inMsg.HasAttachment {
							msg = models.PrivateMessage{
								ID:            primitive.NewObjectIDFromTimestamp(time.Now()),
								Content:       inMsg.Content,
								HasAttachment: true,
								Uid:           *uid,
								CreatedAt:     primitive.NewDateTimeFromTime(time.Now()),
								UpdatedAt:     primitive.NewDateTimeFromTime(time.Now()),
								RecipientId:   recipientId,
								AttachmentProgress: models.AttachmentProgress{
									Failed:  false,
									Pending: true,
									Ratio:   0,
								},
							}
						} else {
							msg = models.PrivateMessage{
								ID:            primitive.NewObjectIDFromTimestamp(time.Now()),
								Content:       inMsg.Content,
								HasAttachment: false,
								Uid:           *uid,
								CreatedAt:     primitive.NewDateTimeFromTime(time.Now()),
								UpdatedAt:     primitive.NewDateTimeFromTime(time.Now()),
								RecipientId:   recipientId,
							}
						}
						if _, err := colls.InboxCollection.UpdateByID(context.TODO(), uid, bson.M{
							"$addToSet": bson.M{
								"messages_sent_to": recipientId,
							},
						}); err != nil {
							err := conn.WriteJSON(map[string]string{
								"TYPE": "RESPONSE_MESSAGE",
								"DATA": `{"msg":"Internal error","err":true}`,
							})
							if err != nil {
								log.Println(err)
							}
						} else {
							if _, err := colls.InboxCollection.UpdateByID(context.TODO(), recipientId, bson.M{
								"$push": bson.M{
									"messages": msg,
								},
							}); err != nil {
								err := conn.WriteJSON(map[string]string{
									"TYPE": "RESPONSE_MESSAGE",
									"DATA": `{"msg":"Internal error","err":true}`,
								})
								if err != nil {
									log.Println(err)
								}
							} else {
								data, err := json.Marshal(msg)
								if err != nil {
									err := conn.WriteJSON(map[string]string{
										"TYPE": "RESPONSE_MESSAGE",
										"DATA": `{"msg":"Internal error","err":true}`,
									})
									if err != nil {
										log.Println(err)
									}
								} else {
									outBytes, err := json.Marshal(socketmodels.OutMessage{
										Type: "PRIVATE_MESSAGE",
										Data: string(data),
									})
									if err != nil {
										err := conn.WriteJSON(map[string]string{
											"TYPE": "RESPONSE_MESSAGE",
											"DATA": `{"msg":"Internal error","err":true}`,
										})
										if err != nil {
											log.Println(err)
										}
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
			}
			if eventType == "ROOM_MESSAGE" {
				if *uid == primitive.NilObjectID {
					err := conn.WriteJSON(map[string]string{
						"TYPE": "RESPONSE_MESSAGE",
						"DATA": `{"msg":"Unauthorized","err":true}`,
					})
					if err != nil {
						log.Println(err)
					}
				} else {
					var inMsg socketmodels.RoomMessage
					if err := json.Unmarshal(p, &inMsg); err != nil {
						err := conn.WriteJSON(map[string]string{
							"TYPE": "RESPONSE_MESSAGE",
							"DATA": `{"msg":"Bad request","err":true}`,
						})
						if err != nil {
							log.Println(err)
						}
					} else {
						roomId, err := primitive.ObjectIDFromHex(inMsg.RoomId)
						if err != nil {
							err := conn.WriteJSON(map[string]string{
								"TYPE": "RESPONSE_MESSAGE",
								"DATA": `{"msg":"Invalid ID","err":true}`,
							})
							if err != nil {
								log.Println(err)
							}
						} else {
							msg := &models.RoomMessage{
								ID:            primitive.NewObjectID(),
								Content:       inMsg.Content,
								HasAttachment: inMsg.HasAttachment,
								Uid:           *uid,
								CreatedAt:     primitive.NewDateTimeFromTime(time.Now()),
								UpdatedAt:     primitive.NewDateTimeFromTime(time.Now()),
							}
							if _, err := colls.RoomMessagesCollection.UpdateByID(context.TODO(), roomId, bson.M{
								"$push": bson.M{
									"messages": msg,
								},
							}); err != nil {
								err := conn.WriteJSON(map[string]string{
									"TYPE": "RESPONSE_MESSAGE",
									"DATA": `{"msg":"Internal error","err":true}`,
								})
								if err != nil {
									log.Println(err)
								}
							} else {
								data, err := json.Marshal(msg)
								if err != nil {
									err := conn.WriteJSON(map[string]string{
										"TYPE": "RESPONSE_MESSAGE",
										"DATA": `{"msg":"Internal error","err":true}`,
									})
									if err != nil {
										log.Println(err)
									}
								} else {
									outBytes, err := json.Marshal(socketmodels.OutMessage{
										Type: "ROOM_MESSAGE",
										Data: string(data),
									})
									if err != nil {
										err := conn.WriteJSON(map[string]string{
											"TYPE": "RESPONSE_MESSAGE",
											"DATA": `{"msg":"Internal error","err":true}`,
										})
										if err != nil {
											log.Println(err)
										}
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
			}
		} else {
			err := conn.WriteJSON(map[string]string{
				"TYPE": "RESPONSE_MESSAGE",
				"DATA": `{"msg":"Invalid socket event","err":true}`,
			})
			if err != nil {
				log.Println(err)
			}
		}
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
		Conn: ws,
		Uid:  uid,
	}
	log.Println("Client connected")
	defer func() {
		h.SocketServer.UnregisterConn <- socketserver.ConnectionInfo{
			Conn: ws,
			Uid:  uid,
		}
		log.Println("Client Disconnected")
	}()
	reader(ws, h.SocketServer, &uid, h.Collections)
}
