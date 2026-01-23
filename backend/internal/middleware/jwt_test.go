package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"piemdm/pkg/jwt"
	"piemdm/pkg/log"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, *jwt.JWT, *log.Logger) {
	// 创建测试配置
	v := viper.New()
	v.Set("security.jwt.key", "test-secret-key-for-middleware-testing")

	// 创建JWT实例
	j := jwt.NewJwt(v)

	// 创建日志实例 - 使用slog的默认logger
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // 测试中只记录错误级别
	}))
	logger := &log.Logger{Logger: slogLogger}

	// 创建Gin引擎
	gin.SetMode(gin.TestMode)
	r := gin.New()

	return r, j, logger
}

func TestStrictAuth(t *testing.T) {
	r, j, logger := setupTestRouter()

	// 生成测试令牌
	token, err := j.GenToken("123", "testuser", "test@example.com", "Yes")
	require.NoError(t, err)

	// 测试路由
	r.GET("/protected", StrictAuth(j, logger), func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, "123", userID)

		userName, exists := c.Get("user_name")
		assert.True(t, exists)
		assert.Equal(t, "testuser", userName)

		admin, exists := c.Get("admin")
		assert.True(t, exists)
		assert.Equal(t, "Yes", admin)

		c.JSON(http.StatusOK, gin.H{"message": "authorized"})
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectAuth     bool
	}{
		{
			name:           "有效令牌授权",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			expectAuth:     true,
		},
		{
			name:           "无Authorization头",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectAuth:     false,
		},
		{
			name:           "无效令牌格式",
			authHeader:     "Bearer invalid.token.string",
			expectedStatus: http.StatusUnauthorized,
			expectAuth:     false,
		},
		{
			name:           "空Bearer令牌",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectAuth:     false,
		},
		{
			name:           "错误前缀",
			authHeader:     "Token " + token,
			expectedStatus: http.StatusUnauthorized,
			expectAuth:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestNoStrictAuth(t *testing.T) {
	r, j, logger := setupTestRouter()

	// 生成测试令牌
	token, err := j.GenToken("456", "nostrict", "nostrict@example.com", "No")
	require.NoError(t, err)

	// 测试路由
	callCount := 0
	r.GET("/optional", NoStrictAuth(j, logger), func(c *gin.Context) {
		callCount++
		userID, exists := c.Get("user_id")
		if exists {
			assert.Equal(t, "456", userID)
			c.JSON(http.StatusOK, gin.H{"message": "authenticated"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "anonymous"})
		}
	})

	tests := []struct {
		name           string
		authHeader     string
		cookieToken    string
		queryToken     string
		expectedStatus int
		expectAuth     bool
	}{
		{
			name:           "Authorization头有效令牌",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			expectAuth:     true,
		},
		{
			name:           "Cookie有效令牌",
			cookieToken:    token,
			expectedStatus: http.StatusOK,
			expectAuth:     true,
		},
		{
			name:           "Query参数有效令牌",
			queryToken:     token,
			expectedStatus: http.StatusOK,
			expectAuth:     true,
		},
		{
			name:           "无任何令牌",
			expectedStatus: http.StatusOK,
			expectAuth:     false,
		},
		{
			name:           "无效令牌",
			authHeader:     "Bearer invalid.token",
			expectedStatus: http.StatusOK,
			expectAuth:     false,
		},
		{
			name:           "多种令牌源，Authorization优先",
			authHeader:     "Bearer " + token,
			cookieToken:    "different.token",
			queryToken:     "another.token",
			expectedStatus: http.StatusOK,
			expectAuth:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount = 0
			req := httptest.NewRequest("GET", "/optional", nil)

			// 设置Authorization头
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// 设置Cookie
			if tt.cookieToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  "accessToken",
					Value: tt.cookieToken,
				})
			}

			// 设置Query参数
			if tt.queryToken != "" {
				q := req.URL.Query()
				q.Set("accessToken", tt.queryToken)
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, 1, callCount)
		})
	}
}

func TestAdminAuth(t *testing.T) {
	r, j, logger := setupTestRouter()

	// 生成管理员令牌
	adminToken, err := j.GenToken("1", "admin", "admin@example.com", "Yes")
	require.NoError(t, err)

	// 生成普通用户令牌
	userToken, err := j.GenToken("2", "user", "user@example.com", "No")
	require.NoError(t, err)

	// 测试路由
	r.GET("/admin-only", AdminAuth(j, logger), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "管理员令牌访问",
			authHeader:     "Bearer " + adminToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "普通用户令牌访问",
			authHeader:     "Bearer " + userToken,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "无令牌访问",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "无效令牌访问",
			authHeader:     "Bearer invalid.token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/admin-only", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUserAuth(t *testing.T) {
	r, j, logger := setupTestRouter()

	// 生成管理员令牌
	adminToken, err := j.GenToken("1", "admin", "admin@example.com", "Yes")
	require.NoError(t, err)

	// 生成普通用户令牌
	userToken, err := j.GenToken("2", "user", "user@example.com", "No")
	require.NoError(t, err)

	// 测试路由
	r.GET("/user-access", UserAuth(j, logger), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "user access granted"})
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "管理员令牌访问",
			authHeader:     "Bearer " + adminToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "普通用户令牌访问",
			authHeader:     "Bearer " + userToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "无令牌访问",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "无效令牌访问",
			authHeader:     "Bearer invalid.token",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/user-access", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestMiddlewareChain(t *testing.T) {
	r, j, logger := setupTestRouter()

	// 生成测试令牌
	token, err := j.GenToken("999", "chaintest", "chain@example.com", "Yes")
	require.NoError(t, err)

	// 测试多个中间件链式调用
	callOrder := []string{}
	r.GET("/chain",
		func(c *gin.Context) {
			callOrder = append(callOrder, "middleware1")
			c.Next()
		},
		StrictAuth(j, logger),
		func(c *gin.Context) {
			callOrder = append(callOrder, "middleware2")
			c.Next()
		},
		func(c *gin.Context) {
			callOrder = append(callOrder, "handler")
			c.JSON(http.StatusOK, gin.H{"message": "chain completed"})
		},
	)

	req := httptest.NewRequest("GET", "/chain", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{"middleware1", "middleware2", "handler"}, callOrder)
}

// 删除MockLogger定义，直接使用log.Logger

func TestRecoveryLoggerFunc(t *testing.T) {
	r, j, logger := setupTestRouter()

	// 生成测试令牌
	token, err := j.GenToken("777", "logtest", "log@example.com", "No")
	require.NoError(t, err)

	r.GET("/log-context", StrictAuth(j, logger), func(c *gin.Context) {
		// 验证上下文已设置
		claims, exists := c.Get("claims")
		assert.True(t, exists)
		assert.IsType(t, &jwt.CustomClaims{}, claims)

		userInfo := claims.(*jwt.CustomClaims)
		assert.Equal(t, "777", userInfo.ID)
		assert.Equal(t, "logtest", userInfo.UserName)
		assert.Equal(t, "No", userInfo.Admin)

		c.JSON(http.StatusOK, gin.H{"message": "context verified"})
	})

	req := httptest.NewRequest("GET", "/log-context", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddlewareContextValues(t *testing.T) {
	r, j, logger := setupTestRouter()

	// 生成测试令牌
	token, err := j.GenToken("888", "contextuser", "context@example.com", "Yes")
	require.NoError(t, err)

	r.GET("/context-values", StrictAuth(j, logger), func(c *gin.Context) {
		// 验证所有上下文值
		claims, exists := c.Get("claims")
		assert.True(t, exists)
		userClaims := claims.(*jwt.CustomClaims)

		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, userClaims.ID, userID)

		userName, exists := c.Get("user_name")
		assert.True(t, exists)
		assert.Equal(t, userClaims.UserName, userName)

		admin, exists := c.Get("admin")
		assert.True(t, exists)
		assert.Equal(t, userClaims.Admin, admin)

		c.JSON(http.StatusOK, gin.H{
			"user_id":   userID,
			"user_name": userName,
			"admin":     admin,
			"email":     userClaims.Email,
		})
	})

	req := httptest.NewRequest("GET", "/context-values", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddlewareErrorHandling(t *testing.T) {
	r, j, logger := setupTestRouter()

	// 测试各种错误场景
	r.GET("/error-test", StrictAuth(j, logger), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "过期令牌",
			authHeader:     "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Nzg5MDQwMDB9.invalid-expired-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "格式错误令牌",
			authHeader:     "Bearer not.a.valid.jwt.token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "使用错误密钥签名的令牌",
			authHeader:     "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/error-test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
