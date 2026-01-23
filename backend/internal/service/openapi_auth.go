package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"piemdm/internal/model"
	"piemdm/internal/repository"
	"piemdm/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/pieworks/piemdm/packages/go/openapi/auth"
	"github.com/pieworks/piemdm/packages/go/openapi/errors"
	"github.com/pieworks/piemdm/packages/go/openapi/spec"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type OpenApiAuthService interface {
	// VerifySignature 验证 Canonical Request 签名
	VerifySignature(c *gin.Context, appId, timestamp, nonce, sign string, body []byte) (*model.Application, error)

	// VerifyIPWhitelist 验证 IP 白名单
	VerifyIPWhitelist(app *model.Application, clientIP string) error

	// VerifyEntityAccess 验证 Entity 访问权限
	VerifyEntityAccess(appId, entityCode string) error

	// CheckAndRecordNonce 检查并记录 Nonce(防重放)
	CheckAndRecordNonce(nonce string, ttl time.Duration) error
}

type openApiAuthService struct {
	logger                *log.Logger
	applicationRepo       repository.ApplicationRepository
	applicationEntityRepo repository.ApplicationEntityRepository
	redisClient           *redis.Client
	conf                  *viper.Viper
}

func NewOpenApiAuthService(
	logger *log.Logger,
	applicationRepo repository.ApplicationRepository,
	applicationEntityRepo repository.ApplicationEntityRepository,
	redisClient *redis.Client,
	conf *viper.Viper,
) OpenApiAuthService {
	return &openApiAuthService{
		logger:                logger,
		applicationRepo:       applicationRepo,
		applicationEntityRepo: applicationEntityRepo,
		redisClient:           redisClient,
		conf:                  conf,
	}
}

// VerifySignature 验证 Canonical Request 签名
func (s *openApiAuthService) VerifySignature(
	c *gin.Context,
	appId, timestamp, nonce, sign string,
	body []byte,
) (*model.Application, error) {
	app, err := s.applicationRepo.FindByAppId(appId)
	if err != nil {
		s.logger.Warn("Application not found", "app_id", appId, "error", err)
		return nil, errors.ErrAuthFailed
	}

	// 验证时间戳 (也可以使用 spec/utils 或 auth 中的工具，这里保留原有逻辑或迁移)
	// TODO: 建议使用 openapi-go 提供的工具函数解析
	ts, err := parseTimestamp(timestamp)
	if err != nil {
		return nil, errors.ErrTokenExpired
	}

	if !isTimestampValid(ts, getTimestampWindow(s.conf)) {
		return nil, errors.ErrTokenExpired
	}

	// 构建 Canonical Request
	canonicalRequest := auth.BuildCanonicalRequest(
		c.Request.Method,
		c.Request.URL.Path,
		c.Request.URL.Query(),
		body,
		timestamp,
		nonce,
	)

	s.logger.Debug("Canonical Request", "request", canonicalRequest)

	// 验证签名
	if !auth.VerifySignature(sign, canonicalRequest, app.AppSecret) {
		s.logger.Warn("Signature mismatch", "app_id", appId)
		return nil, errors.ErrSignatureInvalid
	}

	return app, nil
}

// getTimestampWindow 从配置读取时间窗口,默认 5 分钟
func getTimestampWindow(conf *viper.Viper) time.Duration {
	windowMinutes := conf.GetInt("openapi.timestamp_window_minutes")
	if windowMinutes == 0 {
		return spec.DefaultTimestampWindow
	}
	return time.Duration(windowMinutes) * time.Minute
}

// isTimestampValid 验证时间戳是否在允许的窗口内
func isTimestampValid(ts time.Time, window time.Duration) bool {
	now := time.Now()
	// 允许一定的时间误差
	return now.Sub(ts) <= window && ts.Sub(now) <= window
}

// parseTimestamp 解析时间戳
func parseTimestamp(timestamp string) (time.Time, error) {
	// 支持 Unix 时间戳(秒)
	ts, err := time.Parse("1136239445", timestamp) // 奇怪的格式，go 应该用 20060102...
	// 这里保留原有逻辑，或者修正。原有逻辑可能是想parse int string
	// 修正为标准 Unix Timestamp 解析

	var unixTime int64
	_, err2 := fmt.Sscanf(timestamp, "%d", &unixTime)
	if err2 == nil {
		return time.Unix(unixTime, 0), nil
	}

	return ts, err
}

// VerifyIPWhitelist 验证 IP 白名单
func (s *openApiAuthService) VerifyIPWhitelist(app *model.Application, clientIP string) error {
	if app.IP == "" {
		// 如果没有配置 IP 白名单,则允许所有 IP
		return nil
	}

	// 支持逗号分隔的多个 IP
	allowedIPs := strings.Split(app.IP, ",")
	for _, ip := range allowedIPs {
		if strings.TrimSpace(ip) == clientIP {
			return nil
		}
	}

	s.logger.Warn("IP not in whitelist", "app_id", app.AppId, "client_ip", clientIP, "allowed_ips", app.IP)
	return errors.ErrIpNotAllowed
}

// VerifyEntityAccess 验证 Entity 访问权限
func (s *openApiAuthService) VerifyEntityAccess(appId, entityCode string) error {
	// 查询 ApplicationEntity
	entity, err := s.applicationEntityRepo.FindByAppIdAndEntityCode(appId, entityCode)
	if err != nil {
		s.logger.Warn("Entity access denied", "app_id", appId, "entity_code", entityCode, "error", err)
		return errors.ErrPermissionDenied
	}

	// 检查状态
	if entity.Status != "Normal" {
		return errors.ErrPermissionDenied
	}

	// 检查过期时间
	if !entity.ExpiredAt.IsZero() && time.Now().After(entity.ExpiredAt) {
		return errors.ErrPermissionDenied
	}

	return nil
}

// CheckAndRecordNonce 检查并记录 Nonce(防重放)
func (s *openApiAuthService) CheckAndRecordNonce(nonce string, ttl time.Duration) error {
	key := fmt.Sprintf("openapi:nonce:%s", nonce)

	if ttl == 0 {
		ttl = spec.DefaultNonceTTL
	}

	// 尝试设置 Nonce,如果已存在则返回错误
	ctx := context.Background()
	success, err := s.redisClient.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		s.logger.Error("Failed to check nonce", "error", err)
		return errors.ErrSystemError
	}

	if !success {
		s.logger.Warn("Nonce already used", "nonce", nonce)
		return errors.ErrNonceUsed
	}

	return nil
}
