package routing

import (
	"github.com/bradenrayhorn/switchboard-chat/hub"
	"github.com/bradenrayhorn/switchboard-chat/middleware"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

func MakeRouter(hub *hub.Hub) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.HubMiddleware(hub))
	applyRoutes(router)
	return router
}

func applyRoutes(router *gin.Engine) {
	base := router.Group(viper.GetString("base_url"))
	base.GET("/health-check", func(context *gin.Context) {
		context.String(http.StatusOK, "ok")
	})

	api := base.Group("")
	api.Use(middleware.AuthMiddleware())

	api.GET("/ws", ConnectWebsocket)
}
