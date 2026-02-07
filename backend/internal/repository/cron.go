package repository

import (
	"piemdm/internal/model"

	"github.com/gin-gonic/gin"
)

type CronRepository interface {
	// 基础查询
	FindOne(id uint) (*model.Cron, error)
	Find(sel string, where map[string]any) ([]*model.Cron, error)
	FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.Cron, error)

	// Base CRUD
	Create(c *gin.Context, cron *model.Cron) error
	Update(c *gin.Context, cron *model.Cron) error
	Delete(c *gin.Context, id uint) error

	// Batch operations
	BatchUpdate(c *gin.Context, ids []uint, cron *model.Cron) error
	BatchDelete(c *gin.Context, ids []uint) error
}
type cronRepository struct {
	*Repository
	source Base
}

func NewCronRepository(repository *Repository, source Base) CronRepository {
	return &cronRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *cronRepository) FindOne(id uint) (*model.Cron, error) {
	var cron model.Cron
	if err := r.source.FirstById(&cron, id); err != nil {
		return nil, err
	}
	return &cron, nil
}

func (r *cronRepository) FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.Cron, error) {
	var crons []*model.Cron
	var cron model.Cron

	// if err := r.db.Where("deleted_at is null").Offset((page - 1) * pageSize).Limit(pageSize).Find(&crons).Error; err != nil {
	// 	return nil, err
	// }

	// if err := r.db.Model(cron).Where("deleted_at is null").Count(total).Error; err != nil {
	// 	r.logger.Error("crons repository count err", "err", err)
	// 	return nil, err
	// }
	preloads := []string{}
	err := r.source.FindPage(cron, &crons, page, pageSize, total, where, preloads, "ID desc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return crons, nil
}

func (r *cronRepository) Find(sel string, where map[string]any) ([]*model.Cron, error) {
	var crons []*model.Cron
	var cron model.Cron
	if sel == "" {
		sel = "*"
	}

	err := r.source.Find(cron, &crons, sel, where, "id asc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return crons, nil
}

func (r *cronRepository) Create(c *gin.Context, cron *model.Cron) error {
	if err := r.db.WithContext(c).Create(cron).Error; err != nil {
		return err
	}
	return nil
}

func (r *cronRepository) Update(c *gin.Context, cron *model.Cron) error {
	if err := r.db.WithContext(c).Updates(cron).Error; err != nil {
		return err
	}
	return nil
}

func (r *cronRepository) BatchUpdate(c *gin.Context, ids []uint, cron *model.Cron) error {
	if err := r.db.WithContext(c).Model(&cron).Where("id in ?", ids).Updates(cron).Error; err != nil {
		return err
	}
	return nil
}

func (r *cronRepository) Delete(c *gin.Context, id uint) error {
	if err := r.db.WithContext(c).Where("id = ?", id).Delete(&model.Cron{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *cronRepository) BatchDelete(c *gin.Context, ids []uint) error {
	var crons []model.Cron
	// 使用下面的调用方式，才能把参数传到 Hooks 里面。
	// //单条删除
	// db.Where("id =1").Find(&cron).Delete(&cron)
	// //多条删除
	// db.Where("id in ?", ids).Find(&crons).Delete(&crons)

	// if err := r.db.Delete(&cron, ids).Error; err != nil {
	if err := r.db.WithContext(c).Where("id in ?", ids).Find(&crons).Delete(&crons).Error; err != nil {
		return err
	}
	return nil
}
