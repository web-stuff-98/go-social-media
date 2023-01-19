package filesocketserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

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
	This is for attachment uploads. Downloads are handled from API route using octet stream.

	THIS WORKS, BUT SOCKET PROGRESS UPDATES ARE EXTREMELY SLOW!!!
	NEED TO FIX IMPROVE THIS. TRY MOVING STUFF OUT OF MEMORY INTO REDIS.
	TRY ADJUSTING SOCKET BUFFER SIZE.

	The client streams the attachment in chunks through to the file socket endpoint, with the first 24
	bytes of the chunk being the message ID (24 characters)

	The chunk is buffered in the AttachmentChunk map with the message ID, when a new chunk comes in it
	appends to the chunk currently stored in the map, but first it checks if the chunk is larger than or
	equal to 1mb, if it is then it saves the chunk the the MongoDB attachment chunk collection instead of
	appending it to the buffer, and clears the buffer.

	The ID of the first chunk will be the same as the ID of the message the attachment is for. Each chunk
	will point to the next chunk, the last chunk will point to nil object id (000000000000)

	This works like GridFS except its my implementation, and its broken somehow
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

	Chunks                map[primitive.ObjectID][]byte
	ChunksChan            chan (*ChunkData)
	AttachmentNextChunkId map[primitive.ObjectID]primitive.ObjectID
	/* The name of the subscription to send progress updates to through the regular socketserver (either inbox or room)
	This is set from the attachment metadata HTTP endpoint */
	SubscriptionNames map[primitive.ObjectID][]string
	/* This is also set from the attachment metadata HTTP endpoint  */
	BytesProcessed map[primitive.ObjectID]BytesInfo
	ChunkIDs       map[primitive.ObjectID][]primitive.ObjectID
	ChunksDone     map[primitive.ObjectID]int

	SuccessChan             chan (primitive.ObjectID)
	DeleteChunksChan        chan (primitive.ObjectID)
	CleanupAttachmentMemory chan (primitive.ObjectID)
}

type ChunkData struct {
	MsgID primitive.ObjectID
	Chunk []byte
}

func Init(ss *socketserver.SocketServer, colls *db.Collections) (*FileSocketServer, error) {
	fss := &FileSocketServer{
		Connections:    make(map[*websocket.Conn]primitive.ObjectID),
		RegisterConn:   make(chan ConnectionInfo),
		UnregisterConn: make(chan ConnectionInfo),

		Chunks:                make(map[primitive.ObjectID][]byte),
		ChunksChan:            make(chan *ChunkData),
		AttachmentNextChunkId: make(map[primitive.ObjectID]primitive.ObjectID),
		SubscriptionNames:     make(map[primitive.ObjectID][]string),
		BytesProcessed:        make(map[primitive.ObjectID]BytesInfo),
		ChunkIDs:              make(map[primitive.ObjectID][]primitive.ObjectID),
		ChunksDone:            make(map[primitive.ObjectID]int),

		SuccessChan:             make(chan primitive.ObjectID),
		DeleteChunksChan:        make(chan primitive.ObjectID),
		CleanupAttachmentMemory: make(chan primitive.ObjectID),
	}
	RunServer(fss, ss, colls)
	return fss, nil
}

func RunServer(fss *FileSocketServer, ss *socketserver.SocketServer, colls *db.Collections) {
	/* ----- Connection registration ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in file WS registration : ", r)
				}
			}()
			connData := <-fss.RegisterConn
			if connData.Conn != nil {
				fss.Connections[connData.Conn] = connData.Uid
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
			connData := <-fss.UnregisterConn
			for conn := range fss.Connections {
				if conn == connData.Conn {
					delete(fss.Connections, conn)
					break
				}
			}
		}
	}()
	/* ----- Handle incoming chunk data ----- */
	go func() {
		for {
			/*
				When a chunk comes in, append it to memory... if the chunk is big (1mb) then save it to the chunks collection in the
				database and clear memory, and wait for the next chunks.

				When the client is done uploading the last chunk the server will handle saving it to the database and cleaning up
				the NextChunk ID if there isn't one... its confusing but should work fine.

				Need to add error handling here.
			*/
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in file WS Chunk handling : ", r)
				}
			}()
			chunkData := <-fss.ChunksChan

			log.Println("Bytes incoming:", len(chunkData.Chunk))

			if _, ok := fss.Chunks[chunkData.MsgID]; ok {
				fss.BytesProcessed[chunkData.MsgID] = BytesInfo{
					TotalBytes: fss.BytesProcessed[chunkData.MsgID].TotalBytes,
					BytesDone:  fss.BytesProcessed[chunkData.MsgID].BytesDone + len(chunkData.Chunk),
				}
				if len(fss.Chunks[chunkData.MsgID]) > 1048576 {
					// If the chunk stored in memory is larger than 1mb then save it to the database
					res := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"_id": chunkData.MsgID})
					nextChunkID := fss.ChunkIDs[chunkData.MsgID][fss.ChunksDone[chunkData.MsgID]+1]
					if res.Err() == mongo.ErrNoDocuments {
						// Save the first chunk, create the ID for the next chunk, and send the progress update
						fss.AttachmentNextChunkId[chunkData.MsgID] = nextChunkID
						colls.AttachmentChunksCollection.InsertOne(context.Background(), models.AttachmentChunk{
							ID:        chunkData.MsgID,
							NextChunk: nextChunkID,
							Bytes:     primitive.Binary{Data: append(fss.Chunks[chunkData.MsgID], chunkData.Chunk...)},
						})
						go func() {
							if subscriptionNames, ok := fss.SubscriptionNames[chunkData.MsgID]; ok {
								outBytes, _ := json.Marshal(socketmodels.OutMessage{
									Type: "ATTACHMENT_PROGRESS",
									Data: `{"ID":"` + chunkData.MsgID.Hex() + `","failed":false,"pending":true,"ratio":` + getProgressString(fss.BytesProcessed[chunkData.MsgID]) + `}`,
								})
								ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
									Names: subscriptionNames,
									Data:  outBytes,
								}
							}
						}()
						delete(fss.Chunks, chunkData.MsgID)
					} else {
						// Save the chunk and create the ID for the next chunk, and send progress update
						colls.AttachmentChunksCollection.InsertOne(context.Background(), models.AttachmentChunk{
							ID:        fss.AttachmentNextChunkId[chunkData.MsgID],
							NextChunk: nextChunkID,
							Bytes:     primitive.Binary{Data: append(fss.Chunks[chunkData.MsgID], chunkData.Chunk...)},
						})
						go func() {
							if subscriptionNames, ok := fss.SubscriptionNames[chunkData.MsgID]; ok {
								outBytes, _ := json.Marshal(socketmodels.OutMessage{
									Type: "ATTACHMENT_PROGRESS",
									Data: `{"ID":"` + chunkData.MsgID.Hex() + `","failed":false,"pending":true,"ratio":` + getProgressString(fss.BytesProcessed[chunkData.MsgID]) + `}`,
								})
								ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
									Names: subscriptionNames,
									Data:  outBytes,
								}
							}
						}()
						fss.AttachmentNextChunkId[chunkData.MsgID] = nextChunkID
						delete(fss.Chunks, chunkData.MsgID)
					}
					fss.ChunksDone[chunkData.MsgID] = fss.ChunksDone[chunkData.MsgID] + 1
				} else {
					// Append the bit of the small chunk into existing chunk memory (not large enough to be saved in the database yet)
					fss.Chunks[chunkData.MsgID] = append(fss.Chunks[chunkData.MsgID], chunkData.Chunk...)
				}
			} else {
				// The very first little chunk of data coming through the socket connection
				fss.Chunks[chunkData.MsgID] = chunkData.Chunk
				if subscriptionNames, ok := fss.SubscriptionNames[chunkData.MsgID]; ok {
					outBytes, _ := json.Marshal(socketmodels.OutMessage{
						Type: "ATTACHMENT_PROGRESS",
						Data: `{"ID":"` + chunkData.MsgID.Hex() + `","failed":false,"pending":true,"ratio":` + getProgressString(fss.BytesProcessed[chunkData.MsgID]) + `}`,
					})
					ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
						Names: subscriptionNames,
						Data:  outBytes,
					}
				}
			}
		}
	}()
	/* ----- Handle attachment finished uploading event ----- */
	go func() {
		for {
			/*
				Here we finalize the upload of the attachment. If the bytes didn't go over the 1mb chunking threshold we save the file as a
				single chunk into the database using the bytes stored in memory. Otherwise we find the last chunk in the database using
				recursion and set the NextChunk value on the last chunk to NilObjectID.
			*/
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in file WS Chunk finished event handling : ", r)
				}
			}()
			msgId := <-fss.SuccessChan

			log.Println("Success channel")

			metaData := &models.AttachmentMetadata{}
			colls.AttachmentMetadataCollection.FindOne(context.Background(), bson.M{"_id": msgId}).Decode(&metaData)

			var firstChunk models.AttachmentChunk
			if err := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"_id": msgId}).Decode(&firstChunk); err != nil {
				if err == mongo.ErrNoDocuments {
					log.Println("Couldn't find first chunk in database... saving bytes as first and only chunk.")
					// Couldn't find a chunk in the database, so save the bytes that are in memory as the only chunk
					colls.AttachmentChunksCollection.InsertOne(context.Background(), models.AttachmentChunk{
						ID:        msgId,
						NextChunk: primitive.NilObjectID,
						Bytes:     primitive.Binary{Data: fss.Chunks[msgId]},
					})
					log.Println("Wrote chunk")
					if subscriptionNames, ok := fss.SubscriptionNames[msgId]; ok {
						log.Println("Subscription names are ok")
						outBytes, _ := json.Marshal(socketmodels.OutMessage{
							Type: "ATTACHMENT_COMPLETE",
							Data: `{"ID":"` + msgId.Hex() + `","size":` + strconv.Itoa(metaData.Size) + `,"name":"` + metaData.Name + `","length":` + fmt.Sprintf("%f", metaData.VideoLength) + `,"type":"` + metaData.MimeType + `"}`,
						})
						ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
							Names: subscriptionNames,
							Data:  outBytes,
						}
					} else {
						log.Println("Subscription names not ok")
					}
					// Update the metadata
					colls.AttachmentMetadataCollection.UpdateByID(context.Background(), msgId, bson.M{"$set": bson.M{
						"pending": false,
					}})
					log.Println("Updated metadata")
				} else {
					// Internal error
					log.Println("INTERNAL ERROR :", err)
					if subscriptionNames, ok := fss.SubscriptionNames[msgId]; ok {
						log.Println("Subscription names are ok")
						outBytes, _ := json.Marshal(socketmodels.OutMessage{
							Type: "ATTACHMENT_PROGRESS",
							Data: `{"ID":"` + msgId.Hex() + `","failed":true,"pending":false}`,
						})
						ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
							Names: subscriptionNames,
							Data:  outBytes,
						}
					} else {
						log.Println("Subscription names not ok")
					}
					fss.DeleteChunksChan <- msgId
					// Update the metadata
					colls.AttachmentMetadataCollection.UpdateByID(context.Background(), msgId, bson.M{"$set": bson.M{
						"failed":  true,
						"pending": false,
					}})
				}
			} else {
				// Found the first chunk in the database. Save the last chunk (if theres any bytes remaining),
				// then find the last chunk using recursion and nil its NextChunk ID.....
				log.Println("Found first chunk in database... saving remaining bytes as last chunk, if there are any")
				var nextChunk models.AttachmentChunk
				if err := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"_id": firstChunk.NextChunk}).Decode(&nextChunk); err != nil {
					if err == mongo.ErrNoDocuments {
						bytes, ok := fss.Chunks[msgId]
						if ok {
							// There are bytes remaining... save the 2nd chunk
							colls.AttachmentChunksCollection.InsertOne(context.Background(), models.AttachmentChunk{
								ID:        firstChunk.NextChunk,
								NextChunk: primitive.NilObjectID,
								Bytes:     primitive.Binary{Data: bytes},
							})
							log.Println("Wrote remaining bytes as last chunk")
						} else {
							// The first chunk is the last chunk and there aren't any bytes remaining... so nil the NextChunkID and send the progress update
							colls.AttachmentChunksCollection.UpdateByID(context.Background(), firstChunk.ID, bson.M{
								"$set": bson.M{"next_id": primitive.NilObjectID},
							})
							log.Println("Wrote last and only chunk")
						}
						if subscriptionNames, ok := fss.SubscriptionNames[msgId]; ok {
							log.Println("Subscription names are ok")
							outBytes, _ := json.Marshal(socketmodels.OutMessage{
								Type: "ATTACHMENT_COMPLETE",
								Data: `{"ID":"` + msgId.Hex() + `","size":` + strconv.Itoa(metaData.Size) + `,"name":"` + metaData.Name + `","length":` + fmt.Sprintf("%f", metaData.VideoLength) + `,"type":"` + metaData.MimeType + `"}`,
							})
							ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
								Names: subscriptionNames,
								Data:  outBytes,
							}
						} else {
							log.Println("Subscription names not ok")
						}
						// Update the metadata
						colls.AttachmentMetadataCollection.UpdateByID(context.Background(), msgId, bson.M{"$set": bson.M{
							"pending": false,
						}})
						log.Println("Updated metadata")
					} else {
						// Internal error
						log.Println("INTERNAL ERROR :", err)
						if subscriptionNames, ok := fss.SubscriptionNames[msgId]; ok {
							log.Println("Subscription names are ok")
							outBytes, _ := json.Marshal(socketmodels.OutMessage{
								Type: "ATTACHMENT_PROGRESS",
								Data: `{"ID":"` + msgId.Hex() + `","failed":true,"pending":false}`,
							})
							ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
								Names: subscriptionNames,
								Data:  outBytes,
							}
						} else {
							log.Println("Subscription names not ok")
						}
						fss.DeleteChunksChan <- msgId
						// Update the metadata
						colls.AttachmentMetadataCollection.UpdateByID(context.Background(), msgId, bson.M{"$set": bson.M{
							"pending": false,
							"failed":  true,
						}})
					}
				} else {
					// The first chunk isn't the last chunk...
					// If theres no bytes remaining find the last chunk recursively and nil its ObjectID using the recursive function
					// If there are bytes remaining then the recursive function will automatically save the last chunk
					// then send the progress update
					if err := finalizeChunksChain(&nextChunk.ID, &nextChunk.NextChunk, colls, ss, fss, &msgId, *metaData); err != nil {
						log.Println("INTERNAL ERROR FROM CHUNK CHAIN :", err)
						if subscriptionNames, ok := fss.SubscriptionNames[msgId]; ok {
							log.Println("Subscription names are ok")
							outBytes, _ := json.Marshal(socketmodels.OutMessage{
								Type: "ATTACHMENT_PROGRESS",
								Data: `{"ID":"` + msgId.Hex() + `","failed":true,"pending":false}`,
							})
							ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
								Names: subscriptionNames,
								Data:  outBytes,
							}
						} else {
							log.Println("Subscription names not ok")
						}
						fss.DeleteChunksChan <- msgId
						// Update the metadata
						colls.AttachmentMetadataCollection.UpdateByID(context.Background(), msgId, bson.M{"$set": bson.M{
							"failed":  true,
							"pending": false,
						}})
					} else {
						log.Println("Wrote chunk")
						if subscriptionNames, ok := fss.SubscriptionNames[msgId]; ok {
							log.Println("Subscription names are ok")
							outBytes, _ := json.Marshal(socketmodels.OutMessage{
								Type: "ATTACHMENT_COMPLETE",
								Data: `{"ID":"` + msgId.Hex() + `","size":` + strconv.Itoa(metaData.Size) + `,"name":"` + metaData.Name + `","length":` + fmt.Sprintf("%f", metaData.VideoLength) + `,"type":"` + metaData.MimeType + `"}`,
							})
							ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
								Names: subscriptionNames,
								Data:  outBytes,
							}
						} else {
							log.Println("Subscription names not ok")
						}
						// Update the metadata
						colls.AttachmentMetadataCollection.UpdateByID(context.Background(), msgId, bson.M{"$set": bson.M{
							"pending": false,
						}})
						log.Println("Updated metadata")
					}
				}
			}

			fss.CleanupAttachmentMemory <- msgId
		}
	}()
	/* ----- Cleanup memory after failing / completing upload ----- */
	go func() {
		for {
			oid := <-fss.CleanupAttachmentMemory
			delete(fss.Chunks, oid)
			delete(fss.AttachmentNextChunkId, oid)
			delete(fss.SubscriptionNames, oid)
			delete(fss.BytesProcessed, oid)
			delete(fss.ChunkIDs, oid)
			delete(fss.ChunksDone, oid)
		}
	}()
	/* ----- Handle attachment failed uploading event ----- */
	go func() {
		for {
			oid := <-fss.DeleteChunksChan
			var firstChunk models.AttachmentChunk
			err := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(&firstChunk)
			if err == nil {
				recursivelyFindAndDeleteChunks(&firstChunk.ID, &firstChunk.NextChunk, colls, ss, fss, &firstChunk.ID, models.AttachmentMetadata{})
			}
			fss.CleanupAttachmentMemory <- oid
		}
	}()
}

func finalizeChunksChain(currentChunkId *primitive.ObjectID, nextChunkId *primitive.ObjectID, colls *db.Collections, ss *socketserver.SocketServer, fss *FileSocketServer, msgId *primitive.ObjectID, metaData models.AttachmentMetadata) error {
	var nextChunk models.AttachmentChunk
	if err := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"_id": nextChunkId}).Decode(&nextChunk); err != nil {
		if err == mongo.ErrNoDocuments {
			if remainingBytes, ok := fss.Chunks[*msgId]; ok {
				// If theres remaining bytes save them as the last chunk with nil NextChunkID
				colls.AttachmentChunksCollection.InsertOne(context.Background(), models.AttachmentChunk{
					ID:        *nextChunkId,
					NextChunk: primitive.NilObjectID,
					Bytes:     primitive.Binary{Data: remainingBytes},
				})
			} else {
				// If theres no remaining bytes then just nil the NextChunkID
				if _, err := colls.AttachmentChunksCollection.UpdateByID(context.Background(), currentChunkId, bson.M{
					"$set": bson.M{"next_id": primitive.NilObjectID},
				}); err != nil {
					return err
				}
			}
			if subscriptionNames, ok := fss.SubscriptionNames[*msgId]; ok {
				outBytes, err := json.Marshal(socketmodels.OutMessage{
					Type: "ATTACHMENT_COMPLETE",
					Data: `{"ID":"` + msgId.Hex() + `","size":` + strconv.Itoa(metaData.Size) + `,"name":"` + metaData.Name + `","length":` + fmt.Sprintf("%f", metaData.VideoLength) + `,"type":"` + metaData.MimeType + `"}`,
				})
				if err != nil {
					return err
				}
				go func() {
					ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
						Names: subscriptionNames,
						Data:  outBytes,
					}
				}()
			} else {
				return fmt.Errorf("Couldnt find attachment subscription names while recursively looking for last chunk")
			}
			return nil
		} else {
			return err
		}
	} else {
		return finalizeChunksChain(nextChunkId, &nextChunk.NextChunk, colls, ss, fss, msgId, metaData)
	}
}

func recursivelyFindAndDeleteChunks(currentChunkId *primitive.ObjectID, nextChunkId *primitive.ObjectID, colls *db.Collections, ss *socketserver.SocketServer, fss *FileSocketServer, msgId *primitive.ObjectID, metaData models.AttachmentMetadata) error {
	var nextChunk models.AttachmentChunk
	if err := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"_id": nextChunkId}).Decode(&nextChunk); err != nil {
		if err == mongo.ErrNoDocuments {
			colls.AttachmentChunksCollection.DeleteOne(context.Background(), currentChunkId)
			return nil
		} else {
			return err
		}
	} else {
		colls.AttachmentChunksCollection.DeleteOne(context.Background(), currentChunkId)
		return recursivelyFindAndDeleteChunks(nextChunkId, &nextChunk.NextChunk, colls, ss, fss, msgId, metaData)
	}
}

func getProgressString(info BytesInfo) string {
	return fmt.Sprintf("%v", float32(info.BytesDone)/float32(info.TotalBytes))
}
