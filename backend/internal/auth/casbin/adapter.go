package casbin

import (
	"context"
	"fmt"
	"strconv"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"gorm.io/gorm"
)

// Adapter represents the Gorm adapter for policy storage.
type Adapter struct {
	db *gorm.DB
}

// NewAdapter is the constructor for Adapter.
func NewAdapter(db *gorm.DB) *Adapter {
	return &Adapter{db: db}
}

// LoadPolicy loads all policy rules from the storage.
func (a *Adapter) LoadPolicy(model model.Model) error {
	ctx := context.Background()

	// 1. Load role assignment policies (g, userID, roleCode)
	// We use userID as subject for 'g'.
	type UserRoleResult struct {
		UserID   uint
		RoleCode string
	}
	var userRoles []UserRoleResult

	err := a.db.WithContext(ctx).Table("user_roles").
		Select("user_roles.user_id, roles.code as role_code").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Scan(&userRoles).Error
	if err != nil {
		return err
	}

	for _, ur := range userRoles {
		line := fmt.Sprintf("g, %s, %s", strconv.FormatUint(uint64(ur.UserID), 10), ur.RoleCode)
		persist.LoadPolicyLine(line, model)
	}

	// 2. Load permission policies (p, roleCode, resource, action)
	type RolePermissionResult struct {
		RoleCode string
		Resource string
		Action   string
	}
	var rolePermissions []RolePermissionResult
	err = a.db.WithContext(ctx).Table("role_permissions").
		Select("roles.code as role_code, permissions.resource, permissions.action").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("permissions.status = ?", "Normal").
		Scan(&rolePermissions).Error
	if err != nil {
		return err
	}

	for _, rp := range rolePermissions {
		if rp.Resource == "" || rp.Action == "" {
			continue
		}
		line := fmt.Sprintf("p, %s, %s, %s", rp.RoleCode, rp.Resource, rp.Action)
		persist.LoadPolicyLine(line, model)
	}

	return nil
}

// SavePolicy saves all policy rules to the storage.
func (a *Adapter) SavePolicy(model model.Model) error {
	return nil // Managed by internal APIs
}

// AddPolicy adds a policy rule to the storage.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	return nil // Managed by internal APIs
}

// RemovePolicy removes a policy rule from the storage.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil // Managed by internal APIs
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil // Managed by internal APIs
}
