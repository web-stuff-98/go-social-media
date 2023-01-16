package filesocketserver

import (
	"log"
	"os"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
	This is for attachment uploads. It takes in chunks of bytes from the client websocket connection
	when they upload an attachment.

	The client streams the attachment in 8mb chunks through to the file socket endpoint, with the first 24
	bytes of the chunk being the message ID (24 characters)

	The chunk is buffered in the AttachmentChunk map with the attachment ID, when a new chunk comes in it
	appends to the chunk currently stored in the map, but first it checks if the chunk is over or equal to
	15mb, if it is then it saves the chunk the the MongoDB attachment chunk collection instead of appending
	it to the buffer.

	The ID of the first chunk will be the same as the ID of the message the attachment is for. Each chunk
	will point to the next chunk, the last chunk will point to nil object id (000000000000)

	This works like GridFS except its my implementation.
*/

type ConnectionInfo struct {
	Conn *websocket.Conn
	Uid  primitive.ObjectID
}

type FileSocketServer struct {
	Connections    map[*websocket.Conn]primitive.ObjectID
	RegisterConn   chan ConnectionInfo
	UnregisterConn chan ConnectionInfo

	// oid is the message ID, byte array is the chunk currently being gathered from the smaller 4mb chunks
	AttachmentChunks     map[primitive.ObjectID][]byte
	AttachmentChunksChan chan (*ChunkData)

	FinishAttachmentChan chan (primitive.ObjectID)
}

type ChunkData struct {
	MsgID primitive.ObjectID
	Chunk []byte
}

func Init() (*FileSocketServer, error) {
	fileSocketServer := &FileSocketServer{
		Connections:    make(map[*websocket.Conn]primitive.ObjectID),
		RegisterConn:   make(chan ConnectionInfo),
		UnregisterConn: make(chan ConnectionInfo),

		AttachmentChunks:     make(map[primitive.ObjectID][]byte),
		AttachmentChunksChan: make(chan *ChunkData),

		FinishAttachmentChan: make(chan primitive.ObjectID),
	}
	RunServer(fileSocketServer)
	return fileSocketServer, nil
}

func RunServer(fileSocketServer *FileSocketServer) {
	/* ----- Connection registration ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in file WS registration : ", r)
				}
			}()
			connData := <-fileSocketServer.RegisterConn
			if connData.Conn != nil {
				fileSocketServer.Connections[connData.Conn] = connData.Uid
			}
		}
	}()
	/* ----- Disconnect registration ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in file WS deregistration : ", r)
				}
			}()
			connData := <-fileSocketServer.UnregisterConn
			for conn := range fileSocketServer.Connections {
				if conn == connData.Conn {
					delete(fileSocketServer.Connections, conn)
					break
				}
			}
		}
	}()
	/* ----- Handle incoming chunk data ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in file WS Chunk handling : ", r)
				}
			}()
			chunkData := <-fileSocketServer.AttachmentChunksChan
			if _, ok := fileSocketServer.AttachmentChunks[chunkData.MsgID]; ok {
				fileSocketServer.AttachmentChunks[chunkData.MsgID] = append(fileSocketServer.AttachmentChunks[chunkData.MsgID], chunkData.Chunk...)
			} else {
				fileSocketServer.AttachmentChunks[chunkData.MsgID] = chunkData.Chunk
			}
		}
	}()
	/* ----- Handle attachment finished uploading event ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in file WS Chunk finished event handling : ", r)
				}
			}()
			msgId := <-fileSocketServer.FinishAttachmentChan
			log.Println("Attachment has finished uploading :", msgId)
			os.WriteFile("/tmp/file.pdf", fileSocketServer.AttachmentChunks[msgId], 0644)
		}
	}()
}
