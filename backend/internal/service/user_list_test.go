package service_test

import (
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"

	"piemdm/internal/model"
	"piemdm/internal/transport/request"
	"piemdm/internal/service"
	"piemdm/pkg/helper/sid"
	"piemdm/pkg/log"
	mock_repository "piemdm/test/mocks/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUserService_GetUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 创建 mock repository
	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRoleRepo := mock_repository.NewMockUserRoleRepository(ctrl)

	// 创建测试依赖
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger := &log.Logger{Logger: slogLogger}
	sidGen := sid.NewSid()

	// 创建 service (JWT 不影响 GetUsers 方法,传入 nil)
	userService := service.NewUserService(logger, sidGen, nil, nil, mockUserRepo, mockUserRoleRepo)

	t.Run("测试用户名精确查询", func(t *testing.T) {
		req := &request.ListUsersRequest{
			Page:     1,
			PageSize: 10,
			Username: "testuser",
		}

		expectedUsers := []*model.User{
			{ID: 1, Username: "testuser"},
		}
		var expectedTotal int64 = 1

		// 设置 mock 期望
		mockUserRepo.EXPECT().
			FindPageWithScopes(1, 10, gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(page, pageSize int, total *int64, where map[string]any, scopes ...func(*gorm.DB) *gorm.DB) ([]*model.User, error) {
				*total = expectedTotal
				// 验证 where 条件
				assert.Equal(t, "testuser", where["username"])
				return expectedUsers, nil
			})

		// 创建测试 Context
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", uint(1))
		c.Set("user_name", "testuser")

		// 调用方法
		users, total, err := userService.List(c, req)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, expectedTotal, total)
		assert.Equal(t, expectedUsers, users)
	})

	t.Run("测试用户名 LIKE 查询", func(t *testing.T) {
		req := &request.ListUsersRequest{
			Page:     1,
			PageSize: 10,
			Username: "like test",
		}

		expectedUsers := []*model.User{
			{ID: 1, Username: "testuser1"},
			{ID: 2, Username: "testuser2"},
		}
		var expectedTotal int64 = 2

		// 设置 mock 期望
		mockUserRepo.EXPECT().
			FindPageWithScopes(1, 10, gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(page, pageSize int, total *int64, where map[string]any, scopes ...func(*gorm.DB) *gorm.DB) ([]*model.User, error) {
				*total = expectedTotal
				// 验证 where 条件包含 LIKE 查询
				assert.Equal(t, "test%", where["username LIKE ?"])
				return expectedUsers, nil
			})

		// 创建测试 Context
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", uint(1))
		c.Set("user_name", "testuser")

		// 调用方法
		users, total, err := userService.List(c, req)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, expectedTotal, total)
		assert.Equal(t, expectedUsers, users)
	})

	t.Run("测试日期范围查询", func(t *testing.T) {
		req := &request.ListUsersRequest{
			Page:      1,
			PageSize:  10,
			StartDate: "2024-01-01",
			EndDate:   "2024-12-31",
		}

		expectedUsers := []*model.User{
			{ID: 1, Username: "user1"},
		}
		var expectedTotal int64 = 1

		// 设置 mock 期望
		mockUserRepo.EXPECT().
			FindPageWithScopes(1, 10, gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(page, pageSize int, total *int64, where map[string]any, scopes ...func(*gorm.DB) *gorm.DB) ([]*model.User, error) {
				*total = expectedTotal
				// 验证 where 条件包含日期范围
				assert.Equal(t, "2024-01-01", where["created_at >="])
				assert.Equal(t, "2024-12-31", where["created_at <="])
				return expectedUsers, nil
			})

		// 创建测试 Context
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", uint(1))
		c.Set("user_name", "testuser")

		// 调用方法
		users, total, err := userService.List(c, req)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, expectedTotal, total)
		assert.Equal(t, expectedUsers, users)
	})

	t.Run("测试组合查询", func(t *testing.T) {
		req := &request.ListUsersRequest{
			Page:      1,
			PageSize:  20,
			Username:  "like admin",
			StartDate: "2024-01-01",
			EndDate:   "2024-12-31",
		}

		expectedUsers := []*model.User{
			{ID: 1, Username: "admin1"},
			{ID: 2, Username: "admin2"},
		}
		var expectedTotal int64 = 2

		// 设置 mock 期望
		mockUserRepo.EXPECT().
			FindPageWithScopes(1, 20, gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(page, pageSize int, total *int64, where map[string]any, scopes ...func(*gorm.DB) *gorm.DB) ([]*model.User, error) {
				*total = expectedTotal
				// 验证所有查询条件
				assert.Equal(t, "admin%", where["username LIKE ?"])
				assert.Equal(t, "2024-01-01", where["created_at >="])
				assert.Equal(t, "2024-12-31", where["created_at <="])
				return expectedUsers, nil
			})

		// 创建测试 Context
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", uint(1))
		c.Set("user_name", "testuser")

		// 调用方法
		users, total, err := userService.List(c, req)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, expectedTotal, total)
		assert.Equal(t, expectedUsers, users)
	})
}
