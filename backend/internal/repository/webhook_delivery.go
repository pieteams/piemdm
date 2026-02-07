package repository

import (
	"piemdm/internal/model"
)

type WebhookDeliveryRepository interface {
	// 基础查询
	FindOne(id uint) (*model.WebhookDelivery, error)
	FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.WebhookDelivery, error)

	// Base CRUD
	Create(webhookDelivery *model.WebhookDelivery) error
	Update(webhookDelivery *model.WebhookDelivery) error
	Delete(id uint) error

	// Batch operations
	BatchUpdate(ids []uint, webhookDelivery *model.WebhookDelivery) error
	BatchDelete(ids []uint) error
}
type webhookDeliveryRepository struct {
	*Repository
	source Base
}

func NewWebhookDeliveryRepository(repository *Repository, source Base) WebhookDeliveryRepository {
	return &webhookDeliveryRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *webhookDeliveryRepository) FindOne(id uint) (*model.WebhookDelivery, error) {
	var webhookDelivery model.WebhookDelivery
	if err := r.source.FirstById(&webhookDelivery, id); err != nil {
		return nil, err
	}
	return &webhookDelivery, nil
}

func (r *webhookDeliveryRepository) FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.WebhookDelivery, error) {
	var webhookDeliverys []*model.WebhookDelivery
	var webhookDelivery model.WebhookDelivery

	// if err := r.db.Where("deleted_at is null").Offset((page - 1) * pageSize).Limit(pageSize).Find(&webhookDeliverys).Error; err != nil {
	// 	return nil, err
	// }

	// if err := r.db.Model(webhookDelivery).Where("deleted_at is null").Count(total).Error; err != nil {
	// 	r.logger.Error("webhook_deliveries repository count err", "err", err)
	// 	return nil, err
	// }
	preloads := []string{}
	err := r.source.FindPage(webhookDelivery, &webhookDeliverys, page, pageSize, total, where, preloads, "ID desc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return webhookDeliverys, nil
}

func (r *webhookDeliveryRepository) Create(webhookDelivery *model.WebhookDelivery) error {
	if err := r.source.Create(webhookDelivery); err != nil {
		return err
	}
	return nil
}

func (r *webhookDeliveryRepository) Update(webhookDelivery *model.WebhookDelivery) error {
	if err := r.source.Updates(&webhookDelivery, webhookDelivery); err != nil {
		return err
	}
	return nil
}

func (r *webhookDeliveryRepository) BatchUpdate(ids []uint, webhookDelivery *model.WebhookDelivery) error {
	if err := r.db.Model(&webhookDelivery).Where("id in ?", ids).Updates(webhookDelivery).Error; err != nil {
		return err
	}
	return nil
}

func (r *webhookDeliveryRepository) Delete(id uint) error {
	var webhookDelivery model.WebhookDelivery
	if err := r.db.Where("id = ?", id).First(&webhookDelivery).Error; err != nil {
		return err
	}
	return nil
}

func (r *webhookDeliveryRepository) BatchDelete(ids []uint) error {
	var webhookDeliverys []model.WebhookDelivery
	// 使用下面的调用方式，才能把参数传到 Hooks 里面。
	// //单条删除
	// db.Where("id =1").Find(&webhookDelivery).Delete(&webhookDelivery)
	// //多条删除
	// db.Where("id in ?", ids).Find(&webhookDeliverys).Delete(&webhookDeliverys)

	// if err := r.db.Delete(&webhookDelivery, ids).Error; err != nil {
	if err := r.db.Where("id in ?", ids).Find(&webhookDeliverys).Delete(&webhookDeliverys).Error; err != nil {
		return err
	}
	return nil
}
