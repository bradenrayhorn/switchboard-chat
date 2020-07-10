package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bradenrayhorn/switchboard-chat/database"
	"github.com/bradenrayhorn/switchboard-chat/grpc"
	"github.com/segmentio/ksuid"
	"log"
	"sync"
)

const GroupChannelPrefix = "group-"
const UserChannelPrefix = "user-"

type GroupMessage struct {
	Message  string      `json:"message"`
	GroupId  string      `json:"group_id"`
	ClientId ksuid.KSUID `json:"client_id"`
	UserId   string      `json:"user_id"`
}

type Hub struct {
	Register    chan *Client
	Unregister  chan *Client
	grpcClient  *grpc.Client
	redis       *database.RedisDB
	clientLock  sync.RWMutex
	clients     map[ksuid.KSUID]*Client
	users       *Users
	groups      *Groups
	groupChange chan bool
	usersChange chan bool
}

func NewHub(grpcClient *grpc.Client, redis *database.RedisDB) Hub {
	return Hub{
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		grpcClient:  grpcClient,
		redis:       redis,
		clients:     make(map[ksuid.KSUID]*Client, 0),
		users:       NewUsers(),
		groups:      NewGroups(),
		groupChange: make(chan bool),
		usersChange: make(chan bool),
	}
}

func (h *Hub) Start() {
	go h.startChatRedis()
	go h.startUserRedis()

	h.startProcessing()
}

// Starts Redis goroutine that monitors channels for incoming messages.
// This function opens a new Redis connection whenever a group is added or removed.
func (h *Hub) startChatRedis() {
	for {
		end := make(chan bool)
		go func() {
			log.Println("opening new redis subscription")
			ps := h.redis.Client.Subscribe(context.Background(), h.groups.getGroupIDs(GroupChannelPrefix)...)
		redisLoop:
			for {
				select {
				case msg, ok := <-ps.Channel():
					if !ok {
						log.Println("invalid message received")
						break
					}
					fmt.Printf("got message on %s = %s;\n", msg.Channel, msg.Payload)
					data := GroupMessage{}
					err := json.Unmarshal([]byte(msg.Payload), &data)
					if err == nil {
						h.distributeChatMessage(data)
					}
				case <-end:
					log.Println("closing redis subscription")
					err := ps.Close()
					if err != nil {
						log.Println(err)
					}
					break redisLoop
				}
			}
		}()
		<-h.groupChange
		end <- true
	}
}

// Starts Redis goroutine that monitors channels for changes regarding a user's groups
// This function opens a new Redis connection whenever a user is added or removed.
func (h *Hub) startUserRedis() {
	for {
		end := make(chan bool)
		go func() {
			log.Println("opening new user redis subscription")
			ps := h.redis.Client.Subscribe(context.Background(), h.users.getUserIDs(UserChannelPrefix)...)
		redisLoop:
			for {
				select {
				case msg, ok := <-ps.Channel():
					if !ok {
						log.Println("invalid message received")
						break
					}
					fmt.Printf("got message on %s = %s;\n", msg.Channel, msg.Payload)
					message := RedisMessage{}
					err := json.Unmarshal([]byte(msg.Payload), &message)
					if err == nil {
						switch message.RedisMessageType {
						case RedisGroupsChanged:
							groupMessage := RedisGroupChangedMessage{}
							err := json.Unmarshal([]byte(msg.Payload), &groupMessage)
							if err == nil {
								userID := msg.Channel[len(UserChannelPrefix):]
								groups := groupMessage.Body.Groups
								anyChanged := false
								// make message
								message := SocketMessage{
									SocketMessageType: GroupChange,
								}
								// for every client user is in update groups
								for _, clientID := range h.users.getClientIDs(userID) {
									// update data
									h.clientLock.Lock()
									changed := h.groups.updateClient(groups, h.clients[clientID])
									h.clients[clientID].groupIDs = groups
									if changed {
										anyChanged = true
									}
									// send message to client
									h.clients[clientID].sendMessage(message)
									h.clientLock.Unlock()
								}
								// if any group changed update subscription
								if anyChanged {
									h.groupChange <- true
								}
							}
							break
						}
					}
				case <-end:
					log.Println("closing user redis subscription")
					err := ps.Close()
					if err != nil {
						log.Println(err)
					}
					break redisLoop
				}
			}
		}()
		<-h.usersChange
		end <- true
	}
}

func (h *Hub) startProcessing() {
	for {
		select {
		case newClient := <-h.Register:
			// fetch groups
			groups, err := h.grpcClient.GetGroups(newClient.userID)
			if err != nil {
				log.Println(err)
				newClient.conn.Close()
			} else {
				// add client data
				newClient.hub = h
				newClient.groupIDs = groups
				// add client to users list
				h.users.addUser(newClient.userID, newClient.id)
				h.usersChange <- true
				// add or update groups
				changed := h.groups.addClient(groups, newClient)
				if changed {
					h.groupChange <- true
				}
				// add client to list
				h.clientLock.Lock()
				h.clients[newClient.id] = newClient
				h.clientLock.Unlock()
				// send existing messages
				log.Printf("client count %d", len(h.clients))
			}
		case client := <-h.Unregister:
			client.conn.Close()
			// remove from users list
			h.users.removeUser(client.userID, client.id)
			h.usersChange <- true
			// remove from clients list
			h.clientLock.Lock()
			delete(h.clients, client.id)
			h.clientLock.Unlock()
			// remove from and groups
			changed := h.groups.removeClient(client)
			if changed {
				h.groupChange <- true
			}
			log.Printf("client count %d", len(h.clients))
		}
	}
}

// Sends a chat message to all clients in a group.
func (h *Hub) distributeChatMessage(message GroupMessage) {
	socketMessage := SocketMessage{
		SocketMessageType: Message,
		Body:              message,
	}
	for _, client := range h.groups.getClientMap(message.GroupId) {
		if message.ClientId == client.id {
			continue
		}
		client.sendMessage(socketMessage)
	}
}

// Pushes a message to Redis for processing by Chat.
func (h *Hub) sendMessage(message GroupMessage) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return
	}
	h.redis.Client.Publish(context.Background(), GroupChannelPrefix+message.GroupId, string(bytes))
}
