package changestreams

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/web-stuff-98/go-social-media/pkg/attachmentserver"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/socketmodels"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var deletePipeline = bson.D{
	{
		Key: "$match", Value: bson.D{
			{Key: "operationType", Value: "delete"},
		},
	},
}
var updatePipeline = bson.D{
	{
		Key: "$match", Value: bson.D{
			{Key: "operationType", Value: "update"},
		},
	},
}
var insertPipeline = bson.D{
	{
		Key: "$match", Value: bson.D{
			{Key: "operationType", Value: "insert"},
		},
	},
}

func WatchCollections(DB *mongo.Database, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer) {
	go watchUserDeletes(DB, ss, as)
	go watchUserPfpUpdates(DB, ss)
	go watchNotificationUpdates(DB, ss)

	go watchPostImageInserts(DB, ss) //Watch for inserts in images collection instead of posts collection because post images are required
	go watchPostImageUpdates(DB, ss)
	go watchPostDeletes(DB, ss)
	go watchPostUpdates(DB, ss)

	go watchRoomInserts(DB, ss)
	go watchRoomImageUpdates(DB, ss)
	go watchRoomDeletes(DB, ss, as)
	go watchRoomUpdates(DB, ss)

	go watchRoomDeletes(DB, ss, as)
}

func watchUserDeletes(db *mongo.Database, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic watching user deletes :", r)
		}
	}()
	cs, err := db.Collection("users").Watch(context.Background(), mongo.Pipeline{deletePipeline}, options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		log.Panicln("CS ERR : ", err.Error())
	}
	for cs.Next(context.Background()) {
		var changeEv struct {
			DocumentKey struct {
				ID primitive.ObjectID `bson:"_id"`
			} `bson:"documentKey"`
		}
		err := cs.Decode(&changeEv)
		if err != nil {
			log.Println("CS DECODE ERROR : ", err)
			return
		}
		uid := changeEv.DocumentKey.ID

		db.Collection("posts").DeleteMany(context.Background(), bson.M{"author_id": uid})
		db.Collection("rooms").DeleteMany(context.Background(), bson.M{"author_id": uid})
		db.Collection("pfps").DeleteOne(context.Background(), bson.M{"_id": uid})
		db.Collection("sessions").DeleteOne(context.Background(), bson.M{"_uid": uid})
		db.Collection("notifications").DeleteOne(context.Background(), bson.M{"_id": uid})

		inbox := &models.Inbox{}
		res := db.Collection("inboxes").FindOne(context.Background(), bson.M{"_id": uid}).Decode(&inbox)
		if res.Error() == "" {
			for _, m := range inbox.Messages {
				if m.HasAttachment {
					as.DeleteChunksChan <- m.ID
					db.Collection("attachment_metadata").DeleteOne(context.Background(), bson.M{"_id": m.ID})
				}
			}
		}
		for _, recipient := range inbox.MessagesSentTo {
			inbox := &models.Inbox{}
			if err := db.Collection("inboxes").FindOne(context.Background(), bson.M{"_id": recipient}).Decode(&inbox); err == nil {
				for _, m := range inbox.Messages {
					if m.HasAttachment && m.Uid == uid {
						as.DeleteChunksChan <- m.ID
						db.Collection("attachment_metadata").DeleteOne(context.Background(), bson.M{"_id": m.ID})
					}
				}
				db.Collection("inboxes").UpdateByID(context.Background(), recipient, bson.M{"$pull": bson.M{"messages": bson.M{"uid": uid}, "messages_sent_to": uid}})
			}
			if err := db.Collection("notifications").FindOne(context.Background(), bson.M{"_id": recipient}).Decode(&inbox); err == nil {
				db.Collection("notifications").UpdateByID(context.Background(), recipient, bson.M{"$pull": bson.M{"notifications": bson.M{"type": "MSG:" + uid.Hex()}}})
			}
		}
		for _, roomId := range changeEv.DocumentKey.ID {
			var roomMsgs models.RoomMessages
			if err := db.Collection("room_messages").FindOneAndUpdate(context.Background(), roomId, bson.M{"$pull": bson.M{"messages": bson.M{"uid": uid}}}).Decode(&roomMsgs); err == nil {
				for _, rm := range roomMsgs.Messages {
					if rm.HasAttachment && rm.Uid == uid {
						as.DeleteChunksChan <- rm.ID
						db.Collection("attachment_metadata").DeleteOne(context.Background(), bson.M{"_id": rm.ID})
					}
				}
			}
		}
	}
}

func watchUserPfpUpdates(db *mongo.Database, ss *socketserver.SocketServer) {
	cs, err := db.Collection("pfps").Watch(context.Background(), mongo.Pipeline{updatePipeline}, options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		log.Panicln("CS ERR : ", err.Error())
	}
	for cs.Next(context.Background()) {
		var changeEv struct {
			DocumentKey struct {
				ID primitive.ObjectID `bson:"_id"`
			} `bson:"documentKey"`
			FullDocument models.Pfp `bson:"fullDocument"`
		}
		err := cs.Decode(&changeEv)
		if err != nil {
			log.Println("CS DECODE ERROR : ", err)
			return
		}
		uid := changeEv.DocumentKey.ID
		pfp := &changeEv.FullDocument
		if err != nil {
			log.Println("CS JSON MARSHAL ERROR : ", err)
			return
		}
		if err != nil {
			log.Println("CS JSON MARSHAL ERROR : ", err)
			return
		}
		pfpB64 := map[string]string{
			"ID":        uid.Hex(),
			"base64pfp": "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(pfp.Binary.Data),
		}
		jsonBytes, err := json.Marshal(pfpB64)
		if err != nil {
			log.Println("CS MARSHAL ERROR : ", err)
			return
		}

		outBytes, err := json.Marshal(socketmodels.OutChangeMessage{
			Type:   "CHANGE",
			Method: "UPDATE_IMAGE",
			Entity: "USER",
			Data:   string(jsonBytes),
		})

		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "user=" + uid.Hex(),
			Data: outBytes,
		}
	}
}

func watchNotificationUpdates(db *mongo.Database, ss *socketserver.SocketServer) {
	cs, err := db.Collection("notifications").Watch(context.Background(), mongo.Pipeline{updatePipeline}, options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		log.Panicln("CS ERR : ", err.Error())
	}
	for cs.Next(context.Background()) {
		var changeEv struct {
			DocumentKey struct {
				ID primitive.ObjectID `bson:"_id"`
			} `bson:"documentKey"`
			FullDocument models.Notifications `bson:"fullDocument"`
		}
		err := cs.Decode(&changeEv)
		if err != nil {
			log.Println("CS DECODE ERROR : ", err)
			return
		}
		outNotificationBytes, err := json.Marshal(changeEv.FullDocument.Notifications)
		outBytes, err := json.Marshal(
			socketmodels.OutMessage{
				Type: "NOTIFICATIONS",
				Data: string(outNotificationBytes),
			},
		)
		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "notifications=" + changeEv.DocumentKey.ID.Hex(),
			Data: outBytes,
		}
	}
}

func watchPostDeletes(db *mongo.Database, ss *socketserver.SocketServer) {
	cs, err := db.Collection("posts").Watch(context.Background(), mongo.Pipeline{deletePipeline})
	if err != nil {
		log.Panicln("CS ERR : ", err.Error())
	}
	for cs.Next(context.Background()) {
		var changeEv bson.M
		err := cs.Decode(&changeEv)
		if err != nil {
			log.Fatal(err)
		}
		postId := changeEv["documentKey"].(bson.M)["_id"].(primitive.ObjectID)

		db.Collection("post_images").DeleteOne(context.Background(), bson.M{"_id": postId})
		db.Collection("post_thumbs").DeleteOne(context.Background(), bson.M{"_id": postId})
		db.Collection("post_votes").DeleteOne(context.Background(), bson.M{"_id": postId})
		db.Collection("post_comments").DeleteOne(context.Background(), bson.M{"_id": postId})

		outBytes, err := json.Marshal(socketmodels.OutChangeMessage{
			Type:   "CHANGE",
			Method: "DELETE",
			Entity: "POST",
			Data:   `{"ID":"` + postId.Hex() + `"}`,
		})

		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "post_card=" + postId.Hex(),
			Data: outBytes,
		}
		ss.DestroySubscription <- "post_card=" + postId.Hex()
		ss.DestroySubscription <- "post_page=" + postId.Hex()
	}
}

/* Watch for post image inserts instead of post inserts, because the image is required*/
func watchPostImageInserts(db *mongo.Database, ss *socketserver.SocketServer) {
	cs, err := db.Collection("post_images").Watch(context.Background(), mongo.Pipeline{insertPipeline})
	if err != nil {
		log.Panicln("CS ERR : ", err.Error())
	}
	for cs.Next(context.Background()) {
		var changeEv bson.M
		err := cs.Decode(&changeEv)
		if err != nil {
			log.Fatal(err)
		}
		postId := changeEv["documentKey"].(bson.M)["_id"].(primitive.ObjectID)
		post := &models.Post{}
		if err := db.Collection("posts").FindOne(context.Background(), bson.M{"_id": postId}).Decode(&post); err != nil {
			log.Println("CS INSERT DECODE ERROR : ", err)
			return
		}
		data, err := json.Marshal(post)
		if err != nil {
			log.Println("CS MARSHAL ERROR : ", err)
			return
		}
		outBytes, err := json.Marshal(socketmodels.OutChangeMessage{
			Type:   "CHANGE",
			Method: "INSERT",
			Entity: "POST",
			Data:   string(data),
		})
		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "post_feed",
			Data: outBytes,
		}
	}
}

func watchPostImageUpdates(db *mongo.Database, ss *socketserver.SocketServer) {
	cs, err := db.Collection("post_images").Watch(context.Background(), mongo.Pipeline{updatePipeline})
	if err != nil {
		log.Panicln("CS ERR : ", err.Error())
	}
	for cs.Next(context.Background()) {
		var changeEv bson.M
		err := cs.Decode(&changeEv)
		if err != nil {
			log.Fatal(err)
		}
		postId := changeEv["documentKey"].(bson.M)["_id"].(primitive.ObjectID)
		outBytes, err := json.Marshal(socketmodels.OutChangeMessage{
			Type:   "CHANGE",
			Method: "UPDATE_IMAGE",
			Entity: "POST",
			Data:   `{"ID":"` + postId.Hex() + `"}`,
		})
		ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
			Names: []string{"post_card=" + postId.Hex(), "post_page=" + postId.Hex()},
			Data:  outBytes,
		}
	}
}

func watchPostUpdates(db *mongo.Database, ss *socketserver.SocketServer) {
	cs, err := db.Collection("posts").Watch(context.Background(), mongo.Pipeline{updatePipeline}, options.ChangeStream().SetFullDocument("updateLookup"))
	if err != nil {
		log.Panicln("CS ERR : ", err.Error())
	}
	for cs.Next(context.Background()) {
		var changeEv struct {
			DocumentKey struct {
				ID primitive.ObjectID `bson:"_id"`
			} `bson:"documentKey"`
			FullDocument models.Post `bson:"fullDocument"`
		}
		err := cs.Decode(&changeEv)
		if err != nil {
			log.Println("CS DECODE ERROR : ", err)
			return
		}
		postId := changeEv.DocumentKey.ID
		post := &changeEv.FullDocument
		data, err := json.Marshal(post)
		if err != nil {
			log.Println("CS JSON MARSHAL ERROR : ", err)
			return
		}
		outBytes, err := json.Marshal(socketmodels.OutChangeMessage{
			Type:   "CHANGE",
			Method: "UPDATE",
			Entity: "POST",
			Data:   string(data),
		})

		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "post_card=" + postId.Hex(),
			Data: outBytes,
		}
		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "post_page=" + postId.Hex(),
			Data: outBytes,
		}
	}
}

func watchRoomDeletes(db *mongo.Database, ss *socketserver.SocketServer, as *attachmentserver.AttachmentServer) {
	cs, err := db.Collection("rooms").Watch(context.Background(), mongo.Pipeline{deletePipeline})
	if err != nil {
		log.Panicln("CS ERR : ", err.Error())
	}
	for cs.Next(context.Background()) {
		var changeEv bson.M
		err := cs.Decode(&changeEv)
		if err != nil {
			log.Fatal(err)
		}
		roomId := changeEv["documentKey"].(bson.M)["_id"].(primitive.ObjectID)
		db.Collection("room_images").DeleteOne(context.Background(), bson.M{"_id": roomId})
		db.Collection("room_private_data").DeleteOne(context.Background(), bson.M{"_id": roomId})

		msgs := &models.RoomMessages{}
		db.Collection("room_messages").FindOneAndDelete(context.Background(), bson.M{"_id": roomId}).Decode(&msgs)
		for _, m := range msgs.Messages {
			if m.HasAttachment {
				as.DeleteChunksChan <- m.ID
				db.Collection("attachment_metadata").DeleteOne(context.Background(), bson.M{"_id": m.ID})
			}
		}

		outBytes, err := json.Marshal(socketmodels.OutChangeMessage{
			Type:   "CHANGE",
			Method: "DELETE",
			Entity: "ROOM",
			Data:   `{"ID":"` + roomId.Hex() + `"}`,
		})

		ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
			Names: []string{"room_card=" + roomId.Hex(), "room=" + roomId.Hex(), "room_feed"},
			Data:  outBytes,
		}

		ss.DestroySubscription <- "room=" + roomId.Hex()
		ss.DestroySubscription <- "room_card=" + roomId.Hex()
	}
}

func watchRoomImageUpdates(db *mongo.Database, ss *socketserver.SocketServer) {
	cs, err := db.Collection("room_images").Watch(context.Background(), mongo.Pipeline{updatePipeline})
	if err != nil {
		log.Panicln("CS ERR : ", err.Error())
	}
	for cs.Next(context.Background()) {
		var changeEv bson.M
		err := cs.Decode(&changeEv)
		if err != nil {
			log.Fatal(err)
		}
		roomId := changeEv["documentKey"].(bson.M)["_id"].(primitive.ObjectID)

		outBytes, err := json.Marshal(socketmodels.OutChangeMessage{
			Type:   "CHANGE",
			Method: "UPDATE_IMAGE",
			Entity: "ROOM",
			Data:   `{"ID":"` + roomId.Hex() + `"}`,
		})

		ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
			Names: []string{"room_card=" + roomId.Hex(), "room=" + roomId.Hex()},
			Data:  outBytes,
		}
	}
}

func watchRoomUpdates(db *mongo.Database, ss *socketserver.SocketServer) {
	cs, err := db.Collection("rooms").Watch(context.Background(), mongo.Pipeline{updatePipeline}, options.ChangeStream().SetFullDocument("updateLookup"))
	if err != nil {
		log.Panicln("CS ERR : ", err.Error())
	}
	for cs.Next(context.Background()) {
		var changeEv struct {
			DocumentKey struct {
				ID primitive.ObjectID `bson:"_id"`
			} `bson:"documentKey"`
			FullDocument models.Room `bson:"fullDocument"`
		}
		err := cs.Decode(&changeEv)
		if err != nil {
			log.Println("CS DECODE ERROR : ", err)
			return
		}
		roomId := changeEv.DocumentKey.ID
		room := &changeEv.FullDocument
		data, err := json.Marshal(room)
		if err != nil {
			log.Println("CS JSON MARSHAL ERROR : ", err)
			return
		}

		outBytes, err := json.Marshal(socketmodels.OutChangeMessage{
			Type:   "CHANGE",
			Method: "UPDATE",
			Entity: "ROOM",
			Data:   string(data),
		})

		ss.SendDataToSubscriptions <- socketserver.SubscriptionDataMessageMulti{
			Names: []string{"room_card=" + roomId.Hex(), "room=" + roomId.Hex()},
			Data:  outBytes,
		}
	}
}

func watchRoomInserts(db *mongo.Database, ss *socketserver.SocketServer) {
	cs, err := db.Collection("rooms").Watch(context.Background(), mongo.Pipeline{insertPipeline}, options.ChangeStream().SetFullDocument("updateLookup"))
	if err != nil {
		log.Panicln("CS ERR : ", err.Error())
	}
	for cs.Next(context.Background()) {
		var changeEv struct {
			DocumentKey struct {
				ID primitive.ObjectID `bson:"_id"`
			} `bson:"documentKey"`
			FullDocument models.Room `bson:"fullDocument"`
		}
		err := cs.Decode(&changeEv)
		if err != nil {
			log.Println("CS DECODE ERROR : ", err)
			log.Fatal(err)
		}
		data, err := bson.Marshal(changeEv.FullDocument)
		if err != nil {
			log.Println("CS JSON MARSHAL ERROR : ", err)
			return
		}

		outBytes, err := json.Marshal(socketmodels.OutChangeMessage{
			Type:   "CHANGE",
			Method: "INSERT",
			Entity: "ROOM",
			Data:   string(data),
		})

		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "room_feed",
			Data: outBytes,
		}
	}
}
