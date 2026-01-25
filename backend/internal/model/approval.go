package model

import (
	"time"

	"gorm.io/gorm"
)

// Approval 审批实例模型
type Approval struct {
	ID                 uint   `gorm:"primarykey"`
	Code               string `gorm:"size:128;unique;not null"`                     // 审批实例唯一编码
	Title              string `gorm:"size:128;not null" binding:"required,max=128"` // 审批标题
	ApprovalDefCode    string `gorm:"size:128;not null" binding:"required,max=128"` // 审批定义编码
	EntityCode         string `gorm:"size:64" binding:"max=64"`                     // 关联实体编码
	ExternalInstanceID string `gorm:"size:128;index"`                               // 外部审批实例ID

	// 流程信息
	SerialNumber    string `gorm:"size:128" binding:"max=128"` // 审批单编号
	CurrentTaskID   string `gorm:"size:128" binding:"max=128"` // 当前任务ID
	CurrentTaskName string `gorm:"size:128" binding:"max=128"` // 当前任务名称

	// 表单数据
	FormData   string `gorm:"type:text"` // 表单数据JSON
	FormSchema string `gorm:"type:text"` // 表单结构JSON

	// 审批信息
	Priority    int    `gorm:"default:0"`              // 优先级
	Urgency     string `gorm:"size:16;default:Normal"` // 紧急程度：Low/Normal/High/Urgent
	Description string `gorm:"size:500"`               // 申请说明

	// 状态信息，审批实例状态
	// 可选值有：
	// - Pending：审批中
	// - Approved：通过
	// - Rejected：拒绝
	// - Canceled：撤回
	// - Deleted：删除
	Status string `gorm:"size:16;default:Pending"` // 审批状态

	// 审计字段
	CreatedBy string         `gorm:"size:64" `                // 创建人
	UpdatedBy string         `gorm:"size:64" `                // 更新人
	CreatedAt time.Time      `json:",omitempty"`              // 创建时间
	UpdatedAt time.Time      `json:",omitempty"`              // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"index" json:",omitempty"` // 删除时间
}

// TableName 指定表名
// func (Approval) TableName() string {
// 	return "approval"
// }

// BeforeCreate 创建前钩子
func (m *Approval) BeforeCreate(tx *gorm.DB) (err error) {
	// 设置默认状态
	if m.Status == "" {
		m.Status = ApprovalStatusPending
	}

	// 验证状态
	if !IsValidApprovalStatus(m.Status) {
		return gorm.ErrInvalidValue
	}

	// // 设置提交时间
	// if m.SubmittedAt == nil {
	// 	now := time.Now()
	// 	m.SubmittedAt = &now
	// }

	return nil
}

// BeforeUpdate 更新前钩子
func (m *Approval) BeforeUpdate(tx *gorm.DB) (err error) {
	// 验证状态
	if m.Status != "" && !IsValidApprovalStatus(m.Status) {
		return gorm.ErrInvalidValue
	}

	return nil
}

// BeforeDelete 删除前钩子
func (m *Approval) BeforeDelete(tx *gorm.DB) (err error) {
	// 软删除时更新状态
	return tx.Model(m).Where("id = ?", m.ID).Update("status", ApprovalStatusDeleted).Error
}

// IsPending 检查是否为待审批状态
func (m *Approval) IsPending() bool {
	return m.Status == ApprovalStatusPending
}

// IsCompleted 检查是否已完成
func (m *Approval) IsCompleted() bool {
	return m.Status == ApprovalStatusApproved || m.Status == ApprovalStatusRejected
}

// CanCancel 检查是否可以撤回
func (m *Approval) CanCancel() bool {
	return m.Status == ApprovalStatusPending
}

// // IsExpired 检查是否已过期
// func (m *Approval) IsExpired() bool {
// 	if m.ExpiredAt == nil {
// 		return false
// 	}
// 	return time.Now().After(*m.ExpiredAt)
// }

// // GetProgress 获取进度百分比
// func (m *Approval) GetProgress() float64 {
// 	if m.TaskCount == 0 {
// 		return 0
// 	}
// 	return float64(m.CompletedTasks) / float64(m.TaskCount) * 100
// }
