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
	jwt2 "piemdm/pkg/jwt"
	"piemdm/pkg/log"
	mock_service "piemdm/test/mocks/service"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func createUserTestHandler() *handler.Handler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	testLogger := &log.Logger{
		Logger: logger,
	}
	return handler.NewHandler(testLogger)
}

func TestUserHandler_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	params := model.User{
		EmployeeID:  "EMP001",
		Username:    "testuser",
		Password:    "123456",
		FirstName:   "Test",
		LastName:    "User",
		DisplayName: "Test User",
		Email:       "test@example.com",
		Phone:       "1234567890",
		Language:    "zh-cn",
		Sex:         "Male",
		Avatar:      "avatar.jpg",
		Description: "Test user description",
		Admin:       "No",
		Status:      "Normal",
	}

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockRoleService := mock_service.NewMockRoleService(ctrl)
	mockUserService.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	// 使用nil JWT，因为Register方法不需要JWT
	userHandler := handler.NewUserHandler(createUserTestHandler(), mockUserService, mockRoleService, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/users", userHandler.Create)

	paramsJson, _ := json.Marshal(params)
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(paramsJson))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestUserHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	params := request.LoginRequest{
		Username: "testuser",
		Password: "123456",
	}

	token := "test_token"

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockRoleService := mock_service.NewMockRoleService(ctrl)
	mockUserService.EXPECT().Login(gomock.Any(), &params).Return(token, &model.User{ID: 1, Username: "testuser"}, nil)

	// 使用nil JWT，因为Login方法中的JWT操作会被mock
	userHandler := handler.NewUserHandler(createUserTestHandler(), mockUserService, mockRoleService, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", userHandler.Login)

	paramsJson, _ := json.Marshal(params)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(paramsJson))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestUserHandler_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedUser := &model.User{
		ID:          1,
		Username:    "testuser",
		DisplayName: "Test User",
		Email:       "test@example.com",
	}

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockRoleService := mock_service.NewMockRoleService(ctrl)
	mockUserService.EXPECT().Get(uint(1)).Return(expectedUser, nil)

	userHandler := handler.NewUserHandler(createUserTestHandler(), mockUserService, mockRoleService, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	// 设置claims上下文中间件，模拟JWT认证
	router.Use(func(c *gin.Context) {
		// 模拟JWT claims
		claims := &jwt2.CustomClaims{
			ID: "1",
		}
		c.Set("claims", claims)
		c.Next()
	})
	router.GET("/user/:id", userHandler.Get)

	req, _ := http.NewRequest("GET", "/user/1", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	params := model.User{
		ID:          1,
		EmployeeID:  "EMP001",
		Username:    "testuser",
		Password:    "123456",
		FirstName:   "Updated",
		LastName:    "User",
		DisplayName: "Test User",
		Email:       "updated@example.com",
		Phone:       "1234567890",
		Language:    "zh-cn",
		Sex:         "Male",
		Avatar:      "new_avatar.jpg",
		Description: "Updated user description",
		Admin:       "No",
		Status:      "Normal",
	}

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockRoleService := mock_service.NewMockRoleService(ctrl)
	mockUserService.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	userHandler := handler.NewUserHandler(createUserTestHandler(), mockUserService, mockRoleService, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	// 设置claims上下文中间件，模拟JWT认证
	router.Use(func(c *gin.Context) {
		// 模拟JWT claims
		claims := &jwt2.CustomClaims{
			ID: "1",
		}
		c.Set("claims", claims)
		c.Next()
	})
	router.PUT("/user/:id", userHandler.Update)

	paramsJson, _ := json.Marshal(params)
	req, _ := http.NewRequest("PUT", "/user/1", bytes.NewBuffer(paramsJson))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestUserHandler_GetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedUser := &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	mockUserService := mock_service.NewMockUserService(ctrl)
	mockRoleService := mock_service.NewMockRoleService(ctrl)
	mockUserService.EXPECT().Get(uint(1)).Return(expectedUser, nil)

	userHandler := handler.NewUserHandler(createUserTestHandler(), mockUserService, mockRoleService, nil)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/users/:id", userHandler.Get)

	req, _ := http.NewRequest("GET", "/users/1", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}
