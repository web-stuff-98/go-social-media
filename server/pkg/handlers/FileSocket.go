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
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

/*
	This is for file attachment uploads. Files are streamed from the client in chunks of bytes.
	Only logged in users can connect to this endpoint. The ID of the message the attachment is
	for is the first 24 bytes of the message, so that the message the file bytes belong to can be
	identified.

	The chunks are sent off to be handled by the fileSocketServer
*/

func fileWsReader(conn *websocket.Conn, uid *primitive.ObjectID, fileSocketServer *filesocketserver.FileSocketServer) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		idBytes := p[:24]
		msgId, err := primitive.ObjectIDFromHex(string(idBytes))

		size := binary.Size(p)

		log.Println("Reading", size, "bytes")

		if size == 24 {
			// If the binary size is 24 (just the ID) it means its has finished uploading
			fileSocketServer.FinishAttachmentChan <- msgId
		} else if size > 24 {
			fileSocketServer.AttachmentChunksChan <- &filesocketserver.ChunkData{
				MsgID: msgId,
				Chunk: p[24:],
			}
		}
		// If the size is less than 24, that cannot be possible, so don't do anything

		defer func() {
			r := recover()
			if r != nil {
				log.Println("Recovered from panic in File WS reader loop : ", r)
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
