package server

import (
	"piemdm/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(r *gin.RouterGroup, h *Handlers) {
	userRouter := r.Group("")
	userRouter.Use(middleware.UserAuth(h.JWT, h.Logger))
	{
		// 文件上传
		userRouter.POST("/upload", h.Upload.Upload)

		// 表字段相关路由
		entities := userRouter.Group("/entities")
		{
			// Single resource operations
			entities.GET("/:table_code", h.Entity.List)
			entities.GET("/:table_code/logs", h.Entity.ListEntityLogs)
			entities.GET("/:table_code/histories", h.Entity.ListEntityHistories)
			entities.GET("/:table_code/:id", h.Entity.Get)
			entities.POST("/:table_code", h.Entity.Create)
			entities.PUT("/:table_code/:id", h.Entity.Update)
			entities.DELETE("/:table_code/:id", h.Entity.Delete)

			// Batch operations
			entities.POST("/:table_code/batch", h.Entity.BatchCreate)
			entities.PUT("/:table_code/batch", h.Entity.BatchUpdate)
			entities.POST("/:table_code/batch_delete", h.Entity.BatchDelete)

			// entity
			entities.POST("/:table_code/import", h.Entity.Import)
			entities.GET("/:table_code/export", h.Entity.Export)
			entities.GET("/:table_code/template", h.Entity.Template)
		}

		// 工作流相关路由
		tables := userRouter.Group("/tables")
		{
			tables.GET("", h.Table.List) // 不要使用 "/"
			tables.GET("/:id", h.Table.Get)
		}

		// 表字段相关路由
		tableFields := userRouter.Group("/table_fields")
		{
			tableFields.GET("", h.TableField.List)                  // 不要使用 "/"
			tableFields.GET("/fields", h.TableField.GetTableFields) // 获取所有字段(包括系统字段) - 必须在/:id之前
			tableFields.GET("/:id", h.TableField.Get)
		}

		// 审批节点相关路由
		approvalNodes := userRouter.Group("/approval_nodes")
		{
			approvalNodes.GET("", h.ApprovalNode.List)
			approvalNodes.GET("/:id", h.ApprovalNode.Get)
		}

		// 审批任务相关路由
		approvalTasks := userRouter.Group("/approval_tasks")
		{
			approvalTasks.GET("", h.ApprovalTask.List)
			approvalTasks.GET("/:id", h.ApprovalTask.Get)
		}

		// 工作流相关路由
		approvals := userRouter.Group("/approvals")
		{
			approvals.GET("", h.Approval.List)
			approvals.GET("/:id", h.Approval.Get)
			approvals.POST("/task/:id/approve", h.Approval.ApproveTask)
			approvals.POST("/task/:id/reject", h.Approval.RejectTask)
		}
		// 工作流相关路由
		hookDeliveries := userRouter.Group("/webhook_deliveries")
		{
			hookDeliveries.GET("", h.WebhookDelivery.List)
		}
	}
}
