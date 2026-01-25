//go:generate mockgen -source=approval_definition_service.go -destination=../../test/mocks/service/sys_approval_def.go -package=mock_service
package service

import (
	"context"
	"errors"

	"piemdm/internal/integration/feishu"
	"piemdm/internal/model"
	"piemdm/internal/repository"

	"github.com/gin-gonic/gin"
)

type ApprovalDefinitionService interface {
	// Base CRUD
	Get(id uint) (*model.ApprovalDefinition, error)
	List(page, pageSize int, total *int64, where map[string]any) ([]*model.ApprovalDefinition, error)
	Create(c *gin.Context, approvalDefinition *model.ApprovalDefinition) error
	Update(c *gin.Context, approvalDefinition *model.ApprovalDefinition) error
	Delete(c *gin.Context, id uint) (*model.ApprovalDefinition, error)

	// Batch operations
	BatchUpdate(c *gin.Context, ids []uint, approvalDefinition *model.ApprovalDefinition) error
	BatchDelete(c *gin.Context, ids []uint) error

	// 业务方法
	GetByCode(code string) (*model.ApprovalDefinition, error)
	First(where map[string]any) (*model.ApprovalDefinition, error)
	GetByEntityCode(entityCode string) ([]*model.ApprovalDefinition, error)
	GetActiveByEntityCode(entityCode string) ([]*model.ApprovalDefinition, error)
	GetByCategory(category string) ([]*model.ApprovalDefinition, error)

	// 状态管理
	Activate(id uint) error
	Deactivate(id uint) error
	Publish(id uint) error

	// 版本管理
	GetVersions(code string) ([]*model.ApprovalDefinition, error)
	GetLatestVersion(code string) (*model.ApprovalDefinition, error)
	CreateNewVersion(c *gin.Context, id uint, comment string) (*model.ApprovalDefinition, error)

	// 验证方法
	ValidateDefinition(def *model.ApprovalDefinition) error
	CanDelete(id uint) (bool, error)
	CanEdit(id uint) (bool, error)

	// Feishu Sync
	SyncFeishuDefinition(code string) (string, error)
}

type approvalDefinitionService struct {
	*Service
	approvalDefinitionRepository repository.ApprovalDefinitionRepository
	approvalNodeRepository       repository.ApprovalNodeRepository
	feishuIntegrationService     *feishu.Service
}

func NewApprovalDefinitionService(
	service *Service,
	approvalDefinitionRepository repository.ApprovalDefinitionRepository,
	approvalNodeRepository repository.ApprovalNodeRepository,
	feishuIntegrationService *feishu.Service,
) ApprovalDefinitionService {
	return &approvalDefinitionService{
		Service:                      service,
		approvalDefinitionRepository: approvalDefinitionRepository,
		approvalNodeRepository:       approvalNodeRepository,
		feishuIntegrationService:     feishuIntegrationService,
	}
}

// 基础CRUD操作
func (s *approvalDefinitionService) Get(id uint) (*model.ApprovalDefinition, error) {
	return s.approvalDefinitionRepository.FindOne(id)
}

func (s *approvalDefinitionService) GetByCode(code string) (*model.ApprovalDefinition, error) {
	return s.approvalDefinitionRepository.FirstByCode(code)
}

func (s *approvalDefinitionService) First(where map[string]any) (*model.ApprovalDefinition, error) {
	return s.approvalDefinitionRepository.First(where)
}

func (s *approvalDefinitionService) List(page, pageSize int, total *int64, where map[string]any) ([]*model.ApprovalDefinition, error) {
	return s.approvalDefinitionRepository.FindPage(page, pageSize, total, where)
}

func (s *approvalDefinitionService) Create(c *gin.Context, approvalDefinition *model.ApprovalDefinition) error {
	// 验证审批定义
	if err := s.ValidateDefinition(approvalDefinition); err != nil {
		s.logger.Error("审批定义验证失败", "error", err)
		return err
	}

	return s.approvalDefinitionRepository.Create(c, approvalDefinition)
}

func (s *approvalDefinitionService) Update(c *gin.Context, approvalDefinition *model.ApprovalDefinition) error {
	// 检查是否可以编辑
	canEdit, err := s.CanEdit(approvalDefinition.ID)
	if err != nil {
		return err
	}
	if !canEdit {
		return errors.New("当前状态不允许编辑")
	}

	// 验证审批定义
	if err := s.ValidateDefinition(approvalDefinition); err != nil {
		s.logger.Error("审批定义验证失败", "error", err)
		return err
	}

	return s.approvalDefinitionRepository.Update(c, approvalDefinition)
}

func (s *approvalDefinitionService) BatchUpdate(c *gin.Context, ids []uint, approvalDefinition *model.ApprovalDefinition) error {
	return s.approvalDefinitionRepository.BatchUpdate(c, ids, approvalDefinition)
}

func (s *approvalDefinitionService) Delete(c *gin.Context, id uint) (*model.ApprovalDefinition, error) {
	// 检查是否可以删除
	canDelete, err := s.CanDelete(id)
	if err != nil {
		return nil, err
	}
	if !canDelete {
		return nil, errors.New("当前状态不允许删除")
	}

	// 先获取记录
	def, err := s.approvalDefinitionRepository.FindOne(id)
	if err != nil {
		return nil, err
	}

	// 执行删除
	err = s.approvalDefinitionRepository.BatchDelete(c, []uint{id})
	if err != nil {
		return nil, err
	}

	return def, nil
}

func (s *approvalDefinitionService) BatchDelete(c *gin.Context, ids []uint) error {
	// 检查每个ID是否可以删除
	for _, id := range ids {
		canDelete, err := s.CanDelete(id)
		if err != nil {
			return err
		}
		if !canDelete {
			return errors.New("存在不允许删除的审批定义")
		}
	}

	return s.approvalDefinitionRepository.BatchDelete(c, ids)
}

// 业务方法
func (s *approvalDefinitionService) GetByEntityCode(entityCode string) ([]*model.ApprovalDefinition, error) {
	return s.approvalDefinitionRepository.FindByEntityCode(entityCode)
}

func (s *approvalDefinitionService) GetActiveByEntityCode(entityCode string) ([]*model.ApprovalDefinition, error) {
	return s.approvalDefinitionRepository.FindActiveByEntityCode(entityCode)
}

func (s *approvalDefinitionService) GetByCategory(category string) ([]*model.ApprovalDefinition, error) {
	return s.approvalDefinitionRepository.FindByCategory(category)
}

// 状态管理
func (s *approvalDefinitionService) Activate(id uint) error {
	// 获取审批定义
	def, err := s.approvalDefinitionRepository.FindOne(id)
	if err != nil {
		return err
	}

	// 验证是否可以激活
	if def.Status == model.ApprovalDefStatusNormal {
		return errors.New("审批定义已经是激活状态")
	}

	// 验证流程节点是否完整
	nodes, err := s.approvalNodeRepository.FindByApprovalDefCode(def.Code)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return errors.New("审批定义缺少流程节点")
	}

	// 检查是否有开始节点和结束节点
	hasStart := false
	hasEnd := false
	for _, node := range nodes {
		if node.IsStartNode() {
			hasStart = true
		}
		if node.IsEndNode() {
			hasEnd = true
		}
	}
	if !hasStart {
		return errors.New("审批定义缺少开始节点")
	}
	if !hasEnd {
		return errors.New("审批定义缺少结束节点")
	}

	return s.approvalDefinitionRepository.UpdateStatus(id, model.ApprovalDefStatusNormal)
}

func (s *approvalDefinitionService) Deactivate(id uint) error {
	return s.approvalDefinitionRepository.UpdateStatus(id, model.ApprovalDefStatusFrozen)
}

func (s *approvalDefinitionService) Publish(id uint) error {
	// 发布前先激活
	if err := s.Activate(id); err != nil {
		return err
	}

	return nil
}

// 版本管理
func (s *approvalDefinitionService) GetVersions(code string) ([]*model.ApprovalDefinition, error) {
	return s.approvalDefinitionRepository.FindVersionsByCode(code)
}

func (s *approvalDefinitionService) GetLatestVersion(code string) (*model.ApprovalDefinition, error) {
	return s.approvalDefinitionRepository.GetLatestVersion(code)
}

func (s *approvalDefinitionService) CreateNewVersion(c *gin.Context, id uint, comment string) (*model.ApprovalDefinition, error) {
	// 获取原审批定义
	originalDef, err := s.approvalDefinitionRepository.FindOne(id)
	if err != nil {
		return nil, err
	}

	// 创建新版本
	newDef := *originalDef
	newDef.ID = 0 // 重置ID
	// newDef.Version = originalDef.Version + 1
	// newDef.ParentVersion = originalDef.Version
	// newDef.VersionComment = comment
	newDef.Status = model.ApprovalDefStatusNormal

	if err := s.approvalDefinitionRepository.Create(c, &newDef); err != nil {
		return nil, err
	}

	return &newDef, nil
}

// 验证方法
func (s *approvalDefinitionService) ValidateDefinition(def *model.ApprovalDefinition) error {
	if def.Name == "" {
		return errors.New("审批名称不能为空")
	}
	// if def.EntityCode == "" {
	// 	return errors.New("关联实体编码不能为空")
	// }
	// if def.ApprovalMode != "" && !model.IsValidApprovalMode(def.ApprovalMode) {
	// 	return errors.New("无效的审批模式")
	// }
	return nil
}

func (s *approvalDefinitionService) CanDelete(id uint) (bool, error) {
	def, err := s.approvalDefinitionRepository.FindOne(uint(id))
	if err != nil {
		return false, err
	}
	return def.CanDelete(), nil
}

func (s *approvalDefinitionService) CanEdit(id uint) (bool, error) {
	def, err := s.approvalDefinitionRepository.FindOne(id)
	if err != nil {
		return false, err
	}
	return def.CanEdit(), nil
}

func (s *approvalDefinitionService) SyncFeishuDefinition(code string) (string, error) {
	if s.feishuIntegrationService == nil {
		return "", errors.New("Feishu integration service is not available")
	}

	// 使用 background context 或传入 request context (如果能修改接口签名)
	// 这里简化使用 Background
	return s.feishuIntegrationService.GetApprovalDefinition(context.Background(), code)
}
