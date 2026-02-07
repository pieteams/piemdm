package repository

import (
	"context"
	"piemdm/internal/model"
)

type RoleRepository interface {
	FindOne(id uint) (*model.Role, error)
	Find(sel string, where map[string]any) ([]*model.Role, error)
	FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.Role, error)
	Create(c context.Context, role *model.Role) error
	Update(c context.Context, role *model.Role) error
	BatchUpdate(c context.Context, ids []uint, role *model.Role) error
	Delete(c context.Context, id uint) (*model.Role, error)
	BatchDelete(c context.Context, ids []uint) error

	// Permission management
	GetPermissions(c context.Context, roleID uint) ([]*model.Permission, error)
	AssignPermissions(c context.Context, roleID uint, permissionIDs []uint) error
	RemovePermissions(c context.Context, roleID uint, permissionIDs []uint) error
	UpdatePermissions(c context.Context, roleID uint, permissionIDs []uint) error
}
type soleRepository struct {
	*Repository
	source Base
}

func NewRoleRepository(repository *Repository, source Base) RoleRepository {
	return &soleRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *soleRepository) FindOne(id uint) (*model.Role, error) {
	var role model.Role
	if err := r.source.FirstById(&role, id); err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *soleRepository) FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.Role, error) {
	var roles []*model.Role
	var role model.Role

	// if err := r.db.Where("deleted_at is null").Offset((page - 1) * pageSize).Limit(pageSize).Find(&crons).Error; err != nil {
	// 	return nil, err
	// }

	// if err := r.db.Model(cron).Where("deleted_at is null").Count(total).Error; err != nil {
	// 	r.logger.Error("roles repository count err", "err", err)
	// 	return nil, err
	// }
	preloads := []string{}
	err := r.source.FindPage(role, &roles, page, pageSize, total, where, preloads, "ID desc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return roles, nil
}

func (r *soleRepository) Find(sel string, where map[string]any) ([]*model.Role, error) {
	var roles []*model.Role
	var role model.Role
	if sel == "" {
		sel = "*"
	}

	err := r.source.Find(role, &roles, sel, where, "id asc")
	if err != nil {
		r.logger.Error("获取模型信息失败", "err", err)
	}
	return roles, nil
}

func (r *soleRepository) Create(c context.Context, role *model.Role) error {
	if err := r.db.WithContext(c).Create(role).Error; err != nil {
		return err
	}
	return nil
}

func (r *soleRepository) Update(c context.Context, role *model.Role) error {
	if err := r.db.WithContext(c).Updates(role).Error; err != nil {
		return err
	}
	return nil
}

func (r *soleRepository) BatchUpdate(c context.Context, ids []uint, role *model.Role) error {
	if err := r.db.WithContext(c).Model(&role).Where("id in ?", ids).Updates(role).Error; err != nil {
		return err
	}
	return nil
}

func (r *soleRepository) Delete(c context.Context, id uint) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(c).Where("id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *soleRepository) BatchDelete(c context.Context, ids []uint) error {
	var roles []model.Role
	// 使用下面的调用方式，才能把参数传到 Hooks 里面。
	// //单条删除
	// db.Where("id =1").Find(&cron).Delete(&cron)
	// //多条删除
	// db.Where("id in ?", ids).Find(&crons).Delete(&crons)

	// if err := r.db.Delete(&cron, ids).Error; err != nil {
	if err := r.db.WithContext(c).Where("id in ?", ids).Find(&roles).Delete(&roles).Error; err != nil {
		return err
	}
	return nil
}

func (r *soleRepository) GetPermissions(c context.Context, roleID uint) ([]*model.Permission, error) {
	var role model.Role
	role.ID = roleID
	var permissions []*model.Permission
	if err := r.db.WithContext(c).Model(&role).Association("Permissions").Find(&permissions); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *soleRepository) AssignPermissions(c context.Context, roleID uint, permissionIDs []uint) error {
	var role model.Role
	role.ID = roleID
	var permissions []model.Permission
	for _, pid := range permissionIDs {
		permissions = append(permissions, model.Permission{ID: pid})
	}
	return r.db.WithContext(c).Model(&role).Association("Permissions").Append(permissions)
}

func (r *soleRepository) RemovePermissions(c context.Context, roleID uint, permissionIDs []uint) error {
	var role model.Role
	role.ID = roleID
	var permissions []model.Permission
	for _, pid := range permissionIDs {
		permissions = append(permissions, model.Permission{ID: pid})
	}
	return r.db.WithContext(c).Model(&role).Association("Permissions").Delete(permissions)
}

func (r *soleRepository) UpdatePermissions(c context.Context, roleID uint, permissionIDs []uint) error {
	var role model.Role
	role.ID = roleID
	var permissions []model.Permission
	for _, pid := range permissionIDs {
		permissions = append(permissions, model.Permission{ID: pid})
	}
	return r.db.WithContext(c).Model(&role).Association("Permissions").Replace(permissions)
}
