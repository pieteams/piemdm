package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"piemdm/internal/model"
	"piemdm/internal/service"
	"piemdm/pkg/helper/resp"
	"piemdm/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/pieworks/piemdm/packages/go/openapi/errors"
	"github.com/pieworks/piemdm/packages/go/openapi/spec"
	"github.com/spf13/viper"
)

// CanonicalRequestMiddleware Canonical Request 签名验证中间件
func CanonicalRequestMiddleware(
	authService service.OpenApiAuthService,
	logger *log.Logger,
	conf *viper.Viper,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		appId := c.GetHeader(spec.HeaderAppID)
		timestamp := c.GetHeader(spec.HeaderTimestamp)
		nonce := c.GetHeader(spec.HeaderNonce)
		sign := c.GetHeader(spec.HeaderSignature)

		if appId == "" || timestamp == "" || nonce == "" || sign == "" {
			logger.Warn("Missing required headers")
			// 使用标准错误码
			resp.HandleError(c, errors.ErrAuthFailed.HTTPStatus(), errors.ErrAuthFailed.Error(), nil)
			c.Abort()
			return
		}

		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		app, err := authService.VerifySignature(c, appId, timestamp, nonce, sign, body)
		if err != nil {
			logger.Warn("Signature verification failed", "error", err, "app_id", appId)
			// err 已经是 errors.ErrorCode 类型 (由 service 返回)
			// resp.HandleError 可能需要适配，这里直接传递err.Error()保持兼容，或者解析Code
			status := http.StatusUnauthorized
			if e, ok := err.(errors.ErrorCode); ok {
				status = e.HTTPStatus()
			}
			resp.HandleError(c, status, err.Error(), nil)
			c.Abort()
			return
		}

		clientIP := c.ClientIP()
		if err := authService.VerifyIPWhitelist(app, clientIP); err != nil {
			logger.Warn("IP whitelist check failed", "error", err, "app_id", appId, "client_ip", clientIP)
			status := http.StatusForbidden
			if e, ok := err.(errors.ErrorCode); ok {
				status = e.HTTPStatus()
			}
			resp.HandleError(c, status, err.Error(), nil)
			c.Abort()
			return
		}

		nonceTTL := getNonceTTL(conf)
		if err := authService.CheckAndRecordNonce(nonce, nonceTTL); err != nil {
			logger.Warn("Nonce check failed", "error", err, "nonce", nonce)
			status := http.StatusUnauthorized
			if e, ok := err.(errors.ErrorCode); ok {
				status = e.HTTPStatus()
			}
			resp.HandleError(c, status, err.Error(), nil)
			c.Abort()
			return
		}

		c.Set("application", app)
		logger.Debug("Canonical request verified", "app_id", appId)
		c.Next()
	}
}

// getNonceTTL 从配置读取 Nonce TTL,默认 10 分钟
func getNonceTTL(conf *viper.Viper) time.Duration {
	ttlMinutes := conf.GetInt("openapi.nonce_ttl_minutes")
	if ttlMinutes == 0 {
		return spec.DefaultNonceTTL
	}
	return time.Duration(ttlMinutes) * time.Minute
}

// OpenApiEntityPermissionMiddleware Entity 权限校验中间件
func OpenApiEntityPermissionMiddleware(
	authService service.OpenApiAuthService,
	logger *log.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		app, exists := c.Get("application")
		if !exists {
			logger.Error("Application not found in context")
			resp.HandleError(c, errors.ErrAuthFailed.HTTPStatus(), errors.ErrAuthFailed.Error(), nil)
			c.Abort()
			return
		}

		application, ok := app.(*model.Application)
		if !ok {
			logger.Error("Invalid application type in context")
			resp.HandleError(c, errors.ErrSystemError.HTTPStatus(), errors.ErrSystemError.Error(), nil)
			c.Abort()
			return
		}

		tableCode := c.Param("table")
		if tableCode == "" {
			resp.HandleError(c, errors.ErrParamMissing.HTTPStatus(), errors.ErrParamMissing.Error(), nil)
			c.Abort()
			return
		}

		if err := authService.VerifyEntityAccess(application.AppId, tableCode); err != nil {
			logger.Warn("Entity access denied", "app_id", application.AppId, "table", tableCode, "error", err)
			status := http.StatusForbidden
			if e, ok := err.(errors.ErrorCode); ok {
				status = e.HTTPStatus()
			}
			resp.HandleError(c, status, err.Error(), nil)
			c.Abort()
			return
		}

		logger.Debug("Entity access granted", "app_id", application.AppId, "table", tableCode)
		c.Next()
	}
}
