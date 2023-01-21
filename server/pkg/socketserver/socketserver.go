package socketserver

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
	This is for the chat. It handles JSON messages only.

	Uid can always be left as primitive.NilObjectID, users are not required
	to be authenticated to connect or open subscriptions, but there is an auth
	check for users down below, to make sure users cannot subscribe to other users
	inboxes or subscribe to rooms without being authenticated.
*/

type ConnectionInfo struct {
	Conn *websocket.Conn
	Uid  primitive.ObjectID
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

type SocketServer struct {
	Connections   map[*websocket.Conn]primitive.ObjectID
	Subscriptions map[string]map[*websocket.Conn]primitive.ObjectID

	RegisterConn   chan ConnectionInfo
	UnregisterConn chan ConnectionInfo

	RegisterSubscriptionConn   chan SubscriptionConnectionInfo
	UnregisterSubscriptionConn chan SubscriptionConnectionInfo

	SendDataToSubscription           chan SubscriptionDataMessage
	SendDataToSubscriptionExclusive  chan ExclusiveSubscriptionDataMessage
	SendDataToSubscriptions          chan SubscriptionDataMessageMulti
	SendDataToSubscriptionsExclusive chan ExclusiveSubscriptionDataMessageMulti

	SendDataToUser chan UserDataMessage

	DestroySubscription chan string
}

func Init() (*SocketServer, error) {
	socketServer := &SocketServer{
		Connections:   make(map[*websocket.Conn]primitive.ObjectID),
		Subscriptions: make(map[string]map[*websocket.Conn]primitive.ObjectID),

		RegisterConn:   make(chan ConnectionInfo),
		UnregisterConn: make(chan ConnectionInfo),

		RegisterSubscriptionConn:   make(chan SubscriptionConnectionInfo),
		UnregisterSubscriptionConn: make(chan SubscriptionConnectionInfo),

		SendDataToSubscription:           make(chan SubscriptionDataMessage),
		SendDataToSubscriptionExclusive:  make(chan ExclusiveSubscriptionDataMessage),
		SendDataToSubscriptions:          make(chan SubscriptionDataMessageMulti),
		SendDataToSubscriptionsExclusive: make(chan ExclusiveSubscriptionDataMessageMulti),

		SendDataToUser: make(chan UserDataMessage),

		DestroySubscription: make(chan string),
	}
	RunServer(socketServer)
	return socketServer, nil
}

func RunServer(socketServer *SocketServer) {
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
				// Make sure users cannot subscribe to rooms without being logged in
				if strings.Contains(connData.Name, "room=") {
					if connData.Uid == primitive.NilObjectID {
						allow = false
					}
				}
				if allow {
					if socketServer.Subscriptions[connData.Name] == nil {
						socketServer.Subscriptions[connData.Name] = make(map[*websocket.Conn]primitive.ObjectID)
					}
					socketServer.Subscriptions[connData.Name][connData.Conn] = connData.Uid
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
			delete(socketServer.Subscriptions[connData.Name], connData.Conn)
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
					outBytes, err := json.Marshal(data.Data)
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
			delete(socketServer.Subscriptions, subsName)
		}
	}()

	/* -------- Cleanup ticker -------- */
	cleanupTicker := time.NewTicker(20 * time.Minute)
	quitCleanup := make(chan struct{})
	go func() {
		for {
			select {
			case <-cleanupTicker.C:
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
