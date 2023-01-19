package handlers

/* Dependency injection for handlers */

import (
	"encoding/json"
	"net/http"

	"github.com/web-stuff-98/go-social-media/pkg/attachmentserver"
	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"

	"go.mongodb.org/mongo-driver/mongo"
)

func responseMessage(w http.ResponseWriter, c int, m string) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(c)
	json.NewEncoder(w).Encode(map[string]string{"msg": m})
	return
}

type handler struct {
	DB               *mongo.Database
	Collections      *db.Collections
	SocketServer     *socketserver.SocketServer
	AttachmentServer *attachmentserver.AttachmentServer
}

func New(db *mongo.Database, collections *db.Collections, sserver *socketserver.SocketServer, aserver *attachmentserver.AttachmentServer) handler {
	return handler{db, collections, sserver, aserver}
}
