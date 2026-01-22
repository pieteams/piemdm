package repository_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"piemdm/internal/model"
	"piemdm/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupApprovalDefRepository(t *testing.T) (repository.ApprovalDefinitionRepository, sqlmock.Sqlmock) {
	// 设置Gin为测试模式，避免调试信息输出
	gin.SetMode(gin.TestMode)

	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm connection: %v", err)
	}

	rdb, _ := redismock.NewClientMock()
	repo := repository.NewRepository(db, rdb, nil)
	base := repository.NewBaseRepository(repo)
	approvalDefRepo := repository.NewApprovalDefinitionRepository(repo, base)

	return approvalDefRepo, mock
}

func TestApprovalDefRepository_Create(t *testing.T) {
	approvalDefRepo, mock := setupApprovalDefRepository(t)

	now := time.Now()
	approvalDef := &model.ApprovalDefinition{
		ID:          1,
		Code:        "TEST_APPROVAL_001",
		Name:        "测试审批流程",
		Description: "测试审批流程描述",
		FormData:    `{"fields":[]}`,
		NodeList:    `{"nodes":[]}`,
		Status:      "Normal",
		CreatedBy:   "admin",
		UpdatedBy:   "admin",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `approval_definitions`").
		WithArgs(
			sqlmock.AnyArg(), // Code
			sqlmock.AnyArg(), // Name
			sqlmock.AnyArg(), // Description
			sqlmock.AnyArg(), // FormData
			sqlmock.AnyArg(), // NodeList
			sqlmock.AnyArg(), // Platform
			sqlmock.AnyArg(), // Status
			sqlmock.AnyArg(), // CreatedBy
			sqlmock.AnyArg(), // UpdatedBy
			sqlmock.AnyArg(), // CreatedAt
			sqlmock.AnyArg(), // UpdatedAt
			sqlmock.AnyArg(), // DeletedAt
			sqlmock.AnyArg(), // ID
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	err := approvalDefRepo.Create(ctx, approvalDef)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApprovalDefRepository_Update(t *testing.T) {
	approvalDefRepo, mock := setupApprovalDefRepository(t)

	now := time.Now()
	approvalDef := &model.ApprovalDefinition{
		ID:          1,
		Code:        "TEST_APPROVAL_001",
		Name:        "测试审批流程-更新",
		Description: "测试审批流程描述-更新",
		FormData:    `{"fields":[]}`,
		NodeList:    `{"nodes":[]}`,
		Status:      "Normal",
		UpdatedBy:   "admin",
		UpdatedAt:   now,
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `approval_definitions`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	err := approvalDefRepo.Update(ctx, approvalDef)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApprovalDefRepository_FindOne(t *testing.T) {
	approvalDefRepo, mock := setupApprovalDefRepository(t)

	id := uint(1)

	rows := sqlmock.NewRows([]string{
		"id", "code", "name", "description", "form_data", "node_list", "status", "created_by", "updated_by", "created_at", "updated_at",
	}).AddRow(
		1, "TEST_APPROVAL_001", "测试审批流程", "测试审批流程描述", `{"fields":[]}`, `{"nodes":[]}`, "Normal", "admin", "admin", time.Now(), time.Now(),
	)
	mock.ExpectQuery("SELECT \\* FROM `approval_definitions` WHERE `approval_definitions`.`id` = \\? ORDER BY `approval_definitions`.`id` LIMIT \\?").
		WithArgs(id, 1).
		WillReturnRows(rows)

	approvalDef, err := approvalDefRepo.FindOne(id)
	assert.NoError(t, err)
	assert.NotNil(t, approvalDef)
	assert.Equal(t, "TEST_APPROVAL_001", approvalDef.Code)
	assert.Equal(t, "测试审批流程", approvalDef.Name)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApprovalDefRepository_Delete(t *testing.T) {
	approvalDefRepo, mock := setupApprovalDefRepository(t)

	id := uint(1)

	rows := sqlmock.NewRows([]string{"id", "code"}).AddRow(1, "TEST_CODE")
	mock.ExpectQuery("SELECT \\* FROM `approval_definitions` WHERE id = \\? AND `approval_definitions`.`deleted_at` IS NULL ORDER BY `approval_definitions`.`id` LIMIT \\?").
		WithArgs(id, 1).
		WillReturnRows(rows)

	mock.ExpectBegin()
	// 首先执行BeforeDelete钩子中的更新操作
	mock.ExpectExec("UPDATE `approval_definitions` SET `status`=\\?,`updated_by`=\\?,`updated_at`=\\? WHERE id = \\? AND `approval_definitions`.`deleted_at` IS NULL AND `id` = \\?").
		WithArgs("Deleted", "", sqlmock.AnyArg(), id, id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// 然后执行GORM的软删除操作，设置deleted_at字段
	mock.ExpectExec("UPDATE `approval_definitions` SET `deleted_at`=\\? WHERE `approval_definitions`.`id` = \\? AND `approval_definitions`.`deleted_at` IS NULL").
		WithArgs(sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	deletedApprovalDef, err := approvalDefRepo.Delete(ctx, id)
	assert.NoError(t, err)
	assert.NotNil(t, deletedApprovalDef)
	assert.Equal(t, uint(1), deletedApprovalDef.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}
