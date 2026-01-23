package middleware

import (
	"net/http"

	"piemdm/pkg/helper/resp"
	"piemdm/pkg/jwt"
	"piemdm/pkg/log"

	"github.com/gin-gonic/gin"
)

func StrictAuth(j *jwt.JWT, logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.Request.Header.Get("Authorization")
		if tokenString == "" {
			logger.WithContext(ctx).Warn("请求未携带token,无权限访问",
				"url", ctx.Request.URL,
				"params", ctx.Params)
			resp.HandleError(ctx, http.StatusUnauthorized, "no token", nil)
			ctx.Abort()
			return
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			logger.WithContext(ctx).Error("token error",
				"url", ctx.Request.URL,
				"params", ctx.Params)
			resp.HandleError(ctx, http.StatusUnauthorized, err.Error(), nil)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		ctx.Set("user_id", claims.ID)
		ctx.Set("user_name", claims.UserName)
		ctx.Set("admin", claims.Admin)
		recoveryLoggerFunc(ctx, logger)
		ctx.Next()
	}
}

func NoStrictAuth(j *jwt.JWT, logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.Request.Header.Get("Authorization")

		if tokenString == "" {
			tokenString, _ = ctx.Cookie("accessToken")
		}
		if tokenString == "" {
			tokenString = ctx.Query("accessToken")
		}
		if tokenString == "" {
			ctx.Next()
			return
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			ctx.Next()
			return
		}

		ctx.Set("claims", claims)
		recoveryLoggerFunc(ctx, logger)
		ctx.Next()
	}
}

// AdminAuth 管理员权限中间件
func AdminAuth(j *jwt.JWT, logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 先进行基础认证
		tokenString := ctx.Request.Header.Get("Authorization")
		if tokenString == "" {
			resp.HandleError(ctx, http.StatusUnauthorized, "no token", nil)
			ctx.Abort()
			return
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			resp.HandleError(ctx, http.StatusUnauthorized, err.Error(), nil)
			ctx.Abort()
			return
		}

		// 检查是否为管理员
		if claims.Admin != "Yes" { // 需要实现 isAdmin 函数
			resp.HandleError(ctx, http.StatusForbidden, "admin access required", nil)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		ctx.Set("user_id", claims.ID)
		ctx.Set("user_name", claims.UserName)
		ctx.Set("admin", claims.Admin)
		ctx.Next()
	}
}

// UserAuth 普通用户权限中间件
func UserAuth(j *jwt.JWT, logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 基础认证，允许普通用户和管理员访问
		tokenString := ctx.Request.Header.Get("Authorization")
		if tokenString == "" {
			resp.HandleError(ctx, http.StatusUnauthorized, "no token", nil)
			ctx.Abort()
			return
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			resp.HandleError(ctx, http.StatusUnauthorized, err.Error(), nil)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		ctx.Set("user_id", claims.ID)
		ctx.Set("user_name", claims.UserName)
		ctx.Set("admin", claims.Admin)
		ctx.Next()
	}
}

func recoveryLoggerFunc(ctx *gin.Context, logger *log.Logger) {
	userInfo := ctx.MustGet("claims").(*jwt.CustomClaims)
	logger.NewContext(ctx, "UserId", userInfo.ID)
}
