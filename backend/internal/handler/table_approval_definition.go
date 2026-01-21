package handler

import (
	"net/http"

	"piemdm/internal/model"
	"piemdm/internal/service"
	"piemdm/pkg/helper/resp"

	"github.com/gin-gonic/gin"
)

type TableApprovalDefinitionHandler interface {
	List(c *gin.Context)
	Get(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	BatchCreate(c *gin.Context)
	BatchDelete(c *gin.Context)
}

type tableApprovalDefinitionHandler struct {
	*Handler
	tableApprovalDefinitionService service.TableApprovalDefinitionService
}

func NewTableApprovalDefinitionHandler(handler *Handler, tableApprovalDefinitionService service.TableApprovalDefinitionService) TableApprovalDefinitionHandler {
	return &tableApprovalDefinitionHandler{
		Handler:                        handler,
		tableApprovalDefinitionService: tableApprovalDefinitionService,
	}
}

// List 获取表审批定义列表
// @Summary 获取表审批定义列表
// @Tags 表审批定义
// @Accept json
// @Produce json
// @Param entity_code query string true "实体编码"
// @Param operation query string false "操作类型"
// @Success 200 {array} model.TableApprovalDefinition
// @Router /admin/table_approval_defs [get]
func (h *tableApprovalDefinitionHandler) List(c *gin.Context) {
	var req struct {
		EntityCode string `form:"entity_code" binding:"required,max=64"`
		Operation  string `form:"operation" binding:"max=64"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	list, err := h.tableApprovalDefinitionService.List(req.EntityCode, req.Operation)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, list)
}

// Get 获取表审批定义详情
// @Summary 获取表审批定义详情
// @Tags 表审批定义
// @Accept json
// @Produce json
// @Param id path int true "定义ID"
// @Success 200 {object} model.TableApprovalDefinition
// @Router /admin/table_approval_defs/{id} [get]
func (h *tableApprovalDefinitionHandler) Get(c *gin.Context) {
	var req struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	item, err := h.tableApprovalDefinitionService.Get(req.Id)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, item)
}

// Create 创建表审批定义
// @Summary 创建表审批定义
// @Tags 表审批定义
// @Accept json
// @Produce json
// @Param data body object true "创建请求"
// @Success 200 {object} model.TableApprovalDefinition
// @Router /admin/table_approval_defs [post]
func (h *tableApprovalDefinitionHandler) Create(c *gin.Context) {
	var req struct {
		EntityCode      string `json:"entity_code" binding:"required,max=64"`
		Operation       string `json:"operation" binding:"required,max=64"`
		ApprovalDefCode string `json:"approval_def_code" binding:"required,max=64"`
		Description     string `json:"description" binding:"max=255"`
		Status          string `json:"status" binding:"required,max=8"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	tableApprovalDefinition := model.TableApprovalDefinition{
		EntityCode:      req.EntityCode,
		Operation:       req.Operation,
		ApprovalDefCode: req.ApprovalDefCode,
		Description:     req.Description,
		Status:          req.Status,
	}

	err := h.tableApprovalDefinitionService.Create(c, &tableApprovalDefinition)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, tableApprovalDefinition)
}

// Update 更新表审批定义
// @Summary 更新表审批定义
// @Tags 表审批定义
// @Accept json
// @Produce json
// @Param id path int true "定义ID"
// @Param data body object true "更新请求"
// @Success 200 {object} model.TableApprovalDefinition
// @Router /admin/table_approval_defs/{id} [put]
func (h *tableApprovalDefinitionHandler) Update(c *gin.Context) {
	var req struct {
		ID              uint   `json:"id" binding:"required"`
		EntityCode      string `json:"entity_code" binding:"required,max=64"`
		Operation       string `json:"operation" binding:"required,max=64"`
		ApprovalDefCode string `json:"approval_def_code" binding:"required,max=64"`
	}

	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	tableApprovalDefinition := model.TableApprovalDefinition{
		ID:              req.ID,
		EntityCode:      req.EntityCode,
		Operation:       req.Operation,
		ApprovalDefCode: req.ApprovalDefCode,
	}

	err := h.tableApprovalDefinitionService.Update(c, &tableApprovalDefinition)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, tableApprovalDefinition)
}

// Delete 删除表审批定义
// @Summary 删除表审批定义
// @Tags 表审批定义
// @Accept json
// @Produce json
// @Param id path int true "定义ID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/table_approval_defs/{id} [delete]
func (h *tableApprovalDefinitionHandler) Delete(c *gin.Context) {
	var req struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.tableApprovalDefinitionService.Delete(c, req.ID)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, nil)
}

// BatchCreate 批量创建表审批定义
// @Summary 批量创建表审批定义
// @Tags 表审批定义
// @Accept json
// @Produce json
// @Param data body object true "批量创建请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/table_approval_defs/batch_create [post]
func (h *tableApprovalDefinitionHandler) BatchCreate(c *gin.Context) {
	var req struct {
		List []model.TableApprovalDefinition `form:"list" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.tableApprovalDefinitionService.BatchCreate(c, req.List)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, nil)
}

// BatchDelete 批量删除表审批定义
// @Summary 批量删除表审批定义
// @Tags 表审批定义
// @Accept json
// @Produce json
// @Param data body object true "批量删除请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/table_approval_defs/batch_delete [post]
func (h *tableApprovalDefinitionHandler) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []uint `form:"ids" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.tableApprovalDefinitionService.BatchDelete(c, req.IDs)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, nil)
}
