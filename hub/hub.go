package hub

import (
	"github.com/bradenrayhorn/switchboard-chat/grpc"
	"github.com/segmentio/ksuid"
	"log"
	"sort"
)

type GroupMessage struct {
	Message  string      `json:"message"`
	GroupId  string      `json:"group_id"`
	ClientId ksuid.KSUID `json:"client_id"`
	UserId   string      `json:"user_id"`
}

type Hub struct {
	Register   chan *Client
	Unregister chan *Client
	grpcClient *grpc.Client
	clients    map[ksuid.KSUID]*Client
	groups     map[string]map[ksuid.KSUID]*Client
}

func NewHub(grpcClient *grpc.Client) Hub {
	return Hub{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		grpcClient: grpcClient,
		clients:    make(map[ksuid.KSUID]*Client, 0),
		groups:     make(map[string]map[ksuid.KSUID]*Client, 0),
	}
}

func (h *Hub) Start() {
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
				for _, groupId := range groups {
					if _, ok := h.groups[groupId]; ok {
						h.groups[groupId][newClient.id] = newClient
					} else {
						h.groups[groupId] = map[ksuid.KSUID]*Client{newClient.id: newClient}
					}
				}
				// add client to list
				h.clients[newClient.id] = newClient
				// send existing messages
				// newClient.sendQueue <- "welcome"
				log.Printf("client count %d", len(h.clients))
			}
		case client := <-h.Unregister:
			client.conn.Close()
			delete(h.clients, client.id)
			for _, groupId := range client.groupIds {
				delete(h.groups[groupId], client.id)
			}
			log.Printf("client count %d", len(h.clients))
		}
	}
}

func (h *Hub) sendMessage(message GroupMessage) {
	for _, clients := range h.groups[message.GroupId] {
		if message.ClientId == clients.id {
			continue
		}
		clients.sendQueue <- message
	}
}
