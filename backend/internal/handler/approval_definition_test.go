package handler_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"piemdm/internal/handler"
	"piemdm/internal/model"
	"piemdm/internal/transport/request"
	"piemdm/pkg/log"
	mock_service "piemdm/test/mocks/service"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func createTestHandler() *handler.Handler {
	// 创建一个简单的 logger 用于测试
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	testLogger := &log.Logger{
		Logger: logger,
	}

	return handler.NewHandler(testLogger)
}

func TestApprovalDefHandler_ListApprovalDefs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApprovalDefService := mock_service.NewMockApprovalDefinitionService(ctrl)

	expectedList := []*model.ApprovalDefinition{
		{
			ID:     1,
			Code:   "TEST_DEF_001",
			Name:   "测试审批定义",
			Status: "Normal",
		},
	}

	mockApprovalDefService.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedList, nil)

	approvalDefHandler := handler.NewApprovalDefinitionHandler(createTestHandler(), mockApprovalDefService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/approval-defs", approvalDefHandler.List)

	req, _ := http.NewRequest("GET", "/approval-defs?page=1&pageSize=10", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalDefHandler_GetApprovalDef(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApprovalDefService := mock_service.NewMockApprovalDefinitionService(ctrl)

	expectedApprovalDef := &model.ApprovalDefinition{
		ID:     1,
		Code:   "TEST_DEF_001",
		Name:   "测试审批定义",
		Status: "Normal",
	}

	mockApprovalDefService.EXPECT().Get(uint(1)).Return(expectedApprovalDef, nil)

	approvalDefHandler := handler.NewApprovalDefinitionHandler(createTestHandler(), mockApprovalDefService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/approval-defs/:id", approvalDefHandler.Get)

	req, _ := http.NewRequest("GET", "/approval-defs/1", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalDefHandler_CreateApprovalDef(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	params := request.CreateApprovalDefRequest{
		Name:        "测试审批定义",
		Description: "测试审批定义描述",
		FormData:    `{"fields":[]}`,
		NodeList:    `{"nodes":[]}`,
		Platform:    "Builtin",
		Status:      "Normal",
	}

	mockApprovalDefService := mock_service.NewMockApprovalDefinitionService(ctrl)
	mockApprovalDefService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	approvalDefHandler := handler.NewApprovalDefinitionHandler(createTestHandler(), mockApprovalDefService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/approval-defs", approvalDefHandler.Create)

	paramsJson, _ := json.Marshal(params)
	req, _ := http.NewRequest("POST", "/approval-defs", bytes.NewBuffer(paramsJson))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalDefHandler_UpdateApprovalDef(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	params := request.UpdateApprovalDefRequest{
		ID:          1,
		Name:        "测试审批定义-更新",
		Description: "测试审批定义描述-更新",
		Platform:    "Builtin",
		FormData:    `{"fields":[]}`,
		NodeList:    `{"nodes":[]}`,
		Status:      "Normal",
	}

	mockApprovalDefService := mock_service.NewMockApprovalDefinitionService(ctrl)

	// Handler会先调用GetById获取现有记录
	existingApprovalDef := &model.ApprovalDefinition{
		ID:     1,
		Code:   "TEST_DEF_001",
		Name:   "测试审批定义",
		Status: "Normal",
	}
	mockApprovalDefService.EXPECT().Get(uint(1)).Return(existingApprovalDef, nil)
	mockApprovalDefService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	approvalDefHandler := handler.NewApprovalDefinitionHandler(createTestHandler(), mockApprovalDefService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/approval-defs/:id", approvalDefHandler.Update)

	paramsJson, _ := json.Marshal(params)
	req, _ := http.NewRequest("PUT", "/approval-defs/1", bytes.NewBuffer(paramsJson))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalDefHandler_DeleteApprovalDef(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApprovalDefService := mock_service.NewMockApprovalDefinitionService(ctrl)
	mockApprovalDefService.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(&model.ApprovalDefinition{}, nil)

	approvalDefHandler := handler.NewApprovalDefinitionHandler(createTestHandler(), mockApprovalDefService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/approval-defs/:id", approvalDefHandler.Delete)

	req, _ := http.NewRequest("DELETE", "/approval-defs/1", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalDefHandler_GetApprovalDefByCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedApprovalDef := &model.ApprovalDefinition{
		ID:     1,
		Code:   "TEST_DEF_001",
		Name:   "测试审批定义",
		Status: "Normal",
	}

	mockApprovalDefService := mock_service.NewMockApprovalDefinitionService(ctrl)
	mockApprovalDefService.EXPECT().GetByCode("TEST_DEF_001").Return(expectedApprovalDef, nil)

	approvalDefHandler := handler.NewApprovalDefinitionHandler(createTestHandler(), mockApprovalDefService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/approval-defs/code/:code", approvalDefHandler.GetByCode)

	req, _ := http.NewRequest("GET", "/approval-defs/code/TEST_DEF_001", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalDefHandler_GetApprovalDefsByEntity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedList := []*model.ApprovalDefinition{
		{
			ID:     1,
			Code:   "TEST_DEF_001",
			Name:   "测试审批定义",
			Status: "Normal",
		},
	}

	mockApprovalDefService := mock_service.NewMockApprovalDefinitionService(ctrl)
	mockApprovalDefService.EXPECT().GetByEntityCode("test_entity").Return(expectedList, nil)

	approvalDefHandler := handler.NewApprovalDefinitionHandler(createTestHandler(), mockApprovalDefService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/approval-defs/entity/:entityCode", approvalDefHandler.GetByEntity)

	req, _ := http.NewRequest("GET", "/approval-defs/entity/test_entity", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalDefHandler_ActivateApprovalDef(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApprovalDefService := mock_service.NewMockApprovalDefinitionService(ctrl)
	mockApprovalDefService.EXPECT().Activate(gomock.Any()).Return(nil)

	approvalDefHandler := handler.NewApprovalDefinitionHandler(createTestHandler(), mockApprovalDefService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/approval-defs/:id/activate", approvalDefHandler.Activate)

	req, _ := http.NewRequest("POST", "/approval-defs/1/activate", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalDefHandler_DeactivateApprovalDef(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApprovalDefService := mock_service.NewMockApprovalDefinitionService(ctrl)
	mockApprovalDefService.EXPECT().Deactivate(gomock.Any()).Return(nil)

	approvalDefHandler := handler.NewApprovalDefinitionHandler(createTestHandler(), mockApprovalDefService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/approval-defs/:id/deactivate", approvalDefHandler.Deactivate)

	req, _ := http.NewRequest("POST", "/approval-defs/1/deactivate", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalDefHandler_ValidateApprovalDef(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApprovalDefService := mock_service.NewMockApprovalDefinitionService(ctrl)

	// 先获取审批定义
	expectedApprovalDef := &model.ApprovalDefinition{
		ID:     1,
		Code:   "TEST_DEF_001",
		Name:   "测试审批定义",
		Status: "Normal",
	}
	mockApprovalDefService.EXPECT().Get(uint(1)).Return(expectedApprovalDef, nil)
	mockApprovalDefService.EXPECT().ValidateDefinition(expectedApprovalDef).Return(nil)

	approvalDefHandler := handler.NewApprovalDefinitionHandler(createTestHandler(), mockApprovalDefService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/approval-defs/:id/validate", approvalDefHandler.Validate)

	req, _ := http.NewRequest("POST", "/approval-defs/1/validate", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}
