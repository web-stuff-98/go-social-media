package socketserver

import (
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
	This facilitiates connection registrations and subscriptions. The socket handler
	file handles messages from the client.

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
	Data map[string]string
}
type ExclusiveSubscriptionDataMessage struct {
	Name    string
	Data    map[string]string
	Exclude map[primitive.ObjectID]bool
}

type SocketServer struct {
	Connections   map[*websocket.Conn]primitive.ObjectID
	Subscriptions map[string]map[*websocket.Conn]primitive.ObjectID

	RegisterConn   chan ConnectionInfo
	UnregisterConn chan ConnectionInfo

	RegisterSubscriptionConn   chan SubscriptionConnectionInfo
	UnregisterSubscriptionConn chan SubscriptionConnectionInfo

	SendDataToSubscription          chan SubscriptionDataMessage
	SendDataToSubscriptionExclusive chan ExclusiveSubscriptionDataMessage

	DestroySubscription chan string
}

func Init() (*SocketServer, error) {
	socketServer := &SocketServer{
		Connections:                     make(map[*websocket.Conn]primitive.ObjectID),
		Subscriptions:                   make(map[string]map[*websocket.Conn]primitive.ObjectID),
		RegisterConn:                    make(chan ConnectionInfo),
		UnregisterConn:                  make(chan ConnectionInfo),
		RegisterSubscriptionConn:        make(chan SubscriptionConnectionInfo),
		UnregisterSubscriptionConn:      make(chan SubscriptionConnectionInfo),
		SendDataToSubscription:          make(chan SubscriptionDataMessage),
		SendDataToSubscriptionExclusive: make(chan ExclusiveSubscriptionDataMessage),
		DestroySubscription:             make(chan string),
	}
	RunServer(socketServer)
	return socketServer, nil
}

func RunServer(socketServer *SocketServer) {
	/* ----- Websocket connection registration ----- */
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
	/* ----- Websocket disconnect registration ----- */
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
						conn.WriteJSON(subsData.Data)
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
							conn.WriteJSON(subsData.Data)
						}
					}
					break
				}
			}
		}
	}()
	/* ----- Destroy subscription (for when a post is deleted for example) ----- */
	go func() {
		for {
			subsName := <-socketServer.DestroySubscription
			delete(socketServer.Subscriptions, subsName)
		}
	}()

	/* -------- Cleanup ticker -------- */
	cleanupTicker := time.NewTicker(2 * time.Minute)
	quitCleanup := make(chan struct{})
	go func() {
		for {
			select {
			case <-cleanupTicker.C:
				// Remove subscriptions nobody is connected to
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
