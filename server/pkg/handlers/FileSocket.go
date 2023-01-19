package handlers

import (
	"encoding/binary"
	"log"
	"net/http"

	"github.com/web-stuff-98/go-social-media/pkg/filesocketserver"
	"github.com/web-stuff-98/go-social-media/pkg/helpers"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var fileWsUpgrader = websocket.Upgrader{
	ReadBufferSize:  8096,
	WriteBufferSize: 8096,
}

/*
	This is for file attachment uploads. Files are streamed from the client in chunks of bytes.
	Only logged in users can connect to this endpoint. The ID of the message the attachment is
	for is the first 24 bytes of the message, so that the message the file bytes belong to can be
	identified.

	The chunks are sent off to be handled by the fileSocketServer.

	When the client is done uploading the attachment it will send the message ID on its own,
	then the server will finalize the upload.
*/

func fileWsReader(conn *websocket.Conn, uid *primitive.ObjectID, fileSocketServer *filesocketserver.FileSocketServer) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		size := binary.Size(p)

		idBytes := p[:24]
		msgId, err := primitive.ObjectIDFromHex(string(idBytes))

		if size == 24 {
			// If the binary size is 24 (just the ID) it means it has finished uploading
			log.Println("Sent to success channel")
			fileSocketServer.SuccessChan <- msgId
		} else if size > 24 && size <= 8096 {
			fileSocketServer.ChunksChan <- &filesocketserver.ChunkData{
				MsgID: msgId,
				Chunk: p[24:],
			}
		} else {
			// If the size is less than 24, or larger than it should be, that shouldn't be possible, so break the connection.
			log.Println("Connection broken")
			break
		}

		defer func() {
			r := recover()
			if r != nil {
				log.Println("Recovered from panic in File WS reader loop : ", r)
			} else {
				conn.Close()
			}
		}()
	}
}

func (h handler) FileWebSocketEndpoint(w http.ResponseWriter, r *http.Request) {
	fileWsUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := fileWsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	user, _, err := helpers.GetUserAndSessionFromRequest(r, *h.Collections)
	uid := primitive.NilObjectID
	if user != nil {
		uid = user.ID
	} else {
		responseMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	h.FileSocketServer.RegisterConn <- filesocketserver.ConnectionInfo{
		Conn: ws,
		Uid:  uid,
	}
	log.Println("Client connected")
	defer func() {
		h.FileSocketServer.UnregisterConn <- filesocketserver.ConnectionInfo{
			Conn: ws,
			Uid:  uid,
		}
		log.Println("Client Disconnected")
	}()
	fileWsReader(ws, &user.ID, h.FileSocketServer)
}
