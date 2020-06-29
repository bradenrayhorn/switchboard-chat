package routing

import (
	"github.com/bradenrayhorn/switchboard-chat/hub"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ConnectWebsocket(c *gin.Context) {
	chatHub := c.MustGet("hub").(*hub.Hub)

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := hub.NewClient(ws, c.GetString("user_id"))
	chatHub.Register <- &client

	go client.StartWrite()
	go client.StartRead()
}
