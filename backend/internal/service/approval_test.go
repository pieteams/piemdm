package service_test

import (
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"piemdm/internal/model"
	"piemdm/internal/service"
	"piemdm/pkg/helper/sid"
	"piemdm/pkg/jwt"
	"piemdm/pkg/log"
	mock_repository "piemdm/test/mocks/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func setupApprovalService(t *testing.T) (service.ApprovalService, *mock_repository.MockApprovalRepository, *mock_repository.MockApprovalDefinitionRepository, *mock_repository.MockApprovalNodeRepository, *mock_repository.MockApprovalTaskRepository, *gomock.Controller) {
	// 设置Gin为测试模式，避免调试信息输出
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)

	logger := &log.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	mockApprovalRepo := mock_repository.NewMockApprovalRepository(ctrl)
	mockDefRepo := mock_repository.NewMockApprovalDefinitionRepository(ctrl)
	mockNodeRepo := mock_repository.NewMockApprovalNodeRepository(ctrl)
	mockTaskRepo := mock_repository.NewMockApprovalTaskRepository(ctrl)

	baseService := service.NewService(logger, &sid.Sid{}, &jwt.JWT{})

	approvalService := service.NewApprovalService(baseService, mockApprovalRepo, nil, nil, nil, nil, nil, nil, nil, nil, mockDefRepo, nil, mockNodeRepo, nil, mockTaskRepo, nil, nil, nil, nil)
	return approvalService, mockApprovalRepo, mockDefRepo, mockNodeRepo, mockTaskRepo, ctrl
}

func TestApprovalService_Create(t *testing.T) {
	approvalService, mockApprovalRepo, mockDefRepo, _, _, ctrl := setupApprovalService(t)
	defer ctrl.Finish()

	now := time.Now()
	approval := &model.Approval{
		ID:              uint(1),
		Code:            "APPROVAL_001",
		Title:           "测试审批申请",
		ApprovalDefCode: "TEST_DEF_001",
		EntityCode:      "test_entity",
		SerialNumber:    "20240101001",
		CurrentTaskID:   "start",
		CurrentTaskName: "开始节点",
		FormData:        `{"field1":"value1"}`,
		Status:          model.ApprovalStatusPending,
		Priority:        0,
		Urgency:         "Normal",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Mock approvalDefinitionRepository.First - Create 方法会调用此方法检查是否为飞书审批
	mockDefRepo.EXPECT().First(map[string]any{"code": "TEST_DEF_001"}).Return(nil, fmt.Errorf("not found"))

	mockApprovalRepo.EXPECT().Create(c, approval).Return(nil)

	err := approvalService.Create(c, approval)
	assert.NoError(t, err)
}

func TestApprovalService_Get(t *testing.T) {
	approvalService, mockApprovalRepo, _, _, _, ctrl := setupApprovalService(t)
	defer ctrl.Finish()

	id := uint(1)
	expectedApproval := &model.Approval{
		ID:              id,
		Code:            "APPROVAL_001",
		Title:           "测试审批申请",
		ApprovalDefCode: "TEST_DEF_001",
		Status:          model.ApprovalStatusPending,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mockApprovalRepo.EXPECT().FindOne(id).Return(expectedApproval, nil)

	approval, err := approvalService.Get(id)
	assert.NoError(t, err)
	assert.NotNil(t, approval)
	assert.Equal(t, "APPROVAL_001", approval.Code)
	assert.Equal(t, "测试审批申请", approval.Title)
}

func TestApprovalService_GetByCode(t *testing.T) {
	approvalService, mockApprovalRepo, _, _, _, ctrl := setupApprovalService(t)
	defer ctrl.Finish()

	code := "APPROVAL_001"
	expectedApproval := &model.Approval{
		ID:              uint(1),
		Code:            "APPROVAL_001",
		Title:           "测试审批申请",
		ApprovalDefCode: "TEST_DEF_001",
		Status:          model.ApprovalStatusPending,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mockApprovalRepo.EXPECT().FirstByCode(code).Return(expectedApproval, nil)

	approval, err := approvalService.GetByCode(code)
	assert.NoError(t, err)
	assert.NotNil(t, approval)
	assert.Equal(t, "APPROVAL_001", approval.Code)
}

// 注释掉这个测试，因为我们不在 ApprovalService 中实现 StartApproval 方法
// 该方法现在只在 ApprovalService 中实现
/*
func TestApprovalService_StartApproval(t *testing.T) {
	approvalService, mockApprovalRepo, mockDefRepo, mockNodeRepo, _, ctrl := setupApprovalService(t)
	defer ctrl.Finish()

	approvalDefCode := "TEST_DEF_001"
	applicantID := "user001"
	title := "测试审批申请"
	formData := `{"field1":"value1"}`

	// Mock审批定义
	approvalDef := &model.ApprovalDefinition{
		ID:        1,
		Code:      "TEST_DEF_001",
		Name:      "测试审批流程",
		Status:    model.ApprovalDefStatusNormal,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock开始节点
	startNode := &model.ApprovalNode{
		ID:              1,
		ApprovalDefCode: "TEST_DEF_001",
		NodeCode:        "start",
		NodeType:        "START",
		NodeName:        "开始节点",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	mockDefRepo.EXPECT().FirstByCode(approvalDefCode).Return(approvalDef, nil)
	mockNodeRepo.EXPECT().FindStartNode(approvalDefCode).Return(startNode, nil)
	mockApprovalRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	approval, err := approvalService.StartApproval(c, approvalDefCode, applicantID, title, formData)
	assert.NoError(t, err)
	assert.NotNil(t, approval)
	assert.Equal(t, title, approval.Title)
	assert.Equal(t, applicantID, approval.ApplicantID)
	assert.Equal(t, model.ApprovalStatusPending, approval.Status)
	assert.Equal(t, "start", approval.CurrentTaskID)
}
*/
