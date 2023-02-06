package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/web-stuff-98/go-social-media/pkg/attachmentserver"
	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/helpers"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

/*
	Socket event handling.

	Voting and commenting are done in the API handlers, I could have put that in here but I didn't

	Todo:
	 - sendErrorMessageThroughSocket with http status code and message
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
			err := HandleSocketEvent(eventType.(string), p, conn, *uid, socketServer, attachmentServer, colls)
			if err != nil {
				sendErrorMessageThroughSocket(conn)
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
