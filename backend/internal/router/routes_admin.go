package router

import (
	"piemdm/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerAdminRoutes(r *gin.RouterGroup, h *Handlers) {
	adminRouter := r.Group("/admin")
	adminRouter.Use(middleware.AdminAuth(h.JWT, h.Logger))
	{
		// 权限相关路由 (Table Permission)
		dataPermissions := adminRouter.Group("/table_permissions")
		{
			dataPermissions.GET("", middleware.CasbinMiddleware(h.Enforcer, "table_permission", "list"), h.TablePermission.List)                   // 获取列表 (支持 ?user_id=X)
			dataPermissions.POST("", middleware.CasbinMiddleware(h.Enforcer, "table_permission", "create"), h.TablePermission.Create)              // 创建权限
			dataPermissions.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "table_permission", "update"), h.TablePermission.BatchUpdate)    // 批量更新 (含冻结解冻)
			dataPermissions.DELETE("/batch", middleware.CasbinMiddleware(h.Enforcer, "table_permission", "delete"), h.TablePermission.BatchDelete) // 批量删除
		}

		// 用户相关路由
		users := adminRouter.Group("/users")
		// users.Use(middleware.StrictAuth(jwt, logger))
		// users.Use(middlewares.Auth()) // User routes require authentication
		{
			// Single resource operations
			users.GET("", middleware.CasbinMiddleware(h.Enforcer, "user", "list"), h.User.List)
			users.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "user", "list"), h.User.Get)
			users.POST("", middleware.CasbinMiddleware(h.Enforcer, "user", "create"), h.User.Create)
			users.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "user", "update"), h.User.Update)
			users.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "user", "delete"), h.User.Delete)

			// Batch operations
			users.POST("/batch", middleware.CasbinMiddleware(h.Enforcer, "user", "create"), h.User.BatchCreate)
			users.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "user", "update"), h.User.BatchUpdate)
			users.DELETE("/batch", middleware.CasbinMiddleware(h.Enforcer, "user", "delete"), h.User.BatchDelete)

			// 角色管理
			users.GET("/:id/roles", middleware.CasbinMiddleware(h.Enforcer, "user", "list"), h.User.GetUserRoles)
			users.PUT("/:id/roles", middleware.CasbinMiddleware(h.Enforcer, "user", "update"), h.User.UpdateUserRoles)
		}

		// 定时任务相关路由
		crons := adminRouter.Group("/crons")
		{
			// Single resource operations
			crons.GET("", middleware.CasbinMiddleware(h.Enforcer, "cron", "list"), h.Cron.List) // 不要使用 "/"
			crons.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "cron", "list"), h.Cron.Get)
			crons.POST("", middleware.CasbinMiddleware(h.Enforcer, "cron", "create"), h.Cron.Create) // 不要使用 "/"
			crons.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "cron", "update"), h.Cron.Update)
			crons.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "cron", "delete"), h.Cron.Delete)

			// Batch operations
			crons.POST("/batch", middleware.CasbinMiddleware(h.Enforcer, "cron", "create"), h.Cron.BatchCreate)
			crons.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "cron", "update"), h.Cron.BatchUpdate)
			crons.DELETE("/batch", middleware.CasbinMiddleware(h.Enforcer, "cron", "delete"), h.Cron.BatchDelete)
		}

		// 定时任务日志相关路由
		cronsLog := adminRouter.Group("/cron_logs")
		{
			// Single resource operations
			cronsLog.GET("", middleware.CasbinMiddleware(h.Enforcer, "cron_log", "list"), h.CronLog.List) // 不要使用 "/"
			cronsLog.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "cron_log", "list"), h.CronLog.Get)
			cronsLog.POST("", middleware.CasbinMiddleware(h.Enforcer, "cron_log", "create"), h.CronLog.Create) // 不要使用 "/"
			cronsLog.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "cron_log", "update"), h.CronLog.Update)
			cronsLog.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "cron_log", "delete"), h.CronLog.Delete)

			// Batch operations
			cronsLog.POST("/batch", middleware.CasbinMiddleware(h.Enforcer, "cron_log", "create"), h.CronLog.BatchCreate)
			cronsLog.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "cron_log", "update"), h.CronLog.BatchUpdate)
			cronsLog.DELETE("/batch", middleware.CasbinMiddleware(h.Enforcer, "cron_log", "delete"), h.CronLog.BatchDelete)
		}

		// 定时任务日志相关路由
		webhooks := adminRouter.Group("/webhooks")
		{
			// Single resource operations
			webhooks.GET("", middleware.CasbinMiddleware(h.Enforcer, "webhook", "list"), h.Webhook.List) // 不要使用 "/"
			webhooks.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "webhook", "list"), h.Webhook.Get)
			webhooks.POST("", middleware.CasbinMiddleware(h.Enforcer, "webhook", "create"), h.Webhook.Create) // 不要使用 "/"
			webhooks.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "webhook", "update"), h.Webhook.Update)
			webhooks.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "webhook", "delete"), h.Webhook.Delete)

			// Batch operations
			webhooks.POST("/batch", middleware.CasbinMiddleware(h.Enforcer, "webhook", "create"), h.Webhook.BatchCreate)
			webhooks.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "webhook", "update"), h.Webhook.BatchUpdate)
			webhooks.DELETE("/batch", middleware.CasbinMiddleware(h.Enforcer, "webhook", "delete"), h.Webhook.BatchDelete)
		}

		// 定时任务日志相关路由
		webhookDeliveries := adminRouter.Group("/webhook_deliveries")
		{
			// Single resource operations
			webhookDeliveries.GET("", middleware.CasbinMiddleware(h.Enforcer, "webhook_delivery", "list"), h.WebhookDelivery.List) // 不要使用 "/"
			webhookDeliveries.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "webhook_delivery", "list"), h.WebhookDelivery.Get)
			webhookDeliveries.POST("", middleware.CasbinMiddleware(h.Enforcer, "webhook_delivery", "create"), h.WebhookDelivery.Create) // 不要使用 "/"
			webhookDeliveries.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "webhook_delivery", "update"), h.WebhookDelivery.Update)
			webhookDeliveries.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "webhook_delivery", "delete"), h.WebhookDelivery.Delete)

			// Batch operations
			webhookDeliveries.POST("/batch", middleware.CasbinMiddleware(h.Enforcer, "webhook_delivery", "create"), h.WebhookDelivery.BatchCreate)
			webhookDeliveries.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "webhook_delivery", "update"), h.WebhookDelivery.BatchUpdate)
			webhookDeliveries.DELETE("/batch", middleware.CasbinMiddleware(h.Enforcer, "webhook_delivery", "delete"), h.WebhookDelivery.BatchDelete)
		}

		// 定时任务日志相关路由
		applications := adminRouter.Group("/applications")
		{
			// Single resource operations
			applications.GET("", middleware.CasbinMiddleware(h.Enforcer, "application", "list"), h.Application.List) // 不要使用 "/"
			applications.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "application", "list"), h.Application.Get)
			applications.POST("", middleware.CasbinMiddleware(h.Enforcer, "application", "create"), h.Application.Create) // 不要使用 "/"
			applications.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "application", "update"), h.Application.Update)
			applications.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "application", "delete"), h.Application.Delete)

			// Batch operations
			applications.POST("/batch", middleware.CasbinMiddleware(h.Enforcer, "application", "create"), h.Application.BatchCreate)
			applications.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "application", "update"), h.Application.BatchUpdate)
			applications.DELETE("/batch", middleware.CasbinMiddleware(h.Enforcer, "application", "delete"), h.Application.BatchDelete)
		}

		// 角色相关路由
		roles := adminRouter.Group("/roles")
		{
			// Single resource operations
			roles.GET("", middleware.CasbinMiddleware(h.Enforcer, "role", "list"), h.Role.List) // 不要使用 "/"
			roles.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "role", "list"), h.Role.Get)
			roles.POST("", middleware.CasbinMiddleware(h.Enforcer, "role", "create"), h.Role.Create) // 不要使用 "/"
			roles.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "role", "update"), h.Role.Update)
			roles.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "role", "delete"), h.Role.Delete)

			// 权限管理接口
			roles.GET("/:id/permissions", middleware.CasbinMiddleware(h.Enforcer, "role", "list"), h.Role.GetRolePermissions)                // 获取角色权限
			roles.POST("/:id/permissions", middleware.CasbinMiddleware(h.Enforcer, "role", "assign_permission"), h.Role.AssignPermissions)   // 分配权限(追加)
			roles.DELETE("/:id/permissions", middleware.CasbinMiddleware(h.Enforcer, "role", "assign_permission"), h.Role.RemovePermissions) // 移除权限
			roles.PUT("/:id/permissions", middleware.CasbinMiddleware(h.Enforcer, "role", "assign_permission"), h.Role.UpdatePermissions)    // 同步权限(覆盖)

			// Batch operations
			roles.POST("/batch", middleware.CasbinMiddleware(h.Enforcer, "role", "create"), h.Role.BatchCreate)
			roles.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "role", "update"), h.Role.BatchUpdate)
			roles.DELETE("/batch", middleware.CasbinMiddleware(h.Enforcer, "role", "delete"), h.Role.BatchDelete)

			// 用户管理
			roles.GET("/:id/users", middleware.CasbinMiddleware(h.Enforcer, "role", "list"), h.Role.GetRoleUsers)
			roles.PUT("/:id/users", middleware.CasbinMiddleware(h.Enforcer, "role", "update"), h.Role.UpdateRoleUsers)
		}

		// 权限相关路由
		permissions := adminRouter.Group("/permissions")
		{
			permissions.GET("", middleware.CasbinMiddleware(h.Enforcer, "permission", "list"), h.Permission.List)            // 获取权限列表(分页)
			permissions.GET("/tree", middleware.CasbinMiddleware(h.Enforcer, "permission", "list"), h.Permission.GetTree)    // 获取权限树
			permissions.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "permission", "list"), h.Permission.Get)         // 获取单个权限
			permissions.POST("", middleware.CasbinMiddleware(h.Enforcer, "permission", "create"), h.Permission.Create)       // 创建权限
			permissions.PUT("", middleware.CasbinMiddleware(h.Enforcer, "permission", "update"), h.Permission.Update)        // 更新权限
			permissions.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "permission", "delete"), h.Permission.Delete) // 删除权限
		}

		// 审批定义相关路由
		approvalDefinitions := adminRouter.Group("/approval_defs")
		{
			// 基础CRUD接口
			approvalDefinitions.GET("", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "list"), h.ApprovalDefinition.List)
			approvalDefinitions.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "list"), h.ApprovalDefinition.Get)
			approvalDefinitions.POST("", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "create"), h.ApprovalDefinition.Create)
			approvalDefinitions.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "update"), h.ApprovalDefinition.Update)
			approvalDefinitions.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "delete"), h.ApprovalDefinition.Delete)

			// Batch operations
			approvalDefinitions.POST("/batch", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "create"), h.ApprovalDefinition.BatchCreate)
			approvalDefinitions.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "update"), h.ApprovalDefinition.BatchUpdate)
			approvalDefinitions.DELETE("/batch", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "delete"), h.ApprovalDefinition.BatchDelete)

			// 业务特有接口
			approvalDefinitions.GET("/code/:code", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "list"), h.ApprovalDefinition.GetByCode)
			approvalDefinitions.GET("/entity/:entityCode", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "list"), h.ApprovalDefinition.GetByEntity)
			approvalDefinitions.GET("/entity/:entityCode/active", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "list"), h.ApprovalDefinition.GetActiveByEntity)
			approvalDefinitions.POST("/:id/activate", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "update"), h.ApprovalDefinition.Activate)
			approvalDefinitions.POST("/:id/deactivate", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "update"), h.ApprovalDefinition.Deactivate)
			approvalDefinitions.POST("/:id/publish", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "update"), h.ApprovalDefinition.Publish)
			approvalDefinitions.GET("/code/:code/versions", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "list"), h.ApprovalDefinition.GetVersions)
			approvalDefinitions.POST("/:id/versions", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "create"), h.ApprovalDefinition.CreateVersion)
			approvalDefinitions.POST("/:id/validate", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "list"), h.ApprovalDefinition.Validate)
			approvalDefinitions.POST("/sync-feishu", middleware.CasbinMiddleware(h.Enforcer, "approval_def", "update"), h.ApprovalDefinition.SyncFeishu) // sync feishu
		}

		// 审批节点相关路由
		approvalNodes := adminRouter.Group("/approval_nodes")
		{
			// 基础CRUD接口
			approvalNodes.GET("", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "list"), h.ApprovalNode.List)
			approvalNodes.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "list"), h.ApprovalNode.Get)
			approvalNodes.POST("", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "create"), h.ApprovalNode.Create)
			approvalNodes.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "update"), h.ApprovalNode.Update)
			approvalNodes.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "delete"), h.ApprovalNode.Delete)

			// Batch operations
			approvalNodes.POST("/batch", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "create"), h.ApprovalNode.BatchCreate)
			approvalNodes.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "update"), h.ApprovalNode.BatchUpdate)
			approvalNodes.POST("/batch_delete", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "delete"), h.ApprovalNode.BatchDelete)

			// 业务特有接口
			approvalNodes.GET("/approval_def/:approvalDefCode", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "list"), h.ApprovalNode.GetNodesByApprovalDef)
			approvalNodes.GET("/approval_def/:approvalDefCode/start", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "list"), h.ApprovalNode.GetStartNode)
			approvalNodes.GET("/approval_def/:approvalDefCode/end", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "list"), h.ApprovalNode.GetEndNodes)
			approvalNodes.GET("/approval_def/:approvalDefCode/next/:currentNodeCode", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "list"), h.ApprovalNode.GetNextNodes)
			approvalNodes.POST("/approval_def/:approvalDefCode/validate", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "list"), h.ApprovalNode.ValidateWorkflow)
			approvalNodes.POST("/:id/configure-approvers", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "update"), h.ApprovalNode.ConfigureApprovers)
			approvalNodes.POST("/:id/configure-conditions", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "update"), h.ApprovalNode.ConfigureConditions)
			approvalNodes.POST("/:id/configure-timeouts", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "update"), h.ApprovalNode.ConfigureTimeouts)
			approvalNodes.POST("/:id/activate", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "update"), h.ApprovalNode.ActivateNode)
			approvalNodes.POST("/:id/deactivate", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "update"), h.ApprovalNode.DeactivateNode)
			approvalNodes.POST("/sync", middleware.CasbinMiddleware(h.Enforcer, "approval_node", "create"), h.ApprovalNode.BatchSyncApprovalNodes)
		}

		// 审批任务相关路由
		approvalTasks := adminRouter.Group("/approval_tasks")
		{
			// 基础CRUD接口
			approvalTasks.GET("", middleware.CasbinMiddleware(h.Enforcer, "approval_task", "list"), h.ApprovalTask.List)
			approvalTasks.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "approval_task", "list"), h.ApprovalTask.Get)
			approvalTasks.POST("", middleware.CasbinMiddleware(h.Enforcer, "approval_task", "create"), h.ApprovalTask.Create)
			approvalTasks.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "approval_task", "update"), h.ApprovalTask.Update)
			approvalTasks.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "approval_task", "delete"), h.ApprovalTask.Delete)

			// Batch operations
			approvalTasks.POST("/batch", middleware.CasbinMiddleware(h.Enforcer, "approval_task", "create"), h.ApprovalTask.BatchCreate)
			approvalTasks.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "approval_task", "update"), h.ApprovalTask.BatchUpdate)
			approvalTasks.POST("/batch_delete", middleware.CasbinMiddleware(h.Enforcer, "approval_task", "delete"), h.ApprovalTask.BatchDelete)

			// 任务处理业务接口
			// approvalTasks.POST("/process", approvalTask.ProcessTask)
			// approvalTasks.POST("/:id/approve", approvalTask.ApproveTask)
			// approvalTasks.POST("/:id/reject", approvalTask.RejectTask)
			// approvalTasks.POST("/:id/transfer", approvalTask.TransferTask)
			// approvalTasks.POST("/:id/remind", approvalTask.RemindTask)
			// approvalTasks.POST("/batch-remind", approvalTask.BatchRemindTasks)
			// approvalTasks.GET("/assignee/:assigneeId", approvalTask.GetTasksByAssignee)
			// approvalTasks.GET("/assignee/:assigneeId/pending", approvalTask.GetPendingTasksByAssignee)
			// approvalTasks.GET("/approval/:approvalCode", approvalTask.GetTasksByApproval)
			// approvalTasks.GET("/overdue", approvalTask.GetOverdueTasks)
			// approvalTasks.GET("/expired", approvalTask.GetExpiredTasks)
			// approvalTasks.GET("/statistics", approvalTask.GetTaskStatistics)
		}

		// 审批实例相关路由
		approvals := adminRouter.Group("/approvals")
		{
			// 基础CRUD接口
			approvals.GET("", middleware.CasbinMiddleware(h.Enforcer, "approval", "list"), h.Approval.List)
			approvals.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "approval", "list"), h.Approval.Get)
			approvals.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "approval", "delete"), h.Approval.Delete)

			// Batch operations
			approvals.POST("/batch_delete", middleware.CasbinMiddleware(h.Enforcer, "approval", "delete"), h.Approval.BatchDelete)

			// 审批流程业务接口
			// approvals.POST("/start", approval.StartApproval)
			// approvals.POST("/submit", approval.SubmitApproval)
			// approvals.POST("/cancel", approval.CancelApproval)
			// approvals.POST("/process", approval.ProcessApproval)
			// approvals.GET("/code/:code", approval.GetApprovalByCode)
			// approvals.GET("/applicant/:applicantId", approval.GetApprovalsByApplicant)
			// approvals.GET("/status/:status", approval.GetApprovalsByStatus)
			// approvals.GET("/entity/:entityCode", approval.GetApprovalsByEntity)
			approvals.GET("/statistics", h.Approval.GetStatistics)
			// approvals.GET("/expired", approval.GetExpiredApprovals)
			// approvals.GET("/:id/history", approval.GetApprovalHistory)

			// 保留旧的任务审批接口以兼容前台
			// approvals.POST("/task/:id/approve", approval.TaskApprove)
			// approvals.POST("/task/:id/reject", approval.TaskReject)

		}

		// 通知模板路由
		notificationTemplates := adminRouter.Group("/notification_templates")
		{
			notificationTemplates.GET("", middleware.CasbinMiddleware(h.Enforcer, "notification_template", "list"), h.NotificationTemplate.List)
			notificationTemplates.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "notification_template", "list"), h.NotificationTemplate.Get)
			notificationTemplates.POST("", middleware.CasbinMiddleware(h.Enforcer, "notification_template", "create"), h.NotificationTemplate.Create)
			notificationTemplates.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "notification_template", "update"), h.NotificationTemplate.Update)
			notificationTemplates.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "notification_template", "delete"), h.NotificationTemplate.Delete)
			notificationTemplates.POST("/:id/render", middleware.CasbinMiddleware(h.Enforcer, "notification_template", "list"), h.NotificationTemplate.Render)
			notificationTemplates.POST("/validate", middleware.CasbinMiddleware(h.Enforcer, "notification_template", "list"), h.NotificationTemplate.Validate)
			notificationTemplates.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "notification_template", "update"), h.NotificationTemplate.BatchUpdate)
		}

		// 通知日志路由
		notificationLogs := adminRouter.Group("/notification_logs")
		{
			notificationLogs.GET("", middleware.CasbinMiddleware(h.Enforcer, "notification_log", "list"), h.NotificationLog.List)
			notificationLogs.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "notification_log", "list"), h.NotificationLog.Get)
		}

		// 通知相关路由
		notifications := adminRouter.Group("/notifications")
		{
			// 通知发送
			notifications.POST("/send/approval", middleware.CasbinMiddleware(h.Enforcer, "notification", "create"), h.Notification.SendApproval)
			notifications.POST("/send/batch", middleware.CasbinMiddleware(h.Enforcer, "notification", "create"), h.Notification.SendBatch)
			notifications.POST("/send/template", middleware.CasbinMiddleware(h.Enforcer, "notification", "create"), h.Notification.SendByTemplate)
			notifications.POST("/test", middleware.CasbinMiddleware(h.Enforcer, "notification", "create"), h.Notification.Test)

			// 统计
			notifications.GET("/statistics", middleware.CasbinMiddleware(h.Enforcer, "notification", "list"), h.Notification.GetStatistics)

			// 通知处理
			notifications.POST("/process/pending", middleware.CasbinMiddleware(h.Enforcer, "notification", "create"), h.Notification.ProcessPending)
			notifications.POST("/process/retry", middleware.CasbinMiddleware(h.Enforcer, "notification", "create"), h.Notification.ProcessRetry)
		}

		// 表相关路由
		tables := adminRouter.Group("/tables")
		{
			// Single resource operations
			tables.GET("", middleware.CasbinMiddleware(h.Enforcer, "table", "list"), h.Table.List) // 不要使用 "/"
			tables.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "table", "list"), h.Table.Get)
			tables.POST("", middleware.CasbinMiddleware(h.Enforcer, "table", "create"), h.Table.Create) // 不要使用 "/"
			tables.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "table", "update"), h.Table.Update)
			tables.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "table", "delete"), h.Table.Delete)

			// Batch operations
			tables.POST("/batch", middleware.CasbinMiddleware(h.Enforcer, "table", "create"), h.Table.BatchCreate)
			tables.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "table", "update"), h.Table.BatchUpdate)
			tables.POST("/batch_delete", middleware.CasbinMiddleware(h.Enforcer, "table", "delete"), h.Table.BatchDelete)
		}

		// 表字段相关路由
		tableFields := adminRouter.Group("/table_fields")
		{
			// Single resource operations
			tableFields.GET("", middleware.CasbinMiddleware(h.Enforcer, "table_field", "list"), h.TableField.List)                  // 不要使用 "/"
			tableFields.GET("/fields", middleware.CasbinMiddleware(h.Enforcer, "table_field", "list"), h.TableField.GetTableFields) // 获取所有字段(包括系统字段) - 必须在/:id之前
			tableFields.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "table_field", "list"), h.TableField.Get)
			tableFields.POST("", middleware.CasbinMiddleware(h.Enforcer, "table_field", "create"), h.TableField.Create) // 不要使用 "/"
			tableFields.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "table_field", "update"), h.TableField.Update)
			tableFields.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "table_field", "delete"), h.TableField.Delete)
			tableFields.POST("/public", middleware.CasbinMiddleware(h.Enforcer, "table_field", "update"), h.TableField.Public)

			// Batch operations
			tableFields.POST("/batch", middleware.CasbinMiddleware(h.Enforcer, "table_field", "create"), h.TableField.BatchCreate)
			tableFields.PUT("/batch", middleware.CasbinMiddleware(h.Enforcer, "table_field", "update"), h.TableField.BatchUpdate)
			tableFields.POST("/batch_delete", middleware.CasbinMiddleware(h.Enforcer, "table_field", "delete"), h.TableField.BatchDelete)
		}

		// 表审批定义相关路由
		tableApprovalDefinitons := adminRouter.Group("/table_approval_defs")
		{
			// 基础CRUD接口
			tableApprovalDefinitons.GET("", middleware.CasbinMiddleware(h.Enforcer, "table_approval_def", "list"), h.TableApprovalDefinition.List)
			tableApprovalDefinitons.GET("/:id", middleware.CasbinMiddleware(h.Enforcer, "table_approval_def", "list"), h.TableApprovalDefinition.Get)
			tableApprovalDefinitons.POST("", middleware.CasbinMiddleware(h.Enforcer, "table_approval_def", "create"), h.TableApprovalDefinition.Create)
			tableApprovalDefinitons.PUT("/:id", middleware.CasbinMiddleware(h.Enforcer, "table_approval_def", "update"), h.TableApprovalDefinition.Update)
			tableApprovalDefinitons.DELETE("/:id", middleware.CasbinMiddleware(h.Enforcer, "table_approval_def", "delete"), h.TableApprovalDefinition.Delete)

			// Batch operations
			tableApprovalDefinitons.POST("/batch_create", middleware.CasbinMiddleware(h.Enforcer, "table_approval_def", "create"), h.TableApprovalDefinition.BatchCreate)
			tableApprovalDefinitons.DELETE("/batch_delete", middleware.CasbinMiddleware(h.Enforcer, "table_approval_def", "delete"), h.TableApprovalDefinition.BatchDelete)
		}

		// 实体统计相关路由
		entities := adminRouter.Group("/entities")
		{
			entities.GET("/statistics", h.Entity.GetStatistics)
		}
	}
}
