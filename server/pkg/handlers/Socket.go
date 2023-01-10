package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/web-stuff-98/go-social-media/pkg/helpers"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

/*
	Uid can always be left as primitive.NilObjectID, users are not required
	to be authenticated to connect to the socket, join subscriptions or recieve
	messages. Uid is stored with the connection so it's easy to identify users
	that are logged in.
*/

func reader(conn *websocket.Conn, socketServer *socketserver.SocketServer, uid primitive.ObjectID) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		log.Println(string(p))

		var data map[string]interface{}
		json.Unmarshal(p, &data)
		if data["event_type"] != nil {
			if data["event_type"] == "OPEN_SUBSCRIPTION" {
				socketServer.RegisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
					Name: data["name"].(string),
					Uid:  uid,
					Conn: conn,
				}
			}
			if data["event_type"] == "CLOSE_SUBSCRIPTION" {
				socketServer.UnregisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
					Name: data["name"].(string),
					Uid:  uid,
					Conn: conn,
				}
			}
			if data["event_type"] == "OPEN_SUBSCRIPTIONS" {
				names := data["names"].([]interface{})
				for _, name := range names {
					socketServer.RegisterSubscriptionConn <- socketserver.SubscriptionConnectionInfo{
						Name: name.(string),
						Uid:  uid,
						Conn: conn,
					}
				}
			}
		}

		/*if err := conn.WriteMessage(msgType, p); err != nil {
			log.Println(err)
			return
		}*/
	}
}

func (h handler) WebSocketEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	user, _, err := helpers.GetUserAndSessionFromRequest(r, h.Collections)
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
	reader(ws, h.SocketServer, uid)
}
