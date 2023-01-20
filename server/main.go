package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/web-stuff-98/go-social-media/pkg/attachmentserver"
	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/db/changestreams"
	"github.com/web-stuff-98/go-social-media/pkg/handlers"
	"github.com/web-stuff-98/go-social-media/pkg/handlers/middleware"
	rdb "github.com/web-stuff-98/go-social-media/pkg/redis"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

/*
	Router stuff should be moved into a seperate file...
*/

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("DOTENV ERROR : ", err)
	}
	DB, Collections := db.Init()
	SocketServer, err := socketserver.Init()
	if err != nil {
		log.Fatal("Failed to set up socket server ", err)
	}
	AttachmentServer, err := attachmentserver.Init(Collections, SocketServer)
	if err != nil {
		log.Fatal("Failed to set up attachment server ", err)
	}

	h := handlers.New(DB, Collections, SocketServer, AttachmentServer)

	router := mux.NewRouter()
	redisClient := rdb.Init()

	var origin string
	if os.Getenv("PRODUCTION") == "true" {
		origin = "https://go-social-media-js.herokuapp.com"
	} else {
		origin = "http://localhost:3000"
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{origin},
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PATCH", "DELETE"},
		AllowCredentials: true,
	})

	router.HandleFunc("/api/users/{id}", middleware.BasicRateLimiter(h.GetUser, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       500,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_user",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	router.HandleFunc("/api/users/{id}/pfp", middleware.BasicRateLimiter(h.GetPfp, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       500,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_pfp",
	}, *redisClient, *Collections)).Methods(http.MethodGet)

	router.HandleFunc("/api/account/register", middleware.BasicRateLimiter(h.Register, middleware.SimpleLimiterOpts{
		Window:        time.Second * 1000,
		MaxReqs:       3,
		BlockDuration: time.Second * 10000,
		Message:       "You have been creating too many accounts",
		RouteName:     "register",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	router.HandleFunc("/api/account/login", middleware.BasicRateLimiter(h.Login, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       5,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "login",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	router.HandleFunc("/api/account/logout", middleware.BasicRateLimiter(h.Logout, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       5,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "logout",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	router.HandleFunc("/api/account/refresh", middleware.BasicRateLimiter(h.RefreshToken, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       6,
		BlockDuration: time.Second * 6000,
		Message:       "Too many requests",
		RouteName:     "refresh_token",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	router.HandleFunc("/api/account/delete", middleware.BasicRateLimiter(h.DeleteAccount, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       2,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "delete_account",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	router.HandleFunc("/api/account/pfp", middleware.BasicRateLimiter(h.UploadPfp, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 500,
		Message:       "Too many requests",
		RouteName:     "upload_pfp",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	router.HandleFunc("/api/account/conversations", middleware.BasicRateLimiter(h.GetConversations, middleware.SimpleLimiterOpts{
		Window:        time.Second * 8,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_conversations",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	router.HandleFunc("/api/account/conversation/{id}", middleware.BasicRateLimiter(h.GetConversation, middleware.SimpleLimiterOpts{
		Window:        time.Second * 8,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_conversation",
	}, *redisClient, *Collections)).Methods(http.MethodGet)

	router.HandleFunc("/api/posts/page/{page}", middleware.BasicRateLimiter(h.GetPage, middleware.SimpleLimiterOpts{
		Window:        time.Second * 10,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_page",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	router.HandleFunc("/api/posts/{postId}/comment", middleware.BasicRateLimiter(h.CommentOnPost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       40,
		BlockDuration: time.Second * 3000,
		Message:       "You have been making too many comments",
		RouteName:     "create_comment",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	router.HandleFunc("/api/posts/{postId}/comment/{id}/delete", middleware.BasicRateLimiter(h.DeleteCommentOnPost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       40,
		BlockDuration: time.Second * 3000,
		Message:       "You have been deleting too many comments",
		RouteName:     "delete_comment",
	}, *redisClient, *Collections)).Methods(http.MethodDelete)
	router.HandleFunc("/api/posts/{postId}/comment/{id}/update", middleware.BasicRateLimiter(h.UpdatePostComment, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       30,
		BlockDuration: time.Second * 3000,
		Message:       "You have been editing too many comments",
		RouteName:     "update_comment",
	}, *redisClient, *Collections)).Methods(http.MethodPatch)
	router.HandleFunc("/api/posts/{slug}", middleware.BasicRateLimiter(h.GetPost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 60,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_post",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	router.HandleFunc("/api/posts/{slug}/delete", middleware.BasicRateLimiter(h.DeletePost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "delete_post",
	}, *redisClient, *Collections)).Methods(http.MethodDelete)
	router.HandleFunc("/api/posts/{slug}/update", middleware.BasicRateLimiter(h.UpdatePost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 3000,
		Message:       "You have been editing too many posts",
		RouteName:     "update_post",
	}, *redisClient, *Collections)).Methods(http.MethodPatch)
	router.HandleFunc("/api/posts", middleware.BasicRateLimiter(h.CreatePost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 3000,
		Message:       "You have been creating too many posts",
		RouteName:     "create_post",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	router.HandleFunc("/api/posts/{slug}/image", middleware.BasicRateLimiter(h.UploadPostImage, middleware.SimpleLimiterOpts{
		Window:        time.Second * 30,
		MaxReqs:       20,
		BlockDuration: time.Second * 200,
		Message:       "Too many requests",
		RouteName:     "upload_post_image",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	router.HandleFunc("/api/posts/{id}/image", middleware.BasicRateLimiter(h.GetPostImage, middleware.SimpleLimiterOpts{
		Window:        time.Second * 3,
		MaxReqs:       60,
		BlockDuration: time.Second * 100,
		Message:       "Too many requests",
		RouteName:     "get_post_image",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	router.HandleFunc("/api/posts/{id}/thumb", middleware.BasicRateLimiter(h.GetPostThumb, middleware.SimpleLimiterOpts{
		Window:        time.Second * 3,
		MaxReqs:       60,
		BlockDuration: time.Second * 100,
		Message:       "Too many requests",
		RouteName:     "get_post_thumb",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	router.HandleFunc("/api/posts/{id}/vote", middleware.BasicRateLimiter(h.VoteOnPost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 3,
		MaxReqs:       10,
		BlockDuration: time.Second * 100,
		Message:       "Too many requests",
		RouteName:     "post_vote",
	}, *redisClient, *Collections)).Methods(http.MethodPatch)
	router.HandleFunc("/api/posts/{postId}/{commentId}/vote", middleware.BasicRateLimiter(h.VoteOnPostComment, middleware.SimpleLimiterOpts{
		Window:        time.Second * 3,
		MaxReqs:       10,
		BlockDuration: time.Second * 100,
		Message:       "Too many requests",
		RouteName:     "post_vote_comment",
	}, *redisClient, *Collections)).Methods(http.MethodPatch)

	router.HandleFunc("/api/rooms", middleware.BasicRateLimiter(h.CreateRoom, middleware.SimpleLimiterOpts{
		Window:        time.Second * 240,
		MaxReqs:       3,
		BlockDuration: time.Second * 1000,
		Message:       "You have been creating too many rooms",
		RouteName:     "create_room",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	router.HandleFunc("/api/rooms/page/{page}", middleware.BasicRateLimiter(h.GetRoomPage, middleware.SimpleLimiterOpts{
		Window:        time.Second * 10,
		MaxReqs:       20,
		BlockDuration: time.Second * 1000,
		Message:       "Too many requests",
		RouteName:     "get_room_page",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	router.HandleFunc("/api/rooms/{id}", middleware.BasicRateLimiter(h.GetRoom, middleware.SimpleLimiterOpts{
		Window:        time.Second * 10,
		MaxReqs:       20,
		BlockDuration: time.Second * 1000,
		Message:       "Too many requests",
		RouteName:     "get_room",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	router.HandleFunc("/api/rooms/{id}/image", middleware.BasicRateLimiter(h.UploadRoomImage, middleware.SimpleLimiterOpts{
		Window:        time.Second * 30,
		MaxReqs:       20,
		BlockDuration: time.Second * 200,
		Message:       "Too many requests",
		RouteName:     "upload_room_image",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	router.HandleFunc("/api/rooms/{id}/image", middleware.BasicRateLimiter(h.GetRoomImage, middleware.SimpleLimiterOpts{
		Window:        time.Second * 3,
		MaxReqs:       60,
		BlockDuration: time.Second * 100,
		Message:       "Too many requests",
		RouteName:     "get_room_image",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	router.HandleFunc("/api/rooms/{id}/update", middleware.BasicRateLimiter(h.UpdateRoom, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 3000,
		Message:       "You have been editing too many rooms",
		RouteName:     "update_room",
	}, *redisClient, *Collections)).Methods(http.MethodPatch)
	router.HandleFunc("/api/rooms/{id}/delete", middleware.BasicRateLimiter(h.DeleteRoom, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "delete_room",
	}, *redisClient, *Collections)).Methods(http.MethodDelete)

	router.HandleFunc("/api/attachment/metadata/{msgId}/{recipientId}", middleware.BasicRateLimiter(h.HandleAttachmentMetadata, middleware.SimpleLimiterOpts{
		Window:        time.Second * 30,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "attachment_metadata",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	router.HandleFunc("/api/attachment/download/{id}", middleware.BasicRateLimiter(h.DownloadAttachment, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       4,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "download_attachment",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	/*
		BROKEN
		router.HandleFunc("/api/attachment/video/{id}", middleware.BasicRateLimiter(h.GetVideoPartialContent, middleware.SimpleLimiterOpts{
			Window:        time.Second * 20,
			MaxReqs:       20,
			BlockDuration: time.Second * 3000,
			Message:       "Too many requests",
			RouteName:     "get_video_chunk",
		}, *redisClient, *Collections)).Methods(http.MethodGet)*/
	router.HandleFunc("/api/attachment/chunk/{msgId}", middleware.BasicRateLimiter(h.UploadAttachmentChunk, middleware.SimpleLimiterOpts{
		Window:        time.Second * 60,
		MaxReqs:       60,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "upload_chunk",
	}, *redisClient, *Collections)).Methods(http.MethodPost)

	router.HandleFunc("/api/ws", h.WebSocketEndpoint)

	log.Println("Creating changestreams")
	changestreams.WatchCollections(DB, SocketServer, AttachmentServer)

	//DB.Drop(context.TODO())
	//go seed.SeedDB(Collections, 5, 2, 0)

	log.Println("API open on port", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(fmt.Sprint(":", os.Getenv("PORT")), c.Handler(router)))
}
