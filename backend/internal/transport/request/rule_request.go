package request

// RuleListRequest 规则列表请求参数
type RuleListRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=100"`
	Keyword  string `form:"keyword"`
	Type     string `form:"type"`
	Status   string `form:"status"`
}

// RuleCreateRequest 创建规则请求参数
type RuleCreateRequest struct {
	PolicyType   string `json:"policyType" binding:"required"`   // 策略类型: user/role
	SubjectID    string `json:"subjectId" binding:"required"`    // 主体ID(用户ID或角色ID)
	ResourceCode string `json:"resourceCode" binding:"required"` // 资源代码
	ActionType   string `json:"actionType" binding:"required"`   // 操作类型: read/write/delete
	Effect       string `json:"effect" binding:"required"`       // 权限效果: allow/deny
	DataFilter   string `json:"dataFilter"`                      // 数据过滤条件
	ReservedData string `json:"reservedData"`                    // 预留字段
}

// RuleUpdateRequest 更新规则请求参数
type RuleUpdateRequest struct {
	ID           string `json:"id" binding:"required"` // 规则ID
	PolicyType   string `json:"policyType"`            // 策略类型: user/role
	SubjectID    string `json:"subjectId"`             // 主体ID(用户ID或角色ID)
	ResourceCode string `json:"resourceCode"`          // 资源代码
	ActionType   string `json:"actionType"`            // 操作类型: read/write/delete
	Effect       string `json:"effect"`                // 权限效果: allow/deny
	DataFilter   string `json:"dataFilter"`            // 数据过滤条件
}
