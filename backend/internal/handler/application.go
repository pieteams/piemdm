package handler

import (
	"net/http"

	"piemdm/internal/model"
	"piemdm/internal/service"
	"piemdm/pkg/helper/resp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/xid"
)

type ApplicationHandler interface {
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

type applicationHandler struct {
	*Handler
	applicationService service.ApplicationService
}

func NewApplicationHandler(handler *Handler, applicationService service.ApplicationService) ApplicationHandler {
	return &applicationHandler{
		Handler:            handler,
		applicationService: applicationService,
	}
}

// List 获取应用列表
// @Summary 获取应用列表
// @Tags 应用管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(15)
// @Param appId query string false "应用ID"
// @Param startDate query string false "开始日期"
// @Param endDate query string false "结束日期"
// @Success 200 {array} model.Application
// @Router /admin/applications [get]
func (h *applicationHandler) List(c *gin.Context) {
	var req struct {
		Page      int    `form:"page,default=1"`
		PageSize  int    `form:"pageSize,default=15"`
		ShowType  string `form:"showType"`
		StartDate string `form:"startDate"`
		EndDate   string `form:"endDate"`
		AppId     string `form:"appId"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	page := req.Page
	pageSize := req.PageSize
	where := make(map[string]any)
	var total int64

	if req.AppId != "" {
		where["app_id"] = req.AppId
	}
	if req.StartDate != "" {
		where["created_at >="] = req.StartDate
	}
	if req.EndDate != "" {
		where["created_at <="] = req.EndDate
	}

	applications, err := h.applicationService.List(page, pageSize, &total, where)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	links := resp.GeneratePaginationLinks(c.Request, page, pageSize, int(total))
	c.Header("Link", links.String())

	resp.HandleSuccess(c, applications)
}

// Get 获取应用详情
// @Summary 获取应用详情
// @Tags 应用管理
// @Accept json
// @Produce json
// @Param id path int true "应用ID"
// @Success 200 {object} model.Application
// @Router /admin/applications/{id} [get]
func (h *applicationHandler) Get(c *gin.Context) {
	var req struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	application, err := h.applicationService.Get(req.Id)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, application)
}

// Create 创建应用
// @Summary 创建应用
// @Tags 应用管理
// @Accept json
// @Produce json
// @Param data body object true "创建应用请求"
// @Success 200 {object} model.Application
// @Router /admin/applications [post]
func (h *applicationHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `binding:"required,max=128"` // 系统名称
		IP          string `binding:"max=128"`          // 系统来源IP
		Description string `binding:"max=255"`
		Status      string `binding:"oneof=Normal Frozen"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	application := model.Application{
		AppId:       xid.New().String(),
		AppSecret:   uuid.New().String(),
		Name:        req.Name,
		IP:          req.IP,
		Description: req.Description,
		Status:      req.Status,
	}

	err := h.applicationService.Create(c, &application)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, application)
}

// Update 更新应用
// @Summary 更新应用
// @Tags 应用管理
// @Accept json
// @Produce json
// @Param id path int true "应用ID"
// @Param data body object true "更新应用请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/applications/{id} [put]
func (h *applicationHandler) Update(c *gin.Context) {
	var req struct {
		Id          uint   `binding:"required"`
		Name        string `binding:"required,max=128"` // 系统名称
		IP          string `binding:"max=128"`          // 系统来源IP
		Description string `binding:"max=255"`
		Status      string `binding:"oneof=Normal Frozen"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	application := model.Application{
		ID:          req.Id,
		Name:        req.Name,
		IP:          req.IP,
		Description: req.Description,
		Status:      req.Status,
	}

	err := h.applicationService.Update(c, &application)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, nil)
}

// Delete 删除应用
// @Summary 删除应用
// @Tags 应用管理
// @Accept json
// @Produce json
// @Param id path int true "应用ID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/applications/{id} [delete]
func (h *applicationHandler) Delete(c *gin.Context) {
	var req struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.applicationService.Delete(c, req.Id)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, nil)
}

func (h *applicationHandler) BatchCreate(c *gin.Context) {
}

// BatchUpdate 批量更新应用状态
// @Summary 批量更新应用状态
// @Tags 应用管理
// @Accept json
// @Produce json
// @Param data body object true "批量更新请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/applications/batch [put]
func (h *applicationHandler) BatchUpdate(c *gin.Context) {
	var req struct {
		Ids    []uint `form:"ids" binding:"required"`
		Status string `form:"status" binding:"required,oneof=Normal Frozen"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var application model.Application
	application.Status = req.Status

	err := h.applicationService.BatchUpdate(c, req.Ids, &application)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
	// resp.HandleSuccess(c, nil)
}

// BatchDelete 批量删除应用
// @Summary 批量删除应用
// @Tags 应用管理
// @Accept json
// @Produce json
// @Param data body object true "批量删除请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/applications/batch [delete]
func (h *applicationHandler) BatchDelete(c *gin.Context) {
	var req struct {
		Ids []uint `form:"ids" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	err := h.applicationService.BatchDelete(c, req.Ids)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
}
