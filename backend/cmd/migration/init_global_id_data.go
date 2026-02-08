package main

import (
	"fmt"
	"piemdm/internal/model"
	"piemdm/pkg/log"

	"gorm.io/gorm"
)

// InitGlobalIdData 初始化全局ID数据
type InitGlobalIdData struct {
	db     *gorm.DB
	logger *log.Logger
}

// NewInitGlobalIdData 创建 InitGlobalIdData 实例
func NewInitGlobalIdData(db *gorm.DB, logger *log.Logger) *InitGlobalIdData {
	return &InitGlobalIdData{
		db:     db,
		logger: logger,
	}
}

// Run 执行全局ID数据初始化
func (m *InitGlobalIdData) Run() error {
	m.logger.Info("开始初始化全局ID数据...")

	globalIds := []model.GlobalId{
		{
			Identifier:  "entity",
			LastID:      1000000,
			Step:        1,
			Description: "主数据表编号",
		},
		{
			Identifier:  "approval",
			LastID:      2000,
			Step:        1,
			Description: "工作流相关表编号",
		},
		{
			Identifier:  "entity_draft",
			LastID:      100,
			Step:        1,
			Description: "主数据草稿表编号",
		},
		{
			Identifier:  "application",
			LastID:      100,
			Step:        1,
			Description: "应用程序",
		},
	}

	for _, gid := range globalIds {
		var count int64
		m.db.Model(&model.GlobalId{}).Where("identifier = ?", gid.Identifier).Count(&count)
		if count == 0 {
			if err := m.db.Create(&gid).Error; err != nil {
				m.logger.Error(fmt.Sprintf("创建全局ID失败: %s, err: %v", gid.Identifier, err))
				return err
			}
			m.logger.Info(fmt.Sprintf("成功初始化全局ID: %s (last_id: %d, step: %d)", gid.Identifier, gid.LastID, gid.Step))
		} else {
			m.logger.Info(fmt.Sprintf("全局ID已存在, 跳过: %s", gid.Identifier))
		}
	}

	m.logger.Info("全局ID数据初始化完成")
	return nil
}
