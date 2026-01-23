package handler

import (
	"net/http"

	"piemdm/internal/transaction"
	"piemdm/internal/repository"
	"piemdm/internal/service"
	"piemdm/pkg/helper/resp"

	"github.com/gin-gonic/gin"
)

// NotificationTemplateHandler 通知模板处理器接口
type NotificationTemplateHandler interface {
	// Base CRUD
	List(ctx *gin.Context)
	Get(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)

	// 特殊操作
	Render(ctx *gin.Context)
	Validate(ctx *gin.Context)
	BatchUpdate(ctx *gin.Context)
}

// notificationTemplateHandler 通知模板处理器实现
type notificationTemplateHandler struct {
	*Handler
	notificationTemplateService service.NotificationTemplateService
	transactionManager          transaction.TransactionManager
}

// NewNotificationTemplateHandler 创建通知模板处理器
func NewNotificationTemplateHandler(
	handler *Handler,
	notificationTemplateService service.NotificationTemplateService,
	transactionManager transaction.TransactionManager,
) NotificationTemplateHandler {
	return &notificationTemplateHandler{
		Handler:                     handler,
		notificationTemplateService: notificationTemplateService,
		transactionManager:          transactionManager,
	}
}

// List 获取通知模板列表
// @Summary 获取通知模板列表
// @Description 分页获取通知模板列表
// @Tags 通知模板管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(15)
// @Param templateType query string false "模板类型"
// @Param notificationType query string false "通知类型"
// @Param keyword query string false "关键词搜索"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/v1/notification_templates [get]
func (h *notificationTemplateHandler) List(ctx *gin.Context) {
	page, pageSize := GetPage(ctx)

	req := &repository.ListNotificationTemplateRequest{
		Page:             page,
		PageSize:         pageSize,
		TemplateType:     ctx.Query("templateType"),
		NotificationType: ctx.Query("notificationType"),
		TemplateCode:     ctx.Query("templateCode"),
		TemplateName:     ctx.Query("templateName"),
		Keyword:          ctx.Query("keyword"),
	}

	result, err := h.notificationTemplateService.List(ctx.Request.Context(), req)
	if err != nil {
		h.logger.Error("获取通知模板列表失败", "error", err)
		resp.HandleError(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// 生成分页链接
	links := resp.GeneratePaginationLinks(ctx.Request, result.Page, result.PageSize, int(result.Total))
	ctx.Header("Link", links.String())

	resp.HandleSuccess(ctx, result.Data)
}

// Get 获取通知模板详情
// @Summary 获取通知模板详情
// @Description 根据ID获取通知模板详情
// @Tags 通知模板管理
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 200 {object} piemdm_internal_model.NotificationTemplate
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notification_templates/{id} [get]
func (h *notificationTemplateHandler) Get(ctx *gin.Context) {
	templateID := ctx.Param("id")
	if templateID == "" {
		resp.HandleError(ctx, http.StatusBadRequest, "模板ID不能为空", nil)
		return
	}

	template, err := h.notificationTemplateService.Get(ctx.Request.Context(), templateID)
	if err != nil {
		h.logger.Error("获取通知模板失败", "error", err)
		resp.HandleError(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(ctx, template)
}

// Create 创建通知模板
// @Summary 创建通知模板
// @Description 创建新的通知模板
// @Tags 通知模板管理
// @Accept json
// @Produce json
// @Param template body service.CreateNotificationTemplateRequest true "模板信息"
// @Success 200 {object} piemdm_internal_model.NotificationTemplate
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notification_templates [post]
func (h *notificationTemplateHandler) Create(ctx *gin.Context) {
	var req service.CreateNotificationTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数绑定失败", "error", err)
		resp.HandleError(ctx, http.StatusBadRequest, "参数格式错误", nil)
		return
	}

	// 在事务中执行
	err := h.transactionManager.ExecuteInTransaction(ctx.Request.Context(), func(tx transaction.Transaction) error {
		// 将事务传递给服务层
		txCtx := transaction.WithTransaction(ctx.Request.Context(), tx)

		template, err := h.notificationTemplateService.Create(txCtx, &req)
		if err != nil {
			return err
		}

		resp.HandleSuccess(ctx, template)
		return nil
	})
	if err != nil {
		h.logger.Error("创建通知模板失败", "error", err)
		resp.HandleError(ctx, http.StatusInternalServerError, err.Error(), nil)
	}
}

// Update 更新通知模板
// @Summary 更新通知模板
// @Description 更新指定的通知模板
// @Tags 通知模板管理
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Param template body service.UpdateNotificationTemplateRequest true "模板信息"
// @Success 200 {object} piemdm_internal_model.NotificationTemplate
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notification_templates/{id} [put]
func (h *notificationTemplateHandler) Update(ctx *gin.Context) {
	templateID := ctx.Param("id")
	if templateID == "" {
		resp.HandleError(ctx, http.StatusBadRequest, "模板ID不能为空", nil)
		return
	}

	var req service.UpdateNotificationTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数绑定失败", "error", err)
		resp.HandleError(ctx, http.StatusBadRequest, "参数格式错误", nil)
		return
	}

	// 设置ID
	req.ID = templateID

	// 在事务中执行
	err := h.transactionManager.ExecuteInTransaction(ctx.Request.Context(), func(tx transaction.Transaction) error {
		txCtx := transaction.WithTransaction(ctx.Request.Context(), tx)

		template, err := h.notificationTemplateService.Update(txCtx, &req)
		if err != nil {
			return err
		}

		resp.HandleSuccess(ctx, template)
		return nil
	})
	if err != nil {
		h.logger.Error("更新通知模板失败", "error", err)
		resp.HandleError(ctx, http.StatusInternalServerError, err.Error(), nil)
	}
}

// BatchUpdate 批量更新通知模板状态
// @Summary 批量更新通知模板状态
// @Tags 通知模板管理
// @Accept json
// @Produce json
// @Param data body object true "批量更新请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/notification_templates/batch [put]
func (h *notificationTemplateHandler) BatchUpdate(ctx *gin.Context) {
	var params struct {
		Ids    []string `json:"ids"`
		Status string   `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&params); err != nil {
		h.logger.Error("参数绑定失败", "error", err)
		resp.HandleError(ctx, http.StatusBadRequest, "参数格式错误", nil)
		return
	}

	// 在事务中执行
	err := h.transactionManager.ExecuteInTransaction(ctx.Request.Context(), func(tx transaction.Transaction) error {
		txCtx := transaction.WithTransaction(ctx.Request.Context(), tx)
		return h.notificationTemplateService.BatchUpdateStatus(txCtx, params.Ids, params.Status)
	})
	if err != nil {
		h.logger.Error("批量更新通知模板状态失败", "error", err)
		resp.HandleError(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(ctx, map[string]string{"message": "更新成功"})
}

// Delete 删除通知模板

// @Description 删除指定的通知模板
// @Tags 通知模板管理
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notification_templates/{id} [delete]
func (h *notificationTemplateHandler) Delete(ctx *gin.Context) {
	templateID := ctx.Param("id")
	if templateID == "" {
		resp.HandleError(ctx, http.StatusBadRequest, "模板ID不能为空", nil)
		return
	}

	// 在事务中执行
	err := h.transactionManager.ExecuteInTransaction(ctx.Request.Context(), func(tx transaction.Transaction) error {
		txCtx := transaction.WithTransaction(ctx.Request.Context(), tx)

		return h.notificationTemplateService.Delete(txCtx, templateID)
	})
	if err != nil {
		h.logger.Error("删除通知模板失败", "error", err)
		resp.HandleError(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(ctx, map[string]string{"message": "删除成功"})
}

// Render 渲染通知模板
// @Summary 渲染通知模板
// @Description 使用变量渲染通知模板
// @Tags 通知模板管理
// @Accept json
// @Produce json
// @Param id path string true "模板ID"
// @Param variables body map[string]any true "模板变量"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notification_templates/{id}/render [post]
func (h *notificationTemplateHandler) Render(ctx *gin.Context) {
	templateID := ctx.Param("id")
	if templateID == "" {
		resp.HandleError(ctx, http.StatusBadRequest, "模板ID不能为空", nil)
		return
	}

	var variables map[string]any
	if err := ctx.ShouldBindJSON(&variables); err != nil {
		h.logger.Error("参数绑定失败", "error", err)
		resp.HandleError(ctx, http.StatusBadRequest, "参数格式错误", nil)
		return
	}

	result, err := h.notificationTemplateService.RenderTemplate(ctx.Request.Context(), templateID, variables)
	if err != nil {
		h.logger.Error("渲染通知模板失败", "error", err)
		resp.HandleError(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(ctx, result)
}

// Validate 验证通知模板
// @Summary 验证通知模板
// @Description 验证通知模板的语法和格式
// @Tags 通知模板管理
// @Accept json
// @Produce json
// @Param template body service.ValidateTemplateRequest true "模板内容"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/notification_templates/validate [post]
func (h *notificationTemplateHandler) Validate(ctx *gin.Context) {
	var req service.ValidateTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数绑定失败", "error", err)
		resp.HandleError(ctx, http.StatusBadRequest, "参数格式错误", nil)
		return
	}

	err := h.notificationTemplateService.ValidateTemplate(ctx.Request.Context(), &req)
	if err != nil {
		h.logger.Error("验证通知模板失败", "error", err)
		resp.HandleError(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	resp.HandleSuccess(ctx, map[string]bool{"valid": true})
}
