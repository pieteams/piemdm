package router

import (
	"piemdm/internal/handler"
	"piemdm/internal/repository"
	"piemdm/internal/service"
	"piemdm/pkg/jwt"
	"piemdm/pkg/log"

	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type Handlers struct {
	Logger             *log.Logger
	JWT                *jwt.JWT
	Conf               *viper.Viper
	Enforcer           *casbin.Enforcer
	Entity             handler.EntityHandler
	Approval           handler.ApprovalHandler
	ApprovalService    service.ApprovalService
	ApprovalDefinition handler.ApprovalDefinitionHandler
	ApprovalNode       handler.ApprovalNodeHandler
	ApprovalTask       handler.ApprovalTaskHandler
	Table              handler.TableHandler
	TableField         handler.TableFieldHandler

	Application             handler.ApplicationHandler
	Webhook                 handler.WebhookHandler
	WebhookDelivery         handler.WebhookDeliveryHandler
	Cron                    handler.CronHandler
	CronLog                 handler.CronLogHandler
	Role                    handler.RoleHandler
	Permission              handler.PermissionHandler
	User                    handler.UserHandler
	Notification            handler.NotificationHandler
	NotificationTemplate    handler.NotificationTemplateHandler
	NotificationLog         handler.NotificationLogHandler
	TableApprovalDefinition handler.TableApprovalDefinitionHandler
	Upload                  handler.UploadHandler
	TablePermission         handler.TablePermissionHandler // 新增

	// OpenAPI Handler
	OpenApi handler.OpenApiHandler

	// OpenAPI dependencies (只在 OpenAPI 路由中使用)
	ApplicationRepo       repository.ApplicationRepository
	ApplicationEntityRepo repository.ApplicationEntityRepository
	ApplicationApiLogRepo repository.ApplicationApiLogRepository
	RedisClient           *redis.Client
}
