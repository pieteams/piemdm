package model

import (
	"time"

	"gorm.io/gorm"
)

type ApprovalTask struct {
	ID             uint   `gorm:"primarykey"`
	TaskCode       string `gorm:"size:128;unique"`                                    // 任务唯一编码
	ApprovalCode   string `gorm:"size:128;not null;index" binding:"required,max=128"` // 审批实例编码
	NodeCode       string `gorm:"size:128;index" binding:"max=128"`                   // 节点编码
	NodeName       string `gorm:"size:128" binding:"max=128"`                         // 节点名称
	ExternalTaskID string `gorm:"size:128;index"`                                     // 外部审批任务ID

	// node_type,approver_type,approver_config 三个字段暂时没有用
	// 节点类型：START/APPROVAL/CONDITION/CC/END/PARALLEL/MERGE
	NodeType string `gorm:"size:16" binding:"max=16"` // 节点类型

	// 审批配置
	// 审批方式
	// 可选值有：
	// AND：会签
	// OR：或签
	// AUTO_PASS：自动通过
	// AUTO_REJECT：自动拒绝
	// SEQUENTIAL：按顺序
	ApproverType   string `gorm:"size:32" binding:"max=32"` // 审批类型
	ApproverConfig string `gorm:"type:text"`                // 审批人配置JSON

	// 审批信息
	// Priority     int    `gorm:"default:0"` // 优先级
	// 紧急程度：Low/Normal/High/Urgent
	Urgency     string `gorm:"size:16;default:Normal"`
	Comment     string `gorm:"size:1000"` // 审批意见
	RemindCount int    `gorm:"default:0"` // 催办次数
	Attachments string `gorm:"type:text"` // 附件列表JSON
	// 审批模式在 ApproverConfig 中
	// 审批模式：OR/AND/SEQUENTIAL
	// ApprovalMode string `gorm:"size:16"`

	// 审批人信息
	AssigneeID   string `gorm:"size:64" binding:"max=64"`                           // 审批人ID
	AssigneeName string `gorm:"size:64;index:idx_assignee_status" binding:"max=64"` // 审批人姓名
	// AssigneeDeptID string `gorm:"size:64"` // 审批人部门ID
	// AssigneeDept   string `gorm:"size:128"` // 审批人部门名称

	// 表单数据
	// FormData       string `gorm:"type:text"` // 表单数据JSON
	// FormChanges    string `gorm:"type:text"` // 表单变更JSON
	// ReadOnlyFields string `gorm:"size:500"` // 只读字段列表（逗号分隔）
	// RequiredFields string `gorm:"size:500"` // 必填字段列表（逗号分隔）

	// 通知配置
	// NotifyConfig string `gorm:"type:text"` // 通知配置JSON
	// EnableEmail  bool   `gorm:"default:true"` // 启用邮件通知
	// EnableSMS    bool   `gorm:"default:false"` // 启用短信通知
	// 用户中添加，绑定邮箱，手机号，飞书ID，钉钉ID，微信ID
	// Email      string `gorm:"size:128"` // 邮件地址
	// SMS        string `gorm:"size:128"` // 短信地址
	// FeishuID   string `gorm:"size:128"` // 飞书ID,使用userid可以向个人发消息
	// DingTalkID string `gorm:"size:128"` // 钉钉ID
	// WechatID   string `gorm:"size:128"` // 微信ID

	// 任务状态
	// 审批任务状态
	// 可选值有：
	// - PENDING：审批中
	// - APPROVED：通过
	// - REJECTED：拒绝
	// - TRANSFERRED：已转交
	// - DONE：完成
	Status string `gorm:"size:16;default:Pending;index:idx_assignee_status"` // 任务状态

	// 审计字段
	CreatedBy string `gorm:"size:64" json:",omitempty"`
	UpdatedBy string `gorm:"size:64" json:",omitempty"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" json:",omitempty"`
}

// func (ApprovalTask) TableName() string {
// 	return "approval_tasks"
// }

func (m *ApprovalTask) BeforeDelete(tx *gorm.DB) (err error) {
	if user, ok := tx.Statement.Context.Value("user_name").(string); ok {
		m.UpdatedBy = user
	}
	tx.Model(m).Where("id = ?", m.ID).Updates(map[string]any{
		"status":     "Deleted",
		"updated_by": m.UpdatedBy,
	})
	return
}

func (m *ApprovalTask) BeforeCreate(tx *gorm.DB) (err error) {
	if user, ok := tx.Statement.Context.Value("user_name").(string); ok {
		m.CreatedBy = user
	}
	return
}

func (m *ApprovalTask) BeforeUpdate(tx *gorm.DB) (err error) {
	if user, ok := tx.Statement.Context.Value("user_name").(string); ok {
		m.UpdatedBy = user
	}

	return
}

// IsPending 检查是否为待处理状态
func (m *ApprovalTask) IsPending() bool {
	return m.Status == TaskStatusPending
}

// IsCompleted 检查是否已完成
func (m *ApprovalTask) IsCompleted() bool {
	return m.Status == TaskStatusApproved || m.Status == TaskStatusRejected || m.Status == TaskStatusDone
}

// // CanTransfer 检查是否可以转交
// func (m *ApprovalTask) CanTransfer() bool {
// 	return m.AllowTransfer && m.Status == TaskStatusPending
// }

// // CanReject 检查是否可以拒绝
// func (m *ApprovalTask) CanReject() bool {
// 	return m.AllowReject && m.Status == TaskStatusPending
// }

// // IsExpired 检查是否已过期
// func (m *ApprovalTask) IsExpired() bool {
// 	if m.ExpiredAt == nil {
// 		return false
// 	}
// 	return time.Now().After(*m.ExpiredAt)
// }

// // IsOverdue 检查是否超时
// func (m *ApprovalTask) IsOverdue() bool {
// 	if m.ExpiredAt == nil || m.IsCompleted() {
// 		return false
// 	}
// 	return time.Now().After(*m.ExpiredAt)
// }

// // GetProcessDuration 获取处理时长（分钟）
// func (m *ApprovalTask) GetProcessDuration() int {
// 	if m.StartedAt == nil || m.CompletedAt == nil {
// 		return 0
// 	}
// 	return int(m.CompletedAt.Sub(*m.StartedAt).Minutes())
// }

// // MarkAsStarted 标记为开始处理
// func (m *ApprovalTask) MarkAsStarted() {
// 	if m.StartedAt == nil {
// 		now := time.Now()
// 		m.StartedAt = &now
// 	}
// }

// // MarkAsCompleted 标记为完成
// func (m *ApprovalTask) MarkAsCompleted() {
// 	if m.CompletedAt == nil {
// 		now := time.Now()
// 		m.CompletedAt = &now
// 		m.ProcessDuration = m.GetProcessDuration()
// 	}
// }
