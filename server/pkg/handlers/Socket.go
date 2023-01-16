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

		if data["event_type"] != nil {
			if data["event_type"] == "OPEN_SUBSCRIPTION" {
				// Authorization check for private subscriptions is done inside socketserver
				socketServer.RegisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
					Name: data["name"].(string),
					Uid:  *uid,
					Conn: conn,
				}
			}
			if data["event_type"] == "CLOSE_SUBSCRIPTION" {
				socketServer.UnregisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
					Name: data["name"].(string),
					Uid:  *uid,
					Conn: conn,
				}
			}
			if data["event_type"] == "OPEN_SUBSCRIPTIONS" {
				names := data["names"].([]interface{})
				for _, name := range names {
					socketServer.RegisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
						Name: name.(string),
						Uid:  *uid,
						Conn: conn,
					}
				}
			}
			if data["event_type"] == "PRIVATE_MESSAGE" {
				socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
					Name: "inbox=" + data["recipient_id"].(string),
					Data: map[string]string{
						"content": data["content"].(string),
					},
				}
				recipientId, err := primitive.ObjectIDFromHex(data["recipient_id"].(string))
				if err != nil {
					err := conn.WriteJSON(map[string]string{
						"TYPE": "RESPONSE_MESSAGE",
						"DATA": `{"msg":"Internal error","err":true}`,
					})
					if err != nil {
						log.Println(err)
					}
				} else {
					msg := &models.PrivateMessage{
						ID:          primitive.NewObjectIDFromTimestamp(time.Now()),
						Content:     data["content"].(string),
						Uid:         *uid,
						CreatedAt:   primitive.NewDateTimeFromTime(time.Now()),
						UpdatedAt:   primitive.NewDateTimeFromTime(time.Now()),
						RecipientId: recipientId,
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
								socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
									Name: "inbox=" + recipientId.Hex(),
									Data: map[string]string{
										"TYPE": "PRIVATE_MESSAGE",
										"DATA": string(data),
									},
								}
								// Also send the message to the sender because they need to be able to see their own message
								socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
									Name: "inbox=" + uid.Hex(),
									Data: map[string]string{
										"TYPE": "PRIVATE_MESSAGE",
										"DATA": string(data),
									},
								}
							}
						}
					}
				}
			}
			if data["event_type"] == "ROOM_MESSAGE" {
				if *uid == primitive.NilObjectID {
					err := conn.WriteJSON(map[string]string{
						"TYPE": "RESPONSE_MESSAGE",
						"DATA": `{"msg":"Unauthorized","err":true}`,
					})
					if err != nil {
						log.Println(err)
					}
				} else {
					if data["content"] == nil {
						err := conn.WriteJSON(map[string]string{
							"TYPE": "RESPONSE_MESSAGE",
							"DATA": `{"msg":"Bad request","err":true}`,
						})
						if err != nil {
							log.Println(err)
						}
					} else {
						roomId, err := primitive.ObjectIDFromHex(data["room_id"].(string))
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
								ID:                primitive.NewObjectID(),
								Content:           data["content"].(string),
								Uid:               *uid,
								CreatedAt:         primitive.NewDateTimeFromTime(time.Now()),
								UpdatedAt:         primitive.NewDateTimeFromTime(time.Now()),
								HasAttachment:     false,
								AttachmentPending: false,
								AttachmentType:    "",
								AttachmentError:   false,
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
									socketServer.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
										Name: "room=" + roomId.Hex(),
										Data: map[string]string{
											"TYPE": "ROOM_MESSAGE",
											"DATA": string(data),
										},
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
