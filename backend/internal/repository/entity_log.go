package repository

import (
	"time"

	"piemdm/internal/model"

	"github.com/gin-gonic/gin"
)

type EntityLogRepository interface {
	// 基础查询
	FindOne(tableCode string, id uint) (*model.EntityLog, error)
	FindPage(tableCode string, page, pageSize int, total *int64, where map[string]any) ([]*model.EntityLog, error)

	// Base CRUD
	Create(c *gin.Context, tableCode string, entityLog *model.EntityLog) error
	Update(c *gin.Context, tableCode string, entityLog *model.EntityLog) error
	Delete(c *gin.Context, tableCode string, id uint) error

	// Batch operations
	BatchUpdate(c *gin.Context, tableCode string, ids []uint, entityLog *model.EntityLog) error
	BatchDelete(c *gin.Context, tableCode string, ids []uint) error
}
type entityLogRepository struct {
	*Repository
	source Base
}

func NewEntityLogRepository(repository *Repository, source Base) EntityLogRepository {
	return &entityLogRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *entityLogRepository) FindOne(tableCode string, id uint) (*model.EntityLog, error) {
	var entityLog model.EntityLog
	table := "t_" + tableCode + "_log"
	if err := r.db.Table(table).Where("id = ?", id).Take(&entityLog).Error; err != nil {
		return nil, err
	}
	return &entityLog, nil
}

// TODO 需要封装分页查询逻辑
func (r *entityLogRepository) FindPage(tableCode string, page, pageSize int, total *int64, where map[string]any) ([]*model.EntityLog, error) {
	var entityLogs []*model.EntityLog
	var entityLog model.EntityLog

	// if err := r.db.Where("deleted_at is null").Offset((page - 1) * pageSize).Limit(pageSize).Find(&entityLogs).Error; err != nil {
	// 	return nil, err
	// }

	// if err := r.db.Model(entityLog).Where("deleted_at is null").Count(total).Error; err != nil {
	// 	r.logger.Error("entity_logs repository count err", "err", err)
	// 	return nil, err
	// }
	preloads := []string{}
	err := r.source.FindPage(entityLog, &entityLogs, page, pageSize, total, where, preloads, "ID desc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return entityLogs, nil
}

func (r *entityLogRepository) Create(c *gin.Context, tableCode string, entityLog *model.EntityLog) error {
	table := "t_" + tableCode + "_log"

	if err := r.db.Table(table).WithContext(c).Save(entityLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *entityLogRepository) Update(c *gin.Context, tableCode string, entityLog *model.EntityLog) error {
	table := "t_" + tableCode + "_log"

	if err := r.db.Table(table).WithContext(c).Model(&entityLog).Omit("code", "created_at", "created_by").Updates(&entityLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *entityLogRepository) BatchUpdate(c *gin.Context, tableCode string, ids []uint, entityLog *model.EntityLog) error {
	table := "t_" + tableCode + "_log"

	if err := r.db.Table(table).WithContext(c).Model(&entityLog).Where("id in ?", ids).Updates(entityLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *entityLogRepository) Delete(c *gin.Context, tableCode string, id uint) error {
	var entityLog model.EntityLog
	table := "t_" + tableCode + "_log"

	// if err := r.db.Table(table).Model(&entityLog).Where("id = ?", id).Updates(map[string]any{"status": "Deleted", "deleted_at": time.Now()}).Error; err != nil {
	// 	return err
	// }
	if err := r.db.Table(table).WithContext(c).Where("id = ?", id).Delete(&entityLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *entityLogRepository) BatchDelete(c *gin.Context, tableCode string, ids []uint) error {
	var entityLog model.EntityLog
	table := "t_" + tableCode + "_log"

	if err := r.db.Table(table).WithContext(c).Model(&entityLog).Where("id in ?", ids).Updates(map[string]any{"status": "Deleted", "deleted_at": time.Now()}).Error; err != nil {
		return err
	}
	return nil
}
