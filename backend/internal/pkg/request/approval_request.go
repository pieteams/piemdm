package request

import "time"

// 审批定义相关请求结构体

// CreateApprovalDefRequest 创建审批定义请求
type CreateApprovalDefRequest struct {
	Name string `json:"name" binding:"required,min=1,max=128" comment:"审批名称"`
	// EntityCode  string `json:"entityCode" binding:"required,min=1,max=64" comment:"关联实体编码"`
	// Category    string `json:"category" binding:"max=64" comment:"审批分类"`
	Description string `json:"description" binding:"max=500" comment:"审批描述"`
	FormData    string `json:"formData" binding:"max=10000" comment:"表单配置"`
	NodeList    string `json:"nodeList" binding:"max=10000" comment:"流程节点列表"`
	Platform    string `json:"platform" binding:"required,oneof=Builtin Feishu DingTalk WeChatWork Custom" comment:"审批系统"`

	// 流程配置
	// ApprovalMode string `json:"approvalMode" binding:"required,oneof=OR AND SEQUENTIAL" comment:"审批模式"`
	// TimeoutHours int    `json:"timeoutHours" binding:"min=1,max=720" comment:"超时时间(小时)"`
	// Priority     int    `json:"priority" binding:"min=0,max=10" comment:"优先级"`

	// 通知配置
	// EnableEmail   bool `json:"enableEmail" comment:"启用邮件通知"`
	// EnableSMS     bool `json:"enableSMS" comment:"启用短信通知"`
	// EnableWebhook bool `json:"enableWebhook" comment:"启用Webhook通知"`

	// 权限配置
	// VisibleToAll bool   `json:"visibleToAll" comment:"是否对所有人可见"`
	// AllowedRoles string `json:"allowedRoles" binding:"max=500" comment:"允许的角色列表"`
	// AllowedDepts string `json:"allowedDepts" binding:"max=500" comment:"允许的部门列表"`

	// 配置JSON
	// FormConfig    string `json:"formConfig" binding:"max=10000" comment:"表单配置"`
	// ProcessConfig string `json:"processConfig" binding:"max=10000" comment:"流程配置"`
	// NotifyConfig  string `json:"notifyConfig" binding:"max=5000" comment:"通知配置"`

	Status string `json:"status" binding:"oneof=Normal Frozen Deleted" comment:"状态"`
	// Tags   string `json:"tags" binding:"max=255" comment:"标签"`
}

// UpdateApprovalDefRequest 更新审批定义请求
type UpdateApprovalDefRequest struct {
	ID       uint   `json:"id" binding:"required,gt=0" comment:"审批定义ID"`
	Name     string `json:"name" binding:"required,min=1,max=128" comment:"审批名称"`
	Platform string `json:"platform" binding:"required,oneof=Builtin Feishu DingTalk WeChatWork Custom" comment:"审批系统"`
	// EntityCode  string `json:"entityCode" binding:"required,min=1,max=64" comment:"关联实体编码"`
	// Category    string `json:"category" binding:"max=64" comment:"审批分类"`
	Description string `json:"description" binding:"max=500" comment:"审批描述"`
	FormData    string `json:"formData" binding:"max=10000" comment:"表单配置"`
	NodeList    string `json:"nodeList" binding:"max=10000" comment:"流程节点列表"`

	// 流程配置
	// ApprovalMode string `json:"approvalMode" binding:"required,oneof=OR AND SEQUENTIAL" comment:"审批模式"`
	// TimeoutHours int    `json:"timeoutHours" binding:"min=1,max=720" comment:"超时时间(小时)"`
	// Priority     int    `json:"priority" binding:"min=0,max=10" comment:"优先级"`

	// 通知配置
	// EnableEmail   bool `json:"enableEmail" comment:"启用邮件通知"`
	// EnableSMS     bool `json:"enableSMS" comment:"启用短信通知"`
	// EnableWebhook bool `json:"enableWebhook" comment:"启用Webhook通知"`

	// 权限配置
	// VisibleToAll bool   `json:"visibleToAll" comment:"是否对所有人可见"`
	// AllowedRoles string `json:"allowedRoles" binding:"max=500" comment:"允许的角色列表"`
	// AllowedDepts string `json:"allowedDepts" binding:"max=500" comment:"允许的部门列表"`

	// 配置JSON
	// FormConfig    string `json:"formConfig" binding:"max=10000" comment:"表单配置"`
	// ProcessConfig string `json:"processConfig" binding:"max=10000" comment:"流程配置"`
	// NotifyConfig  string `json:"notifyConfig" binding:"max=5000" comment:"通知配置"`

	Status string `json:"status" binding:"oneof=Normal Frozen Deleted" comment:"状态"`
	// Tags   string `json:"tags" binding:"max=255" comment:"标签"`

	// 注意：不包含以下系统管理的字段
	// - Code: 系统自动生成，不允许修改
	// - Version: 版本控制，系统管理
	// - CreatedBy/UpdatedBy: 审计字段，系统自动设置
	// - CreatedAt/UpdatedAt: 时间戳，系统自动管理
}

// ActivateApprovalDefRequest 激活审批定义请求
type ActivateApprovalDefRequest struct {
	ID uint `json:"id" binding:"required,gt=0" comment:"审批定义ID"`
}

// CreateVersionRequest 创建版本请求
type CreateVersionRequest struct {
	ID      uint   `json:"id" binding:"required,gt=0" comment:"审批定义ID"`
	Comment string `json:"comment" binding:"required,min=1,max=255" comment:"版本说明"`
}

// 审批节点相关请求结构体

// CreateApprovalNodeRequest 创建审批节点请求
type CreateApprovalNodeRequest struct {
	ApprovalDefCode string `json:"ApprovalDefCode" binding:"max=128" comment:"审批定义编码"`
	NodeCode        string `json:"NodeCode" binding:"required,min=1,max=64" comment:"节点编码"`
	NodeName        string `json:"NodeName" binding:"required,min=1,max=128" comment:"节点名称"`
	NodeType        string `json:"NodeType" binding:"required,oneof=START APPROVAL CONDITION AUTO_REJECT AUTO_APPROVE CC END" comment:"节点类型"`
	ApproverType    string `json:"ApproverType" binding:"oneof=USERS ROLES DEPARTMENTS AUTO_REJECT AUTO_APPROVE SYSTEM" comment:"审批人类型"`
	ApproverConfig  string `json:"ApproverConfig" binding:"max=2000" comment:"审批人配置"`
	// ApprovalMode    string `json:"ApprovalMode" binding:"oneof=OR AND SEQUENTIAL" comment:"审批模式"`
	ConditionConfig string `json:"ConditionConfig" binding:"max=5000" comment:"条件配置"`
	// // TimeoutHours    int    `json:"TimeoutHours" binding:"min=0,max=720" comment:"超时时间(小时)"`
	SortOrder int `json:"SortOrder" binding:"min=0" comment:"排序"`
	// // IsRequired       bool   `json:"IsRequired" comment:"是否必须"`
	// // AllowReject      bool   `json:"AllowReject" comment:"允许拒绝"`
	// // AllowTransfer    bool   `json:"AllowTransfer" comment:"允许转交"`
	// // AutoApprove      bool   `json:"AutoApprove" comment:"自动审批"`
	// // NotifyConfig     string `json:"NotifyConfig" binding:"max=2000" comment:"通知配置"`
	// // FormConfig       string `json:"FormConfig" binding:"max=5000" comment:"表单配置"`
	// // PermissionConfig string `json:"PermissionConfig" binding:"max=2000" comment:"权限配置"`
	// // Remark           string `json:"Remark" binding:"max=500" comment:"备注"`
	Status string `json:"status" binding:"oneof=Normal Frozen Deleted" comment:"状态"`
}

// UpdateApprovalNodeRequest 更新审批节点请求
type UpdateApprovalNodeRequest struct {
	ID               uint   `json:"id" binding:"required,gt=0" comment:"节点ID"`
	NodeName         string `json:"nodeName" binding:"required,min=1,max=128" comment:"节点名称"`
	NodeType         string `json:"nodeType" binding:"required,oneof=START APPROVAL CONDITION PARALLEL MERGE CC END" comment:"节点类型"`
	ApproverType     string `json:"approverType" binding:"oneof=USER ROLE DEPT EXPRESSION" comment:"审批人类型"`
	ApproverConfig   string `json:"approverConfig" binding:"max=2000" comment:"审批人配置"`
	ApprovalMode     string `json:"approvalMode" binding:"oneof=OR AND SEQUENTIAL" comment:"审批模式"`
	ConditionConfig  string `json:"conditionConfig" binding:"max=5000" comment:"条件配置"`
	TimeoutHours     int    `json:"timeoutHours" binding:"min=0,max=720" comment:"超时时间(小时)"`
	SortOrder        int    `json:"sortOrder" binding:"min=0" comment:"排序"`
	IsRequired       bool   `json:"isRequired" comment:"是否必须"`
	AllowReject      bool   `json:"allowReject" comment:"允许拒绝"`
	AllowTransfer    bool   `json:"allowTransfer" comment:"允许转交"`
	AutoApprove      bool   `json:"autoApprove" comment:"自动审批"`
	NotifyConfig     string `json:"notifyConfig" binding:"max=2000" comment:"通知配置"`
	FormConfig       string `json:"formConfig" binding:"max=5000" comment:"表单配置"`
	PermissionConfig string `json:"permissionConfig" binding:"max=2000" comment:"权限配置"`
	Remark           string `json:"remark" binding:"max=500" comment:"备注"`
}

// 审批实例相关请求结构体

// StartApprovalRequest 启动审批请求
type StartApprovalRequest struct {
	ApprovalDefCode string     `json:"approvalDefCode" binding:"required,min=1,max=128" comment:"审批定义编码"`
	Title           string     `json:"title" binding:"required,min=1,max=255" comment:"审批标题"`
	EntityID        string     `json:"entityId" binding:"max=64" comment:"关联实体ID"`
	FormData        string     `json:"formData" binding:"max=20000" comment:"表单数据"`
	Priority        int        `json:"priority" binding:"min=0,max=10" comment:"优先级"`
	Urgency         string     `json:"urgency" binding:"oneof=LOW NORMAL HIGH URGENT" comment:"紧急程度"`
	ExpectedDate    *time.Time `json:"expectedDate" comment:"期望完成时间"`
	Remark          string     `json:"remark" binding:"max=1000" comment:"备注"`
}

// SubmitApprovalRequest 提交审批请求
type SubmitApprovalRequest struct {
	Code string `json:"code" binding:"required,min=1,max=128" comment:"审批编码"`
}

// CancelApprovalRequest 取消审批请求
type CancelApprovalRequest struct {
	Code   string `json:"code" binding:"required,min=1,max=128" comment:"审批编码"`
	Reason string `json:"reason" binding:"required,min=1,max=500" comment:"撤回原因"`
}

// ProcessApprovalRequest 处理审批请求
type ProcessApprovalRequest struct {
	ApprovalCode string `json:"approvalCode" binding:"required,min=1,max=128" comment:"审批编码"`
	NodeCode     string `json:"nodeCode" binding:"required,min=1,max=64" comment:"节点编码"`
	Action       string `json:"action" binding:"required,oneof=APPROVE REJECT TRANSFER DELEGATE" comment:"操作类型"`
	Comment      string `json:"comment" binding:"max=1000" comment:"审批意见"`
	Reason       string `json:"reason" binding:"max=500" comment:"操作原因"`
}

// 审批任务相关请求结构体

// ProcessTaskRequest 处理任务请求
type ProcessTaskRequest struct {
	TaskID  uint   `json:"taskId" binding:"required,gt=0" comment:"任务ID"`
	Action  string `json:"action" binding:"required,oneof=APPROVE REJECT TRANSFER DELEGATE" comment:"操作类型"`
	Comment string `json:"comment" binding:"max=1000" comment:"审批意见"`
	Reason  string `json:"reason" binding:"max=500" comment:"操作原因"`
}

// ApprovalTaskActionRequest 审批任务操作请求（用于ApproveTask和RejectTask接口，TaskID从URL获取）
type ApprovalTaskActionRequest struct {
	Action  string `json:"action" binding:"oneof=APPROVE REJECT" comment:"操作类型"`
	Comment string `json:"comment" binding:"max=1000" comment:"审批意见"`
	Reason  string `json:"reason" binding:"max=500" comment:"操作原因"`
}

// TransferTaskRequest 转交任务请求
type TransferTaskRequest struct {
	TaskID     uint   `json:"taskId" binding:"required,gt=0" comment:"任务ID"`
	ToUserID   string `json:"toUserId" binding:"required,min=1,max=64" comment:"转交给用户ID"`
	ToUserName string `json:"toUserName" binding:"required,min=1,max=128" comment:"转交给用户名"`
	Reason     string `json:"reason" binding:"required,min=1,max=500" comment:"转交原因"`
}

// RemindTaskRequest 催办任务请求
type RemindTaskRequest struct {
	TaskID uint `json:"taskId" binding:"required,gt=0" comment:"任务ID"`
}

// BatchRemindRequest 批量催办请求
type BatchRemindRequest struct {
	AssigneeID string `json:"assigneeId" binding:"required,min=1,max=64" comment:"审批人ID"`
}

// 通用请求结构体

// BatchOperationRequest Batch operations请求
type BatchOperationRequest struct {
	IDs    []uint `json:"ids" binding:"required,min=1,dive,gt=0" comment:"ID列表"`
	Status string `json:"status" binding:"required,oneof=Normal Frozen Deleted" comment:"状态"`
}

// 查询请求结构体

// QueryApprovalDefRequest 查询审批定义请求
type QueryApprovalDefRequest struct {
	Code string `form:"code" binding:"max=128" comment:"审批编码"`
	Name string `form:"name" binding:"max=128" comment:"审批名称"`
	// EntityCode string `form:"entityCode" binding:"max=64" comment:"实体编码"`
	// Category   string `form:"category" binding:"max=64" comment:"分类"`
	Status    string `form:"status" binding:"omitempty,oneof=Normal Frozen Deleted" comment:"状态"`
	StartDate string `form:"startDate" binding:"omitempty,datetime=2006-01-02" comment:"开始日期"`
	EndDate   string `form:"endDate" binding:"omitempty,datetime=2006-01-02" comment:"结束日期"`
	Page      int    `form:"page,default=1" comment:"页码"`
	PageSize  int    `form:"pageSize,default=15" comment:"页大小"`
}

// QueryApprovalRequest 查询审批请求
type QueryApprovalRequest struct {
	Code        string `form:"code" binding:"max=128" comment:"审批编码"`
	Title       string `form:"title" binding:"max=255" comment:"标题"`
	Status      string `form:"status" binding:"oneof=DRAFT PENDING APPROVED REJECTED CANCELLED" comment:"状态"`
	ApplicantID string `form:"applicantId" binding:"max=64" comment:"申请人ID"`
	// EntityCode  string `form:"entityCode" binding:"max=64" comment:"实体编码"`
	ShowType  string `form:"showType" binding:"oneof=ALL MY_APPLY MY_APPROVE MY_CC" comment:"显示类型"`
	StartDate string `form:"startDate" binding:"omitempty,datetime=2006-01-02" comment:"开始日期"`
	EndDate   string `form:"endDate" binding:"omitempty,datetime=2006-01-02" comment:"结束日期"`
	Page      int    `form:"page,default=1"  comment:"页码"`
	PageSize  int    `form:"pageSize,default=15"  comment:"页大小"`
}

// QueryApprovalTaskRequest 查询审批任务请求
type QueryApprovalTaskRequest struct {
	ApprovalCode string `form:"approvalCode" binding:"max=128" comment:"审批编码"`
	NodeCode     string `form:"nodeCode" binding:"max=64" comment:"节点编码"`
	AssigneeID   string `form:"assigneeId" binding:"max=64" comment:"审批人ID"`
	Status       string `form:"status" binding:"oneof=PENDING APPROVED REJECTED TRANSFERRED TIMEOUT" comment:"状态"`
	StartDate    string `form:"startDate" binding:"omitempty,datetime=2006-01-02" comment:"开始日期"`
	EndDate      string `form:"endDate" binding:"omitempty,datetime=2006-01-02" comment:"结束日期"`
	Page         int    `form:"page,default=1"  comment:"页码"`
	PageSize     int    `form:"pageSize,default=15" comment:"页大小"`
}
