package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"context"
	"piemdm/internal/integration/feishu"
	"piemdm/internal/model"
	"piemdm/internal/repository"
	"piemdm/pkg/notification"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ApprovalService interface {
	ApprovalWorkflowService // 继承审批工作流接口
	// 基础CRUD操作
	Get(id uint) (*model.Approval, error)
	GetByCode(code string) (*model.Approval, error)
	List(page, pageSize int, total *int64, where map[string]any) ([]*model.Approval, error)
	Create(c *gin.Context, approval *model.Approval) error
	Update(c *gin.Context, approval *model.Approval) error
	BatchUpdate(c *gin.Context, ids []uint, sysAppvoval *model.Approval) error
	Delete(c *gin.Context, id uint) (*model.Approval, error)
	BatchDelete(c *gin.Context, ids []uint) error

	// ==================== 流程审批引擎核心方法 ====================
	// 从 approval.go 迁移的方法

	// 业务查询方法
	GetByApplicant(applicantID string) ([]*model.Approval, error)
	GetByStatus(status string) ([]*model.Approval, error)
	GetPendingByApplicant(applicantID string) ([]*model.Approval, error)
	GetByEntityCode(entityCode string) ([]*model.Approval, error)
	GetByEntityID(entityCode, entityID string) ([]*model.Approval, error)
	GetApprovalHistory(id uint) ([]*model.Approval, error)

	// 基于审批任务的查询方法
	FindPendingByAssignee(assigneeName string, page, pageSize int, total *int64, timeRange string) ([]*model.Approval, error)
	FindProcessedByAssignee(assigneeName string, page, pageSize int, total *int64, timeRange string) ([]*model.Approval, error)

	// 审批流程引擎核心方法
	StartApprovalFlow(c *gin.Context, approvalDefCode, applicantID, title, formData string) (*model.Approval, error)
	// SubmitApprovalFlow(c *gin.Context, approvalCode string) error
	// CancelApprovalFlow(c *gin.Context, approvalCode, reason string) error
	ProcessApprovalFlow(c *gin.Context, taskID uint, assigneeID, comment, reason string) error
	ProcessApprovalByCodeFlow(c *gin.Context, approvalCode, nodeCode, assigneeID, action, comment, reason string) error

	// 任务处理方法
	StartApproval(c *gin.Context, approvalDefCode, applicantID, title, formData string) (*model.Approval, error)
	ApproveTask(c *gin.Context, taskId uint, comment string) error
	RejectTask(c *gin.Context, taskId uint, comment string) error
	ProcessTask(c *gin.Context, taskId uint, action, comment string) error
	// 向后兼容的方法
	// ApproveTask(c *gin.Context, taskId uint, comment string) error
	TransferTask(c *gin.Context, taskID uint, fromUserID, toUserID, fromUserName, toUserName, reason string) error

	// 任务管理方法
	// RemindTaskFlow(c *gin.Context, taskID uint) error
	// BatchRemindTasksFlow(c *gin.Context, assigneeID string) error

	// 流程控制
	MoveToNextNodeFlow(c *gin.Context, approvalCode, currentNodeCode string) error
	CompleteApprovalFlow(c *gin.Context, approvalCode, result, reason string) error
	// RejectApprovalFlow(c *gin.Context, approvalCode, reason string) error

	// 状态管理
	UpdateApprovalStatusFlow(c *gin.Context, id uint, status string) error
	UpdateCurrentNodeFlow(c *gin.Context, id uint, nodeID, nodeName string) error

	// 验证方法
	ValidateApprovalFlow(c *gin.Context, approval *model.Approval) error
	CanCancelFlow(c *gin.Context, approvalCode string) (bool, error)
	CanProcessFlow(c *gin.Context, approvalCode, assigneeID string) (bool, error)

	// 统一检查方法
	CheckExistingActiveDraft(c *gin.Context, tableCodeDraft string, entityID any) error

	// 统计方法
	GetApprovalStatisticsFlow(applicantID string) (map[string]int64, error)
	GetExpiredApprovalsFlow() ([]*model.Approval, error)

	// 其它业务方法
	SyncFeishuSubscriptions(ctx context.Context) error
}

type approvalService struct {
	*Service
	approvalRepository                repository.ApprovalRepository
	approvalTaskService               ApprovalTaskService
	approvalDefinitionService         ApprovalDefinitionService
	approvalNodeService               ApprovalNodeService
	entityRepository                  repository.EntityRepository
	tableFieldService                 TableFieldService
	entityLogService                  EntityLogService
	webhookService                    WebhookService
	tableFieldRepository              repository.TableFieldRepository
	approvalDefinitionRepository      repository.ApprovalDefinitionRepository
	tableApprovalDefinitionRepository repository.TableApprovalDefinitionRepository
	approvalNodeRepository            repository.ApprovalNodeRepository
	globalIdService                   GlobalIdService
	// ApprovalTaskService               ApprovalTaskService

	// 新增字段，用于流程审批引擎
	approvalTaskRepository repository.ApprovalTaskRepository
	userRepository         repository.UserRepository

	// 通知服务
	notificationService notification.NotificationService

	// 集成服务
	feishuIntegrationService *feishu.Service

	// 自动编码服务
	autocodeService AutocodeService
}

func NewApprovalService(
	service *Service,
	approvalRepository repository.ApprovalRepository,
	approvalTaskService ApprovalTaskService,
	approvalDefinitionService ApprovalDefinitionService,
	approvalNodeService ApprovalNodeService,
	entityRepository repository.EntityRepository,
	tableFieldService TableFieldService,
	entityLogService EntityLogService,
	webhookService WebhookService,
	tableFieldRepository repository.TableFieldRepository,
	approvalDefinitionRepository repository.ApprovalDefinitionRepository,
	tableApprovalDefinitionRepository repository.TableApprovalDefinitionRepository,
	approvalNodeRepository repository.ApprovalNodeRepository,
	globalIdService GlobalIdService,
	approvalTaskRepository repository.ApprovalTaskRepository,
	userRepository repository.UserRepository,
	notificationService notification.NotificationService,
	feishuIntegrationService *feishu.Service,
	autocodeService AutocodeService,
) ApprovalService {
	s := &approvalService{
		Service:                           service,
		approvalRepository:                approvalRepository,
		approvalTaskService:               approvalTaskService,
		approvalDefinitionService:         approvalDefinitionService,
		approvalNodeService:               approvalNodeService,
		entityRepository:                  entityRepository,
		tableFieldService:                 tableFieldService,
		entityLogService:                  entityLogService,
		webhookService:                    webhookService,
		tableFieldRepository:              tableFieldRepository,
		approvalDefinitionRepository:      approvalDefinitionRepository,
		tableApprovalDefinitionRepository: tableApprovalDefinitionRepository,
		approvalNodeRepository:            approvalNodeRepository,
		globalIdService:                   globalIdService,
		approvalTaskRepository:            approvalTaskRepository,
		userRepository:                    userRepository,
		notificationService:               notificationService,
		feishuIntegrationService:          feishuIntegrationService,
		autocodeService:                   autocodeService,
	}

	// 注册飞书回调
	if feishuIntegrationService != nil {
		feishuIntegrationService.SetApprovalCallback(func(ctx context.Context, externalInstID string, status string) error {
			return s.handleFeishuApprovalUpdate(ctx, externalInstID, status)
		})
	}

	return s
}

// SyncFeishuSubscriptions 同步并激活所有活跃的飞书审批定义订阅
func (s *approvalService) SyncFeishuSubscriptions(ctx context.Context) error {
	if s.feishuIntegrationService == nil {
		return nil
	}

	// 1. 获取所有活跃的飞书审批定义 Code
	codes, err := s.approvalDefinitionRepository.FindActiveFeishuCodes()
	if err != nil {
		return fmt.Errorf("failed to load active feishu codes: %w", err)
	}

	// 2. 批量订阅
	s.feishuIntegrationService.SubscribeAllApprovals(ctx, codes)

	return nil
}

// 基础CRUD操作
func (s *approvalService) Get(id uint) (*model.Approval, error) {
	return s.approvalRepository.FindOne(id)
}

// createDummyContext 创建一个极简的 Gin 上下文用于后台任务
func (s *approvalService) createDummyContext(ctx context.Context, userName string) *gin.Context {
	c, _ := gin.CreateTestContext(nil)
	c.Request, _ = http.NewRequestWithContext(ctx, "POST", "/feishu/callback", nil)
	c.Set("user_id", uint(0))
	c.Set("user_name", userName)
	return c
}

// handleFeishuApprovalUpdate 处理来自飞书集成的审批状态变更
func (s *approvalService) handleFeishuApprovalUpdate(ctx context.Context, externalInstID string, status string) error {
	slog.Info("Handling feishu approval update", "externalID", externalInstID, "status", status)

	// 1. 查找本地审批记录
	approval, err := s.approvalRepository.FirstByExternalInstanceID(externalInstID)
	if err != nil {
		return fmt.Errorf("failed to find approval by external ID %s: %w", externalInstID, err)
	}

	if approval == nil {
		slog.Warn("No local approval record found for feishu instance", "externalID", externalInstID)
		return nil
	}

	// 如果本地已经是终态，忽略
	if approval.IsCompleted() {
		slog.Debug("Approval already completed, ignoring feishu update", "code", approval.Code)
		return nil
	}

	// 2. 构造上下文 (针对内部逻辑中大量依赖 c.GetString("user_name") 的情况)
	c := s.createDummyContext(ctx, "feishu_sync")

	// 3. 根据飞书状态映射本地逻辑
	// 飞书状态: REJECTED, APPROVED, PENDING ...
	switch status {
	case "APPROVED":
		// 如果是通过，执行本地流程完成逻辑
		slog.Info("Feishu approval APPROVED, completing local flow", "code", approval.Code)
		return s.completeApprovalFlow(c, approval)
	case "REJECTED":
		// 如果是拒绝，执行本地拒绝逻辑
		slog.Info("Feishu approval REJECTED, handling local rejection", "code", approval.Code)
		return s.handleRejection(c, approval, nil, "rejected by feishu")
	default:
		slog.Debug("Ignoring non-terminal feishu status", "status", status, "code", approval.Code)
	}

	return nil
}

func (s *approvalService) GetByCode(code string) (*model.Approval, error) {
	return s.approvalRepository.FirstByCode(code)
}

func (s *approvalService) List(page, pageSize int, total *int64, where map[string]any) ([]*model.Approval, error) {
	return s.approvalRepository.FindPage(page, pageSize, total, where)
}

func (s *approvalService) Create(c *gin.Context, approval *model.Approval) error {
	// 验证审批实例
	if err := s.validateApproval(c, approval); err != nil {
		s.logger.Error("审批实例验证失败", "error", err)
		return err
	}

	// 生成唯一编码
	if approval.Code == "" {
		uuid := uuid.New()
		approval.Code = strings.ToUpper(uuid.String())
	}

	// 检查是否为飞书审批
	approvalDef, err := s.approvalDefinitionRepository.First(map[string]any{"code": approval.ApprovalDefCode})
	if err == nil && approvalDef.Platform == "Feishu" {
		return s.createFeishuInstance(c, approval, approvalDef)
	}

	return s.approvalRepository.Create(c, approval)
}

func (s *approvalService) Update(c *gin.Context, approval *model.Approval) error {
	return s.approvalRepository.Update(c, approval)
}

func (s *approvalService) BatchUpdate(c *gin.Context, ids []uint, sysAppvoval *model.Approval) error {
	return s.approvalRepository.BatchUpdate(c, ids, sysAppvoval)
}

func (s *approvalService) Delete(c *gin.Context, id uint) (*model.Approval, error) {
	return s.approvalRepository.Delete(c, id)
}

func (s *approvalService) BatchDelete(c *gin.Context, ids []uint) error {
	return s.approvalRepository.BatchDelete(c, ids)
}

// createFeishuInstance 创建飞书审批实例
func (s *approvalService) createFeishuInstance(c *gin.Context, approval *model.Approval, def *model.ApprovalDefinition) error {
	ctx := c.Request.Context()
	feishuUserID, err := s.getFeishuID(c)
	if err != nil {
		return fmt.Errorf("failed to get initiator feishu ID: %w", err)
	}

	// 1. 准备表单数据 (过滤掉非飞书定义的字段)
	formValues := "[]"
	if approval.FormData != "" {
		// 校验：飞书平台必须提供表单定义 JSON，否则无法通过过滤防止 60022 错误
		if def.FormData == "" {
			return fmt.Errorf("飞书审批必须配置“表单结构定义”(form_content JSON)")
		}

		var data map[string]any
		if err := json.Unmarshal([]byte(approval.FormData), &data); err == nil {
			// 从定义中获取有效字段 ID
			allowedIDs := s.feishuIntegrationService.ExtractFieldIDs(def.FormData)
			if len(allowedIDs) == 0 {
				return fmt.Errorf("无法从“表单结构定义”中提取到有效的控件 ID")
			}
			qv, err := s.feishuIntegrationService.GenerateFormValues(data, allowedIDs)
			if err == nil {
				formValues = qv
			}
		}
	}

	// 2. 调用 Feishu Service
	externalID, err := s.feishuIntegrationService.CreateApprovalInstance(ctx, def.Code, formValues, feishuUserID, approval.Title)
	if err != nil {
		return fmt.Errorf("create feishu instance failed: %w", err)
	}

	approval.ExternalInstanceID = externalID
	approval.Status = model.ApprovalStatusPending

	return s.approvalRepository.Create(c, approval)
}

// HandleExternalStatusUpdate 处理外部审批状态变更
func (s *approvalService) HandleExternalStatusUpdate(ctx context.Context, externalInstID string, status string) error {
	// 1. 查找审批记录
	approval, err := s.approvalRepository.First(map[string]any{"external_instance_id": externalInstID})
	if err != nil {
		return fmt.Errorf("approval not found for external id: %s", externalInstID)
	}

	// 2. 映射状态
	var newStatus string
	switch status {
	case "APPROVED":
		newStatus = model.ApprovalStatusApproved
	case "REJECTED":
		newStatus = model.ApprovalStatusRejected
	case "CANCELED":
		newStatus = model.ApprovalStatusCanceled
	case "DELETED":
		newStatus = model.ApprovalStatusDeleted
	default:
		// 其他状态暂不处理或视为Pending
		return nil
	}

	// 3. 如果状态改变，执行相应逻辑
	if approval.Status != newStatus {
		// Mock Context since we are in background
		c := &gin.Context{}
		c.Set("user_name", "System/Feishu")

		if newStatus == model.ApprovalStatusApproved {
			return s.completeApprovalFlow(c, approval)
		} else if newStatus == model.ApprovalStatusRejected {
			// 创建虚构的任务和评论来复用 handleRejection
			dummyTask := &model.ApprovalTask{}
			return s.handleRejection(c, approval, dummyTask, "Rejected on Feishu")
		} else {
			// 其他状态直接更新
			approval.Status = newStatus
			approval.UpdatedBy = "System/Feishu"
			return s.approvalRepository.Update(c, approval)
		}
	}

	return nil
}

func (s *approvalService) ApproveTask(c *gin.Context, taskId uint, comment string) error {
	return s.processApprovalTask(c, taskId, "APPROVE", comment)
}

func (s *approvalService) RejectTask(c *gin.Context, taskId uint, comment string) error {
	return s.processApprovalTask(c, taskId, "REJECT", comment)
}

// processApprovalTask 处理审批任务
func (s *approvalService) processApprovalTask(c *gin.Context, taskId uint, action, comment string) error {
	currentUser := c.GetString("user_name")

	// 1. 先根据ID获取任务信息,获取 approval_code 和 node_code
	clickedTask, err := s.approvalTaskService.First(map[string]any{
		"id": taskId,
	})
	if err != nil || clickedTask == nil {
		s.logger.Error("任务不存在",
			"taskId", taskId,
			"error", err)
		return fmt.Errorf("任务不存在")
	}

	// 2. 根据 approval_code, node_code, assignee_name, status 查找当前用户的待审批任务
	task, err := s.approvalTaskService.First(map[string]any{
		"approval_code": clickedTask.ApprovalCode,
		"node_code":     clickedTask.NodeCode,
		"assignee_name": currentUser,
		"status":        model.TaskStatusPending,
	})

	if err != nil || task == nil {
		// 检查点击的任务是否就是当前用户的任务
		if clickedTask.AssigneeName == currentUser {
			if clickedTask.Status != model.TaskStatusPending {
				s.logger.Warn("任务已经被处理",
					"taskId", taskId,
					"status", clickedTask.Status)
				return fmt.Errorf("该任务已经被处理,状态: %s", clickedTask.Status)
			}
		} else {
			s.logger.Warn("当前节点没有分配给当前用户的待审批任务",
				"currentUser", currentUser,
				"nodeCode", clickedTask.NodeCode,
				"clickedTaskAssignee", clickedTask.AssigneeName)
			return fmt.Errorf("当前节点没有分配给您的待审批任务")
		}
		return fmt.Errorf("您没有权限或已经审批过该任务")
	}

	// 3. 获取审批实例
	approval, err := s.approvalRepository.First(map[string]any{
		"code": task.ApprovalCode,
	})
	if err != nil {
		return fmt.Errorf("获取审批实例失败: %v", err)
	}

	// 3. 获取审批定义
	approvalDef, err := s.approvalDefinitionService.First(map[string]any{
		"code": approval.ApprovalDefCode,
	})
	if err != nil {
		return fmt.Errorf("获取审批定义失败: %v", err)
	}

	// 4. 获取所有审批节点
	approvalNodes, err := s.getApprovalNodes(approvalDef.Code)
	if err != nil {
		return fmt.Errorf("获取审批节点失败: %v", err)
	}

	// 5. 更新当前任务状态
	if err := s.updateTaskStatus(task, action, comment, currentUser); err != nil {
		return fmt.Errorf("更新任务状态失败: %v", err)
	}

	// 6. 处理后续流程
	switch action {
	case "APPROVE":
		return s.handleApprovalFlow(c, approval, task, approvalNodes)
	case "REJECT":
		return s.handleRejection(c, approval, task, comment)
	case "TRANSFER":
		return s.handleTransfer(c, task, comment)
	default:
		return fmt.Errorf("不支持的操作类型: %s", action)
	}
}

// handleApprovalFlow 处理审批通过流程
func (s *approvalService) handleApprovalFlow(c *gin.Context, approval *model.Approval, currentTask *model.ApprovalTask, approvalNodes []*model.ApprovalNode) error {
	// 1. 找到当前节点
	currentNode := s.findNodeByCode(approvalNodes, currentTask.NodeCode)
	if currentNode == nil {
		return fmt.Errorf("未找到当前节点: %s", currentTask.NodeCode)
	}

	// 2. 检查当前节点是否完成(根据OR/AND模式)
	isCompleted, err := s.isNodeCompleted(currentNode, approval.Code)
	if err != nil {
		return fmt.Errorf("检查节点完成状态失败: %v", err)
	}

	if !isCompleted {
		// 节点未完成(AND模式且还有人未审批),流程暂停
		return nil
	}

	// 3. OR模式下取消其他待处理任务
	if err := s.cancelPendingTasksInNode(approval.Code, currentNode.NodeCode); err != nil {
		s.logger.Error("取消待处理任务失败", "error", err)
		// 不阻断流程
	}

	// 4. 获取并处理下一个节点(循环直到找到需要处理的节点或流程结束)
	for {
		nextNode, err := s.getNextApprovalNode(approvalNodes, currentNode, approval.FormData)
		if err != nil {
			return fmt.Errorf("获取下一个节点失败: %v", err)
		}

		// 5. 如果没有下一个节点，说明流程结束
		if nextNode == nil {
			return s.completeApprovalFlow(c, approval)
		}

		switch nextNode.NodeType {
		case "APPROVAL":
			// 检查节点条件
			if nextNode.ConditionConfig != "" && !s.evaluateNodeCondition(nextNode, approval) {
				// 将当前节点设置为这个跳过的节点,继续循环查找下一个节点
				currentNode = nextNode
				continue
			}
			// 条件满足,创建任务
			return s.createNextApprovalTask(c, nextNode, approval)

		case "CC":
			return s.handleNotificationNode(c, nextNode, approval, approvalNodes)

		case "CONDITION":
			// 条件节点应该在 getNextApprovalNode 中已经处理
			return fmt.Errorf("条件节点不应该直接处理")

		case "END":
			// 创建END任务
			if err := s.createEndTask(c, approval); err != nil {
				s.logger.Error("创建END任务失败", "error", err)
				// END任务创建失败不应该阻断流程完成
			}
			return s.completeApprovalFlow(c, approval)

		default:
			return fmt.Errorf("不支持的节点类型: %s", nextNode.NodeType)
		}
	}
}

// handleNotificationNode 处理通知节点
func (s *approvalService) handleNotificationNode(c *gin.Context, ccNode *model.ApprovalNode, approval *model.Approval, approvalNodes []*model.ApprovalNode) error {
	// 1. 发送通知
	if err := s.sendNotification(c, ccNode, approval); err != nil {
		s.logger.Error("发送通知失败", "error", err)
		// 通知失败不应该阻断流程
	}

	// 2. 创建抄送任务记录
	if err := s.createCCTask(c, ccNode, approval); err != nil {
		s.logger.Error("创建抄送任务失败", "error", err)
	}

	// 3. 继续寻找下一个审批节点
	nextNode, err := s.getNextApprovalNode(approvalNodes, ccNode, approval.FormData)
	if err != nil {
		return fmt.Errorf("获取通知节点后的下一个节点失败: %v", err)
	}

	// 抄送节点后的下一个节点"
	if nextNode == nil {
		return s.completeApprovalFlow(c, approval)
	}

	// 4. 递归处理下一个节点
	switch nextNode.NodeType {
	case "APPROVAL":
		return s.createNextApprovalTask(c, nextNode, approval)
	case "CC":
		return s.handleNotificationNode(c, nextNode, approval, approvalNodes)
	case "END":
		// 创建END任务
		if err := s.createEndTask(c, approval); err != nil {
			s.logger.Error("创建END任务失败", "error", err)
			// END任务创建失败不应该阻断流程完成
		}
		return s.completeApprovalFlow(c, approval)
	default:
		return fmt.Errorf("通知节点后不支持的节点类型: %s", nextNode.NodeType)
	}
}

// handleRejection 处理审批拒绝
func (s *approvalService) handleRejection(c *gin.Context, approval *model.Approval, task *model.ApprovalTask, comment string) error {
	// 需要更新 实例，任务，草稿
	// 1. 更新审批实例状态
	// now := time.Now()
	approval.Status = model.ApprovalStatusRejected
	// approval.CompletedAt = &now
	approval.UpdatedBy = c.GetString("user_name")
	if err := s.approvalRepository.Update(c, approval); err != nil {
		return fmt.Errorf("更新审批实例状态失败: %v", err)
	}

	// 2. 取消所有待处理任务
	if err := s.cancelPendingTasks(approval.Code); err != nil {
		s.logger.Error("取消待处理任务失败", "error", err)
	}

	// 3. 更新draft状态
	tableCodeDraft := fmt.Sprintf("%s_draft", approval.EntityCode)
	updateMap := map[string]any{
		"draft_status": "Drafted",
	}
	whereMap := map[string]any{
		"approval_code": approval.Code,
	}
	if err := s.entityRepository.Update(c, tableCodeDraft, updateMap, whereMap); err != nil {
		s.logger.Error("handleRejection: failed to revert draft status", "err", err, "approvalCode", approval.Code)
		// Non-blocking error, but should be logged
	}

	return nil
}

// completeApprovalFlow 完成审批流程
func (s *approvalService) completeApprovalFlow(c *gin.Context, approval *model.Approval) error {
	s.logger.Debug("service-approval-completeApprovalFlow", "approval", approval)
	// 1. 更新审批实例状态
	// now := time.Now()
	approval.Status = model.ApprovalStatusApproved
	// approval.CompletedAt = &now
	approval.CurrentTaskName = "已完成"
	approval.UpdatedBy = c.GetString("user_name")

	if err := s.approvalRepository.Update(c, approval); err != nil {
		return fmt.Errorf("更新审批实例状态失败: %v", err)
	}

	// 1.5. 取消所有待处理任务(确保数据一致性)
	if err := s.cancelPendingTasks(approval.Code); err != nil {
		s.logger.Error("取消待处理任务失败", "error", err)
		// 不阻断流程,仅记录错误
	}

	s.logger.Debug("service-approval-completeApprovalFlow2", "approval", approval)
	// 2. 执行审批通过后的业务逻辑
	if err := s.approved(c, approval, c.GetString("user_name")); err != nil {
		return fmt.Errorf("执行审批通过逻辑失败: %v", err)
	}

	// 审批流程完成
	return nil
}

// createNextApprovalTask 创建下一个审批任务
// 注意: 调用此方法前应该已经检查过节点条件
func (s *approvalService) createNextApprovalTask(c *gin.Context, node *model.ApprovalNode, approval *model.Approval) error {
	// 1. 获取审批人列表
	approvers, err := s.getApprovers(node)
	if err != nil {
		return fmt.Errorf("获取审批人列表失败: %v", err)
	}

	if len(approvers) == 0 {
		return fmt.Errorf("未找到审批人")
	}

	// 3. 为每个审批人创建任务
	var createdTaskIds []uint
	for _, approver := range approvers {
		nextTask := model.ApprovalTask{
			ApprovalCode: approval.Code,
			NodeCode:     node.NodeCode,
			NodeName:     node.NodeName,
			AssigneeID:   approver.ID,
			AssigneeName: approver.Name,
			Status:       model.TaskStatusPending,
			CreatedBy:    c.GetString("user_name"),
			UpdatedBy:    c.GetString("user_name"),
		}

		// 4. 处理自动审批节点
		switch node.ApproverType {
		case "AUTO_APPROVE":
			nextTask.Status = model.TaskStatusApproved
			nextTask.Comment = "系统自动通过"
		case "AUTO_REJECT":
			nextTask.Status = model.TaskStatusRejected
			nextTask.Comment = "系统自动驳回"
		}

		// 5. 保存任务
		if err := s.approvalTaskService.Create(&nextTask); err != nil {
			return fmt.Errorf("创建审批任务失败: %v", err)
		}

		// 记录创建的任务ID
		createdTaskIds = append(createdTaskIds, nextTask.ID)
	}

	// 6. 更新审批实例的当前节点
	if len(createdTaskIds) > 0 {
		// 使用第一个创建的任务ID作为current_task_id
		// 对于多审批人,使用哪个任务ID都可以,因为查询时会根据approval_code+node_code+assignee_name来查找
		approval.CurrentTaskID = strconv.FormatUint(uint64(createdTaskIds[0]), 10)
		approval.CurrentTaskName = node.NodeName
		approval.UpdatedBy = c.GetString("user_name")

		if err := s.approvalRepository.Update(c, approval); err != nil {
			return fmt.Errorf("更新审批实例当前节点失败: %v", err)
		}
	}

	return nil
}

// 辅助方法
func (s *approvalService) updateTaskStatus(task *model.ApprovalTask, action, comment, userName string) error {
	// now := time.Now()
	// task.CompletedAt = &now
	task.Comment = comment
	task.UpdatedBy = userName

	switch action {
	case "APPROVE":
		task.Status = model.TaskStatusApproved
	case "REJECT":
		task.Status = model.TaskStatusRejected
	case "TRANSFER":
		task.Status = model.TaskStatusTransferred
	default:
		return fmt.Errorf("不支持的操作: %s", action)
	}

	return s.approvalTaskService.Update(task)
}

func (s *approvalService) findNodeByCode(nodes []*model.ApprovalNode, nodeCode string) *model.ApprovalNode {
	for _, node := range nodes {
		if node.NodeCode == nodeCode {
			return node
		}
	}
	return nil
}

func (s *approvalService) getApprovalNodes(approvalDefCode string) ([]*model.ApprovalNode, error) {
	// 调用审批节点 repository 获取活跃的审批节点
	var total int64
	nodes, err := s.approvalNodeService.List(1, 100, &total, map[string]any{
		"approval_def_code": approvalDefCode,
		"status":            "Normal",
	})
	if err != nil {
		return nil, fmt.Errorf("获取审批节点失败: %v", err)
	}

	// 按照节点顺序排序
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].SortOrder < nodes[j].SortOrder
	})

	// 获取审批节点成功
	return nodes, nil
}

func (s *approvalService) sendNotification(c *gin.Context, node *model.ApprovalNode, approval *model.Approval) error {
	// 如果通知服务未初始化，跳过通知发送
	if s.notificationService == nil {
		s.logger.Warn("通知服务未初始化，跳过通知发送")
		return nil
	}

	// 检查是否已经发送过通知，防止重复发送
	existingTasks, err := s.approvalTaskService.GetByApprovalCode(approval.Code)
	if err != nil {
		s.logger.Error("获取现有任务失败", "error", err)
	} else {
		for _, task := range existingTasks {
			if task.NodeCode == node.NodeCode && task.Status == model.TaskStatusDone {
				// 通知已发送，跳过重复发送
				return nil
			}
		}
	}

	// 解析抄送配置，获取接收者列表
	var ccConfig struct {
		Type     string   `json:"type"`
		Users    []string `json:"users"`
		Mode     string   `json:"mode"`
		CcTiming string   `json:"ccTiming"`
	}

	if err := json.Unmarshal([]byte(node.ApproverConfig), &ccConfig); err != nil {
		s.logger.Error("解析抄送配置失败", "error", err, "config", node.ApproverConfig)
		return nil
	}

	var recipients []string
	var usernames []string

	for _, user := range ccConfig.Users {
		user = strings.TrimSpace(user)
		if user != "" {
			usernames = append(usernames, user)
		}
	}

	if len(usernames) > 0 {
		// 查询用户邮箱
		where := map[string]any{
			"username": usernames,
		}
		users, err := s.userRepository.Find("email", where)
		if err != nil {
			s.logger.Error("查询抄送用户失败", "error", err, "usernames", usernames)
		} else {
			for _, user := range users {
				if user.Email != "" {
					recipients = append(recipients, user.Email)
				}
			}
		}
	}

	// 如果没有找到邮箱，尝试使用默认域名（兼容旧逻辑，可选）
	// for _, user := range usernames {
	// 	found := false
	// 	for _, recipient := range recipients {
	// 		if strings.HasPrefix(recipient, user+"@") {
	// 			found = true
	// 			break
	// 		}
	// 	}
	// 	if !found {
	// 		recipients = append(recipients, user+"@example.com")
	// 	}
	// }

	// 如果没有配置接收者，使用默认接收者（申请人）
	if len(recipients) == 0 {
		// 实际项目中应该查询申请人的邮箱地址
		recipients = append(recipients, "applicant@example.com")
	}

	// 创建抄送通知消息
	message := notification.CreateCCNotificationMessage(
		recipients,
		approval.Title,
		approval.CreatedBy,
		"处理中", // 当前状态
	)

	// 发送通知
	_, err = s.notificationService.Send(c.Request.Context(), message)
	if err != nil {
		s.logger.Error("发送通知失败",
			"error", err,
			"recipients", recipients)
		// 通知发送失败不应该阻断审批流程
		return nil
	}

	return nil
}

func (s *approvalService) createCCTask(c *gin.Context, node *model.ApprovalNode, approval *model.Approval) error {
	// 开始创建抄送任务

	// 检查是否已经存在相同的抄送任务，防止重复创建
	existingTasks, err := s.approvalTaskService.GetByApprovalCode(approval.Code)
	if err != nil {
		s.logger.Error("获取现有任务失败", "error", err)
	} else {
		for _, task := range existingTasks {
			if task.NodeCode == node.NodeCode {
				// 抄送任务已存在，跳过创建
				return nil
			}
		}
	}

	// 解析抄送配置
	var ccConfig struct {
		CcUsers  string `json:"ccUsers"`
		CcTiming string `json:"ccTiming"`
	}

	// // 优先使用 CustomConfig，如果没有则使用 NotifyConfig
	// configStr := node.CustomConfig
	// if configStr == "" {
	// 	configStr = node.NotifyConfig
	// }

	// if configStr != "" {
	// 	if err := json.Unmarshal([]byte(configStr), &ccConfig); err != nil {
	// 		s.logger.Error("解析抄送配置失败", "error", err)
	// 		return fmt.Errorf("解析抄送配置失败: %v", err)
	// 	}
	// }

	// 解析抄送用户列表
	ccUserList := strings.Split(ccConfig.CcUsers, "，")
	if len(ccUserList) == 0 || (len(ccUserList) == 1 && ccUserList[0] == "") {
		ccUserList = strings.Split(ccConfig.CcUsers, ",")
	}

	// 为每个抄送用户创建任务
	// now := time.Now()
	for _, ccUser := range ccUserList {
		ccUser = strings.TrimSpace(ccUser)
		if ccUser == "" {
			continue
		}

		// 创建抄送任务
		ccTask := model.ApprovalTask{
			ApprovalCode: approval.Code,
			NodeCode:     node.NodeCode,
			NodeName:     node.NodeName,
			AssigneeID:   ccUser,
			AssigneeName: ccUser,               // 这里可以根据实际需要查询用户真实姓名
			Status:       model.TaskStatusDone, // 抄送任务直接标记为完成
			Comment:      "抄送通知",
			CreatedBy:    c.GetString("user_name"),
			UpdatedBy:    c.GetString("user_name"),
			// StartedAt:    &now,
			// CompletedAt:  &now,
		}

		// 保存抄送任务
		if err := s.approvalTaskService.Create(&ccTask); err != nil {
			s.logger.Error("创建抄送任务失败",
				"ccUser", ccUser,
				"error", err)
			// 抄送任务创建失败不应该阻断流程
			continue
		}
		// 创建抄送任务成功
	}

	return nil
}

// createEndTask 创建END任务记录
func (s *approvalService) createEndTask(c *gin.Context, approval *model.Approval) error {
	// 开始创建END任务

	// 获取审批定义的所有节点，找到END节点
	approvalNodes, err := s.approvalNodeService.GetByApprovalDefCode(approval.ApprovalDefCode)
	if err != nil {
		s.logger.Error("获取审批节点失败", "error", err)
		return fmt.Errorf("获取审批节点失败: %v", err)
	}

	// 查找END节点
	var endNode *model.ApprovalNode
	for _, node := range approvalNodes {
		if node.NodeType == "END" {
			endNode = node
			break
		}
	}

	if endNode == nil {
		s.logger.Warn("未找到END节点",
			"approvalDefCode", approval.ApprovalDefCode)
		return fmt.Errorf("未找到END节点")
	}

	// 检查是否已经存在END任务，防止重复创建
	existingTasks, err := s.approvalTaskService.GetByApprovalCode(approval.Code)
	if err != nil {
		s.logger.Error("获取现有任务失败", "error", err)
	} else {
		for _, task := range existingTasks {
			if task.NodeCode == endNode.NodeCode {
				// END任务已存在，跳过创建
				return nil
			}
		}
	}

	// 创建END任务
	// now := time.Now()
	endTask := model.ApprovalTask{
		ApprovalCode: approval.Code,
		NodeCode:     endNode.NodeCode,
		NodeName:     endNode.NodeName,
		AssigneeID:   "system",
		AssigneeName: "系统",
		Status:       model.TaskStatusDone, // END任务直接标记为完成
		Comment:      "流程完成",
		CreatedBy:    c.GetString("user_name"),
		UpdatedBy:    c.GetString("user_name"),
		// StartedAt:    &now,
		// CompletedAt:  &now,
	}

	// 保存END任务
	if err := s.approvalTaskService.Create(&endTask); err != nil {
		s.logger.Error("创建END任务失败", "error", err)
		return fmt.Errorf("创建END任务失败: %v", err)
	}

	// 创建END任务成功
	return nil
}

func (s *approvalService) cancelPendingTasks(approvalCode string) error {
	// 获取所有待处理的任务
	// 使用 GetByApprovalCode 方法获取指定审批的所有任务
	allTasks, err := s.approvalTaskService.GetByApprovalCode(approvalCode)
	if err != nil {
		s.logger.Warn("获取审批任务失败", "error", err)
		return nil
	}

	// 过滤出待处理的任务
	var pendingTasks []*model.ApprovalTask
	for _, task := range allTasks {
		if task.Status == model.TaskStatusPending {
			pendingTasks = append(pendingTasks, task)
		}
	}

	// 批量取消任务
	// now := time.Now()
	for _, task := range pendingTasks {
		task.Status = model.TaskStatusCanceled
		task.Comment = "审批被拒绝，任务自动取消"
		//  task.CompletedAt = &now
		task.UpdatedBy = "system"

		if err := s.approvalTaskService.Update(task); err != nil {
			s.logger.Error("取消任务失败",
				"taskId", fmt.Sprintf("%d", task.ID),
				"error", err)
		}
	}

	return nil
}

func (s *approvalService) parseApproverConfig(node *model.ApprovalNode, assigneeID, assigneeName *string) error {
	// 处理特殊的审批人类型
	switch node.ApproverType {
	case "AUTO_REJECT":
		*assigneeID = "system"
		*assigneeName = "系统自动驳回"
		return nil
	case "AUTO_APPROVE":
		*assigneeID = "system"
		*assigneeName = "系统自动通过"
		return nil
	}

	// 解析审批人配置JSON
	if node.ApproverConfig == "" {
		*assigneeID = "system"
		*assigneeName = "系统"
		return nil
	}

	var approverConfig struct {
		Type  string   `json:"type"`  // USERS, ROLES, EXPRESSION
		Users []string `json:"users"` // 用户列表
		Roles []string `json:"roles"` // 角色列表
		Mode  string   `json:"mode"`  // OR, AND
	}

	if err := json.Unmarshal([]byte(node.ApproverConfig), &approverConfig); err != nil {
		s.logger.Error("解析审批人配置失败", "error", err)
		*assigneeID = "system"
		*assigneeName = "系统"
		return err
	}

	// 根据配置类型处理
	switch approverConfig.Type {
	case "USERS":
		if len(approverConfig.Users) > 0 {
			// 取第一个用户作为审批人（简化处理，实际可能需要支持多人审批）
			*assigneeID = approverConfig.Users[0]
			*assigneeName = s.getUserNameByID(approverConfig.Users[0])
		} else {
			*assigneeID = "system"
			*assigneeName = "系统"
		}
	case "ROLES":
		if len(approverConfig.Roles) > 0 {
			// 根据角色查找审批人
			userID, userName, err := s.getUserByRole(approverConfig.Roles[0])
			if err != nil {
				s.logger.Error("根据角色查找审批人失败", "error", err)
				*assigneeID = "system"
				*assigneeName = "系统"
			} else {
				*assigneeID = userID
				*assigneeName = userName
			}
		} else {
			*assigneeID = "system"
			*assigneeName = "系统"
		}
	case "EXPRESSION":
		// TODO: 实现表达式解析逻辑
		*assigneeID = "expression_user"
		*assigneeName = "表达式用户"
	default:
		*assigneeID = "system"
		*assigneeName = "系统"
	}

	return nil
}

// getApprovers 获取审批人列表
func (s *approvalService) getApprovers(node *model.ApprovalNode) ([]Approver, error) {
	// 处理特殊的审批人类型
	switch node.ApproverType {
	case "AUTO_REJECT", "AUTO_APPROVE":
		return []Approver{{ID: "system", Name: "系统"}}, nil
	}

	// 解析审批人配置JSON
	if node.ApproverConfig == "" {
		return []Approver{{ID: "system", Name: "系统"}}, nil
	}

	var approverConfig struct {
		Type  string   `json:"type"`  // USERS, ROLES, EXPRESSION
		Users []string `json:"users"` // 用户列表
		Roles []string `json:"roles"` // 角色列表
		Mode  string   `json:"mode"`  // OR, AND
	}

	if err := json.Unmarshal([]byte(node.ApproverConfig), &approverConfig); err != nil {
		s.logger.Error("解析审批人配置失败", "error", err)
		return nil, err
	}

	var approvers []Approver

	// 根据配置类型处理
	switch approverConfig.Type {
	case "USERS":
		// 为每个用户创建审批人信息
		for _, userID := range approverConfig.Users {
			userName := s.getUserNameByID(userID)
			approvers = append(approvers, Approver{
				ID:   userID,
				Name: userName,
			})
		}

	case "ROLES":
		// 根据角色获取用户列表
		for _, roleID := range approverConfig.Roles {
			userID, userName, err := s.getUserByRole(roleID)
			if err != nil {
				s.logger.Error("根据角色获取用户失败", "error", err)
				continue
			}
			approvers = append(approvers, Approver{
				ID:   userID,
				Name: userName,
			})
		}

	case "EXPRESSION":
		// TODO: 实现表达式解析逻辑
		approvers = append(approvers, Approver{
			ID:   "expression_user",
			Name: "表达式用户",
		})

	default:
		approvers = append(approvers, Approver{
			ID:   "system",
			Name: "系统",
		})
	}

	return approvers, nil
}

// evaluateNodeCondition 评估节点条件
func (s *approvalService) evaluateNodeCondition(node *model.ApprovalNode, approval *model.Approval) bool {
	// 如果节点没有条件配置,默认返回true
	if node.ConditionConfig == "" {
		return true
	}

	// 解析条件配置
	var conditionConfig struct {
		FieldName     string `json:"fieldName"`
		Operator      string `json:"operator"`
		FieldValue    string `json:"fieldValue"`
		ConditionType string `json:"conditionType"` // instance 或 data
	}

	if err := json.Unmarshal([]byte(node.ConditionConfig), &conditionConfig); err != nil {
		s.logger.Error("解析条件配置失败",
			"error", err,
			"conditionConfig", node.ConditionConfig)
		return false
	}

	var fieldValue any
	// var dataSource string

	// 根据条件类型获取字段值
	switch conditionConfig.ConditionType {
	case "instance":
		// 从审批实例获取字段值
		// dataSource = "审批实例"
		switch conditionConfig.FieldName {
		case "created_by":
			fieldValue = approval.CreatedBy
		case "status":
			fieldValue = approval.Status
		case "code":
			fieldValue = approval.Code
		case "title":
			fieldValue = approval.Title
		// 可以根据需要添加更多字段
		default:
			s.logger.Warn("不支持的审批实例字段",
				"fieldName", conditionConfig.FieldName)
			return false
		}

	case "data":
		// 从表单数据获取字段值
		// dataSource = "表单数据"
		if approval.FormData == "" {
			s.logger.Warn("表单数据为空，无法评估条件",
				"fieldName", conditionConfig.FieldName)
			return false
		}

		var formDataMap map[string]any
		if err := json.Unmarshal([]byte(approval.FormData), &formDataMap); err != nil {
			s.logger.Error("解析表单数据失败",
				"error", err,
				"formData", approval.FormData)
			return false
		}

		var ok bool
		fieldValue, ok = formDataMap[conditionConfig.FieldName]
		if !ok {
			s.logger.Warn("表单数据中未找到字段",
				"fieldName", conditionConfig.FieldName,
				"formDataMap", formDataMap)
			return false
		}

	default:
		s.logger.Error("不支持的条件类型",
			"conditionType", conditionConfig.ConditionType)
		return false
	}

	// 评估条件
	result := s.evaluateConditionByOperator(conditionConfig.Operator, fieldValue, conditionConfig.FieldValue)

	return result
}

// evaluateConditionByOperator 评估单个条件
func (s *approvalService) evaluateConditionByOperator(operator string, actualValue any, expectedValue string) bool {
	// 将实际值转换为字符串进行比较
	actualStr := fmt.Sprintf("%v", actualValue)

	switch operator {
	case "eq": // 等于
		return actualStr == expectedValue
	case "ne": // 不等于
		return actualStr != expectedValue
	case "gt": // 大于
		return s.compareNumeric(actualStr, expectedValue) > 0
	case "gte": // 大于等于
		return s.compareNumeric(actualStr, expectedValue) >= 0
	case "lt": // 小于
		return s.compareNumeric(actualStr, expectedValue) < 0
	case "lte": // 小于等于
		return s.compareNumeric(actualStr, expectedValue) <= 0
	case "contains": // 包含
		return strings.Contains(actualStr, expectedValue)
	case "not_contains": // 不包含
		return !strings.Contains(actualStr, expectedValue)
	default:
		s.logger.Warn("不支持的操作符", "operator", operator)
		return false
	}
}

// compareNumeric 数值比较
func (s *approvalService) compareNumeric(actual, expected string) int {
	actualNum, err1 := strconv.ParseFloat(actual, 64)
	expectedNum, err2 := strconv.ParseFloat(expected, 64)

	if err1 != nil || err2 != nil {
		// 如果无法转换为数字,则进行字符串比较
		return strings.Compare(actual, expected)
	}

	if actualNum > expectedNum {
		return 1
	} else if actualNum < expectedNum {
		return -1
	}
	return 0
}

func (s *approvalService) getNextApprovalNode(nodes []*model.ApprovalNode, currentNode *model.ApprovalNode, formData string) (*model.ApprovalNode, error) {
	// 解析表单数据
	var formDataMap map[string]any
	if formData != "" {
		// 先尝试修复可能的JSON格式问题
		fixedFormData := strings.ReplaceAll(formData, `""`, `","`)
		if err := json.Unmarshal([]byte(fixedFormData), &formDataMap); err != nil {
			s.logger.Error("解析表单数据失败", "error", err, "formData", formData)
			formDataMap = make(map[string]any)
		}
	} else {
		formDataMap = make(map[string]any)
	}

	// 调用 entityService 的方法获取下一个节点
	// 这里需要注入 entityService 或者将逻辑移到公共的地方
	return s.getNextNodeByOrder(nodes, currentNode, formDataMap)
}

// getNextNodeByOrder 按排序获取下一个节点
func (s *approvalService) getNextNodeByOrder(approvalNodes []*model.ApprovalNode, currentNode *model.ApprovalNode, formData map[string]any) (*model.ApprovalNode, error) {
	if currentNode == nil {
		return nil, fmt.Errorf("当前节点为空")
	}

	// 检查当前节点是否是条件分支中的节点
	if s.isConditionalBranchNode(approvalNodes, currentNode) {

		// 条件分支节点完成后，需要找到条件节点之后的下一个主流程节点
		conditionNode := s.findConditionNodeContaining(approvalNodes, currentNode)
		if conditionNode != nil {
			// 查找条件节点之后的下一个非条件分支节点
			for _, node := range approvalNodes {

				if node.SortOrder > conditionNode.SortOrder {
					// 跳过条件分支中的节点
					isConditionBranchNode := s.isConditionalBranchNode(approvalNodes, node)

					if isConditionBranchNode {
						continue
					}

					// 根据节点类型处理
					switch node.NodeType {
					case "CONDITION":
						return s.evaluateConditionNodeInApproval(approvalNodes, node, formData)
					case "APPROVAL":
						// TODO 如果有 condition_config 需要判断，调试是否满足
						s.logger.Debug("node: ", "node", node)
						// 如果满足则执行，如果不满足就跳过
						if node.ApproverType == "AUTO_REJECT" || node.ApproverType == "AUTO_APPROVE" {
							return nil, nil // 自动驳回或自动通过，流程结束
						}
						return node, nil
					case "CC":
						return node, nil
					case "END":
						return node, nil
					default:
						// 继续查找下一个节点
						return s.getNextNodeByOrder(approvalNodes, node, formData)
					}
				}
			}
		}

		// 如果有明确指定的下一个节点
		if currentNode.NodeCode != "" {
			return s.findNodeByCode(approvalNodes, currentNode.NodeCode), nil
		}

		// 否则流程结束
		return nil, nil
	}

	// 按 sort_order 排序查找下一个正常流程节点
	var nextNode *model.ApprovalNode
	for _, node := range approvalNodes {
		if node.SortOrder == currentNode.SortOrder+1 {
			nextNode = node
			break
		}
	}

	if nextNode == nil {
		return nil, nil // 已到达最后节点
	}

	// 根据节点类型进行处理
	switch nextNode.NodeType {
	case "CONDITION":
		// 条件节点：根据条件判断跳转到对应的审批节点
		return s.evaluateConditionNodeInApproval(approvalNodes, nextNode, formData)
	case "APPROVAL", "CC":
		// 审批节点和通知节点：直接返回
		return nextNode, nil
	case "END":
		// 结束节点：返回END节点，让调用方处理END任务创建
		return nextNode, nil
	case "START":
		// 开始节点：继续查找下一个节点
		return s.getNextNodeByOrder(approvalNodes, nextNode, formData)
	default:
		s.logger.Warn("未知的节点类型", "nodeType", nextNode.NodeType)
		// 对于未知节点类型，继续查找下一个节点
		return s.getNextNodeByOrder(approvalNodes, nextNode, formData)
	}
}

// evaluateConditionNodeInApproval 在审批服务中评估条件节点
func (s *approvalService) evaluateConditionNodeInApproval(approvalNodes []*model.ApprovalNode, conditionNode *model.ApprovalNode, formData map[string]any) (*model.ApprovalNode, error) {

	// 解析条件配置 - 支持新的数据结构
	var conditionConfig struct {
		Branches []struct {
			Name      string `json:"name"`
			Priority  int    `json:"priority,omitempty"` // 可选字段
			Condition struct {
				FieldName  string `json:"fieldName"`
				Operator   string `json:"operator"`
				FieldValue string `json:"fieldValue"`
			} `json:"condition"`
			Nodes []any `json:"nodes"` // 支持完整节点对象或简单节点信息
		} `json:"branches"`
	}

	if err := json.Unmarshal([]byte(conditionNode.ConditionConfig), &conditionConfig); err != nil {
		s.logger.Error("解析条件配置失败", "error", err)
		return s.getDefaultNextNodeInApproval(approvalNodes, conditionNode), nil
	}

	// 按优先级排序分支（如果没有优先级，保持原顺序）
	sort.Slice(conditionConfig.Branches, func(i, j int) bool {
		// 如果都没有设置优先级，保持原顺序
		if conditionConfig.Branches[i].Priority == 0 && conditionConfig.Branches[j].Priority == 0 {
			return i < j
		}
		return conditionConfig.Branches[i].Priority < conditionConfig.Branches[j].Priority
	})

	// 遍历条件分支，按顺序评估
	for _, branch := range conditionConfig.Branches {

		// 评估条件
		conditionMet := s.evaluateConditionInApproval(branch.Condition, formData)

		if conditionMet {
			// 查找目标节点
			for _, nodeData := range branch.Nodes {

				// 解析节点数据
				var nodeCode string
				// var nodeType string

				// 支持多种格式：简单格式、完整格式
				switch node := nodeData.(type) {
				case map[string]any:
					// 完整节点对象格式
					if code, ok := node["nodeCode"].(string); ok {
						nodeCode = code
					}
					// if nType, ok := node["nodeType"].(string); ok {
					// 	nodeType := nType
					// }
				case string:
					// 简单字符串格式（仅节点代码）
					nodeCode = node
				default:
					s.logger.Warn("不支持的节点数据格式", "nodeData", nodeData)
					continue
				}

				if nodeCode == "" {
					s.logger.Warn("节点代码为空", "nodeData", nodeData)
					continue
				}

				// 在所有审批节点中查找目标节点
				targetNode := s.findNodeByCode(approvalNodes, nodeCode)
				if targetNode != nil {

					// 支持所有类型的节点，不仅仅是 APPROVAL 和 CC
					switch targetNode.NodeType {
					case "APPROVAL":
						// 检查是否是自动驳回节点
						if targetNode.ApproverType == "AUTO_REJECT" {
							// 自动驳回节点，流程结束
							return nil, nil
						}
						return targetNode, nil
					case "CC", "END":
						return targetNode, nil
					case "CONDITION":
						// 如果目标节点也是条件节点，递归评估
						return s.evaluateConditionNodeInApproval(approvalNodes, targetNode, formData)
					default:
						s.logger.Warn("不支持的目标节点类型",
							"nodeType", targetNode.NodeType,
							"nodeCode", targetNode.NodeCode)
						continue
					}
				} else {
					s.logger.Warn("未找到目标节点", "nodeCode", nodeCode)
				}
			}
		}
	}

	// 如果没有匹配的条件分支，返回默认节点
	defaultNode := s.getDefaultNextNodeInApproval(approvalNodes, conditionNode)
	if defaultNode != nil {
		// 返回默认节点
	} else {
		s.logger.Error("没有找到默认节点，流程结束")
	}

	return defaultNode, nil
}

// isConditionalBranchNode 判断节点是否是条件分支中的节点
func (s *approvalService) isConditionalBranchNode(approvalNodes []*model.ApprovalNode, targetNode *model.ApprovalNode) bool {

	// 遍历所有条件节点，检查目标节点是否在其条件分支中
	for _, node := range approvalNodes {
		if node.NodeType == "CONDITION" && node.ConditionConfig != "" {
			// 解析条件配置
			var conditionConfig struct {
				Branches []struct {
					Name      string `json:"name"`
					Priority  int    `json:"priority,omitempty"`
					Condition struct {
						FieldName  string `json:"fieldName"`
						Operator   string `json:"operator"`
						FieldValue string `json:"fieldValue"`
					} `json:"condition"`
					Nodes []any `json:"nodes"`
				} `json:"branches"`
			}

			if err := json.Unmarshal([]byte(node.ConditionConfig), &conditionConfig); err != nil {
				s.logger.Error("解析条件配置失败",
					"nodeCode", node.NodeCode,
					"error", err)
				continue
			}

			// 检查目标节点是否在任何条件分支中
			for _, branch := range conditionConfig.Branches {
				for _, nodeData := range branch.Nodes {
					var nodeCode string

					switch node := nodeData.(type) {
					case map[string]any:
						if code, ok := node["nodeCode"].(string); ok {
							nodeCode = code
						}
					case string:
						nodeCode = node
					}

					if nodeCode == targetNode.NodeCode {
						return true
					}
				}
			}
		}
	}

	return false
}

// findConditionNodeContaining 找到包含指定节点的条件节点
func (s *approvalService) findConditionNodeContaining(approvalNodes []*model.ApprovalNode, targetNode *model.ApprovalNode) *model.ApprovalNode {
	for _, node := range approvalNodes {
		if node.NodeType == "CONDITION" && node.ConditionConfig != "" {
			// 解析条件配置
			var conditionConfig struct {
				Branches []struct {
					Name      string `json:"name"`
					Priority  int    `json:"priority,omitempty"`
					Condition struct {
						FieldName  string `json:"fieldName"`
						Operator   string `json:"operator"`
						FieldValue string `json:"fieldValue"`
					} `json:"condition"`
					Nodes []any `json:"nodes"`
				} `json:"branches"`
			}

			if err := json.Unmarshal([]byte(node.ConditionConfig), &conditionConfig); err != nil {
				continue
			}

			// 检查目标节点是否在任何条件分支中
			for _, branch := range conditionConfig.Branches {
				for _, nodeData := range branch.Nodes {
					var nodeCode string
					switch n := nodeData.(type) {
					case map[string]any:
						if code, ok := n["nodeCode"].(string); ok {
							nodeCode = code
						}
					case string:
						nodeCode = n
					}

					if nodeCode == targetNode.NodeCode {
						return node
					}
				}
			}
		}
	}
	return nil
}

// getNextNodeAfterCondition 获取条件节点之后的下一个节点
func (s *approvalService) getNextNodeAfterCondition(approvalNodes []*model.ApprovalNode, conditionNode *model.ApprovalNode, formData map[string]any) (*model.ApprovalNode, error) {

	// 查找条件节点之后的下一个主流程节点
	for _, node := range approvalNodes {
		if node.SortOrder > conditionNode.SortOrder {

			// 跳过条件分支中的节点
			isConditionBranchNode := s.isConditionalBranchNode(approvalNodes, node)

			if !isConditionBranchNode {

				// 根据节点类型处理
				switch node.NodeType {
				case "CONDITION":
					// 如果是另一个条件节点，递归评估
					return s.evaluateConditionNodeInApproval(approvalNodes, node, formData)
				case "APPROVAL":
					// 检查是否是自动驳回节点
					if node.ApproverType == "AUTO_REJECT" {
						// 自动驳回节点，流程结束
						return nil, nil
					}
					return node, nil
				case "CC":
					return node, nil
				case "END":
					return nil, nil // 流程结束
				default:
					// 继续查找下一个节点
					return s.getNextNodeAfterCondition(approvalNodes, node, formData)
				}
			}
		}
	}

	// 没有找到下一个节点，流程结束
	return nil, nil
}

// 辅助方法
func (s *approvalService) getUserNameByID(userID string) string {
	// TODO: 实现根据用户ID获取用户名的逻辑
	return userID // 简化处理
}

func (s *approvalService) getUserByRole(roleCode string) (string, string, error) {
	// TODO: 实现根据角色获取用户的逻辑
	return "role_user", "角色用户", nil
}

func (s *approvalService) getDefaultNextNodeInApproval(approvalNodes []*model.ApprovalNode, currentNode *model.ApprovalNode) *model.ApprovalNode {
	// 返回排序在当前节点之后的第一个审批节点
	for _, node := range approvalNodes {
		if node.SortOrder > currentNode.SortOrder && (node.NodeType == "APPROVAL" || node.NodeType == "CC") {
			return node
		}
	}
	return nil
}

func (s *approvalService) evaluateConditionInApproval(condition any, formData map[string]any) bool {
	// 将 condition 转换为 ConditionRule 结构
	conditionBytes, err := json.Marshal(condition)
	if err != nil {
		s.logger.Error("序列化条件失败", "error", err)
		return false
	}

	var conditionRule struct {
		FieldName  string `json:"fieldName"`
		Operator   string `json:"operator"`
		FieldValue string `json:"fieldValue"`
	}

	if err := json.Unmarshal(conditionBytes, &conditionRule); err != nil {
		s.logger.Error("反序列化条件失败", "error", err)
		return false
	}

	// 如果字段名为空，根据操作符判断是否为默认条件
	if conditionRule.FieldName == "" {
		// 空字段名通常表示默认条件或其他情况分支
		switch conditionRule.Operator {
		case "default", "eq", "":
			// 默认条件或空值等于条件，返回true
			return true
		default:
			// 其他操作符但字段名为空，视为不匹配
			return false
		}
	}

	// 从表单数据中获取字段值
	fieldValue, exists := formData[conditionRule.FieldName]
	if !exists {
		// 对于不存在的字段，只有 is_empty 或 is_null 条件为真
		switch conditionRule.Operator {
		case "is_empty", "is_null":
			return true
		default:
			return false
		}
	}

	// 转换为字符串进行比较
	fieldValueStr := s.convertToStringInApproval(fieldValue)
	expectedValue := conditionRule.FieldValue

	// 根据操作符进行比较
	switch conditionRule.Operator {
	case "eq", "equal", "==":
		result := s.evaluateEqualInApproval(fieldValueStr, expectedValue)
		return result
	case "ne", "not_equal", "!=":
		result := !s.evaluateEqualInApproval(fieldValueStr, expectedValue)
		return result
	case "gt", "greater_than", ">":
		result := s.evaluateGreaterThanInApproval(fieldValueStr, expectedValue)
		return result
	case "gte", "greater_than_equal", ">=":
		// 大于等于：大于或等于
		gtResult := s.evaluateGreaterThanInApproval(fieldValueStr, expectedValue)
		eqResult := s.evaluateEqualInApproval(fieldValueStr, expectedValue)
		result := gtResult || eqResult
		return result
	case "lt", "less_than", "<":
		result := s.evaluateLessThanInApproval(fieldValueStr, expectedValue)
		return result
	case "lte", "less_than_equal", "<=":
		// 小于等于：小于或等于
		ltResult := s.evaluateLessThanInApproval(fieldValueStr, expectedValue)
		eqResult := s.evaluateEqualInApproval(fieldValueStr, expectedValue)
		result := ltResult || eqResult
		return result
	case "contains", "like":
		result := strings.Contains(strings.ToLower(fieldValueStr), strings.ToLower(expectedValue))
		return result
	case "not_contains", "not_like":
		result := !strings.Contains(strings.ToLower(fieldValueStr), strings.ToLower(expectedValue))
		return result
	case "starts_with":
		result := strings.HasPrefix(strings.ToLower(fieldValueStr), strings.ToLower(expectedValue))
		return result
	case "ends_with":
		result := strings.HasSuffix(strings.ToLower(fieldValueStr), strings.ToLower(expectedValue))
		return result
	case "in":
		result := s.evaluateInInApproval(fieldValueStr, expectedValue)
		return result
	case "not_in":
		result := !s.evaluateInInApproval(fieldValueStr, expectedValue)
		return result
	case "is_empty", "is_null":
		result := s.evaluateIsEmptyInApproval(fieldValue)
		return result
	case "is_not_empty", "is_not_null":
		result := !s.evaluateIsEmptyInApproval(fieldValue)
		return result
	case "default":
		// 默认条件总是返回true
		return true
	default:
		s.logger.Warn("不支持的条件操作符", "operator", conditionRule.Operator)
		return false
	}
}

// convertToStringInApproval 将任意类型转换为字符串
func (s *approvalService) convertToStringInApproval(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// evaluateEqualInApproval 等于比较
func (s *approvalService) evaluateEqualInApproval(fieldValue, expectedValue string) bool {
	// 尝试数值比较
	if num1, err1 := strconv.ParseFloat(fieldValue, 64); err1 == nil {
		if num2, err2 := strconv.ParseFloat(expectedValue, 64); err2 == nil {
			return num1 == num2
		}
	}

	// 尝试布尔值比较
	if bool1, err1 := strconv.ParseBool(fieldValue); err1 == nil {
		if bool2, err2 := strconv.ParseBool(expectedValue); err2 == nil {
			return bool1 == bool2
		}
	}

	// 尝试日期比较
	if date1, err1 := time.Parse("2006-01-02", fieldValue); err1 == nil {
		if date2, err2 := time.Parse("2006-01-02", expectedValue); err2 == nil {
			return date1.Equal(date2)
		}
	}

	// 字符串比较（不区分大小写）
	return strings.EqualFold(fieldValue, expectedValue)
}

// evaluateGreaterThanInApproval 大于比较
func (s *approvalService) evaluateGreaterThanInApproval(fieldValue, expectedValue string) bool {
	// 尝试数值比较
	if num1, err1 := strconv.ParseFloat(fieldValue, 64); err1 == nil {
		if num2, err2 := strconv.ParseFloat(expectedValue, 64); err2 == nil {
			return num1 > num2
		}
	}

	// 尝试日期比较
	if date1, err1 := time.Parse("2006-01-02", fieldValue); err1 == nil {
		if date2, err2 := time.Parse("2006-01-02", expectedValue); err2 == nil {
			return date1.After(date2)
		}
	}

	// 字符串长度比较
	return len(fieldValue) > len(expectedValue)
}

// evaluateLessThanInApproval 小于比较
func (s *approvalService) evaluateLessThanInApproval(fieldValue, expectedValue string) bool {
	// 尝试数值比较
	if num1, err1 := strconv.ParseFloat(fieldValue, 64); err1 == nil {
		if num2, err2 := strconv.ParseFloat(expectedValue, 64); err2 == nil {
			return num1 < num2
		}
	}

	// 尝试日期比较
	if date1, err1 := time.Parse("2006-01-02", fieldValue); err1 == nil {
		if date2, err2 := time.Parse("2006-01-02", expectedValue); err2 == nil {
			return date1.Before(date2)
		}
	}

	// 字符串长度比较
	return len(fieldValue) < len(expectedValue)
}

// evaluateInInApproval 在列表中
func (s *approvalService) evaluateInInApproval(fieldValue, expectedValue string) bool {
	// 解析期望值为数组，支持逗号分隔或JSON数组格式
	var values []string

	// 尝试解析为JSON数组
	if strings.HasPrefix(expectedValue, "[") && strings.HasSuffix(expectedValue, "]") {
		if err := json.Unmarshal([]byte(expectedValue), &values); err == nil {
			for _, v := range values {
				if s.evaluateEqualInApproval(fieldValue, v) {
					return true
				}
			}
			return false
		}
	}

	// 按逗号分隔
	values = strings.Split(expectedValue, ",")
	for _, v := range values {
		v = strings.TrimSpace(v)
		if s.evaluateEqualInApproval(fieldValue, v) {
			return true
		}
	}
	return false
}

// evaluateIsEmptyInApproval 判断是否为空
func (s *approvalService) evaluateIsEmptyInApproval(fieldValue any) bool {
	if fieldValue == nil {
		return true
	}

	switch v := fieldValue.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		str := s.convertToStringInApproval(fieldValue)
		return strings.TrimSpace(str) == "" || str == "0" || str == "false"
	}
}

func (s *approvalService) approved(c *gin.Context, approval *model.Approval, userName string) error {
	tableDraft := approval.EntityCode + "_draft"

	draftMap := make(map[string]any)
	draftMap["approval_code"] = approval.Code
	s.logger.Debug("service-approval-approved", "draftMap", draftMap)

	// 查出需要处理的数据
	drafts, err := s.entityRepository.Find(tableDraft, "*", draftMap)
	if err != nil {
		return err
	}

	// 处理每个 draft 数据
	for _, draft := range drafts {
		s.logger.Debug("service-approval-approved222: ", "draft", draft)
		draft["draft_status"] = "Published"
		where := map[string]any{
			"id": draft["id"],
		}
		// 更新 draft 中的数据状态为 Published
		if err := s.entityRepository.Update(c, tableDraft, draft, where); err != nil {
			s.logger.Error("service-approval-approved", "update draft error", err)
		}
		draft["id"] = uint(draft["entity_id"].(int64))
		where["id"] = uint(draft["entity_id"].(int64))

		// 保存操作类型，后面清理 map
		operation := ""
		if v, ok := draft["operation"].(string); ok {
			operation = v
		}

		delete(draft, "entity_id")
		delete(draft, "approval_code")
		delete(draft, "draft_status")
		delete(draft, "date_version")
		delete(draft, "instance_code")
		delete(draft, "operation")

		tableCode := approval.EntityCode

		// 处理 data(entiry) 表中的数据
		// 注意：draft["operation"] 可能包含历史遗留的 operation 代码或 operation 名称
		switch operation {
		// "C", "MC" - 历史遗留的 Create 代码（根据 GetOperationInfo 应该是 "I"）
		// "Create", "BatchCreate" - operation 名称
		case "C", "MC", "Create", "BatchCreate":
			s.logger.Debug("service-approval-approved333: ", "tableCode", tableCode, "draft", draft)
			if err := s.entityRepository.Create(c, tableCode, draft); err != nil {
				return fmt.Errorf("create entity error: %s", err.Error())
			}
		// "U", "MU" - 历史遗留的 Update 代码（正确）
		// "Update", "BatchUpdate" - operation 名称
		case "U", "MU", "Update", "BatchUpdate":
			// 获取需要修改的数据(1条)
			id := where["id"].(uint)
			s.logger.Info("service-approval-approved", "update entity", tableCode, "id", id)
			origin, err := s.entityRepository.FindOne(tableCode, id)
			if err != nil {
				return err
			}

			draft["id"] = where["id"]
			draft["updated_by"] = draft["created_by"]
			draft["updated_at"] = time.Now()
			delete(draft, "created_by")
			delete(draft, "created_at")
			// 修改数据
			if err := s.entityRepository.Update(c, tableCode, draft, where); err != nil {
				return fmt.Errorf("change Entity Error: %s", err.Error())
			}

			fieldWhere := map[string]any{}
			fieldWhere["table_code"] = tableCode
			tableFields, err := s.tableFieldService.Find("", fieldWhere)
			if err != nil {
				return err
			}

			field := make(map[string]string)
			// 整理字段编码和字段名称
			for _, value := range tableFields {
				field[value.Code] = value.Name
			}
			// 添加系统字段的中文名称映射
			field["action"] = "操作代码"
			field["operation"] = "操作类型"
			field["status"] = "状态"
			field["send_status"] = "发送状态"
			exclude := []string{
				"id",
				"updated_at",
				"updated_by",
				"created_at",
				"created_by",
				"deleted_at",
			}
			for key := range draft {
				var excluded bool
				for _, v := range exclude {
					if key == v {
						excluded = true
					}
				}
				// 比较原来的值和现在的值有什么变化，如果并变化则记录日志
				if draft[key] != origin[key] && !excluded {
					var entityLog model.EntityLog
					entityLog.EntityID = uint(origin["id"].(uint64))
					entityLog.FieldCode = key
					entityLog.FieldName = field[key]
					entityLog.BeforeUpdate = s.convertToStringInApproval(origin[key])
					entityLog.AfterUpdate = s.convertToStringInApproval(draft[key])
					entityLog.UpdateBy = draft["updated_by"].(string)
					entityLog.Reason = approval.Description

					err := s.entityLogService.Create(c, tableCode, &entityLog)
					if err != nil {
						return err
					}
				}
			}
		// 注意：这里混合了历史遗留的 operation 代码和 operation 名称
		// "F", "MF" - 历史遗留的 Freeze 代码（根据 GetOperationInfo 应该是 "B"）
		// "T", "MT" - 历史遗留的 Terminate 代码
		// "L", "ML" - 历史遗留的 Lock 代码（根据 GetOperationInfo 应该是 "U"）
		// "UL" - 历史遗留的 Unlock 代码（根据 GetOperationInfo 应该是 "U"）
		// "MUL" - 历史遗留的批量 Unlock 代码（根据 GetOperationInfo 应该是 "U"）
		// "MUF" - 历史遗留的批量 Unfreeze 代码（根据 GetOperationInfo 应该是 "C"）
		// "D", "MD" - Delete 代码（正确）
		// "Freeze", "Unfreeze", "Delete", "Lock", "Unlock" - operation 名称
		case "F", "MF", "T", "MT", "L", "ML", "UL", "MUL", "MUF", "D", "MD", "Freeze", "Unfreeze", "Delete", "Lock", "Unlock", "BatchFreeze", "BatchUnfreeze", "BatchDelete", "BatchLock", "BatchUnlock":

			id := where["id"].(uint)
			origin, err := s.entityRepository.FindOne(tableCode, id)
			if err != nil {
				return err
			}

			entityMap := make(map[string]any)
			entityMap["status"] = draft["status"]
			entityMap["operation"] = draft["operation"]
			entityMap["updated_by"] = draft["created_by"]
			entityMap["updated_at"] = time.Now()
			if entityMap["operation"] == "D" || entityMap["operation"] == "MD" {
				entityMap["deleted_at"] = time.Now()
			}

			if err := s.entityRepository.Update(c, tableCode, entityMap, where); err != nil {
				return fmt.Errorf("change Entity Error: %s", err.Error())
			}

			fieldWhere := map[string]any{}
			fieldWhere["table_code"] = tableCode
			tableFields, err := s.tableFieldService.Find("", fieldWhere)
			if err != nil {
				return err
			}

			field := map[string]string{}
			for _, value := range tableFields {
				field[value.Code] = value.Name
			}
			// 添加系统字段的中文名称映射
			field["action"] = "操作代码"
			field["operation"] = "操作类型"
			field["status"] = "状态"
			field["send_status"] = "发送状态"
			exclude := []string{
				"id",
				"updated_at",
				"updated_by",
				"created_at",
				"created_by",
				"deleted_at",
			}
			for key := range entityMap {
				var excluded bool
				for _, v := range exclude {
					if key == v {
						excluded = true
					}
				}
				if entityMap[key] != origin[key] && !excluded {
					var entityLog model.EntityLog
					entityLog.EntityID = uint(origin["id"].(uint64))
					entityLog.FieldCode = key
					entityLog.FieldName = field[key]
					entityLog.BeforeUpdate = s.convertToStringInApproval(origin[key])
					entityLog.AfterUpdate = s.convertToStringInApproval(entityMap[key])
					entityLog.UpdateBy = entityMap["updated_by"].(string)
					entityLog.Reason = approval.Description

					err := s.entityLogService.Create(c, tableCode, &entityLog)
					if err != nil {
						return err
					}
				}
			}
		case "E":
		case "V":
		case "CA":
		case "TE":
		case "ME":
		}

		webhookWhere := map[string]any{
			"table_code": tableCode,
			"status":     "Normal",
		}
		webhooks, err := s.webhookService.Find("", webhookWhere)
		if err != nil {
			return err
		}

		for _, webhookItem := range webhooks {
			webhookReq := model.WebhookReq{
				HookID:       webhookItem.ID,
				ApprovalCode: approval.Code,
				TableCode:    tableCode,
				EntityID:     uint(draft["id"].(int64)),
			}
			s.SaveToQueue(webhookReq)
		}

	}

	return nil
}

func (s *approvalService) SaveToQueue(webhookReq model.WebhookReq) error {
	return s.approvalRepository.SaveToQueue(webhookReq)
}

func (s *approvalService) handleTransfer(c *gin.Context, task *model.ApprovalTask, comment string) error {
	// 处理任务转交逻辑
	// 这里需要根据具体需求实现转交逻辑

	// 暂时返回 nil，后续根据需求完善
	return nil
}

// ==================== 审批工作流相关方法 (从entity.go迁移) ====================

// GetOperationInfo 获取操作信息
func (s *approvalService) GetOperationInfo(operation string, operationInfo *map[string]string) error {
	// ==============================
	// operationType
	// SAP-ECC中习惯使用的标识方法。
	// 新建: Create
	// 修改: Update
	// 冻结: Freeze
	// 解冻: UnFreeze
	// 锁定: Lock
	// 解锁: Unlock
	// 删除: Delete
	// 扩展: Extension
	// 作废: Void
	// 撤销: Cancel
	// 终止: Terminate
	// 批量创建: BatchCreate
	// 批量修改: BatchUpdate
	// 批量冻结: BatchFreeze
	// 批量解冻: BatchUnfreeze
	// 批量锁定: BatchLock
	// 批量解锁: BatchUnlock
	// 批量删除: BatchDeletion
	// 批量扩展: BatchExtension
	// ==============================
	// action，是针对已有系统的另外一套操作类型
	// I插入，U更新，B冻结，C解冻，D删除
	// ==============================
	// status
	// 正常: Normal，正常的状态都使用这个，解锁、解冻后的状态也是这个。option操作了操作的类型。
	// 已冻结: Frozen
	// 已锁定: Locked
	// 已删除: Deleted
	// 已作废: Voided
	// 已终止: Terminated
	// var operationName, status, action string
	info := *operationInfo
	switch operation {
	case "Create": // 新建
		info["operationName"] = "新建"
		info["status"] = "Normal"
		info["action"] = "I"
	case "Update": // 修改
		info["operationName"] = "修改"
		info["status"] = "Normal"
		info["action"] = "U"
	case "Freeze": // 冻结
		info["operationName"] = "冻结"
		info["status"] = "Frozen"
		info["action"] = "B"
	case "Unfreeze": // 解冻
		info["operationName"] = "解冻"
		info["status"] = "Normal"
		info["action"] = "C"
	case "Lock": // 锁定
		info["operationName"] = "锁定"
		info["status"] = "Locked"
		info["action"] = "U"
	case "Unlock": // 解锁
		info["operationName"] = "解锁"
		info["status"] = "Normal"
		info["action"] = "U"
	case "Delete": // 删除
		info["operationName"] = "删除"
		info["status"] = "Deleted"
		info["action"] = "D"
	case "Extend": // 扩展
		info["operationName"] = "扩展"
		info["status"] = "Extended"
		info["action"] = "I"
	case "Void": // 作废
		info["operationName"] = "作废"
		info["status"] = "Voided"
		info["action"] = "U"
	case "Cancel": // 撤销
		info["operationName"] = "撤销"
		info["status"] = "Normal"
		info["action"] = "U"
	case "Terminate": // 终止
		info["operationName"] = "终止"
		info["status"] = "Terminated"
		info["action"] = "U"
	case "BatchCreate": // 批量创建
		info["operationName"] = "批量创建"
		info["status"] = "Normal"
		info["action"] = "I"
	case "BatchUpdate": // 批量更新
		info["operationName"] = "批量更新"
		info["status"] = "Normal"
		info["action"] = "U"
	case "BatchFreeze": // 批量冻结
		info["operationName"] = "批量冻结"
		info["status"] = "Frozen"
		info["action"] = "B"
	case "BatchUnfreeze": // 批量解冻
		info["operationName"] = "批量解冻"
		info["status"] = "Normal"
		info["action"] = "C"
	case "BatchLock": // 批量锁定
		info["operationName"] = "批量锁定"
		info["status"] = "Locked"
		info["action"] = "U"
	case "BatchUnlock": // 批量解锁
		info["operationName"] = "批量解锁"
		info["status"] = "Normal"
		info["action"] = "U"
	case "BatchDelete": // 批量删除
		info["operationName"] = "批量删除"
		info["status"] = "Deleted"
		info["action"] = "D"
	case "BatchExtend": // 批量扩展
		info["operationName"] = "批量扩展"
		info["status"] = "Normal"
		info["action"] = "I"
	default:
		return fmt.Errorf("operation %s not in list。", operation)
	}

	return nil
}

// CreateApprovalFlow 创建审批流程
func (s *approvalService) CreateApprovalFlow(c *gin.Context, tableCode string, approvalInfo map[string]string, formData map[string]any) error {
	// 1. 根据 tableCode 和 operation 获取审批流定义映射
	tableApprovalDefs, err := s.tableApprovalDefinitionRepository.List(tableCode, approvalInfo["operation"])
	if err != nil {
		s.logger.Error("获取审批映射失败", "error", err)
		return err
	}

	if len(tableApprovalDefs) == 0 {
		return nil // 跳过审批
	}

	tableApprovalDef := tableApprovalDefs[0]

	// 2. 获取审批流定义
	approvalDef, err := s.approvalDefinitionRepository.First(map[string]any{
		"code":   tableApprovalDef.ApprovalDefCode,
		"status": "Normal",
	})
	if err != nil {
		return err
	}

	// 3. 检查平台并分发
	if approvalDef.Platform == "Feishu" {
		// 补充缺失的 entityCode
		approvalInfo["entityCode"] = tableCode
		return s.createFeishuInstanceFromInfo(c, approvalDef, approvalInfo, formData)
	}

	// 4. 获取本地审批节点列表 (仅 Builtin 需要)
	approvalNodes, err := s.approvalNodeRepository.FindActiveByApprovalDefCode(approvalDef.Code)
	if err != nil {
		return err
	}

	// 4. 创建审批实例
	return s.CreateApprovalInstance(c, approvalDef, approvalNodes, approvalInfo)
}

// createFeishuInstanceFromInfo 专门用于从 approvalInfo map 创建飞书审批实例
func (s *approvalService) createFeishuInstanceFromInfo(c *gin.Context, def *model.ApprovalDefinition, approvalInfo map[string]string, formData map[string]any) error {
	ctx := c.Request.Context()

	// 构造标题: 审批定义名称 - 操作名称
	title := def.Name
	if opName, ok := approvalInfo["operationName"]; ok && opName != "" {
		title = fmt.Sprintf("%s-%s", def.Name, opName)
	}

	feishuUserID, err := s.getFeishuID(c)
	if err != nil {
		return fmt.Errorf("failed to get initiator feishu ID: %w", err)
	}

	// 生成飞书表单值
	// fix: Form中添加下面的信息，把字段整理成换行的json内容
	// 替换测试测试为 字段名：字段值\n字段名：字段值\n字段名：字段值
	formValues := "[]"
	if formData != nil {
		var summary string
		operation := approvalInfo["operation"]

		// 检查是否为批量操作
		isBatch := false
		batchCount := 0
		var actionName string

		switch operation {
		case "Import":
			isBatch = true
			actionName = "导入"
			if records, ok := formData["records"].([]any); ok {
				batchCount = len(records)
			}
		case "BatchCreate":
			if records, ok := formData["records"].([]any); ok {
				isBatch = true
				actionName = "批量创建"
				batchCount = len(records)
			}
		case "BatchUpdate":
			if records, ok := formData["records"].([]any); ok {
				isBatch = true
				actionName = "批量更新"
				batchCount = len(records)
			} else if ids, ok := formData["ids"].([]any); ok {
				isBatch = true
				actionName = "批量更新"
				batchCount = len(ids)
			}
		case "BatchFreeze":
			isBatch = true
			actionName = "冻结"
			if ids, ok := formData["ids"].([]any); ok {
				batchCount = len(ids)
			}
		case "BatchUnfreeze":
			isBatch = true
			actionName = "解冻"
			if ids, ok := formData["ids"].([]any); ok {
				batchCount = len(ids)
			}
		case "BatchDelete":
			isBatch = true
			actionName = "删除"
			if ids, ok := formData["ids"].([]any); ok {
				batchCount = len(ids)
			}
		}

		var originalValues map[string]any

		// Debug: Log entry
		s.logger.Info("createFeishuInstanceFromInfo: checking for single record optimization", "isBatch", isBatch, "batchCount", batchCount, "operation", operation)

		if isBatch && batchCount == 1 {
			// 如果批量操作只有一条记录，尝试转换为单条显示
			if records, ok := formData["records"].([]any); ok && len(records) > 0 {
				if record, ok := records[0].(map[string]any); ok {
					s.logger.Info("createFeishuInstanceFromInfo: processing single record from records", "record_op", record["operation"])

					// 将记录合并到 formData
					for k, v := range record {
						formData[k] = v
					}
					isBatch = false // 视为非批量操作，显示详细字段
				}
			} else if operation == "BatchFreeze" || operation == "BatchUnfreeze" || operation == "BatchDelete" || (operation == "BatchUpdate" && formData["ids"] != nil) {
				// 获取第一条记录的ID
				if ids, ok := formData["ids"].([]any); ok && len(ids) > 0 {
					var id uint
					switch v := ids[0].(type) {
					case float64:
						id = uint(v)
					case int:
						id = uint(v)
					case uint:
						id = v
					case string:
						// try parse
						if val, err := strconv.ParseUint(v, 10, 64); err == nil {
							id = uint(val)
						}
					}

					if id > 0 && approvalInfo["entityCode"] != "" {
						// 查询详细数据 using entityRepository
						// Note: entityRepository.FindOne returns map[string]any
						entityData, err := s.entityRepository.FindOne(approvalInfo["entityCode"], id)
						if err == nil {
							// 将实体数据合并到 formData
							for k, v := range entityData {
								formData[k] = v
							}
							isBatch = false // 视为非批量操作，显示详细字段
						} else {
							s.logger.Warn("failed to fetch single entity for summary", "err", err)
						}
					}
				}
			}
		}

		// 通用逻辑: 如果是更新操作(包括单条更新 或 优化后的批量更新)，查询原值用于对比
		// 此时 formData 应该包含了 ID
		if !isBatch && strings.Contains(operation, "Update") {
			var id uint
			// 尝试解析 ID
			if val, ok := formData["entity_id"]; ok {
				idInt, _ := strconv.ParseUint(fmt.Sprintf("%v", val), 10, 64)
				id = uint(idInt)
			} else if val, ok := formData["id"]; ok {
				// 注意: 单条修改有时 id 是 entity_id, 有时是 draft id?
				// 通常 formData 来自 entityMap, id 可能是 database id
				idInt, _ := strconv.ParseUint(fmt.Sprintf("%v", val), 10, 64)
				id = uint(idInt)
			}

			if id > 0 && approvalInfo["entityCode"] != "" {
				s.logger.Info("createFeishuInstanceFromInfo: fetching original values for diff", "id", id, "entityCode", approvalInfo["entityCode"])
				if existing, err := s.entityRepository.FindOne(approvalInfo["entityCode"], id); err == nil {
					originalValues = existing
					s.logger.Info("createFeishuInstanceFromInfo: original values found", "count", len(existing))
				} else {
					s.logger.Warn("createFeishuInstanceFromInfo: failed to find original values", "error", err)
				}
			}
		}

		if isBatch {
			summary = fmt.Sprintf("批量%s %d 条，不方便在这里显示，请到MDM系统查看明细", actionName, batchCount)
		} else {
			// 非批量操作，使用详细字段列表
			// 获取字段定义以映射名称并排序
			var tableFields []*model.TableField
			fieldMap := make(map[string]string)
			entityCode := approvalInfo["entityCode"]
			if entityCode != "" {
				// Find 接口签名: Find(selectString string, where map[string]any) -> []*model.TableField
				fields, err := s.tableFieldService.Find("", map[string]any{
					"table_code": entityCode,
				})
				if err == nil {
					tableFields = fields
					// sort by Sort field
					sort.Slice(tableFields, func(i, j int) bool {
						return tableFields[i].Sort < tableFields[j].Sort
					})

					for _, f := range fields {
						// f is *model.TableField, access fields directly
						if f.Code != "" {
							name := f.Name
							if name == "" {
								name = f.Code
							}
							fieldMap[f.Code] = name
						}
					}
				} else {
					s.logger.Warn("failed to fetch table fields for summary", "entityCode", entityCode, "error", err)
				}
			}

			var sb strings.Builder
			processedKeys := make(map[string]bool)

			// 1. 按照由于定义的顺序遍历
			for _, field := range tableFields {
				val, ok := formData[field.Code]
				if !ok {
					continue
				}

				// skip system fields if defined in table fields(unlikely but safe)
				if field.Code == "id" || field.Code == "approval_code" || field.Code == "operation" || field.Code == "action" || field.Code == "entity_id" || field.Code == "title" {
					processedKeys[field.Code] = true
					continue
				}

				valStr := fmt.Sprintf("%v", val)

				// 如果有原值且不一致，显示原值
				if originalValues != nil {
					if oldVal, exists := originalValues[field.Code]; exists {
						oldValStr := fmt.Sprintf("%v", oldVal)
						if oldValStr != valStr {
							s.logger.Info("createFeishuInstanceFromInfo: diff detected", "field", field.Code, "new", valStr, "old", oldValStr)
							valStr = fmt.Sprintf("%s(原值：%s)", valStr, oldValStr)
						}
					}
				}

				displayName := field.Code
				if name, ok := fieldMap[field.Code]; ok && name != "" {
					displayName = name
				}
				sb.WriteString(fmt.Sprintf("%s: %s\n", displayName, valStr))
				processedKeys[field.Code] = true
			}

			// 2. 遍历剩余的 formData (未在定义中找到的字段)
			for k, v := range formData {
				if processedKeys[k] {
					continue
				}
				// skip system fields
				if k == "id" || k == "approval_code" || k == "operation" || k == "action" || k == "entity_id" || k == "title" {
					continue
				}
				// skip additional internal fields
				if k == "records" || k == "ids" || k == "created_at" || k == "updated_at" || k == "deleted_at" {
					continue
				}
				if k == "status" || k == "created_by" || k == "updated_by" || k == "draft_status" || k == "send_status" {
					continue
				}
				valStr := fmt.Sprintf("%v", v)

				// 如果有原值且不一致，显示原值
				if originalValues != nil {
					if oldVal, exists := originalValues[k]; exists {
						oldValStr := fmt.Sprintf("%v", oldVal)
						if oldValStr != valStr {
							s.logger.Info("createFeishuInstanceFromInfo: diff detected (uncategorized)", "field", k, "new", valStr, "old", oldValStr)
							valStr = fmt.Sprintf("%s(原值：%s)", valStr, oldValStr)
						}
					}
				}

				// 使用字段名称，如果不存在则回退到字段代码
				displayName := k
				if name, ok := fieldMap[k]; ok && name != "" {
					displayName = name
				}
				sb.WriteString(fmt.Sprintf("%s: %s\n", displayName, valStr))
			}
			summary = sb.String()
		}

		if summary == "" {
			summary = "无详细内容"
		}

		// Construct fixed JSON for the summary widget
		// Dynamic lookup for the first textarea widget
		textareaID := s.feishuIntegrationService.ExtractFirstTextareaID(def.FormData)
		if textareaID == "" {
			s.logger.Error("failed to find any textarea widget for summary injection", "defCode", def.Code)
			// fallback or error? user request implies we should use this mechanism.
			// for now, let's just return error or skip injection to avoid 60022 error (invalid widget id)
			return fmt.Errorf("审批定义中未找到 textarea 控件，无法注入审批内容摘要。请在飞书审批表单中添加一个多行文本控件")
		}

		widgetData := []map[string]string{
			{
				"id":    textareaID,
				"type":  "textarea",
				"value": summary,
			},
		}
		jsonBytes, _ := json.Marshal(widgetData)
		formValues = string(jsonBytes)
	}

	// 1. 调用飞书服务创建实例
	externalID, err := s.feishuIntegrationService.CreateApprovalInstance(ctx, def.Code, formValues, feishuUserID, title)
	if err != nil {
		// 如果错误提示用户ID不正确 (1390001)，给出更友好的提示
		if strings.Contains(err.Error(), "1390001") || strings.Contains(err.Error(), "用户ID不正确") || strings.Contains(err.Error(), "invalid user_id") {
			return fmt.Errorf("创建飞书审批失败: 飞书提示用户ID不存在。请确认本地用户的 username [%s] 和飞书中的 user_id 一致。原始错误: %w", feishuUserID, err)
		}
		return fmt.Errorf("create feishu instance failed: %w", err)
	}

	// 2. 创建本地审批记录
	approvalId := s.globalIdService.GetNewID("approval")
	approvalInstance := model.Approval{
		ID:                 approvalId,
		Code:               approvalInfo["approvalCode"],
		Title:              title,
		ApprovalDefCode:    def.Code,
		EntityCode:         approvalInfo["entityCode"],
		SerialNumber:       s.GenerateSerialNumber(),
		FormData:           def.FormData,
		Description:        approvalInfo["reason"],
		Status:             model.ApprovalStatusPending,
		ExternalInstanceID: externalID,
		CreatedBy:          c.GetString("user_name"),
		UpdatedBy:          c.GetString("user_name"),
	}

	return s.approvalRepository.Create(c, &approvalInstance)
}

// getFeishuID 从上下文获取用户ID并查询对应的飞书ID
func (s *approvalService) getFeishuID(c *gin.Context) (string, error) {
	// 从上下文中直接获取用户名 (由 JWT 中间件设置)
	username := c.GetString("user_name")
	if username == "" {
		// 备选: 如果上下文没有 user_name (罕见情况), 尝试通过 user_id 查询
		userIdVal, exists := c.Get("user_id")
		if !exists {
			return "", fmt.Errorf("user not logged in")
		}

		var uid uint
		switch v := userIdVal.(type) {
		case string:
			id, _ := strconv.ParseUint(v, 10, 32)
			uid = uint(id)
		case uint:
			uid = v
		case float64:
			uid = uint(v)
		default:
			return "", fmt.Errorf("invalid user id type in context")
		}

		user, err := s.userRepository.FindOne(uid)
		if err != nil {
			return "", fmt.Errorf("user not found: %w", err)
		}
		username = user.Username
	}

	// 业务约定: 直接使用 Username 作为飞书 UserID
	if username == "" {
		return "", fmt.Errorf("user username is empty")
	}
	return username, nil
}

// CreateApprovalInstance 创建审批实例的具体逻辑
func (s *approvalService) CreateApprovalInstance(c *gin.Context, approvalDef *model.ApprovalDefinition,
	approvalNodes []*model.ApprovalNode, approvalInfo map[string]string,
) error {
	// 1. 找到开始节点
	var startNode *model.ApprovalNode
	for _, node := range approvalNodes {
		if node.NodeType == "START" && node.SortOrder == 0 {
			startNode = node
			break
		}
	}

	if startNode == nil {
		return fmt.Errorf("未找到开始节点")
	}

	// 2. 创建审批实例
	approvalId := s.globalIdService.GetNewID("approval")
	// now := time.Now()

	approvalInstance := model.Approval{
		ID:              approvalId,
		Code:            approvalInfo["approvalCode"],
		Title:           approvalDef.Name + "-" + approvalInfo["operationName"],
		ApprovalDefCode: approvalDef.Code,
		EntityCode:      approvalInfo["entityCode"],
		// CurrentNodeID:   strconv.FormatUint(uint64(startNode.ID), 10), // 将在创建任务后设置
		// CurrentNodeName: startNode.NodeName, // 将在创建任务后设置
		SerialNumber: s.GenerateSerialNumber(),
		FormData:     approvalDef.FormData,
		Description:  approvalInfo["reason"],
		Status:       "Pending",
		// StartedAt:     &now,
		// ApplicantID:   c.GetString("user_name"),
		// ApplicantName: c.GetString("user_name"),
		CreatedBy: c.GetString("user_name"),
		UpdatedBy: c.GetString("user_name"),
	}

	if err := s.approvalRepository.Create(c, &approvalInstance); err != nil {
		return err
	}

	// 3. 创建开始任务
	if err := s.CreateStartTask(c, startNode, approvalInfo); err != nil {
		return err
	}

	// 4. 创建下一个审批任务
	nextNode, err := s.GetNextApprovalNode(approvalNodes, startNode, nil)
	if err != nil {
		return err
	}

	if nextNode != nil {
		createdTask, err := s.CreateApprovalTask(c, nextNode, approvalInfo)
		if err != nil {
			return err
		}

		// 更新审批实例的当前任务信息
		approvalInstance.CurrentTaskID = strconv.FormatUint(uint64(createdTask.ID), 10)
		approvalInstance.CurrentTaskName = createdTask.NodeName
		return s.approvalRepository.Update(c, &approvalInstance)
	}
	return nil
}

// CreateStartTask 创建开始任务
func (s *approvalService) CreateStartTask(c *gin.Context, startNode *model.ApprovalNode, approvalInfo map[string]string) error {
	// taskId := s.globalIdService.GetNewID("approval_task")
	// now := time.Now()

	startTask := model.ApprovalTask{
		// ID:           taskId,
		ApprovalCode: approvalInfo["approvalCode"],
		NodeCode:     startNode.NodeCode,
		NodeName:     startNode.NodeName,
		AssigneeID:   c.GetString("user_id"),
		AssigneeName: c.GetString("user_name"),
		Comment:      approvalInfo["reason"],
		Status:       model.TaskStatusDone, // 开始任务直接完成
		CreatedBy:    c.GetString("user_name"),
		UpdatedBy:    c.GetString("user_name"),
		// StartedAt:    &now,
		// CompletedAt:  &now,
	}

	return s.approvalTaskService.Create(&startTask)
}

// CreateApprovalTask 创建审批任务
func (s *approvalService) CreateApprovalTask(c *gin.Context, node *model.ApprovalNode, approvalInfo map[string]string) (*model.ApprovalTask, error) {
	// taskId := s.globalIdService.GetNewID("approval_task")
	// now := time.Now()

	// 根据审批节点类型处理审批人
	var assigneeID, assigneeName string
	if err := s.parseApproverConfig(node, &assigneeID, &assigneeName); err != nil {
		s.logger.Error("解析审批人配置失败", "error", err)
		assigneeID = "system"
		assigneeName = "系统"
	}

	// 创建审批任务
	approvalTask := model.ApprovalTask{
		// ID:           taskId,
		ApprovalCode: approvalInfo["approvalCode"],
		NodeCode:     node.NodeCode,
		NodeName:     node.NodeName,
		AssigneeID:   assigneeID,
		AssigneeName: assigneeName,
		Comment:      "",
		Status:       model.TaskStatusPending,
		CreatedBy:    c.GetString("user_name"),
		UpdatedBy:    c.GetString("user_name"),
		// StartedAt:    &now,
	}

	// 处理特殊节点类型
	switch node.ApproverType {
	case "AUTO_APPROVE":
		approvalTask.Status = model.TaskStatusApproved
		approvalTask.Comment = "系统自动通过"
		// approvalTask.CompletedAt = &now
	case "AUTO_REJECT":
		approvalTask.Status = model.TaskStatusRejected
		approvalTask.Comment = "系统自动驳回"
		// approvalTask.CompletedAt = &now
	}

	if err := s.approvalTaskService.Create(&approvalTask); err != nil {
		return nil, err
	}
	return &approvalTask, nil
}

// GenerateSerialNumber 生成审批单编号
func (s *approvalService) GenerateSerialNumber() string {
	// 生成审批单编号：AP + 年月日 + 6位随机数
	now := time.Now()
	dateStr := now.Format("20060102")
	randomStr := fmt.Sprintf("%06d", now.UnixNano()%1000000)
	return fmt.Sprintf("AP%s%s", dateStr, randomStr)
}

// GetNextApprovalNode 获取下一个审批节点
func (s *approvalService) GetNextApprovalNode(approvalNodes []*model.ApprovalNode, currentNode *model.ApprovalNode, formData map[string]any) (*model.ApprovalNode, error) {
	if currentNode == nil {
		return nil, fmt.Errorf("current node is nil")
	}

	// 按 sort_order 排序查找下一个节点
	var nextNode *model.ApprovalNode
	for _, node := range approvalNodes {
		if node.SortOrder == currentNode.SortOrder+1 {
			nextNode = node
			break
		}
	}

	if nextNode == nil {
		return nil, nil // 已到达最后节点
	}

	// 根据节点类型进行处理
	switch nextNode.NodeType {
	case "CONDITION":
		// 条件节点：根据条件判断跳转到对应的审批节点
		return s.EvaluateConditionNode(approvalNodes, nextNode, formData)
	case "APPROVAL":
		// 审批节点：直接返回
		return nextNode, nil
	case "CC":
		// 通知节点：直接返回，由调用方处理通知逻辑
		return nextNode, nil
	case "END":
		// 结束节点：返回END节点，让调用方处理END任务创建
		return nextNode, nil
	case "START":
		// 开始节点：继续查找下一个节点
		return s.GetNextApprovalNode(approvalNodes, nextNode, formData)
	default:
		s.logger.Warn("未知的节点类型", "nodeType", nextNode.NodeType)
		// 对于未知节点类型，继续查找下一个节点
		return s.GetNextApprovalNode(approvalNodes, nextNode, formData)
	}
}

// // ConditionConfig 条件配置结构
// type ConditionConfig struct {
// 	Branches []ConditionBranch `json:"branches"`
// }

// // ConditionBranch 条件分支
// type ConditionBranch struct {
// 	Name      string          `json:"name"`
// 	Condition ConditionRule   `json:"condition"`
// 	Nodes     []ConditionNode `json:"nodes"`
// }

// // ConditionRule 条件规则
// type ConditionRule struct {
// 	FieldName  string `json:"fieldName"`
// 	Operator   string `json:"operator"`
// 	FieldValue string `json:"fieldValue"`
// }

// // ConditionNode 条件节点
// type ConditionNode struct {
// 	NodeCode     string `json:"nodeCode"`
// 	NodeName     string `json:"nodeName"`
// 	NodeType     string `json:"nodeType"`
// 	ApproverType string `json:"approverType"`
// }

// // ConditionExpression 复杂条件表达式结构
// type ConditionExpression struct {
// 	Type      string                 `json:"type"`      // "simple", "and", "or"
// 	Condition *ConditionRule         `json:"condition"` // 简单条件
// 	Children  []*ConditionExpression `json:"children"`  // 子表达式
// }

// EvaluateConditionNode 评估条件节点并返回对应的审批节点
func (s *approvalService) EvaluateConditionNode(approvalNodes []*model.ApprovalNode, conditionNode *model.ApprovalNode, formData map[string]any) (*model.ApprovalNode, error) {
	// 解析 condition_config JSON
	var conditionConfig struct {
		Branches []struct {
			Name      string `json:"name"`
			Priority  int    `json:"priority"`
			Condition struct {
				FieldName  string `json:"fieldName"`
				Operator   string `json:"operator"`
				FieldValue string `json:"fieldValue"`
			} `json:"condition"`
			Nodes []struct {
				NodeCode string `json:"nodeCode"`
				NodeType string `json:"nodeType"`
			} `json:"nodes"`
		} `json:"branches"`
	}

	if err := json.Unmarshal([]byte(conditionNode.ConditionConfig), &conditionConfig); err != nil {
		s.logger.Error("解析条件配置失败", "error", err)
		// 如果解析失败，返回下一个审批节点作为默认行为
		return s.getDefaultNextNode(approvalNodes, conditionNode), nil
	}

	// 按优先级排序分支
	sort.Slice(conditionConfig.Branches, func(i, j int) bool {
		return conditionConfig.Branches[i].Priority < conditionConfig.Branches[j].Priority
	})

	// 遍历条件分支，按优先级评估
	for _, branch := range conditionConfig.Branches {

		var conditionResult bool

		// 评估条件
		conditionResult = s.evaluateConditionInApproval(branch.Condition, formData)

		// 如果条件匹配，查找分支中的目标节点
		if conditionResult {

			// 查找分支中的目标节点
			for _, conditionNodeInfo := range branch.Nodes {
				targetNode := s.FindNodeByCode(approvalNodes, conditionNodeInfo.NodeCode)
				if targetNode != nil {
					// 检查目标节点的 approver_config 条件
					if targetNode.ApproverConfig != "" {
						// 解析 approver_config 中的条件
						conditionMet, err := s.evaluateApproverConfigCondition(targetNode, formData)
						if err != nil {
							s.logger.Error("评估审批人配置条件失败", "error", err)
							continue
						}
						if !conditionMet {
							// 审批人配置条件不满足，跳过节点
							continue
						}
					}

					// 根据目标节点类型处理
					switch targetNode.NodeType {
					case "APPROVAL":
						return targetNode, nil
					case "CC":
						return targetNode, nil
					case "END":
						return nil, nil // 流程结束
					case "CONDITION":
						// 如果目标是另一个条件节点，递归评估
						return s.EvaluateConditionNode(approvalNodes, targetNode, formData)
					default:
						s.logger.Warn("条件分支指向未知节点类型",
							"nodeCode", conditionNodeInfo.NodeCode,
							"nodeType", targetNode.NodeType)
					}
				} else {
					s.logger.Warn("未找到条件分支指向的节点",
						"nodeCode", conditionNodeInfo.NodeCode)
				}
			}
		}
	}

	// 如果没有匹配的条件，返回默认节点
	return s.getDefaultNextNode(approvalNodes, conditionNode), nil
}

// evaluateApproverConfigCondition 评估审批人配置中的条件
func (s *approvalService) evaluateApproverConfigCondition(node *model.ApprovalNode, formData map[string]any) (bool, error) {
	if node.ApproverConfig == "" {
		return true, nil // 没有配置条件，默认通过
	}

	// 解析审批人配置
	var approverConfig struct {
		Type       string `json:"type"`
		Condition  any    `json:"condition"`  // 条件配置
		Expression any    `json:"expression"` // 表达式配置
	}

	if err := json.Unmarshal([]byte(node.ApproverConfig), &approverConfig); err != nil {
		s.logger.Error("解析审批人配置条件失败", "error", err)
		return false, err
	}

	// 检查是否有条件配置
	if approverConfig.Condition != nil {
		// 使用现有的条件评估逻辑
		return s.evaluateConditionInApproval(approverConfig.Condition, formData), nil
	}

	if approverConfig.Expression != nil {
		// 解析表达式配置
		expressionBytes, err := json.Marshal(approverConfig.Expression)
		if err != nil {
			s.logger.Error("序列化表达式配置失败", "error", err)
			return false, err
		}

		var expression ConditionExpression
		if err := json.Unmarshal(expressionBytes, &expression); err != nil {
			s.logger.Error("解析表达式配置失败", "error", err)
			return false, err
		}

		// 使用现有的复杂表达式评估逻辑
		return s.evaluateComplexCondition(&expression, formData), nil
	}

	// 没有条件配置，默认通过
	return true, nil
}

// FindNodeByCode 根据节点编码查找节点
func (s *approvalService) FindNodeByCode(approvalNodes []*model.ApprovalNode, nodeCode string) *model.ApprovalNode {
	for _, node := range approvalNodes {
		if node.NodeCode == nodeCode {
			return node
		}
	}
	return nil
}

// getDefaultNextNode 获取默认的下一个节点
func (s *approvalService) getDefaultNextNode(approvalNodes []*model.ApprovalNode, currentNode *model.ApprovalNode) *model.ApprovalNode {
	// 返回排序在当前节点之后的第一个审批节点或通知节点
	for _, node := range approvalNodes {
		if node.SortOrder > currentNode.SortOrder && (node.NodeType == "APPROVAL" || node.NodeType == "CC") {
			return node
		}
	}
	return nil
}

/*
条件评估支持的格式示例：

1. 简单条件：
{
  "fieldName": "amount",
  "operator": "gt",
  "fieldValue": "1000"
}

2. 复杂表达式：
{
  "type": "and",
  "children": [
    {
      "type": "simple",
      "condition": {
        "fieldName": "amount",
        "operator": "gt",
        "fieldValue": "1000"
      }
    },
    {
      "type": "or",
      "children": [
        {
          "type": "simple",
          "condition": {
            "fieldName": "department",
            "operator": "eq",
            "fieldValue": "finance"
          }
        },
        {
          "type": "simple",
          "condition": {
            "fieldName": "level",
            "operator": "gte",
            "fieldValue": "3"
          }
        }
      ]
    }
  ]
}

3. 条件分支配置：
{
  "branches": [
    {
      "name": "高金额审批",
      "priority": 1,
      "expression": {
        "type": "and",
        "children": [
          {
            "type": "simple",
            "condition": {
              "fieldName": "amount",
              "operator": "gt",
              "fieldValue": "10000"
            }
          },
          {
            "type": "simple",
            "condition": {
              "fieldName": "type",
              "operator": "eq",
              "fieldValue": "expense"
            }
          }
        ]
      },
      "nodes": [
        {
          "nodeCode": "finance_manager_approval",
          "nodeType": "APPROVAL"
        }
      ]
    },
    {
      "name": "普通审批",
      "priority": 2,
      "condition": {
        "fieldName": "amount",
        "operator": "lte",
        "fieldValue": "10000"
      },
      "nodes": [
        {
          "nodeCode": "direct_manager_approval",
          "nodeType": "APPROVAL"
        }
      ]
    },
    {
      "name": "默认分支",
      "priority": 999,
      "nodes": [
        {
          "nodeCode": "default_approval",
          "nodeType": "APPROVAL"
        }
      ]
    }
  ]
}

支持的操作符：
 比较操作符：eq, ne, gt, gte, lt, lte
 字符串操作符：contains, starts_with, ends_with, regex
 列表操作符：in, not_in
 空值操作符：is_empty, is_not_empty
 范围操作符：between
 逻辑操作符：and, or, not（用于复杂表达式）


支持的数据类型：
 字符串：直接比较或转换后比较
 数值：自动转换为 float64 进行比较
 布尔值：支持 true/false 字符串转换
 日期：支持 "2006-01-02" 格式的日期比较
 数组：支持 JSON 数组或逗号分隔的字符串
*/

// ConditionExpression 复杂条件表达式结构
type ConditionExpression struct {
	Type      string                 `json:"type"`      // "simple", "and", "or"
	Condition *ConditionRule         `json:"condition"` // 简单条件
	Children  []*ConditionExpression `json:"children"`  // 子表达式
}

// evaluateComplexCondition 评估复杂条件表达式
func (s *approvalService) evaluateComplexCondition(expression *ConditionExpression, formData map[string]any) bool {
	if expression == nil {
		return true
	}

	switch expression.Type {
	case "simple":
		// 简单条件
		if expression.Condition != nil {
			return s.evaluateCondition(*expression.Condition, formData)
		}
		return true

	case "and":
		// AND 逻辑：所有子条件都必须为真
		if len(expression.Children) == 0 {
			return true
		}
		for _, child := range expression.Children {
			if !s.evaluateComplexCondition(child, formData) {
				return false
			}
		}
		return true

	case "or":
		// OR 逻辑：至少一个子条件为真
		if len(expression.Children) == 0 {
			return false
		}
		for _, child := range expression.Children {
			if s.evaluateComplexCondition(child, formData) {
				return true
			}
		}
		return false

	case "not":
		// NOT 逻辑：子条件的反值
		if len(expression.Children) == 1 {
			return !s.evaluateComplexCondition(expression.Children[0], formData)
		}
		return false

	default:
		s.logger.Warn("不支持的表达式类型", "type", expression.Type)
		return false
	}
}

// // evaluateCondition 评估单个条件
func (s *approvalService) evaluateCondition(condition ConditionRule, formData map[string]any) bool {
	// 如果字段名为空或操作符为 default，表示默认条件
	if condition.FieldName == "" || condition.Operator == "default" {
		return true
	}

	// 从表单数据中获取字段值
	fieldValue, exists := formData[condition.FieldName]
	if !exists {
		s.logger.Warn("字段不存在", "fieldName", condition.FieldName)
		return false
	}

	// 转换为字符串进行比较
	fieldValueStr := s.convertToStringForCondition(fieldValue)
	expectedValue := condition.FieldValue

	// 根据操作符进行比较
	switch condition.Operator {
	case "eq", "equal", "==":
		return s.evaluateEqualForCondition(fieldValueStr, expectedValue)
	case "ne", "not_equal", "!=":
		return !s.evaluateEqualForCondition(fieldValueStr, expectedValue)
	case "gt", "greater_than", ">":
		return s.evaluateGreaterThanForCondition(fieldValueStr, expectedValue)
	case "gte", "greater_than_equal", ">=":
		return s.evaluateGreaterThanForCondition(fieldValueStr, expectedValue) || s.evaluateEqualForCondition(fieldValueStr, expectedValue)
	case "lt", "less_than", "<":
		return s.evaluateLessThanForCondition(fieldValueStr, expectedValue)
	case "lte", "less_than_equal", "<=":
		return s.evaluateLessThanForCondition(fieldValueStr, expectedValue) || s.evaluateEqualForCondition(fieldValueStr, expectedValue)
	case "contains", "like":
		return strings.Contains(strings.ToLower(fieldValueStr), strings.ToLower(expectedValue))
	case "not_contains", "not_like":
		return !strings.Contains(strings.ToLower(fieldValueStr), strings.ToLower(expectedValue))
	case "starts_with":
		return strings.HasPrefix(strings.ToLower(fieldValueStr), strings.ToLower(expectedValue))
	case "ends_with":
		return strings.HasSuffix(strings.ToLower(fieldValueStr), strings.ToLower(expectedValue))
	case "in":
		return s.evaluateInForCondition(fieldValueStr, expectedValue)
	case "not_in":
		return !s.evaluateInForCondition(fieldValueStr, expectedValue)
	case "is_empty", "is_null":
		return s.evaluateIsEmptyForCondition(fieldValue)
	case "is_not_empty", "is_not_null":
		return !s.evaluateIsEmptyForCondition(fieldValue)
	case "regex", "regexp":
		return s.evaluateRegexForCondition(fieldValueStr, expectedValue)
	case "between":
		return s.evaluateBetweenForCondition(fieldValueStr, expectedValue)
	default:
		s.logger.Warn("不支持的条件操作符", "operator", condition.Operator)
		return false
	}
}

// convertToStringForCondition 将任意类型转换为字符串
func (s *approvalService) convertToStringForCondition(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// evaluateEqualForCondition 等于比较
func (s *approvalService) evaluateEqualForCondition(fieldValue, expectedValue string) bool {
	// 尝试数值比较
	if num1, err1 := strconv.ParseFloat(fieldValue, 64); err1 == nil {
		if num2, err2 := strconv.ParseFloat(expectedValue, 64); err2 == nil {
			return num1 == num2
		}
	}

	// 尝试布尔值比较
	if bool1, err1 := strconv.ParseBool(fieldValue); err1 == nil {
		if bool2, err2 := strconv.ParseBool(expectedValue); err2 == nil {
			return bool1 == bool2
		}
	}

	// 尝试日期比较
	if date1, err1 := time.Parse("2006-01-02", fieldValue); err1 == nil {
		if date2, err2 := time.Parse("2006-01-02", expectedValue); err2 == nil {
			return date1.Equal(date2)
		}
	}

	// 字符串比较（不区分大小写）
	return strings.EqualFold(fieldValue, expectedValue)
}

// evaluateGreaterThanForCondition 大于比较
func (s *approvalService) evaluateGreaterThanForCondition(fieldValue, expectedValue string) bool {
	// 尝试数值比较
	if num1, err1 := strconv.ParseFloat(fieldValue, 64); err1 == nil {
		if num2, err2 := strconv.ParseFloat(expectedValue, 64); err2 == nil {
			return num1 > num2
		}
	}

	// 尝试日期比较
	if date1, err1 := time.Parse("2006-01-02", fieldValue); err1 == nil {
		if date2, err2 := time.Parse("2006-01-02", expectedValue); err2 == nil {
			return date1.After(date2)
		}
	}

	// 字符串长度比较
	return len(fieldValue) > len(expectedValue)
}

// evaluateLessThanForCondition 小于比较
func (s *approvalService) evaluateLessThanForCondition(fieldValue, expectedValue string) bool {
	// 尝试数值比较
	if num1, err1 := strconv.ParseFloat(fieldValue, 64); err1 == nil {
		if num2, err2 := strconv.ParseFloat(expectedValue, 64); err2 == nil {
			return num1 < num2
		}
	}

	// 尝试日期比较
	if date1, err1 := time.Parse("2006-01-02", fieldValue); err1 == nil {
		if date2, err2 := time.Parse("2006-01-02", expectedValue); err2 == nil {
			return date1.Before(date2)
		}
	}

	// 字符串长度比较
	return len(fieldValue) < len(expectedValue)
}

// evaluateInForCondition 在列表中
func (s *approvalService) evaluateInForCondition(fieldValue, expectedValue string) bool {
	// 解析期望值为数组，支持逗号分隔或JSON数组格式
	var values []string

	// 尝试解析为JSON数组
	if strings.HasPrefix(expectedValue, "[") && strings.HasSuffix(expectedValue, "]") {
		if err := json.Unmarshal([]byte(expectedValue), &values); err == nil {
			for _, v := range values {
				if s.evaluateEqualForCondition(fieldValue, v) {
					return true
				}
			}
			return false
		}
	}

	// 按逗号分隔
	values = strings.Split(expectedValue, ",")
	for _, v := range values {
		v = strings.TrimSpace(v)
		if s.evaluateEqualForCondition(fieldValue, v) {
			return true
		}
	}
	return false
}

// evaluateIsEmptyForCondition 判断是否为空
func (s *approvalService) evaluateIsEmptyForCondition(fieldValue any) bool {
	if fieldValue == nil {
		return true
	}

	switch v := fieldValue.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		str := s.convertToStringForCondition(fieldValue)
		return strings.TrimSpace(str) == "" || str == "0" || str == "false"
	}
}

// evaluateRegexForCondition 正则表达式匹配
func (s *approvalService) evaluateRegexForCondition(fieldValue, pattern string) bool {
	matched, err := regexp.MatchString(pattern, fieldValue)
	if err != nil {
		s.logger.Error("正则表达式匹配失败",
			"pattern", pattern,
			"fieldValue", fieldValue,
			"error", err)
		return false
	}
	return matched
}

// evaluateBetweenForCondition 范围比较
func (s *approvalService) evaluateBetweenForCondition(fieldValue, rangeValue string) bool {
	// 解析范围值，格式："min,max" 或 "[min,max]"
	rangeValue = strings.Trim(rangeValue, "[]")
	parts := strings.Split(rangeValue, ",")
	if len(parts) != 2 {
		s.logger.Warn("范围值格式错误", "rangeValue", rangeValue)
		return false
	}

	minValue := strings.TrimSpace(parts[0])
	maxValue := strings.TrimSpace(parts[1])

	// 数值范围比较
	if num, err := strconv.ParseFloat(fieldValue, 64); err == nil {
		if minNum, err1 := strconv.ParseFloat(minValue, 64); err1 == nil {
			if maxNum, err2 := strconv.ParseFloat(maxValue, 64); err2 == nil {
				return num >= minNum && num <= maxNum
			}
		}
	}

	// 日期范围比较
	if date, err := time.Parse("2006-01-02", fieldValue); err == nil {
		if minDate, err1 := time.Parse("2006-01-02", minValue); err1 == nil {
			if maxDate, err2 := time.Parse("2006-01-02", maxValue); err2 == nil {
				return (date.Equal(minDate) || date.After(minDate)) &&
					(date.Equal(maxDate) || date.Before(maxDate))
			}
		}
	}

	return false
}

// ParseApproverConfig 解析审批人配置
func (s *approvalService) ParseApproverConfig(node *model.ApprovalNode, assigneeID, assigneeName *string) error {
	return s.parseApproverConfig(node, assigneeID, assigneeName)
}

// CreateDraftWithApproval 创建草稿并启动审批流程
func (s *approvalService) CreateDraftWithApproval(c *gin.Context, tableCode, reason string, entityMap map[string]any) error {
	// 生成操作类型和名称
	operation := "Create"
	entityMap["operation"] = "Create"
	tableCodeDraft := fmt.Sprintf("%s_draft", tableCode)
	operationInfo := make(map[string]string)
	if err := s.GetOperationInfo(operation, &operationInfo); err != nil {
		return err
	}

	// 生成entity struct iterface
	where := make(map[string]any)
	where["table_code"] = tableCodeDraft
	entity := s.tableFieldRepository.BuildEntity(tableCodeDraft)

	// 生成审批流 global id
	uuid := uuid.New()
	approvalCode := strings.ToUpper(uuid.String())

	// 保存到草稿箱，并提交审批流
	// 生成 draft global id
	draftGid := s.globalIdService.GetNewID("entity_draft")
	// 生成 entity global id
	gid := s.globalIdService.GetNewID("entity")

	// 合并entityMap到entity - 使用snake_case字段名
	for k, v := range entityMap {
		entity[k] = v
	}
	entity["id"] = draftGid
	entity["operation"] = operation
	entity["action"] = operationInfo["action"]
	entity["send_status"] = 0
	entity["entity_id"] = gid
	entity["approval_code"] = approvalCode
	entity["draft_status"] = "Drafted"
	entity["status"] = operationInfo["status"]
	entity["created_by"] = c.GetString("user_name")
	entity["updated_by"] = c.GetString("user_name")

	// Save entity to draft table
	if err := s.entityRepository.Create(c, tableCodeDraft, entity); err != nil {
		return err
	}

	// make infomation of approval
	approvalInfo := make(map[string]string)
	approvalInfo["operation"] = operation
	approvalInfo["approvalCode"] = approvalCode
	approvalInfo["operationName"] = operationInfo["operationName"]
	approvalInfo["action"] = operationInfo["action"]
	approvalInfo["reason"] = reason
	approvalInfo["entityCode"] = tableCode

	// 创建审批流程
	if err := s.CreateApprovalFlow(c, tableCode, approvalInfo, entityMap); err != nil {
		return err
	}

	// 审批流程创建成功后,更新draft_status为Pending
	updateMap := map[string]any{
		"draft_status": "Pending",
	}
	whereMap := map[string]any{
		"approval_code": approvalCode,
	}
	if err := s.entityRepository.Update(c, tableCodeDraft, updateMap, whereMap); err != nil {
		s.logger.Error("更新draft_status失败", "err", err, "approvalCode", approvalCode)
		// 不阻断流程,仅记录错误
	}

	return nil
}

// UpdateDraftWithApproval 更新草稿并启动审批流程
func (s *approvalService) UpdateDraftWithApproval(c *gin.Context, tableCode, reason string, entityMap map[string]any) error {
	operation := "Update"
	entityMap["operation"] = "Update"
	operationInfo := make(map[string]string)
	if err := s.GetOperationInfo(operation, &operationInfo); err != nil {
		return err
	}

	tableCodeDraft := fmt.Sprintf("%s_draft", tableCode)

	// Check if there is an active draft for this entity
	if err := s.CheckExistingActiveDraft(c, tableCodeDraft, entityMap["id"]); err != nil {
		return err
	}

	// 生成entity struce iterface
	where := make(map[string]any)
	where["table_code"] = tableCodeDraft
	entity := s.tableFieldRepository.BuildEntity(tableCodeDraft)

	// 生成审批流 global id
	uuid := uuid.New()
	approvalCode := strings.ToUpper(uuid.String())

	// 保存到草稿箱，并提交审批流
	// 生成 draft global id
	draftGid := s.globalIdService.GetNewID("entity_draft")

	// 合并entityMap到entity - 使用snake_case字段名
	for k, v := range entityMap {
		entity[k] = v
	}
	entity["operation"] = operation
	entity["action"] = operationInfo["action"]
	entity["send_status"] = 0
	entity["entity_id"] = entityMap["id"]
	entity["id"] = draftGid
	entity["approval_code"] = approvalCode
	entity["draft_status"] = "Pending"
	entity["status"] = operationInfo["status"]
	entity["created_by"] = c.GetString("user_name")

	// 生成自动编码 (实际上是用于保留原值)
	// 如果是 update 操作, generateAutocodes 会查询原数据并填充到 entity 中
	if err := s.generateAutocodes(c, tableCode, entity); err != nil {
		s.logger.Warn("Failed to generate/restore autocodes in UpdateDraftWithApproval", "err", err)
		// 不阻断
	}

	// Save change information to draft table
	if err := s.entityRepository.Create(c, tableCodeDraft, entity); err != nil {
		return err
	}

	// Make Change Request information to approval
	approvalInfo := make(map[string]string)
	approvalInfo["operation"] = operation
	approvalInfo["approvalCode"] = approvalCode
	approvalInfo["operationName"] = operationInfo["operationName"]
	approvalInfo["action"] = operationInfo["action"]
	approvalInfo["reason"] = reason
	approvalInfo["entityCode"] = tableCode

	// Create workflow
	return s.CreateApprovalFlow(c, tableCode, approvalInfo, entityMap)
}

// UpdateByIdsWithApproval 批量更新并启动审批流程
func (s *approvalService) UpdateByIdsWithApproval(c *gin.Context, tableCode, reason string, ids []uint, entityMap map[string]any) error {
	s.logger.Info("service UpdateByIdsWithApproval", "tableCode", tableCode, "reason", reason, "ids", ids, "entityMap", entityMap)
	operation := entityMap["operation"].(string)
	s.logger.Info("service UpdateByIdsWithApproval", "operation", operation)
	operationInfo := map[string]string{}
	if err := s.GetOperationInfo(operation, &operationInfo); err != nil {
		return err
	}
	s.logger.Info("service UpdateByIdsWithApproval2223", "operationInfo", operationInfo)

	// Map operation string to code
	// 根据 GetOperationInfo 的映射关系设置 action 代码
	switch operation {
	case "Freeze":
		entityMap["action"] = "B"
	case "Unfreeze":
		entityMap["action"] = "C"
	case "Delete":
		entityMap["action"] = "D"
	case "Lock":
		entityMap["action"] = "U"
	case "Unlock":
		entityMap["action"] = "U"
	default:
		// For other operations, use the action code from operationInfo if available, or default to "U"
		if action, ok := operationInfo["action"]; ok {
			entityMap["action"] = action
		} else {
			entityMap["action"] = "U"
		}
	}
	s.logger.Info("service UpdateByIdsWithApproval444", "entityMap", entityMap)
	s.logger.Info("service UpdateByIdsWithApproval445", "ids", ids)

	// 获取 origin data by ids
	where := map[string]any{
		"id in": ids,
	}
	s.logger.Info("service UpdateByIdsWithApproval446", "where", where)

	origins, err := s.entityRepository.Find(tableCode, "*", where)
	if err != nil {
		s.logger.Error("GetAll", "err", err)
	}
	s.logger.Info("service UpdateByIdsWithApproval447", "origins", origins)

	tableCodeDraft := fmt.Sprintf("%s_draft", tableCode)

	// Check if there is an active draft for these entities
	for _, id := range ids {
		if err := s.CheckExistingActiveDraft(c, tableCodeDraft, id); err != nil {
			return err
		}
	}
	uuid := uuid.New()
	approvalCode := strings.ToUpper(uuid.String())

	for _, row := range origins {
		// 生成entity struce iterface
		entity := s.tableFieldRepository.BuildEntity(tableCodeDraft)
		// entityMap := make(map[string]any)
		currentRow := row

		// 合并entityMap到entity - 使用snake_case字段名
		for k, v := range currentRow {
			entity[k] = v
		}
		entity["operation"] = operation
		entity["action"] = operationInfo["action"]
		entity["send_status"] = 0
		entity["entity_id"] = currentRow["id"]
		entity["id"] = nil
		entity["approval_code"] = approvalCode
		entity["draft_status"] = "Pending"
		entity["status"] = operationInfo["status"]
		entity["created_by"] = c.GetString("user_name")

		s.logger.Debug("entity after merge", "entity", entity)

		if err := s.entityRepository.Create(c, tableCodeDraft, entity); err != nil {
			return err
		}
	}

	approvalInfo := map[string]string{}
	approvalInfo["operation"] = operation
	approvalInfo["approvalCode"] = approvalCode
	approvalInfo["operationName"] = operationInfo["operationName"]
	approvalInfo["action"] = operationInfo["action"]
	approvalInfo["reason"] = reason
	approvalInfo["entityCode"] = tableCode

	s.logger.Debug("service UpdateByIdsWithApproval", "approvalInfo", approvalInfo)
	// return s.UpdateApprovalFlow(c, map[string]string{"tableCode": tableCode}, approvalInfo)
	// return s.UpdateApprovalFlow(c, tableCode, approvalInfo)
	// 将 ids 转换为 []any 并放入 entityMap，供 createFeishuInstanceFromInfo 使用
	idsAny := make([]any, len(ids))
	for i, v := range ids {
		idsAny[i] = v
	}
	entityMap["ids"] = idsAny

	return s.CreateApprovalFlow(c, tableCode, approvalInfo, entityMap)
}

// func (s *approvalService) UpdateApprovalFlow(c *gin.Context, tableCode, approvalInfo map[string]string) error {
// 	return s.UpdateApprovalFlow(c, tableCode, approvalInfo)
// }

// ImportWithApproval 导入数据并启动审批流程
func (s *approvalService) ImportWithApproval(c *gin.Context, tableCode, reason, operation string, r io.Reader) error {
	s.logger.Debug("s ImportWithApproval0: ",
		"tableCode", tableCode,
		"reason", reason,
		"operation", operation)

	operationInfo := map[string]string{}
	if err := s.GetOperationInfo(operation, &operationInfo); err != nil {
		return err
	}
	s.logger.Debug("h ImportWithApproval: ",
		"tableCode", tableCode,
		"reason", reason,
		"operation", operation)

	// TODO 如果是 draft，需要转换成主表名，然后取表定义
	// fieldTable := tableCode
	// if strings.HasSuffix(tableCode, "_draft") {
	// 	fieldTable = tableCode[0 : len(tableCode)-6]
	// }

	xlsx, err := excelize.OpenReader(r)
	if err != nil {
		return err
	}

	rows := xlsx.GetRows("Sheet1")
	// generate uuid for approval instance
	uuid := uuid.New()
	approvalCode := strings.ToUpper(uuid.String())

	tableCodeDraft := fmt.Sprintf("%s_draft", tableCode)
	s.logger.Debug("s ImportWithApproval1: ", "tableCodeDraft", tableCodeDraft)

	// Fetch field definitions for type mapping
	fieldWhere := map[string]any{"table_code": tableCode} // Use existing tableCode
	// Query both FieldType (UI) and Type (DB)
	tFields, err := s.tableFieldService.Find("code,field_type,type", fieldWhere)
	if err != nil {
		s.logger.Error("ImportWithApproval: fetch fields failed", "err", err)
		return err
	}

	// Map field code to both types
	fieldTypeMap := make(map[string]struct {
		uiType string
		dbType string
	})

	for _, f := range tFields {
		fieldTypeMap[f.Code] = struct{ uiType, dbType string }{
			uiType: f.FieldType,
			dbType: f.Type,
		}
	}

	s.logger.Debug("ImportWithApproval - FieldTypeMap", "count", len(fieldTypeMap), "sample", fmt.Sprintf("%v", tFields[0]))

	var fields []string
	var importedRecords []any
	// 保存到草稿箱，并提交审批流
	for irow, row := range rows {
		// 生成entity struce iterface
		entity := s.tableFieldRepository.BuildEntity(tableCodeDraft)
		entityMap := make(map[string]any)
		if irow == 0 {
			fields = append(fields, row...)
		} else {
			for index, cell := range row {
				if index >= len(fields) {
					break
				}
				fieldCode := fields[index]

				// 解决编码从文本转到unit
				if fieldCode == "ID" || fieldCode == "id" {
					if num, err := strconv.Atoi(cell); err == nil {
						entityMap[fieldCode] = num
						continue
					}
				}

				// Handle empty strings based on field type
				if cell == "" {
					if types, ok := fieldTypeMap[fieldCode]; ok {
						uiType := strings.ToLower(types.uiType)
						dbType := strings.ToLower(types.dbType)

						// Check both UI type and DB type for indicators of numeric/date values
						isDateOrNum := false
						if strings.Contains(uiType, "date") ||
							strings.Contains(uiType, "time") ||
							strings.Contains(uiType, "int") ||
							strings.Contains(uiType, "number") ||
							strings.Contains(uiType, "decimal") ||
							strings.Contains(uiType, "float") ||
							strings.Contains(uiType, "double") {
							isDateOrNum = true
						}

						if !isDateOrNum {
							if strings.Contains(dbType, "date") ||
								strings.Contains(dbType, "time") ||
								strings.Contains(dbType, "int") ||
								strings.Contains(dbType, "decimal") ||
								strings.Contains(dbType, "float") ||
								strings.Contains(dbType, "double") {
								isDateOrNum = true
							}
						}

						if isDateOrNum {
							entityMap[fieldCode] = nil
							continue
						}
					} else {
						// Only log if it's not a known system/ignored field or likely header mismatch
						// But honestly, fields in Excel not in DB is normal if template is old or changed.
						// s.logger.Warn("ImportWithApproval: field not found in definitions", "field", fieldCode, "table", tableCode)
					}
				}

				entityMap[fieldCode] = cell
			}

			// If this is an update operation, check for existing active draft
			if operation != "BatchCreate" {
				if idVal, ok := entityMap["id"]; ok {
					if err := s.CheckExistingActiveDraft(c, tableCodeDraft, idVal); err != nil {
						return err
					}
				}
			}

			// 生成 draft global id
			draftGid := s.globalIdService.GetNewID("entity_draft")
			// 生成 entity global id
			gid := s.globalIdService.GetNewID("entity")
			// 合并entityMap到entity - 使用snake_case字段名
			for k, v := range entityMap {
				entity[k] = v
			}
			entity["operation"] = operation
			entity["action"] = operationInfo["action"]
			entity["send_status"] = 0
			// new global id for new entity
			if operation == "BatchCreate" {
				entity["entity_id"] = gid
			} else {
				entity["entity_id"] = entityMap["id"]
			}
			entity["id"] = draftGid
			entity["approval_code"] = approvalCode
			entity["draft_status"] = "Pending"
			entity["status"] = operationInfo["status"]
			entity["created_by"] = "jasen"
			s.logger.Debug("s ImportWithApproval2: ", "entity", entity)

			// 生成自动编码字段 (使用 Draft 表名)
			// Fix: 使用主表名查找配置，但写入 entity (将会存入 draft 表)
			// 这里的 entityMap 是 draft 的数据，但字段名应该和主表一致
			// 注意: 我们传入 tableCode (主表) 给 generateAutocodes 来查找配置
			if err := s.generateAutocodes(c, tableCode, entity); err != nil {
				return err
			}

			if err := s.entityRepository.Create(c, tableCodeDraft, entity); err != nil {
				return err
			}

			// Add to imported records for approval summary
			// Fix: Use entity instead of entityMap to ensure we have generated/restored fields
			importedRecords = append(importedRecords, entity)
		}
	}

	// 保存审批实例
	approvalInfo := make(map[string]string)
	approvalInfo["operation"] = operation
	approvalInfo["approvalCode"] = approvalCode
	approvalInfo["operationName"] = operationInfo["operationName"]
	approvalInfo["action"] = operationInfo["action"]
	approvalInfo["reason"] = reason
	approvalInfo["entityCode"] = tableCode

	// Pass records for Feishu summary logic
	formData := map[string]any{
		"records": importedRecords,
	}
	return s.CreateApprovalFlow(c, tableCode, approvalInfo, formData)
}

// CheckExistingActiveDraft 检查是否存在处于审批中的草稿
func (s *approvalService) CheckExistingActiveDraft(c *gin.Context, tableCodeDraft string, entityID any) error {
	where := map[string]any{
		"entity_id":    entityID,
		"draft_status": "Pending",
	}

	// Check if records exist - query approval_code too
	existingDrafts, err := s.entityRepository.Find(tableCodeDraft, "id,approval_code", where)
	if err != nil {
		s.logger.Error("CheckExistingActiveDraft query error", "err", err)
		return fmt.Errorf("check existing draft failed: %v", err)
	}

	for _, draft := range existingDrafts {
		// Double check with approval status
		if approvalCode, ok := draft["approval_code"].(string); ok && approvalCode != "" {
			approval, err := s.approvalRepository.FirstByCode(approvalCode)
			if err == nil {
				// If approval is NOT pending, then this draft is stale.
				// We should ideally update it, but for now just don't block.
				if !approval.IsPending() {
					s.logger.Warn("Found stale draft with Pending status but Approval is finished",
						"draftID", draft["id"],
						"approvalCode", approvalCode,
						"approvalStatus", approval.Status)

					// Auto hash-correct the stale status
					draftStatus := "Drafted"
					switch approval.Status {
					case model.ApprovalStatusRejected:
						draftStatus = "Rejected"
					case model.ApprovalStatusApproved:
						draftStatus = "Approved"
					}

					// Run update in background/async or just do it here
					updateMap := map[string]any{"draft_status": draftStatus}
					whereUpdate := map[string]any{"id": draft["id"]}
					_ = s.entityRepository.Update(c, tableCodeDraft, updateMap, whereUpdate)

					continue
				}
			}
		}

		// If we get here, we found a truly pending draft (or couldn't verify approval), so we block
		return fmt.Errorf("current data is in approval process, cannot be submitted again")
	}

	return nil
}

// ==================== 流程审批引擎方法实现 (从 approval.go 迁移) ====================

// 业务查询方法
func (s *approvalService) GetByApplicant(applicantID string) ([]*model.Approval, error) {
	return s.approvalRepository.FindByApplicantID(applicantID)
}

func (s *approvalService) GetByStatus(status string) ([]*model.Approval, error) {
	return s.approvalRepository.FindByStatus(status)
}

func (s *approvalService) GetPendingByApplicant(applicantID string) ([]*model.Approval, error) {
	return s.approvalRepository.FindPendingByApplicant(applicantID)
}

func (s *approvalService) GetByEntityCode(entityCode string) ([]*model.Approval, error) {
	return s.approvalRepository.FindByEntityCode(entityCode)
}

func (s *approvalService) GetByEntityID(entityCode, entityID string) ([]*model.Approval, error) {
	return s.approvalRepository.FindByEntityID(entityCode, entityID)
}

func (s *approvalService) GetApprovalHistory(id uint) ([]*model.Approval, error) {
	return s.approvalRepository.FindHistory(id)
}

// FindPendingByAssignee 查询待指定用户审批的实例(分页)
func (s *approvalService) FindPendingByAssignee(assigneeName string, page, pageSize int, total *int64, timeRange string) ([]*model.Approval, error) {
	return s.approvalRepository.FindPendingByAssignee(assigneeName, page, pageSize, total, timeRange)
}

// FindProcessedByAssignee 查询指定用户已审批过的实例(分页)
func (s *approvalService) FindProcessedByAssignee(assigneeName string, page, pageSize int, total *int64, timeRange string) ([]*model.Approval, error) {
	return s.approvalRepository.FindProcessedByAssignee(assigneeName, page, pageSize, total, timeRange)
}

// 审批流程引擎核心方法
func (s *approvalService) StartApprovalFlow(c *gin.Context, approvalDefCode, applicantID, title, formData string) (*model.Approval, error) {
	// 获取审批定义
	approvalDef, err := s.approvalDefinitionRepository.FirstByCode(approvalDefCode)
	if err != nil {
		return nil, fmt.Errorf("审批定义不存在: %v", err)
	}

	if !approvalDef.IsActive() {
		return nil, errors.New("审批定义未激活")
	}

	// 生成审批实例
	uuid := uuid.New()

	// 强制统一格式：审批定义名称-新建 (忽略入参 title)
	title = fmt.Sprintf("%s-新建", approvalDef.Name)

	approval := &model.Approval{
		Code:            strings.ToUpper(uuid.String()),
		Title:           title,
		ApprovalDefCode: approvalDefCode,
		SerialNumber:    s.generateSerialNumberFlow(),
		FormData:        formData,
		Status:          model.ApprovalStatusPending,
		CreatedBy:       applicantID,
		UpdatedBy:       applicantID,
		Priority:        0,
		Urgency:         "Normal",
	}

	if err := s.approvalRepository.Create(c, approval); err != nil {
		return nil, fmt.Errorf("创建审批实例失败: %v", err)
	}

	return approval, nil
}

func (s *approvalService) SubmitApprovalFlow(c *gin.Context, approvalCode string) error {
	approval, err := s.approvalRepository.FirstByCode(approvalCode)
	if err != nil {
		return err
	}

	if approval.Status != model.ApprovalStatusPending {
		return errors.New("只有待审批状态的申请才能提交")
	}

	// 更新提交时间和开始时间
	// now := time.Now()
	//  approval.SubmittedAt = &now
	// approval.StartedAt = &now

	// 移动到第一个审批节点
	if err := s.MoveToNextNodeFlow(c, approvalCode, approval.CurrentTaskID); err != nil {
		return fmt.Errorf("移动到下一节点失败: %v", err)
	}

	return nil
}

func (s *approvalService) CancelApprovalFlow(c *gin.Context, approvalCode, reason string) error {
	canCancel, err := s.CanCancelFlow(c, approvalCode)
	if err != nil {
		return err
	}
	if !canCancel {
		return errors.New("当前状态不允许撤回")
	}

	approval, err := s.approvalRepository.FirstByCode(approvalCode)
	if err != nil {
		return fmt.Errorf("审批不存在: %v", err)
	}

	// 更新状态
	if err := s.approvalRepository.UpdateStatusByCode(c, approvalCode, model.ApprovalStatusCanceled); err != nil {
		return err
	}

	// 取消所有待处理的任务
	tasks, err := s.approvalTaskRepository.FindByApprovalCode(approvalCode)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.IsPending() {
			if err := s.approvalTaskRepository.UpdateStatus(task.ID, model.TaskStatusCanceled); err != nil {
				s.logger.Error("取消任务失败", "error", err)
			}
		}
	}

	// 更新draft状态
	tableCodeDraft := fmt.Sprintf("%s_draft", approval.EntityCode)
	updateMap := map[string]any{
		"draft_status": "Drafted",
	}
	whereMap := map[string]any{
		"approval_code": approval.Code,
	}
	if err := s.entityRepository.Update(c, tableCodeDraft, updateMap, whereMap); err != nil {
		s.logger.Error("CancelApprovalFlow: failed to revert draft status", "err", err, "approvalCode", approval.Code)
	}

	return nil
}

func (s *approvalService) ProcessApprovalFlow(c *gin.Context, taskID uint, assigneeID, comment, reason string) error {
	// 根据任务ID获取任务
	task, err := s.approvalTaskRepository.FindOne(taskID)
	if err != nil {
		return fmt.Errorf("任务不存在: %v", err)
	}

	// 验证任务状态
	if !task.IsPending() {
		return errors.New("任务已处理，无法重复操作")
	}

	// 验证处理人权限
	if task.AssigneeID != assigneeID {
		return errors.New("无权限处理此任务")
	}

	// 获取审批实例
	approval, err := s.approvalRepository.FirstByCode(task.ApprovalCode)
	if err != nil {
		return fmt.Errorf("审批实例不存在: %v", err)
	}

	// 验证审批状态
	if !approval.IsPending() {
		return errors.New("审批已完成，无法处理")
	}

	// 这里需要根据具体的操作类型来处理，由于handler中没有传递action，
	// 我们需要从handler中获取操作类型
	return errors.New("ProcessApprovalFlow方法需要重构以支持不同的操作类型")
}

// 任务处理方法
func (s *approvalService) ApproveTaskFlow(c *gin.Context, taskID uint, assigneeID, comment, reason string) error {
	// 根据任务ID获取任务
	task, err := s.approvalTaskRepository.FindOne(taskID)
	if err != nil {
		return fmt.Errorf("任务不存在: %v", err)
	}

	// 验证任务状态和权限
	if !task.IsPending() {
		return errors.New("任务已处理，无法重复操作")
	}
	if task.AssigneeID != assigneeID {
		return errors.New("无权限处理此任务")
	}

	// 获取审批实例
	approval, err := s.approvalRepository.FirstByCode(task.ApprovalCode)
	if err != nil {
		return fmt.Errorf("审批实例不存在: %v", err)
	}

	// 完成任务
	if err := s.approvalTaskRepository.CompleteTask(uint(taskID), model.TaskStatusApproved, comment, reason); err != nil {
		return err
	}
	// 任务状态更新成功
	// 检查节点是否完成
	nodeCompleted, err := s.isNodeCompletedFlow(approval.Code, task.NodeCode)
	if err != nil {
		s.logger.Error("检查节点完成状态失败", "error", err)
		return err
	}

	if nodeCompleted {
		// 节点已完成，开始移动到下一节点
		return s.MoveToNextNodeFlow(c, approval.Code, task.NodeCode)
	}

	// 节点未完成，等待其他任务完成
	return nil
}

func (s *approvalService) RejectTaskFlow(c *gin.Context, taskID uint, assigneeID, comment, reason string) error {
	// 根据任务ID获取任务
	task, err := s.approvalTaskRepository.FindOne(taskID)
	if err != nil {
		return fmt.Errorf("任务不存在: %v", err)
	}

	// 验证任务状态和权限
	if !task.IsPending() {
		return errors.New("任务已处理，无法重复操作")
	}
	if task.AssigneeID != assigneeID {
		return errors.New("无权限处理此任务")
	}

	// 完成任务
	if err := s.approvalTaskRepository.CompleteTask(taskID, model.TaskStatusRejected, comment, reason); err != nil {
		return err
	}

	// 拒绝整个审批
	return s.RejectApprovalFlow(c, task.ApprovalCode, reason)
}

// 任务管理方法
func (s *approvalService) TransferTask(c *gin.Context, taskID uint, fromUserID, toUserID, fromUserName, toUserName, reason string) error {
	// 根据任务ID获取任务
	task, err := s.approvalTaskRepository.FindOne(taskID)
	if err != nil {
		return fmt.Errorf("任务不存在: %v", err)
	}

	// 验证任务状态和权限
	if !task.IsPending() {
		return errors.New("任务已处理，无法转交")
	}
	if task.AssigneeID != fromUserID {
		return errors.New("无权限转交此任务")
	}

	// 执行转交
	if err := s.approvalTaskRepository.TransferTask(uint(taskID), fromUserID, toUserID, fromUserName, toUserName, reason); err != nil {
		return err
	}

	return nil
}

func (s *approvalService) RemindTaskFlow(c *gin.Context, taskID uint) error {
	// 根据任务ID获取任务
	task, err := s.approvalTaskRepository.FindOne(taskID)
	if err != nil {
		return fmt.Errorf("任务不存在: %v", err)
	}

	// 验证任务状态
	if !task.IsPending() {
		return errors.New("任务已处理，无需催办")
	}

	// 更新催办次数和时间
	if err := s.approvalTaskRepository.UpdateRemindInfo(taskID); err != nil {
		return err
	}

	return nil
}

func (s *approvalService) BatchRemindTasksFlow(c *gin.Context, assigneeID string) error {
	// 获取用户的所有待处理任务
	tasks, err := s.approvalTaskRepository.FindPendingByAssignee(assigneeID)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		return errors.New("没有待处理的任务")
	}

	// 批量更新催办信息
	var taskIDs []uint
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.ID)
	}

	if err := s.approvalTaskRepository.BatchUpdateRemindInfo(taskIDs); err != nil {
		return err
	}

	return nil
}

// 流程控制
func (s *approvalService) MoveToNextNodeFlow(c *gin.Context, approvalCode, currentNodeCode string) error {

	approval, err := s.approvalRepository.FirstByCode(approvalCode)
	if err != nil {
		s.logger.Error("获取审批实例失败", "error", err)
		return err
	}

	// 获取下一个节点
	nextNodes, err := s.approvalNodeRepository.FindNextNodes(approval.ApprovalDefCode, currentNodeCode)
	if err != nil {
		s.logger.Error("查找下一个节点失败", "error", err)
		return err
	}

	if len(nextNodes) == 0 {
		// 没有下一个节点，完成审批
		return s.CompleteApprovalFlow(c, approvalCode, model.ApprovalStatusApproved, "审批流程完成")
	}

	// 简化处理：取第一个下一节点
	nextNode := nextNodes[0]

	// 更新当前节点
	if err := s.approvalRepository.UpdateCurrentNode(c, approval.ID, nextNode.NodeCode, nextNode.NodeName); err != nil {
		s.logger.Error("更新当前节点失败", "error", err)
		return err
	}

	// 如果是结束节点，完成审批
	if nextNode.IsEndNode() {
		return s.CompleteApprovalFlow(c, approvalCode, model.ApprovalStatusApproved, "审批流程完成")
	}

	// 如果是审批节点，创建审批任务
	if nextNode.IsApprovalNode() {
		return s.createApprovalTasksFlow(approval, nextNode)
	}

	return nil
}

func (s *approvalService) CompleteApprovalFlow(c *gin.Context, approvalCode, result, reason string) error {
	approval, err := s.approvalRepository.FirstByCode(approvalCode)
	if err != nil {
		return err
	}

	// 更新审批实例状态
	// now := time.Now()
	approval.Status = result
	// approval.Result = result
	// approval.ResultReason = reason
	// approval.CompletedAt = &now

	if err := s.approvalRepository.Update(c, approval); err != nil {
		return err
	}

	// 同步更新关联实体的 draft_status
	if approval.EntityCode != "" {
		tableCodeDraft := fmt.Sprintf("%s_draft", approval.EntityCode)
		draftStatus := "Pending"
		switch result {
		case model.ApprovalStatusApproved:
			draftStatus = "Approved"
		case model.ApprovalStatusRejected:
			draftStatus = "Rejected"
		case model.ApprovalStatusCanceled:
			draftStatus = "Drafted"
		}

		updateMap := map[string]any{
			"draft_status": draftStatus,
		}
		whereMap := map[string]any{
			"approval_code": approvalCode,
		}
		if err := s.entityRepository.Update(c, tableCodeDraft, updateMap, whereMap); err != nil {
			s.logger.Error("CompleteApprovalFlow: failed to sync draft status",
				"err", err,
				"approvalCode", approvalCode,
				"draftStatus", draftStatus)
			// 不阻断流程,仅记录错误
		}
	}

	return nil
}

func (s *approvalService) RejectApprovalFlow(c *gin.Context, approvalCode, reason string) error {
	return s.CompleteApprovalFlow(c, approvalCode, model.ApprovalStatusRejected, reason)
}

// 状态管理
func (s *approvalService) UpdateApprovalStatusFlow(c *gin.Context, id uint, status string) error {
	return s.approvalRepository.UpdateStatus(c, id, status)
}

func (s *approvalService) UpdateCurrentNodeFlow(c *gin.Context, id uint, nodeID, nodeName string) error {
	return s.approvalRepository.UpdateCurrentNode(c, id, nodeID, nodeName)
}

// 验证方法
func (s *approvalService) ValidateApprovalFlow(c *gin.Context, approval *model.Approval) error {
	if approval.Title == "" {
		return errors.New("审批标题不能为空")
	}
	if approval.ApprovalDefCode == "" {
		return errors.New("审批定义编码不能为空")
	}
	return nil
}

func (s *approvalService) CanCancelFlow(c *gin.Context, approvalCode string) (bool, error) {
	approval, err := s.approvalRepository.FirstByCode(approvalCode)
	if err != nil {
		return false, err
	}
	return approval.CanCancel(), nil
}

func (s *approvalService) CanProcessFlow(c *gin.Context, approvalCode, assigneeID string) (bool, error) {
	// 查找用户的待处理任务
	tasks, err := s.approvalTaskRepository.FindPendingByAssignee(assigneeID)
	if err != nil {
		return false, err
	}

	for _, task := range tasks {
		if task.ApprovalCode == approvalCode {
			return true, nil
		}
	}
	return false, nil
}

// 统计方法
func (s *approvalService) GetApprovalStatisticsFlow(applicantID string) (map[string]int64, error) {
	stats := make(map[string]int64)

	// 统计各状态的审批数量
	statuses := []string{
		model.ApprovalStatusPending,
		model.ApprovalStatusApproved,
		model.ApprovalStatusRejected,
		model.ApprovalStatusCanceled,
		model.ApprovalStatusDeleted,
		model.ApprovalStatusExpired,
	}

	for _, status := range statuses {
		count, err := s.approvalRepository.CountByStatus(status)
		if err != nil {
			return nil, err
		}
		stats[status] = count
	}

	return stats, nil
}

func (s *approvalService) GetExpiredApprovalsFlow() ([]*model.Approval, error) {
	return s.approvalRepository.FindExpiredApprovals()
}

func (s *approvalService) ProcessApprovalByCodeFlow(c *gin.Context, approvalCode, nodeCode, assigneeID, action, comment, reason string) error {
	// 验证是否可以处理
	canProcess, err := s.CanProcessFlow(c, approvalCode, assigneeID)
	if err != nil {
		return err
	}
	if !canProcess {
		return errors.New("无权限处理此审批")
	}

	// 获取审批实例
	approval, err := s.approvalRepository.FirstByCode(approvalCode)
	if err != nil {
		return fmt.Errorf("审批实例不存在: %v", err)
	}

	// 验证审批状态
	if !approval.IsPending() {
		return errors.New("审批已完成，无法处理")
	}

	// 查找当前任务
	task, err := s.findCurrentTaskByCodeFlow(approvalCode, nodeCode, assigneeID)
	if err != nil {
		return err
	}

	// 根据操作类型处理
	switch action {
	case model.OperationApprove:
		return s.approveTaskByCodeFlow(c, approval, task, comment, reason)
	case model.OperationReject:
		return s.rejectTaskByCodeFlow(c, approval, task, comment, reason)
	case model.OperationTransfer:
		// TODO: 实现转交逻辑
		return errors.New("转交功能暂未实现")
	default:
		return errors.New("无效的操作类型")
	}
}

// ==================== 私有辅助方法 ====================

// 生成审批单编号
func (s *approvalService) generateSerialNumberFlow() string {
	now := time.Now()
	return fmt.Sprintf("AP%s%06d", now.Format("20060102"), now.Unix()%1000000)
}

// 查找当前任务
func (s *approvalService) findCurrentTaskFlow(approvalCode, assigneeID string) (*model.ApprovalTask, error) {
	where := map[string]any{
		"approval_code": approvalCode,
		"assignee_id":   assigneeID,
		"status":        model.TaskStatusPending,
	}

	tasks, err := s.approvalTaskRepository.Find(where)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, errors.New("未找到待处理任务")
	}

	return tasks[0], nil
}

// 检查节点是否完成
func (s *approvalService) isNodeCompletedFlow(approvalCode, nodeCode string) (bool, error) {

	tasks, err := s.approvalTaskRepository.FindByApprovalCode(approvalCode)
	if err != nil {
		s.logger.Error("查找审批任务失败", "error", err)
		return false, err
	}

	// 查找到的审批任务
	var pendingTasks []string
	for _, task := range tasks {
		// 检查任务状态
		if task.NodeCode == nodeCode && task.IsPending() {
			pendingTasks = append(pendingTasks, fmt.Sprintf("Task-%d", task.ID))
		}
	}

	isCompleted := len(pendingTasks) == 0
	// 节点完成状态检查结果
	return isCompleted, nil
}

// isNodeCompleted 判断节点是否完成(支持OR/AND模式)
func (s *approvalService) isNodeCompleted(node *model.ApprovalNode, approvalCode string) (bool, error) {
	// 解析审批人配置
	var approverConfig struct {
		Mode string `json:"mode"` // OR, AND
	}

	if node.ApproverConfig != "" {
		if err := json.Unmarshal([]byte(node.ApproverConfig), &approverConfig); err != nil {
			s.logger.Error("解析审批人配置失败", "error", err)
			// 默认使用 OR 模式
			approverConfig.Mode = "OR"
		}
	}

	// 如果没有指定模式，默认为 OR
	if approverConfig.Mode == "" {
		approverConfig.Mode = "OR"
	}

	// 获取当前节点的所有任务
	tasks, err := s.approvalTaskRepository.FindByApprovalCode(approvalCode)
	if err != nil {
		s.logger.Error("查找审批任务失败", "error", err)
		return false, err
	}

	// 过滤出当前节点的任务
	var nodeTasks []*model.ApprovalTask
	for _, task := range tasks {
		if task.NodeCode == node.NodeCode {
			nodeTasks = append(nodeTasks, task)
		}
	}

	if len(nodeTasks) == 0 {
		s.logger.Warn("节点没有任务",
			"nodeCode", node.NodeCode)
		return false, nil
	}

	// 记录节点任务状态
	taskStatuses := make([]string, 0, len(nodeTasks))
	for _, task := range nodeTasks {
		taskStatuses = append(taskStatuses, fmt.Sprintf("Task-%d(%s):%s", task.ID, task.AssigneeName, task.Status))
	}

	var isCompleted bool
	switch approverConfig.Mode {
	case "OR":
		// OR 模式:任意一个任务完成即可
		for _, task := range nodeTasks {
			if task.Status == model.TaskStatusApproved {
				isCompleted = true
				break
			}
		}

	case "AND":
		// AND 模式:所有任务都必须完成
		isCompleted = true
		for _, task := range nodeTasks {
			if task.Status != model.TaskStatusApproved {
				isCompleted = false
				break
			}
		}

	default:
		// 默认为 OR 模式
		for _, task := range nodeTasks {
			if task.Status == model.TaskStatusApproved {
				isCompleted = true
				break
			}
		}
	}

	return isCompleted, nil
}

// cancelPendingTasksInNode 取消节点中的待处理任务
func (s *approvalService) cancelPendingTasksInNode(approvalCode, nodeCode string) error {
	tasks, err := s.approvalTaskRepository.FindByApprovalCode(approvalCode)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.NodeCode == nodeCode && task.IsPending() {
			task.Status = model.TaskStatusCanceled
			task.Comment = "其他审批人已完成审批"
			task.UpdatedBy = "system"
			if err := s.approvalTaskService.Update(task); err != nil {
				s.logger.Error("取消任务失败", "error", err)
			}
		}
	}

	return nil
}

// 创建审批任务
func (s *approvalService) createApprovalTasksFlow(approval *model.Approval, node *model.ApprovalNode) error {
	// 开始创建审批任务

	// 解析审批人配置
	assigneeIDs, err := s.parseApproverConfigFlow(node.ApproverType, node.ApproverConfig)
	if err != nil {
		s.logger.Error("解析审批人配置失败", "error", err)
		return fmt.Errorf("解析审批人配置失败: %v", err)
	}
	// 解析审批人配置成功
	if len(assigneeIDs) == 0 {
		s.logger.Error("未找到有效的审批人")
		return errors.New("未找到有效的审批人")
	}

	// 为每个审批人创建任务
	for _, assigneeID := range assigneeIDs {
		uuid := uuid.New()
		task := &model.ApprovalTask{
			ApprovalCode: approval.Code,
			NodeCode:     node.NodeCode,
			NodeName:     node.NodeName,
			TaskCode:     strings.ToUpper(uuid.String()),
			// Title:        fmt.Sprintf("审批任务：%s", approval.Title),
			AssigneeID: assigneeID,
			Status:     model.TaskStatusPending,
			// Priority:   approval.Priority,
			Urgency: approval.Urgency,
		}

		// // 设置过期时间
		// if node.TimeoutHours > 0 {
		// 	expiredAt := time.Now().Add(time.Duration(node.TimeoutHours) * time.Hour)
		// 	task.ExpiredAt = &expiredAt
		// }

		if err := s.approvalTaskRepository.Create(task); err != nil {
			s.logger.Error("创建审批任务失败", "error", err)
			return fmt.Errorf("创建审批任务失败: %v", err)
		}
		// 创建审批任务成功
	}
	// 创建审批任务完成
	return nil
}

// 解析审批人配置
func (s *approvalService) parseApproverConfigFlow(approverType, approverConfig string) ([]string, error) {
	switch approverType {
	case model.ApproverTypeUsers:
		// 解析用户配置 {"users":["user1","user2"]}
		var config struct {
			Users []string `json:"users"`
		}
		if err := json.Unmarshal([]byte(approverConfig), &config); err != nil {
			return nil, fmt.Errorf("解析用户配置失败: %v", err)
		}
		return config.Users, nil

	case model.ApproverTypeRoles:
		// TODO: 实现角色解析
		return []string{"default_user"}, nil

	case model.ApproverTypeDepartments:
		// TODO: 实现部门解析
		return []string{"default_user"}, nil

	default:
		// 默认返回测试用户
		return []string{"default_user"}, nil
	}
}

// 根据编码查找当前任务
func (s *approvalService) findCurrentTaskByCodeFlow(approvalCode, nodeCode, assigneeID string) (*model.ApprovalTask, error) {
	where := map[string]any{
		"approval_code": approvalCode,
		"node_code":     nodeCode,
		"assignee_id":   assigneeID,
		"status":        model.TaskStatusPending,
	}

	tasks, err := s.approvalTaskRepository.Find(where)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, errors.New("未找到待处理任务")
	}

	return tasks[0], nil
}

// 根据编码同意任务
func (s *approvalService) approveTaskByCodeFlow(c *gin.Context, approval *model.Approval, task *model.ApprovalTask, comment, reason string) error {
	// 完成任务
	if err := s.approvalTaskRepository.CompleteTask(uint(task.ID), model.TaskStatusApproved, comment, reason); err != nil {
		return err
	}

	// 检查节点是否完成
	nodeCompleted, err := s.isNodeCompletedFlow(approval.Code, task.NodeCode)
	if err != nil {
		return err
	}

	if nodeCompleted {
		// 移动到下一节点
		return s.MoveToNextNodeFlow(c, approval.Code, task.NodeCode)
	}

	// Approved task by code
	return nil
}

// 根据编码拒绝任务
func (s *approvalService) rejectTaskByCodeFlow(c *gin.Context, approval *model.Approval, task *model.ApprovalTask, comment, reason string) error {
	// 完成任务
	if err := s.approvalTaskRepository.CompleteTask(uint(task.ID), model.TaskStatusRejected, comment, reason); err != nil {
		return err
	}

	// 拒绝整个审批
	return s.RejectApprovalFlow(c, approval.Code, reason)
}

// ==================== 向后兼容的方法实现 ====================

// ProcessTask 统一的任务处理方法
func (s *approvalService) ProcessTask(c *gin.Context, taskId uint, action, comment string) error {
	return s.processApprovalTask(c, taskId, action, comment)
}

// StartApproval 启动审批流程（向后兼容）
func (s *approvalService) StartApproval(c *gin.Context, approvalDefCode, applicantID, title, formData string) (*model.Approval, error) {
	return s.StartApprovalFlow(c, approvalDefCode, applicantID, title, formData)
}

func (s *approvalService) createApproval(c *gin.Context, tableCode string, approvalInfo map[string]string) error {
	// 1. 根据 tableCode 和 operation 获取审批流定义映射
	tableApprovalDefs, err := s.approvalDefinitionRepository.FindPage(1, 1000, nil, map[string]any{
		"table_code": tableCode,
		"operation":  approvalInfo["operation"],
		"status":     "Normal",
	})
	if err != nil {
		s.logger.Error("获取审批映射失败", "error", err)
		return err
	}

	if len(tableApprovalDefs) == 0 {
		// 未找到审批映射，跳过审批流程
		return nil // 跳过审批
	}

	tableApprovalDef := tableApprovalDefs[0]

	// 2. 获取审批流定义
	approvalDef, err := s.approvalDefinitionRepository.First(map[string]any{
		"code":   tableApprovalDef.Code,
		"status": "Normal",
	})
	if err != nil {
		return err
	}

	// 3. 获取审批节点列表
	approvalNodes, err := s.approvalNodeRepository.FindActiveByApprovalDefCode(approvalDef.Code)
	if err != nil {
		return err
	}

	// 4. 创建审批实例
	return s.createApprovalInstance(c, approvalDef, approvalNodes, approvalInfo)
}

// 创建审批实例的具体逻辑
func (s *approvalService) createApprovalInstance(c *gin.Context, approvalDef *model.ApprovalDefinition,
	approvalNodes []*model.ApprovalNode, approvalInfo map[string]string,
) error {
	// 1. 找到开始节点
	var startNode *model.ApprovalNode
	for _, node := range approvalNodes {
		if node.NodeType == "START" && node.SortOrder == 0 {
			startNode = node
			break
		}
	}

	if startNode == nil {
		return fmt.Errorf("未找到开始节点")
	}

	// 2. 创建审批实例
	approvalId := s.globalIdService.GetNewID("approval")
	// now := time.Now()

	approvalInstance := model.Approval{
		ID:              approvalId,
		Code:            approvalInfo["approvalCode"],
		Title:           approvalDef.Name + "-" + approvalInfo["operationName"],
		ApprovalDefCode: approvalDef.Code,
		EntityCode:      approvalInfo["entityCode"],
		CurrentTaskID:   strconv.FormatUint(uint64(startNode.ID), 10),
		CurrentTaskName: startNode.NodeName,
		SerialNumber:    s.generateSerialNumber(),
		FormData:        approvalDef.FormData,
		Description:     approvalInfo["reason"],
		Status:          "Pending",
		// StartedAt:       &now,
		// ApplicantID:     c.GetString("user_name"),
		// ApplicantName:   c.GetString("user_name"),
		CreatedBy: c.GetString("user_name"),
		UpdatedBy: c.GetString("user_name"),
	}

	if err := s.approvalRepository.Create(c, &approvalInstance); err != nil {
		return err
	}

	// 3. 创建开始任务
	if err := s.createStartTask(c, startNode, approvalInfo); err != nil {
		return err
	}

	// 4. 创建下一个审批任务
	nextNode, err := s.getNextApprovalNode(approvalNodes, startNode, approvalInfo["formData"])
	if err != nil {
		return err
	}

	if nextNode != nil {
		return s.createApprovalTask(c, nextNode, approvalInfo)
	}

	return nil
}

// createStartTask 创建开始任务
func (s *approvalService) createStartTask(c *gin.Context, startNode *model.ApprovalNode, approvalInfo map[string]string) error {
	taskId := s.globalIdService.GetNewID("approval_task")
	// now := time.Now()

	startTask := model.ApprovalTask{
		ID:           taskId,
		ApprovalCode: approvalInfo["approvalCode"],
		NodeCode:     startNode.NodeCode,
		NodeName:     startNode.NodeName,
		AssigneeID:   c.GetString("user_name"),
		AssigneeName: c.GetString("user_name"),
		Comment:      approvalInfo["reason"],
		Status:       "COMPLETED", // 开始任务直接完成
		CreatedBy:    c.GetString("user_name"),
		UpdatedBy:    c.GetString("user_name"),
		// StartedAt:    &now,
		// CompletedAt:  &now,
	}

	return s.approvalTaskService.Create(&startTask)
}

// createApprovalTask 创建审批任务
func (s *approvalService) createApprovalTask(c *gin.Context, node *model.ApprovalNode, approvalInfo map[string]string) error {
	taskId := s.globalIdService.GetNewID("approval_task")
	// now := time.Now()

	// 根据审批节点类型处理审批人
	var assigneeID, assigneeName string
	if err := s.parseApproverConfig(node, &assigneeID, &assigneeName); err != nil {
		s.logger.Error("解析审批人配置失败", "error", err)
		assigneeID = "system"
		assigneeName = "系统"
	}

	// 创建审批任务
	approvalTask := model.ApprovalTask{
		ID:           taskId,
		ApprovalCode: approvalInfo["approvalCode"],
		NodeCode:     node.NodeCode,
		NodeName:     node.NodeName,
		AssigneeID:   assigneeID,
		AssigneeName: assigneeName,
		Comment:      "",
		Status:       "PENDING",
		CreatedBy:    c.GetString("user_name"),
		UpdatedBy:    c.GetString("user_name"),
		// StartedAt:    &now,
	}

	// 处理特殊节点类型
	switch node.ApproverType {
	case "AUTO_APPROVE":
		approvalTask.Status = "APPROVED"
		approvalTask.Comment = "系统自动通过"
		// approvalTask.CompletedAt = &now
	case "AUTO_REJECT":
		approvalTask.Status = "REJECTED"
		approvalTask.Comment = "系统自动驳回"
		// approvalTask.CompletedAt = &now
	}

	return s.approvalTaskService.Create(&approvalTask)
}

func (s *approvalService) Import(c *gin.Context, tableCode, reason, operation string, r io.Reader) error {
	operationInfo := map[string]string{}
	if err := s.GetOperationInfo(operation, &operationInfo); err != nil {
		return err
	}

	// 如果是 draft，需要转换成主表名，然后取表定义
	fieldTable := tableCode
	if strings.HasSuffix(tableCode, "_draft") {
		fieldTable = tableCode[0 : len(tableCode)-6]
	}

	// 整理查询表定义的条件
	stfWhere := make(map[string]any)
	stfWhere["status"] = "Normal"
	stfWhere["table_code"] = fieldTable
	// tableFields, err := s.tableFieldService.Find("*", stfWhere)
	// if err != nil {
	//      s.logger.Error("tableFieldService.Find", "err", err)
	// }

	xlsx, err := excelize.OpenReader(r)
	if err != nil {
		return err
	}

	rows := xlsx.GetRows("Sheet1")
	// generate uuid for approval instance
	uuid := uuid.New()
	approvalCode := strings.ToUpper(uuid.String())

	tableCodeDraft := fmt.Sprintf("%s_draft", tableCode)
	var fields []string
	// 保存到草稿箱，并提交审批流
	for irow, row := range rows {
		// 生成entity struce iterface
		entity := s.tableFieldRepository.BuildEntity(tableCodeDraft)
		entityMap := make(map[string]any)
		if irow == 0 {
			fields = append(fields, row...)
		} else {
			for index, cell := range row {
				// 解决编码从文本转到unit
				if fields[index] == "ID" {
					if num, err := strconv.Atoi(cell); err != nil {
						entityMap[fields[index]] = cell
					} else {
						entityMap[fields[index]] = num
					}
				} else {
					entityMap[fields[index]] = cell
				}
			}
			// 生成 draft global id
			draftGid := s.globalIdService.GetNewID("entity_draft")
			// 生成 entity global id
			gid := s.globalIdService.GetNewID("entity")
			// 合并entityMap到entity - 使用snake_case字段名
			for k, v := range entityMap {
				entity[k] = v
			}
			entity["operation"] = operation
			entity["action"] = operationInfo["action"]
			entity["send_status"] = 0
			// new global id for new entity
			if operation == "BatchCreate" {
				entity["entity_id"] = gid
			} else {
				entity["entity_id"] = entityMap["ID"]
			}
			entity["id"] = draftGid
			entity["approval_code"] = approvalCode
			entity["draft_status"] = "Drafted"
			entity["status"] = operationInfo["status"]
			entity["created_by"] = c.GetString("user_name")

			s.logger.Debug("entity after merge", "entity", entity)

			if err := s.entityRepository.Create(c, tableCodeDraft, entity); err != nil {
				return err
			}
		}
	}

	// 保存审批实例
	approvalInfo := make(map[string]string)
	approvalInfo["operation"] = operation
	approvalInfo["approvalCode"] = approvalCode
	approvalInfo["operationName"] = operationInfo["operationName"]
	approvalInfo["action"] = operationInfo["action"]
	approvalInfo["reason"] = reason
	approvalInfo["entityCode"] = tableCode

	s.createApproval(c, tableCode, approvalInfo)
	return nil
	// return s.approvalService.ImportWithApproval(c, tableCode, reason, operation, r)
}

func (s *approvalService) generateSerialNumber() string {
	// 生成审批单编号：AP + 年月日 + 6位随机数
	now := time.Now()
	dateStr := now.Format("20060102")
	randomStr := fmt.Sprintf("%06d", now.UnixNano()%1000000)
	return fmt.Sprintf("AP%s%s", dateStr, randomStr)
}

// ConditionConfig 条件配置结构
type ConditionConfig struct {
	Branches []ConditionBranch `json:"branches"`
}

// ConditionBranch 条件分支
type ConditionBranch struct {
	Name      string          `json:"name"`
	Condition ConditionRule   `json:"condition"`
	Nodes     []ConditionNode `json:"nodes"`
}

// ConditionRule 条件规则
type ConditionRule struct {
	FieldName  string `json:"fieldName"`
	Operator   string `json:"operator"`
	FieldValue string `json:"fieldValue"`
}

// ConditionNode 条件节点
type ConditionNode struct {
	NodeCode     string `json:"nodeCode"`
	NodeName     string `json:"nodeName"`
	NodeType     string `json:"nodeType"`
	ApproverType string `json:"approverType"`
}

// evaluateConditionNode 评估条件节点并返回对应的审批节点
func (s *approvalService) evaluateConditionNode(approvalNodes []*model.ApprovalNode, conditionNode *model.ApprovalNode, formData map[string]any) (*model.ApprovalNode, error) {
	// 解析 condition_config JSON
	var conditionConfig struct {
		Branches []struct {
			Name       string               `json:"name"`
			Priority   int                  `json:"priority"`
			Condition  *ConditionRule       `json:"condition"`  // 简单条件（向后兼容）
			Expression *ConditionExpression `json:"expression"` // 复杂表达式
			Nodes      []struct {
				NodeCode string `json:"nodeCode"`
				NodeType string `json:"nodeType"`
			} `json:"nodes"`
		} `json:"branches"`
	}

	if err := json.Unmarshal([]byte(conditionNode.ConditionConfig), &conditionConfig); err != nil {
		s.logger.Error("解析条件配置失败", "error", err)
		// 如果解析失败，返回下一个审批节点作为默认行为
		return s.getDefaultNextNode(approvalNodes, conditionNode), nil
	}

	// 按优先级排序分支
	sort.Slice(conditionConfig.Branches, func(i, j int) bool {
		return conditionConfig.Branches[i].Priority < conditionConfig.Branches[j].Priority
	})

	// 遍历条件分支，按优先级评估
	for _, branch := range conditionConfig.Branches {

		// 评估条件分支
		var conditionResult bool

		// 优先使用复杂表达式
		if branch.Expression != nil {
			// 使用复杂表达式评估
			conditionResult = s.evaluateComplexCondition(branch.Expression, formData)
		} else if branch.Condition != nil {
			// 向后兼容：使用简单条件
			// 使用简单条件评估
			conditionResult = s.evaluateCondition(*branch.Condition, formData)
		} else {
			// 如果没有条件定义，视为默认分支（总是匹配）
			conditionResult = true
		}

		// 如果条件匹配，查找分支中的目标节点
		if conditionResult {
			// 条件匹配成功"
			// 查找分支中的目标节点
			for _, conditionNodeInfo := range branch.Nodes {
				targetNode := s.findNodeByCode(approvalNodes, conditionNodeInfo.NodeCode)
				if targetNode != nil {
					// 根据目标节点类型处理
					switch targetNode.NodeType {
					case "APPROVAL":
						return targetNode, nil
					case "CC":
						return targetNode, nil
					case "END":
						return nil, nil // 流程结束
					case "CONDITION":
						// 如果目标是另一个条件节点，递归评估
						return s.evaluateConditionNode(approvalNodes, targetNode, formData)
					default:
						s.logger.Warn("条件分支指向未知节点类型",
							"nodeCode", conditionNodeInfo.NodeCode,
							"nodeType", targetNode.NodeType)
					}
				} else {
					s.logger.Warn("未找到条件分支指向的节点",
						"nodeCode", conditionNodeInfo.NodeCode)
				}
			}
		}
	}

	// 如果没有匹配的条件，返回默认节点
	return s.getDefaultNextNode(approvalNodes, conditionNode), nil
}

// parseConditionExpression 解析条件表达式
func (s *approvalService) parseConditionExpression(conditionConfig string) (*ConditionExpression, error) {
	if conditionConfig == "" {
		return nil, nil
	}

	var expression ConditionExpression
	if err := json.Unmarshal([]byte(conditionConfig), &expression); err != nil {
		// 如果不是复杂表达式，尝试解析为简单条件
		var simpleCondition ConditionRule
		if err2 := json.Unmarshal([]byte(conditionConfig), &simpleCondition); err2 == nil {
			return &ConditionExpression{
				Type:      "simple",
				Condition: &simpleCondition,
			}, nil
		}
		return nil, fmt.Errorf("解析条件表达式失败: %v", err)
	}

	return &expression, nil
}

// validateConditionExpression 验证条件表达式
func (s *approvalService) validateConditionExpression(expression *ConditionExpression) error {
	if expression == nil {
		return nil
	}

	switch expression.Type {
	case "simple":
		if expression.Condition == nil {
			return fmt.Errorf("简单条件表达式缺少条件定义")
		}
		return s.validateConditionRule(*expression.Condition)

	case "and", "or":
		if len(expression.Children) == 0 {
			return fmt.Errorf("%s 表达式至少需要一个子条件", expression.Type)
		}
		for i, child := range expression.Children {
			if err := s.validateConditionExpression(child); err != nil {
				return fmt.Errorf("子表达式[%d]验证失败: %v", i, err)
			}
		}
		return nil

	case "not":
		if len(expression.Children) != 1 {
			return fmt.Errorf("NOT 表达式只能有一个子条件")
		}
		return s.validateConditionExpression(expression.Children[0])

	default:
		return fmt.Errorf("不支持的表达式类型: %s", expression.Type)
	}
}

// validateConditionRule 验证条件规则
func (s *approvalService) validateConditionRule(rule ConditionRule) error {
	if rule.FieldName == "" {
		return fmt.Errorf("字段名不能为空")
	}

	// 验证操作符
	validOperators := []string{
		"eq", "equal", "==",
		"ne", "not_equal", "!=",
		"gt", "greater_than", ">",
		"gte", "greater_than_equal", ">=",
		"lt", "less_than", "<",
		"lte", "less_than_equal", "<=",
		"contains", "like",
		"not_contains", "not_like",
		"starts_with", "ends_with",
		"in", "not_in",
		"is_empty", "is_null",
		"is_not_empty", "is_not_null",
		"regex", "regexp",
		"between",
		"default",
	}

	if slices.Contains(validOperators, rule.Operator) {
		return nil
	}

	return fmt.Errorf("不支持的操作符: %s", rule.Operator)
}

// evaluateGreaterThan 大于比较
func (s *approvalService) evaluateGreaterThan(fieldValue, expectedValue string) bool {
	// 尝试数值比较
	if num1, err1 := strconv.ParseFloat(fieldValue, 64); err1 == nil {
		if num2, err2 := strconv.ParseFloat(expectedValue, 64); err2 == nil {
			return num1 > num2
		}
	}

	// 尝试日期比较
	if date1, err1 := time.Parse("2006-01-02", fieldValue); err1 == nil {
		if date2, err2 := time.Parse("2006-01-02", expectedValue); err2 == nil {
			return date1.After(date2)
		}
	}

	// 字符串长度比较
	return len(fieldValue) > len(expectedValue)
}

// evaluateGreaterThanEqual 大于等于比较
func (s *approvalService) evaluateGreaterThanEqual(fieldValue, expectedValue string) bool {
	return s.evaluateGreaterThan(fieldValue, expectedValue) || s.evaluateEqual(fieldValue, expectedValue)
}

// evaluateLessThan 小于比较
func (s *approvalService) evaluateLessThan(fieldValue, expectedValue string) bool {
	// 尝试数值比较
	if num1, err1 := strconv.ParseFloat(fieldValue, 64); err1 == nil {
		if num2, err2 := strconv.ParseFloat(expectedValue, 64); err2 == nil {
			return num1 < num2
		}
	}

	// 尝试日期比较
	if date1, err1 := time.Parse("2006-01-02", fieldValue); err1 == nil {
		if date2, err2 := time.Parse("2006-01-02", expectedValue); err2 == nil {
			return date1.Before(date2)
		}
	}

	// 字符串长度比较
	return len(fieldValue) < len(expectedValue)
}

// evaluateLessThanEqual 小于等于比较
func (s *approvalService) evaluateLessThanEqual(fieldValue, expectedValue string) bool {
	return s.evaluateLessThan(fieldValue, expectedValue) || s.evaluateEqual(fieldValue, expectedValue)
}

// evaluateContains 包含比较
func (s *approvalService) evaluateContains(fieldValue, expectedValue string) bool {
	return strings.Contains(strings.ToLower(fieldValue), strings.ToLower(expectedValue))
}

// evaluateStartsWith 开头匹配
func (s *approvalService) evaluateStartsWith(fieldValue, expectedValue string) bool {
	return strings.HasPrefix(strings.ToLower(fieldValue), strings.ToLower(expectedValue))
}

// evaluateEndsWith 结尾匹配
func (s *approvalService) evaluateEndsWith(fieldValue, expectedValue string) bool {
	return strings.HasSuffix(strings.ToLower(fieldValue), strings.ToLower(expectedValue))
}

// evaluateIn 在列表中
func (s *approvalService) evaluateIn(fieldValue, expectedValue string) bool {
	// 解析期望值为数组，支持逗号分隔或JSON数组格式
	var values []string

	// 尝试解析为JSON数组
	if strings.HasPrefix(expectedValue, "[") && strings.HasSuffix(expectedValue, "]") {
		if err := json.Unmarshal([]byte(expectedValue), &values); err == nil {
			for _, v := range values {
				if s.evaluateEqual(fieldValue, v) {
					return true
				}
			}
			return false
		}
	}

	// 按逗号分隔
	values = strings.Split(expectedValue, ",")
	for _, v := range values {
		v = strings.TrimSpace(v)
		if s.evaluateEqual(fieldValue, v) {
			return true
		}
	}
	return false
}

// evaluateIsEmpty 判断是否为空
func (s *approvalService) evaluateIsEmpty(fieldValue any) bool {
	if fieldValue == nil {
		return true
	}

	switch v := fieldValue.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		str := s.convertToString(fieldValue)
		return strings.TrimSpace(str) == "" || str == "0" || str == "false"
	}
}

// evaluateRegex 正则表达式匹配
func (s *approvalService) evaluateRegex(fieldValue, pattern string) bool {
	matched, err := regexp.MatchString(pattern, fieldValue)
	if err != nil {
		s.logger.Error("正则表达式匹配失败",
			"pattern", pattern,
			"fieldValue", fieldValue,
			"error", err)
		return false
	}
	return matched
}

// evaluateBetween 范围比较
func (s *approvalService) evaluateBetween(fieldValue, rangeValue string) bool {
	// 解析范围值，格式："min,max" 或 "[min,max]"
	rangeValue = strings.Trim(rangeValue, "[]")
	parts := strings.Split(rangeValue, ",")
	if len(parts) != 2 {
		s.logger.Error("范围值格式错误", "rangeValue", rangeValue)
		return false
	}

	minValue := strings.TrimSpace(parts[0])
	maxValue := strings.TrimSpace(parts[1])

	// 数值范围比较
	if num, err := strconv.ParseFloat(fieldValue, 64); err == nil {
		if minNum, err1 := strconv.ParseFloat(minValue, 64); err1 == nil {
			if maxNum, err2 := strconv.ParseFloat(maxValue, 64); err2 == nil {
				return num >= minNum && num <= maxNum
			}
		}
	}

	// 日期范围比较
	if date, err := time.Parse("2006-01-02", fieldValue); err == nil {
		if minDate, err1 := time.Parse("2006-01-02", minValue); err1 == nil {
			if maxDate, err2 := time.Parse("2006-01-02", maxValue); err2 == nil {
				return (date.Equal(minDate) || date.After(minDate)) &&
					(date.Equal(maxDate) || date.Before(maxDate))
			}
		}
	}

	return false
}

// ApproverConfig 审批人配置结构
type ApproverConfig struct {
	Type  string   `json:"type"`  // USERS, ROLES, AUTO_REJECT
	Users []string `json:"users"` // 用户列表
	Mode  string   `json:"mode"`  // OR, AND
}

// Approver 审批人信息
type Approver struct {
	ID   string
	Name string
}

// evaluateEqual 等于比较
func (s *approvalService) evaluateEqual(fieldValue, expectedValue string) bool {
	// 尝试数值比较
	if num1, err1 := strconv.ParseFloat(fieldValue, 64); err1 == nil {
		if num2, err2 := strconv.ParseFloat(expectedValue, 64); err2 == nil {
			return num1 == num2
		}
	}

	// 尝试布尔值比较
	if bool1, err1 := strconv.ParseBool(fieldValue); err1 == nil {
		if bool2, err2 := strconv.ParseBool(expectedValue); err2 == nil {
			return bool1 == bool2
		}
	}

	// 尝试日期比较
	if date1, err1 := time.Parse("2006-01-02", fieldValue); err1 == nil {
		if date2, err2 := time.Parse("2006-01-02", expectedValue); err2 == nil {
			return date1.Equal(date2)
		}
	}

	// 字符串比较（不区分大小写）
	return strings.EqualFold(fieldValue, expectedValue)
}

// convertToString 将任意类型转换为字符串
func (s *approvalService) convertToString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// 私有辅助方法
// func (s *approvalService) generateSerialNumber() string {
// 	now := time.Now()
// 	return fmt.Sprintf("AP%s%06d", now.Format("20060102"), now.Unix()%1000000)
// }

func (s *approvalService) validateApproval(c *gin.Context, approval *model.Approval) error {
	if approval.Title == "" {
		return errors.New("审批标题不能为空")
	}
	if approval.ApprovalDefCode == "" {
		return errors.New("审批定义编码不能为空")
	}
	return nil
}

// generateAutocodes 为 autocode 类型的字段生成编码
func (s *approvalService) generateAutocodes(c *gin.Context, tableCode string, entityMap map[string]any) error {
	return s.autocodeService.GenerateOrRestoreAutocodes(c, tableCode, entityMap, s.tableFieldService, s.entityRepository, s.logger)
}
