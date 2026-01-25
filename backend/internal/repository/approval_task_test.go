package repository_test

import (
	"testing"
	"time"

	"piemdm/internal/model"
	"piemdm/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupApprovalTaskRepository(t *testing.T) (repository.ApprovalTaskRepository, sqlmock.Sqlmock) {
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
	approvalTaskRepo := repository.NewApprovalTaskRepository(repo, base)

	return approvalTaskRepo, mock
}

func TestApprovalTaskRepository_Create(t *testing.T) {
	approvalTaskRepo, mock := setupApprovalTaskRepository(t)

	now := time.Now()
	approvalTask := &model.ApprovalTask{
		ID:           1,
		ApprovalCode: "APPROVAL_001",
		NodeCode:     "node_001",
		NodeName:     "管理员审批",
		TaskCode:     "TASK_001",
		// Priority:       1,
		Urgency:      "Normal",
		Status:       model.TaskStatusPending,
		AssigneeID:   "user001",
		AssigneeName: "张三",
		CreatedBy:    "admin",
		UpdatedBy:    "admin",
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `approval_tasks`").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := approvalTaskRepo.Create(approvalTask)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApprovalTaskRepository_Update(t *testing.T) {
	approvalTaskRepo, mock := setupApprovalTaskRepository(t)

	now := time.Now()
	approvalTask := &model.ApprovalTask{
		ID:           1,
		ApprovalCode: "APPROVAL_001",
		NodeCode:     "node_001",
		NodeName:     "管理员审批",
		TaskCode:     "TASK_001",
		// Priority:       1,
		Urgency:      "Normal",
		Status:       model.TaskStatusApproved,
		AssigneeID:   "user001",
		AssigneeName: "张三",
		CreatedBy:    "admin",
		UpdatedBy:    "admin",
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `approval_tasks`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := approvalTaskRepo.Update(approvalTask)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApprovalTaskRepository_FindOne(t *testing.T) {
	approvalTaskRepo, mock := setupApprovalTaskRepository(t)

	id := uint(1)

	rows := sqlmock.NewRows([]string{
		"id", "approval_code", "node_code", "node_name", "task_code",
		"priority", "urgency", "status", "assignee_id", "assignee_name",
		"created_by", "updated_by", "created_at", "updated_at",
	}).AddRow(
		1, "APPROVAL_001", "node_001", "管理员审批", "TASK_001",
		1, "Normal", "Pending", "user001", "张三",
		"admin", "admin", time.Now(), time.Now(),
	)
	mock.ExpectQuery("SELECT \\* FROM `approval_tasks` WHERE `approval_tasks`.`id` = \\? ORDER BY `approval_tasks`.`id` LIMIT \\?").WithArgs(id, 1).WillReturnRows(rows)

	approvalTask, err := approvalTaskRepo.FindOne(id)
	assert.NoError(t, err)
	assert.NotNil(t, approvalTask)
	assert.Equal(t, "APPROVAL_001", approvalTask.ApprovalCode)
	assert.Equal(t, "node_001", approvalTask.NodeCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApprovalTaskRepository_FindByApprovalCode(t *testing.T) {
	approvalTaskRepo, mock := setupApprovalTaskRepository(t)

	approvalCode := "APPROVAL_001"

	rows := sqlmock.NewRows([]string{
		"id", "approval_code", "node_code", "node_name", "task_code",
		"priority", "urgency", "status", "assignee_id", "assignee_name",
		"created_by", "updated_by", "created_at", "updated_at",
	}).AddRow(
		1, "APPROVAL_001", "node_001", "管理员审批", "TASK_001",
		1, "Normal", "Pending", "user001", "张三",
		"admin", "admin", time.Now(), time.Now(),
	).AddRow(
		2, "APPROVAL_001", "node_002", "财务审批", "TASK_002",
		1, "Normal", "Pending", "user002", "李四",
		"admin", "admin", time.Now(), time.Now(),
	)
	mock.ExpectQuery("SELECT \\* FROM `approval_tasks`").WillReturnRows(rows)

	approvalTasks, err := approvalTaskRepo.FindByApprovalCode(approvalCode)
	assert.NoError(t, err)
	assert.NotNil(t, approvalTasks)
	assert.Len(t, approvalTasks, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApprovalTaskRepository_FindByAssigneeID(t *testing.T) {
	approvalTaskRepo, mock := setupApprovalTaskRepository(t)

	assigneeID := "user001"

	rows := sqlmock.NewRows([]string{
		"id", "approval_code", "node_code", "node_name", "task_code",
		"priority", "urgency", "status", "assignee_id", "assignee_name",
		"created_by", "updated_by", "created_at", "updated_at",
	}).AddRow(
		1, "APPROVAL_001", "node_001", "管理员审批", "TASK_001",
		1, "Normal", "Pending", "user001", "张三",
		"admin", "admin", time.Now(), time.Now(),
	).AddRow(
		2, "APPROVAL_002", "node_001", "管理员审批", "TASK_003",
		1, "Normal", "Pending", "user001", "张三",
		"admin", "admin", time.Now(), time.Now(),
	)
	mock.ExpectQuery("SELECT \\* FROM `approval_tasks`").WillReturnRows(rows)

	approvalTasks, err := approvalTaskRepo.FindByAssigneeID(assigneeID)
	assert.NoError(t, err)
	assert.NotNil(t, approvalTasks)
	assert.Len(t, approvalTasks, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApprovalTaskRepository_FindByStatus(t *testing.T) {
	approvalTaskRepo, mock := setupApprovalTaskRepository(t)

	status := model.TaskStatusPending

	rows := sqlmock.NewRows([]string{
		"id", "approval_code", "node_code", "node_name", "task_code",
		"priority", "urgency", "status", "assignee_id", "assignee_name",
		"created_by", "updated_by", "created_at", "updated_at",
	}).AddRow(
		1, "APPROVAL_001", "node_001", "管理员审批", "TASK_001",
		1, "Normal", "Pending", "user001", "张三",
		"admin", "admin", time.Now(), time.Now(),
	).AddRow(
		2, "APPROVAL_002", "node_001", "管理员审批", "TASK_002",
		1, "Normal", "Pending", "user002", "李四",
		"admin", "admin", time.Now(), time.Now(),
	)
	mock.ExpectQuery("SELECT \\* FROM `approval_tasks`").WillReturnRows(rows)

	approvalTasks, err := approvalTaskRepo.FindByStatus(status)
	assert.NoError(t, err)
	assert.NotNil(t, approvalTasks)
	assert.Len(t, approvalTasks, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApprovalTaskRepository_UpdateStatus(t *testing.T) {
	approvalTaskRepo, mock := setupApprovalTaskRepository(t)

	id := uint(1)
	status := model.TaskStatusApproved

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `approval_tasks` SET `status`=\\?,`updated_at`=\\? WHERE id = \\? AND `approval_tasks`.`deleted_at` IS NULL").
		WithArgs(status, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := approvalTaskRepo.UpdateStatus(id, status)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApprovalTaskRepository_CountByStatus(t *testing.T) {
	approvalTaskRepo, mock := setupApprovalTaskRepository(t)

	status := model.TaskStatusPending

	rows := sqlmock.NewRows([]string{"count"}).AddRow(3)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `approval_tasks`").WillReturnRows(rows)

	count, err := approvalTaskRepo.CountByStatus(status)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApprovalTaskRepository_Delete(t *testing.T) {
	approvalTaskRepo, mock := setupApprovalTaskRepository(t)

	id := uint(1)

	// 先查询记录是否存在
	rows := sqlmock.NewRows([]string{
		"id", "approval_code", "node_code", "node_name", "task_code",
		"priority", "urgency", "status", "assignee_id", "assignee_name",
		"created_by", "updated_by", "created_at", "updated_at",
	}).AddRow(
		1, "APPROVAL_001", "node_001", "管理员审批", "TASK_001",
		1, "Normal", "Pending", "user001", "张三",
		"admin", "admin", time.Now(), time.Now(),
	)
	mock.ExpectQuery("SELECT \\* FROM `approval_tasks` WHERE id = \\? AND `approval_tasks`.`deleted_at` IS NULL ORDER BY `approval_tasks`.`id` LIMIT \\?").
		WithArgs(id, 1).
		WillReturnRows(rows)

	// 模拟软删除操作 - GORM软删除会设置deleted_at字段，同时BeforeDelete钩子会更新状态为Deleted
	mock.ExpectBegin()
	// 首先执行BeforeDelete钩子中的更新操作
	mock.ExpectExec("UPDATE `approval_tasks` SET `status`=\\?,`updated_by`=\\?,`updated_at`=\\? WHERE id = \\? AND `approval_tasks`.`deleted_at` IS NULL AND `id` = \\?").
		WithArgs("Deleted", "admin", sqlmock.AnyArg(), id, id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// 然后执行GORM的软删除操作，设置deleted_at字段
	mock.ExpectExec("UPDATE `approval_tasks` SET `deleted_at`=\\? WHERE `approval_tasks`.`id` = \\? AND `approval_tasks`.`deleted_at` IS NULL").
		WithArgs(sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	deletedApprovalTask, err := approvalTaskRepo.Delete(id)
	assert.NoError(t, err)
	assert.NotNil(t, deletedApprovalTask)
	assert.Equal(t, "APPROVAL_001", deletedApprovalTask.ApprovalCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}
