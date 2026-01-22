package notification

import (
	"log/slog"
	"piemdm/pkg/log"

	"github.com/spf13/viper"
)

// NewNotificationServiceFromConfig 从配置创建通知服务
func NewNotificationServiceFromConfig(config *viper.Viper, logger *slog.Logger) (NotificationService, error) {
	// 解析通知配置
	var notificationConfig NotificationConfig
	if err := config.UnmarshalKey("notification", &notificationConfig); err != nil {
		logger.Warn("failed to parse notification config, using default", "error", err)
		// 使用默认配置
		// ...
	}

	// 从 integrations 加载飞书和钉钉配置到 notificationConfig 中
	// 这样保持内部代码结构不变，但配置来源改变了
	if err := config.UnmarshalKey("integrations.feishu", &notificationConfig.Feishu); err != nil {
		logger.Warn("failed to parse feishu integration config", "error", err)
	}
	if err := config.UnmarshalKey("integrations.dingtalk", &notificationConfig.Dingtalk); err != nil {
		logger.Warn("failed to parse dingtalk integration config", "error", err)
	}

	// 修复配置解析问题：手动设置可能解析失败的字段
	// 这是一个临时修复，用于解决 YAML 标签解析问题
	if notificationConfig.Email.Enabled {
		// 手动读取 use_tls 配置，因为结构体解析可能失败
		if config.IsSet("notification.email.use_tls") {
			notificationConfig.Email.UseTLS = config.GetBool("notification.email.use_tls")
		}

		// 确保其他关键字段也正确设置
		if notificationConfig.Email.Host == "" {
			notificationConfig.Email.Host = config.GetString("notification.email.host")
		}
		if notificationConfig.Email.Port == 0 {
			notificationConfig.Email.Port = config.GetInt("notification.email.port")
		}
		if notificationConfig.Email.Username == "" {
			notificationConfig.Email.Username = config.GetString("notification.email.username")
		}
		if notificationConfig.Email.Password == "" {
			notificationConfig.Email.Password = config.GetString("notification.email.password")
		}
		if notificationConfig.Email.From == "" {
			notificationConfig.Email.From = config.GetString("notification.email.from")
		}
		if notificationConfig.Email.FromName == "" {
			notificationConfig.Email.FromName = config.GetString("notification.email.from_name")
		}
	}

	// 创建通知工厂
	factory := NewNotificationFactory(&notificationConfig, logger)

	// 创建通知服务
	service, err := factory.CreateService()
	if err != nil {
		logger.Error("failed to create notification service", "error", err)
		return nil, err
	}

	return service, nil
}

// NewNotificationServiceProvider Wire提供者函数
func NewNotificationServiceProvider(config *viper.Viper, logger *log.Logger) NotificationService {
	service, err := NewNotificationServiceFromConfig(config, logger.Logger)
	if err != nil {
		logger.Error("failed to create notification service, returning nil", "error", err)
		return nil
	}
	return service
}
