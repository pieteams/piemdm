package main

import (
	"piemdm/internal/model"
	"piemdm/pkg/helper/sid"
	"piemdm/pkg/log"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Migrate struct {
	db     *gorm.DB
	logger *log.Logger
	sid    *sid.Sid
}

func NewMigrate(db *gorm.DB, logger *log.Logger, sid *sid.Sid) *Migrate {
	return &Migrate{
		db:     db,
		logger: logger,
		sid:    sid,
	}
}
func (m *Migrate) Run() error {
	// Execute migration for approval related tables
	m.logger.Info("Starting migration for approval related tables...")
	// Auto migrate approval related models
	err := m.db.AutoMigrate(
		&model.ApplicationEntityField{},
		&model.ApplicationEntity{},
		&model.Application{},
		&model.ApprovalDefinition{},

		&model.ApprovalNode{},
		&model.ApprovalTask{},
		&model.Approval{},
		&model.CronAction{},
		&model.CronLock{},
		&model.CronLog{},
		&model.CronParam{},
		&model.Cron{},
		&model.NotificationLog{},
		&model.NotificationTemplate{},
		&model.Permission{}, // Permission table
		&model.Role{},
		&model.RolePermission{}, // Role-Permission relation table

		&model.TableFieldGroup{},
		&model.TableField{},
		&model.TableRelation{},
		&model.Table{},
		&model.TablePermission{}, // Table permission
		&model.User{},
		&model.UserRole{}, // User-Role relation table
		&model.WebhookDelivery{},
		&model.Webhook{},
		&model.TableApprovalDefinition{},
		&model.ApplicationApiLog{},
		&model.GlobalId{},
	)
	if err != nil {
		return errors.Wrap(err, "Failed to auto migrate approval tables")
	}
	m.logger.Info("Approval related tables migration completed")

	// Initialize permission data
	m.logger.Info("Starting permission data initialization...")
	permissionInit := NewInitPermissionData(m.db, m.logger)
	if err := permissionInit.Run(); err != nil {
		m.logger.Error("Failed to initialize permission data: " + err.Error())
		// Note: Do not return failure here, continue process
	}

	// Initialize role data
	m.logger.Info("Starting role data initialization...")
	roleInit := NewInitRoleData(m.db, m.logger)
	if err := roleInit.Run(); err != nil {
		m.logger.Error("Failed to initialize role data: " + err.Error())
	}

	// Initialize notification data
	m.logger.Info("Starting notification data initialization...")
	notificationInit := NewInitNotificationData(m.db, m.logger)
	if err := notificationInit.Run(); err != nil {
		m.logger.Error("Failed to initialize notification data: " + err.Error())
	}

	// Initialize global id data
	m.logger.Info("Starting global id data initialization...")
	globalIdInit := NewInitGlobalIdData(m.db, m.logger)
	if err := globalIdInit.Run(); err != nil {
		m.logger.Error("Failed to initialize global id data: " + err.Error())
	}

	return nil
}
