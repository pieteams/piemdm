package repository

import (
	"piemdm/internal/model"

	"github.com/gin-gonic/gin"
)

type WebhookRepository interface {
	// 基础查询
	FindOne(id uint) (*model.Webhook, error)
	Find(sel string, where map[string]any) ([]*model.Webhook, error)
	FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.Webhook, error)

	// Base CRUD
	Create(c *gin.Context, webhook *model.Webhook) error
	Update(c *gin.Context, webhook *model.Webhook) error
	Delete(c *gin.Context, id uint) error

	// Batch operations
	BatchUpdate(c *gin.Context, ids []uint, webhook *model.Webhook) error
	BatchDelete(c *gin.Context, ids []uint) error
}
type webhookRepository struct {
	*Repository
	source Base
}

func NewWebhookRepository(repository *Repository, source Base) WebhookRepository {
	return &webhookRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *webhookRepository) FindOne(id uint) (*model.Webhook, error) {
	var webhook model.Webhook

	if err := r.source.FirstById(&webhook, id); err != nil {
		return nil, err
	}

	return &webhook, nil
}

func (r *webhookRepository) Find(sel string, where map[string]any) ([]*model.Webhook, error) {
	var sysTableFields []*model.Webhook
	var sysTableField model.Webhook
	if sel == "" {
		sel = "*"
	}

	err := r.source.Find(sysTableField, &sysTableFields, sel, where, "id asc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return sysTableFields, nil
}

func (r *webhookRepository) FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.Webhook, error) {
	var webhooks []*model.Webhook
	var webhook model.Webhook

	preloads := []string{}
	err := r.source.FindPage(webhook, &webhooks, page, pageSize, total, where, preloads, "ID desc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return webhooks, nil
}

func (r *webhookRepository) Create(c *gin.Context, webhook *model.Webhook) error {
	if err := r.db.WithContext(c).Create(webhook).Error; err != nil {
		return err
	}
	return nil
}

func (r *webhookRepository) Update(c *gin.Context, webhook *model.Webhook) error {
	if err := r.db.WithContext(c).Updates(webhook).Error; err != nil {
		return err
	}
	return nil
}

func (r *webhookRepository) BatchUpdate(c *gin.Context, ids []uint, webhook *model.Webhook) error {
	if err := r.db.WithContext(c).Model(&webhook).Where("id in ?", ids).Updates(webhook).Error; err != nil {
		return err
	}
	return nil
}

func (r *webhookRepository) Delete(c *gin.Context, id uint) error {
	var webhook model.Webhook
	if err := r.db.WithContext(c).Where("id = ?", id).Delete(&webhook).Error; err != nil {
		return err
	}
	return nil
}

func (r *webhookRepository) BatchDelete(c *gin.Context, ids []uint) error {
	var webhooks []model.Webhook
	// 使用下面的调用方式，才能把参数传到 Hooks 里面。
	// //单条删除
	// db.Where("id =1").Find(&webhook).Delete(&webhook)
	// //多条删除
	// db.Where("id in ?", ids).Find(&webhooks).Delete(&webhooks)

	// if err := r.db.Delete(&webhook, ids).Error; err != nil {
	if err := r.db.WithContext(c).Where("id in ?", ids).Find(&webhooks).Delete(&webhooks).Error; err != nil {
		return err
	}
	return nil
}
