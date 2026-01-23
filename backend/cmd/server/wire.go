//go:build wireinject
// +build wireinject

package main

import (
	"piemdm/internal/handler"
	"piemdm/internal/auth/casbin" // Import casbin package
	"piemdm/internal/transaction"
	"piemdm/internal/repository"
	"piemdm/internal/server"
	"piemdm/internal/service"
	"piemdm/pkg/cron"
	"piemdm/pkg/cron/job"
	"piemdm/pkg/helper/sid"
	"piemdm/pkg/jwt"
	"piemdm/pkg/log"
	"piemdm/pkg/notification"
	"piemdm/pkg/webhook"
	"piemdm/pkg/webhook/task"

	"github.com/google/wire"
	"github.com/spf13/viper"
)

var HandlerSet = wire.NewSet(
	handler.NewHandler,
	handler.NewUserHandler,
	handler.NewApprovalHandler,
	handler.NewApprovalDefinitionHandler,
	handler.NewApprovalNodeHandler,
	handler.NewApprovalTaskHandler,
	handler.NewTableHandler,
	handler.NewTableFieldHandler,

	handler.NewApplicationHandler,
	handler.NewWebhookHandler,
	handler.NewWebhookDeliveryHandler,
	handler.NewCronHandler,
	handler.NewCronLogHandler,
	handler.NewEntityHandler,
	handler.NewRoleHandler,
	handler.NewPermissionHandler,
	handler.NewNotificationHandler,
	handler.NewNotificationTemplateHandler,
	handler.NewNotificationLogHandler,
	handler.NewTableApprovalDefinitionHandler,
	handler.NewUploadHandler,
	handler.NewTablePermissionHandler,

	// OpenAPI
	handler.NewOpenApiHandler,
)

var ServiceSet = wire.NewSet(
	service.NewService,
	service.NewUserService,
	service.NewApprovalService,
	service.NewApprovalDefinitionService,
	service.NewApprovalNodeService,
	service.NewApprovalTaskService,
	service.NewTableService,
	service.NewTableFieldService,

	service.NewApplicationService,
	service.NewWebhookService,
	service.NewWebhookDeliveryService,
	service.NewCronService,
	service.NewCronLogService,
	service.NewCronParamService,
	service.NewEntityService,
	service.NewEntityLogService,
	service.NewGlobalIdService,
	service.NewRoleService,
	service.NewPermissionService,
	service.NewNotificationService,
	service.NewNotificationTemplateService,
	service.NewNotificationLogService,
	service.NewTableApprovalDefinitionService,
	service.NewAutocodeService,
	service.NewUploadService,
	service.NewTablePermissionService,

	// OpenAPI
	service.NewOpenApiAuthService,
	service.NewApplicationApiLogService,
)

var RepositorySet = wire.NewSet(
	repository.NewDB,
	repository.NewRedis,
	repository.NewRepository,
	repository.NewBaseRepository,
	repository.NewUserRepository,
	repository.NewApprovalRepository,
	repository.NewApprovalDefinitionRepository,
	repository.NewApprovalNodeRepository,
	repository.NewApprovalTaskRepository,
	repository.NewTableRepository,
	repository.NewTableFieldRepository,

	repository.NewApplicationRepository,
	repository.NewWebhookRepository,
	repository.NewWebhookDeliveryRepository,
	repository.NewCronRepository,
	repository.NewCronParamRepository,
	repository.NewCronLogRepository,
	repository.NewEntityRepository,
	repository.NewEntityLogRepository,
	repository.NewGlobalIdRepository,
	repository.NewRoleRepository,
	repository.NewPermissionRepository,
	repository.NewNotificationTemplateRepository,
	repository.NewNotificationLogRepository,
	repository.NewTableApprovalDefinitionRepository,
	repository.NewTablePermissionRepository,
	repository.NewUserRoleRepository,

	// OpenAPI
	repository.NewApplicationApiLogRepository,
	repository.NewApplicationEntityRepository,
)

var CasbinSet = wire.NewSet(
	casbin.InitEnforcer,
)

var CronSet = wire.NewSet(
	job.NewScanner,
	cron.NewCron,
)

var WebhookSet = wire.NewSet(
	task.NewScanner,
	webhook.NewWebhook,
)

var TransactionSet = wire.NewSet(
	transaction.NewTransactionManager,
)

var NotificationSet = wire.NewSet(
	notification.NewNotificationServiceProvider,
)

func newApp(*viper.Viper, *log.Logger) (*server.Server, func(), error) {
	panic(wire.Build(
		RepositorySet,
		ServiceSet,
		HandlerSet,
		TransactionSet,
		NotificationSet,
		CasbinSet,
		server.NewServer,
		server.NewServerHTTP,
		sid.NewSid,
		jwt.NewJwt,
	))
}

func newCronApp(*viper.Viper, *log.Logger) (*cron.Cron, func(), error) {
	panic(wire.Build(
		RepositorySet,
		ServiceSet,
		CronSet,
		NotificationSet,
		sid.NewSid,
		jwt.NewJwt,
	))
}

func newWebhookApp(*viper.Viper, *log.Logger) (*webhook.Webhook, func(), error) {
	panic(wire.Build(
		RepositorySet,
		ServiceSet,
		WebhookSet,
		NotificationSet,
		sid.NewSid,
		jwt.NewJwt,
	))
}
