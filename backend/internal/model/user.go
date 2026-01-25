package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID               uint       `gorm:"primaryKey"`
	EmployeeID       string     `gorm:"size:64;unique;not null" binding:"required,max=64"` // 工号
	Username         string     `gorm:"size:64;unique;not null" binding:"required,max=64"` // 用户名
	Password         string     `gorm:"size:128;not null" binding:"required,max=128"`      // 密码
	FirstName        string     `gorm:"size:64" binding:"max=64"`                          // 名称
	LastName         string     `gorm:"size:64" binding:"max=64"`                          // 姓氏
	DisplayName      string     `gorm:"size:64" binding:"required,max=64"`                 // 显示名称
	Email            string     `gorm:"size:255" binding:"required,max=255"`               // 邮箱
	Phone            string     `gorm:"size:64" binding:"max=64"`                          // 电话
	Language         string     `gorm:"size:8;default:zh-cn" binding:"max=8"`              // 语言
	Sex              string     `gorm:"size:8;default:Male" binding:"max=8"`               // 性别
	Avatar           string     `gorm:"size:255" binding:"max=255"`                        // 头像地址
	Description      string     `gorm:"size:255" binding:"max=255"`                        // 简介
	Admin            string     `gorm:"size:3;default:No" binding:"required,max=3"`        // 是否为管理员
	SuperiorUsername string     `gorm:"size:64" binding:"required,max=64"`                 // 上级用户名
	LastLogin        *time.Time // 最后登录时间

	// 状态：Normal 正常 Frozen 已冻结 Deleted 已删除
	Status    string `gorm:"size:8;default:Normal"`
	CreatedBy string `gorm:"size:64"` // 创建人
	UpdatedBy string `gorm:"size:64"` // 更新人
	UpdatedAt *time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Roles       []*Role  `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	Permissions []string `gorm:"-" json:"permissions,omitempty"` // 权限列表(不持久化)
}

func (m *User) BeforeDelete(tx *gorm.DB) (err error) {
	if user, ok := tx.Statement.Context.Value("user_name").(string); ok {
		m.UpdatedBy = user
	}
	tx.Model(m).Where("id = ?", m.ID).Updates(map[string]any{
		"status":     "Deleted",
		"updated_by": m.UpdatedBy,
	})
	return
}

func (m *User) BeforeCreate(tx *gorm.DB) (err error) {
	if user, ok := tx.Statement.Context.Value("user_name").(string); ok {
		m.CreatedBy = user
	}
	return
}

func (m *User) BeforeUpdate(tx *gorm.DB) (err error) {
	if user, ok := tx.Statement.Context.Value("user_name").(string); ok {
		m.UpdatedBy = user
	}

	return
}
