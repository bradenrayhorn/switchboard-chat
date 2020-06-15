package routing

import (
	"github.com/bradenrayhorn/switchboard-chat/hub"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func ConnectWebsocket(c *gin.Context) {
	chatHub := c.MustGet("hub").(*hub.Hub)

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := hub.NewClient(ws, c.GetString("user_id"))
	chatHub.Register <- &client

	go client.StartWrite()
	go client.StartRead()
}
