// Package server provides HTTP server configuration and routing for the PieMDM API.
// It contains the main server setup, route definitions, middleware configuration,
// and HTTP request handling infrastructure using Gin framework.
package router

import (
	"net/http"

	"piemdm/internal/handler"
	"piemdm/internal/middleware"
	"piemdm/internal/repository"
	"piemdm/pkg/jwt"
	"piemdm/pkg/log"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"

	// Swagger imports
	_ "piemdm/docs" // swagger docs

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewServerHTTP(
	logger *log.Logger,
	jwt *jwt.JWT,
	entity handler.EntityHandler,
	approval handler.ApprovalHandler,
	approvalDefinition handler.ApprovalDefinitionHandler,
	approvalNode handler.ApprovalNodeHandler,
	approvalTask handler.ApprovalTaskHandler,
	table handler.TableHandler,
	tableField handler.TableFieldHandler,

	application handler.ApplicationHandler,
	webhook handler.WebhookHandler,
	webhookDelivery handler.WebhookDeliveryHandler,
	cron handler.CronHandler,
	cronLog handler.CronLogHandler,
	role handler.RoleHandler,
	permission handler.PermissionHandler,
	user handler.UserHandler,
	notification handler.NotificationHandler,
	notificationTemplate handler.NotificationTemplateHandler,
	notificationLog handler.NotificationLogHandler,
	tableApprovalDefinition handler.TableApprovalDefinitionHandler,
	upload handler.UploadHandler,
	tablePermission handler.TablePermissionHandler,

	// OpenAPI
	openApi handler.OpenApiHandler,

	// OpenAPI dependencies (repositories)
	applicationRepo repository.ApplicationRepository,
	applicationEntityRepo repository.ApplicationEntityRepository,
	applicationApiLogRepo repository.ApplicationApiLogRepository,
	redisClient *redis.Client,

	conf *viper.Viper,
	enforcer *casbin.Enforcer,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	// r.RedirectTrailingSlash = false // Add this line
	// r.RedirectTrailingSlash = false // Disable automatic redirection

	// Create a limiter struct.
	// TODO: Only for sync data APIs? Or all APIs?
	// limiter := tollbooth.NewLimiter(10, nil)
	// limiter.SetBasicAuthExpirationTTL(time.Hour)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.Use(
		middleware.CORSMiddleware(),
		middleware.ResponseLogMiddleware(logger),
		middleware.RequestLogMiddleware(logger),
		// middleware.SignMiddleware(log),
		// middleware.RateLimitMiddleWare(limiter),
	)

	r.StaticFS("/export", http.Dir(conf.GetString("app.runtime-root-path")+conf.GetString("app.export-save-path")))
	r.StaticFS("/uploads", http.Dir("./uploads"))

	// Swagger UI route - Only enable in development/staging environment
	// Security: Disable in production to prevent API information leakage
	if conf.GetString("app.env") != "production" {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API version grouping
	v1 := r.Group("/api/v1")

	handlers := &Handlers{
		Logger:                  logger,
		JWT:                     jwt,
		Conf:                    conf,
		Enforcer:                enforcer,
		Entity:                  entity,
		Approval:                approval,
		ApprovalDefinition:      approvalDefinition,
		ApprovalNode:            approvalNode,
		ApprovalTask:            approvalTask,
		Table:                   table,
		TableField:              tableField,
		Application:             application,
		Webhook:                 webhook,
		WebhookDelivery:         webhookDelivery,
		Cron:                    cron,
		CronLog:                 cronLog,
		Role:                    role,
		Permission:              permission,
		User:                    user,
		Notification:            notification,
		NotificationTemplate:    notificationTemplate,
		NotificationLog:         notificationLog,
		TableApprovalDefinition: tableApprovalDefinition,
		Upload:                  upload,
		TablePermission:         tablePermission,

		// OpenAPI
		OpenApi: openApi,

		// OpenAPI dependencies
		ApplicationRepo:       applicationRepo,
		ApplicationEntityRepo: applicationEntityRepo,
		ApplicationApiLogRepo: applicationApiLogRepo,
		RedisClient:           redisClient,
	}

	// 注册 API 路由
	registerPublicRoutes(v1, handlers)
	registerAdminRoutes(v1, handlers)
	registerUserRoutes(v1, handlers)

	// 注册 OpenAPI 路由 (顶级路由,与 /api 同级)
	// OpenAPI 路由从 Handlers 获取所有依赖,完全自包含
	registerOpenApiRoutes(r, handlers)

	return r
}
