package middleware

import (
	"bytes"
	"io"
	"time"

	"piemdm/internal/model"
	"piemdm/internal/service"
	"piemdm/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// responseWriter 包装 gin.ResponseWriter 以捕获响应
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// OpenApiAuditMiddleware OpenAPI 审计日志中间件
func OpenApiAuditMiddleware(
	auditService service.ApplicationApiLogService,
	logger *log.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 生成 Request ID (Trace ID)
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// 2. 记录请求开始时间
		startTime := time.Now()

		// 3. 读取请求体(用于审计日志)
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重新设置请求体
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 4. 包装 ResponseWriter 以捕获响应
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// 5. 执行请求
		c.Next()

		// 6. 计算请求耗时
		duration := time.Since(startTime)

		// 7. 获取 Application ID
		applicationID := c.GetHeader("X-App-Id")
		if applicationID == "" {
			applicationID = "unknown"
		}

		// 8. 确定日志类型和结果
		logType := "ACCESS"
		outcome := "SUCCESS"
		errorCode := ""
		errorMessage := ""

		httpStatus := c.Writer.Status()

		// 根据 HTTP 状态码判断结果
		if httpStatus >= 400 && httpStatus < 500 {
			if httpStatus == 401 || httpStatus == 403 {
				logType = "SECURITY"
				outcome = "AUTH_FAILED"
			} else if httpStatus == 429 {
				logType = "SECURITY"
				outcome = "RATE_LIMITED"
			} else {
				outcome = "FAILED"
			}
		} else if httpStatus >= 500 {
			logType = "ERROR"
			outcome = "FAILED"
		}

		// 从响应中提取错误信息(如果有)
		responsePayload := ""
		if outcome != "SUCCESS" {
			responsePayload = blw.body.String()
			// 简单解析错误信息(实际应该解析 JSON)
			// 这里简化处理,实际应该根据 resp.Response 结构解析
			if len(responsePayload) > 500 {
				responsePayload = responsePayload[:500] + "..."
			}
		}

		// 9. 创建审计日志
		auditLog := &model.ApplicationApiLog{
			RequestId:     requestID,
			ApplicationId: applicationID,
			LogType:       logType,

			HttpMethod:     c.Request.Method,
			RequestPath:    c.Request.URL.Path,
			QueryParams:    c.Request.URL.RawQuery,
			RequestPayload: string(requestBody),
			ClientIp:       c.ClientIP(),
			UserAgent:      c.Request.UserAgent(),

			HttpStatus:      httpStatus,
			ResponsePayload: responsePayload,
			DurationMs:      int(duration.Milliseconds()),

			Outcome:      outcome,
			ErrorCode:    errorCode,
			ErrorMessage: errorMessage,

			AffectedResourceIds: nil,
		}

		// 10. 异步写入审计日志
		go func() {
			if err := auditService.Create(c.Copy(), auditLog); err != nil {
				logger.Error("Failed to create audit log", "error", err, "request_id", requestID)
			}
		}()

		logger.Debug("Audit log created",
			"request_id", requestID,
			"app_id", applicationID,
			"path", c.Request.URL.Path,
			"status", httpStatus,
			"duration_ms", duration.Milliseconds())
	}
}
