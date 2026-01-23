package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"piemdm/internal/model"
	"piemdm/internal/transport/request"
	"piemdm/internal/service"
	"piemdm/pkg/helper/resp"

	"github.com/gin-gonic/gin"
)

type ApprovalHandler interface {
	// Base CRUD
	List(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)

	// Batch operations
	BatchDelete(c *gin.Context)

	// 业务特有方法
	ApproveTask(c *gin.Context)
	RejectTask(c *gin.Context)
	GetStatistics(c *gin.Context)
}

type approvalHandler struct {
	*Handler
	approvalService service.ApprovalService
}

func NewApprovalHandler(handler *Handler, approvalService service.ApprovalService) ApprovalHandler {
	return &approvalHandler{
		Handler:         handler,
		approvalService: approvalService,
	}
}

// Delete 删除审批实例
// @Summary 删除审批实例
// @Tags 审批实例
// @Accept json
// @Produce json
// @Param id path int true "审批实例ID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/approvals/{id} [delete]
func (h *approvalHandler) Delete(c *gin.Context) {
	var params struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	deletedApproval, err := h.approvalService.Delete(c, params.ID)
	if err != nil {
		h.logger.Error("删除审批实例失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "删除成功",
		"data":    deletedApproval,
	})
}

// BatchDelete 批量删除审批实例
// @Summary 批量删除审批实例
// @Tags 审批实例
// @Accept json
// @Produce json
// @Param data body object true "批量删除请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/approvals/batch [delete]
func (h *approvalHandler) BatchDelete(c *gin.Context) {
	var params struct {
		Ids []uint `form:"ids"`
	}
	if err := c.ShouldBind(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	err := h.approvalService.BatchDelete(c, params.Ids)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
}

// ApproveTask 审批任务
// @Summary 审批任务
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Param data body object true "审批参数"
// @Success 200 {object} map[string]interface{}
// @Router /admin/approval_tasks/{id}/approve [post]
func (h *approvalHandler) ApproveTask(c *gin.Context) {
	taskId64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	taskId := uint(taskId64)

	var params struct {
		Comment string `form:"comment" json:"comment"`
		Action  string `form:"action" json:"action"`
	}

	if err := c.ShouldBind(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	h.logger.Debug("handler approval ApproveTask 2025-12-06", "params", params)

	// taskId, _ := strconv.ParseInt(params.TaskId, 10, 64)
	// h.logger.Debug("handler approval ApproveTask", "taskId", taskId)
	// if err := h.approvalService.ApproveTask(c, taskId, params.Comment); err != nil {
	if err := h.approvalService.ProcessTask(c, taskId, params.Action, params.Comment); err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
}

// RejectTask 拒绝任务
// @Summary 拒绝任务
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Param data body object true "拒绝参数"
// @Success 200 {object} map[string]interface{}
// @Router /admin/approval_tasks/{id}/reject [post]
func (h *approvalHandler) RejectTask(c *gin.Context) {
	taskId64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	taskId := uint(taskId64)
	h.logger.Debug("handler approval RejectTask", "taskId", taskId)
	var params struct {
		Comment string `form:"comment" json:"comment"`
		Action  string `form:"action" json:"action"`
	}
	h.logger.Debug("handler approval RejectTask", "params", params)
	if err := c.ShouldBind(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 如果没有传递action，默认为REJECT
	action := params.Action
	if action == "" {
		action = "REJECT"
	}

	h.logger.Debug("handler approval TaskReject", "taskId", taskId, "action", action, "comment", params.Comment)
	if err := h.approvalService.ProcessTask(c, taskId, action, params.Comment); err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
}

// ListApprovals 获取审批实例列表
// @Summary 获取审批实例列表
// @Tags 审批实例
// @Accept json
// @Produce json
// @Param query query request.QueryApprovalRequest false "查询参数"
// @Success 200 {array} model.Approval
// @Router /api/v1/approvals [get]
func (h *approvalHandler) List(c *gin.Context) {
	page, pageSize := GetPage(c)

	var total int64
	var approvals []*model.Approval
	var err error

	if strings.Contains(c.Request.URL.Path, "/admin") {
		// 管理员视图 - 保持原有逻辑
		where := make(map[string]any)
		code := c.Query("code")
		startDate := c.Query("startDate")
		endDate := c.Query("endDate")

		if code != "" {
			where["code"] = code
		}
		if startDate != "" {
			where["created_at >="] = startDate
		}
		if endDate != "" {
			where["created_at <="] = endDate
		}

		approvals, err = h.approvalService.List(page, pageSize, &total, where)
	} else {
		// 用户视图 - 使用新逻辑
		userName := c.GetString("user_name")
		showType := c.Query("showType")
		timeRange := c.Query("timeRange")

		switch showType {
		case "Approved", "ProcessedByMe":
			// 我审批过的 - 使用新方法,关联approval_task表
			approvals, err = h.approvalService.FindProcessedByAssignee(userName, page, pageSize, &total, timeRange)

		case "Created":
			// 我提交的 - 保持原逻辑
			where := make(map[string]any)
			where["created_by"] = userName

			// 应用时间范围过滤
			switch timeRange {
			case "Today":
				where["created_at >="] = time.Now().Format("2006-01-02 00:00:00")
			case "LastWeek":
				where["created_at >="] = time.Now().AddDate(0, 0, -7).Format("2006-01-02 00:00:00")
			case "LastMonth":
				where["created_at >="] = time.Now().AddDate(0, 0, -30).Format("2006-01-02 00:00:00")
			}

			approvals, err = h.approvalService.List(page, pageSize, &total, where)

		case "All":
			// 全部相关的 - 暂时只显示我提交的
			// TODO: 未来可以实现联合查询(我提交的 + 我审批的)
			where := make(map[string]any)
			where["created_by"] = userName

			// 应用时间范围过滤
			switch timeRange {
			case "Today":
				where["created_at >="] = time.Now().Format("2006-01-02 00:00:00")
			case "LastWeek":
				where["created_at >="] = time.Now().AddDate(0, 0, -7).Format("2006-01-02 00:00:00")
			case "LastMonth":
				where["created_at >="] = time.Now().AddDate(0, 0, -30).Format("2006-01-02 00:00:00")
			}

			approvals, err = h.approvalService.List(page, pageSize, &total, where)

		default:
			// 待我审批的(Pending或空) - 使用新方法,关联approval_task表
			approvals, err = h.approvalService.FindPendingByAssignee(userName, page, pageSize, &total, timeRange)
		}
	}

	if err != nil {
		h.logger.Error("获取审批实例列表失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	links := resp.GeneratePaginationLinks(c.Request, page, pageSize, int(total))
	c.Header("Link", links.String())

	resp.HandleSuccess(c, approvals)
}

// GetApproval 获取审批实例详情
// @Summary 获取审批实例详情
// @Tags 审批实例
// @Accept json
// @Produce json
// @Param id path int true "审批实例ID"
// @Success 200 {object} model.Approval
// @Router /api/v1/approvals/{id} [get]
func (h *approvalHandler) Get(c *gin.Context) {
	var params struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	approval, err := h.approvalService.Get(params.Id)
	if err != nil {
		h.logger.Error("获取审批实例详情失败", "error", err, "id", params.Id)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, approval)
}

// CreateApproval 创建审批实例（已废弃，使用StartApproval）
func (h *approvalHandler) CreateApproval(c *gin.Context) {
	resp.HandleError(c, http.StatusBadRequest, "请使用StartApproval接口启动审批流程", nil)
}

// UpdateApproval 更新审批实例
func (h *approvalHandler) UpdateApproval(c *gin.Context) {
	resp.HandleError(c, http.StatusBadRequest, "审批实例不支持直接更新，请使用相应的业务接口", nil)
}

// DeleteApproval 删除审批实例
func (h *approvalHandler) Delete_Old(c *gin.Context) {
	var params struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	deletedApproval, err := h.approvalService.Delete(c, params.Id)
	if err != nil {
		h.logger.Error("删除审批实例失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "删除成功",
		"data":    deletedApproval,
	})
}

// BatchCreateApprovals 批量创建审批实例（不支持）
func (h *approvalHandler) BatchCreateApprovals(c *gin.Context) {
	resp.HandleError(c, http.StatusBadRequest, "不支持批量创建审批实例", nil)
}

// BatchUpdateApprovals 批量更新审批实例（不支持）
func (h *approvalHandler) BatchUpdateApprovals(c *gin.Context) {
	resp.HandleError(c, http.StatusBadRequest, "不支持批量更新审批实例", nil)
}

// BatchDeleteApprovals 批量删除审批实例
func (h *approvalHandler) BatchDelete_Old(c *gin.Context) {
	var req request.BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.approvalService.BatchDelete(c, req.IDs)
	if err != nil {
		h.logger.Error("批量删除审批实例失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "批量删除成功",
		"count":   len(req.IDs),
	})
}

// StartApproval 启动审批流程
// @Summary 启动审批流程
// @Tags 审批实例
// @Accept json
// @Produce json
// @Param request body request.StartApprovalRequest true "启动审批请求"
// @Success 200 {object} model.Approval
// @Router /api/v1/approvals/start [post]
func (h *approvalHandler) StartApproval(c *gin.Context) {
	var req request.StartApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 获取当前用户ID（这里简化处理，实际应该从JWT或Session中获取）
	applicantID := c.GetString("user_id")
	if applicantID == "" {
		applicantID = "default_user" // 临时处理
	}

	approval, err := h.approvalService.StartApprovalFlow(c, req.ApprovalDefCode, applicantID, req.Title, req.FormData)
	if err != nil {
		h.logger.Error("启动审批流程失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, approval)
}

// ProcessApproval 处理审批
// @Summary 处理审批
// @Tags 审批实例
// @Accept json
// @Produce json
// @Param request body request.ProcessApprovalRequest true "处理审批请求"
// @Success 200 {object} map[string]any
// @Router /api/v1/approvals/process [post]
func (h *approvalHandler) ProcessApproval(c *gin.Context) {
	var req request.ProcessApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 获取当前用户ID
	assigneeID := c.GetString("user_id")
	if assigneeID == "" {
		assigneeID = "default_user" // 临时处理
	}

	err := h.approvalService.ProcessApprovalByCodeFlow(c, req.ApprovalCode, req.NodeCode, assigneeID, req.Action, req.Comment, req.Reason)
	if err != nil {
		h.logger.Error("处理审批失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "处理成功",
	})
}

// GetApprovalByCode 根据编码获取审批实例
// @Summary 根据编码获取审批实例
// @Tags 审批实例
// @Accept json
// @Produce json
// @Param code path string true "审批编码"
// @Success 200 {object} model.Approval
// @Router /api/v1/approvals/code/{code} [get]
func (h *approvalHandler) GetByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批编码不能为空", nil)
		return
	}

	approval, err := h.approvalService.GetByCode(code)
	if err != nil {
		h.logger.Error("根据编码获取审批实例失败", "error", err, "code", code)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, approval)
}

// GetApprovalsByApplicant 根据申请人获取审批实例列表
// @Summary 根据申请人获取审批实例列表
// @Tags 审批实例
// @Accept json
// @Produce json
// @Param applicantId path string true "申请人ID"
// @Success 200 {array} model.Approval
// @Router /api/v1/approvals/applicant/{applicantId} [get]
func (h *approvalHandler) GetsByApplicant(c *gin.Context) {
	applicantID := c.Param("applicantId")
	if applicantID == "" {
		resp.HandleError(c, http.StatusBadRequest, "申请人ID不能为空", nil)
		return
	}

	approvals, err := h.approvalService.GetByApplicant(applicantID)
	if err != nil {
		h.logger.Error("根据申请人获取审批实例失败", "error", err, "applicantId", applicantID)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, approvals)
}

// GetApprovalsByStatus 根据状态获取审批实例列表
// @Summary 根据状态获取审批实例列表
// @Tags 审批实例
// @Accept json
// @Produce json
// @Param status path string true "审批状态"
// @Success 200 {array} model.Approval
// @Router /api/v1/approvals/status/{status} [get]
func (h *approvalHandler) GetsByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批状态不能为空", nil)
		return
	}

	approvals, err := h.approvalService.GetByStatus(status)
	if err != nil {
		h.logger.Error("根据状态获取审批实例失败", "error", err, "status", status)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, approvals)
}

// GetApprovalsByEntity 根据实体获取审批实例列表
// @Summary 根据实体获取审批实例列表
// @Tags 审批实例
// @Accept json
// @Produce json
// @Param entityCode path string true "实体编码"
// @Success 200 {array} model.Approval
// @Router /api/v1/approvals/entity/{entityCode} [get]
func (h *approvalHandler) GetsByEntity(c *gin.Context) {
	entityCode := c.Param("entityCode")
	if entityCode == "" {
		resp.HandleError(c, http.StatusBadRequest, "实体编码不能为空", nil)
		return
	}

	approvals, err := h.approvalService.GetByEntityCode(entityCode)
	if err != nil {
		h.logger.Error("根据实体获取审批实例失败", "error", err, "entityCode", entityCode)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, approvals)
}

// GetApprovalStatistics 获取审批统计信息
// @Summary 获取审批统计信息
// @Tags 审批实例
// @Accept json
// @Produce json
// @Param applicantId query string false "申请人ID"
// @Success 200 {object} map[string]int64
// @Router /api/v1/approvals/statistics [get]
func (h *approvalHandler) GetStatistics(c *gin.Context) {
	applicantID := c.Query("applicantId")
	// 如果是管理员，且没有传 applicantId，则返回全局统计
	// 这里目前简化处理，如果是通过 admin 路由进来的，可以认为是管理统计

	statistics, err := h.approvalService.GetApprovalStatisticsFlow(applicantID)
	if err != nil {
		h.logger.Error("获取审批统计信息失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, statistics)
}

// GetExpiredApprovals 获取过期的审批实例
// @Summary 获取过期的审批实例
// @Tags 审批实例
// @Accept json
// @Produce json
// @Success 200 {array} model.Approval
// @Router /api/v1/approvals/expired [get]
// func (h *approvalHandler) GetExpiredApprovals(c *gin.Context) {
// 	approvals, err := h.approvalService.GetExpiredApprovalsFlow()
// 	if err != nil {
// 		h.logger.Error("获取过期审批实例失败", "error", err)
// 		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
// 		return
// 	}

// 	resp.HandleSuccess(c, approvals)
// }

// GetApprovalHistory 获取审批历史
// @Summary 获取审批历史
// @Tags 审批实例
// @Accept json
// @Produce json
// @Param id path int true "审批实例ID"
// @Success 200 {array} model.Approval
// @Router /api/v1/approvals/{id}/history [get]
func (h *approvalHandler) GetHistory(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批实例ID不能为空", nil)
		return
	}
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "审批实例ID格式错误", nil)
		return
	}
	id := uint(id64)
	histories, err := h.approvalService.GetApprovalHistory(id)
	if err != nil {
		h.logger.Error("获取审批历史失败", "error", err, "id", id)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, histories)
}
