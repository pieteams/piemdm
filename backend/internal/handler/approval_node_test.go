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

func createNodeTestHandler() *handler.Handler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	testLogger := &log.Logger{
		Logger: logger,
	}
	return handler.NewHandler(testLogger)
}

func TestApprovalNodeHandler_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApprovalNodeService := mock_service.NewMockApprovalNodeService(ctrl)

	expectedList := []*model.ApprovalNode{
		{
			ID:              1,
			ApprovalDefCode: "TEST_DEF_001",
			NodeCode:        "START_001",
			NodeName:        "开始节点",
			NodeType:        model.NodeTypeStart,
			SortOrder:       1,
		},
	}

	mockApprovalNodeService.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedList, nil)

	approvalNodeHandler := handler.NewApprovalNodeHandler(createNodeTestHandler(), mockApprovalNodeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/approval-nodes", approvalNodeHandler.List)

	req, _ := http.NewRequest("GET", "/approval-nodes?page=1&pageSize=10", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalNodeHandler_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApprovalNodeService := mock_service.NewMockApprovalNodeService(ctrl)

	expectedNode := &model.ApprovalNode{
		ID:              1,
		ApprovalDefCode: "TEST_DEF_001",
		NodeCode:        "START_001",
		NodeName:        "开始节点",
		NodeType:        model.NodeTypeStart,
		SortOrder:       1,
	}

	mockApprovalNodeService.EXPECT().Get(uint(1)).Return(expectedNode, nil)

	approvalNodeHandler := handler.NewApprovalNodeHandler(createNodeTestHandler(), mockApprovalNodeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/approval-nodes/:id", approvalNodeHandler.Get)

	req, _ := http.NewRequest("GET", "/approval-nodes/1", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalNodeHandler_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	params := request.CreateApprovalNodeRequest{
		ApprovalDefCode: "TEST_DEF_001",
		NodeCode:        "APPROVAL_001",
		NodeName:        "审批节点",
		NodeType:        "APPROVAL",
		ApproverType:    "USERS",
		SortOrder:       2,
		ApproverConfig:  `{"type":"USERS","users":["user001"]}`,
		ConditionConfig: `{}`,
		Status:          "Normal",
	}

	mockApprovalNodeService := mock_service.NewMockApprovalNodeService(ctrl)
	mockApprovalNodeService.EXPECT().Create(gomock.Any()).Return(nil)

	approvalNodeHandler := handler.NewApprovalNodeHandler(createNodeTestHandler(), mockApprovalNodeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/approval-nodes", approvalNodeHandler.Create)

	paramsJson, _ := json.Marshal(params)
	req, _ := http.NewRequest("POST", "/approval-nodes", bytes.NewBuffer(paramsJson))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalNodeHandler_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	params := request.UpdateApprovalNodeRequest{
		ID:              1,
		NodeName:        "审批节点-更新",
		NodeType:        "APPROVAL",
		ApproverType:    "USER",
		ApproverConfig:  `{"type":"USERS","users":["user001","user002"]}`,
		ApprovalMode:    "AND",
		TimeoutHours:    48,
		AllowReject:     true,
		AllowTransfer:   false,
		NotifyConfig:    `{"email":true,"sms":true}`,
		FormConfig:      `{"fields":[]}`,
		ConditionConfig: `{}`,
		Remark:          "测试审批节点-更新",
	}

	mockApprovalNodeService := mock_service.NewMockApprovalNodeService(ctrl)

	// Handler会先调用Get获取现有记录
	existingNode := &model.ApprovalNode{
		ID:              1,
		ApprovalDefCode: "TEST_DEF_001",
		NodeCode:        "APPROVAL_001",
		NodeName:        "审批节点",
		NodeType:        model.NodeTypeApproval,
		SortOrder:       2,
	}
	mockApprovalNodeService.EXPECT().Get(uint(1)).Return(existingNode, nil)
	mockApprovalNodeService.EXPECT().Update(gomock.Any()).Return(nil)

	approvalNodeHandler := handler.NewApprovalNodeHandler(createNodeTestHandler(), mockApprovalNodeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/approval-nodes/:id", approvalNodeHandler.Update)

	paramsJson, _ := json.Marshal(params)
	req, _ := http.NewRequest("PUT", "/approval-nodes/1", bytes.NewBuffer(paramsJson))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalNodeHandler_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApprovalNodeService := mock_service.NewMockApprovalNodeService(ctrl)
	mockApprovalNodeService.EXPECT().Delete(uint(1)).Return(&model.ApprovalNode{}, nil)

	approvalNodeHandler := handler.NewApprovalNodeHandler(createNodeTestHandler(), mockApprovalNodeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/approval-nodes/:id", approvalNodeHandler.Delete)

	req, _ := http.NewRequest("DELETE", "/approval-nodes/1", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalNodeHandler_GetNodesByApprovalDef(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedList := []*model.ApprovalNode{
		{
			ID:              1,
			ApprovalDefCode: "TEST_DEF_001",
			NodeCode:        "START_001",
			NodeName:        "开始节点",
			NodeType:        model.NodeTypeStart,
			SortOrder:       1,
		},
		{
			ID:              2,
			ApprovalDefCode: "TEST_DEF_001",
			NodeCode:        "APPROVAL_001",
			NodeName:        "审批节点",
			NodeType:        model.NodeTypeApproval,
			SortOrder:       2,
		},
	}

	mockApprovalNodeService := mock_service.NewMockApprovalNodeService(ctrl)
	mockApprovalNodeService.EXPECT().GetByApprovalDefCode("TEST_DEF_001").Return(expectedList, nil)

	approvalNodeHandler := handler.NewApprovalNodeHandler(createNodeTestHandler(), mockApprovalNodeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/approval-nodes/approval-def/:approvalDefCode", approvalNodeHandler.GetNodesByApprovalDef)

	req, _ := http.NewRequest("GET", "/approval-nodes/approval-def/TEST_DEF_001", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalNodeHandler_GetStartNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedNode := &model.ApprovalNode{
		ID:              1,
		ApprovalDefCode: "TEST_DEF_001",
		NodeCode:        "START_001",
		NodeName:        "开始节点",
		NodeType:        model.NodeTypeStart,
		SortOrder:       1,
	}

	mockApprovalNodeService := mock_service.NewMockApprovalNodeService(ctrl)
	mockApprovalNodeService.EXPECT().GetStartNode("TEST_DEF_001").Return(expectedNode, nil)

	approvalNodeHandler := handler.NewApprovalNodeHandler(createNodeTestHandler(), mockApprovalNodeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/approval-nodes/approval-def/:approvalDefCode/start", approvalNodeHandler.GetStartNode)

	req, _ := http.NewRequest("GET", "/approval-nodes/approval-def/TEST_DEF_001/start", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalNodeHandler_GetEndNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedList := []*model.ApprovalNode{
		{
			ID:              3,
			ApprovalDefCode: "TEST_DEF_001",
			NodeCode:        "END_001",
			NodeName:        "结束节点",
			NodeType:        model.NodeTypeEnd,
			SortOrder:       3,
		},
	}

	mockApprovalNodeService := mock_service.NewMockApprovalNodeService(ctrl)
	mockApprovalNodeService.EXPECT().GetEndNodes("TEST_DEF_001").Return(expectedList, nil)

	approvalNodeHandler := handler.NewApprovalNodeHandler(createNodeTestHandler(), mockApprovalNodeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/approval-nodes/approval-def/:approvalDefCode/end", approvalNodeHandler.GetEndNodes)

	req, _ := http.NewRequest("GET", "/approval-nodes/approval-def/TEST_DEF_001/end", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestApprovalNodeHandler_ValidateWorkflow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApprovalNodeService := mock_service.NewMockApprovalNodeService(ctrl)
	mockApprovalNodeService.EXPECT().ValidateWorkflow("TEST_DEF_001").Return(nil)

	approvalNodeHandler := handler.NewApprovalNodeHandler(createNodeTestHandler(), mockApprovalNodeService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/approval-nodes/approval-def/:approvalDefCode/validate", approvalNodeHandler.ValidateWorkflow)

	req, _ := http.NewRequest("POST", "/approval-nodes/approval-def/TEST_DEF_001/validate", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}
