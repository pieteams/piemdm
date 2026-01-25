package handler

import (
	"net/http"

	"piemdm/internal/model"
	"piemdm/internal/service"
	"piemdm/pkg/helper/resp"

	"github.com/gin-gonic/gin"
)

type WebhookHandler interface {
	// Base CRUD
	List(c *gin.Context)
	Get(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)

	// Batch operations
	BatchCreate(c *gin.Context)
	BatchUpdate(c *gin.Context)
	BatchDelete(c *gin.Context)
}

type webhookHandler struct {
	*Handler
	webhookService service.WebhookService
}

func NewWebhookHandler(handler *Handler, webhookService service.WebhookService) WebhookHandler {
	return &webhookHandler{
		Handler:        handler,
		webhookService: webhookService,
	}
}

// List 获取Webhook列表
// @Summary 获取Webhook列表
// @Tags Webhook管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(15)
// @Param tableCode query string false "表代码"
// @Param startDate query string false "开始日期"
// @Param endDate query string false "结束日期"
// @Success 200 {array} model.Webhook
// @Router /admin/webhooks [get]
func (h *webhookHandler) List(c *gin.Context) {
	var req struct {
		Page      int    `form:"page,default=1"`
		PageSize  int    `form:"pageSize,default=15"`
		TableCode string `form:"tableCode"`
		StartDate string `form:"startDate"`
		EndDate   string `form:"endDate"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	page := req.Page
	pageSize := req.PageSize

	where := make(map[string]any)
	var total int64

	if req.TableCode != "" {
		where["table_code"] = req.TableCode
	}
	if req.StartDate != "" {
		where["created_at >="] = req.StartDate
	}
	if req.EndDate != "" {
		where["created_at <="] = req.EndDate
	}

	webhooks, err := h.webhookService.List(page, pageSize, &total, where)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	links := resp.GeneratePaginationLinks(c.Request, page, pageSize, int(total))
	c.Header("Link", links.String())

	resp.HandleSuccess(c, webhooks)
}

// Get 获取Webhook详情
// @Summary 获取Webhook详情
// @Tags Webhook管理
// @Accept json
// @Produce json
// @Param id path int true "WebhookID"
// @Success 200 {object} model.Webhook
// @Router /admin/webhooks/{id} [get]
func (h *webhookHandler) Get(c *gin.Context) {
	var req struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	webhook, err := h.webhookService.Get(req.Id)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, webhook)
}

// Create 创建Webhook
// @Summary 创建Webhook
// @Tags Webhook管理
// @Accept json
// @Produce json
// @Param data body object true "Webhook信息"
// @Success 200 {object} model.Webhook
// @Router /admin/webhooks [post]
func (h *webhookHandler) Create(c *gin.Context) {
	var req struct {
		Url         string `binding:"required,max=256"` // 调用Url
		TableCode   string `binding:"required,max=64"`  // 主数据Code
		Username    string `binding:"max=64"`           // 用户名
		ContentType string `binding:"required,max=128"` // 系统名称
		Secret      string `binding:"required,max=128"` // AppSecret
		Events      string `binding:"required,max=256"` // 监控什么事件
		Description string `binding:"max=255"`          // 描述
		Status      string `binding:"required,max=128"` // 状态
	}

	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	webhook := model.Webhook{
		Url:         req.Url,
		TableCode:   req.TableCode,
		Username:    req.Username,
		ContentType: req.ContentType,
		Secret:      req.Secret,
		Events:      req.Events,
		Description: req.Description,
		Status:      req.Status,
	}

	err := h.webhookService.Create(c, &webhook)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, webhook)
}

// Update 更新Webhook
// @Summary 更新Webhook
// @Tags Webhook管理
// @Accept json
// @Produce json
// @Param id path int true "WebhookID"
// @Param data body object true "Webhook信息"
// @Success 200 {object} map[string]interface{}
// @Router /admin/webhooks/{id} [put]
func (h *webhookHandler) Update(c *gin.Context) {
	var req struct {
		ID          uint   `binding:"required"`
		Url         string `binding:"required,max=256"` // 调用Url
		TableCode   string `binding:"required,max=64"`  // 主数据Code
		Username    string `binding:"max=64"`           // 用户名
		ContentType string `binding:"required,max=128"` // 系统名称
		Secret      string `binding:"required,max=128"` // AppSecret
		Events      string `binding:"required,max=256"` // 监控什么事件
		Description string `binding:"max=255"`          // 描述
		Status      string `binding:"required,max=128"` // 状态
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 从上下文获取 user_id
	webhook := model.Webhook{
		ID:          req.ID,
		Url:         req.Url,
		TableCode:   req.TableCode,
		Username:    req.Username,
		ContentType: req.ContentType,
		Secret:      req.Secret,
		Events:      req.Events,
		Description: req.Description,
		Status:      req.Status,
	}

	err3 := h.webhookService.Update(c, &webhook)
	if err3 != nil {
		resp.HandleError(c, http.StatusInternalServerError, err3.Error(), nil)
		return
	}
	resp.HandleSuccess(c, nil)
}

// Delete 删除Webhook
// @Summary 删除Webhook
// @Tags Webhook管理
// @Accept json
// @Produce json
// @Param id path int true "WebhookID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/webhooks/{id} [delete]
func (h *webhookHandler) Delete(c *gin.Context) {
	var req struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.webhookService.Delete(c, req.ID)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, nil)
}

func (h *webhookHandler) BatchCreate(c *gin.Context) {
	resp.HandleSuccess(c, nil)
}

// BatchUpdate 批量更新Webhook状态
// @Summary 批量更新Webhook状态
// @Tags Webhook管理
// @Accept json
// @Produce json
// @Param data body object true "批量更新请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/webhooks/batch [put]
func (h *webhookHandler) BatchUpdate(c *gin.Context) {
	var req struct {
		Ids    []uint `form:"ids" binding:"required"`
		Status string `form:"status" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var webhook model.Webhook
	webhook.Status = req.Status

	err := h.webhookService.BatchUpdate(c, req.Ids, &webhook)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
}

// BatchDelete 批量删除Webhook
// @Summary 批量删除Webhook
// @Tags Webhook管理
// @Accept json
// @Produce json
// @Param data body object true "批量删除请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/webhooks/batch [delete]
func (h *webhookHandler) BatchDelete(c *gin.Context) {
	var req struct {
		Ids []uint `form:"ids" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	err := h.webhookService.BatchDelete(c, req.Ids)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
}
