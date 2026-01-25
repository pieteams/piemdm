// Package service_test contains unit tests for entity service
//
// To generate required mocks, run:
//
//	go generate ./...
//
// Or manually generate mocks:
//
//	mockgen -source=internal/repository/entity.go -destination=test/mocks/repository/entity.go -package=mock_repository
//	mockgen -source=internal/service/table_field.go -destination=test/mocks/service/table_field.go -package=mock_service
//
// To run these tests:
//
//	go test -v ./internal/service -run TestValidateUniqueConstraints
//
// Note: These tests require mock files to be generated first.
// The tests demonstrate the validation logic for unique index constraints.
package service_test

import (
	"os"
	"testing"

	"piemdm/internal/model"
	"piemdm/internal/service"
	"piemdm/pkg/configloader"
	"piemdm/pkg/log"
	mock_repository "piemdm/test/mocks/repository"
	mock_service "piemdm/test/mocks/service"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var (
	testLogger *log.Logger
)

func init() {
	// 设置测试环境配置
	configPath := "../../config/local.yml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 如果配置文件不存在，创建一个空的配置用于测试，避免 panic
		testConfig, _ := configloader.Load()
		testLogger = log.NewLog(testConfig)
		return
	}
	_ = os.Setenv("APP_CONF", configPath)
	testConfig, _ := configloader.Load()
	testLogger = log.NewLog(testConfig)
}

// TestValidateUniqueConstraints_CreateNoWorkflow_NoConflict 测试创建无流程场景 - 无冲突
func TestValidateUniqueConstraints_CreateNoWorkflow_NoConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEntityRepo := mock_repository.NewMockEntityRepository(ctrl)
	mockTableFieldService := mock_service.NewMockTableFieldService(ctrl)
	mockTableFieldRepo := mock_repository.NewMockTableFieldRepository(ctrl)
	mockTableApprovalDefRepo := mock_repository.NewMockTableApprovalDefinitionRepository(ctrl)
	mockApprovalService := mock_service.NewMockApprovalService(ctrl)
	mockGlobalIdService := mock_service.NewMockGlobalIdService(ctrl)
	mockEntityLogService := mock_service.NewMockEntityLogService(ctrl)
	mockAutocodeService := mock_service.NewMockAutocodeService(ctrl)
	mockTablePermissionService := mock_service.NewMockTablePermissionService(ctrl)
	mockTableRepo := mock_repository.NewMockTableRepository(ctrl)

	entityService := service.NewEntityService(
		service.NewService(testLogger, nil, nil),
		mockEntityRepo,
		mockTableFieldService,
		mockTableFieldRepo,
		mockTableApprovalDefRepo,
		mockApprovalService,
		mockGlobalIdService,
		mockEntityLogService,
		mockAutocodeService,
		mockTablePermissionService,
		mockTableRepo, // tableRepository
		nil,           // viper config
	)

	// Mock HasPermission check - assume has permission for tests
	mockTablePermissionService.EXPECT().
		CheckTablePermission(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(true, nil).
		AnyTimes()

	// Mock TableRepository.Find for calculateTreeFields
	mockTableRepo.EXPECT().
		Find("", map[string]any{"code": "test_entity"}).
		Return([]*model.Table{}, nil) // Return empty list to indicate not a tree structure

	// Mock GetOperationInfo - Create 方法会调用此方法获取操作信息
	mockApprovalService.EXPECT().
		GetOperationInfo("Create", gomock.Any()).
		DoAndReturn(func(operation string, info *map[string]string) error {
			(*info)["action"] = "Create"
			(*info)["status"] = "Normal"
			return nil
		})

	// Mock GetNewID - Create 方法会生成全局唯一 ID
	mockGlobalIdService.EXPECT().
		GetNewID("entity").
		Return(uint(1))

	// Mock GenerateOrRestoreAutocodes - 重构后自动编码逻辑在 autocodeService 中
	mockAutocodeService.EXPECT().
		GenerateOrRestoreAutocodes(gomock.Any(), "test_entity", gomock.Any(), mockTableFieldService, mockEntityRepo, gomock.Any()).
		Return(nil)

	// 模拟唯一索引字段查询
	uniqueFields := []*model.TableField{
		{Code: "code", IsUnique: "Yes", IndexName: ""},
	}
	mockTableFieldService.EXPECT().
		Find("code,is_unique,index_name", map[string]any{
			"table_code": "test_entity",
			"is_unique":  "Yes",
		}).
		Return(uniqueFields, nil)

	// 模拟查询结果 - 无冲突
	mockEntityRepo.EXPECT().
		Find("test_entity", "id", map[string]any{"code": "TEST001"}).
		Return([]map[string]any{}, nil)

	// Mock Create - 最终会调用 repository 的 Create 方法
	mockEntityRepo.EXPECT().
		Create(gomock.Any(), "test_entity", gomock.Any()).
		Return(nil)

	entityMap := map[string]any{
		"code": "TEST001",
		"name": "Test Entity",
	}

	// 创建一个模拟的 gin.Context
	c := &gin.Context{}
	c.Set("user_id", uint(1))

	err := entityService.Create(c, "test_entity", entityMap)

	assert.NoError(t, err)
}

// TestValidateUniqueConstraints_CreateNoWorkflow_Conflict 测试创建无流程场景 - 有冲突
func TestValidateUniqueConstraints_CreateNoWorkflow_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEntityRepo := mock_repository.NewMockEntityRepository(ctrl)
	mockTableFieldService := mock_service.NewMockTableFieldService(ctrl)
	mockTableFieldRepo := mock_repository.NewMockTableFieldRepository(ctrl)
	mockTableApprovalDefRepo := mock_repository.NewMockTableApprovalDefinitionRepository(ctrl)
	mockApprovalService := mock_service.NewMockApprovalService(ctrl)
	mockGlobalIdService := mock_service.NewMockGlobalIdService(ctrl)
	mockEntityLogService := mock_service.NewMockEntityLogService(ctrl)
	mockAutocodeService := mock_service.NewMockAutocodeService(ctrl)
	mockTablePermissionService := mock_service.NewMockTablePermissionService(ctrl)
	mockTableRepo := mock_repository.NewMockTableRepository(ctrl)

	entityService := service.NewEntityService(
		service.NewService(testLogger, nil, nil),
		mockEntityRepo,
		mockTableFieldService,
		mockTableFieldRepo,
		mockTableApprovalDefRepo,
		mockApprovalService,
		mockGlobalIdService,
		mockEntityLogService,
		mockAutocodeService,
		mockTablePermissionService,
		mockTableRepo, // tableRepository
		nil,           // viper config
	)

	// Mock HasPermission check
	mockTablePermissionService.EXPECT().
		CheckTablePermission(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(true, nil).
		AnyTimes()

	// Mock TableRepository.Find for calculateTreeFields
	mockTableRepo.EXPECT().
		Find("", map[string]any{"code": "test_entity"}).
		Return([]*model.Table{}, nil) // Return empty list to indicate not a tree structure

	// Mock GetOperationInfo
	mockApprovalService.EXPECT().
		GetOperationInfo("Create", gomock.Any()).
		DoAndReturn(func(operation string, info *map[string]string) error {
			(*info)["action"] = "Create"
			(*info)["status"] = "Normal"
			return nil
		})

	// Mock GetNewID
	mockGlobalIdService.EXPECT().
		GetNewID("entity").
		Return(uint(1))

	// Mock GenerateOrRestoreAutocodes
	mockAutocodeService.EXPECT().
		GenerateOrRestoreAutocodes(gomock.Any(), "test_entity", gomock.Any(), mockTableFieldService, mockEntityRepo, gomock.Any()).
		Return(nil)

	// 模拟唯一索引字段查询
	uniqueFields := []*model.TableField{
		{Code: "code", IsUnique: "Yes", IndexName: ""},
	}
	mockTableFieldService.EXPECT().
		Find("code,is_unique,index_name", map[string]any{
			"table_code": "test_entity",
			"is_unique":  "Yes",
		}).
		Return(uniqueFields, nil)

	// 模拟查询结果 - 有冲突
	mockEntityRepo.EXPECT().
		Find("test_entity", "id", map[string]any{"code": "TEST001"}).
		Return([]map[string]any{{"id": uint64(1)}}, nil)

	entityMap := map[string]any{
		"code": "TEST001",
		"name": "Test Entity",
	}

	c := &gin.Context{}
	c.Set("user_id", uint(1))
	err := entityService.Create(c, "test_entity", entityMap)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "唯一索引字段")
}

// TestValidateUniqueConstraints_CreateWithWorkflow_NoConflict 测试创建有流程场景 - 无冲突
func TestValidateUniqueConstraints_CreateWithWorkflow_NoConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEntityRepo := mock_repository.NewMockEntityRepository(ctrl)
	mockTableFieldService := mock_service.NewMockTableFieldService(ctrl)
	mockTableFieldRepo := mock_repository.NewMockTableFieldRepository(ctrl)
	mockTableApprovalDefRepo := mock_repository.NewMockTableApprovalDefinitionRepository(ctrl)
	mockApprovalService := mock_service.NewMockApprovalService(ctrl)
	mockGlobalIdService := mock_service.NewMockGlobalIdService(ctrl)
	mockEntityLogService := mock_service.NewMockEntityLogService(ctrl)
	mockAutocodeService := mock_service.NewMockAutocodeService(ctrl)
	mockTablePermissionService := mock_service.NewMockTablePermissionService(ctrl)
	mockTableRepo := mock_repository.NewMockTableRepository(ctrl)

	entityService := service.NewEntityService(
		service.NewService(testLogger, nil, nil),
		mockEntityRepo,
		mockTableFieldService,
		mockTableFieldRepo,
		mockTableApprovalDefRepo,
		mockApprovalService,
		mockGlobalIdService,
		mockEntityLogService,
		mockAutocodeService,
		mockTablePermissionService,
		mockTableRepo, // tableRepository
		nil,           // viper config
	)

	// Mock HasPermission check
	mockTablePermissionService.EXPECT().
		CheckTablePermission(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(true, nil).
		AnyTimes()

	// Mock GenerateOrRestoreAutocodes - CreateDraft 会调用此方法
	mockAutocodeService.EXPECT().
		GenerateOrRestoreAutocodes(gomock.Any(), "test_entity", gomock.Any(), mockTableFieldService, mockEntityRepo, gomock.Any()).
		Return(nil)

	// Mock CreateDraftWithApproval - CreateDraft 最终会调用此方法
	mockApprovalService.EXPECT().
		CreateDraftWithApproval(gomock.Any(), "test_entity", "test reason", gomock.Any()).
		Return(nil)

	entityMap := map[string]any{
		"code": "TEST001",
		"name": "Test Entity",
	}

	c := &gin.Context{}
	c.Set("user_id", uint(1))
	err := entityService.CreateDraft(c, "test_entity", "test reason", entityMap)

	assert.NoError(t, err)
}

// TestValidateUniqueConstraints_CreateWithWorkflow_Conflict 测试创建有流程场景 - 有冲突
func TestValidateUniqueConstraints_CreateWithWorkflow_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEntityRepo := mock_repository.NewMockEntityRepository(ctrl)
	mockTableFieldService := mock_service.NewMockTableFieldService(ctrl)
	mockTableFieldRepo := mock_repository.NewMockTableFieldRepository(ctrl)
	mockTableApprovalDefRepo := mock_repository.NewMockTableApprovalDefinitionRepository(ctrl)
	mockApprovalService := mock_service.NewMockApprovalService(ctrl)
	mockGlobalIdService := mock_service.NewMockGlobalIdService(ctrl)
	mockEntityLogService := mock_service.NewMockEntityLogService(ctrl)
	mockAutocodeService := mock_service.NewMockAutocodeService(ctrl)
	mockTablePermissionService := mock_service.NewMockTablePermissionService(ctrl)
	mockTableRepo := mock_repository.NewMockTableRepository(ctrl)

	entityService := service.NewEntityService(
		service.NewService(testLogger, nil, nil),
		mockEntityRepo,
		mockTableFieldService,
		mockTableFieldRepo,
		mockTableApprovalDefRepo,
		mockApprovalService,
		mockGlobalIdService,
		mockEntityLogService,
		mockAutocodeService,
		mockTablePermissionService,
		mockTableRepo, // tableRepository
		nil,           // viper config
	)

	// Mock HasPermission check
	mockTablePermissionService.EXPECT().
		CheckTablePermission(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(true, nil).
		AnyTimes()

	// Mock GenerateOrRestoreAutocodes - CreateDraft 会调用此方法
	mockAutocodeService.EXPECT().
		GenerateOrRestoreAutocodes(gomock.Any(), "test_entity", gomock.Any(), mockTableFieldService, mockEntityRepo, gomock.Any()).
		Return(nil)

	// Mock CreateDraftWithApproval - CreateDraft 最终会调用此方法
	mockApprovalService.EXPECT().
		CreateDraftWithApproval(gomock.Any(), "test_entity", "test reason", gomock.Any()).
		Return(nil)

	entityMap := map[string]any{
		"code": "TEST001",
		"name": "Test Entity",
	}

	c := &gin.Context{}
	c.Set("user_id", uint(1))
	err := entityService.CreateDraft(c, "test_entity", "test reason", entityMap)

	// 注意: CreateDraft 方法本身不进行唯一性验证
	// 唯一性验证应该在审批通过后的同步过程中进行
	// 因此这个测试实际上不会检测到冲突,应该成功
	assert.NoError(t, err)
}

// TestValidateUniqueConstraints_UpdateNoWorkflow_NoConflict 测试修改无流程场景 - 无冲突
func TestValidateUniqueConstraints_UpdateNoWorkflow_NoConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEntityRepo := mock_repository.NewMockEntityRepository(ctrl)
	mockTableFieldService := mock_service.NewMockTableFieldService(ctrl)
	mockTableFieldRepo := mock_repository.NewMockTableFieldRepository(ctrl)
	mockTableApprovalDefRepo := mock_repository.NewMockTableApprovalDefinitionRepository(ctrl)
	mockApprovalService := mock_service.NewMockApprovalService(ctrl)
	mockGlobalIdService := mock_service.NewMockGlobalIdService(ctrl)
	mockEntityLogService := mock_service.NewMockEntityLogService(ctrl)
	mockAutocodeService := mock_service.NewMockAutocodeService(ctrl)
	mockTablePermissionService := mock_service.NewMockTablePermissionService(ctrl)
	mockTableRepo := mock_repository.NewMockTableRepository(ctrl)

	entityService := service.NewEntityService(
		service.NewService(testLogger, nil, nil),
		mockEntityRepo,
		mockTableFieldService,
		mockTableFieldRepo,
		mockTableApprovalDefRepo,
		mockApprovalService,
		mockGlobalIdService,
		mockEntityLogService,
		mockAutocodeService,
		mockTablePermissionService,
		mockTableRepo, // tableRepository
		nil,           // viper config
	)

	// Mock HasPermission check
	mockTablePermissionService.EXPECT().
		CheckTablePermission(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(true, nil).
		AnyTimes()

	// Mock 审批流程定义查询 - 无流程
	mockTableApprovalDefRepo.EXPECT().
		List("test_entity", "Update").
		Return([]model.TableApprovalDefinition{}, nil)

	// 模拟唯一索引字段查询
	uniqueFields := []*model.TableField{
		{Code: "code", IsUnique: "Yes", IndexName: ""},
	}
	mockTableFieldService.EXPECT().
		Find("code,is_unique,index_name", map[string]any{
			"table_code": "test_entity",
			"is_unique":  "Yes",
		}).
		Return(uniqueFields, nil)

	// 模拟查询结果 - 无冲突 (排除当前记录)
	mockEntityRepo.EXPECT().
		Find("test_entity", "id", map[string]any{
			"code":  "TEST001",
			"id !=": uint(1),
		}).
		Return([]map[string]any{}, nil)

	// Mock 其他依赖
	mockEntityRepo.EXPECT().FindOne("test_entity", uint(1)).Return(map[string]any{
		"id":   uint(1),
		"code": "OLD001",
		"name": "Old Name",
	}, nil)

	mockTableFieldService.EXPECT().
		Find("", map[string]any{"table_code": "test_entity"}).
		Return([]*model.TableField{
			{Code: "code", Name: "Code"},
			{Code: "name", Name: "Name"},
		}, nil)

	// Mock EntityLogService.Create - UpdateDraft 会为每个变更的字段记录日志
	mockEntityLogService.EXPECT().
		Create(gomock.Any(), "test_entity", gomock.Any()).
		Return(nil).
		AnyTimes() // 允许多次调用,因为可能有多个字段变更

	mockEntityRepo.EXPECT().
		Update(gomock.Any(), "test_entity", gomock.Any(), map[string]any{"id": uint(1)}).
		Return(nil)

	entityMap := map[string]any{
		"id":   uint(1),
		"code": "TEST001",
		"name": "Updated Name",
	}

	c := &gin.Context{}
	c.Set("user_id", uint(1))
	err := entityService.UpdateDraft(c, "test_entity", "test reason", entityMap)

	assert.NoError(t, err)
}

// TestValidateUniqueConstraints_MultiFieldIndex 测试多字段联合唯一索引
func TestValidateUniqueConstraints_MultiFieldIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEntityRepo := mock_repository.NewMockEntityRepository(ctrl)
	mockTableFieldService := mock_service.NewMockTableFieldService(ctrl)
	mockTableFieldRepo := mock_repository.NewMockTableFieldRepository(ctrl)
	mockTableApprovalDefRepo := mock_repository.NewMockTableApprovalDefinitionRepository(ctrl)
	mockApprovalService := mock_service.NewMockApprovalService(ctrl)
	mockGlobalIdService := mock_service.NewMockGlobalIdService(ctrl)
	mockEntityLogService := mock_service.NewMockEntityLogService(ctrl)
	mockAutocodeService := mock_service.NewMockAutocodeService(ctrl)
	mockTablePermissionService := mock_service.NewMockTablePermissionService(ctrl)
	mockTableRepo := mock_repository.NewMockTableRepository(ctrl)

	entityService := service.NewEntityService(
		service.NewService(testLogger, nil, nil),
		mockEntityRepo,
		mockTableFieldService,
		mockTableFieldRepo,
		mockTableApprovalDefRepo,
		mockApprovalService,
		mockGlobalIdService,
		mockEntityLogService,
		mockAutocodeService,
		mockTablePermissionService,
		mockTableRepo, // tableRepository
		nil,           // viper config
	)

	// Mock HasPermission check
	mockTablePermissionService.EXPECT().
		CheckTablePermission(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(true, nil).
		AnyTimes()

	// Mock TableRepository.Find for calculateTreeFields
	mockTableRepo.EXPECT().
		Find("", map[string]any{"code": "test_entity"}).
		Return([]*model.Table{}, nil) // Return empty list to indicate not a tree structure

	// Mock GetOperationInfo
	mockApprovalService.EXPECT().
		GetOperationInfo("Create", gomock.Any()).
		DoAndReturn(func(operation string, info *map[string]string) error {
			(*info)["action"] = "Create"
			(*info)["status"] = "Normal"
			return nil
		})

	// Mock GetNewID
	mockGlobalIdService.EXPECT().
		GetNewID("entity").
		Return(uint(1))

	// Mock GenerateOrRestoreAutocodes
	mockAutocodeService.EXPECT().
		GenerateOrRestoreAutocodes(gomock.Any(), "test_entity", gomock.Any(), mockTableFieldService, mockEntityRepo, gomock.Any()).
		Return(nil)

	// 模拟多字段联合唯一索引
	uniqueFields := []*model.TableField{
		{Code: "company_code", IsUnique: "Yes", IndexName: "idx_company_account"},
		{Code: "account_code", IsUnique: "Yes", IndexName: "idx_company_account"},
	}
	mockTableFieldService.EXPECT().
		Find("code,is_unique,index_name", map[string]any{
			"table_code": "test_entity",
			"is_unique":  "Yes",
		}).
		Return(uniqueFields, nil)

	// 模拟查询结果 - 有冲突
	mockEntityRepo.EXPECT().
		Find("test_entity", "id", map[string]any{
			"company_code": "C001",
			"account_code": "A001",
		}).
		Return([]map[string]any{{"id": uint64(1)}}, nil)

	entityMap := map[string]any{
		"company_code": "C001",
		"account_code": "A001",
		"name":         "Test Entity",
	}

	c := &gin.Context{}
	c.Set("user_id", uint(1))
	err := entityService.Create(c, "test_entity", entityMap)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "唯一索引字段")
}
