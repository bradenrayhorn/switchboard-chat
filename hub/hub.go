package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bradenrayhorn/switchboard-chat/database"
	"github.com/bradenrayhorn/switchboard-chat/grpc"
	"github.com/segmentio/ksuid"
	"log"
	"sort"
	"sync"
)

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
	clients     map[ksuid.KSUID]*Client
	groups      map[string]map[ksuid.KSUID]*Client
	groupLock   sync.RWMutex
	groupChange chan bool
}

func NewHub(grpcClient *grpc.Client, redis *database.RedisDB) Hub {
	return Hub{
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		grpcClient:  grpcClient,
		redis:       redis,
		clients:     make(map[ksuid.KSUID]*Client, 0),
		groups:      make(map[string]map[ksuid.KSUID]*Client, 0),
		groupChange: make(chan bool),
	}
}

func (h *Hub) Start() {
	go h.startRedis()

	h.startProcessing()
}

// Starts Redis goroutine that monitors channels for incoming messages.
// This function opens a new Redis connection whenever a group is added or removed.
func (h *Hub) startRedis() {
	for {
		end := make(chan bool)
		go func() {
			log.Println("opening new redis subscription")
			ps := h.redis.Client.Subscribe(context.Background(), h.getGroupIDs()...)
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
						h.distributeMessage(data)
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

func (h *Hub) startProcessing() {
	for {
		select {
		case newClient := <-h.Register:
			// fetch groups
			groups, err := h.grpcClient.GetGroups(newClient.userId)
			if err != nil {
				log.Println(err)
				newClient.conn.Close()
			} else {
				// add client data
				newClient.hub = h
				sort.Strings(groups)
				newClient.groupIds = groups
				// add or update groups
				h.groupLock.Lock()
				for _, groupId := range groups {
					if _, ok := h.groups[groupId]; ok {
						h.groups[groupId][newClient.id] = newClient
					} else {
						h.groups[groupId] = map[ksuid.KSUID]*Client{newClient.id: newClient}
						h.groupChange <- true
					}
				}
				h.groupLock.Unlock()
				// add client to list
				h.clients[newClient.id] = newClient
				// send existing messages
				log.Printf("client count %d", len(h.clients))
			}
		case client := <-h.Unregister:
			client.conn.Close()
			delete(h.clients, client.id)
			h.groupLock.Lock()
			for _, groupId := range client.groupIds {
				delete(h.groups[groupId], client.id)
				if len(h.groups[groupId]) == 0 {
					delete(h.groups, groupId)
				}
			}
			h.groupLock.Unlock()
			log.Printf("client count %d", len(h.clients))
		}
	}
}

// Distributes a message to all clients connected to this chat instance.
func (h *Hub) distributeMessage(message GroupMessage) {
	h.groupLock.RLock()
	for _, clients := range h.groups[message.GroupId] {
		if message.ClientId == clients.id {
			continue
		}
		clients.sendQueue <- message
	}
	h.groupLock.RUnlock()
}

// Pushes a message to Redis for processing by Chat.
func (h *Hub) sendMessage(message GroupMessage) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return
	}
	h.redis.Client.Publish(context.Background(), message.GroupId, string(bytes))
}

func (h *Hub) getGroupIDs() []string {
	groupIDs := make([]string, len(h.groups))
	i := 0
	h.groupLock.RLock()
	for k := range h.groups {
		groupIDs[i] = k
		i++
	}
	h.groupLock.RUnlock()
	return groupIDs
}
