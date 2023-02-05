package socketserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/web-stuff-98/go-social-media/pkg/db"
	"github.com/web-stuff-98/go-social-media/pkg/db/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
	This is for the chat. It handles JSON messages only.

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
type UserDataMessage struct {
	Uid  primitive.ObjectID
	Data interface{}
	Type string
}
type VidChatOpenData struct {
	Conn *websocket.Conn
	Id   primitive.ObjectID
}

type SocketServer struct {
	Connections                 map[*websocket.Conn]primitive.ObjectID
	Subscriptions               map[string]map[*websocket.Conn]primitive.ObjectID
	ConnectionSubscriptionCount map[*websocket.Conn]uint8 //Max subscriptions is 128... nice number half max uint8

	RegisterConn   chan ConnectionInfo
	UnregisterConn chan ConnectionInfo

	RegisterSubscriptionConn   chan SubscriptionConnectionInfo
	UnregisterSubscriptionConn chan SubscriptionConnectionInfo

	SendDataToSubscription           chan SubscriptionDataMessage
	SendDataToSubscriptionExclusive  chan ExclusiveSubscriptionDataMessage
	SendDataToSubscriptions          chan SubscriptionDataMessageMulti
	SendDataToSubscriptionsExclusive chan ExclusiveSubscriptionDataMessageMulti

	VidChatOpenChan  chan VidChatOpenData
	VidChatCloseChan chan *websocket.Conn
	VidChatStatus    map[*websocket.Conn]VidChatOpenData

	OpenConversations map[primitive.ObjectID]map[primitive.ObjectID]struct{}

	SendDataToUser chan UserDataMessage

	DestroySubscription chan string
}

func Init(colls *db.Collections) (*SocketServer, error) {
	socketServer := &SocketServer{
		Connections:                 make(map[*websocket.Conn]primitive.ObjectID),
		Subscriptions:               make(map[string]map[*websocket.Conn]primitive.ObjectID),
		ConnectionSubscriptionCount: make(map[*websocket.Conn]uint8),

		RegisterConn:   make(chan ConnectionInfo),
		UnregisterConn: make(chan ConnectionInfo),

		RegisterSubscriptionConn:   make(chan SubscriptionConnectionInfo),
		UnregisterSubscriptionConn: make(chan SubscriptionConnectionInfo),

		SendDataToSubscription:           make(chan SubscriptionDataMessage),
		SendDataToSubscriptionExclusive:  make(chan ExclusiveSubscriptionDataMessage),
		SendDataToSubscriptions:          make(chan SubscriptionDataMessageMulti),
		SendDataToSubscriptionsExclusive: make(chan ExclusiveSubscriptionDataMessageMulti),

		VidChatOpenChan:  make(chan VidChatOpenData),
		VidChatCloseChan: make(chan *websocket.Conn),
		VidChatStatus:    make(map[*websocket.Conn]VidChatOpenData),

		OpenConversations: make(map[primitive.ObjectID]map[primitive.ObjectID]struct{}),

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
				socketServer.Connections[connData.Conn] = connData.Uid
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
			for conn := range socketServer.Connections {
				if conn == connData.Conn {
					delete(socketServer.Connections, conn)
					delete(socketServer.VidChatStatus, conn)
					delete(socketServer.ConnectionSubscriptionCount, conn)
					if connData.Uid != primitive.NilObjectID {
						delete(socketServer.OpenConversations, connData.Uid)
					}
					for _, r := range socketServer.Subscriptions {
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
						if err != nil {
							room := &models.Room{}
							if err := colls.RoomCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&room); err != nil {
								allow = false
							} else {
								if room.Author != connData.Uid {
									allow = false
								}
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
						} else {
							allow = false
						}
					}
				}
				// Make sure users cannot open too many subscriptions
				count, countOk := socketServer.ConnectionSubscriptionCount[connData.Conn]
				if count >= 128 {
					allow = false
				}
				// Passed all checks, add the connection to the subscription
				if allow {
					if socketServer.Subscriptions[connData.Name] == nil {
						socketServer.Subscriptions[connData.Name] = make(map[*websocket.Conn]primitive.ObjectID)
					}
					socketServer.Subscriptions[connData.Name][connData.Conn] = connData.Uid
					if countOk {
						socketServer.ConnectionSubscriptionCount[connData.Conn]++
					} else {
						socketServer.ConnectionSubscriptionCount[connData.Conn] = 1
					}
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
				if _, ok := socketServer.Subscriptions[connData.Name]; ok {
					delete(socketServer.Subscriptions[connData.Name], connData.Conn)
				}
				delete(socketServer.VidChatStatus, connData.Conn)
				if _, ok := socketServer.ConnectionSubscriptionCount[connData.Conn]; ok {
					socketServer.ConnectionSubscriptionCount[connData.Conn]--
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
			for k, s := range socketServer.Subscriptions {
				if k == subsData.Name {
					for conn := range s {
						conn.WriteMessage(websocket.TextMessage, subsData.Data)
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
			for k, s := range socketServer.Subscriptions {
				if k == subsData.Name {
					for conn, oid := range s {
						if subsData.Exclude[oid] != true {
							conn.WriteMessage(websocket.TextMessage, subsData.Data)
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
				for k, s := range socketServer.Subscriptions {
					if k == v {
						for conn := range s {
							conn.WriteMessage(websocket.TextMessage, subsData.Data)
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
				for k, s := range socketServer.Subscriptions {
					if k == v {
						for conn, oid := range s {
							if subsData.Exclude[oid] != true {
								conn.WriteMessage(websocket.TextMessage, subsData.Data)
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
			for conn, uid := range socketServer.Connections {
				if data.Uid == uid {
					var m map[string]interface{}
					outBytesNoTypeKey, err := json.Marshal(data.Data)
					json.Unmarshal(outBytesNoTypeKey, &m)
					m["TYPE"] = data.Type
					outBytes, err := json.Marshal(m)
					if err == nil {
						conn.WriteMessage(websocket.TextMessage, outBytes)
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
			for c := range socketServer.Subscriptions[subsName] {
				if _, ok := socketServer.ConnectionSubscriptionCount[c]; ok {
					socketServer.ConnectionSubscriptionCount[c]--
				}
			}
			delete(socketServer.Subscriptions, subsName)
		}
	}()
	/* ----- Vid chat opened chan ----- */
	go func() {
		for {
			data := <-socketServer.VidChatOpenChan
			socketServer.VidChatStatus[data.Conn] = data
		}
	}()
	/* ----- Vid chat closed chan ----- */
	go func() {
		for {
			data := <-socketServer.VidChatCloseChan
			delete(socketServer.VidChatStatus, data)
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
				for k, v := range socketServer.Subscriptions {
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
