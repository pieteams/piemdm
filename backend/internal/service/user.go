//go:generate mockgen -source=user.go -destination=../../test/mocks/service/user.go -package=mock_service
package service

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"piemdm/internal/model"
	"piemdm/internal/transport/request"
	"piemdm/internal/repository"
	"piemdm/pkg/helper/sid"
	"piemdm/pkg/jwt"
	"piemdm/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	// Base CRUD
	Get(id uint) (*model.User, error)
	List(c *gin.Context, req *request.ListUsersRequest) ([]*model.User, int64, error)
	Find(sel string, where map[string]any) ([]*model.User, error)
	Create(c *gin.Context, user *model.User) error
	Update(c *gin.Context, user *model.User) error
	Delete(c *gin.Context, id uint) error

	// Batch operations
	BatchUpdate(c *gin.Context, ids []uint, user *model.User) error
	BatchDelete(c *gin.Context, ids []uint) error

	// 特殊操作
	Login(c *gin.Context, req *request.LoginRequest) (string, *model.User, error)
	GetUserPermissions(c *gin.Context, userID uint) ([]string, error)
	HasPermission(c *gin.Context, userID uint, permissionCode string) (bool, error)
	GetUserMaxDataScope(c *gin.Context, userID uint) (string, error)
	GetSubordinateUsernames(c *gin.Context, username string) ([]string, error)
	WithDataScope(c *gin.Context) func(db *gorm.DB) *gorm.DB

	// 角色管理
	GetUserRoles(userID uint) ([]model.Role, error)
	UpdateUserRoles(userID uint, roleIDs []uint) error
}

type userService struct {
	logger       *log.Logger
	sid          *sid.Sid
	jwt          *jwt.JWT
	rdb          *redis.Client
	userRepo     repository.UserRepository
	userRoleRepo repository.UserRoleRepository
}

func NewUserService(logger *log.Logger, sid *sid.Sid, jwt *jwt.JWT, rdb *redis.Client, userRepo repository.UserRepository, userRoleRepo repository.UserRoleRepository) UserService {
	return &userService{
		logger:       logger,
		sid:          sid,
		jwt:          jwt,
		rdb:          rdb,
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
	}
}

func (s *userService) Login(c *gin.Context, req *request.LoginRequest) (string, *model.User, error) {
	user, err := s.userRepo.FindByUsername(c, req.Username)
	if err != nil {
		return "", nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to hash password")
	}
	token, err := s.jwt.GenToken(strconv.FormatUint(uint64(user.ID), 10), user.Username, user.Email, user.Admin)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to generate JWT token")
	}

	// 填充用户权限
	permissions, err := s.userRepo.GetUserPermissions(c, user.ID)
	if err != nil {
		// 记录错误但不阻塞登录
		s.logger.Error("获取用户权限失败", "err", err)
	} else {
		user.Permissions = permissions
	}

	return token, user, nil
}

func (s *userService) Get(id uint) (*model.User, error) {
	return s.userRepo.FindOne(id)
}

func (s *userService) List(c *gin.Context, req *request.ListUsersRequest) ([]*model.User, int64, error) {
	where := make(map[string]any)
	var total int64

	// 处理用户名查询(支持 like 查询)
	if req.Username != "" {
		trimmedUsername := strings.TrimSpace(req.Username)
		if likeValue, ok := strings.CutPrefix(trimmedUsername, "like "); ok {
			// 提取 like 后面的内容并去除空格
			where["username LIKE ?"] = strings.TrimSpace(likeValue) + "%"
		} else {
			where["username"] = req.Username
		}
	}

	// 处理日期范围过滤
	if req.StartDate != "" {
		where["created_at >="] = req.StartDate
	}
	if req.EndDate != "" {
		where["created_at <="] = req.EndDate
	}

	// 使用新的 FindPageWithScopes 和 WithDataScope
	// 注意: WithDataScope 返回的是 GORM scope 函数
	users, err := s.userRepo.FindPageWithScopes(req.Page, req.PageSize, &total, where, s.WithDataScope(c))
	return users, total, err
}

func (s *userService) Find(sel string, where map[string]any) ([]*model.User, error) {
	return s.userRepo.Find(sel, where)
}

func (s *userService) Create(c *gin.Context, user *model.User) error {
	// 检查用户名是否已存在
	if user, err := s.userRepo.FindByUsername(c, user.Username); err == nil && user != nil {
		return errors.New("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "failed to hash password")
	}
	// Generate user ID
	employeeID, err := s.sid.GenString()
	if err != nil {
		return errors.Wrap(err, "failed to generate user ID")
	}
	// Create a user
	if user.EmployeeID == "" {
		user.EmployeeID = employeeID
	}
	user.Password = string(hashedPassword)
	return s.userRepo.Create(c, user)
}

func (s *userService) Update(c *gin.Context, user *model.User) error {
	// 如果没有ID，无法更新
	if user.ID == 0 {
		return errors.New("user ID is required")
	}

	// 检查用户是否存在
	userOld, err := s.userRepo.FindOne(user.ID)
	if err != nil {
		return errors.Wrap(err, "user not found")
	}

	// 只有当密码变化时才重新哈希
	if user.Password != "" && user.Password != userOld.Password {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, "failed to hash password")
		}
		user.Password = string(hashedPassword)
	} else {
		// 如果密码为空或未变化，保留原密码
		user.Password = userOld.Password
	}

	return s.userRepo.Update(c, user)
}

func (s *userService) BatchUpdate(c *gin.Context, ids []uint, user *model.User) error {
	return s.userRepo.BatchUpdate(c, ids, user)
}

func (s *userService) Delete(c *gin.Context, id uint) error {
	return s.userRepo.Delete(c, id)
}

func (s *userService) BatchDelete(c *gin.Context, ids []uint) error {
	return s.userRepo.BatchDelete(c, ids)
}

func (s *userService) GetUserPermissions(c *gin.Context, userID uint) ([]string, error) {
	return s.userRepo.GetUserPermissions(c, userID)
}

func (s *userService) HasPermission(c *gin.Context, userID uint, permissionCode string) (bool, error) {
	permissions, err := s.userRepo.GetUserPermissions(c, userID)
	if err != nil {
		return false, err
	}
	for _, p := range permissions {
		if p == permissionCode {
			return true, nil
		}
	}
	return false, nil
}

func (s *userService) GetUserMaxDataScope(c *gin.Context, userID uint) (string, error) {
	roles, err := s.userRepo.GetUserRoles(c, userID)
	if err != nil {
		return "", err
	}

	// 优先级: All > Subordinate > Self
	hasAll := false
	hasSubordinate := false

	for _, role := range roles {
		if role.DataScope == "All" {
			hasAll = true
			break
		}
		if role.DataScope == "Subordinate" {
			hasSubordinate = true
		}
	}

	if hasAll {
		return "All", nil
	}
	if hasSubordinate {
		return "Subordinate", nil
	}
	return "Self", nil
}

func (s *userService) GetSubordinateUsernames(c *gin.Context, username string) ([]string, error) {
	// 尝试从缓存获取
	cacheKey := "piemdm:users:subordinates:" + username
	val, err := s.rdb.Get(c, cacheKey).Result()
	if err == nil {
		var cachedSubs []string
		if err := json.Unmarshal([]byte(val), &cachedSubs); err == nil {
			return cachedSubs, nil
		}
	}

	// 递归查找下属
	var subordinates []string

	// 1. 查找直接下属
	users, err := s.userRepo.Find("username", map[string]any{"superior_username": username})
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		subordinates = append(subordinates, u.Username)
		// 递归
		subs, err := s.GetSubordinateUsernames(c, u.Username)
		if err == nil {
			subordinates = append(subordinates, subs...)
		}
	}

	// 写入缓存(1小时过期)
	if data, err := json.Marshal(subordinates); err == nil {
		s.rdb.Set(c, cacheKey, data, time.Hour)
	}

	return subordinates, nil
}

// GetUserRoles 获取用户的角色
func (s *userService) GetUserRoles(userID uint) ([]model.Role, error) {
	return s.userRoleRepo.FindRolesByUserID(userID)
}

// UpdateUserRoles 更新用户的角色
func (s *userService) UpdateUserRoles(userID uint, roleIDs []uint) error {
	return s.userRoleRepo.UpdateUserRoles(userID, roleIDs)
}
