package handler

import (
	"net/http"
	"strconv"

	"piemdm/internal/model"
	"piemdm/internal/transport/request"
	"piemdm/internal/service"
	"piemdm/pkg/helper/resp"

	"github.com/gin-gonic/gin"
)

type ApprovalTaskHandler interface {
	// 基础CRUD接口
	List(c *gin.Context)
	Get(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	BatchCreate(c *gin.Context)
	BatchUpdate(c *gin.Context)
	BatchDelete(c *gin.Context)

	// 任务处理业务接口
	// ProcessTask(c *gin.Context)
	// ApproveTask(c *gin.Context)
	// RejectTask(c *gin.Context)
	// TransferTask(c *gin.Context)
	// RemindTask(c *gin.Context)
	// BatchRemindTasks(c *gin.Context)
	// GetTasksByAssignee(c *gin.Context)
	// GetPendingTasksByAssignee(c *gin.Context)
	// GetTasksByApproval(c *gin.Context)
	// GetOverdueTasks(c *gin.Context)
	// GetExpiredTasks(c *gin.Context)
	// GetTaskStatistics(c *gin.Context)
}

type approvalTaskHandler struct {
	*Handler
	approvalTaskService service.ApprovalTaskService
	approvalService     service.ApprovalService
}

func NewApprovalTaskHandler(handler *Handler, approvalTaskService service.ApprovalTaskService, approvalService service.ApprovalService) ApprovalTaskHandler {
	return &approvalTaskHandler{
		Handler:             handler,
		approvalTaskService: approvalTaskService,
		approvalService:     approvalService,
	}
}

// ListApprovalTasks 获取审批任务列表
// @Summary 获取审批任务列表
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param query query request.QueryApprovalTaskRequest false "查询参数"
// @Success 200 {array} model.ApprovalTask
// @Router /api/v1/approval-tasks [get]
func (h *approvalTaskHandler) List(c *gin.Context) {
	page, pageSize := GetPage(c)
	where := make(map[string]any)
	var total int64

	// 构建查询条件
	approvalCode := c.Query("approval_code")
	nodeType := c.Query("node_type")
	assigneeID := c.Query("assignee_id")
	status := c.Query("status")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// 构建查询条件
	if approvalCode != "" {
		where["approval_code"] = approvalCode
	}
	if nodeType != "" {
		where["node_code"] = nodeType
	}
	if assigneeID != "" {
		where["assignee_id"] = assigneeID
	}
	if status != "" {
		where["status"] = status
	}
	if startDate != "" {
		where["created_at >="] = startDate
	}
	if endDate != "" {
		where["created_at <="] = endDate
	}

	tasks, err := h.approvalTaskService.List(page, pageSize, &total, where, "created_at ASC")
	if err != nil {
		h.logger.Error("获取审批任务列表失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	links := resp.GeneratePaginationLinks(c.Request, page, pageSize, int(total))
	c.Header("Link", links.String())

	resp.HandleSuccess(c, tasks)
}

// GetApprovalTask 获取审批任务详情
// @Summary 获取审批任务详情
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param id path int true "审批任务ID"
// @Success 200 {object} model.ApprovalTask
// @Router /api/v1/approval-tasks/{id} [get]
func (h *approvalTaskHandler) Get(c *gin.Context) {
	var params struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	task, err := h.approvalTaskService.Get(params.Id)
	if err != nil {
		h.logger.Error("获取审批任务详情失败", "error", err, "id", params.Id)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, task)
}

// CreateApprovalTask 创建审批任务
// @Summary 创建审批任务
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param request body model.ApprovalTask true "创建审批任务请求"
// @Success 200 {object} model.ApprovalTask
// @Router /api/v1/approval-tasks [post]
func (h *approvalTaskHandler) Create(c *gin.Context) {
	var task model.ApprovalTask
	if err := c.ShouldBindJSON(&task); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.approvalTaskService.Create(&task)
	if err != nil {
		h.logger.Error("创建审批任务失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, task)
}

// UpdateApprovalTask 更新审批任务
// @Summary 更新审批任务
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param id path int true "审批任务ID"
// @Param request body model.ApprovalTask true "更新审批任务请求"
// @Success 200 {object} model.ApprovalTask
// @Router /api/v1/approval-tasks/{id} [put]
func (h *approvalTaskHandler) Update(c *gin.Context) {
	var params struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var task model.ApprovalTask
	if err := c.ShouldBindJSON(&task); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 设置ID
	task.ID = uint(params.Id)

	err := h.approvalTaskService.Update(&task)
	if err != nil {
		h.logger.Error("更新审批任务失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, task)
}

// DeleteApprovalTask 删除审批任务
// @Summary 删除审批任务
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param id path int true "审批任务ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-tasks/{id} [delete]
func (h *approvalTaskHandler) Delete(c *gin.Context) {
	var params struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	deletedTask, err := h.approvalTaskService.Delete(params.Id)
	if err != nil {
		h.logger.Error("删除审批任务失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "删除成功",
		"data":    deletedTask,
	})
}

// BatchCreateApprovalTasks 批量创建审批任务（不支持）
func (h *approvalTaskHandler) BatchCreate(c *gin.Context) {
	resp.HandleError(c, http.StatusBadRequest, "不支持批量创建审批任务", nil)
}

// BatchUpdateApprovalTasks 批量更新审批任务
func (h *approvalTaskHandler) BatchUpdate(c *gin.Context) {
	var req request.BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var task model.ApprovalTask
	task.Status = req.Status

	err := h.approvalTaskService.BatchUpdate(req.IDs, &task)
	if err != nil {
		h.logger.Error("批量更新审批任务失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "批量更新成功",
		"count":   len(req.IDs),
	})
}

// BatchDeleteApprovalTasks 批量删除审批任务
func (h *approvalTaskHandler) BatchDelete(c *gin.Context) {
	var req request.BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.approvalTaskService.BatchDelete(req.IDs)
	if err != nil {
		h.logger.Error("批量删除审批任务失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "批量删除成功",
		"count":   len(req.IDs),
	})
}

// TransferTask 转交审批任务
// @Summary 转交审批任务
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Param request body request.TransferTaskRequest true "转交请求"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-tasks/{id}/transfer [post]
func (h *approvalTaskHandler) TransferTask(c *gin.Context) {
	idStr := c.Param("id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "无效的任务ID", nil)
		return
	}

	var req request.TransferTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 获取当前用户ID
	fromUserID := c.GetString("user_id")
	if fromUserID == "" {
		fromUserID = "default_user" // 临时处理
	}
	fromUserName := c.GetString("user_name")
	if fromUserName == "" {
		fromUserName = "默认用户" // 临时处理
	}

	err = h.approvalService.TransferTask(c, uint(taskID), fromUserID, req.ToUserID, fromUserName, req.ToUserName, req.Reason)
	if err != nil {
		h.logger.Error("转交审批任务失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "转交成功",
	})
}

// RemindTask 催办审批任务
// @Summary 催办审批任务
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-tasks/{id}/remind [post]
func (h *approvalTaskHandler) RemindTask(c *gin.Context) {
	idStr := c.Param("id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "无效的任务ID", nil)
		return
	}

	err = h.approvalTaskService.RemindTask(uint(taskID))
	if err != nil {
		h.logger.Error("催办审批任务失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "催办成功",
	})
}

// BatchRemindTasks 批量催办审批任务
// @Summary 批量催办审批任务
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param request body request.BatchRemindRequest true "批量催办请求"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-tasks/batch-remind [post]
func (h *approvalTaskHandler) BatchRemindTasks(c *gin.Context) {
	var req request.BatchRemindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.approvalTaskService.BatchRemindTasks(req.AssigneeID)
	if err != nil {
		h.logger.Error("批量催办审批任务失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "批量催办成功",
	})
}

// GetTasksByAssignee 根据审批人获取任务列表
// @Summary 根据审批人获取任务列表
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param assigneeId path string true "审批人ID"
// @Success 200 {array} model.ApprovalTask
// @Router /api/v1/approval-tasks/assignee/{assigneeId} [get]
func (h *approvalTaskHandler) GetTasksByAssignee(c *gin.Context) {
	assigneeID := c.Param("assigneeId")
	if assigneeID == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批人ID不能为空", nil)
		return
	}

	tasks, err := h.approvalTaskService.GetByAssignee(assigneeID)
	if err != nil {
		h.logger.Error("根据审批人获取任务失败", "error", err, "assigneeId", assigneeID)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, tasks)
}

// GetPendingTasksByAssignee 根据审批人获取待处理任务列表
// @Summary 根据审批人获取待处理任务列表
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param assigneeId path string true "审批人ID"
// @Success 200 {array} model.ApprovalTask
// @Router /api/v1/approval-tasks/assignee/{assigneeId}/pending [get]
func (h *approvalTaskHandler) GetPendingTasksByAssignee(c *gin.Context) {
	assigneeID := c.Param("assigneeId")
	if assigneeID == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批人ID不能为空", nil)
		return
	}

	tasks, err := h.approvalTaskService.GetPendingByAssignee(assigneeID)
	if err != nil {
		h.logger.Error("根据审批人获取待处理任务失败", "error", err, "assigneeId", assigneeID)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, tasks)
}

// GetTasksByApproval 根据审批实例获取任务列表
// @Summary 根据审批实例获取任务列表
// @Tags 审批任务
// @Accept json
// @Produce json
// @Param approvalCode path string true "审批编码"
// @Success 200 {array} model.ApprovalTask
// @Router /api/v1/approval-tasks/approval/{approvalCode} [get]
func (h *approvalTaskHandler) GetTasksByApproval(c *gin.Context) {
	approvalCode := c.Param("approvalCode")
	if approvalCode == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批编码不能为空", nil)
		return
	}

	tasks, err := h.approvalTaskService.GetByApprovalCode(approvalCode)
	if err != nil {
		h.logger.Error("根据审批实例获取任务失败", "error", err, "approvalCode", approvalCode)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, tasks)
}

// GetOverdueTasks 获取逾期任务列表
// @Summary 获取逾期任务列表
// @Tags 审批任务
// @Accept json
// @Produce json
// @Success 200 {array} model.ApprovalTask
// @Router /api/v1/approval-tasks/overdue [get]
func (h *approvalTaskHandler) GetOverdueTasks(c *gin.Context) {
	tasks, err := h.approvalTaskService.GetOverdueTasks()
	if err != nil {
		h.logger.Error("获取逾期任务失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, tasks)
}

// GetExpiredTasks 获取过期任务列表
// @Summary 获取过期任务列表
// @Tags 审批任务
// @Accept json
// @Produce json
// @Success 200 {array} model.ApprovalTask
// @Router /api/v1/approval-tasks/expired [get]
func (h *approvalTaskHandler) GetExpiredTasks(c *gin.Context) {
	tasks, err := h.approvalTaskService.GetExpiredTasks()
	if err != nil {
		h.logger.Error("获取过期任务失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, tasks)
}
