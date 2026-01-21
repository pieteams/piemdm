package repository

import (
	"piemdm/internal/model"

	"github.com/gin-gonic/gin"
)

type TableApprovalDefinitionRepository interface {
	List(entityCode string, operation string) ([]model.TableApprovalDefinition, error)
	Get(id uint) (*model.TableApprovalDefinition, error)
	Create(c *gin.Context, item *model.TableApprovalDefinition) error
	Update(c *gin.Context, item *model.TableApprovalDefinition) error
	Delete(c *gin.Context, id uint) error
	BatchCreate(c *gin.Context, list []model.TableApprovalDefinition) error
	BatchDelete(c *gin.Context, ids []uint) error
}

type tableApprovalDefinitionRepository struct {
	*Repository
	source Base
}

func NewTableApprovalDefinitionRepository(repository *Repository, source Base) TableApprovalDefinitionRepository {
	return &tableApprovalDefinitionRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *tableApprovalDefinitionRepository) List(entityCode, operation string) ([]model.TableApprovalDefinition, error) {
	var tableApprovalDefinitions []model.TableApprovalDefinition
	var tableApprovalDefinition model.TableApprovalDefinition

	where := map[string]any{
		"entity_code": entityCode,
	}
	if operation != "" {
		where["operation"] = operation
	}

	err := r.source.Find(tableApprovalDefinition, &tableApprovalDefinitions, "*", where, "id asc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return tableApprovalDefinitions, nil
}

func (r *tableApprovalDefinitionRepository) Get(id uint) (*model.TableApprovalDefinition, error) {
	var item model.TableApprovalDefinition
	err := r.db.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *tableApprovalDefinitionRepository) Create(c *gin.Context, item *model.TableApprovalDefinition) error {
	return r.db.WithContext(c).Create(item).Error
}

func (r *tableApprovalDefinitionRepository) Update(c *gin.Context, item *model.TableApprovalDefinition) error {
	return r.db.WithContext(c).Model(&model.TableApprovalDefinition{}).Where("id = ?", item.ID).Updates(item).Error
}

func (r *tableApprovalDefinitionRepository) Delete(c *gin.Context, id uint) error {
	var tableApprovalDefinition model.TableApprovalDefinition
	if err := r.db.WithContext(c).Where("id = ?", id).Find(&tableApprovalDefinition).Delete(&tableApprovalDefinition).Error; err != nil {
		return err
	}
	return nil
}

func (r *tableApprovalDefinitionRepository) BatchCreate(c *gin.Context, list []model.TableApprovalDefinition) error {
	return r.db.WithContext(c).Create(&list).Error
}

func (r *tableApprovalDefinitionRepository) BatchDelete(c *gin.Context, ids []uint) error {
	return r.db.WithContext(c).Delete(&model.TableApprovalDefinition{}, ids).Error
}
