package repository

import (
	"piemdm/internal/model"

	"github.com/gin-gonic/gin"
)

type ApplicationRepository interface {
	// 基础查询
	FindOne(id uint) (*model.Application, error)
	FindByAppId(appId string) (*model.Application, error)
	FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.Application, error)

	// Base CRUD
	Create(c *gin.Context, application *model.Application) error
	Update(c *gin.Context, application *model.Application) error
	Delete(c *gin.Context, id uint) error

	// Batch operations
	BatchUpdate(c *gin.Context, ids []uint, application *model.Application) error
	BatchDelete(c *gin.Context, ids []uint) error
}
type applicationRepository struct {
	*Repository
	source Base
}

func NewApplicationRepository(repository *Repository, source Base) ApplicationRepository {
	return &applicationRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *applicationRepository) FindOne(id uint) (*model.Application, error) {
	var application model.Application
	if err := r.source.FirstById(&application, id); err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *applicationRepository) FindByAppId(appId string) (*model.Application, error) {
	var application model.Application
	if err := r.db.Where("app_id = ? AND status = ?", appId, "Normal").First(&application).Error; err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *applicationRepository) FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.Application, error) {
	var applications []*model.Application
	var application model.Application

	// if err := r.db.Where("deleted_at is null").Offset((page - 1) * pageSize).Limit(pageSize).Find(&applications).Error; err != nil {
	// 	return nil, err
	// }

	// if err := r.db.Model(application).Where("deleted_at is null").Count(total).Error; err != nil {
	// 	r.logger.Error("applications repository count err", "err", err)
	// 	return nil, err
	// }
	preloads := []string{}
	err := r.source.FindPage(application, &applications, page, pageSize, total, where, preloads, "ID desc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return applications, nil
}

func (r *applicationRepository) Create(c *gin.Context, application *model.Application) error {
	if err := r.db.WithContext(c).Create(application).Error; err != nil {
		return err
	}
	return nil
}

func (r *applicationRepository) Update(c *gin.Context, application *model.Application) error {
	if err := r.db.WithContext(c).Updates(application).Error; err != nil {
		return err
	}
	return nil
}

func (r *applicationRepository) BatchUpdate(c *gin.Context, ids []uint, application *model.Application) error {
	if err := r.db.WithContext(c).Model(&application).Where("id in ?", ids).Updates(application).Error; err != nil {
		return err
	}
	return nil
}

func (r *applicationRepository) Delete(c *gin.Context, id uint) error {
	if err := r.db.WithContext(c).Where("id = ?", id).Delete(&model.Application{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *applicationRepository) BatchDelete(c *gin.Context, ids []uint) error {
	var applications []model.Application
	// 使用下面的调用方式，才能把参数传到 Hooks 里面。
	// //单条删除
	// db.Where("id =1").Find(&application).Delete(&application)
	// //多条删除
	// db.Where("id in ?", ids).Find(&applications).Delete(&applications)

	// if err := r.db.Delete(&application, ids).Error; err != nil {
	if err := r.db.WithContext(c).Where("id in ?", ids).Find(&applications).Delete(&applications).Error; err != nil {
		return err
	}
	return nil
}
