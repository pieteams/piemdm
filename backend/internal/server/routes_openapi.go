package server

import (
	"piemdm/internal/middleware"
	"piemdm/internal/service"

	"github.com/gin-gonic/gin"
)

func registerOpenApiRoutes(r *gin.Engine, h *Handlers) {
	// 创建 OpenAPI 专用的 service (从 Handlers 获取所有依赖)
	openApiAuthService := service.NewOpenApiAuthService(
		h.Logger,
		h.ApplicationRepo,
		h.ApplicationEntityRepo,
		h.RedisClient,
		h.Conf,
	)
	svc := service.NewService(h.Logger, nil, nil)
	applicationApiLogService := service.NewApplicationApiLogService(svc, h.ApplicationApiLogRepo)

	// OpenAPI 路由组 - 与 /api 同级
	openapiV1 := r.Group("/openapi/v1")

	// 应用中间件
	openapiV1.Use(
		middleware.CanonicalRequestMiddleware(openApiAuthService, h.Logger, h.Conf),
		middleware.OpenApiAuditMiddleware(applicationApiLogService, h.Logger),
	)

	// Entity 相关路由
	entities := openapiV1.Group("/entities")
	entities.Use(middleware.OpenApiEntityPermissionMiddleware(openApiAuthService, h.Logger))

	// Base CRUD 操作
	entities.GET("/:table", h.OpenApi.List)    // 列表查询
	entities.GET("/:table/:id", h.OpenApi.Get) // 详情查询
	// entities.POST("/:table", h.OpenApiHandler.Create)    // 创建 (Phase 2)
	// entities.PUT("/:table/:id", h.OpenApiHandler.Update) // 更新 (Phase 2)
	// entities.DELETE("/:table/:id", h.OpenApiHandler.Delete) // 删除 (Phase 2)
}
