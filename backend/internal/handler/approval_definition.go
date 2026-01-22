package handler

import (
	"net/http"
	"strconv"

	"piemdm/internal/model"
	"piemdm/internal/pkg/request"
	"piemdm/internal/service"
	"piemdm/pkg/helper/resp"

	"github.com/gin-gonic/gin"
)

type ApprovalDefinitionHandler interface {
	// 基础CRUD接口
	List(c *gin.Context)
	Get(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	BatchCreate(c *gin.Context)
	BatchUpdate(c *gin.Context)
	BatchDelete(c *gin.Context)

	// 业务特有接口
	GetByCode(c *gin.Context)
	GetByEntity(c *gin.Context)
	GetActiveByEntity(c *gin.Context)
	Activate(c *gin.Context)
	Deactivate(c *gin.Context)
	Publish(c *gin.Context)
	GetVersions(c *gin.Context)
	CreateVersion(c *gin.Context)
	Validate(c *gin.Context)
}

type approvalDefinitionHandler struct {
	*Handler
	approvalDefinitionService service.ApprovalDefinitionService
}

func NewApprovalDefinitionHandler(handler *Handler, approvalDefinitionService service.ApprovalDefinitionService) ApprovalDefinitionHandler {
	return &approvalDefinitionHandler{
		Handler:                   handler,
		approvalDefinitionService: approvalDefinitionService,
	}
}

// List 获取审批定义列表
// @Summary 获取审批定义列表
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param query query request.QueryApprovalDefRequest false "查询参数"
// @Success 200 {array} model.ApprovalDefinition
// @Router /api/v1/approval-defs [get]
func (h *approvalDefinitionHandler) List(c *gin.Context) {
	var params request.QueryApprovalDefRequest
	if err := c.ShouldBindQuery(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 15
	}

	where := make(map[string]any)
	var total int64

	// 构建查询条件
	if params.Code != "" {
		where["code LIKE ?"] = "%" + params.Code + "%"
	}
	if params.Name != "" {
		where["name LIKE ?"] = "%" + params.Name + "%"
	}
	// if params.EntityCode != "" {
	// 	where["entity_code"] = params.EntityCode
	// }
	// if params.Category != "" {
	// 	where["category"] = params.Category
	// }
	if params.Status != "" {
		where["status"] = params.Status
	}
	if params.StartDate != "" {
		where["created_at >="] = params.StartDate
	}
	if params.EndDate != "" {
		where["created_at <="] = params.EndDate
	}

	approvalDefs, err := h.approvalDefinitionService.List(params.Page, params.PageSize, &total, where)
	if err != nil {
		h.logger.Error("获取审批定义列表失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	links := resp.GeneratePaginationLinks(c.Request, params.Page, params.PageSize, int(total))
	c.Header("Link", links.String())

	resp.HandleSuccess(c, approvalDefs)
}

// Get 获取审批定义详情
// @Summary 获取审批定义详情
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param id path int true "审批定义ID"
// @Success 200 {object} model.ApprovalDefinition
// @Router /api/v1/approval-defs/{id} [get]
func (h *approvalDefinitionHandler) Get(c *gin.Context) {
	var params struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	approvalDefinition, err := h.approvalDefinitionService.Get(params.Id)
	if err != nil {
		h.logger.Error("获取审批定义详情失败", "error", err, "id", params.Id)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, approvalDefinition)
}

// Create 创建审批定义
// @Summary 创建审批定义
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param request body request.CreateApprovalDefRequest true "创建审批定义请求"
// @Success 200 {object} model.ApprovalDefinition
// @Router /api/v1/approval-defs [post]
func (h *approvalDefinitionHandler) Create(c *gin.Context) {
	var req request.CreateApprovalDefRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为模型
	approvalDefinition := &model.ApprovalDefinition{
		Name:        req.Name,
		Description: req.Description,
		FormData:    req.FormData,
		NodeList:    req.NodeList,
		Platform:    req.Platform,
		Status:      req.Status,
	}

	err := h.approvalDefinitionService.Create(c, approvalDefinition)
	if err != nil {
		h.logger.Error("创建审批定义失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, approvalDefinition)
}

// Update 更新审批定义
// @Summary 更新审批定义
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param id path int true "审批定义ID"
// @Param request body request.UpdateApprovalDefRequest true "更新审批定义请求"
// @Success 200 {object} model.ApprovalDefinition
// @Router /api/v1/approval-defs/{id} [put]
func (h *approvalDefinitionHandler) Update(c *gin.Context) {
	var req request.UpdateApprovalDefRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 获取现有记录
	existingDef, err := h.approvalDefinitionService.Get(uint(req.ID))
	if err != nil {
		resp.HandleError(c, http.StatusNotFound, "审批定义不存在", nil)
		return
	}

	// 更新字段 - 只更新用户可以修改的字段
	existingDef.Name = req.Name
	existingDef.Description = req.Description
	existingDef.FormData = req.FormData
	existingDef.NodeList = req.NodeList
	existingDef.Platform = req.Platform
	existingDef.Status = req.Status

	// 系统字段由系统管理，不允许用户修改
	// - Code: 系统自动生成
	// - Version: 版本控制系统管理
	// - CreatedBy/UpdatedBy: 审计字段
	// - CreatedAt/UpdatedAt: 时间戳

	err = h.approvalDefinitionService.Update(c, existingDef)
	if err != nil {
		h.logger.Error("更新审批定义失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, existingDef)
}

// Delete 删除审批定义
// @Summary 删除审批定义
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param id path int true "审批定义ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-defs/{id} [delete]
func (h *approvalDefinitionHandler) Delete(c *gin.Context) {
	var req struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	deletedDef, err := h.approvalDefinitionService.Delete(c, req.Id)
	if err != nil {
		h.logger.Error("删除审批定义失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "删除成功",
		"data":    deletedDef,
	})
}

// BatchCreate 批量创建审批定义
func (h *approvalDefinitionHandler) BatchCreate(c *gin.Context) {
	var req []request.CreateApprovalDefRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var results []*model.ApprovalDefinition
	for _, item := range req {
		approvalDefinition := &model.ApprovalDefinition{
			Name:        item.Name,
			Description: item.Description,
			FormData:    item.FormData,
			NodeList:    item.NodeList,
			Status:      model.ApprovalDefStatusNormal,
		}

		if err := h.approvalDefinitionService.Create(c, approvalDefinition); err != nil {
			h.logger.Error("批量创建审批定义失败", "error", err)
			resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
			return
		}
		results = append(results, approvalDefinition)
	}

	resp.HandleSuccess(c, results)
}

// BatchUpdate 批量更新审批定义
func (h *approvalDefinitionHandler) BatchUpdate(c *gin.Context) {
	var req request.BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var approvalDefinition model.ApprovalDefinition
	approvalDefinition.Status = req.Status

	err := h.approvalDefinitionService.BatchUpdate(c, req.IDs, &approvalDefinition)
	if err != nil {
		h.logger.Error("批量更新审批定义失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "批量更新成功",
		"count":   len(req.IDs),
	})
}

// BatchDelete 批量删除审批定义
func (h *approvalDefinitionHandler) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []uint `form:"ids" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.approvalDefinitionService.BatchDelete(c, req.IDs)
	if err != nil {
		h.logger.Error("批量删除审批定义失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "批量删除成功",
		"count":   len(req.IDs),
	})
}

// GetByCode 根据编码获取审批定义
// @Summary 根据编码获取审批定义
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param code path string true "审批定义编码"
// @Success 200 {object} model.ApprovalDefinition
// @Router /api/v1/approval-defs/code/{code} [get]
func (h *approvalDefinitionHandler) GetByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批定义编码不能为空", nil)
		return
	}

	approvalDef, err := h.approvalDefinitionService.GetByCode(code)
	if err != nil {
		h.logger.Error("根据编码获取审批定义失败", "error", err, "code", code)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, approvalDef)
}

// GetByEntity 根据实体编码获取审批定义列表
// @Summary 根据实体编码获取审批定义列表
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param entityCode path string true "实体编码"
// @Success 200 {array} model.ApprovalDefinition
// @Router /api/v1/approval-defs/entity/{entityCode} [get]
func (h *approvalDefinitionHandler) GetByEntity(c *gin.Context) {
	entityCode := c.Param("entityCode")
	if entityCode == "" {
		resp.HandleError(c, http.StatusBadRequest, "实体编码不能为空", nil)
		return
	}

	approvalDefs, err := h.approvalDefinitionService.GetByEntityCode(entityCode)
	if err != nil {
		h.logger.Error("根据实体编码获取审批定义失败", "error", err, "entityCode", entityCode)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, approvalDefs)
}

// GetActiveByEntity 根据实体编码获取激活的审批定义列表
// @Summary 根据实体编码获取激活的审批定义列表
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param entityCode path string true "实体编码"
// @Success 200 {array} model.ApprovalDefinition
// @Router /api/v1/approval-defs/entity/{entityCode}/active [get]
func (h *approvalDefinitionHandler) GetActiveByEntity(c *gin.Context) {
	entityCode := c.Param("entityCode")
	if entityCode == "" {
		resp.HandleError(c, http.StatusBadRequest, "实体编码不能为空", nil)
		return
	}

	approvalDefs, err := h.approvalDefinitionService.GetActiveByEntityCode(entityCode)
	if err != nil {
		h.logger.Error("根据实体编码获取激活审批定义失败", "error", err, "entityCode", entityCode)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, approvalDefs)
}

// Activate 激活审批定义
// @Summary 激活审批定义
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param id path int true "审批定义ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-defs/{id}/activate [post]
func (h *approvalDefinitionHandler) Activate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	err = h.approvalDefinitionService.Activate(uint(id))
	if err != nil {
		h.logger.Error("激活审批定义失败", "error", err, "id", id)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "激活成功",
	})
}

// Deactivate 停用审批定义
// @Summary 停用审批定义
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param id path int true "审批定义ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-defs/{id}/deactivate [post]
func (h *approvalDefinitionHandler) Deactivate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	err = h.approvalDefinitionService.Deactivate(uint(id))
	if err != nil {
		h.logger.Error("停用审批定义失败", "error", err, "id", id)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "停用成功",
	})
}

// Publish 发布审批定义
// @Summary 发布审批定义
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param id path int true "审批定义ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-defs/{id}/publish [post]
func (h *approvalDefinitionHandler) Publish(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	err = h.approvalDefinitionService.Publish(uint(id))
	if err != nil {
		h.logger.Error("发布审批定义失败", "error", err, "id", id)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "发布成功",
	})
}

// GetVersions 获取审批定义版本列表
// @Summary 获取审批定义版本列表
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param code path string true "审批定义编码"
// @Success 200 {array} model.ApprovalDefinition
// @Router /api/v1/approval-defs/code/{code}/versions [get]
func (h *approvalDefinitionHandler) GetVersions(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批定义编码不能为空", nil)
		return
	}

	versions, err := h.approvalDefinitionService.GetVersions(code)
	if err != nil {
		h.logger.Error("获取审批定义版本失败", "error", err, "code", code)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, versions)
}

// CreateVersion 创建审批定义新版本
// @Summary 创建审批定义新版本
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param id path int true "审批定义ID"
// @Param request body request.CreateVersionRequest true "创建版本请求"
// @Success 200 {object} model.ApprovalDefinition
// @Router /api/v1/approval-defs/{id}/versions [post]
func (h *approvalDefinitionHandler) CreateVersion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	var req request.CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	newVersion, err := h.approvalDefinitionService.CreateNewVersion(c, uint(id), req.Comment)
	if err != nil {
		h.logger.Error("创建审批定义版本失败", "error", err, "id", id)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, newVersion)
}

// Validate 验证审批定义
// @Summary 验证审批定义
// @Tags 审批定义
// @Accept json
// @Produce json
// @Param id path int true "审批定义ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-defs/{id}/validate [post]
func (h *approvalDefinitionHandler) Validate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "无效的ID", nil)
		return
	}

	approvalDef, err := h.approvalDefinitionService.Get(uint(id))
	if err != nil {
		resp.HandleError(c, http.StatusNotFound, "审批定义不存在", nil)
		return
	}

	err = h.approvalDefinitionService.ValidateDefinition(approvalDef)
	if err != nil {
		resp.HandleSuccess(c, gin.H{
			"valid":   false,
			"message": err.Error(),
		})
		return
	}

	resp.HandleSuccess(c, gin.H{
		"valid":   true,
		"message": "验证通过",
	})
}
