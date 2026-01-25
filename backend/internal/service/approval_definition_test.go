package service_test

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"piemdm/internal/model"
	"piemdm/internal/service"
	"piemdm/pkg/helper/sid"
	"piemdm/pkg/jwt"
	"piemdm/pkg/log"
	mock_repository "piemdm/test/mocks/repository"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func setupApprovalDefService(t *testing.T) (service.ApprovalDefinitionService, *mock_repository.MockApprovalDefinitionRepository, *mock_repository.MockApprovalNodeRepository, *gomock.Controller) {
	ctrl := gomock.NewController(t)

	logger := &log.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	mockDefRepo := mock_repository.NewMockApprovalDefinitionRepository(ctrl)
	mockNodeRepo := mock_repository.NewMockApprovalNodeRepository(ctrl)

	baseService := service.NewService(logger, &sid.Sid{}, &jwt.JWT{})

	approvalDefService := service.NewApprovalDefinitionService(baseService, mockDefRepo, mockNodeRepo, nil)
	return approvalDefService, mockDefRepo, mockNodeRepo, ctrl
}

func TestApprovalDefService_Create(t *testing.T) {
	approvalDefService, mockDefRepo, _, ctrl := setupApprovalDefService(t)
	defer ctrl.Finish()

	now := time.Now()
	approvalDef := &model.ApprovalDefinition{
		ID:          1,
		Code:        "TEST_APPROVAL_001",
		Name:        "测试审批流程",
		Description: "测试审批流程描述",
		FormData:    `{"fields":[]}`,
		NodeList:    `{"nodes":[]}`,
		Status:      "Draft",
		CreatedBy:   "admin",
		UpdatedBy:   "admin",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mockDefRepo.EXPECT().Create(gomock.Any(), approvalDef).Return(nil)

	err := approvalDefService.Create(nil, approvalDef)
	assert.NoError(t, err)
}

func TestApprovalDefService_GetById(t *testing.T) {
	approvalDefService, mockDefRepo, _, ctrl := setupApprovalDefService(t)
	defer ctrl.Finish()

	id := uint(1)
	expectedApprovalDef := &model.ApprovalDefinition{
		ID:        1,
		Code:      "TEST_APPROVAL_001",
		Name:      "测试审批流程",
		Status:    "Normal",
		CreatedBy: "admin",
		UpdatedBy: "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockDefRepo.EXPECT().FindOne(id).Return(expectedApprovalDef, nil)

	approvalDef, err := approvalDefService.Get(uint(id))
	assert.NoError(t, err)
	assert.NotNil(t, approvalDef)
	assert.Equal(t, "TEST_APPROVAL_001", approvalDef.Code)
	assert.Equal(t, "测试审批流程", approvalDef.Name)
}

func TestApprovalDefService_GetByCode(t *testing.T) {
	approvalDefService, mockDefRepo, _, ctrl := setupApprovalDefService(t)
	defer ctrl.Finish()

	code := "TEST_APPROVAL_001"
	expectedApprovalDef := &model.ApprovalDefinition{
		ID:        1,
		Code:      "TEST_APPROVAL_001",
		Name:      "测试审批流程",
		Status:    "Normal",
		CreatedBy: "admin",
		UpdatedBy: "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockDefRepo.EXPECT().FirstByCode(code).Return(expectedApprovalDef, nil)

	approvalDef, err := approvalDefService.GetByCode(code)
	assert.NoError(t, err)
	assert.NotNil(t, approvalDef)
	assert.Equal(t, "TEST_APPROVAL_001", approvalDef.Code)
}

func TestApprovalDefService_Delete(t *testing.T) {
	approvalDefService, mockDefRepo, _, ctrl := setupApprovalDefService(t)
	defer ctrl.Finish()

	id := uint(1)
	approvalDef := &model.ApprovalDefinition{
		ID:     1,
		Code:   "TEST_APPROVAL_001",
		Status: "Draft",
	}

	// 第一次调用 FindOne 用于 CanDelete 检查
	mockDefRepo.EXPECT().FindOne(id).Return(approvalDef, nil)
	// 第二次调用 FindOne 用于获取删除前的记录
	mockDefRepo.EXPECT().FindOne(id).Return(approvalDef, nil)
	// 调用 BatchDelete 执行实际删除
	mockDefRepo.EXPECT().BatchDelete(gomock.Any(), gomock.Any()).Return(nil)

	deletedDef, err := approvalDefService.Delete(nil, id)
	assert.NoError(t, err)
	assert.NotNil(t, deletedDef)
}
