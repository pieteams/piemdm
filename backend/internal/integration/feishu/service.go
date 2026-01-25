package feishu

import (
	"context"
	"encoding/json"
	"fmt"

	"piemdm/internal/configs"
	"piemdm/pkg/log"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkapproval "github.com/larksuite/oapi-sdk-go/v3/service/approval/v4"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

// ApprovalCallback 审批回调处理函数定义
type ApprovalCallback func(ctx context.Context, externalInstID string, status string) error

// Service 飞书集成服务
type Service struct {
	client           *Client
	logger           *log.Logger
	wsClient         *larkws.Client
	approvalCallback ApprovalCallback
}

// NewService 创建飞书集成服务
func NewService(cfg configs.FeishuConfig, logger *log.Logger) *Service {
	client := NewClient(cfg)
	return &Service{
		client: client,
		logger: logger,
	}
}

// SetApprovalCallback 设置审批回调
func (s *Service) SetApprovalCallback(callback ApprovalCallback) {
	s.approvalCallback = callback
}

// GetConfig 获取配置
func (s *Service) GetConfig() configs.FeishuConfig {
	return s.client.GetConfig()
}

// StartEventLoop 启动长连接事件监听
func (s *Service) StartEventLoop(ctx context.Context) error {
	if !s.client.CheckConfig() {
		s.logger.Warn("Feishu config is disabled or invalid, skip event loop")
		return nil
	}

	// 使用 dispatcher.NewEventDispatcher 创建事件分发器
	eventHandler := dispatcher.NewEventDispatcher("", "").
		OnCustomizedEvent("approval_instance", func(ctx context.Context, event *larkevent.EventReq) error {

			// 解析事件内容
			var body struct {
				Event struct {
					InstanceCode string `json:"instance_code"`
					Status       string `json:"status"`
				} `json:"event"`
			}
			if err := json.Unmarshal(event.Body, &body); err != nil {
				s.logger.Error("Failed to unmarshal customized approval event", "error", err)
				return nil
			}

			// 如果有 code 和 status，触发回调
			if body.Event.InstanceCode != "" && body.Event.Status != "" {
				if s.approvalCallback != nil {
					return s.approvalCallback(ctx, body.Event.InstanceCode, body.Event.Status)
				}
			}

			return nil
		})

	// 监听审批实例状态变更事件
	eventHandler.OnP2ApprovalUpdatedV4(func(ctx context.Context, event *larkapproval.P2ApprovalUpdatedV4) error {

		// 检查事件内容是否为空
		if event.Event == nil || event.Event.Object == nil {
			return nil
		}

		obj := event.Event.Object

		// SDK v3.5.2 ApprovalEvent struct definition:
		// ApprovalCode *string (Likely Definition Code)
		// ApprovalId *string
		// Extra *string (Contains detailed event info including instance_code and status)

		instanceCode := ""
		status := "UNKNOWN"

		if obj.Extra != nil && *obj.Extra != "" {
			var extraMap map[string]interface{}
			if err := json.Unmarshal([]byte(*obj.Extra), &extraMap); err == nil {
				if v, ok := extraMap["instance_code"].(string); ok {
					instanceCode = v
				} else if v, ok := extraMap["approval_instance_code"].(string); ok {
					// Fallback check
					instanceCode = v
				}

				if v, ok := extraMap["status"].(string); ok {
					status = v
				}
			} else {
				s.logger.Error("Failed to parse Feishu event extra", "error", err, "extra", *obj.Extra)
			}
		}

		// Fallback: Use ApprovalCode as InstanceCode if not found in Extra (Dangerous assumption, but handled by logging)
		if instanceCode == "" && obj.ApprovalCode != nil {
			// In some contexts ApprovalCode might be InstanceCode, but usually it's Definition Code.
			// We log this potential issue.
			s.logger.Warn("InstanceCode not found in Extra, using object.ApprovalCode", "code", *obj.ApprovalCode)
			instanceCode = *obj.ApprovalCode
		}

		s.logger.Debug("Received Feishu approval event", "instance_code", instanceCode, "status", status)

		if instanceCode == "" {
			s.logger.Warn("Skipping Feishu event: instanceCode is empty")
			return nil
		}

		if s.approvalCallback == nil {
			s.logger.Warn("Skipping Feishu event: approvalCallback is nil")
			return nil
		}

		s.logger.Debug("Executing approval callback", "instanceCode", instanceCode, "status", status)
		err := s.approvalCallback(ctx, instanceCode, status)
		if err != nil {
			s.logger.Error("Feishu approval callback failed", "error", err, "instanceCode", instanceCode)
		} else {
			s.logger.Debug("Feishu approval callback executed successfully", "instanceCode", instanceCode)
		}
		return err
	})

	// 创建 WebSocket 客户端
	wsClient := larkws.NewClient(
		s.client.config.AppID,
		s.client.config.AppSecret,
		larkws.WithEventHandler(eventHandler),
		larkws.WithLogLevel(larkcore.LogLevelDebug),
	)

	s.wsClient = wsClient

	// 启动长连接
	go func() {
		s.logger.Info("Attempting to connect to Feishu WebSocket...", "app_id", s.client.config.AppID)
		err := wsClient.Start(ctx)
		if err != nil {
			s.logger.Error("Feishu WebSocket connection failed or closed", "error", err, "app_id", s.client.config.AppID)
		} else {
			s.logger.Info("Feishu WebSocket server stopped gracefully")
		}
	}()

	return nil
}

// StopEventLoop 停止长连接
func (s *Service) StopEventLoop() {
	// 目前 larkws 没有公开 Stop 方法, 依赖 context 取消或者让进程自然退出
	// 如果未来 SDK 支持 Stop，可以在这里调用
}

// GenerateFormValues 将 map 转换为飞书表单值 JSON 字符串，并根据 allowedFields 进行过滤和类型补充
func (s *Service) GenerateFormValues(data map[string]any, allowedFields map[string]string) (string, error) {
	// 飞书表单实例数据格式: [{"id": "widget_id", "type": "input", "value": "widget_value"}]
	var values []map[string]any

	for k, v := range data {
		// 排除一些系统字段
		if k == "id" || k == "approval_code" || k == "operation" || k == "action" || k == "entity_id" {
			continue
		}

		// 检查字段是否在允许列表中
		fieldType, allowed := allowedFields[k]
		if len(allowedFields) > 0 && !allowed {
			continue
		}

		// 转换值为字符串 (简单处理)
		// 注意: 某些控件类型可能需要特殊格式 (如 dateInterval), 这里暂且简化为 string
		valStr := fmt.Sprintf("%v", v)

		values = append(values, map[string]any{
			"id":    k,
			"type":  fieldType,
			"value": valStr,
		})
	}

	jsonBytes, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// ExtractFieldIDs 从飞书审批定义 JSON 中提取所有控件 ID 和类型
func (s *Service) ExtractFieldIDs(definitionJSON string) map[string]string {
	if definitionJSON == "" {
		s.logger.Warn("ExtractFieldIDs: definitionJSON is empty")
		return nil
	}

	// 1. 尝试解析完整定义格式: {"form": {"form_content": "[...]"}}
	var def struct {
		Form struct {
			FormContent string `json:"form_content"`
		} `json:"form"`
	}

	var formContent string
	if err := json.Unmarshal([]byte(definitionJSON), &def); err == nil && def.Form.FormContent != "" {
		formContent = def.Form.FormContent
	} else {
		// 2. 尝试直接作为 formContent 数组解析
		formContent = definitionJSON
	}

	var controls []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}

	if err := json.Unmarshal([]byte(formContent), &controls); err != nil {
		s.logger.Error("failed to unmarshal feishu form content", "error", err, "content", formContent)
		return nil
	}

	fieldMap := make(map[string]string)
	for _, c := range controls {
		if c.ID != "" {
			fieldMap[c.ID] = c.Type
		}
	}

	// s.logger.Debug("ExtractFieldIDs: success", "count", len(fieldMap), "fields", fieldMap)
	return fieldMap
}

// ExtractFirstTextareaID 从飞书审批定义 JSON 中提取第一个 textarea 控件的 ID
func (s *Service) ExtractFirstTextareaID(definitionJSON string) string {
	if definitionJSON == "" {
		return ""
	}

	// 1. 尝试解析完整定义格式
	var def struct {
		Form struct {
			FormContent string `json:"form_content"`
		} `json:"form"`
	}

	var formContent string
	if err := json.Unmarshal([]byte(definitionJSON), &def); err == nil && def.Form.FormContent != "" {
		formContent = def.Form.FormContent
	} else {
		formContent = definitionJSON
	}

	var controls []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}

	if err := json.Unmarshal([]byte(formContent), &controls); err != nil {
		s.logger.Error("ExtractFirstTextareaID: failed to unmarshal feishu form", "error", err)
		return ""
	}

	for _, c := range controls {
		if c.Type == "textarea" {
			return c.ID
		}
	}

	return ""
}

// SubscribeApproval 订阅审批事件
func (s *Service) SubscribeApproval(ctx context.Context, approvalCode string) error {
	// s.logger.Debug("Initiating Feishu approval subscription", "approvalCode", approvalCode)
	if !s.client.CheckConfig() {
		return fmt.Errorf("feishu config invalid")
	}

	// 1. 构建请求 (使用 SubscribeApprovalReq 而非 SubscribeInstanceReq)
	req := larkapproval.NewSubscribeApprovalReqBuilder().
		ApprovalCode(approvalCode).
		Build()

	// 2. 发起请求 (Subscribe 是在 Approval 资源上，而非 Instance)
	resp, err := s.client.GetLarkClient().Approval.Approval.Subscribe(ctx, req)
	if err != nil {
		return err
	}

	// 3. 处理响应
	if !resp.Success() {
		// 如果已经订阅过，API 可能会返回错误，我们需要根据具体 code 判断是否静默忽略
		// 为了简单，我们先记录日志
		// s.logger.Debug("Feishu approval subscription status", "code", resp.Code, "msg", resp.Msg, "approvalCode", approvalCode)
	} else {
		s.logger.Debug("Feishu approval subscription successful", "approvalCode", approvalCode)
	}

	return nil
}

// SubscribeAllApprovals 批量订阅审批事件
func (s *Service) SubscribeAllApprovals(ctx context.Context, codes []string) {
	if len(codes) == 0 {
		return
	}
	// s.logger.Debug("Proactively subscribing to Feishu approvals", "count", len(codes))
	for _, code := range codes {
		if err := s.SubscribeApproval(ctx, code); err != nil {
			s.logger.Error("Failed to proactively subscribe to Feishu approval", "code", code, "error", err)
		}
	}
}

// CreateApprovalInstance 创建审批实例
func (s *Service) CreateApprovalInstance(ctx context.Context, approvalDefCode, formContent, userID, title string) (string, error) {
	if !s.client.CheckConfig() {
		return "", fmt.Errorf("feishu config invalid")
	}

	// 0. 自动订阅 (确保能收到事件)
	_ = s.SubscribeApproval(ctx, approvalDefCode)

	// 1. 构建请求
	req := larkapproval.NewCreateInstanceReqBuilder().
		InstanceCreate(larkapproval.NewInstanceCreateBuilder().
			ApprovalCode(approvalDefCode).
			UserId(userID).
			Form(formContent).
			Title(title).
			Build()).
		Build()

	// 2. 发起请求
	resp, err := s.client.GetLarkClient().Approval.Instance.Create(ctx, req)
	if err != nil {
		return "", err
	}

	// 3. 处理响应
	if !resp.Success() {
		return "", fmt.Errorf("create feishu instance failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return *resp.Data.InstanceCode, nil
}

// GetApprovalDefinition 获取审批定义详情 (包括表单结构)
func (s *Service) GetApprovalDefinition(ctx context.Context, approvalCode string) (string, error) {
	if !s.client.CheckConfig() {
		return "", fmt.Errorf("feishu config invalid")
	}

	// 1. 构建请求
	req := larkapproval.NewGetApprovalReqBuilder().
		ApprovalCode(approvalCode).
		Build()

	// 2. 发起请求
	resp, err := s.client.GetLarkClient().Approval.Approval.Get(ctx, req)
	if err != nil {
		return "", err
	}

	// 3. 处理响应
	if !resp.Success() {
		return "", fmt.Errorf("get feishu approval definition failed: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	// 提取表单定义
	// 不同的 SDK 版本结构可能不同，通常在 resp.Data.Form 或者 resp.Data.WidgetList
	// 查看 SDK 源码或文档: Response Data has Form(string) or WidgetList
	// 假设 SDK struct:
	/*
		type GetApprovalRespData struct {
			ApprovalName *string `json:"approval_name,omitempty"`
			Status *string `json:"status,omitempty"`
			Form *string `json:"form,omitempty"`  // This is usually the serialized JSON of widget list or form config
			NodeList *[]*Node `json:"node_list,omitempty"`
			...
		}
	*/

	// 如果 form 是 string (json)
	if resp.Data.Form != nil {
		return *resp.Data.Form, nil
	}

	// 如果没有 Form 只有 WidgetList，手动序列化
	// (SDK V4 might organize it differently)
	// For now, assume Form is populated as it is standard in V4 Get Approval

	return "", fmt.Errorf("no form content found in feishu approval definition")
}
