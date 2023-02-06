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


	All messages are put through the queue channel before being sent, to avoid
	the concurrent write to websocket error which I did not anticipate. Write to
	websocket can only be done from one goroutine.

	Uid can always be left as primitive.NilObjectID, users are not required
	to be authenticated to connect or open subscriptions, but there is an auth
	check for users down below, to make sure users cannot subscribe to other users
	inboxes/notifications or subscribe to rooms without being authenticated.
*/

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

type Connections struct {
	conns map[*websocket.Conn]primitive.ObjectID
	mutex sync.Mutex
}
type Subscriptions struct {
	subs  map[string]map[*websocket.Conn]primitive.ObjectID
	mutex sync.Mutex
}
type ConnectionsSubscriptionCount struct {
	counts map[*websocket.Conn]uint8 //Max subscriptions is 128... nice number half max uint8
	mutex  sync.Mutex
}
type VidChatStatus struct {
	data  map[*websocket.Conn]VidChatOpenData
	mutex sync.Mutex
}
type OpenConversations struct {
	data  map[primitive.ObjectID]map[primitive.ObjectID]struct{}
	mutex sync.Mutex
}

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

	// websocket Write/Read must be done from 1 goroutine. Queue all of them to be executed in a loop.
	MessageSendQueue chan QueuedMessage

	VidChatOpenChan  chan VidChatOpenData
	VidChatCloseChan chan *websocket.Conn
	VidChatStatus    VidChatStatus

	OpenConversations OpenConversations

	SendDataToUser chan UserDataMessage

	DestroySubscription chan string
}

func Init(colls *db.Collections) (*SocketServer, error) {
	socketServer := &SocketServer{
		Connections: Connections{
			conns: make(map[*websocket.Conn]primitive.ObjectID),
		},
		Subscriptions: Subscriptions{
			subs: make(map[string]map[*websocket.Conn]primitive.ObjectID),
		},
		ConnectionSubscriptionCount: ConnectionsSubscriptionCount{
			counts: make(map[*websocket.Conn]uint8),
		},

		RegisterConn:   make(chan ConnectionInfo),
		UnregisterConn: make(chan ConnectionInfo),

		RegisterSubscriptionConn:   make(chan SubscriptionConnectionInfo),
		UnregisterSubscriptionConn: make(chan SubscriptionConnectionInfo),

		SendDataToSubscription:           make(chan SubscriptionDataMessage),
		SendDataToSubscriptionExclusive:  make(chan ExclusiveSubscriptionDataMessage),
		SendDataToSubscriptions:          make(chan SubscriptionDataMessageMulti),
		SendDataToSubscriptionsExclusive: make(chan ExclusiveSubscriptionDataMessageMulti),

		MessageSendQueue: make(chan QueuedMessage),

		VidChatOpenChan:  make(chan VidChatOpenData),
		VidChatCloseChan: make(chan *websocket.Conn),
		VidChatStatus: VidChatStatus{
			data: make(map[*websocket.Conn]VidChatOpenData),
		},

		OpenConversations: OpenConversations{
			data: make(map[primitive.ObjectID]map[primitive.ObjectID]struct{}),
		},

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
					log.Println("Recovered from panic in WS registration : ", r)
				}
			}()
			connData := <-socketServer.RegisterConn
			if connData.Conn != nil {
				socketServer.Connections.mutex.Lock()
				socketServer.Connections.conns[connData.Conn] = connData.Uid
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
					log.Println("Recovered from panic in WS deregistration : ", r)
				}
			}()
			connData := <-socketServer.UnregisterConn
			for conn := range socketServer.Connections.conns {
				if conn == connData.Conn {
					socketServer.Connections.mutex.Lock()
					socketServer.Subscriptions.mutex.Lock()
					socketServer.VidChatStatus.mutex.Lock()
					socketServer.ConnectionSubscriptionCount.mutex.Lock()
					socketServer.OpenConversations.mutex.Lock()
					defer func() {
						socketServer.Connections.mutex.Unlock()
						socketServer.Subscriptions.mutex.Unlock()
						socketServer.VidChatStatus.mutex.Unlock()
						socketServer.ConnectionSubscriptionCount.mutex.Unlock()
						socketServer.OpenConversations.mutex.Unlock()
					}()
					delete(socketServer.Connections.conns, conn)
					delete(socketServer.VidChatStatus.data, conn)
					delete(socketServer.ConnectionSubscriptionCount.counts, conn)
					if connData.Uid != primitive.NilObjectID {
						delete(socketServer.OpenConversations.data, connData.Uid)
					}
					for _, r := range socketServer.Subscriptions.subs {
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
		}
	}()
	/* ----- Send messages in queue ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in queued socket messages : ", r)
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
					log.Println("Recovered from panic in subscription registration : ", r)
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
				count, countOk := socketServer.ConnectionSubscriptionCount.counts[connData.Conn]
				if count >= 128 {
					allow = false
				}
				// Passed all checks, add the connection to the subscription
				if allow {
					socketServer.Subscriptions.mutex.Lock()
					if socketServer.Subscriptions.subs[connData.Name] == nil {
						socketServer.Subscriptions.subs[connData.Name] = make(map[*websocket.Conn]primitive.ObjectID)
					}
					socketServer.Subscriptions.subs[connData.Name][connData.Conn] = connData.Uid
					socketServer.ConnectionSubscriptionCount.mutex.Lock()
					if countOk {
						socketServer.ConnectionSubscriptionCount.counts[connData.Conn]++
					} else {
						socketServer.ConnectionSubscriptionCount.counts[connData.Conn] = 1
					}
					defer func() {
						socketServer.Subscriptions.mutex.Unlock()
						socketServer.ConnectionSubscriptionCount.mutex.Unlock()
					}()
				}
			}
		}
	}()
	/* ----- Subscription disconnect registration ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in subscription disconnect registration : ", r)
				}
			}()
			connData := <-socketServer.UnregisterSubscriptionConn
			var err error
			if connData.Conn == nil {
				err = fmt.Errorf("Connection was nil")
			}
			if err != nil {
				if _, ok := socketServer.Subscriptions.subs[connData.Name]; ok {
					socketServer.Subscriptions.mutex.Lock()
					delete(socketServer.Subscriptions.subs[connData.Name], connData.Conn)
					socketServer.Subscriptions.mutex.Unlock()
				}
				socketServer.VidChatStatus.mutex.Lock()
				delete(socketServer.VidChatStatus.data, connData.Conn)
				socketServer.VidChatStatus.mutex.Unlock()
				if _, ok := socketServer.ConnectionSubscriptionCount.counts[connData.Conn]; ok {
					socketServer.ConnectionSubscriptionCount.mutex.Lock()
					socketServer.ConnectionSubscriptionCount.counts[connData.Conn]--
					socketServer.ConnectionSubscriptionCount.mutex.Unlock()
				}
			}
		}
	}()
	/* ----- Send data to subscription ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in subscription data channel : ", r)
				}
			}()
			subsData := <-socketServer.SendDataToSubscription
			for k, s := range socketServer.Subscriptions.subs {
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
		}
	}()
	/* ----- Send data to subscription excluding uids ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in exclusive subscription data channel : ", r)
				}
			}()
			subsData := <-socketServer.SendDataToSubscriptionExclusive
			for k, s := range socketServer.Subscriptions.subs {
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
		}
	}()
	/* ----- Send data to multiple subscriptions ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in subscription data channel : ", r)
				}
			}()
			subsData := <-socketServer.SendDataToSubscriptions
			for _, v := range subsData.Names {
				for k, s := range socketServer.Subscriptions.subs {
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
		}
	}()
	/* ----- Send data to multiple subscriptions excluding uids ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in exclusive subscription data channel : ", r)
				}
			}()
			subsData := <-socketServer.SendDataToSubscriptionsExclusive
			for _, v := range subsData.Names {
				for k, s := range socketServer.Subscriptions.subs {
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
		}
	}()
	/* ----- Send data to a specific user ----- */
	go func() {
		for {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("Recovered from panic in send data to user channel : ", r)
				}
			}()
			data := <-socketServer.SendDataToUser
			for conn, uid := range socketServer.Connections.conns {
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
		}
	}()
	/* ----- Destroy subscription ----- */
	go func() {
		for {
			subsName := <-socketServer.DestroySubscription
			for c := range socketServer.Subscriptions.subs[subsName] {
				if _, ok := socketServer.ConnectionSubscriptionCount.counts[c]; ok {
					socketServer.ConnectionSubscriptionCount.mutex.Lock()
					socketServer.ConnectionSubscriptionCount.counts[c]--
					socketServer.ConnectionSubscriptionCount.mutex.Unlock()
				}
			}
			socketServer.Subscriptions.mutex.Lock()
			delete(socketServer.Subscriptions.subs, subsName)
			socketServer.Subscriptions.mutex.Unlock()
		}
	}()
	/* ----- Vid chat opened chan ----- */
	go func() {
		for {
			data := <-socketServer.VidChatOpenChan
			defer func() {
				socketServer.VidChatStatus.mutex.Unlock()
			}()
			socketServer.VidChatStatus.mutex.Lock()
			socketServer.VidChatStatus.data[data.Conn] = data
		}
	}()
	/* ----- Vid chat closed chan ----- */
	go func() {
		for {
			data := <-socketServer.VidChatCloseChan
			defer func() {
				socketServer.VidChatStatus.mutex.Unlock()
			}()
			socketServer.VidChatStatus.mutex.Lock()
			delete(socketServer.VidChatStatus.data, data)
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
				for k, v := range socketServer.Subscriptions.subs {
					if len(v) == 0 {
						socketServer.DestroySubscription <- k
					}
				}
			case <-quitCleanup:
				cleanupTicker.Stop()
				return
			}
		}
	}()
}
