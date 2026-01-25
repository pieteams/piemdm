package router

import (
	"piemdm/internal/handler"
	"piemdm/internal/integration/feishu"
	"piemdm/internal/service"

	"github.com/gin-gonic/gin"
)

type Server struct {
	ServerHTTP          *gin.Engine
	notificationHandler handler.NotificationHandler
	FeishuService       *feishu.Service
	ApprovalService     service.ApprovalService
}

func NewServer(serverHTTP *gin.Engine, notificationHandler handler.NotificationHandler, feishuService *feishu.Service, approvalService service.ApprovalService) *Server {
	return &Server{
		ServerHTTP:          serverHTTP,
		notificationHandler: notificationHandler,
		FeishuService:       feishuService,
		ApprovalService:     approvalService,
	}
}
