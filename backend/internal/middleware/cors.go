package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		c.Header("Access-Control-Allow-Credentials", "true")
		// c.Header("Access-Control-Allow-Origin", "http://127.0.0.1:8081")
		// c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// c.Writer.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8081") // 替换为前端地址
		// c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// 主要是下面3项
		// Access-Control-Allow-Origin: *
		// Access-Control-Allow-Headers: Authorization, Content-Type
		// Access-Control-Allow-Methods: GET, POST, PATCH, PUT, DELETE

		if method == "OPTIONS" {
			c.Header("Access-Control-Allow-Methods", c.GetHeader("Access-Control-Request-Method"))
			c.Header("Access-Control-Allow-Headers", c.GetHeader("Access-Control-Request-Headers"))
			c.Header("Access-Control-Max-Age", "7200")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 显式禁用自动重定向
		// c.Request.URL.Path = strings.TrimSuffix(c.Request.URL.Path, "/")
		// 强制统一路径格式（去掉末尾斜杠）
		// c.Request.URL.Path = strings.TrimSuffix(c.Request.URL.Path, "/")
		c.Next()
	}
}
