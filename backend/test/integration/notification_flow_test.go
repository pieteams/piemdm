package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"piemdm/internal/model"
	"piemdm/internal/repository"
	"piemdm/internal/service"
	"piemdm/pkg/configloader"
	"piemdm/pkg/log"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type NotificationFlowTestSuite struct {
	suite.Suite
	repo   *repository.Repository
	gormDB *gorm.DB
	v      *viper.Viper
	logger *log.Logger

	notificationService service.NotificationService
	templateRepo        repository.NotificationTemplateRepository
	logRepo             repository.NotificationLogRepository
}

func (suite *NotificationFlowTestSuite) SetupSuite() {
	// 1. 设置配置文件路径 (参考 openapi_test.go)
	configPath := "../../config/local.yml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		suite.T().Skip("配置文件不存在，跳过集成测试")
	}
	os.Setenv("APP_CONF", configPath)
	suite.v, _ = configloader.Load()

	suite.logger = log.NewLog(suite.v)

	// 2. 初始化数据库连接
	db := repository.NewDB(suite.v)
	rdb := repository.NewRedis(suite.v)
	suite.gormDB = db
	suite.repo = repository.NewRepository(db, rdb, suite.logger)

	// 3. 运行自动迁移（确保表结构最新，特别是 Status 字段）
	err := db.AutoMigrate(&model.NotificationTemplate{}, &model.NotificationLog{})
	require.NoError(suite.T(), err, "自动迁移失败")

	// 4. 强制初始化测试数据
	// 使用 Assign 确保即使数据库里已有该 Code 的记录，也会被强制更新为测试所需的状态
	testTemplate := &model.NotificationTemplate{
		TemplateCode:     "test_flow_template",
		TemplateName:     "集成测试全链路通知模板",
		TemplateType:     model.TemplateTypeApprovalStart,
		NotificationType: "email",
		TitleTemplate:    "测试通知: {{.title}}",
		ContentTemplate:  "这是一条来自 {{.sender}} 的测试通知。",
		Status:           "Normal",
		CreatedBy:        "test-suite",
	}

	var existing model.NotificationTemplate
	err = db.Where("template_code = ?", testTemplate.TemplateCode).First(&existing).Error
	if err != nil {
		// 如果找不到则创建
		testTemplate.ID = "test_flow_template" // 赋予固定 ID 或由 GORM 生成
		err = db.Create(testTemplate).Error
	} else {
		// 如果找到了则确保状态和模板内容被修正（使用指针传递给 Updates）
		err = db.Model(&existing).Updates(&model.NotificationTemplate{
			Status:          "Normal",
			TitleTemplate:   testTemplate.TitleTemplate,
			ContentTemplate: testTemplate.ContentTemplate,
		}).Error
	}
	require.NoError(suite.T(), err, "初始化测试模板数据失败")

	// 5. 初始化相关组件
	suite.templateRepo = repository.NewNotificationTemplateRepository(suite.gormDB)
	suite.logRepo = repository.NewNotificationLogRepository(suite.gormDB)

	templateService := service.NewNotificationTemplateService(suite.templateRepo, suite.logger)
	suite.notificationService = service.NewNotificationService(templateService, suite.logRepo, suite.logger)

	// 6. 注册一个模拟 Channel 模拟真实发送
	mockChannel := service.NewEmailChannel(service.EmailConfig{
		Enabled:   true,
		SMTPHost:  "mock.smtp.com",
		SMTPPort:  25,
		FromEmail: "test@example.com",
	}, suite.logger)
	err = suite.notificationService.RegisterChannel(mockChannel)
	require.NoError(suite.T(), err, "注册 Mock Channel 失败")
}

func (suite *NotificationFlowTestSuite) TestSendByTemplateAndVerifyLog() {
	t := suite.T()
	ctx := context.Background()

	req := &service.TemplateNotificationRequest{
		TemplateCode:     "test_flow_template",
		NotificationType: "email",
		RecipientID:      "user_001",
		RecipientType:    "user",
		Variables: map[string]any{
			"title":  "全链路测试报告",
			"sender": "Antigravity AI",
		},
		ApprovalID: "APP-I-001",
		TaskID:     "TSK-I-001",
	}

	// 1. 执行发送
	err := suite.notificationService.SendByTemplate(ctx, req)
	require.NoError(t, err)

	// 2. 稍微等待异步或事务完成（如果是同步写入则不需要）
	// 2. 稍微等待异步或事务完成
	time.Sleep(500 * time.Millisecond)

	// 3. 从数据库校验日志记录
	var logs []model.NotificationLog
	// 直接通过 DB 查询，不经过 repo 以验证真实存储
	// 注意：这里使用 ID 倒序，确保查到最新的一条
	suite.gormDB.Where("approval_id = ?", req.ApprovalID).Order("id desc").Find(&logs)

	require.GreaterOrEqual(t, len(logs), 1, "数据库中应存有一条通知日志")

	logRecord := logs[0]
	assert.Equal(t, req.TemplateCode, logRecord.TemplateCode)
	assert.Contains(t, logRecord.Title, "测试通知: 全链路测试报告")
	assert.Contains(t, logRecord.Content, "这是一条来自 Antigravity AI 的测试通知")
	assert.Equal(t, model.NotificationStatusSent, logRecord.Status)
}

func TestNotificationFlowSuite(t *testing.T) {
	suite.Run(t, new(NotificationFlowTestSuite))
}
