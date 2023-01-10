package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Collections struct {
	UserCollection         *mongo.Collection
	InboxCollection        *mongo.Collection
	SessionCollection      *mongo.Collection
	PfpCollection          *mongo.Collection
	PostCollection         *mongo.Collection
	PostImageCollection    *mongo.Collection
	PostThumbCollection    *mongo.Collection
	PostCommentsCollection *mongo.Collection
	RoomCollection         *mongo.Collection
	RoomImageCollection    *mongo.Collection
}

func Init() (*mongo.Database, Collections) {
	log.Println("Connecting to MongoDB...")
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	DB := client.Database(os.Getenv("MONGODB_DB"))
	colls := Collections{
		UserCollection:         DB.Collection("users"),
		InboxCollection:        DB.Collection("inboxes"),
		SessionCollection:      DB.Collection("sessions"),
		PfpCollection:          DB.Collection("pfps"),
		PostCollection:         DB.Collection("posts"),
		PostImageCollection:    DB.Collection("post_images"),
		PostThumbCollection:    DB.Collection("post_thumbs"),
		PostCommentsCollection: DB.Collection("post_comments"),
		RoomCollection:         DB.Collection("rooms"),
		RoomImageCollection:    DB.Collection("room_images"),
	}
	return DB, colls
}
