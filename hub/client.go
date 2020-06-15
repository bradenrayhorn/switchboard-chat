package hub

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
	"sort"
)

type Client struct {
	id        ksuid.KSUID
	conn      *websocket.Conn
	hub       *Hub
	sendQueue chan interface{}
	userId    string
	groupIds  []string
}

func NewClient(conn *websocket.Conn, userId string) Client {
	return Client{
		id:        ksuid.New(),
		conn:      conn,
		sendQueue: make(chan interface{}),
		userId:    userId,
		groupIds:  make([]string, 0),
	}
}

func (c *Client) StartRead() {
	defer func() {
		c.conn.Close()
		c.hub.Unregister <- c
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// parse message
		data := GroupMessage{}
		err = json.Unmarshal(message, &data)
		if err == nil {
			if i := sort.SearchStrings(c.groupIds, data.GroupId); i < len(c.groupIds) && c.groupIds[i] == data.GroupId {
				data.UserId = c.userId
				data.ClientId = c.id
				c.hub.sendMessage(data)
			}
		}
	}
}

func (c *Client) StartWrite() {
	defer func() {
		c.conn.Close()
	}()
	for {
		jsonToSend, ok := <-c.sendQueue
		if !ok {
			fmt.Println("broke conn")
			break
		}
		err := c.conn.WriteJSON(jsonToSend)
		if err != nil {
			fmt.Println("broke conn: ")
			break
		}
	}
}
