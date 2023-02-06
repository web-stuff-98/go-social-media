package socketserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
	Uid can always be left as primitive.NilObjectID, users are not required
	to be authenticated to connect or open subscriptions, but there is an auth
	check for users down below, to make sure users cannot subscribe to other users
	inboxes/notifications or subscribe to rooms without being authenticated.
*/

/*--------------- SOCKET SERVER STRUCT ---------------*/
type SocketServer struct {
	Connections                 Connections
	Subscriptions               Subscriptions
	ConnectionSubscriptionCount ConnectionsSubscriptionCount

	RegisterConn   chan ConnectionInfo
	UnregisterConn chan ConnectionInfo

	RegisterSubscriptionConn   chan SubscriptionConnectionInfo
	UnregisterSubscriptionConn chan SubscriptionConnectionInfo

	SendDataToSubscription           chan SubscriptionDataMessage
	SendDataToSubscriptionExclusive  chan ExclusiveSubscriptionDataMessage
	SendDataToSubscriptions          chan SubscriptionDataMessageMulti
	SendDataToSubscriptionsExclusive chan ExclusiveSubscriptionDataMessageMulti
	RemoveUserFromSubscription       chan RemoveUserFromSubscription

	// websocket Write/Read must be done from 1 goroutine. Queue all of them to be executed in a loop.
	MessageSendQueue chan QueuedMessage

	VidChatOpenChan            chan VidChatOpenData
	VidChatCloseChan           chan *websocket.Conn
	VidChatStatus              VidChatStatus
	VidChatGetAllUsersInRoom   chan VidChatGetAllUsersInRoom
	VidChatGetOtherUserVidOpen chan VidChatGetOtherUserVidOpen

	// openConversations is a map including all the users that have the "inbox" section of their chat open,
	// with an inner map containing the UIDs of users they have conversation open with (with an empty struct)
	OpenConversations            OpenConversations
	GetUserConversationsOpenWith chan GetUserConversationsOpenWith
	UserOpenConversationWith     chan UserOpenCloseConversationWith
	UserCloseConversationWith    chan UserOpenCloseConversationWith

	SendDataToUser chan UserDataMessage

	DestroySubscription chan string
}

/*--------------- MUTEX PROTECTED MAPS ---------------*/
type Connections struct {
	data  map[*websocket.Conn]primitive.ObjectID
	mutex sync.Mutex
}
type Subscriptions struct {
	data  map[string]map[*websocket.Conn]primitive.ObjectID
	mutex sync.Mutex
}
type ConnectionsSubscriptionCount struct {
	data  map[*websocket.Conn]uint8 //Max subscriptions is 128... nice number half max uint8
	mutex sync.Mutex
}
type VidChatStatus struct {
	data  map[*websocket.Conn]VidChatOpenData
	mutex sync.Mutex
}
type OpenConversations struct {
	data  map[primitive.ObjectID]map[primitive.ObjectID]struct{}
	mutex sync.Mutex
}

/*--------------- MISC STRUCTS ---------------*/
type ConnectionInfo struct {
	Conn        *websocket.Conn
	Uid         primitive.ObjectID
	VidChatOpen bool
}
type SubscriptionConnectionInfo struct {
	Name string
	Uid  primitive.ObjectID
	Conn *websocket.Conn
}
type SubscriptionDataMessage struct {
	Name string
	Data []byte
}
type ExclusiveSubscriptionDataMessage struct {
	Name    string
	Data    []byte
	Exclude map[primitive.ObjectID]bool
}
type SubscriptionDataMessageMulti struct {
	Names []string
	Data  []byte
}
type ExclusiveSubscriptionDataMessageMulti struct {
	Names   []string
	Data    []byte
	Exclude map[primitive.ObjectID]bool
}
type QueuedMessage struct {
	Conn *websocket.Conn
	Data []byte
}
type UserDataMessage struct {
	Uid  primitive.ObjectID
	Data interface{}
	Type string
}
type VidChatOpenData struct {
	Conn *websocket.Conn
	Id   primitive.ObjectID
}
type RemoveUserFromSubscription struct {
	Name string
	Uid  primitive.ObjectID
}
type UserOpenCloseConversationWith struct {
	Uid     primitive.ObjectID
	ConvUid primitive.ObjectID
}

/*--------------- CHANNEL STRUCTS ---------------*/
type GetUserConversationsOpenWith struct {
	RecvChan chan<- bool
	Uid      primitive.ObjectID
	UidB     primitive.ObjectID
}
type VidChatGetAllUsersInRoom struct {
	RecvChan  chan<- []string // Hexes of other user ids
	Uid       primitive.ObjectID
	RoomIdHex string
}
type VidChatGetOtherUserVidOpen struct {
	RecvChan chan<- []string // Array will be empty or contain the other user Id hex
	Uid      primitive.ObjectID
	UidB     primitive.ObjectID
}

func Init(colls *db.Collections) (*SocketServer, error) {
	socketServer := &SocketServer{
		Connections: Connections{
			data: make(map[*websocket.Conn]primitive.ObjectID),
		},
		Subscriptions: Subscriptions{
			data: make(map[string]map[*websocket.Conn]primitive.ObjectID),
		},
		ConnectionSubscriptionCount: ConnectionsSubscriptionCount{
			data: make(map[*websocket.Conn]uint8),
		},

		RegisterConn:   make(chan ConnectionInfo),
		UnregisterConn: make(chan ConnectionInfo),

		RegisterSubscriptionConn:   make(chan SubscriptionConnectionInfo),
		UnregisterSubscriptionConn: make(chan SubscriptionConnectionInfo),

		SendDataToSubscription:           make(chan SubscriptionDataMessage),
		SendDataToSubscriptionExclusive:  make(chan ExclusiveSubscriptionDataMessage),
		SendDataToSubscriptions:          make(chan SubscriptionDataMessageMulti),
		SendDataToSubscriptionsExclusive: make(chan ExclusiveSubscriptionDataMessageMulti),
		RemoveUserFromSubscription:       make(chan RemoveUserFromSubscription),

		MessageSendQueue: make(chan QueuedMessage),

		VidChatOpenChan:  make(chan VidChatOpenData),
		VidChatCloseChan: make(chan *websocket.Conn),
		VidChatStatus: VidChatStatus{
			data: make(map[*websocket.Conn]VidChatOpenData),
		},
		VidChatGetAllUsersInRoom:   make(chan VidChatGetAllUsersInRoom),
		VidChatGetOtherUserVidOpen: make(chan VidChatGetOtherUserVidOpen),

		OpenConversations: OpenConversations{
			data: make(map[primitive.ObjectID]map[primitive.ObjectID]struct{}),
		},
		GetUserConversationsOpenWith: make(chan GetUserConversationsOpenWith),
		UserOpenConversationWith:     make(chan UserOpenCloseConversationWith),
		UserCloseConversationWith:    make(chan UserOpenCloseConversationWith),

		SendDataToUser: make(chan UserDataMessage),

		DestroySubscription: make(chan string),
	}
	RunServer(socketServer, colls)
	return socketServer, nil
}

func RunServer(socketServer *SocketServer, colls *db.Collections) {
	/* ----- Connection registration ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in WS registration :", r)
				}
			}()
			connData := <-socketServer.RegisterConn
			if connData.Conn != nil {
				socketServer.Connections.mutex.Lock()
				socketServer.Connections.data[connData.Conn] = connData.Uid
				socketServer.Connections.mutex.Unlock()
			}
		}
	}()
	/* ----- Disconnect registration ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in WS deregistration :", r)
				}
			}()
			connData := <-socketServer.UnregisterConn
			socketServer.Connections.mutex.Lock()
			socketServer.Subscriptions.mutex.Lock()
			socketServer.VidChatStatus.mutex.Lock()
			socketServer.ConnectionSubscriptionCount.mutex.Lock()
			socketServer.OpenConversations.mutex.Lock()
			for conn := range socketServer.Connections.data {
				if conn == connData.Conn {
					delete(socketServer.Connections.data, conn)
					delete(socketServer.VidChatStatus.data, conn)
					delete(socketServer.ConnectionSubscriptionCount.data, conn)
					if connData.Uid != primitive.NilObjectID {
						delete(socketServer.OpenConversations.data, connData.Uid)
					}
					for _, r := range socketServer.Subscriptions.data {
						for c := range r {
							if c == connData.Conn {
								delete(r, c)
								break
							}
						}
					}
					break
				}
			}
			socketServer.Connections.mutex.Unlock()
			socketServer.Subscriptions.mutex.Unlock()
			socketServer.VidChatStatus.mutex.Unlock()
			socketServer.ConnectionSubscriptionCount.mutex.Unlock()
			socketServer.OpenConversations.mutex.Unlock()
		}
	}()
	/* ----- Send messages in queue ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in queued socket messages :", r)
				}
			}()
			data := <-socketServer.MessageSendQueue
			data.Conn.WriteMessage(websocket.TextMessage, data.Data)
		}
	}()
	/* ----- Subscription connection registration (also check the authorization if subscription requires it) ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in subscription registration :", r)
				}
			}()
			connData := <-socketServer.RegisterSubscriptionConn
			if connData.Conn != nil {
				allow := true
				// Make sure users cannot subscribe to other users inboxes
				if strings.Contains(connData.Name, "inbox=") {
					rawUid := strings.ReplaceAll(connData.Name, "inbox=", "")
					uid, err := primitive.ObjectIDFromHex(rawUid)
					if err != nil {
						allow = false
					}
					if uid != connData.Uid {
						allow = false
					}
				}
				// Make sure users cannot subscribe to other users notifications
				if strings.Contains(connData.Name, "notifications=") {
					rawUid := strings.ReplaceAll(connData.Name, "notifications=", "")
					uid, err := primitive.ObjectIDFromHex(rawUid)
					if err != nil {
						allow = false
					}
					if uid != connData.Uid {
						allow = false
					}
				}
				// Make sure users cannot subscribe to rooms if they aren't logged in, banned, or not a member (if rooms private)
				if strings.Contains(connData.Name, "room=") {
					if connData.Uid == primitive.NilObjectID {
						allow = false
					}
					rawRoomId := strings.ReplaceAll(connData.Name, "room=", "")
					roomId, err := primitive.ObjectIDFromHex(rawRoomId)
					if err != nil {
						allow = false
					} else {
						var room models.Room
						if err := colls.RoomCollection.FindOne(context.Background(), bson.M{"_id": roomId}).Decode(&room); err != nil {
							allow = false
							return
						} else {
							var roomPrivateData models.RoomPrivateData
							if err := colls.RoomPrivateDataCollection.FindOne(context.Background(), bson.M{"_id": roomId}).Decode(&roomPrivateData); err != nil {
								allow = false
								return
							}
							for _, oi := range roomPrivateData.Banned {
								if oi == connData.Uid {
									allow = false
									break
								}
							}
							if room.Private == true {
								isMember := false
								for _, oi := range roomPrivateData.Members {
									if oi == connData.Uid {
										allow = true
										break
									}
								}
								if connData.Uid != room.Author && !isMember {
									allow = false
									return
								}
							}
						}
					}
				}
				// Make sure users cannot subscribe to room private data if not a member or the author
				if strings.Contains(connData.Name, "room_private_data=") {
					if connData.Uid == primitive.NilObjectID {
						allow = false
					} else {
						id, err := primitive.ObjectIDFromHex(strings.ReplaceAll(connData.Name, "room_private_data=", ""))
						if err == nil {
							room := &models.Room{}
							if err := colls.RoomCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&room); err != nil {
								allow = false
							} else {
								if room.Author != connData.Uid {
									allow = false
								}
								roomPrivateData := &models.RoomPrivateData{}
								foundInMembers := false
								if err := colls.RoomPrivateDataCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&roomPrivateData); err != nil {
									allow = false
								} else {
									for _, oi := range roomPrivateData.Members {
										if oi == connData.Uid {
											foundInMembers = true
											break
										}
									}
								}
								if foundInMembers || room.Author == connData.Uid {
									allow = true
								}
							}
						} else {
							allow = false
						}
					}
				}
				// Make sure users cannot open too many subscriptions
				socketServer.ConnectionSubscriptionCount.mutex.Lock()
				socketServer.Subscriptions.mutex.Lock()
				count, countOk := socketServer.ConnectionSubscriptionCount.data[connData.Conn]
				if count >= 128 {
					allow = false
				}
				// Passed all checks, add the connection to the subscription
				if allow {
					if socketServer.Subscriptions.data[connData.Name] == nil {
						socketServer.Subscriptions.data[connData.Name] = make(map[*websocket.Conn]primitive.ObjectID)
					}
					socketServer.Subscriptions.data[connData.Name][connData.Conn] = connData.Uid
					if countOk {
						socketServer.ConnectionSubscriptionCount.data[connData.Conn]++
					} else {
						socketServer.ConnectionSubscriptionCount.data[connData.Conn] = 1
					}
				}
				socketServer.Subscriptions.mutex.Unlock()
				socketServer.ConnectionSubscriptionCount.mutex.Unlock()
			}
		}
	}()
	/* ----- Subscription disconnect registration ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in subscription disconnect registration :", r)
				}
			}()
			connData := <-socketServer.UnregisterSubscriptionConn
			var err error
			if connData.Conn == nil {
				err = fmt.Errorf("Connection was nil")
			}
			if err != nil {
				socketServer.Subscriptions.mutex.Lock()
				socketServer.VidChatStatus.mutex.Lock()
				if _, ok := socketServer.Subscriptions.data[connData.Name]; ok {
					delete(socketServer.Subscriptions.data[connData.Name], connData.Conn)
				}
				delete(socketServer.VidChatStatus.data, connData.Conn)
				socketServer.Subscriptions.mutex.Unlock()
				socketServer.VidChatStatus.mutex.Unlock()
				socketServer.ConnectionSubscriptionCount.mutex.Lock()
				if _, ok := socketServer.ConnectionSubscriptionCount.data[connData.Conn]; ok {
					socketServer.ConnectionSubscriptionCount.data[connData.Conn]--
				}
				socketServer.ConnectionSubscriptionCount.mutex.Unlock()
			}
		}
	}()
	/* ----- Send data to subscription ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in subscription data channel :", r)
				}
			}()
			subsData := <-socketServer.SendDataToSubscription
			socketServer.Subscriptions.mutex.Lock()
			for k, s := range socketServer.Subscriptions.data {
				if k == subsData.Name {
					for conn := range s {
						socketServer.MessageSendQueue <- QueuedMessage{
							Conn: conn,
							Data: subsData.Data,
						}
					}
					break
				}
			}
			socketServer.Subscriptions.mutex.Unlock()
		}
	}()
	/* ----- Send data to subscription excluding uids ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in exclusive subscription data channel :", r)
				}
			}()
			subsData := <-socketServer.SendDataToSubscriptionExclusive
			socketServer.Subscriptions.mutex.Lock()
			for k, s := range socketServer.Subscriptions.data {
				if k == subsData.Name {
					for conn, oid := range s {
						if subsData.Exclude[oid] != true {
							socketServer.MessageSendQueue <- QueuedMessage{
								Conn: conn,
								Data: subsData.Data,
							}
						}
					}
					break
				}
			}
			socketServer.Subscriptions.mutex.Unlock()
		}
	}()
	/* ----- Send data to multiple subscriptions ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in subscription data channel :", r)
				}
			}()
			subsData := <-socketServer.SendDataToSubscriptions
			socketServer.Subscriptions.mutex.Lock()
			for _, v := range subsData.Names {
				for k, s := range socketServer.Subscriptions.data {
					if k == v {
						for conn := range s {
							socketServer.MessageSendQueue <- QueuedMessage{
								Conn: conn,
								Data: subsData.Data,
							}
						}
						break
					}
				}
			}
			socketServer.Subscriptions.mutex.Unlock()
		}
	}()
	/* ----- Send data to multiple subscriptions excluding uids ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in exclusive subscription data channel :", r)
				}
			}()
			subsData := <-socketServer.SendDataToSubscriptionsExclusive
			socketServer.Subscriptions.mutex.Lock()
			for _, v := range subsData.Names {
				for k, s := range socketServer.Subscriptions.data {
					if k == v {
						for conn, oid := range s {
							if subsData.Exclude[oid] != true {
								socketServer.MessageSendQueue <- QueuedMessage{
									Conn: conn,
									Data: subsData.Data,
								}
							}
						}
						break
					}
				}
			}
			socketServer.Subscriptions.mutex.Unlock()
		}
	}()
	/* ----- Send data to a specific user ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in send data to user channel :", r)
				}
			}()
			data := <-socketServer.SendDataToUser
			socketServer.Connections.mutex.Lock()
			for conn, uid := range socketServer.Connections.data {
				if data.Uid == uid {
					var m map[string]interface{}
					outBytesNoTypeKey, err := json.Marshal(data.Data)
					json.Unmarshal(outBytesNoTypeKey, &m)
					m["TYPE"] = data.Type
					outBytes, err := json.Marshal(m)
					if err == nil {
						socketServer.MessageSendQueue <- QueuedMessage{
							Conn: conn,
							Data: outBytes,
						}
					} else {
						log.Println("Error marshaling data to be sent to user :", err)
					}
					break
				}
			}
			socketServer.Connections.mutex.Unlock()
		}
	}()
	/* ----- Remove a user from subscription ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in remove user from subscription channel :", r)
				}
			}()
			data := <-socketServer.RemoveUserFromSubscription
			socketServer.Subscriptions.mutex.Lock()
			if subs, ok := socketServer.Subscriptions.data[data.Name]; ok {
				for c, oi := range subs {
					if oi == data.Uid {
						defer func() {
							socketServer.Subscriptions.mutex.Unlock()
						}()
						socketServer.Subscriptions.mutex.Lock()
						delete(socketServer.Subscriptions.data[data.Name], c)
						break
					}
				}
			}
			socketServer.Subscriptions.mutex.Unlock()
		}
	}()
	/* ----- Destroy subscription ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in destroy subscription channel :", r)
				}
			}()
			subsName := <-socketServer.DestroySubscription
			socketServer.Subscriptions.mutex.Lock()
			socketServer.ConnectionSubscriptionCount.mutex.Lock()
			for c := range socketServer.Subscriptions.data[subsName] {
				if _, ok := socketServer.ConnectionSubscriptionCount.data[c]; ok {
					socketServer.ConnectionSubscriptionCount.data[c]--
				}
			}
			delete(socketServer.Subscriptions.data, subsName)
			socketServer.Subscriptions.mutex.Unlock()
			socketServer.ConnectionSubscriptionCount.mutex.Unlock()
		}
	}()
	/* ----- Vid chat opened chan ----- */
	go func() {
		for {
			data := <-socketServer.VidChatOpenChan
			socketServer.VidChatStatus.mutex.Lock()
			socketServer.VidChatStatus.data[data.Conn] = data
			socketServer.VidChatStatus.mutex.Unlock()
		}
	}()
	/* ----- Vid chat closed chan ----- */
	go func() {
		for {
			data := <-socketServer.VidChatCloseChan
			socketServer.VidChatStatus.mutex.Lock()
			delete(socketServer.VidChatStatus.data, data)
			socketServer.VidChatStatus.mutex.Unlock()
		}
	}()
	/* ----- Get user has conversations open with other user chan ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in get user has conversations open with other user chan :", r)
				}
			}()
			data := <-socketServer.GetUserConversationsOpenWith
			socketServer.OpenConversations.mutex.Lock()
			if openConvs, ok := socketServer.OpenConversations.data[data.Uid]; ok {
				for oi := range openConvs {
					if oi == data.UidB {
						data.RecvChan <- true
						break
					}
				}
			}
			socketServer.OpenConversations.mutex.Unlock()
			data.RecvChan <- false
		}
	}()
	/* ----- Open conversation with other user chan ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in open conversation with other user chan :", r)
				}
			}()
			data := <-socketServer.UserOpenConversationWith
			socketServer.OpenConversations.mutex.Lock()
			if _, ok := socketServer.OpenConversations.data[data.Uid]; ok {
				socketServer.OpenConversations.data[data.Uid][data.ConvUid] = struct{}{}
			} else {
				convs := make(map[primitive.ObjectID]struct{})
				convs[data.ConvUid] = struct{}{}
				socketServer.OpenConversations.data[data.Uid] = convs
			}
			socketServer.OpenConversations.mutex.Unlock()
		}
	}()
	/* ----- Close conversation with other user chan ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in close conversation with other user chan :", r)
				}
			}()
			data := <-socketServer.UserCloseConversationWith
			socketServer.OpenConversations.mutex.Lock()
			if _, ok := socketServer.OpenConversations.data[data.Uid]; ok {
				delete(socketServer.OpenConversations.data[data.Uid], data.ConvUid)
			}
			socketServer.OpenConversations.mutex.Unlock()
		}
	}()
	/* ----- Get UID hexes of other users in a room (vidChat) chan ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in room vidChat get all users chan :", r)
				}
			}()
			data := <-socketServer.VidChatGetAllUsersInRoom
			allUsers := []string{}
			socketServer.Subscriptions.mutex.Lock()
			socketServer.VidChatStatus.mutex.Lock()
			socketServer.Connections.mutex.Lock()
			for k, v := range socketServer.Subscriptions.data {
				if strings.ReplaceAll(k, "room=", "") == data.RoomIdHex {
					for _, oi := range v {
						if oi != data.Uid {
							for c, oi2 := range socketServer.Connections.data {
								if oi2 == oi {
									if status, ok := socketServer.VidChatStatus.data[c]; ok {
										if status.Id.Hex() == data.RoomIdHex {
											allUsers = append(allUsers, oi.Hex())
										}
									}
									break
								}
							}
						}
					}
					break
				}
			}
			socketServer.Subscriptions.mutex.Unlock()
			socketServer.VidChatStatus.mutex.Unlock()
			socketServer.Connections.mutex.Unlock()
			data.RecvChan <- allUsers
		}
	}()
	/* ----- Get UID hex of other user in a conversation (vidChat) chan ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in room vidChat get all users (conversation) chan :", r)
				}
			}()
			data := <-socketServer.VidChatGetOtherUserVidOpen
			allUsers := []string{}
			hasOpen := false
			socketServer.Connections.mutex.Lock()
			socketServer.VidChatStatus.mutex.Lock()
			for c, oi := range socketServer.Connections.data {
				if oi == data.UidB {
					if status, ok := socketServer.VidChatStatus.data[c]; ok {
						if status.Id == data.Uid {
							hasOpen = true
						}
					}
					break
				}
			}
			socketServer.Connections.mutex.Unlock()
			socketServer.VidChatStatus.mutex.Unlock()
			if hasOpen {
				allUsers = []string{data.UidB.Hex()}
			}
			data.RecvChan <- allUsers
		}
	}()

	/* -------- Cleanup ticker -------- */
	cleanupTicker := time.NewTicker(20 * time.Minute)
	quitCleanup := make(chan struct{})
	defer func() {
		quitCleanup <- struct{}{}
	}()
	go func() {
		for {
			select {
			case <-cleanupTicker.C:
				// Destroy empty subscriptions
				socketServer.Subscriptions.mutex.Lock()
				for k, v := range socketServer.Subscriptions.data {
					if len(v) == 0 {
						socketServer.DestroySubscription <- k
					}
				}
				socketServer.Subscriptions.mutex.Unlock()
			case <-quitCleanup:
				cleanupTicker.Stop()
				return
			}
		}
	}()
}
