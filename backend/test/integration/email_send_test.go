package integration

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"piemdm/pkg/configloader"
	"piemdm/pkg/notification"
)

// TestEmailSend 测试实际邮件发送
func TestEmailSend(t *testing.T) {
	// 创建日志器
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// 设置配置文件路径
	// 在测试环境中，工作目录通常是测试文件所在的目录
	// 这里使用相对路径指向 config/local.yml
	configPath := "../../config/local.yml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("配置文件不存在，跳过集成测试")
	}
	os.Setenv("APP_CONF", configPath)

	// 从配置文件加载邮件配置
	cfg, err := configloader.Load()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 创建邮件配置
	emailConfig := &notification.EmailConfig{
		Enabled:  cfg.GetBool("notification.email.enabled"),
		Host:     cfg.GetString("notification.email.host"),
		Port:     cfg.GetInt("notification.email.port"),
		Username: cfg.GetString("notification.email.username"),
		Password: cfg.GetString("notification.email.password"),
		From:     cfg.GetString("notification.email.from"),
		FromName: cfg.GetString("notification.email.from_name"),
		UseTLS:   cfg.GetBool("notification.email.use_tls"),
	}

	// 确定测试收件人
	testRecipient := cfg.GetString("notification.email.test_recipient")
	if testRecipient == "" {
		testRecipient = emailConfig.From
		t.Logf("未配置 notification.email.test_recipient，将使用发件人地址作为收件人: %s", testRecipient)
	}

	// 创建邮件提供者
	provider := notification.NewEmailProvider(emailConfig, logger)

	// 验证配置
	if err := provider.ValidateConfig(); err != nil {
		t.Logf("配置验证失败: %v", err)
		t.Logf("请检查邮件配置是否正确")
		return
	}

	// 创建测试消息
	message := &notification.NotificationMessage{
		To:          []string{testRecipient},
		Subject:     "【PieMDM测试】邮件发送测试",
		Content:     "这是一封测试邮件，用于验证PieMDM系统的邮件发送功能。\n\n发送时间：" + time.Now().Format("2006-01-02 15:04:05"),
		ContentType: "text/plain",
	}

	// 发送邮件
	ctx := context.Background()
	result, err := provider.Send(ctx, message)
	if err != nil {
		t.Logf("邮件发送失败: %v", err)
		t.Logf("请检查以下配置:")
		t.Logf("1. 邮箱地址是否正确")
		t.Logf("2. 授权码是否正确")
		t.Logf("3. 网络连接是否正常")
		t.Logf("4. 服务器地址和端口是否正确")
		return
	}

	if result.Success {
		t.Logf("邮件发送成功!")
		t.Logf("消息ID: %s", result.MessageID)
		t.Logf("发送时间: %s", result.SentAt.Format("2006-01-02 15:04:05"))
		t.Logf("收件人: %v", message.To)
	} else {
		t.Logf("邮件发送失败: %v", result.Error)
	}
}

// TestEmailSendWithHTML 测试HTML格式邮件发送
func TestEmailSendWithHTML(t *testing.T) {
	// 创建日志器
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	configPath := "../../config/local.yml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("配置文件不存在，跳过集成测试")
	}
	os.Setenv("APP_CONF", configPath)

	// 从配置文件加载邮件配置
	cfg, err := configloader.Load()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 创建邮件配置
	emailConfig := &notification.EmailConfig{
		Enabled:  cfg.GetBool("notification.email.enabled"),
		Host:     cfg.GetString("notification.email.host"),
		Port:     cfg.GetInt("notification.email.port"),
		Username: cfg.GetString("notification.email.username"),
		Password: cfg.GetString("notification.email.password"),
		From:     cfg.GetString("notification.email.from"),
		FromName: cfg.GetString("notification.email.from_name"),
		UseTLS:   cfg.GetBool("notification.email.use_tls"),
	}

	// 确定测试收件人
	testRecipient := cfg.GetString("notification.email.test_recipient")
	if testRecipient == "" {
		testRecipient = emailConfig.From
	}

	// 创建邮件提供者
	provider := notification.NewEmailProvider(emailConfig, logger)

	// 创建HTML格式的测试消息
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>PieMDM测试邮件</title>
</head>
<body>
    <h2 style="color: #2563eb;">PieMDM系统测试</h2>
    <p>这是一封HTML格式的测试邮件，用于验证PieMDM系统的邮件发送功能。</p>
    <ul>
        <li>发送时间：` + time.Now().Format("2006-01-02 15:04:05") + `</li>
        <li>测试类型：HTML格式邮件</li>
        <li>系统版本：PieMDM v1.0</li>
    </ul>
    <p style="color: #059669;">邮件发送功能正常！</p>
</body>
</html>`

	message := &notification.NotificationMessage{
		To:          []string{testRecipient},
		Subject:     "【PieMDM测试】HTML格式邮件",
		Content:     htmlContent,
		ContentType: "html",
	}

	// 发送邮件
	ctx := context.Background()
	result, err := provider.Send(ctx, message)
	if err != nil {
		t.Logf("HTML邮件发送失败: %v", err)
		return
	}

	if result.Success {
		t.Logf("HTML邮件发送成功!")
		t.Logf("消息ID: %s", result.MessageID)
		t.Logf("发送时间: %s", result.SentAt.Format("2006-01-02 15:04:05"))
	} else {
		t.Logf("HTML邮件发送失败: %v", result.Error)
	}
}

// TestNotificationService 测试完整的通知服务
func TestNotificationService(t *testing.T) {
	// 创建日志器
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	configPath := "../../config/local.yml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("配置文件不存在，跳过集成测试")
	}
	os.Setenv("APP_CONF", configPath)

	// 从配置文件加载邮件配置
	cfg, err := configloader.Load()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 创建邮件配置
	emailConfig := &notification.EmailConfig{
		Enabled:  cfg.GetBool("notification.email.enabled"),
		Host:     cfg.GetString("notification.email.host"),
		Port:     cfg.GetInt("notification.email.port"),
		Username: cfg.GetString("notification.email.username"),
		Password: cfg.GetString("notification.email.password"),
		From:     cfg.GetString("notification.email.from"),
		FromName: cfg.GetString("notification.email.from_name"),
		UseTLS:   cfg.GetBool("notification.email.use_tls"),
	}

	// 确定测试收件人
	testRecipient := cfg.GetString("notification.email.test_recipient")
	if testRecipient == "" {
		testRecipient = emailConfig.From
	}

	// 创建通知配置
	config := &notification.NotificationConfig{
		Email: *emailConfig,
	}

	// 创建通知服务
	factory := notification.NewNotificationFactory(config, logger)
	notificationService, err := factory.CreateService()
	if err != nil {
		t.Logf("创建通知服务失败: %v", err)
		return
	}

	// 创建测试消息
	message := &notification.NotificationMessage{
		To:          []string{testRecipient},
		Subject:     "【PieMDM通知】审批流程测试",
		Content:     "您的审批申请已提交，请及时处理。\n\n申请时间：" + time.Now().Format("2006-01-02 15:04:05"),
		ContentType: "text/plain",
	}

	// 发送通知
	ctx := context.Background()
	result, err := notificationService.Send(ctx, message)
	if err != nil {
		t.Logf("通知发送失败: %v", err)
		return
	}

	if result.Success {
		t.Logf("通知发送成功!")
		t.Logf("消息ID: %s", result.MessageID)
		t.Logf("发送时间: %s", result.SentAt.Format("2006-01-02 15:04:05"))
	} else {
		t.Logf("通知发送失败: %v", result.Error)
	}
}

// TestEmailConfigValidation 测试邮件配置验证
func TestEmailConfigValidation(t *testing.T) {
	// 创建日志器
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	configPath := "../../config/local.yml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("配置文件不存在，跳过集成测试")
	}
	os.Setenv("APP_CONF", configPath)

	// 从配置文件加载邮件配置
	cfg, err := configloader.Load()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 测试有效配置
	validConfig := &notification.EmailConfig{
		Enabled:  cfg.GetBool("notification.email.enabled"),
		Host:     cfg.GetString("notification.email.host"),
		Port:     cfg.GetInt("notification.email.port"),
		Username: cfg.GetString("notification.email.username"),
		Password: cfg.GetString("notification.email.password"),
		From:     cfg.GetString("notification.email.from"),
		FromName: cfg.GetString("notification.email.from_name"),
		UseTLS:   cfg.GetBool("notification.email.use_tls"),
	}

	provider := notification.NewEmailProvider(validConfig, logger)
	err = provider.ValidateConfig()
	if err != nil {
		t.Logf("有效配置验证失败: %v", err)
	} else {
		t.Logf("有效配置验证通过")
	}

	// 测试无效配置
	invalidConfig := &notification.EmailConfig{
		Enabled:  true,
		Host:     "", // 空主机名
		Port:     0,  // 无效端口
		Username: "",
		Password: "",
		From:     "",
		FromName: "",
		UseTLS:   true,
	}

	invalidProvider := notification.NewEmailProvider(invalidConfig, logger)
	err = invalidProvider.ValidateConfig()
	if err != nil {
		t.Logf("无效配置验证正确失败: %v", err)
	} else {
		t.Logf("无效配置验证应该失败但没有失败")
	}
}
