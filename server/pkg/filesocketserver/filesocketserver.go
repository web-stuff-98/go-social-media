package filesocketserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/socketmodels"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

/*

	NEED TO ADD ERROR HANDLING TO JSON MARSHAL
	------------------------------------------
	NEED TO ADD ERROR HANDLING TO MONGO OPERATIONS
	----------------------------------------------
	NEED TO ADD PANIC RECOVERY
	--------------------------
	NEED TO ADD CLEANUP FUNCTION & SELF CLEANUP AFTER FAILURE & COMPLETION
	----------------------------------------------------------------------


	This is for attachment uploads. It takes in chunks of bytes from the client websocket connection
	when they upload an attachment, it buffers the chunks in memory and saves them to the database every
	2mb.

	The client streams the attachment in chunks through to the file socket endpoint, with the first 24
	bytes of the chunk being the message ID (24 characters)

	The chunk is buffered in the AttachmentChunk map with the attachment ID, when a new chunk comes in it
	appends to the chunk currently stored in the map, but first it checks if the chunk is larger than or
	equal to 2mb, if it is then it saves the chunk the the MongoDB attachment chunk collection instead of
	appending it to the buffer, and clears the buffer.

	The ID of the first chunk will be the same as the ID of the message the attachment is for. Each chunk
	will point to the next chunk, the last chunk will point to nil object id (000000000000)

	This works like GridFS except its my implementation.

	I commented it a lot because its confusing and Its not completely finished
*/

type ConnectionInfo struct {
	Conn *websocket.Conn
	Uid  primitive.ObjectID
}

type BytesInfo struct {
	TotalBytes int
	BytesDone  int
}

type FileSocketServer struct {
	Connections    map[*websocket.Conn]primitive.ObjectID
	RegisterConn   chan ConnectionInfo
	UnregisterConn chan ConnectionInfo

	// oid is the message ID, byte array is the big chunk/final chunk currently being gathered from the smaller chunks
	AttachmentChunks      map[primitive.ObjectID][]byte
	AttachmentChunksChan  chan (*ChunkData)
	AttachmentNextChunkId map[primitive.ObjectID]primitive.ObjectID
	/* The name of the subscription to send progress updates to through the regular socketserver (either inbox or room)
	This is set from the attachment metadata HTTP endpoint */
	AttachmentSubscriptionNames map[primitive.ObjectID][]string
	/* This is also set from the attachment metadata HTTP endpoint  */
	AttachmentBytesProcessed map[primitive.ObjectID]BytesInfo

	FinishAttachmentChan chan (primitive.ObjectID)
}

type ChunkData struct {
	MsgID primitive.ObjectID
	Chunk []byte
}

func Init(socketServer *socketserver.SocketServer, colls *db.Collections) (*FileSocketServer, error) {
	fileSocketServer := &FileSocketServer{
		Connections:    make(map[*websocket.Conn]primitive.ObjectID),
		RegisterConn:   make(chan ConnectionInfo),
		UnregisterConn: make(chan ConnectionInfo),

		AttachmentChunks:            make(map[primitive.ObjectID][]byte),
		AttachmentChunksChan:        make(chan *ChunkData),
		AttachmentNextChunkId:       make(map[primitive.ObjectID]primitive.ObjectID),
		AttachmentSubscriptionNames: make(map[primitive.ObjectID][]string),
		AttachmentBytesProcessed:    make(map[primitive.ObjectID]BytesInfo),

		FinishAttachmentChan: make(chan primitive.ObjectID),
	}
	RunServer(fileSocketServer, socketServer, colls)
	return fileSocketServer, nil
}

func RunServer(fileSocketServer *FileSocketServer, socketServer *socketserver.SocketServer, colls *db.Collections) {
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
				When a chunk comes in, append it to memory... if the chunk is big (2mb) then save it to the chunks collection in the
				database and clear memory, and wait for the next chunks.

				When the client is done uploading the last chunk the server will handle saving it to the database and cleaning up
				the NextChunk ID if there isn't one... its confusing but should work fine.

				Need to add error handling here.
			*/
			chunkData := <-fileSocketServer.AttachmentChunksChan
			if _, ok := fileSocketServer.AttachmentChunks[chunkData.MsgID]; ok {
				// Keep track of number of bytes processed
				fileSocketServer.AttachmentBytesProcessed[chunkData.MsgID] = BytesInfo{
					TotalBytes: fileSocketServer.AttachmentBytesProcessed[chunkData.MsgID].TotalBytes,
					BytesDone:  fileSocketServer.AttachmentBytesProcessed[chunkData.MsgID].BytesDone + len(chunkData.Chunk),
				}
				if len(fileSocketServer.AttachmentChunks[chunkData.MsgID]) > 1024*1024*2 {
					// If the chunk stored in memory is larger than 2mb then move on to saving it to the database
					count, err := colls.AttachmentChunksCollection.CountDocuments(context.Background(), bson.M{"_id": chunkData.MsgID})
					if err != nil {
						log.Panicln("Error finding chunk :", err)
					}
					if count == 0 {
						// Save the first chunk, create the ID for the next chunk, and send the progress update
						nextChunkID := primitive.NewObjectID()
						fileSocketServer.AttachmentNextChunkId[chunkData.MsgID] = nextChunkID
						colls.AttachmentChunksCollection.InsertOne(context.Background(), models.AttachmentChunk{
							ID:        chunkData.MsgID,
							NextChunk: nextChunkID,
							Bytes:     primitive.Binary{Data: append(fileSocketServer.AttachmentChunks[chunkData.MsgID], chunkData.Chunk...)},
						})
						if subscriptionNames, ok := fileSocketServer.AttachmentSubscriptionNames[chunkData.MsgID]; ok {
							outBytes, _ := json.Marshal(socketmodels.OutMessage{
								Type: "ATTACHMENT_PROGRESS",
								Data: `{"ID":"` + chunkData.MsgID.Hex() + `","failed":false,"pending":true,"ratio":` + getProgressString(fileSocketServer.AttachmentBytesProcessed[chunkData.MsgID]) + `}`,
							})
							socketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
								Names: subscriptionNames,
								Data:  outBytes,
							}
							log.Println(subscriptionNames)
						} else {
							log.Println("Not ok")
						}
					} else {
						// Save the chunk and create the ID for the next chunk, and send progress update
						nextChunkID := primitive.NewObjectID()
						colls.AttachmentChunksCollection.InsertOne(context.Background(), models.AttachmentChunk{
							ID:        fileSocketServer.AttachmentNextChunkId[chunkData.MsgID],
							NextChunk: nextChunkID,
							Bytes:     primitive.Binary{Data: append(fileSocketServer.AttachmentChunks[chunkData.MsgID], chunkData.Chunk...)},
						})
						fileSocketServer.AttachmentNextChunkId[chunkData.MsgID] = nextChunkID
						if subscriptionNames, ok := fileSocketServer.AttachmentSubscriptionNames[chunkData.MsgID]; ok {
							outBytes, _ := json.Marshal(socketmodels.OutMessage{
								Type: "ATTACHMENT_PROGRESS",
								Data: `{"ID":"` + chunkData.MsgID.Hex() + `","failed":false,"pending":true,"ratio":` + getProgressString(fileSocketServer.AttachmentBytesProcessed[chunkData.MsgID]) + `}`,
							})
							socketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
								Names: subscriptionNames,
								Data:  outBytes,
							}
							log.Println(subscriptionNames)
						} else {
							log.Println("Not ok")
						}
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
				Here we finalize the upload of the attachment. If the bytes didn't go over the 2mb chunking threshold we save the file as a
				single chunk into the database using the bytes stored in memory. Otherwise we find the last chunk in the database using
				recursion and set the NextChunk value on the last chunk to NilObjectID.
			*/
			msgId := <-fileSocketServer.FinishAttachmentChan
			log.Println("Attachment has finished uploading :", msgId)

			var firstChunk models.AttachmentChunk
			if err := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"_id": msgId}).Decode(&firstChunk); err != nil {
				if err == mongo.ErrNoDocuments {
					// Couldn't find a chunk in the database, so save the bytes that are in memory as the only chunk
					colls.AttachmentChunksCollection.InsertOne(context.Background(), models.AttachmentChunk{
						ID:        msgId,
						NextChunk: primitive.NilObjectID,
						Bytes:     primitive.Binary{Data: fileSocketServer.AttachmentChunks[msgId]},
					})
				} else {
					//Send internal error and panic
				}
			} else {
				// Found the first chunk in the database. Find the last chunk using recursion and nil its NextChunk ID.....
				var nextChunk models.AttachmentChunk
				if err := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"_id": firstChunk.NextChunk}).Decode(&nextChunk); err != nil {
					if err == mongo.ErrNoDocuments {
						// The first chunk is the last chunk... so nil the NextChunkID and send the progress update
						colls.AttachmentChunksCollection.UpdateByID(context.Background(), firstChunk.ID, bson.M{
							"$set": bson.M{"next_id": primitive.NilObjectID},
						})
						if subscriptionNames, ok := fileSocketServer.AttachmentSubscriptionNames[msgId]; ok {
							outBytes, _ := json.Marshal(socketmodels.OutMessage{
								Type: "ATTACHMENT_PROGRESS",
								Data: `{"ID":"` + msgId.Hex() + `","failed":false,"pending":false,"ratio":1}`,
							})
							socketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
								Names: subscriptionNames,
								Data:  outBytes,
							}
							log.Println(subscriptionNames)
						} else {
							log.Println("Not ok")
						}
					} else {
						//Send internal error and panic
					}
				} else {
					// The first chunk isn't the last chunk... so find the last chunk recursively and nil its ObjectID using the recursive function
					// then send the progress update
					if err := recursivelyFindAndNilNextChunkOnLastChunk(&nextChunk.ID, &nextChunk.NextChunk, colls, socketServer, fileSocketServer, &msgId); err != nil {
						//Send internal error and panic
					}
				}
			}
		}
	}()
}

func recursivelyFindAndNilNextChunkOnLastChunk(currentChunkId *primitive.ObjectID, nextChunkId *primitive.ObjectID, colls *db.Collections, ss *socketserver.SocketServer, fss *FileSocketServer, msgId *primitive.ObjectID) error {
	var nextChunk models.AttachmentChunk
	if err := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"_id": nextChunkId}).Decode(&nextChunk); err != nil {
		if err == mongo.ErrNoDocuments {
			if _, err := colls.AttachmentChunksCollection.UpdateByID(context.Background(), currentChunkId, bson.M{
				"$set": bson.M{"next_id": primitive.NilObjectID},
			}); err != nil {
				return err
			}
			if subscriptionNames, ok := fss.AttachmentSubscriptionNames[*msgId]; ok {
				outBytes, err := json.Marshal(socketmodels.OutMessage{
					Type: "ATTACHMENT_PROGRESS",
					Data: `{"ID":"` + msgId.Hex() + `","failed":false,"pending":false,"ratio":1}`,
				})
				if err != nil {
					return err
				}
				ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
					Names: subscriptionNames,
					Data:  outBytes,
				}
			} else {
				return fmt.Errorf("Couldnt find attachmnet subscription names while recursively looking for last chunk")
			}
			return nil
		} else {
			return err
		}
	} else {
		return recursivelyFindAndNilNextChunkOnLastChunk(nextChunkId, &nextChunk.NextChunk, colls, ss, fss, msgId)
	}
}

func getProgressString(info BytesInfo) string {
	return fmt.Sprintf("%v", float32(info.BytesDone)/float32(info.TotalBytes))
}
