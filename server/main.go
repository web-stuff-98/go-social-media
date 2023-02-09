package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/web-stuff-98/go-social-media/pkg/attachmentserver"
	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/db/changestreams"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"github.com/web-stuff-98/go-social-media/pkg/handlers"
	"github.com/web-stuff-98/go-social-media/pkg/handlers/middleware"
	rdb "github.com/web-stuff-98/go-social-media/pkg/redis"
	"github.com/web-stuff-98/go-social-media/pkg/socketserver"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

/*
	Router stuff should be moved into a seperate file...

	https://github.com/gorilla/mux#serving-single-page-applications

	Rate limiting stuff maybe should not be route specific
*/

type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path = filepath.Join(h.staticPath, path)
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func main() {
	protectedUids := make(map[primitive.ObjectID]struct{})
	protectedPids := make(map[primitive.ObjectID]struct{})
	protectedRids := make(map[primitive.ObjectID]struct{})

	if err := godotenv.Load(); err != nil {
		log.Fatal("DOTENV ERROR : ", err)
	}
	DB, Collections := db.Init()
	SocketServer, err := socketserver.Init(Collections)
	if err != nil {
		log.Fatal("Failed to set up socket server ", err)
	}
	AttachmentServer, err := attachmentserver.Init(Collections, SocketServer)
	if err != nil {
		log.Fatal("Failed to set up attachment server ", err)
	}

	h := handlers.New(DB, Collections, SocketServer, AttachmentServer, &handlers.ProtectedIDs{
		Uids: protectedUids,
		Pids: protectedPids,
		Rids: protectedRids,
	})

	router := mux.NewRouter()
	redisClient := rdb.Init()

	var origins []string
	if os.Getenv("PRODUCTION") == "true" {
		origins = []string{"https://go-social-media-js.herokuapp.com"}
	} else {
		origins = []string{"http://localhost:3000", "http://localhost:8080"}
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PATCH", "DELETE"},
		AllowCredentials: true,
	})

	api := router.PathPrefix("/api/").Subrouter()
	api.HandleFunc("/users/{id}", middleware.BasicRateLimiter(h.GetUser, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       500,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_user",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	api.HandleFunc("/users/{id}/pfp", middleware.BasicRateLimiter(h.GetPfp, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       500,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_pfp",
	}, *redisClient, *Collections)).Methods(http.MethodGet)

	api.HandleFunc("/account/register", middleware.BasicRateLimiter(h.Register, middleware.SimpleLimiterOpts{
		Window:        time.Second * 1000,
		MaxReqs:       3,
		BlockDuration: time.Second * 10000,
		Message:       "You have been creating too many accounts",
		RouteName:     "register",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/account/login", middleware.BasicRateLimiter(h.Login, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       5,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "login",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/account/logout", middleware.BasicRateLimiter(h.Logout, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       5,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "logout",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/account/refresh", middleware.BasicRateLimiter(h.RefreshToken, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       6,
		BlockDuration: time.Second * 6000,
		Message:       "Too many requests",
		RouteName:     "refresh_token",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/account/delete", middleware.BasicRateLimiter(h.DeleteAccount, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       2,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "delete_account",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/account/pfp", middleware.BasicRateLimiter(h.UploadPfp, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 500,
		Message:       "Too many requests",
		RouteName:     "upload_pfp",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/account/conversations", middleware.BasicRateLimiter(h.GetConversations, middleware.SimpleLimiterOpts{
		Window:        time.Second * 8,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_conversations",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	api.HandleFunc("/account/conversation/{id}", middleware.BasicRateLimiter(h.GetConversation, middleware.SimpleLimiterOpts{
		Window:        time.Second * 8,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_conversation",
	}, *redisClient, *Collections)).Methods(http.MethodGet)

	api.HandleFunc("/posts/newest", middleware.BasicRateLimiter(h.GetNewestPosts, middleware.SimpleLimiterOpts{
		Window:        time.Second * 10,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_new_posts",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	api.HandleFunc("/posts/page/{page}", middleware.BasicRateLimiter(h.GetPage, middleware.SimpleLimiterOpts{
		Window:        time.Second * 10,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_page",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	api.HandleFunc("/posts/{postId}/comment", middleware.BasicRateLimiter(h.CommentOnPost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       40,
		BlockDuration: time.Second * 3000,
		Message:       "You have been making too many comments",
		RouteName:     "create_comment",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/posts/{postId}/comment/{id}/delete", middleware.BasicRateLimiter(h.DeleteCommentOnPost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       40,
		BlockDuration: time.Second * 3000,
		Message:       "You have been deleting too many comments",
		RouteName:     "delete_comment",
	}, *redisClient, *Collections)).Methods(http.MethodDelete)
	api.HandleFunc("/posts/{postId}/comment/{id}/update", middleware.BasicRateLimiter(h.UpdatePostComment, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       30,
		BlockDuration: time.Second * 3000,
		Message:       "You have been editing too many comments",
		RouteName:     "update_comment",
	}, *redisClient, *Collections)).Methods(http.MethodPatch)
	api.HandleFunc("/posts/{slug}", middleware.BasicRateLimiter(h.GetPost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 60,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_post",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	api.HandleFunc("/posts/{slug}/delete", middleware.BasicRateLimiter(h.DeletePost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "delete_post",
	}, *redisClient, *Collections)).Methods(http.MethodDelete)
	api.HandleFunc("/posts/{slug}/update", middleware.BasicRateLimiter(h.UpdatePost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 3000,
		Message:       "You have been editing too many posts",
		RouteName:     "update_post",
	}, *redisClient, *Collections)).Methods(http.MethodPatch)
	api.HandleFunc("/posts", middleware.BasicRateLimiter(h.CreatePost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 3000,
		Message:       "You have been creating too many posts",
		RouteName:     "create_post",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/posts/{slug}/image", middleware.BasicRateLimiter(h.UploadPostImage, middleware.SimpleLimiterOpts{
		Window:        time.Second * 30,
		MaxReqs:       20,
		BlockDuration: time.Second * 200,
		Message:       "Too many requests",
		RouteName:     "upload_post_image",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/posts/{id}/image", middleware.BasicRateLimiter(h.GetPostImage, middleware.SimpleLimiterOpts{
		Window:        time.Second * 3,
		MaxReqs:       60,
		BlockDuration: time.Second * 100,
		Message:       "Too many requests",
		RouteName:     "get_post_image",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	api.HandleFunc("/posts/{id}/thumb", middleware.BasicRateLimiter(h.GetPostThumb, middleware.SimpleLimiterOpts{
		Window:        time.Second * 3,
		MaxReqs:       60,
		BlockDuration: time.Second * 100,
		Message:       "Too many requests",
		RouteName:     "get_post_thumb",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	api.HandleFunc("/posts/{id}/vote", middleware.BasicRateLimiter(h.VoteOnPost, middleware.SimpleLimiterOpts{
		Window:        time.Second * 2,
		MaxReqs:       10,
		BlockDuration: time.Second * 100,
		Message:       "Too many requests",
		RouteName:     "post_vote",
	}, *redisClient, *Collections)).Methods(http.MethodPatch)
	api.HandleFunc("/posts/{postId}/{commentId}/vote", middleware.BasicRateLimiter(h.VoteOnPostComment, middleware.SimpleLimiterOpts{
		Window:        time.Second * 2,
		MaxReqs:       10,
		BlockDuration: time.Second * 100,
		Message:       "Too many requests",
		RouteName:     "post_vote_comment",
	}, *redisClient, *Collections)).Methods(http.MethodPatch)

	api.HandleFunc("/rooms", middleware.BasicRateLimiter(h.CreateRoom, middleware.SimpleLimiterOpts{
		Window:        time.Second * 240,
		MaxReqs:       3,
		BlockDuration: time.Second * 1000,
		Message:       "You have been creating too many rooms",
		RouteName:     "create_room",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/rooms/page/{page}", middleware.BasicRateLimiter(h.GetRoomPage, middleware.SimpleLimiterOpts{
		Window:        time.Second * 10,
		MaxReqs:       20,
		BlockDuration: time.Second * 1000,
		Message:       "Too many requests",
		RouteName:     "get_room_page",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	api.HandleFunc("/rooms/own", middleware.BasicRateLimiter(h.GetOwnRooms, middleware.SimpleLimiterOpts{
		Window:        time.Second * 10,
		MaxReqs:       20,
		BlockDuration: time.Second * 1000,
		Message:       "Too many requests",
		RouteName:     "get_own_rooms",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	api.HandleFunc("/rooms/{id}", middleware.BasicRateLimiter(h.GetRoom, middleware.SimpleLimiterOpts{
		Window:        time.Second * 10,
		MaxReqs:       20,
		BlockDuration: time.Second * 1000,
		Message:       "Too many requests",
		RouteName:     "get_room",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	api.HandleFunc("/rooms/{id}/image", middleware.BasicRateLimiter(h.UploadRoomImage, middleware.SimpleLimiterOpts{
		Window:        time.Second * 30,
		MaxReqs:       20,
		BlockDuration: time.Second * 200,
		Message:       "Too many requests",
		RouteName:     "upload_room_image",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/rooms/{id}/image", middleware.BasicRateLimiter(h.GetRoomImage, middleware.SimpleLimiterOpts{
		Window:        time.Second * 3,
		MaxReqs:       60,
		BlockDuration: time.Second * 100,
		Message:       "Too many requests",
		RouteName:     "get_room_image",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	api.HandleFunc("/rooms/{id}/update", middleware.BasicRateLimiter(h.UpdateRoom, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 3000,
		Message:       "You have been editing too many rooms",
		RouteName:     "update_room",
	}, *redisClient, *Collections)).Methods(http.MethodPatch)
	api.HandleFunc("/rooms/{id}/invite", middleware.BasicRateLimiter(h.InviteToRoom, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 3000,
		Message:       "You have been sending too many invitations",
		RouteName:     "invite_room",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/rooms/{id}/ban", middleware.BasicRateLimiter(h.BanUserFromRoom, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "ban_room",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/rooms/{id}/unban", middleware.BasicRateLimiter(h.UnBanUserFromRoom, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "unban_room",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/rooms/{id}/invite/accept/{msgId}", middleware.BasicRateLimiter(h.AcceptRoomInvite, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "invite_room_accept",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/rooms/{id}/invite/decline/{msgId}", middleware.BasicRateLimiter(h.DeclineRoomInvite, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       10,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "invite_room_decline",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/rooms/{id}/private-data", middleware.BasicRateLimiter(h.GetRoomPrivateData, middleware.SimpleLimiterOpts{
		Window:        time.Second * 3,
		MaxReqs:       60,
		BlockDuration: time.Second * 100,
		Message:       "Too many requests",
		RouteName:     "get_room_private_data",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	api.HandleFunc("/rooms/{id}/delete", middleware.BasicRateLimiter(h.DeleteRoom, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "delete_room",
	}, *redisClient, *Collections)).Methods(http.MethodDelete)

	api.HandleFunc("/attachment/metadata/{msgId}/{recipientId}", middleware.BasicRateLimiter(h.HandleAttachmentMetadata, middleware.SimpleLimiterOpts{
		Window:        time.Second * 30,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "attachment_metadata",
	}, *redisClient, *Collections)).Methods(http.MethodPost)
	api.HandleFunc("/attachment/download/{id}", middleware.BasicRateLimiter(h.DownloadAttachment, middleware.SimpleLimiterOpts{
		Window:        time.Second * 120,
		MaxReqs:       4,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "download_attachment",
	}, *redisClient, *Collections)).Methods(http.MethodGet)
	/*api.HandleFunc("/attachment/video/{id}", middleware.BasicRateLimiter(h.GetVideoPartialContent, middleware.SimpleLimiterOpts{
		Window:        time.Second * 20,
		MaxReqs:       20,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "get_video_chunk",
	}, *redisClient, *Collections)).Methods(http.MethodGet)*/
	api.HandleFunc("/attachment/chunk/{msgId}", middleware.BasicRateLimiter(h.UploadAttachmentChunk, middleware.SimpleLimiterOpts{
		Window:        time.Second * 60,
		MaxReqs:       60,
		BlockDuration: time.Second * 3000,
		Message:       "Too many requests",
		RouteName:     "upload_chunk",
	}, *redisClient, *Collections)).Methods(http.MethodPost)

	api.HandleFunc("/ws", h.WebSocketEndpoint)

	spa := spaHandler{staticPath: "build", indexPath: "index.html"}
	router.PathPrefix("/").Handler(spa)

	log.Println("Watching changestreams...")
	changestreams.WatchCollections(DB, SocketServer, AttachmentServer)

	if os.Getenv("PRODUCTION") == "true" {
		//DB.Drop(context.Background())
		//go seed.SeedDB(Collections, 50, 300, 50, protectedUids, protectedPids, protectedRids)
		// Seeds already been generated, so just get everything already in the database instead
		pcursor, _ := Collections.PostCollection.Find(context.Background(), bson.M{})
		for pcursor.Next(context.Background()) {
			post := &models.Post{}
			pcursor.Decode(&post)
			protectedPids[post.ID] = struct{}{}
		}
		rcursor, _ := Collections.RoomCollection.Find(context.Background(), bson.M{})
		for rcursor.Next(context.Background()) {
			room := &models.Room{}
			rcursor.Decode(&room)
			protectedPids[room.ID] = struct{}{}
		}
		ucursor, _ := Collections.UserCollection.Find(context.Background(), bson.M{})
		for ucursor.Next(context.Background()) {
			user := &models.User{}
			ucursor.Decode(&user)
			protectedPids[user.ID] = struct{}{}
		}
	} else {
		//DB.Drop(context.Background())
		//go seed.SeedDB(Collections, 10, 10, 5, protectedUids, protectedPids, protectedRids)
	}

	deleteAccountTicker := time.NewTicker(20 * time.Minute)
	go func() {
		for {
			select {
			case <-deleteAccountTicker.C:
				h.Collections.UserCollection.DeleteMany(context.Background(), bson.M{
					"$and": []bson.M{
						{"created_at": bson.M{"$lt": primitive.NewDateTimeFromTime(time.Now().Add(-time.Minute * 20))}},
						{"_id": bson.M{"$nin": protectedUids}},
					},
				})
			}
		}
	}()

	log.Println("API open on port", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(fmt.Sprint(":", os.Getenv("PORT")), c.Handler(router)))
}
