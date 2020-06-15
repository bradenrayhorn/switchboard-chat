package middleware

import (
	"github.com/bradenrayhorn/switchboard-chat/hub"
	"github.com/gin-gonic/gin"
)

func HubMiddleware(hub *hub.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("hub", hub)
		c.Next()
	}
}
