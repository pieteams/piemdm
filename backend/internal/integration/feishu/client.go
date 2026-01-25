package feishu

import (
	"context"
	"time"

	"piemdm/internal/configs"

	lark "github.com/larksuite/oapi-sdk-go/v3"
)

// Client 飞书客户端包装
type Client struct {
	larkClient *lark.Client
	config     configs.FeishuConfig
}

// NewClient 创建飞书客户端
func NewClient(cfg configs.FeishuConfig) *Client {
	larkClient := lark.NewClient(
		cfg.AppID,
		cfg.AppSecret,
		lark.WithReqTimeout(10*time.Second), // 设置请求超时
	)

	return &Client{
		larkClient: larkClient,
		config:     cfg,
	}
}

// GetLarkClient 获取原生 Lark Client
func (c *Client) GetLarkClient() *lark.Client {
	return c.larkClient
}

// GetConfig 获取配置
func (c *Client) GetConfig() configs.FeishuConfig {
	return c.config
}

// CheckConfig 检查配置是否有效
func (c *Client) CheckConfig() bool {
	return c.config.Enabled && c.config.AppID != "" && c.config.AppSecret != ""
}

// SendMessage 发送消息示例 (保留接口)
func (c *Client) SendMessage(ctx context.Context, receiveID string, content string) error {
	// TODO: 实现发送逻辑
	return nil
}
