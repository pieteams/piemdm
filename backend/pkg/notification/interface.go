package notification

import (
	"context"
	"time"
)

// NotificationMessage 通知消息结构
type NotificationMessage struct {
	To          []string          `json:"to"`           // 接收者列表
	Subject     string            `json:"subject"`      // 主题/标题
	Content     string            `json:"content"`      // 内容
	ContentType string            `json:"content_type"` // 内容类型：text, html, markdown
	Template    string            `json:"template"`     // 模板名称
	Variables   map[string]any    `json:"variables"`    // 模板变量
	Priority    int               `json:"priority"`     // 优先级：1-低，2-中，3-高
	Attachments []Attachment      `json:"attachments"`  // 附件列表
	Metadata    map[string]string `json:"metadata"`     // 元数据
}

// Attachment 附件结构
type Attachment struct {
	Name        string `json:"name"`         // 文件名
	ContentType string `json:"content_type"` // 文件类型
	Content     []byte `json:"content"`      // 文件内容
	URL         string `json:"url"`          // 文件URL（可选）
}

// NotificationResult 通知发送结果
type NotificationResult struct {
	Success   bool           `json:"success"`    // 是否成功
	MessageID string         `json:"message_id"` // 消息ID
	Error     error          `json:"error"`      // 错误信息
	Details   map[string]any `json:"details"`    // 详细信息
	SentAt    time.Time      `json:"sent_at"`    // 发送时间
}

// NotificationProvider 通知提供者接口
type NotificationProvider interface {
	// GetName 获取提供者名称
	GetName() string

	// IsEnabled 检查是否启用
	IsEnabled() bool

	// Send 发送通知
	Send(ctx context.Context, message *NotificationMessage) (*NotificationResult, error)

	// ValidateConfig 验证配置
	ValidateConfig() error

	// GetSupportedTypes 获取支持的内容类型
	GetSupportedTypes() []string
}

// NotificationService 通知服务接口
type NotificationService interface {
	// RegisterProvider 注册通知提供者
	RegisterProvider(provider NotificationProvider) error

	// GetProvider 获取指定的通知提供者
	GetProvider(name string) (NotificationProvider, error)

	// GetEnabledProviders 获取所有启用的通知提供者
	GetEnabledProviders() []NotificationProvider

	// Send 发送通知（使用默认提供者）
	Send(ctx context.Context, message *NotificationMessage) (*NotificationResult, error)

	// SendWithProvider 使用指定提供者发送通知
	SendWithProvider(ctx context.Context, providerName string, message *NotificationMessage) (*NotificationResult, error)

	// SendToMultipleProviders 使用多个提供者发送通知
	SendToMultipleProviders(ctx context.Context, providerNames []string, message *NotificationMessage) ([]*NotificationResult, error)

	// SendAsync 异步发送通知
	SendAsync(ctx context.Context, message *NotificationMessage) error
}

// NotificationTemplate 通知模板接口
type NotificationTemplate interface {
	// GetName 获取模板名称
	GetName() string

	// Render 渲染模板
	Render(variables map[string]any) (string, error)

	// GetContentType 获取内容类型
	GetContentType() string
}

// NotificationQueue 通知队列接口
type NotificationQueue interface {
	// Enqueue 入队
	Enqueue(ctx context.Context, message *NotificationMessage, providerName string) error

	// Dequeue 出队
	Dequeue(ctx context.Context) (*QueueItem, error)

	// Retry 重试
	Retry(ctx context.Context, item *QueueItem) error

	// MarkComplete 标记完成
	MarkComplete(ctx context.Context, item *QueueItem) error

	// MarkFailed 标记失败
	MarkFailed(ctx context.Context, item *QueueItem, err error) error
}

// QueueItem 队列项
type QueueItem struct {
	ID           string               `json:"id"`
	Message      *NotificationMessage `json:"message"`
	ProviderName string               `json:"provider_name"`
	Attempts     int                  `json:"attempts"`
	MaxAttempts  int                  `json:"max_attempts"`
	CreatedAt    time.Time            `json:"created_at"`
	ScheduledAt  time.Time            `json:"scheduled_at"`
	Error        string               `json:"error,omitempty"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	// 邮件配置
	Email EmailConfig `yaml:"email" mapstructure:"email"`

	// 短信配置
	SMS SMSConfig `yaml:"sms" mapstructure:"sms"`

	// 飞书配置
	Feishu FeishuConfig `yaml:"feishu" mapstructure:"feishu"`

	// 钉钉配置
	Dingtalk DingtalkConfig `yaml:"dingtalk" mapstructure:"dingtalk"`

	// 自定义配置
	Custom CustomConfig `yaml:"custom" mapstructure:"custom"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
	From     string `yaml:"from" mapstructure:"from"`
	FromName string `yaml:"from_name" mapstructure:"from_name"`
	UseTLS   bool   `yaml:"use_tls" mapstructure:"use_tls"`
}

// SMSConfig 短信配置
type SMSConfig struct {
	Enabled  bool             `yaml:"enabled" mapstructure:"enabled"`
	Provider string           `yaml:"provider" mapstructure:"provider"`
	Aliyun   AliyunSMSConfig  `yaml:"aliyun" mapstructure:"aliyun"`
	Tencent  TencentSMSConfig `yaml:"tencent" mapstructure:"tencent"`
	Custom   CustomSMSConfig  `yaml:"custom" mapstructure:"custom"`
}

// AliyunSMSConfig 阿里云短信配置
type AliyunSMSConfig struct {
	AccessKeyID     string `yaml:"access_key_id" mapstructure:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret" mapstructure:"access_key_secret"`
	SignName        string `yaml:"sign_name" mapstructure:"sign_name"`
	TemplateCode    string `yaml:"template_code" mapstructure:"template_code"`
}

// TencentSMSConfig 腾讯云短信配置
type TencentSMSConfig struct {
	SecretID   string `yaml:"secret_id" mapstructure:"secret_id"`
	SecretKey  string `yaml:"secret_key" mapstructure:"secret_key"`
	SDKAppID   string `yaml:"sms_sdk_app_id" mapstructure:"sms_sdk_app_id"`
	SignName   string `yaml:"sign_name" mapstructure:"sign_name"`
	TemplateID string `yaml:"template_id" mapstructure:"template_id"`
}

// CustomSMSConfig 自定义短信配置
type CustomSMSConfig struct {
	APIURL string `yaml:"api_url" mapstructure:"api_url"`
	APIKey string `yaml:"api_key" mapstructure:"api_key"`
	Method string `yaml:"method" mapstructure:"method"`
}

// FeishuConfig 飞书配置
type FeishuConfig struct {
	Enabled    bool   `yaml:"enabled" mapstructure:"enabled"`
	AppID      string `yaml:"app_id" mapstructure:"app_id"`
	AppSecret  string `yaml:"app_secret" mapstructure:"app_secret"`
	WebhookURL string `yaml:"webhook_url" mapstructure:"webhook_url"`
}

// DingtalkConfig 钉钉配置
type DingtalkConfig struct {
	Enabled    bool   `yaml:"enabled" mapstructure:"enabled"`
	AppKey     string `yaml:"app_key" mapstructure:"app_key"`
	AppSecret  string `yaml:"app_secret" mapstructure:"app_secret"`
	WebhookURL string `yaml:"webhook_url" mapstructure:"webhook_url"`
}

// CustomConfig 自定义配置
type CustomConfig struct {
	Enabled   bool             `yaml:"enabled" mapstructure:"enabled"`
	Endpoints []CustomEndpoint `yaml:"endpoints" mapstructure:"endpoints"`
}

// CustomEndpoint 自定义端点配置
type CustomEndpoint struct {
	Name    string            `yaml:"name" mapstructure:"name"`
	URL     string            `yaml:"url" mapstructure:"url"`
	Method  string            `yaml:"method" mapstructure:"method"`
	Headers map[string]string `yaml:"headers" mapstructure:"headers"`
	Timeout string            `yaml:"timeout" mapstructure:"timeout"`
}
