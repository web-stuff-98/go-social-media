package filesocketserver

import (
	"context"
	"log"

	"github.com/gorilla/websocket"
	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
	This is for attachment uploads. It takes in chunks of bytes from the client websocket connection
	when they upload an attachment, it buffers the chunks in memory and saves them to the database every
	8mb.

	The client streams the attachment in chunks through to the file socket endpoint, with the first 24
	bytes of the chunk being the message ID (24 characters)

	The chunk is buffered in the AttachmentChunk map with the attachment ID, when a new chunk comes in it
	appends to the chunk currently stored in the map, but first it checks if the chunk is larger than or
	equal to 8, if it is then it saves the chunk the the MongoDB attachment chunk collection instead of
	appending it to the buffer, and clears the buffer.

	The ID of the first chunk will be the same as the ID of the message the attachment is for. Each chunk
	will point to the next chunk, the last chunk will point to nil object id (000000000000)

	This works like GridFS except its my implementation.

	It can be optimized by changing primitive.ObjectID to just string, so the conversion doesn't happen,
	also the buffer sizes could tweaked.

	I commented it because its confusing because the bytes come in and get stored in memory then are
	chunked and saved into the database if the bytes are over the 8mb threshold.
*/

type ConnectionInfo struct {
	Conn *websocket.Conn
	Uid  primitive.ObjectID
}

type FileSocketServer struct {
	Connections    map[*websocket.Conn]primitive.ObjectID
	RegisterConn   chan ConnectionInfo
	UnregisterConn chan ConnectionInfo

	// oid is the message ID, byte array is the big chunk/final chunk currently being gathered from the smaller chunks
	AttachmentChunks      map[primitive.ObjectID][]byte
	AttachmentChunksChan  chan (*ChunkData)
	AttachmentNextChunkId map[primitive.ObjectID]primitive.ObjectID

	FinishAttachmentChan chan (primitive.ObjectID)
}

type ChunkData struct {
	MsgID primitive.ObjectID
	Chunk []byte
}

func Init(colls *db.Collections) (*FileSocketServer, error) {
	fileSocketServer := &FileSocketServer{
		Connections:    make(map[*websocket.Conn]primitive.ObjectID),
		RegisterConn:   make(chan ConnectionInfo),
		UnregisterConn: make(chan ConnectionInfo),

		AttachmentChunks:      make(map[primitive.ObjectID][]byte),
		AttachmentChunksChan:  make(chan *ChunkData),
		AttachmentNextChunkId: make(map[primitive.ObjectID]primitive.ObjectID),

		FinishAttachmentChan: make(chan primitive.ObjectID),
	}
	RunServer(fileSocketServer, colls)
	return fileSocketServer, nil
}

func RunServer(fileSocketServer *FileSocketServer, colls *db.Collections) {
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
			/*
				When a chunk comes in, append it to memory... if the chunk is big (8mb) then save it to the chunks collection in the
				database and clear memory, and wait for the next chunks.

				When the client is done uploading the last chunk the server will handle saving it to the database and cleaning up
				the NextChunk ID if there isn't one... its confusing but should work fine.

				Need to add error handling here.
			*/
			chunkData := <-fileSocketServer.AttachmentChunksChan
			if _, ok := fileSocketServer.AttachmentChunks[chunkData.MsgID]; ok {
				if len(fileSocketServer.AttachmentChunks[chunkData.MsgID]) > 1024*1024*8 {
					// If the chunk stored in memory is larger than 8mb then move on to saving it to the database
					count, err := colls.AttachmentChunksCollection.CountDocuments(context.Background(), bson.M{"_id": chunkData.MsgID})
					if err != nil {
						log.Panicln("Error finding chunk :", err)
					}
					if count == 0 {
						// Save the FIRST chunk and create the ID for the next big chunk (if there isn't one it will be nulled when finished)
						nextChunkID := primitive.NewObjectID()
						fileSocketServer.AttachmentNextChunkId[chunkData.MsgID] = nextChunkID
						colls.AttachmentChunksCollection.InsertOne(context.Background(), models.AttachmentChunk{
							ID:        chunkData.MsgID,
							NextChunk: nextChunkID,
							Bytes:     primitive.Binary{Data: append(fileSocketServer.AttachmentChunks[chunkData.MsgID], chunkData.Chunk...)},
						})
					} else {
						// Save the chunk and create the ID for the next big chunk (if there isn't one it will be nulled when finished)
						nextChunkID := primitive.NewObjectID()
						colls.AttachmentChunksCollection.InsertOne(context.Background(), models.AttachmentChunk{
							ID:        fileSocketServer.AttachmentNextChunkId[chunkData.MsgID],
							NextChunk: nextChunkID,
							Bytes:     primitive.Binary{Data: append(fileSocketServer.AttachmentChunks[chunkData.MsgID], chunkData.Chunk...)},
						})
						fileSocketServer.AttachmentNextChunkId[chunkData.MsgID] = nextChunkID
					}
				} else {
					// Append the bit of the small chunk into existing chunk memory (not large enough to be saved in the database yet)
					fileSocketServer.AttachmentChunks[chunkData.MsgID] = append(fileSocketServer.AttachmentChunks[chunkData.MsgID], chunkData.Chunk...)
				}
			} else {
				// The very first little chunk of data coming through the socket connection
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
			/*
				Here we finalize the upload of the attachment. If the bytes didn't go over the 8mb chunking threshold we save the file as a
				single chunk into the database using the bytes stored in memory. Otherwise we find the last chunk in the database using
				recursion and set the NextChunk value on the last chunk to NilObjectID.
			*/
			msgId := <-fileSocketServer.FinishAttachmentChan
			log.Println("Attachment has finished uploading :", msgId)

			var firstChunk models.AttachmentChunk
			if err := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"_id": msgId}).Decode(&firstChunk); err != nil {
				if err == mongo.ErrNoDocuments {
					// Couldn't find a chunk in the database, so save the bytes that are in memory as the only chunk (small file).
					colls.AttachmentChunksCollection.InsertOne(context.Background(), models.AttachmentChunk{
						ID:        msgId,
						NextChunk: primitive.NilObjectID,
						Bytes:     primitive.Binary{Data: fileSocketServer.AttachmentChunks[msgId]},
					})
				} else {
					//Send internal error and panic
				}
			} else {
				// Found the next chunk in the database. Find the last chunk using recursion and nil its NextChunk ID.....
				var nextChunk models.AttachmentChunk
				if err := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"next_id": firstChunk.NextChunk}).Decode(&nextChunk); err != nil {
					if err == mongo.ErrNoDocuments {
						//Its the last chunk... so nil the NextChunkID
						colls.AttachmentChunksCollection.UpdateByID(context.Background(), firstChunk.ID, bson.M{
							"$set": bson.M{"next_id": primitive.NilObjectID}})
					} else {
						//Send internal error and panic
					}
				} else {
					//It isn't the last chunk... so find the last chunk recursively and nil its ObjectID using the recursive function
					if err := recursivelyFindAndNilNextChunkOnLastChunk(&nextChunk.ID, &nextChunk.NextChunk, colls); err != nil {
						//Send internal error and panic
					}
				}
			}
		}
	}()
}

func recursivelyFindAndNilNextChunkOnLastChunk(currentChunkId *primitive.ObjectID, nextChunkId *primitive.ObjectID, colls *db.Collections) error {
	var nextChunk models.AttachmentChunk
	if err := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"_id": nextChunkId}).Decode(&nextChunk); err != nil {
		if err == mongo.ErrNoDocuments {
			colls.AttachmentChunksCollection.UpdateByID(context.Background(), currentChunkId, bson.M{
				"$set": bson.M{"next_id": primitive.NilObjectID},
			})
			return nil
		} else {
			return err
		}
	} else {
		return recursivelyFindAndNilNextChunkOnLastChunk(nextChunkId, &nextChunk.NextChunk, colls)
	}
}
