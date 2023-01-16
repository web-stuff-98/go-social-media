package db

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/exp/maps"
)

/*
	Also included is a cleanup function. The cleanup function removes comments
	that are parented to comments that no longer exist, and votes that are parented to
	comments that no longer exist, to save space. It also deletes users accounts after
	20 minutes, along with all their stuff (i havent wrote that last bit yet though)
*/

type Collections struct {
	UserCollection    *mongo.Collection
	InboxCollection   *mongo.Collection
	SessionCollection *mongo.Collection
	PfpCollection     *mongo.Collection

	PostCollection         *mongo.Collection
	PostVoteCollection     *mongo.Collection
	PostImageCollection    *mongo.Collection
	PostThumbCollection    *mongo.Collection
	PostCommentsCollection *mongo.Collection

	RoomCollection         *mongo.Collection
	RoomMessagesCollection *mongo.Collection
	RoomImageCollection    *mongo.Collection

	AttachmentMetadataCollection *mongo.Collection
	AttachmentChunksCollection   *mongo.Collection
}

func Init() (*mongo.Database, *Collections) {
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
	colls := &Collections{
		UserCollection:    DB.Collection("users"),
		InboxCollection:   DB.Collection("inboxes"),
		SessionCollection: DB.Collection("sessions"),
		PfpCollection:     DB.Collection("pfps"),

		PostCollection:         DB.Collection("posts"),
		PostVoteCollection:     DB.Collection("post_votes"),
		PostImageCollection:    DB.Collection("post_images"),
		PostThumbCollection:    DB.Collection("post_thumbs"),
		PostCommentsCollection: DB.Collection("post_comments"),

		RoomCollection:         DB.Collection("rooms"),
		RoomMessagesCollection: DB.Collection("room_messages"),
		RoomImageCollection:    DB.Collection("room_images"),

		AttachmentMetadataCollection: DB.Collection("attachment_metadata"),
		AttachmentChunksCollection:   DB.Collection("attachment_chunks"),
	}
	cleanUp(colls)
	return DB, colls
}

func cleanUp(colls *Collections) {
	cleanupTicker := time.NewTicker(24 * time.Hour)
	quitCleanup := make(chan struct{})
	go func() {
		for {
			select {
			case <-cleanupTicker.C:
				cleanUpPosts(colls)
			case <-quitCleanup:
				cleanupTicker.Stop()
				return
			}
		}
	}()
}

func getChildrenOfOrphanedComment(orphanId string, cmts map[string]string) map[string]struct{} {
	childIds := make(map[string]struct{})
	for commentId, parentId := range cmts {
		if parentId == orphanId {
			childIds[commentId] = struct{}{}
		}
	}
	for childId, _ := range childIds {
		childChildren := getChildrenOfOrphanedComment(childId, cmts)
		maps.Copy(childChildren, childIds)
	}
	return childIds
}

func cleanUpPosts(colls *Collections) {
	cmtsCursor, err := colls.PostCommentsCollection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Fatal("ERROR IN POSTS CLEANUP CURSOR :", err)
	}
	for cmtsCursor.Next(context.Background()) {
		postCmts := &models.PostComments{}
		cmtsCursor.Decode(&postCmts)
		allCmts := make(map[string]string)
		for _, c := range postCmts.Comments {
			allCmts[c.ID.Hex()] = c.ParentID
		}
		// Get all orphaned comments
		orphanedCmts := make(map[string]struct{})
		for commentId, parentId := range allCmts {
			_, ok := allCmts[parentId]
			if !ok && parentId != "" {
				orphanedCmts[commentId] = struct{}{}
			}
		}
		// Get children of orphaned comments
		for cmtId, _ := range orphanedCmts {
			children := getChildrenOfOrphanedComment(cmtId, allCmts)
			maps.Copy(children, orphanedCmts)
		}
		deleteIds := []primitive.ObjectID{}
		for id, _ := range orphanedCmts {
			oid, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				log.Fatal("Invalid ID : ", err)
			}
			deleteIds = append(deleteIds, oid)
		}
		// Delete all orphaned comments and votes on those orphaned comments
		if _, err := colls.PostCommentsCollection.UpdateByID(context.Background(), postCmts.ID, bson.M{"$pull": bson.M{
			"comments": bson.M{"_id": bson.M{"$in": deleteIds}},
			"votes":    bson.M{"$elemMatch": bson.M{"parent_id": bson.M{"$in": deleteIds}}},
		}}); err != nil {
			log.Fatal("ERROR IN POSTS CLEANUP DELETE COMMENTS OPERATION :", err)
		}
	}
}
