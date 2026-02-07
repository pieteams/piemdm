package model

import (
	"time"

	"gorm.io/gorm"
)

type Table struct {
	ID uint `gorm:"primaryKey"`
	// 表名,英文
	Code string `gorm:"size:64;unique;not null" binding:"required,max=64,containsany=abcdefghijklmnopqrstuvwxyz0123456789_,lowercase" label:"表名"`
	// 名称，可以中文
	Name string `gorm:"size:128;not null" binding:"required,max=128"`
	// 展示模式：List 列表, Tree 树形
	DisplayMode string `gorm:"size:16;default:List;" json:"DisplayMode" validate:"oneof=List Tree"`
	// 表类型：Entity 主表实体, Item 行项目
	TableType string `gorm:"size:8;default:Entity" json:"TableType" validate:"oneof=Entity Item"`
	// 父表Code（仅当TableType=Item时有效，不带t_前缀）
	ParentTable string `gorm:"size:64;" json:"ParentTable" binding:"max=64"`
	// 父表关联字段
	ParentField string `gorm:"size:64;" json:"ParentField" binding:"max=64"`
	// 本表关联字段
	SelfField string `gorm:"size:64;" json:"SelfField" binding:"max=64"`
	// 分类表Code,用于左树右表布局
	TreeTable   string `gorm:"size:64;" json:"TreeTable" binding:"max=64"`
	Sort        uint   `gorm:"size:10;default:0" binding:"max=9999"`
	Description string `gorm:"size:255" binding:"max=255"`

	// 状态：Normal 正常 Frozen 已冻结 Deleted 已删除
	Status    string `gorm:"size:8;default:Normal"`
	CreatedBy string `gorm:"size:64"`
	UpdatedBy string `gorm:"size:64"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m *Table) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Model(m).Where("id = ?", m.ID).Update("status", "Deleted")
	return
}

// GORM 钩子：自动设置操作人
func (m *Table) BeforeCreate(tx *gorm.DB) (err error) {
	if user, ok := tx.Statement.Context.Value("user_name").(string); ok {
		m.CreatedBy = user
	}
	return
}

func (m *Table) BeforeUpdate(tx *gorm.DB) (err error) {
	if user, ok := tx.Statement.Context.Value("user_name").(string); ok {
		m.UpdatedBy = user
	}

	return
}
