package attachmentserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/socketmodels"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
	For attachment uploads
*/

/*--------------- ATTACHMENT SERVER STRUCT ---------------*/
type AttachmentServer struct {
	Uploaders Uploaders

	UploadFailedChan   chan UploadStatusInfo
	UploadCompleteChan chan UploadStatusInfo
	UploadProgressChan chan UploadStatusInfo
	UploadStatusChan   chan UploadStatus

	GetUploaderStatus chan GetUploaderStatus

	DeleteChunksChan chan primitive.ObjectID
}

/*--------------- CHANNEL STRUCTS ---------------*/
type GetUploaderStatus struct {
	RecvChan chan<- map[primitive.ObjectID]Upload
	Uid      primitive.ObjectID
}

/*--------------- OTHER STRUCTS ---------------*/
type Upload struct {
	ChunksDone        int
	TotalChunks       int // +1... starts at 0
	ChunkIDs          []primitive.ObjectID
	SubscriptionNames []string // Where to send progress updates to (inboxes / roomID)
	LastUpdate        time.Time
}

type UploadStatusInfo struct {
	MsgID primitive.ObjectID
	Uid   primitive.ObjectID
}

type UploadStatus struct {
	Status Upload
	MsgId  primitive.ObjectID
	Uid    primitive.ObjectID
}

/*--------------- MUTEX PROTECTED MAPS ---------------*/
type Uploaders struct {
	data  map[primitive.ObjectID]map[primitive.ObjectID]Upload
	mutex sync.RWMutex
}

func Init(colls *db.Collections, SocketServer *socketserver.SocketServer) (*AttachmentServer, error) {
	AttachmentServer := &AttachmentServer{
		Uploaders: Uploaders{
			data: make(map[primitive.ObjectID]map[primitive.ObjectID]Upload),
		},

		UploadFailedChan:   make(chan UploadStatusInfo),
		UploadCompleteChan: make(chan UploadStatusInfo),
		UploadProgressChan: make(chan UploadStatusInfo),
		UploadStatusChan:   make(chan UploadStatus),

		GetUploaderStatus: make(chan GetUploaderStatus),

		DeleteChunksChan: make(chan primitive.ObjectID),
	}
	RunServer(colls, SocketServer, AttachmentServer)
	cleanUp(AttachmentServer, colls)
	return AttachmentServer, nil
}

func RunServer(colls *db.Collections, SocketServer *socketserver.SocketServer, AttachmentServer *AttachmentServer) {
	/* ------ Handle delete attachment chunks ------ */
	go func() {
		for {
			msgId := <-AttachmentServer.DeleteChunksChan
			var metaData models.AttachmentMetadata
			if err := colls.AttachmentMetadataCollection.FindOne(context.Background(), bson.M{"_id": msgId}).Decode(&metaData); err != nil {
				if err == mongo.ErrNoDocuments {
					// If message metadata could not be found, find the first chunk using the message Id instead
					var firstChunk models.AttachmentChunk
					if err := colls.AttachmentChunksCollection.FindOne(context.Background(), bson.M{"_id": msgId}).Decode(&firstChunk); err == nil {
						// Found the chunk. Recursively delete chained chunks
						AttachmentServer.Uploaders.mutex.Lock()
						recursivelyDeleteChunks(firstChunk.ID, colls)
						AttachmentServer.Uploaders.mutex.Unlock()
					}
				}
			} else {
				colls.AttachmentChunksCollection.DeleteMany(context.Background(), bson.M{"_id": bson.M{"$in": metaData.ChunkIDs}})
			}
		}
	}()
	/* ------ Handle attachment failed ------ */
	go func() {
		for {
			info := <-AttachmentServer.UploadFailedChan
			AttachmentServer.Uploaders.mutex.RLock()
			if _, uploaderOk := AttachmentServer.Uploaders.data[info.Uid]; uploaderOk {
				if upload, uploadOk := AttachmentServer.Uploaders.data[info.Uid][info.MsgID]; uploadOk {
					outBytes, _ := json.Marshal(socketmodels.OutMessage{
						Type: "ATTACHMENT_PROGRESS",
						Data: `{"ID":"` + info.MsgID.Hex() + `","failed":true,"pending":false}`,
					})
					SocketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
						Names: upload.SubscriptionNames,
						Data:  outBytes,
					}
				}
			}
			AttachmentServer.Uploaders.mutex.RUnlock()
			colls.AttachmentMetadataCollection.UpdateByID(context.Background(), info.MsgID, bson.M{"$set": bson.M{"failed": true, "pending": false}})
			AttachmentServer.DeleteChunksChan <- info.MsgID
		}
	}()
	/* ------ Handle attachment complete ------ */
	go func() {
		for {
			info := <-AttachmentServer.UploadCompleteChan
			AttachmentServer.Uploaders.mutex.RLock()
			if _, uploaderOk := AttachmentServer.Uploaders.data[info.Uid]; uploaderOk {
				if upload, uploadOk := AttachmentServer.Uploaders.data[info.Uid][info.MsgID]; uploadOk {
					outBytes, _ := json.Marshal(socketmodels.OutMessage{
						Type: "ATTACHMENT_PROGRESS",
						Data: `{"ID":"` + info.MsgID.Hex() + `","failed":false,"pending":false}`,
					})
					SocketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
						Names: upload.SubscriptionNames,
						Data:  outBytes,
					}
				}
			}
			AttachmentServer.Uploaders.mutex.RUnlock()
			colls.AttachmentMetadataCollection.UpdateByID(context.Background(), info.MsgID, bson.M{"$set": bson.M{"failed": false, "pending": false}})
		}
	}()
	/* ------ Handle attachment progress ------ */
	go func() {
		for {
			info := <-AttachmentServer.UploadProgressChan
			AttachmentServer.Uploaders.mutex.RLock()
			if _, uploaderOk := AttachmentServer.Uploaders.data[info.Uid]; uploaderOk {
				if upload, uploadOk := AttachmentServer.Uploaders.data[info.Uid][info.MsgID]; uploadOk {
					outBytes, _ := json.Marshal(socketmodels.OutMessage{
						Type: "ATTACHMENT_PROGRESS",
						Data: `{"ID":"` + info.MsgID.Hex() + `","failed":false,"pending":true,"ratio":` + getProgressString(upload) + `}`,
					})
					SocketServer.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
						Names: upload.SubscriptionNames,
						Data:  outBytes,
					}
				}
			}
			AttachmentServer.Uploaders.mutex.RUnlock()
		}
	}()
	/* ------ Handle set attachment status ------ */
	go func() {
		for {
			data := <-AttachmentServer.UploadStatusChan
			AttachmentServer.Uploaders.mutex.Lock()
			if _, ok := AttachmentServer.Uploaders.data[data.Uid]; !ok {
				AttachmentServer.Uploaders.data[data.Uid] = make(map[primitive.ObjectID]Upload)
			}
			AttachmentServer.Uploaders.data[data.Uid][data.MsgId] = data.Status
			AttachmentServer.Uploaders.mutex.Unlock()
		}
	}()
	/* ------ Handle get uploader status ------ */
	go func() {
		for {
			data := <-AttachmentServer.GetUploaderStatus
			AttachmentServer.Uploaders.mutex.RLock()
			uploads, ok := AttachmentServer.Uploaders.data[data.Uid]
			AttachmentServer.Uploaders.mutex.RUnlock()
			if ok {
				data.RecvChan <- uploads
			} else {
				data.RecvChan <- make(map[primitive.ObjectID]Upload)
			}
		}
	}()
}

func recursivelyDeleteChunks(chunkID primitive.ObjectID, colls *db.Collections) error {
	var chunk models.AttachmentChunk
	res := colls.AttachmentChunksCollection.FindOneAndDelete(context.Background(), bson.M{"_id": chunkID}).Decode(&chunk)
	if res.Error() != "" {
		if res.Error() == mongo.ErrNoDocuments.Error() {
			return nil
		} else {
			return fmt.Errorf(res.Error())
		}
	} else {
		return recursivelyDeleteChunks(chunk.NextChunk, colls)
	}
}

func cleanUp(as *AttachmentServer, colls *db.Collections) {
	cleanupTicker := time.NewTicker(5 * time.Minute)
	quitCleanup := make(chan struct{})
	defer func() {
		quitCleanup <- struct{}{}
	}()
	go func() {
		for {
			select {
			case <-cleanupTicker.C:
				// Go through every Uploader, delete ones that aren't uploading anything
				as.Uploaders.mutex.Lock()
				for oi, v := range as.Uploaders.data {
					if len(v) == 0 {
						delete(as.Uploaders.data, oi)
					} else {
						// If upload info stored in memory hasn't been updated delete it
						for msgId, u := range v {
							if u.LastUpdate.Before(time.Now().Add(-time.Minute * 10)) {
								// If the upload never finished delete the chunks aswell
								if u.ChunksDone < u.TotalChunks-1 {
									recursivelyDeleteChunks(msgId, colls)
								}
								delete(as.Uploaders.data[oi], msgId)
							}
						}
					}
				}
				as.Uploaders.mutex.Unlock()
			case <-quitCleanup:
				cleanupTicker.Stop()
				return
			}
		}
	}()
}

func getProgressString(upload Upload) string {
	return fmt.Sprintf("%v", float32(upload.ChunksDone)/float32(upload.TotalChunks))
}
