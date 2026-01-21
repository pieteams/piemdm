package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"testing"
	"time"

	"piemdm/internal/model"
	"piemdm/internal/pkg/request"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ApprovalWorkflowTestSuite 审批流程集成测试套件
type ApprovalWorkflowTestSuite struct {
	suite.Suite
	router *gin.Engine
}

// SetupSuite 测试套件初始化
func (suite *ApprovalWorkflowTestSuite) SetupSuite() {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	suite.router = gin.New()

	// 添加基本的测试路由
	suite.setupTestRoutes()
}

// setupTestRoutes 设置测试路由
func (suite *ApprovalWorkflowTestSuite) setupTestRoutes() {
	api := suite.router.Group("/api/v1")

	// 模拟审批定义路由
	api.POST("/approval-defs", suite.mockCreateApprovalDef)
	api.POST("/approval-defs/activate", suite.mockActivateApprovalDef)

	// 模拟审批节点路由
	api.POST("/approval-nodes", suite.mockCreateApprovalNode)

	// 模拟审批实例路由
	api.POST("/approvals/start", suite.mockStartApproval)
	api.POST("/approvals/submit", suite.mockSubmitApproval)
	api.POST("/approvals/cancel", suite.mockCancelApproval)
	api.GET("/approvals/code/:code", suite.mockGetApprovalByCode)

	// 模拟审批任务路由
	api.GET("/approval-tasks/assignee/:assigneeId/pending", suite.mockGetPendingTasks)
	api.GET("/approval-tasks/assignee/:assigneeId/completed", suite.mockGetCompletedTasks)
	api.POST("/approval-tasks/:id/approve", suite.mockApproveTask)
	api.POST("/approval-tasks/:id/reject", suite.mockRejectTask)
	api.POST("/approval-tasks/:id/transfer", suite.mockTransferTask)
}

// 模拟数据存储
var (
	mockApprovalDefs   = make(map[string]*model.ApprovalDefinition)
	mockApprovalNodes  = make(map[string][]*model.ApprovalNode)
	mockApprovals      = make(map[string]*model.Approval)
	mockApprovalTasks  = make(map[string][]*model.ApprovalTask) // 待办任务
	mockCompletedTasks = make(map[string][]*model.ApprovalTask) // 已完成任务
)

// getTestApprovalNodes 根据SQL数据生成测试审批节点数据
func getTestApprovalNodes(approvalDefCode string) []*model.ApprovalNode {
	return []*model.ApprovalNode{
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "566fcc47-e977-40fb-b202-b91ed2caf37c",
			NodeName:        "提交",
			NodeType:        model.NodeTypeStart,
			SortOrder:       0,
			ApproverType:    model.ApproverTypeUsers,
			ApproverConfig:  "{}",
			Status:          "Active",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "ec7e63bc-b1b0-475e-8961-1facfd88db2c",
			NodeName:        "上级审批",
			NodeType:        model.NodeTypeApproval,
			SortOrder:       1,
			ApproverType:    model.ApproverTypeUsers,
			Status:          "Active",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "a88b9b61-490e-4507-afd0-89f18a6f121e",
			NodeName:        "条件分支",
			NodeType:        model.NodeTypeCondition,
			SortOrder:       2,
			ApproverType:    model.ApproverTypeUsers,
			ConditionConfig: `{"branches":[{"name":"条件分支 1","condition":{"fieldName":"price","operator":"eq","fieldValue":"3000"},"nodes":[{"nodeCode":"fa8b52fe-659d-43fb-812d-a9a6b0f3f6dc","nodeName":"财务审批","nodeType":"APPROVAL"}]},{"name":"其他情况","condition":{"fieldName":"","operator":"eq","fieldValue":""},"nodes":[{"nodeCode":"73e194ac-3de1-4545-91d7-4f727acca413","nodeName":"自动驳回","nodeType":"APPROVAL"}]}]}`,
			Status:          "Active",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "fa8b52fe-659d-43fb-812d-a9a6b0f3f6dc",
			NodeName:        "财务审批",
			NodeType:        model.NodeTypeApproval,
			SortOrder:       3,
			ApproverType:    model.ApproverTypeUsers,
			Status:          "Active",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "73e194ac-3de1-4545-91d7-4f727acca413",
			NodeName:        "自动驳回",
			NodeType:        model.NodeTypeApproval,
			SortOrder:       4,
			ApproverType:    model.ApproverTypeAutoReject,
			Status:          "Active",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "565a5219-23bc-4d74-af86-42ba2ff79d77",
			NodeName:        "抄送法务",
			NodeType:        model.NodeTypeCC,
			SortOrder:       5,
			ApproverType:    model.ApproverTypeUsers,
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "d96b927a-9a05-41ce-bb22-f9f4d0aa51b5",
			NodeName:        "结束",
			NodeType:        model.NodeTypeEnd,
			SortOrder:       6,
			ApproverType:    model.ApproverTypeUsers,
			Status:          "Active",
		},
	}
}

// Mock handlers
func (suite *ApprovalWorkflowTestSuite) mockCreateApprovalDef(c *gin.Context) {
	var req request.CreateApprovalDefRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approvalDef := &model.ApprovalDefinition{
		Code:        "TEST_APPROVAL_1749042000727964000",
		Name:        req.Name,
		Description: req.Description,
		FormData:    req.FormData,
		NodeList:    req.NodeList,
		Status:      model.ApprovalDefStatusNormal,
	}
	approvalDef.ID = 1

	mockApprovalDefs[approvalDef.Code] = approvalDef

	// 自动创建测试审批节点数据
	mockApprovalNodes[approvalDef.Code] = getTestApprovalNodes(approvalDef.Code)

	c.JSON(http.StatusOK, gin.H{"data": approvalDef})
}

func (suite *ApprovalWorkflowTestSuite) mockActivateApprovalDef(c *gin.Context) {
	var req request.ActivateApprovalDefRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "激活成功"})
}

func (suite *ApprovalWorkflowTestSuite) mockCreateApprovalNode(c *gin.Context) {
	var req request.CreateApprovalNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node := &model.ApprovalNode{
		ApprovalDefCode: req.ApprovalDefCode,
		NodeCode:        req.NodeCode,
		NodeName:        req.NodeName,
		NodeType:        req.NodeType,
		ApproverType:    req.ApproverType,
		ApproverConfig:  req.ApproverConfig,
		ConditionConfig: req.ConditionConfig,
		SortOrder:       req.SortOrder,
		Status:          "Active",
	}
	node.ID = uint(len(mockApprovalNodes[req.ApprovalDefCode]) + 1)

	mockApprovalNodes[req.ApprovalDefCode] = append(mockApprovalNodes[req.ApprovalDefCode], node)
	c.JSON(http.StatusOK, gin.H{"data": node})
}

func (suite *ApprovalWorkflowTestSuite) mockStartApproval(c *gin.Context) {
	var req request.StartApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	approval := &model.Approval{
		Code:            fmt.Sprintf("APPROVAL_%d", time.Now().Unix()),
		ApprovalDefCode: req.ApprovalDefCode,
		Title:           req.Title,
		FormData:        req.FormData, // 保存表单数据
		// EntityCode:      req.EntityCode,
		Status: model.ApprovalStatusPending,
	}
	approval.ID = 1

	mockApprovals[approval.Code] = approval
	c.JSON(http.StatusOK, gin.H{"data": approval})
}

func (suite *ApprovalWorkflowTestSuite) mockSubmitApproval(c *gin.Context) {
	var req request.SubmitApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if approval, exists := mockApprovals[req.Code]; exists {
		approval.Status = model.ApprovalStatusPending

		// 找到提交节点后的第一个需要处理的节点
		if nodes, nodeExists := mockApprovalNodes[approval.ApprovalDefCode]; nodeExists {
			suite.processFirstNodeAfterSubmit(approval, nodes)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "提交成功"})
}

func (suite *ApprovalWorkflowTestSuite) mockCancelApproval(c *gin.Context) {
	var req request.CancelApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if approval, exists := mockApprovals[req.Code]; exists {
		approval.Status = model.ApprovalStatusCanceled
	}

	c.JSON(http.StatusOK, gin.H{"message": "撤回成功"})
}

func (suite *ApprovalWorkflowTestSuite) mockGetApprovalByCode(c *gin.Context) {
	code := c.Param("code")
	if approval, exists := mockApprovals[code]; exists {
		c.JSON(http.StatusOK, gin.H{"data": approval})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "审批实例不存在"})
	}
}

func (suite *ApprovalWorkflowTestSuite) mockGetPendingTasks(c *gin.Context) {
	assigneeID := c.Param("assigneeId")
	tasks := mockApprovalTasks[assigneeID]

	var pendingTasks []*model.ApprovalTask
	for _, task := range tasks {
		if task.Status == model.TaskStatusPending {
			pendingTasks = append(pendingTasks, task)
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": pendingTasks})
}

func (suite *ApprovalWorkflowTestSuite) mockGetCompletedTasks(c *gin.Context) {
	assigneeID := c.Param("assigneeId")
	completedTasks := mockCompletedTasks[assigneeID]

	c.JSON(http.StatusOK, gin.H{"data": completedTasks})
}

func (suite *ApprovalWorkflowTestSuite) mockApproveTask(c *gin.Context) {
	taskID := c.Param("id")

	// 查找并处理任务
	for userID, tasks := range mockApprovalTasks {
		for i, task := range tasks {
			if fmt.Sprintf("%d", task.ID) == taskID {
				// 1. 修改当前task状态为已审批
				task.Status = model.TaskStatusApproved
				// now := time.Now()
				// task.CompletedAt = &now

				// 2. 将已审批的task移动到已完成任务列表
				mockCompletedTasks[userID] = append(mockCompletedTasks[userID], task)

				// 3. 从待办任务中移除
				mockApprovalTasks[userID] = append(tasks[:i], tasks[i+1:]...)

				// 4. 创建下一个task（通过处理下一个节点）
				suite.processNextNode(task)
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "审批成功"})
}

// processNextNode 处理下一个节点的逻辑
func (suite *ApprovalWorkflowTestSuite) processNextNode(currentTask *model.ApprovalTask) {
	approval := mockApprovals[currentTask.ApprovalCode]
	if approval == nil {
		return
	}

	nodes := mockApprovalNodes[approval.ApprovalDefCode]
	if nodes == nil {
		return
	}

	// 检查当前节点是否是条件分支中的节点，如果是则需要跳过条件分支中的其他节点
	nextNode := suite.findNextNodeAfterConditionBranch(currentTask.NodeCode, nodes)
	if nextNode == nil {
		// 按正常顺序找下一个节点
		nextNode = suite.findNextNodeByOrder(currentTask.NodeCode, nodes)
	}

	if nextNode == nil {
		// 没有下一个节点，审批完成
		approval.Status = model.ApprovalStatusApproved
		return
	}

	// 处理不同类型的节点
	switch nextNode.NodeType {
	case "CONDITION":
		// 条件分支节点
		suite.processConditionNode(nextNode, approval)
	case "APPROVAL":
		// 直接审批节点
		suite.createApprovalTask(nextNode, approval)
	case "CC":
		// 抄送节点 - 发送通知后继续下一个
		suite.processCCNode(nextNode, approval)
	case "END":
		// 结束节点
		approval.Status = model.ApprovalStatusApproved
	}
}

// processConditionNode 处理条件分支节点
func (suite *ApprovalWorkflowTestSuite) processConditionNode(conditionNode *model.ApprovalNode, approval *model.Approval) {
	// 解析条件配置
	var conditionConfig struct {
		Branches []struct {
			Name      string `json:"name"`
			Condition struct {
				FieldName  string `json:"fieldName"`
				Operator   string `json:"operator"`
				FieldValue string `json:"fieldValue"`
			} `json:"condition"`
			Nodes []struct {
				NodeCode string `json:"nodeCode"`
				NodeName string `json:"nodeName"`
				NodeType string `json:"nodeType"`
			} `json:"nodes"`
		} `json:"branches"`
	}

	if err := json.Unmarshal([]byte(conditionNode.ConditionConfig), &conditionConfig); err != nil {
		// 解析失败，流程结束
		approval.Status = model.ApprovalStatusRejected
		return
	}

	// 解析表单数据
	var formData map[string]any
	if err := json.Unmarshal([]byte(approval.FormData), &formData); err != nil {
		// 解析失败，流程结束
		approval.Status = model.ApprovalStatusRejected
		return
	}

	// 遍历分支条件，找到匹配的分支
	for _, branch := range conditionConfig.Branches {
		if suite.evaluateCondition(branch.Condition, formData) {
			// 条件匹配，执行该分支的节点
			suite.executeBranchNodes(branch.Nodes, approval)
			return
		}
	}

	// 没有匹配的分支，流程结束
	approval.Status = model.ApprovalStatusRejected
}

// evaluateCondition 评估条件是否满足
func (suite *ApprovalWorkflowTestSuite) evaluateCondition(condition struct {
	FieldName  string `json:"fieldName"`
	Operator   string `json:"operator"`
	FieldValue string `json:"fieldValue"`
}, formData map[string]any,
) bool {
	// 如果条件为空（如"其他情况"），则作为默认分支
	if condition.FieldName == "" {
		return true
	}

	// 获取表单字段值
	fieldValue, exists := formData[condition.FieldName]
	if !exists {
		return false
	}

	// 转换为字符串进行比较
	fieldValueStr := fmt.Sprintf("%v", fieldValue)

	// 根据操作符进行判断
	switch condition.Operator {
	case "eq":
		return fieldValueStr == condition.FieldValue
	case "ne":
		return fieldValueStr != condition.FieldValue
	case "gt":
		// 数值比较
		if fieldVal, err := strconv.ParseFloat(fieldValueStr, 64); err == nil {
			if condVal, err := strconv.ParseFloat(condition.FieldValue, 64); err == nil {
				return fieldVal > condVal
			}
		}
		return false
	case "lt":
		// 数值比较
		if fieldVal, err := strconv.ParseFloat(fieldValueStr, 64); err == nil {
			if condVal, err := strconv.ParseFloat(condition.FieldValue, 64); err == nil {
				return fieldVal < condVal
			}
		}
		return false
	default:
		return false
	}
}

// executeBranchNodes 执行分支节点
func (suite *ApprovalWorkflowTestSuite) executeBranchNodes(branchNodes []struct {
	NodeCode string `json:"nodeCode"`
	NodeName string `json:"nodeName"`
	NodeType string `json:"nodeType"`
}, approval *model.Approval,
) {
	if len(branchNodes) == 0 {
		return
	}

	// 执行分支中的第一个节点
	targetNodeCode := branchNodes[0].NodeCode
	nodes := mockApprovalNodes[approval.ApprovalDefCode]

	for _, node := range nodes {
		if node.NodeCode == targetNodeCode {
			// 根据节点类型执行相应操作
			switch node.NodeType {
			case "APPROVAL":
				// 检查是否为自动驳回类型
				if node.ApproverType == "AUTO_REJECT" {
					approval.Status = model.ApprovalStatusRejected
				} else {
					suite.createApprovalTask(node, approval)
				}
			case "CC":
				// 抄送节点，跳过到下一个
				suite.processNextNodeAfter(node, approval)
			case "END":
				// 结束节点
				approval.Status = model.ApprovalStatusApproved
			}
			break
		}
	}
}

// createApprovalTask 创建审批任务
func (suite *ApprovalWorkflowTestSuite) createApprovalTask(node *model.ApprovalNode, approval *model.Approval) {
	// 解析审批人配置
	var assigneeID string
	if node.ApproverType == "USERS" {
		// 检查是否为自动驳回节点（通过节点名称或配置判断）
		if node.NodeName == "自动驳回" || node.NodeCode == "73e194ac-3de1-4545-91d7-4f727acca413" || node.NodeCode == "reject-001" {
			// 自动驳回
			approval.Status = model.ApprovalStatusRejected
			return
		}

		// 从ApproverConfig中解析用户ID
		if node.NodeCode == "fa8b52fe-659d-43fb-812d-a9a6b0f3f6dc" || node.NodeCode == "finance-001" {
			assigneeID = "pp2" // 财务审批人
		} else {
			assigneeID = "pp1" // 默认审批人
		}
	} else if node.ApproverType == "AUTO_REJECT" {
		// 自动驳回
		approval.Status = model.ApprovalStatusRejected
		return
	}

	if assigneeID != "" {
		task := &model.ApprovalTask{
			ApprovalCode: approval.Code,
			NodeCode:     node.NodeCode,
			AssigneeID:   assigneeID,
			Status:       model.TaskStatusPending,
		}
		task.ID = uint(len(mockApprovalTasks) + 1)
		mockApprovalTasks[assigneeID] = append(mockApprovalTasks[assigneeID], task)
	}
}

// processCCNode 处理抄送节点
func (suite *ApprovalWorkflowTestSuite) processCCNode(ccNode *model.ApprovalNode, approval *model.Approval) {
	// 1. 发送抄送通知（模拟）
	suite.sendCCNotification(ccNode, approval)

	// 2. 继续处理下一个节点
	nodes := mockApprovalNodes[approval.ApprovalDefCode]
	if nodes == nil {
		return
	}

	// 找到抄送节点后的下一个节点
	nextNode := suite.findNextNodeByOrder(ccNode.NodeCode, nodes)
	if nextNode == nil {
		// 没有下一个节点，流程结束
		approval.Status = model.ApprovalStatusApproved
		return
	}

	// 递归处理下一个节点
	switch nextNode.NodeType {
	case "APPROVAL":
		suite.createApprovalTask(nextNode, approval)
	case "CONDITION":
		suite.processConditionNode(nextNode, approval)
	case "CC":
		suite.processCCNode(nextNode, approval)
	case "END":
		approval.Status = model.ApprovalStatusApproved
	}
}

// sendCCNotification 发送抄送通知（模拟实现）
func (suite *ApprovalWorkflowTestSuite) sendCCNotification(ccNode *model.ApprovalNode, approval *model.Approval) {
	// 解析抄送配置
	var notifyConfig struct {
		CCUsers  string `json:"ccUsers"`
		CCTiming string `json:"ccTiming"`
	}

	if err := json.Unmarshal([]byte(ccNode.ApproverConfig), &notifyConfig); err != nil {
		return // 解析失败，跳过通知
	}

	// 模拟发送通知日志
	fmt.Printf("INFO: 发送抄送通知 - 节点: %s, 抄送人员: %s, 审批: %s\n",
		ccNode.NodeName, notifyConfig.CCUsers, approval.Code)
}

// processNextNodeAfter 处理指定节点后的下一个节点
func (suite *ApprovalWorkflowTestSuite) processNextNodeAfter(currentNode *model.ApprovalNode, approval *model.Approval) {
	nodes := mockApprovalNodes[approval.ApprovalDefCode]
	for i, node := range nodes {
		if node.NodeCode == currentNode.NodeCode {
			if i+1 < len(nodes) {
				nextNode := nodes[i+1]
				if nextNode.NodeType == "END" {
					approval.Status = model.ApprovalStatusApproved
				}
			} else {
				// CC节点后没有下一个节点，审批完成
				approval.Status = model.ApprovalStatusApproved
			}
			break
		}
	}
}

func (suite *ApprovalWorkflowTestSuite) mockRejectTask(c *gin.Context) {
	taskID := c.Param("id")

	// 更新任务状态
	for _, tasks := range mockApprovalTasks {
		for _, task := range tasks {
			if fmt.Sprintf("%d", task.ID) == taskID {
				task.Status = model.TaskStatusRejected

				// 更新审批状态
				if approval, exists := mockApprovals[task.ApprovalCode]; exists {
					approval.Status = model.ApprovalStatusRejected
				}
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "拒绝成功"})
}

func (suite *ApprovalWorkflowTestSuite) mockTransferTask(c *gin.Context) {
	taskID := c.Param("id")

	var req request.TransferTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转交任务
	for assigneeID, tasks := range mockApprovalTasks {
		for i, task := range tasks {
			if fmt.Sprintf("%d", task.ID) == taskID {
				// 从原审批人移除
				mockApprovalTasks[assigneeID] = append(tasks[:i], tasks[i+1:]...)

				// 添加到新审批人
				task.AssigneeID = req.ToUserID
				mockApprovalTasks[req.ToUserID] = append(mockApprovalTasks[req.ToUserID], task)
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "转交成功"})
}

// SetupTest 每个测试用例前的准备
func (suite *ApprovalWorkflowTestSuite) SetupTest() {
	// 清理模拟数据
	mockApprovalDefs = make(map[string]*model.ApprovalDefinition)
	mockApprovals = make(map[string]*model.Approval)
	mockApprovalTasks = make(map[string][]*model.ApprovalTask)
	mockCompletedTasks = make(map[string][]*model.ApprovalTask)
}

// TestCompleteApprovalWorkflow 测试完整的审批流程
func (suite *ApprovalWorkflowTestSuite) TestCompleteApprovalWorkflow() {
	t := suite.T()

	// 1. 创建审批定义
	approvalDefCode := suite.createApprovalDefinition(t)
	assert.NotEmpty(t, approvalDefCode)

	// 2. 创建审批节点
	suite.createApprovalNodes(t, approvalDefCode)

	// 3. 激活审批定义
	suite.activateApprovalDefinition(t, approvalDefCode)

	// 4. 启动审批流程（满足条件，进入财务审批）
	approvalCode := suite.startApprovalProcessWithCondition(t, approvalDefCode, true)
	assert.NotEmpty(t, approvalCode)

	// 5. 提交审批申请
	suite.submitApproval(t, approvalCode)

	// 6. 查询待办任务（上级审批人pp1）
	tasks := suite.getPendingTasks(t, "pp1")
	assert.NotEmpty(t, tasks)

	// 7. 审批任务（上级审批）
	for _, task := range tasks {
		suite.approveTask(t, task.ID, "pp1")
	}

	// 8. 查询财务审批任务
	financeTasks := suite.getPendingTasks(t, "pp2")
	if len(financeTasks) > 0 {
		// 财务审批
		for _, task := range financeTasks {
			suite.approveTask(t, task.ID, "pp2")
		}
	}

	// 9. 验证审批完成
	approval := suite.getApprovalByCode(t, approvalCode)
	assert.Equal(t, model.ApprovalStatusApproved, approval.Status)
}

// TestApprovalRejectionWorkflow 测试审批拒绝流程
func (suite *ApprovalWorkflowTestSuite) TestApprovalRejectionWorkflow() {
	t := suite.T()

	// 1. 创建审批定义和节点
	approvalDefCode := suite.createApprovalDefinition(t)
	suite.createApprovalNodes(t, approvalDefCode)
	suite.activateApprovalDefinition(t, approvalDefCode)

	// 2. 启动审批流程
	approvalCode := suite.startApprovalProcess(t, approvalDefCode)
	suite.submitApproval(t, approvalCode)

	// 3. 拒绝审批任务
	tasks := suite.getPendingTasks(t, "pp1")
	assert.NotEmpty(t, tasks)

	suite.rejectTask(t, tasks[0].ID, "pp1", "不符合要求")

	// 4. 验证审批被拒绝
	approval := suite.getApprovalByCode(t, approvalCode)
	assert.Equal(t, model.ApprovalStatusRejected, approval.Status)
}

// TestApprovalCancellationWorkflow 测试审批撤回流程
func (suite *ApprovalWorkflowTestSuite) TestApprovalCancellationWorkflow() {
	t := suite.T()

	// 1. 创建审批定义和节点
	approvalDefCode := suite.createApprovalDefinition(t)
	suite.createApprovalNodes(t, approvalDefCode)
	suite.activateApprovalDefinition(t, approvalDefCode)

	// 2. 启动审批流程
	approvalCode := suite.startApprovalProcess(t, approvalDefCode)
	suite.submitApproval(t, approvalCode)

	// 3. 撤回审批申请
	suite.cancelApproval(t, approvalCode, "申请有误，需要重新提交")

	// 4. 验证审批被撤回
	approval := suite.getApprovalByCode(t, approvalCode)
	assert.Equal(t, model.ApprovalStatusCanceled, approval.Status)
}

// TestConditionBranchAutoRejectWorkflow 测试条件分支自动驳回流程
func (suite *ApprovalWorkflowTestSuite) TestConditionBranchAutoRejectWorkflow() {
	t := suite.T()

	// 1. 创建审批定义和节点
	approvalDefCode := suite.createApprovalDefinition(t)
	suite.createApprovalNodes(t, approvalDefCode)
	suite.activateApprovalDefinition(t, approvalDefCode)

	// 2. 启动审批流程（模拟不满足条件的情况）
	approvalCode := suite.startApprovalProcessWithCondition(t, approvalDefCode, false)
	suite.submitApproval(t, approvalCode)

	// 3. 上级审批通过
	tasks := suite.getPendingTasks(t, "pp1")
	assert.NotEmpty(t, tasks)
	suite.approveTask(t, tasks[0].ID, "pp1")

	// 4. 验证审批被自动驳回（条件分支进入自动驳回节点）
	approval := suite.getApprovalByCode(t, approvalCode)
	assert.Equal(t, model.ApprovalStatusRejected, approval.Status)
}

// TestDirectConditionBranchWorkflow 测试提交后直接进入条件分支的流程
func (suite *ApprovalWorkflowTestSuite) TestDirectConditionBranchWorkflow() {
	t := suite.T()

	// 创建一个特殊的审批定义：提交 -> 条件分支 -> 财务审批/自动驳回
	approvalDefCode := suite.createApprovalDefinition(t)
	suite.createDirectConditionNodes(t, approvalDefCode)
	suite.activateApprovalDefinition(t, approvalDefCode)

	// 启动审批流程（满足条件，应该直接进入财务审批）
	approvalCode := suite.startApprovalProcessWithCondition(t, approvalDefCode, true)
	suite.submitApproval(t, approvalCode)

	// 验证直接创建了财务审批任务
	financeTasks := suite.getPendingTasks(t, "pp2")
	assert.NotEmpty(t, financeTasks, "应该直接创建财务审批任务")

	// 完成财务审批
	suite.approveTask(t, financeTasks[0].ID, "pp2")

	// 验证审批完成
	approval := suite.getApprovalByCode(t, approvalCode)
	assert.Equal(t, model.ApprovalStatusApproved, approval.Status)
}

// TestCCNodeWorkflow 测试抄送节点流程
func (suite *ApprovalWorkflowTestSuite) TestCCNodeWorkflow() {
	t := suite.T()

	// 创建一个包含CC节点的审批定义：提交 -> 审批 -> 抄送 -> 结束
	approvalDefCode := suite.createApprovalDefinition(t)
	suite.createCCNodeWorkflow(t, approvalDefCode)
	suite.activateApprovalDefinition(t, approvalDefCode)

	// 启动审批流程
	approvalCode := suite.startApprovalProcess(t, approvalDefCode)
	suite.submitApproval(t, approvalCode)

	// 完成审批任务
	tasks := suite.getPendingTasks(t, "pp1")
	assert.NotEmpty(t, tasks, "应该有审批任务")
	suite.approveTask(t, tasks[0].ID, "pp1")

	// 验证审批完成（CC节点应该自动发送通知并完成流程）
	approval := suite.getApprovalByCode(t, approvalCode)
	assert.Equal(t, model.ApprovalStatusApproved, approval.Status, "CC节点处理后应该完成审批")
}

// TestTaskStatusManagement 测试任务状态管理
func (suite *ApprovalWorkflowTestSuite) TestTaskStatusManagement() {
	t := suite.T()

	// 创建审批定义和节点
	approvalDefCode := suite.createApprovalDefinition(t)
	suite.createApprovalNodes(t, approvalDefCode)
	suite.activateApprovalDefinition(t, approvalDefCode)

	// 启动审批流程
	approvalCode := suite.startApprovalProcess(t, approvalDefCode)
	suite.submitApproval(t, approvalCode)

	// 验证初始状态：pp1有待办任务
	pendingTasks := suite.getPendingTasks(t, "pp1")
	assert.Len(t, pendingTasks, 1, "pp1应该有1个待办任务")
	assert.Equal(t, model.TaskStatusPending, pendingTasks[0].Status)

	// 验证初始状态：pp1没有已完成任务
	completedTasks := suite.getCompletedTasks(t, "pp1")
	assert.Len(t, completedTasks, 0, "pp1初始应该没有已完成任务")

	// 审批任务
	taskID := pendingTasks[0].ID
	suite.approveTask(t, taskID, "pp1")

	// 验证审批后状态：pp1的待办任务减少
	pendingTasksAfter := suite.getPendingTasks(t, "pp1")
	assert.Len(t, pendingTasksAfter, 0, "pp1审批后应该没有待办任务")

	// 验证审批后状态：pp1有已完成任务
	completedTasksAfter := suite.getCompletedTasks(t, "pp1")
	assert.Len(t, completedTasksAfter, 1, "pp1审批后应该有1个已完成任务")
	assert.Equal(t, model.TaskStatusApproved, completedTasksAfter[0].Status)

	// 验证下一个任务被创建（如果流程继续）
	// 检查是否有其他用户的待办任务（如pp2的财务审批）
	pp2PendingTasks := suite.getPendingTasks(t, "pp2")
	if len(pp2PendingTasks) > 0 {
		assert.Equal(t, model.TaskStatusPending, pp2PendingTasks[0].Status, "下一个任务应该是待办状态")
	}
}

// TestTaskTransferWorkflow 测试任务转交流程
func (suite *ApprovalWorkflowTestSuite) TestTaskTransferWorkflow() {
	t := suite.T()

	// 1. 创建审批定义和节点
	approvalDefCode := suite.createApprovalDefinition(t)
	suite.createApprovalNodes(t, approvalDefCode)
	suite.activateApprovalDefinition(t, approvalDefCode)

	// 2. 启动审批流程
	approvalCode := suite.startApprovalProcess(t, approvalDefCode)
	suite.submitApproval(t, approvalCode)

	// 3. 转交任务
	tasks := suite.getPendingTasks(t, "pp1")
	assert.NotEmpty(t, tasks)

	suite.transferTask(t, tasks[0].ID, "pp1", "pp2", "临时出差，转交处理")

	// 4. 验证任务转交成功
	newTasks := suite.getPendingTasks(t, "pp2")
	assert.NotEmpty(t, newTasks)
	assert.Equal(t, tasks[0].ApprovalCode, newTasks[0].ApprovalCode)
}

// 辅助方法：创建审批定义
func (suite *ApprovalWorkflowTestSuite) createApprovalDefinition(t *testing.T) string {
	reqBody := request.CreateApprovalDefRequest{
		Name:           "测试审批流程",
		Description:    "这是一个测试审批流程",
		FormData:       `{"fields":[{"name":"reason","type":"text","label":"申请原因","required":true}]}`,
		NodeList:       `{"nodes":[{"id":"start","type":"start"},{"id":"approve1","type":"approval"},{"id":"end","type":"end"}]}`,
		Status:         "Normal",
		ApprovalSystem: "Builtin",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/approval-defs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user_id", "admin")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Create approval def failed: status=%d, body=%s", w.Code, w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data model.ApprovalDefinition `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	return response.Data.Code
}

// 辅助方法：创建审批节点
func (suite *ApprovalWorkflowTestSuite) createApprovalNodes(t *testing.T, approvalDefCode string) {
	// 按照SQL数据中的顺序创建所有节点
	nodes := []request.CreateApprovalNodeRequest{
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "566fcc47-e977-40fb-b202-b91ed2caf37c",
			NodeName:        "提交",
			NodeType:        "START",
			ApproverType:    "USERS",
			ApproverConfig:  "{}",
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 0,
			// IsRequired:       true,
			// AllowReject:      false,
			// AllowTransfer:    false,
			// AutoApprove:      false,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "ec7e63bc-b1b0-475e-8961-1facfd88db2c",
			NodeName:        "上级审批",
			NodeType:        "APPROVAL",
			ApproverType:    "USERS",
			ApproverConfig:  `{"type":"USERS","users":["pp1"],"mode":"OR"}`,
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 1,
			// IsRequired:       true,
			// AllowReject:      true,
			// AllowTransfer:    true,
			// AutoApprove:      false,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "a88b9b61-490e-4507-afd0-89f18a6f121e",
			NodeName:        "条件分支",
			NodeType:        "CONDITION",
			ApproverType:    "USERS",
			ApproverConfig:  "{}",
			// ApprovalMode:     "OR",
			ConditionConfig: `{"branches":[{"name":"条件分支 1","condition":{"fieldName":"price","operator":"eq","fieldValue":"3000"},"nodes":[{"nodeCode":"fa8b52fe-659d-43fb-812d-a9a6b0f3f6dc","nodeName":"财务审批","nodeType":"APPROVAL"}]},{"name":"其他情况","condition":{"fieldName":"","operator":"eq","fieldValue":""},"nodes":[{"nodeCode":"73e194ac-3de1-4545-91d7-4f727acca413","nodeName":"自动驳回","nodeType":"APPROVAL"}]}]}`,
			// TimeoutHours:     72,
			SortOrder: 2,
			// IsRequired:       true,
			// AllowReject:      false,
			// AllowTransfer:    false,
			// AutoApprove:      false,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "fa8b52fe-659d-43fb-812d-a9a6b0f3f6dc",
			NodeName:        "财务审批",
			NodeType:        "APPROVAL",
			ApproverType:    "USERS",
			ApproverConfig:  `{"type":"USERS","users":["pp2"],"mode":"OR"}`,
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 3,
			// IsRequired:       true,
			// AllowReject:      true,
			// AllowTransfer:    true,
			// AutoApprove:      false,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "73e194ac-3de1-4545-91d7-4f727acca413",
			NodeName:        "自动驳回",
			NodeType:        "APPROVAL",
			ApproverType:    "USERS",
			ApproverConfig:  `{"type":"USERS","users":[],"mode":"OR"}`,
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 4,
			// IsRequired:       true,
			// AllowReject:      true,
			// AllowTransfer:    true,
			// AutoApprove:      true,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "565a5219-23bc-4d74-af86-42ba2ff79d77",
			NodeName:        "抄送法务",
			NodeType:        "CC",
			ApproverType:    "USERS",
			ApproverConfig:  "{}",
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 5,
			// IsRequired:       true,
			// AllowReject:      false,
			// AllowTransfer:    false,
			// AutoApprove:      true,
			// NotifyConfig:     `{"ccUsers":"ss1，ss2","ccTiming":"final"}`,
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "d96b927a-9a05-41ce-bb22-f9f4d0aa51b5",
			NodeName:        "结束",
			NodeType:        "END",
			ApproverType:    "USERS",
			ApproverConfig:  "{}",
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 6,
			// IsRequired:       true,
			// AllowReject:      false,
			// AllowTransfer:    false,
			// AutoApprove:      false,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
	}

	// 创建所有节点
	for _, node := range nodes {
		suite.createNode(t, node)
	}
}

// 辅助方法：创建单个节点
func (suite *ApprovalWorkflowTestSuite) createNode(t *testing.T, nodeReq request.CreateApprovalNodeRequest) {
	body, _ := json.Marshal(nodeReq)
	req := httptest.NewRequest("POST", "/api/v1/approval-nodes", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user_id", "admin")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Create node failed for %s: status=%d, body=%s", nodeReq.NodeName, w.Code, w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)
}

// 辅助方法：激活审批定义
func (suite *ApprovalWorkflowTestSuite) activateApprovalDefinition(t *testing.T, code string) {
	reqBody := request.ActivateApprovalDefRequest{
		ID: 1, // 简化处理，实际应该查询获取
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/approval-defs/activate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user_id", "admin")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 辅助方法：启动审批流程（带条件参数）
func (suite *ApprovalWorkflowTestSuite) startApprovalProcessWithCondition(t *testing.T, approvalDefCode string, meetCondition bool) string {
	var formData string
	if meetCondition {
		// 满足条件：price = 3000，进入财务审批
		formData = `{"reason":"测试申请原因","price":"3000"}`
	} else {
		// 不满足条件：price != 3000，进入自动驳回
		formData = `{"reason":"测试申请原因","price":"1000"}`
	}

	reqBody := request.StartApprovalRequest{
		ApprovalDefCode: approvalDefCode,
		Title:           "测试审批申请",
		EntityID:        "test_entity_001",
		FormData:        formData,
		Priority:        1,
		Urgency:         "NORMAL",
		Remark:          "测试用途",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/approvals/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user_id", "test_user")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Start approval failed: status=%d, body=%s", w.Code, w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data model.Approval `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	return response.Data.Code
}

// 辅助方法：启动审批流程
func (suite *ApprovalWorkflowTestSuite) startApprovalProcess(t *testing.T, approvalDefCode string) string {
	reqBody := request.StartApprovalRequest{
		ApprovalDefCode: approvalDefCode,
		Title:           "测试审批申请",
		EntityID:        "test_entity_001",
		FormData:        `{"reason":"测试申请原因"}`,
		Priority:        1,
		Urgency:         "NORMAL",
		Remark:          "测试用途",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/approvals/start", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user_id", "test_user")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Start approval failed: status=%d, body=%s", w.Code, w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data model.Approval `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	return response.Data.Code
}

// 辅助方法：提交审批申请
func (suite *ApprovalWorkflowTestSuite) submitApproval(t *testing.T, approvalCode string) {
	reqBody := request.SubmitApprovalRequest{
		Code: approvalCode,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/approvals/submit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user_id", "test_user")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 辅助方法：获取已完成任务
func (suite *ApprovalWorkflowTestSuite) getCompletedTasks(t *testing.T, assigneeID string) []*model.ApprovalTask {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/approval-tasks/assignee/%s/completed", assigneeID), nil)
	req.Header.Set("user_id", assigneeID)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []*model.ApprovalTask `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	return response.Data
}

// 辅助方法：获取待办任务
func (suite *ApprovalWorkflowTestSuite) getPendingTasks(t *testing.T, assigneeID string) []*model.ApprovalTask {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/approval-tasks/assignee/%s/pending", assigneeID), nil)
	req.Header.Set("user_id", assigneeID)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data []*model.ApprovalTask `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	return response.Data
}

// 辅助方法：审批任务
func (suite *ApprovalWorkflowTestSuite) approveTask(t *testing.T, taskID uint, assigneeID string) {
	reqBody := request.ApprovalTaskActionRequest{
		Action:  "APPROVE",
		Comment: "同意",
		Reason:  "符合要求",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/approval-tasks/%d/approve", taskID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user_id", assigneeID)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 辅助方法：拒绝任务
func (suite *ApprovalWorkflowTestSuite) rejectTask(t *testing.T, taskID uint, assigneeID, reason string) {
	reqBody := request.ApprovalTaskActionRequest{
		Action:  "REJECT",
		Comment: "拒绝",
		Reason:  reason,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/approval-tasks/%d/reject", taskID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user_id", assigneeID)

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 辅助方法：撤回审批
func (suite *ApprovalWorkflowTestSuite) cancelApproval(t *testing.T, approvalCode, reason string) {
	reqBody := request.CancelApprovalRequest{
		Code:   approvalCode,
		Reason: reason,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/approvals/cancel", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user_id", "test_user")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 辅助方法：转交任务
func (suite *ApprovalWorkflowTestSuite) transferTask(t *testing.T, taskID uint, fromUserID, toUserID, reason string) {
	reqBody := request.TransferTaskRequest{
		TaskID:     taskID,
		ToUserID:   toUserID,
		ToUserName: "用户2",
		Reason:     reason,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/approval-tasks/%d/transfer", taskID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user_id", fromUserID)
	req.Header.Set("user_name", "用户1")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 辅助方法：根据编码获取审批实例
func (suite *ApprovalWorkflowTestSuite) getApprovalByCode(t *testing.T, approvalCode string) *model.Approval {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/approvals/code/%s", approvalCode), nil)
	req.Header.Set("user_id", "test_user")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data model.Approval `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	return &response.Data
}

// findNextNodeAfterConditionBranch 查找条件分支节点完成后的下一个节点
func (suite *ApprovalWorkflowTestSuite) findNextNodeAfterConditionBranch(currentNodeCode string, nodes []*model.ApprovalNode) *model.ApprovalNode {
	// 找到所有条件分支节点
	var conditionNodes []*model.ApprovalNode
	for _, node := range nodes {
		if node.NodeType == "CONDITION" {
			conditionNodes = append(conditionNodes, node)
		}
	}

	// 检查当前节点是否在某个条件分支的配置中
	for _, conditionNode := range conditionNodes {
		branchNodeCodes := suite.extractBranchNodeCodes(conditionNode.ConditionConfig)

		// 如果当前节点在条件分支的节点列表中
		if slices.Contains(branchNodeCodes, currentNodeCode) {
			// 找到条件分支节点后的下一个节点（按SortOrder）
			return suite.findNextNodeAfterCondition(conditionNode, nodes)
		}
	}

	return nil // 当前节点不是条件分支中的节点
}

// findNextNodeByOrder 按SortOrder顺序查找下一个节点
func (suite *ApprovalWorkflowTestSuite) findNextNodeByOrder(currentNodeCode string, nodes []*model.ApprovalNode) *model.ApprovalNode {
	for i, node := range nodes {
		if node.NodeCode == currentNodeCode {
			// 找到当前节点，获取下一个节点
			if i+1 < len(nodes) {
				return nodes[i+1]
			}
			break
		}
	}
	return nil
}

// extractBranchNodeCodes 提取条件分支配置中的所有节点代码
func (suite *ApprovalWorkflowTestSuite) extractBranchNodeCodes(conditionConfig string) []string {
	var config struct {
		Branches []struct {
			Nodes []struct {
				NodeCode string `json:"nodeCode"`
			} `json:"nodes"`
		} `json:"branches"`
	}

	if err := json.Unmarshal([]byte(conditionConfig), &config); err != nil {
		return nil
	}

	var nodeCodes []string
	for _, branch := range config.Branches {
		for _, node := range branch.Nodes {
			nodeCodes = append(nodeCodes, node.NodeCode)
		}
	}
	return nodeCodes
}

// findNextNodeAfterCondition 查找条件分支节点后的下一个节点
func (suite *ApprovalWorkflowTestSuite) findNextNodeAfterCondition(conditionNode *model.ApprovalNode, nodes []*model.ApprovalNode) *model.ApprovalNode {
	// 找到条件分支节点在列表中的位置
	for i, node := range nodes {
		if node.NodeCode == conditionNode.NodeCode {
			// 跳过所有条件分支中定义的节点，找到条件分支后的第一个非分支节点
			branchNodeCodes := suite.extractBranchNodeCodes(conditionNode.ConditionConfig)

			for j := i + 1; j < len(nodes); j++ {
				nextNode := nodes[j]
				// 检查这个节点是否在条件分支的节点列表中
				if !slices.Contains(branchNodeCodes, nextNode.NodeCode) {
					return nextNode
				}
			}
			break
		}
	}
	return nil
}

// createDirectConditionNodes 创建直接条件分支的节点（用于测试）
func (suite *ApprovalWorkflowTestSuite) createDirectConditionNodes(t *testing.T, approvalDefCode string) {
	// 创建简化的节点：提交 -> 条件分支 -> 财务审批/自动驳回 -> 结束
	nodes := []request.CreateApprovalNodeRequest{
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "start-001",
			NodeName:        "提交",
			NodeType:        "START",
			ApproverType:    "USERS",
			ApproverConfig:  "{}",
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 0,
			// IsRequired:       true,
			// AllowReject:      false,
			// AllowTransfer:    false,
			// AutoApprove:      false,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "condition-001",
			NodeName:        "条件分支",
			NodeType:        "CONDITION",
			ApproverType:    "USERS",
			ApproverConfig:  "{}",
			// ApprovalMode:     "OR",
			ConditionConfig: `{"branches":[{"name":"条件分支 1","condition":{"fieldName":"price","operator":"eq","fieldValue":"3000"},"nodes":[{"nodeCode":"finance-001","nodeName":"财务审批","nodeType":"APPROVAL"}]},{"name":"其他情况","condition":{"fieldName":"","operator":"eq","fieldValue":""},"nodes":[{"nodeCode":"reject-001","nodeName":"自动驳回","nodeType":"APPROVAL"}]}]}`,
			// TimeoutHours:     72,
			SortOrder: 1,
			// IsRequired:       true,
			// AllowReject:      false,
			// AllowTransfer:    false,
			// AutoApprove:      false,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "finance-001",
			NodeName:        "财务审批",
			NodeType:        "APPROVAL",
			ApproverType:    "USERS",
			ApproverConfig:  `{"type":"USERS","users":["pp2"],"mode":"OR"}`,
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 2,
			// IsRequired:       true,
			// AllowReject:      true,
			// AllowTransfer:    true,
			// AutoApprove:      false,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "reject-001",
			NodeName:        "自动驳回",
			NodeType:        "APPROVAL",
			ApproverType:    "USERS",
			ApproverConfig:  `{"type":"USERS","users":[],"mode":"OR"}`,
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 3,
			// IsRequired:       true,
			// AllowReject:      true,
			// AllowTransfer:    true,
			// AutoApprove:      true,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "end-001",
			NodeName:        "结束",
			NodeType:        "END",
			ApproverType:    "USERS",
			ApproverConfig:  "{}",
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 4,
			// IsRequired:       true,
			// AllowReject:      false,
			// AllowTransfer:    false,
			// AutoApprove:      false,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
	}

	// 清除原有的节点数据，使用新的节点
	mockApprovalNodes[approvalDefCode] = []*model.ApprovalNode{}

	// 创建所有节点
	for _, node := range nodes {
		suite.createNode(t, node)
	}
}

// createCCNodeWorkflow 创建包含CC节点的工作流（用于测试）
func (suite *ApprovalWorkflowTestSuite) createCCNodeWorkflow(t *testing.T, approvalDefCode string) {
	// 创建简化的节点：提交 -> 审批 -> 抄送 -> 结束
	nodes := []request.CreateApprovalNodeRequest{
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "start-cc-001",
			NodeName:        "提交",
			NodeType:        "START",
			ApproverType:    "USERS",
			ApproverConfig:  "{}",
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 0,
			// IsRequired:       true,
			// AllowReject:      false,
			// AllowTransfer:    false,
			// AutoApprove:      false,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "approval-cc-001",
			NodeName:        "主管审批",
			NodeType:        "APPROVAL",
			ApproverType:    "USERS",
			ApproverConfig:  `{"type":"USERS","users":["pp1"],"mode":"OR"}`,
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 1,
			// IsRequired:       true,
			// AllowReject:      true,
			// AllowTransfer:    true,
			// AutoApprove:      false,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "cc-001",
			NodeName:        "抄送财务",
			NodeType:        "CC",
			ApproverType:    "USERS",
			ApproverConfig:  "{}",
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 2,
			// IsRequired:       true,
			// AllowReject:      false,
			// AllowTransfer:    false,
			// AutoApprove:      true,
			// NotifyConfig:     `{"ccUsers":"finance1,finance2","ccTiming":"final"}`,
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
		{
			ApprovalDefCode: approvalDefCode,
			NodeCode:        "end-cc-001",
			NodeName:        "结束",
			NodeType:        "END",
			ApproverType:    "USERS",
			ApproverConfig:  "{}",
			// ApprovalMode:     "OR",
			ConditionConfig: "{}",
			// TimeoutHours:     72,
			SortOrder: 3,
			// IsRequired:       true,
			// AllowReject:      false,
			// AllowTransfer:    false,
			// AutoApprove:      false,
			// NotifyConfig:     "{}",
			// FormConfig:       "{}",
			// PermissionConfig: "{}",
			// Remark:           "",
			Status: "Normal",
		},
	}

	// 清除原有的节点数据，使用新的节点
	mockApprovalNodes[approvalDefCode] = []*model.ApprovalNode{}

	// 创建所有节点
	for _, node := range nodes {
		suite.createNode(t, node)
	}
}

// processFirstNodeAfterSubmit 处理提交后的第一个节点
func (suite *ApprovalWorkflowTestSuite) processFirstNodeAfterSubmit(approval *model.Approval, nodes []*model.ApprovalNode) {
	// 找到提交节点（START类型）
	var startNode *model.ApprovalNode
	for _, node := range nodes {
		if node.NodeType == "START" {
			startNode = node
			break
		}
	}

	if startNode == nil {
		return // 没有找到开始节点
	}

	// 找到开始节点后的下一个节点
	nextNode := suite.findNextNodeByOrder(startNode.NodeCode, nodes)
	if nextNode == nil {
		// 没有下一个节点，流程结束
		approval.Status = model.ApprovalStatusApproved
		return
	}

	// 根据下一个节点的类型进行处理
	switch nextNode.NodeType {
	case "APPROVAL":
		// 创建审批任务
		suite.createApprovalTask(nextNode, approval)
	case "CONDITION":
		// 处理条件分支节点
		suite.processConditionNode(nextNode, approval)
	case "CC":
		// 抄送节点，发送通知后继续下一个节点
		suite.processCCNode(nextNode, approval)
	case "END":
		// 直接结束
		approval.Status = model.ApprovalStatusApproved
	}
}

// TestApprovalWorkflowTestSuite 运行测试套件
func TestApprovalWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(ApprovalWorkflowTestSuite))
}
