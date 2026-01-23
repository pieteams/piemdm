package router

import (
	"piemdm/internal/handler"

	"github.com/gin-gonic/gin"
)

type Server struct {
	ServerHTTP          *gin.Engine
	notificationHandler handler.NotificationHandler
}

func NewServer(serverHTTP *gin.Engine, notificationHandler handler.NotificationHandler) *Server {
	return &Server{
		ServerHTTP:          serverHTTP,
		notificationHandler: notificationHandler,
	}
}
