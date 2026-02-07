package repository

import (
	"piemdm/internal/model"
)

type CronParamRepository interface {
	// 基础查询
	FindOne(id uint) (*model.CronParam, error)
	Find(sel string, where map[string]any) ([]*model.CronParam, error)
	FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.CronParam, error)

	// Base CRUD
	Create(cronParam *model.CronParam) error
	Update(cronParam *model.CronParam) error
	Delete(id uint) (*model.CronParam, error)

	// Batch operations
	BatchUpdate(ids []uint, cronParam *model.CronParam) error
	BatchDelete(ids []uint) error
}
type cronParamRepository struct {
	*Repository
	source Base
}

func NewCronParamRepository(repository *Repository, source Base) CronParamRepository {
	return &cronParamRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *cronParamRepository) FindOne(id uint) (*model.CronParam, error) {
	var cronParam model.CronParam
	if err := r.source.FirstById(&cronParam, id); err != nil {
		return nil, err
	}
	return &cronParam, nil
}

func (r *cronParamRepository) Find(sel string, where map[string]any) ([]*model.CronParam, error) {
	var cronParams []*model.CronParam
	var cronParam model.CronParam
	if sel == "" {
		sel = "*"
	}

	// err := r.source.FindPage(cronParam, &cronParams, 1, 20, total, where, []string{}, "ID desc")
	// preloads := []string{}

	err := r.source.Find(cronParam, &cronParams, sel, where, "sort asc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return cronParams, nil
}

func (r *cronParamRepository) FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.CronParam, error) {
	var cronParams []*model.CronParam
	var cronParam model.CronParam

	// if err := r.db.Where("deleted_at is null").Offset((page - 1) * pageSize).Limit(pageSize).Find(&cronParams).Error; err != nil {
	// 	return nil, err
	// }

	// if err := r.db.Model(cronParam).Where("deleted_at is null").Count(total).Error; err != nil {
	// 	r.logger.Error("cron_params repository count err", "err", err)
	// 	return nil, err
	// }
	preloads := []string{}
	err := r.source.FindPage(cronParam, &cronParams, page, pageSize, total, where, preloads, "ID desc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return cronParams, nil
}

func (r *cronParamRepository) Create(cronParam *model.CronParam) error {
	if err := r.source.Create(cronParam); err != nil {
		return err
	}
	return nil
}

func (r *cronParamRepository) Update(cronParam *model.CronParam) error {
	if err := r.source.Updates(&cronParam, cronParam); err != nil {
		return err
	}
	return nil
}

func (r *cronParamRepository) BatchUpdate(ids []uint, cronParam *model.CronParam) error {
	if err := r.db.Model(&cronParam).Where("id in ?", ids).Updates(cronParam).Error; err != nil {
		return err
	}
	return nil
}

func (r *cronParamRepository) Delete(id uint) (*model.CronParam, error) {
	var cronParam model.CronParam
	if err := r.db.Where("id = ?", id).First(&cronParam).Error; err != nil {
		return nil, err
	}
	return &cronParam, nil
}

func (r *cronParamRepository) BatchDelete(ids []uint) error {
	var cronParams []model.CronParam
	// 使用下面的调用方式，才能把参数传到 Hooks 里面。
	// //单条删除
	// db.Where("id =1").Find(&cronParam).Delete(&cronParam)
	// //多条删除
	// db.Where("id in ?", ids).Find(&cronParams).Delete(&cronParams)

	// if err := r.db.Delete(&cronParam, ids).Error; err != nil {
	if err := r.db.Where("id in ?", ids).Find(&cronParams).Delete(&cronParams).Error; err != nil {
		return err
	}
	return nil
}
