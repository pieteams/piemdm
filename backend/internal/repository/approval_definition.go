//go:generate mockgen -source=approval_definition_repository.go -destination=../../test/mocks/repository/approval_definition.go -package=mock_repository

package repository

import (
	"piemdm/internal/model"

	"github.com/gin-gonic/gin"
)

type ApprovalDefinitionRepository interface {
	// 基础查询
	FindOne(id uint) (*model.ApprovalDefinition, error)
	FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.ApprovalDefinition, error)

	// Base CRUD
	Create(c *gin.Context, approvalDefinition *model.ApprovalDefinition) error
	Update(c *gin.Context, approvalDefinition *model.ApprovalDefinition) error
	Delete(c *gin.Context, id uint) (*model.ApprovalDefinition, error)

	// Batch operations
	BatchUpdate(c *gin.Context, ids []uint, approvalDefinition *model.ApprovalDefinition) error
	BatchDelete(c *gin.Context, ids []uint) error

	// 业务查询方法
	FirstByCode(code string) (*model.ApprovalDefinition, error)
	First(where map[string]any) (*model.ApprovalDefinition, error)
	FindByEntityCode(entityCode string) ([]*model.ApprovalDefinition, error)
	FindByStatus(status string) ([]*model.ApprovalDefinition, error)
	FindActiveByEntityCode(entityCode string) ([]*model.ApprovalDefinition, error)
	FindByCategory(category string) ([]*model.ApprovalDefinition, error)

	// 状态管理
	UpdateStatus(id uint, status string) error
	UpdateStatusByIds(ids []uint, status string) error

	// 版本管理
	FindVersionsByCode(code string) ([]*model.ApprovalDefinition, error)
	GetLatestVersion(code string) (*model.ApprovalDefinition, error)

	// 统计查询
	CountByStatus(status string) (int64, error)
	CountByEntityCode(entityCode string) (int64, error)
	FindActiveFeishuCodes() ([]string, error)
}

type approvalDefinitionRepository struct {
	*Repository
	source Base
}

func NewApprovalDefinitionRepository(repository *Repository, source Base) ApprovalDefinitionRepository {
	return &approvalDefinitionRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *approvalDefinitionRepository) FindOne(id uint) (*model.ApprovalDefinition, error) {
	var approvalDefinition model.ApprovalDefinition
	if err := r.source.FirstById(&approvalDefinition, id); err != nil {
		return nil, err
	}
	return &approvalDefinition, nil
}

func (r *approvalDefinitionRepository) FirstByCode(code string) (*model.ApprovalDefinition, error) {
	var approvalDefinition model.ApprovalDefinition
	if err := r.db.Where("code = ? AND deleted_at IS NULL", code).First(&approvalDefinition).Error; err != nil {
		return nil, err
	}
	return &approvalDefinition, nil
}

func (r *approvalDefinitionRepository) First(where map[string]any) (*model.ApprovalDefinition, error) {
	var approvalDefinition model.ApprovalDefinition
	if err := r.db.Where(where).First(&approvalDefinition).Error; err != nil {
		return nil, err
	}
	return &approvalDefinition, nil
}

func (r *approvalDefinitionRepository) FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.ApprovalDefinition, error) {
	var approvalDefinitions []*model.ApprovalDefinition
	var approvalDefinition model.ApprovalDefinition

	preloads := []string{}
	err := r.source.FindPage(approvalDefinition, &approvalDefinitions, page, pageSize, total, where, preloads, "created_at DESC")
	if err != nil {
		r.logger.Error("获取审批定义分页数据失败", "err", err)
	}
	return approvalDefinitions, nil
}

func (r *approvalDefinitionRepository) Create(c *gin.Context, approvalDefinition *model.ApprovalDefinition) error {
	if err := r.db.WithContext(c).Create(approvalDefinition).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalDefinitionRepository) Update(c *gin.Context, approvalDefinition *model.ApprovalDefinition) error {
	if err := r.db.WithContext(c).Updates(approvalDefinition).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalDefinitionRepository) BatchUpdate(c *gin.Context, ids []uint, approvalDefinition *model.ApprovalDefinition) error {
	if err := r.db.WithContext(c).Model(&approvalDefinition).Where("id in ?", ids).Updates(approvalDefinition).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalDefinitionRepository) Delete(c *gin.Context, id uint) (*model.ApprovalDefinition, error) {
	var approvalDefinition model.ApprovalDefinition
	if err := r.db.WithContext(c).Where("id = ?", id).First(&approvalDefinition).Error; err != nil {
		return nil, err
	}

	// 使用软删除，不物理删除数据
	if err := r.db.WithContext(c).Delete(&approvalDefinition).Error; err != nil {
		return nil, err
	}

	return &approvalDefinition, nil
}

func (r *approvalDefinitionRepository) BatchDelete(c *gin.Context, ids []uint) error {
	var approvalDefinitions []model.ApprovalDefinition

	if err := r.db.WithContext(c).Where("id in ?", ids).Find(&approvalDefinitions).Delete(&approvalDefinitions).Error; err != nil {
		return err
	}
	return nil
}

// 业务查询方法
func (r *approvalDefinitionRepository) FindByEntityCode(entityCode string) ([]*model.ApprovalDefinition, error) {
	var approvalDefinitions []*model.ApprovalDefinition
	if err := r.db.Where("entity_code = ? AND deleted_at IS NULL", entityCode).
		Order("sort_order ASC, created_at DESC").Find(&approvalDefinitions).Error; err != nil {
		return nil, err
	}
	return approvalDefinitions, nil
}

func (r *approvalDefinitionRepository) FindByStatus(status string) ([]*model.ApprovalDefinition, error) {
	var approvalDefinitions []*model.ApprovalDefinition
	if err := r.db.Where("status = ? AND deleted_at IS NULL", status).
		Order("sort_order ASC, created_at DESC").Find(&approvalDefinitions).Error; err != nil {
		return nil, err
	}
	return approvalDefinitions, nil
}

func (r *approvalDefinitionRepository) FindActiveByEntityCode(entityCode string) ([]*model.ApprovalDefinition, error) {
	var approvalDefinitions []*model.ApprovalDefinition
	if err := r.db.Where("entity_code = ? AND status = ? AND deleted_at IS NULL",
		entityCode, model.ApprovalDefStatusNormal).
		Order("sort_order ASC, created_at DESC").Find(&approvalDefinitions).Error; err != nil {
		return nil, err
	}
	return approvalDefinitions, nil
}

func (r *approvalDefinitionRepository) FindByCategory(category string) ([]*model.ApprovalDefinition, error) {
	var approvalDefinitions []*model.ApprovalDefinition
	if err := r.db.Where("category = ? AND deleted_at IS NULL", category).
		Order("sort_order ASC, created_at DESC").Find(&approvalDefinitions).Error; err != nil {
		return nil, err
	}
	return approvalDefinitions, nil
}

// 状态管理
func (r *approvalDefinitionRepository) UpdateStatus(id uint, status string) error {
	if err := r.db.Model(&model.ApprovalDefinition{}).Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalDefinitionRepository) UpdateStatusByIds(ids []uint, status string) error {
	if err := r.db.Model(&model.ApprovalDefinition{}).Where("id in ?", ids).
		Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

// 版本管理
func (r *approvalDefinitionRepository) FindVersionsByCode(code string) ([]*model.ApprovalDefinition, error) {
	var approvalDefinitions []*model.ApprovalDefinition
	if err := r.db.Where("code = ? AND deleted_at IS NULL", code).
		Order("version DESC").Find(&approvalDefinitions).Error; err != nil {
		return nil, err
	}
	return approvalDefinitions, nil
}

func (r *approvalDefinitionRepository) GetLatestVersion(code string) (*model.ApprovalDefinition, error) {
	var approvalDefinition model.ApprovalDefinition
	if err := r.db.Where("code = ? AND deleted_at IS NULL", code).
		Order("version DESC").First(&approvalDefinition).Error; err != nil {
		return nil, err
	}
	return &approvalDefinition, nil
}

// 统计查询
func (r *approvalDefinitionRepository) CountByStatus(status string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.ApprovalDefinition{}).
		Where("status = ? AND deleted_at IS NULL", status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *approvalDefinitionRepository) CountByEntityCode(entityCode string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.ApprovalDefinition{}).
		Where("entity_code = ? AND deleted_at IS NULL", entityCode).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *approvalDefinitionRepository) FindActiveFeishuCodes() ([]string, error) {
	var codes []string
	if err := r.db.Model(&model.ApprovalDefinition{}).
		Where("platform = ? AND status = ? AND deleted_at IS NULL", "Feishu", model.ApprovalDefStatusNormal).
		Pluck("code", &codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}
