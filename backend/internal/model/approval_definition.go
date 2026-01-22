package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ApprovalDefinition 审批定义模型
type ApprovalDefinition struct {
	ID   uint   `gorm:"primarykey"`
	Code string `gorm:"size:128;unique;not null"`                     // 审批定义唯一编码
	Name string `gorm:"size:128;not null" binding:"required,max=128"` // 审批名称
	// 不在关联实体编码，这样不同的实体可以有相同的审批定义
	// EntityCode  string `gorm:"size:64;not null" binding:"required,max=64"` // 关联实体编码
	// Category    string `gorm:"size:64" binding:"max=64"`   // 审批分类

	// 表单配置
	// FormConfig string `gorm:"type:text"` // 表单配置JSON
	// FormSchema string `gorm:"type:text"` // 表单结构定义
	FormData string `gorm:"type:text"` // 表单结构定义

	// 流程配置
	// ProcessConfig string `gorm:"type:text"` // 流程配置JSON
	// ProcessNodes  string `gorm:"type:text"` // 流程节点配置
	NodeList    string `gorm:"type:text"`                  // 流程节点列表
	Description string `gorm:"size:500" binding:"max=500"` // 审批描述

	// 审批配置
	// ApprovalMode 应该在节点配置
	// ApprovalMode   string `gorm:"size:16;default:SEQUENTIAL"` // 审批模式：OR/AND/SEQUENTIAL
	// TimeoutHours   int    `gorm:"default:72"`                 // 超时时间（小时）
	// 催办不用配置，做成功能，人工催办和自动催办
	// AutoRemind     bool   `gorm:"default:true"`               // 是否自动催办
	// RemindInterval int `gorm:"default:24"` // 催办间隔（小时）

	// 通知配置
	// 通知配置在节点上面配置，做成功能
	// NotifyConfig  string `gorm:"type:text"`     // 通知配置JSON
	// EnableEmail   bool   `gorm:"default:true"`  // 启用邮件通知
	// EnableSMS     bool   `gorm:"default:false"` // 启用短信通知
	// EnableWebhook bool   `gorm:"default:false"` // 启用Webhook通知

	// 权限配置
	// 现在的版本不用处理这个功能。
	// VisibleToAll bool   `gorm:"default:false"` // 是否对所有人可见
	// AllowedRoles string `gorm:"size:500"`      // 允许的角色列表（逗号分隔）
	// AllowedDepts string `gorm:"size:500"`      // 允许的部门列表（逗号分隔）

	// 版本管理
	// 现在版本不做版本管理，做成功能
	// Version        int    `gorm:"default:1"` // 版本号
	// ParentVersion  int    `gorm:"default:0"` // 父版本号
	// VersionComment string `gorm:"size:255"`  // 版本说明

	// 状态和元数据，
	// Builtin 系统内置,Custom 自定义,Feishu, DingTalk, WeChatWork
	Platform string `gorm:"size:16;default:Builtin"`
	// 状态:Normal 正常, Frozen 已冻结, Deleted 已删除
	Status string `gorm:"size:8;default:Normal"`
	// Priority  int    `gorm:"default:0"`             // 优先级
	// SortOrder int    `gorm:"default:0"`             // 排序
	// Tags      string `gorm:"size:255"`              // 标签（逗号分隔）

	// 审计字段
	CreatedBy string         `gorm:"size:64"` // 创建人
	UpdatedBy string         `gorm:"size:64"` // 更新人
	CreatedAt time.Time      // 创建时间
	UpdatedAt time.Time      // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index"` // 删除时间
}

// TableName 指定表名
// func (ApprovalDefinition) TableName() string {
// 	return "approval_def"
// }

// BeforeCreate 创建前钩子
func (m *ApprovalDefinition) BeforeCreate(tx *gorm.DB) (err error) {
	// Code 处理逻辑
	// 1. 如果是 Builtin (或默认)，强制自动生成，忽略前端输入，防止恶意传入
	// 2. 如果是其他平台，Code 必须由用户提供
	if m.Platform == "" || m.Platform == "Builtin" {
		uuid := uuid.New()
		m.Code = strings.ToUpper(uuid.String())
	} else {
		// 非内置平台，Code 不能为空
		if m.Code == "" {
			return gorm.ErrInvalidValue
		}
	}

	if user, ok := tx.Statement.Context.Value("user_name").(string); ok {
		m.CreatedBy = user
	}

	// 设置默认状态
	if m.Status == "" {
		m.Status = ApprovalDefStatusNormal
	}

	// 验证状态
	if !IsValidApprovalDefStatus(m.Status) {
		return gorm.ErrInvalidValue
	}

	// // 验证审批模式
	// if m.ApprovalMode != "" && !IsValidApprovalMode(m.ApprovalMode) {
	// 	return gorm.ErrInvalidValue
	// }

	return nil
}

// BeforeUpdate 更新前钩子
func (m *ApprovalDefinition) BeforeUpdate(tx *gorm.DB) (err error) {
	// 验证状态
	if m.Status != "" && !IsValidApprovalDefStatus(m.Status) {
		return gorm.ErrInvalidValue
	}

	if user, ok := tx.Statement.Context.Value("user_name").(string); ok {
		m.UpdatedBy = user
	}

	// // 验证审批模式
	// if m.ApprovalMode != "" && !IsValidApprovalMode(m.ApprovalMode) {
	// 	return gorm.ErrInvalidValue
	// }

	return nil
}

// BeforeDelete 删除前钩子
func (m *ApprovalDefinition) BeforeDelete(tx *gorm.DB) (err error) {
	if user, ok := tx.Statement.Context.Value("user_name").(string); ok {
		m.UpdatedBy = user
	}
	tx.Model(m).Where("id = ?", m.ID).Updates(map[string]any{
		"status":     ApprovalDefStatusDeleted,
		"updated_by": m.UpdatedBy,
	})
	return
}

// IsActive 检查是否为激活状态
func (m *ApprovalDefinition) IsActive() bool {
	return m.Status == ApprovalDefStatusNormal
}

// CanEdit 检查是否可以编辑
func (m *ApprovalDefinition) CanEdit() bool {
	return m.Status == ApprovalDefStatusNormal || m.Status == ApprovalDefStatusFrozen
}

// CanDelete 检查是否可以删除
func (m *ApprovalDefinition) CanDelete() bool {
	return m.Status != ApprovalDefStatusDeleted
}
