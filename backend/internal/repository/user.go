//go:generate mockgen -source=user.go -destination=../../test/mocks/repository/user.go -package=mock_repository
package repository

import (
	"piemdm/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type UserRepository interface {
	// 基础查询
	FindOne(id uint) (*model.User, error)
	FindByUserID(c *gin.Context, userID string) (*model.User, error)
	FindByUsername(c *gin.Context, username string) (*model.User, error)
	Find(sel string, where map[string]any) ([]*model.User, error)
	FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.User, error)
	FindPageWithScopes(page, pageSize int, total *int64, where map[string]any, scopes ...func(*gorm.DB) *gorm.DB) ([]*model.User, error)

	// Base CRUD
	Create(c *gin.Context, user *model.User) error
	Update(c *gin.Context, user *model.User) error
	Delete(c *gin.Context, id uint) error

	// Batch operations
	BatchUpdate(c *gin.Context, ids []uint, user *model.User) error
	BatchDelete(c *gin.Context, ids []uint) error

	// 权限相关
	GetUserPermissions(c *gin.Context, userID uint) ([]string, error)
	GetUserRoles(c *gin.Context, userID uint) ([]*model.Role, error)
}

type userRepository struct {
	*Repository
	source Base
}

func NewUserRepository(r *Repository, source Base) UserRepository {
	return &userRepository{
		Repository: r,
		source:     source,
	}
}

func (r *userRepository) Create(c *gin.Context, user *model.User) error {
	if err := r.db.WithContext(c).Create(user).Error; err != nil {
		return errors.New("failed to create user: " + err.Error())
	}
	return nil
}

func (r *userRepository) Update(c *gin.Context, user *model.User) error {
	if err := r.db.WithContext(c).Updates(user).Error; err != nil {
		return errors.New("failed to update user: " + err.Error())
	}
	return nil
}

func (r *userRepository) FindByUserID(c *gin.Context, userID string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(c).Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.Errorf("failed to get user by ID: %v", err)
	}

	return &user, nil
}

func (r *userRepository) FindByUsername(c *gin.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get user by username")
	}

	return &user, nil
}

func (r *userRepository) FindOne(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.User, error) {
	var users []*model.User
	var user model.User

	// if err := r.source.Where("deleted_at is null").Offset((page - 1) * pageSize).Limit(pageSize).Find(&crons).Error; err != nil {
	// 	return nil, err
	// }

	// if err := r.source.Model(cron).Where("deleted_at is null").Count(total).Error; err != nil {
	// 	r.logger.Error("users repository count err", "err", err)
	// 	return nil, err
	// }
	preloads := []string{}
	// if err := r.db.Where("deleted_at is null").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
	// 	return nil, err
	// }
	err := r.source.FindPage(user, &users, page, pageSize, total, where, preloads, "ID desc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return users, nil
}

func (r *userRepository) FindPageWithScopes(page, pageSize int, total *int64, where map[string]any, scopes ...func(*gorm.DB) *gorm.DB) ([]*model.User, error) {
	var users []*model.User
	var user model.User
	preloads := []string{}
	err := r.source.FindPageWithScopes(user, &users, page, pageSize, total, where, preloads, scopes, "ID desc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return users, nil
}

func (r *userRepository) Find(sel string, where map[string]any) ([]*model.User, error) {
	var users []*model.User
	// var user model.User
	if sel == "" {
		sel = "*"
	}

	err := r.db.Where(where).Select(sel).Find(&users).Error
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return users, nil
}

func (r *userRepository) BatchUpdate(c *gin.Context, ids []uint, user *model.User) error {
	if err := r.db.WithContext(c).Model(&user).Where("id in ?", ids).Updates(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) Delete(c *gin.Context, id uint) error {
	if err := r.db.WithContext(c).Where("id = ?", id).Delete(&model.User{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) BatchDelete(c *gin.Context, ids []uint) error {
	var users []model.User
	// 使用下面的调用方式，才能把参数传到 Hooks 里面。
	// //单条删除
	// db.Where("id =1").Find(&user).Delete(&user)
	// //多条删除
	// db.Where("id in ?", ids).Find(&users).Delete(&users)

	// if err := r.source.Delete(&user, ids).Error; err != nil {
	if err := r.db.WithContext(c).Where("id in ?", ids).Find(&users).Delete(&users).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetUserPermissions(c *gin.Context, userID uint) ([]string, error) {
	var codes []string
	err := r.db.WithContext(c).Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Pluck("permissions.code", &codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}

func (r *userRepository) GetUserRoles(c *gin.Context, userID uint) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.WithContext(c).Table("roles").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}
