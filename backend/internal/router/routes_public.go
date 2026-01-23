package router

import (
	"github.com/gin-gonic/gin"
)

func registerPublicRoutes(r *gin.RouterGroup, h *Handlers) {
	noAuth := r.Group("")
	{

		noAuth.POST("/auth/login", h.User.Login)
		noAuth.POST("/auth/validate", h.User.ValidateToken)
		// noAuth.GET("/auth/profile", user.GetProfile)
		// noAuth.POST("/auth/register", user.Register)
		// noAuth.POST("/auth/refresh", user.RefreshToken)
	}
}
