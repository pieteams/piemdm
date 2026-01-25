package service

import (
	"io"

	"piemdm/internal/model"

	"github.com/gin-gonic/gin"
)

// ApprovalWorkflowService 审批工作流服务接口
// 用于处理实体数据的审批流程，避免与EntityService循环引用
type ApprovalWorkflowService interface {
	// 操作信息相关
	GetOperationInfo(operation string, operationInfo *map[string]string) error

	// 审批流程创建
	CreateApprovalFlow(c *gin.Context, tableCode string, approvalInfo map[string]string, formData map[string]any) error
	CreateApprovalInstance(c *gin.Context, approvalDefinition *model.ApprovalDefinition, approvalNodes []*model.ApprovalNode, approvalInfo map[string]string) error

	// 任务创建
	CreateStartTask(c *gin.Context, startNode *model.ApprovalNode, approvalInfo map[string]string) error
	CreateApprovalTask(c *gin.Context, node *model.ApprovalNode, approvalInfo map[string]string) (*model.ApprovalTask, error)

	// 节点处理
	GetNextApprovalNode(approvalNodes []*model.ApprovalNode, currentNode *model.ApprovalNode, formData map[string]any) (*model.ApprovalNode, error)
	EvaluateConditionNode(approvalNodes []*model.ApprovalNode, conditionNode *model.ApprovalNode, formData map[string]any) (*model.ApprovalNode, error)
	FindNodeByCode(approvalNodes []*model.ApprovalNode, nodeCode string) *model.ApprovalNode

	// 审批人配置
	ParseApproverConfig(node *model.ApprovalNode, assigneeID, assigneeName *string) error

	// 工具方法
	GenerateSerialNumber() string

	// 草稿相关方法
	CreateDraftWithApproval(c *gin.Context, tableCode, reason string, entityMap map[string]any) error
	UpdateDraftWithApproval(c *gin.Context, tableCode, reason string, entityMap map[string]any) error
	UpdateByIdsWithApproval(c *gin.Context, tableCode, reason string, ids []uint, entityMap map[string]any) error
	ImportWithApproval(c *gin.Context, tableCode, reason, operation string, r io.Reader) error
}
