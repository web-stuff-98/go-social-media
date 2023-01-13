package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/db/changestreams"
	"github.com/web-stuff-98/go-social-media/pkg/handlers"
	"github.com/web-stuff-98/go-social-media/pkg/seed"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("DOTENV ERROR : ", err)
	}
	DB, Collections := db.Init()
	SocketServer, err := socketserver.Init()
	if err != nil {
		log.Fatal("Failed to set up socket server ", err)
	}
	h := handlers.New(DB, Collections, SocketServer)
	router := mux.NewRouter()

	Collections.PostCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"title": "text",
		},
		Options: options.Index().SetName("title_text"),
	})
	Collections.RoomCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.M{
			"name": "text",
		},
		Options: options.Index().SetName("name_text"),
	})

	var origin string
	if os.Getenv("PRODUCTION") == "true" {
		origin = "https://go-social-media-js.herokuapp.com"
	} else {
		origin = "http://localhost:3000"
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{origin},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PATCH", "DELETE"},
	})

	router.HandleFunc("/api/users/{id}", h.GetUser).Methods(http.MethodGet)
	router.HandleFunc("/api/users/{id}/pfp", h.GetPfp).Methods(http.MethodGet)

	router.HandleFunc("/api/account/register", h.Register).Methods(http.MethodPost)
	router.HandleFunc("/api/account/login", h.Login).Methods(http.MethodPost)
	router.HandleFunc("/api/account/logout", h.Logout).Methods(http.MethodPost)
	router.HandleFunc("/api/account/refresh", h.RefreshToken).Methods(http.MethodPost)
	router.HandleFunc("/api/account/delete", h.DeleteAccount).Methods(http.MethodPost)
	router.HandleFunc("/api/account/pfp", h.UploadPfp).Methods(http.MethodPost)
	router.HandleFunc("/api/account/conversations", h.GetConversations).Methods(http.MethodGet)
	router.HandleFunc("/api/account/conversation/{id}", h.GetConversation).Methods(http.MethodGet)

	router.HandleFunc("/api/posts/page/{page}", h.GetPage).Methods(http.MethodGet)
	router.HandleFunc("/api/posts/{postId}/comment", h.CommentOnPost).Methods(http.MethodPost)
	router.HandleFunc("/api/posts/{postId}/comment/{id}/delete", h.DeleteCommentOnPost).Methods(http.MethodDelete)
	router.HandleFunc("/api/posts/{postId}/comment/{id}/update", h.UpdatePostComment).Methods(http.MethodPatch)
	router.HandleFunc("/api/posts/{slug}", h.GetPost).Methods(http.MethodGet)
	router.HandleFunc("/api/posts/{slug}/delete", h.DeletePost).Methods(http.MethodDelete)
	router.HandleFunc("/api/posts/{slug}/update", h.UpdatePost).Methods(http.MethodPatch)
	router.HandleFunc("/api/posts", h.CreatePost).Methods(http.MethodPost)
	router.HandleFunc("/api/posts/{slug}/image", h.UploadPostImage).Methods(http.MethodPost)
	router.HandleFunc("/api/posts/{id}/image", h.GetPostImage).Methods(http.MethodGet)
	router.HandleFunc("/api/posts/{id}/thumb", h.GetPostThumb).Methods(http.MethodGet)
	router.HandleFunc("/api/posts/{id}/vote", h.VoteOnPost).Methods(http.MethodPatch)

	router.HandleFunc("/api/rooms", h.CreateRoom).Methods(http.MethodPost)
	router.HandleFunc("/api/rooms/page/{page}", h.GetRoomPage).Methods(http.MethodGet)
	router.HandleFunc("/api/rooms/{id}", h.GetRoom).Methods(http.MethodGet)
	router.HandleFunc("/api/rooms/{id}/image", h.UploadRoomImage).Methods(http.MethodPost)
	router.HandleFunc("/api/rooms/{id}/image", h.GetRoomImage).Methods(http.MethodGet)
	router.HandleFunc("/api/rooms/{id}/update", h.UpdateRoom).Methods(http.MethodPatch)
	router.HandleFunc("/api/rooms/{id}/delete", h.DeleteRoom).Methods(http.MethodDelete)

	router.HandleFunc("/api/ws", h.WebSocketEndpoint)

	log.Println("Creating changestreams")
	changestreams.WatchCollections(DB, SocketServer)

	DB.Drop(context.TODO())
	go seed.SeedDB(&Collections, 20, 50)

	log.Println("API open on port", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(fmt.Sprint(":", os.Getenv("PORT")), c.Handler(router)))
}
