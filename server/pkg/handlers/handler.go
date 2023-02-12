package handlers

/* Dependency injection for handlers */

import (
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis/v9"
	"github.com/web-stuff-98/go-social-media/pkg/attachmentserver"
	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func responseMessage(w http.ResponseWriter, c int, m string) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(c)
	json.NewEncoder(w).Encode(map[string]string{"msg": m})
	return
}

type ProtectedIDs struct {
	Uids map[primitive.ObjectID]struct{}
	Rids map[primitive.ObjectID]struct{}
	Pids map[primitive.ObjectID]struct{}
}

type handler struct {
	DB               *mongo.Database
	RedisClient      *redis.Client
	Collections      *db.Collections
	SocketServer     *socketserver.SocketServer
	AttachmentServer *attachmentserver.AttachmentServer
	ProtectedIDs     *ProtectedIDs
}

func New(db *mongo.Database, rdb *redis.Client, collections *db.Collections, sserver *socketserver.SocketServer, aserver *attachmentserver.AttachmentServer, protectedIDs *ProtectedIDs) handler {
	return handler{db, rdb, collections, sserver, aserver, protectedIDs}
}
