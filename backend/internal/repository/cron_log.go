package repository

import (
	"piemdm/internal/model"
)

type CronLogRepository interface {
	// 基础查询
	FindOne(id uint) (*model.CronLog, error)
	FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.CronLog, error)

	// Base CRUD
	Create(cronLog *model.CronLog) error
	Update(cronLog *model.CronLog) error
	Delete(id uint) (*model.CronLog, error)

	// Batch operations
	BatchUpdate(ids []uint, cronLog *model.CronLog) error
	BatchDelete(ids []uint) error
}
type cronLogRepository struct {
	*Repository
	source Base
}

func NewCronLogRepository(repository *Repository, source Base) CronLogRepository {
	return &cronLogRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *cronLogRepository) FindOne(id uint) (*model.CronLog, error) {
	var cronLog model.CronLog
	if err := r.source.FirstById(&cronLog, id); err != nil {
		return nil, err
	}
	return &cronLog, nil
}

func (r *cronLogRepository) FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.CronLog, error) {
	var cronLogs []*model.CronLog
	var cronLog model.CronLog

	// if err := r.db.Where("deleted_at is null").Offset((page - 1) * pageSize).Limit(pageSize).Find(&cronLogs).Error; err != nil {
	// 	return nil, err
	// }

	// if err := r.db.Model(cronLog).Where("deleted_at is null").Count(total).Error; err != nil {
	// 	r.logger.Error("cron_logs repository count err", "err", err)
	// 	return nil, err
	// }
	preloads := []string{}
	err := r.source.FindPage(cronLog, &cronLogs, page, pageSize, total, where, preloads, "ID desc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return cronLogs, nil
}

func (r *cronLogRepository) Create(cronLog *model.CronLog) error {
	if err := r.source.Create(cronLog); err != nil {
		return err
	}
	return nil
}

func (r *cronLogRepository) Update(cronLog *model.CronLog) error {
	if err := r.source.Updates(&cronLog, cronLog); err != nil {
		return err
	}
	return nil
}

func (r *cronLogRepository) BatchUpdate(ids []uint, cronLog *model.CronLog) error {
	if err := r.db.Model(&cronLog).Where("id in ?", ids).Updates(cronLog).Error; err != nil {
		return err
	}
	return nil
}

func (r *cronLogRepository) Delete(id uint) (*model.CronLog, error) {
	var cronLog model.CronLog
	if err := r.db.Where("id = ?", id).First(&cronLog).Error; err != nil {
		return nil, err
	}
	return &cronLog, nil
}

func (r *cronLogRepository) BatchDelete(ids []uint) error {
	var cronLogs []model.CronLog
	// 使用下面的调用方式，才能把参数传到 Hooks 里面。
	// //单条删除
	// db.Where("id =1").Find(&cronLog).Delete(&cronLog)
	// //多条删除
	// db.Where("id in ?", ids).Find(&cronLogs).Delete(&cronLogs)

	// if err := r.db.Delete(&cronLog, ids).Error; err != nil {
	if err := r.db.Where("id in ?", ids).Find(&cronLogs).Delete(&cronLogs).Error; err != nil {
		return err
	}
	return nil
}
