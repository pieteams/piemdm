package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"piemdm/internal/model"
	"piemdm/internal/transport/request"
	"piemdm/internal/service"
	"piemdm/pkg/helper/resp"

	"github.com/gin-gonic/gin"
)

// convertConfigToString 将配置map转换为JSON字符串
func convertConfigToString(config map[string]any) (string, error) {
	if config == nil {
		return "", nil
	}

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

type ApprovalNodeHandler interface {
	// 基础CRUD接口
	List(c *gin.Context)
	Get(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	BatchCreate(c *gin.Context)
	BatchUpdate(c *gin.Context)
	BatchDelete(c *gin.Context)

	// 节点配置业务接口
	GetNodesByApprovalDef(c *gin.Context)
	GetStartNode(c *gin.Context)
	GetEndNodes(c *gin.Context)
	GetNextNodes(c *gin.Context)
	ValidateWorkflow(c *gin.Context)
	ConfigureApprovers(c *gin.Context)
	ConfigureConditions(c *gin.Context)
	ConfigureTimeouts(c *gin.Context)
	ActivateNode(c *gin.Context)
	DeactivateNode(c *gin.Context)
	BatchSyncApprovalNodes(c *gin.Context)
}

type approvalNodeHandler struct {
	*Handler
	approvalNodeService service.ApprovalNodeService
}

func NewApprovalNodeHandler(handler *Handler, approvalNodeService service.ApprovalNodeService) ApprovalNodeHandler {
	return &approvalNodeHandler{
		Handler:             handler,
		approvalNodeService: approvalNodeService,
	}
}

// ListApprovalNodes 获取审批节点列表
// @Summary 获取审批节点列表
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param approvalDefCode query string false "审批定义编码"
// @Param nodeType query string false "节点类型"
// @Param page query int false "页码"
// @Param pageSize query int false "页大小"
// @Success 200 {array} model.ApprovalNode
// @Router /api/v1/approval-nodes [get]
func (h *approvalNodeHandler) List(c *gin.Context) {
	page, pageSize := GetPage(c)
	where := make(map[string]any)
	var total int64

	// 构建查询条件
	approvalDefCode := c.Query("approvalDefCode")
	nodeType := c.Query("nodeType")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if approvalDefCode != "" {
		where["approval_def_code"] = approvalDefCode
	}
	if nodeType != "" {
		where["node_type"] = nodeType
	}
	if startDate != "" {
		where["created_at >="] = startDate
	}
	if endDate != "" {
		where["created_at <="] = endDate
	}

	nodes, err := h.approvalNodeService.List(page, pageSize, &total, where)
	if err != nil {
		h.logger.Error("获取审批节点列表失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	links := resp.GeneratePaginationLinks(c.Request, page, pageSize, int(total))
	c.Header("Link", links.String())

	resp.HandleSuccess(c, nodes)
}

// GetApprovalNode 获取审批节点详情
// @Summary 获取审批节点详情
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param id path int true "审批节点ID"
// @Success 200 {object} model.ApprovalNode
// @Router /api/v1/approval-nodes/{id} [get]
func (h *approvalNodeHandler) Get(c *gin.Context) {
	var params struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	node, err := h.approvalNodeService.Get(params.Id)
	if err != nil {
		h.logger.Error("获取审批节点详情失败", "error", err, "id", params.Id)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, node)
}

// CreateApprovalNode 创建审批节点
// @Summary 创建审批节点
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param request body request.CreateApprovalNodeRequest true "创建审批节点请求"
// @Success 200 {object} model.ApprovalNode
// @Router /api/v1/approval-nodes [post]
func (h *approvalNodeHandler) Create(c *gin.Context) {
	var req request.CreateApprovalNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 转换为模型
	node := &model.ApprovalNode{
		// ApprovalDefCode: req.ApprovalDefCode,
		NodeCode:        req.NodeCode,
		NodeName:        req.NodeName,
		NodeType:        req.NodeType,
		ApproverType:    req.ApproverType,
		ApproverConfig:  req.ApproverConfig,
		ConditionConfig: req.ConditionConfig,
		SortOrder:       req.SortOrder,
		Status:          model.ApprovalDefStatusNormal,
	}

	err := h.approvalNodeService.Create(node)
	if err != nil {
		h.logger.Error("创建审批节点失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, node)
}

// UpdateApprovalNode 更新审批节点
// @Summary 更新审批节点
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param id path int true "审批节点ID"
// @Param request body request.UpdateApprovalNodeRequest true "更新审批节点请求"
// @Success 200 {object} model.ApprovalNode
// @Router /api/v1/approval-nodes/{id} [put]
func (h *approvalNodeHandler) Update(c *gin.Context) {
	var params struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var req request.UpdateApprovalNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 获取现有记录
	existingNode, err := h.approvalNodeService.Get(params.Id)
	if err != nil {
		resp.HandleError(c, http.StatusNotFound, "审批节点不存在", nil)
		return
	}

	// 更新字段
	existingNode.NodeName = req.NodeName
	existingNode.NodeType = req.NodeType
	existingNode.ApproverType = req.ApproverType
	existingNode.ApproverConfig = req.ApproverConfig
	existingNode.ConditionConfig = req.ConditionConfig
	existingNode.SortOrder = req.SortOrder

	err = h.approvalNodeService.Update(existingNode)
	if err != nil {
		h.logger.Error("更新审批节点失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, existingNode)
}

// DeleteApprovalNode 删除审批节点
// @Summary 删除审批节点
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param id path int true "审批节点ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-nodes/{id} [delete]
func (h *approvalNodeHandler) Delete(c *gin.Context) {
	var params struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	deletedNode, err := h.approvalNodeService.Delete(params.ID)
	if err != nil {
		h.logger.Error("删除审批节点失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "删除成功",
		"data":    deletedNode,
	})
}

// BatchCreateApprovalNodes 批量创建审批节点
func (h *approvalNodeHandler) BatchCreate(c *gin.Context) {
	var req []request.CreateApprovalNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var results []*model.ApprovalNode
	for _, item := range req {
		node := &model.ApprovalNode{
			// ApprovalDefCode: item.ApprovalDefCode,
			NodeCode:        item.NodeCode,
			NodeName:        item.NodeName,
			NodeType:        item.NodeType,
			ApproverType:    item.ApproverType,
			ApproverConfig:  item.ApproverConfig,
			ConditionConfig: item.ConditionConfig,
			SortOrder:       item.SortOrder,
			Status:          model.ApprovalDefStatusNormal,
		}

		if err := h.approvalNodeService.Create(node); err != nil {
			h.logger.Error("批量创建审批节点失败", "error", err)
			resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
			return
		}
		results = append(results, node)
	}

	resp.HandleSuccess(c, results)
}

// BatchUpdateApprovalNodes 批量更新审批节点
func (h *approvalNodeHandler) BatchUpdate(c *gin.Context) {
	var req request.BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var node model.ApprovalNode
	node.Status = req.Status

	err := h.approvalNodeService.BatchUpdate(req.IDs, &node)
	if err != nil {
		h.logger.Error("批量更新审批节点失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "批量更新成功",
		"count":   len(req.IDs),
	})
}

// BatchDeleteApprovalNodes 批量删除审批节点
func (h *approvalNodeHandler) BatchDelete(c *gin.Context) {
	var req request.BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.approvalNodeService.BatchDelete(req.IDs)
	if err != nil {
		h.logger.Error("批量删除审批节点失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "批量删除成功",
		"count":   len(req.IDs),
	})
}

// GetNodesByApprovalDef 根据审批定义获取节点列表
// @Summary 根据审批定义获取节点列表
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param approvalDefCode path string true "审批定义编码"
// @Success 200 {array} model.ApprovalNode
// @Router /api/v1/approval-nodes/approval-def/{approvalDefCode} [get]
func (h *approvalNodeHandler) GetNodesByApprovalDef(c *gin.Context) {
	approvalDefCode := c.Param("approvalDefCode")
	if approvalDefCode == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批定义编码不能为空", nil)
		return
	}

	nodes, err := h.approvalNodeService.GetByApprovalDefCode(approvalDefCode)
	if err != nil {
		h.logger.Error("根据审批定义获取节点失败", "error", err, "approvalDefCode", approvalDefCode)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, nodes)
}

// GetStartNode 获取开始节点
// @Summary 获取开始节点
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param approvalDefCode path string true "审批定义编码"
// @Success 200 {object} model.ApprovalNode
// @Router /api/v1/approval-nodes/approval-def/{approvalDefCode}/start [get]
func (h *approvalNodeHandler) GetStartNode(c *gin.Context) {
	approvalDefCode := c.Param("approvalDefCode")
	if approvalDefCode == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批定义编码不能为空", nil)
		return
	}

	node, err := h.approvalNodeService.GetStartNode(approvalDefCode)
	if err != nil {
		h.logger.Error("获取开始节点失败", "error", err, "approvalDefCode", approvalDefCode)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, node)
}

// GetEndNodes 获取结束节点列表
// @Summary 获取结束节点列表
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param approvalDefCode path string true "审批定义编码"
// @Success 200 {array} model.ApprovalNode
// @Router /api/v1/approval-nodes/approval-def/{approvalDefCode}/end [get]
func (h *approvalNodeHandler) GetEndNodes(c *gin.Context) {
	approvalDefCode := c.Param("approvalDefCode")
	if approvalDefCode == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批定义编码不能为空", nil)
		return
	}

	nodes, err := h.approvalNodeService.GetEndNodes(approvalDefCode)
	if err != nil {
		h.logger.Error("获取结束节点失败", "error", err, "approvalDefCode", approvalDefCode)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, nodes)
}

// GetNextNodes 获取下一个节点列表
// @Summary 获取下一个节点列表
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param approvalDefCode path string true "审批定义编码"
// @Param currentNodeCode path string true "当前节点编码"
// @Success 200 {array} model.ApprovalNode
// @Router /api/v1/approval-nodes/approval-def/{approvalDefCode}/next/{currentNodeCode} [get]
func (h *approvalNodeHandler) GetNextNodes(c *gin.Context) {
	approvalDefCode := c.Param("approvalDefCode")
	currentNodeCode := c.Param("currentNodeCode")

	if approvalDefCode == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批定义编码不能为空", nil)
		return
	}
	if currentNodeCode == "" {
		resp.HandleError(c, http.StatusBadRequest, "当前节点编码不能为空", nil)
		return
	}

	nodes, err := h.approvalNodeService.GetNextNodes(approvalDefCode, currentNodeCode)
	if err != nil {
		h.logger.Error("获取下一个节点失败", "error", err,
			"approvalDefCode", approvalDefCode,
			"currentNodeCode", currentNodeCode)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, nodes)
}

// ValidateWorkflow 验证工作流
// @Summary 验证工作流
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param approvalDefCode path string true "审批定义编码"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-nodes/approval-def/{approvalDefCode}/validate [post]
func (h *approvalNodeHandler) ValidateWorkflow(c *gin.Context) {
	approvalDefCode := c.Param("approvalDefCode")
	if approvalDefCode == "" {
		resp.HandleError(c, http.StatusBadRequest, "审批定义编码不能为空", nil)
		return
	}

	err := h.approvalNodeService.ValidateWorkflow(approvalDefCode)
	if err != nil {
		resp.HandleSuccess(c, gin.H{
			"valid":   false,
			"message": err.Error(),
		})
		return
	}

	resp.HandleSuccess(c, gin.H{
		"valid":   true,
		"message": "工作流验证通过",
	})
}

// ConfigureApprovers 配置审批人
// @Summary 配置审批人
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param id path int true "节点ID"
// @Param request body map[string]any true "审批人配置"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-nodes/{id}/configure-approvers [post]
func (h *approvalNodeHandler) ConfigureApprovers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "无效的节点ID", nil)
		return
	}

	var config map[string]any
	if err := c.ShouldBindJSON(&config); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 将配置转换为JSON字符串
	configStr, err := convertConfigToString(config)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "配置格式错误", nil)
		return
	}

	err = h.approvalNodeService.ConfigureApprovers(uint(id), configStr)
	if err != nil {
		h.logger.Error("配置审批人失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "配置成功",
	})
}

// ConfigureConditions 配置条件
// @Summary 配置条件
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param id path int true "节点ID"
// @Param request body map[string]any true "条件配置"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-nodes/{id}/configure-conditions [post]
func (h *approvalNodeHandler) ConfigureConditions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "无效的节点ID", nil)
		return
	}

	var config map[string]any
	if err := c.ShouldBindJSON(&config); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 将配置转换为JSON字符串
	configStr, err := convertConfigToString(config)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "配置格式错误", nil)
		return
	}

	err = h.approvalNodeService.ConfigureConditions(uint(id), configStr)
	if err != nil {
		h.logger.Error("配置条件失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "配置成功",
	})
}

// ConfigureTimeouts 配置超时
// @Summary 配置超时
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param id path int true "节点ID"
// @Param request body map[string]any true "超时配置"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-nodes/{id}/configure-timeouts [post]
func (h *approvalNodeHandler) ConfigureTimeouts(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "无效的节点ID", nil)
		return
	}

	var config map[string]any
	if err := c.ShouldBindJSON(&config); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 从配置中提取超时时间
	timeoutHours, ok := config["timeoutHours"]
	if !ok {
		resp.HandleError(c, http.StatusBadRequest, "缺少timeoutHours参数", nil)
		return
	}

	// 转换为int类型
	var timeoutInt int
	switch v := timeoutHours.(type) {
	case float64:
		timeoutInt = int(v)
	case int:
		timeoutInt = v
	case int64:
		timeoutInt = int(v)
	default:
		resp.HandleError(c, http.StatusBadRequest, "timeoutHours必须是数字类型", nil)
		return
	}

	err = h.approvalNodeService.ConfigureTimeouts(uint(id), timeoutInt)
	if err != nil {
		h.logger.Error("配置超时失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "配置成功",
	})
}

// ActivateNode 激活节点
// @Summary 激活节点
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param id path int true "节点ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-nodes/{id}/activate [post]
func (h *approvalNodeHandler) ActivateNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "无效的节点ID", nil)
		return
	}

	err = h.approvalNodeService.ActivateNode(uint(id))
	if err != nil {
		h.logger.Error("激活节点失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "激活成功",
	})
}

// DeactivateNode 停用节点
// @Summary 停用节点
// @Tags 审批节点
// @Accept json
// @Produce json
// @Param id path int true "节点ID"
// @Success 200 {object} map[string]any
// @Router /api/v1/approval-nodes/{id}/deactivate [post]
func (h *approvalNodeHandler) DeactivateNode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		resp.HandleError(c, http.StatusBadRequest, "无效的节点ID", nil)
		return
	}

	err = h.approvalNodeService.DeactivateNode(uint(id))
	if err != nil {
		h.logger.Error("停用节点失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "停用成功",
	})
}

// BatchSyncApprovalNodesRequest 批量同步请求
type BatchSyncApprovalNodesRequest struct {
	ApprovalDefCode string                              `json:"approvalDefCode" binding:"required"`
	Nodes           []request.CreateApprovalNodeRequest `json:"nodes"`
}

// BatchSyncResult 批量同步结果
type BatchSyncResult struct {
	Created []int64 `json:"created"` // 新创建的节点ID
	Updated []int64 `json:"updated"` // 更新的节点ID
	Deleted []int64 `json:"deleted"` // 删除的节点ID
	Total   int     `json:"total"`   // 总处理数量
	Success bool    `json:"success"`
}

// BatchSyncApprovalNodes 批量同步审批节点
func (h *approvalNodeHandler) BatchSyncApprovalNodes(c *gin.Context) {
	h.logger.Info("批量同步审批节点", "approvalDefCode", c.Query("approvalDefCode"))
	var req []request.CreateApprovalNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	approvalDefCode := c.Query("approvalDefCode")
	var results []*model.ApprovalNode
	for _, item := range req {
		node := &model.ApprovalNode{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        item.NodeCode,
			NodeName:        item.NodeName,
			NodeType:        item.NodeType,
			ApproverType:    item.ApproverType,
			ApproverConfig:  item.ApproverConfig,
			ConditionConfig: item.ConditionConfig,
			SortOrder:       item.SortOrder,
			Status:          model.ApprovalDefStatusNormal,
		}
		results = append(results, node)
	}

	h.logger.Info("批量同步审批节点", "results", results)

	message, err := h.approvalNodeService.BatchSyncNodes(approvalDefCode, results)
	if err != nil {
		h.logger.Error("批量同步审批节点失败", "error", err)
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": message,
	})
}
