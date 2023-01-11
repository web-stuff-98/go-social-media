package changestreams

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"

	"github.com/web-stuff-98/go-social-media/pkg/db/models"
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

func WatchCollections(DB *mongo.Database, ss *socketserver.SocketServer) {
	go watchUserDeletes(DB, ss)
	go watchUserPfpUpdates(DB, ss)

	go watchPostImageInserts(DB, ss) //Watch for changes in images collection instead of posts collection because we need to wait for the image to be uploaded
	go watchPostImageUpdates(DB, ss)
	go watchPostDeletes(DB, ss)
	go watchPostUpdates(DB, ss)

	go watchRoomDeletes(DB, ss)
}

func watchUserDeletes(db *mongo.Database, ss *socketserver.SocketServer) {
	cs, err := db.Collection("users").Watch(context.Background(), mongo.Pipeline{deletePipeline})
	if err != nil {
		log.Panicln("CS ERR : ", err.Error())
	}
	for cs.Next(context.Background()) {
		var changeEv bson.M
		err := cs.Decode(&changeEv)
		if err != nil {
			log.Fatal(err)
		}
		uid := changeEv["documentKey"].(bson.M)["_id"].(primitive.ObjectID)
		db.Collection("posts").DeleteMany(context.TODO(), bson.M{"author_id": uid})
		db.Collection("rooms").DeleteMany(context.TODO(), bson.M{"author_id": uid})
		db.Collection("pfps").DeleteOne(context.TODO(), bson.M{"id": uid})
		db.Collection("sessions").DeleteOne(context.TODO(), bson.M{"_uid": uid})
		db.Collection("inboxes").DeleteOne(context.TODO(), bson.M{"_id": uid})
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
		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "user=" + uid.Hex(),
			Data: map[string]string{
				"TYPE":   "CHANGE",
				"METHOD": "UPDATE_IMAGE",
				"ENTITY": "USER",
				"DATA":   string(jsonBytes),
			},
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
		db.Collection("post_images").DeleteOne(context.TODO(), bson.M{"_id": postId})
		db.Collection("post_thumbs").DeleteOne(context.TODO(), bson.M{"_id": postId})
		db.Collection("post_votes").DeleteOne(context.TODO(), bson.M{"_id": postId})
		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "post_card=" + postId.Hex(),
			Data: map[string]string{
				"TYPE":   "CHANGE",
				"METHOD": "DELETE",
				"ENTITY": "POST",
				"DATA":   `{"ID":"` + postId.Hex() + `"}`,
			},
		}
		ss.DestroySubscription <- "post_card=" + postId.Hex()
		ss.DestroySubscription <- "post_page=" + postId.Hex()
	}
}

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
		if err := db.Collection("posts").FindOne(context.TODO(), bson.M{"_id": postId}).Decode(&post); err != nil {
			log.Println("CS INSERT DECODE ERROR : ", err)
			return
		}
		data, err := json.Marshal(post)
		if err != nil {
			log.Println("CS MARSHAL ERROR : ", err)
			return
		}
		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "post_feed",
			Data: map[string]string{
				"TYPE":   "CHANGE",
				"METHOD": "INSERT",
				"ENTITY": "POST",
				"DATA":   string(data),
			},
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
		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "post_card=" + postId.Hex(),
			Data: map[string]string{
				"TYPE":   "CHANGE",
				"METHOD": "UPDATE_IMAGE",
				"ENTITY": "POST",
				"DATA":   `{"ID":"` + postId.Hex() + `"}`,
			},
		}
		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "post_page=" + postId.Hex(),
			Data: map[string]string{
				"TYPE":   "CHANGE",
				"METHOD": "UPDATE_IMAGE",
				"ENTITY": "POST",
				"DATA":   `{"ID":"` + postId.Hex() + `"}`,
			},
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
		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "post_card=" + postId.Hex(),
			Data: map[string]string{
				"TYPE":   "CHANGE",
				"METHOD": "UPDATE",
				"ENTITY": "POST",
				"DATA":   string(data),
			},
		}
		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "post_page=" + postId.Hex(),
			Data: map[string]string{
				"TYPE":   "CHANGE",
				"METHOD": "UPDATE",
				"ENTITY": "POST",
				"DATA":   string(data),
			},
		}
	}
}

/*func watchPostInserts(db *mongo.Database, ss *socketserver.SocketServer) {
	cs, err := db.Collection("posts").Watch(context.Background(), mongo.Pipeline{insertPipeline})
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
		log.Println("POST", postId, "WAS CREATED")
		data, err := bson.MarshalExtJSON(changeEv["fullDocument"].(bson.M), true, true)
		ss.SendDataToSubscription <- socketserver.SubscriptionDataMessage{
			Name: "post_feed",
			Data: map[string]string{
				"TYPE":   "CHANGE",
				"METHOD": "INSERT",
				"ENTITY": "POST",
				"DATA":   string(data),
			},
		}
	}
}*/

func watchRoomDeletes(db *mongo.Database, ss *socketserver.SocketServer) {
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
		db.Collection("room_images").DeleteOne(context.TODO(), bson.M{"_id": roomId})
		db.Collection("room_messages").DeleteOne(context.TODO(), bson.M{"_id": roomId})
		ss.DestroySubscription <- "room=" + roomId.Hex()
		ss.DestroySubscription <- "room_card=" + roomId.Hex()
	}
}
