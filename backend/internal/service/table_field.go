package service

import (
	"fmt"
	"sort"
	"strings"

	"piemdm/internal/constants"
	"piemdm/internal/model"
	"piemdm/internal/repository"

	"github.com/gin-gonic/gin"
)

type TableFieldService interface {
	Get(id uint) (*model.TableField, error)
	Find(sel string, where map[string]any) ([]*model.TableField, error)
	List(page, pageSize int, total *int64, where map[string]any) ([]*model.TableField, error)
	Create(c *gin.Context, tableField *model.TableField) error
	BatchCreate(c *gin.Context, fields []*model.TableField) error
	Update(c *gin.Context, tableField *model.TableField) error
	BatchUpdate(c *gin.Context, ids []uint, tableField *model.TableField) error
	Delete(c *gin.Context, id uint) (*model.TableField, error)
	BatchDelete(c *gin.Context, ids []uint) error
	Public(c *gin.Context, tableCode string) error
	GetTableFields(tableCode string) ([]*model.FieldMetadata, error)
	GetTableOptions(tableCode string, filter map[string]any) ([]map[string]any, error)
}

type tableFieldService struct {
	*Service
	tableFieldRepository repository.TableFieldRepository
	tableRepository      repository.TableRepository
}

func NewTableFieldService(service *Service, tableFieldRepository repository.TableFieldRepository, tableRepository repository.TableRepository) TableFieldService {
	return &tableFieldService{
		Service:              service,
		tableFieldRepository: tableFieldRepository,
		tableRepository:      tableRepository,
	}
}

func (s *tableFieldService) Get(id uint) (*model.TableField, error) {
	return s.tableFieldRepository.FindOne(id)
}

func (s *tableFieldService) Find(sel string, where map[string]any) ([]*model.TableField, error) {
	return s.tableFieldRepository.Find(sel, where)
}

func (s *tableFieldService) List(page, pageSize int, total *int64, where map[string]any) ([]*model.TableField, error) {
	return s.tableFieldRepository.FindPage(page, pageSize, total, where)
}

// applyPreset 应用字段类型预设配置
func (s *tableFieldService) applyPreset(field *model.TableField) error {
	// 如果没有指定 field_type，跳过预设应用
	if field.FieldType == "" {
		return nil
	}

	preset, ok := constants.GetFieldPreset(field.FieldType)
	if !ok {
		return fmt.Errorf("不支持的字段类型: %s", field.FieldType)
	}

	// 强制使用预设的数据类型和长度
	field.Type = preset.DataType
	field.Length = preset.Length

	// 初始化 Options
	if field.Options == nil {
		field.Options = &model.FieldOptions{}
	}

	// 补全 UI 配置
	if field.Options.UI == nil {
		field.Options.UI = preset.UI
	} else {
		// 如果用户没有指定 widget，使用预设的
		if field.Options.UI.Widget == "" {
			field.Options.UI.Widget = preset.UI.Widget
		}
		// 合并 widgetProps
		if preset.UI != nil && preset.UI.WidgetProps != nil {
			if field.Options.UI.WidgetProps == nil {
				field.Options.UI.WidgetProps = preset.UI.WidgetProps
			} else {
				// 合并预设的 widgetProps（用户配置优先）
				for k, v := range preset.UI.WidgetProps {
					if _, exists := field.Options.UI.WidgetProps[k]; !exists {
						field.Options.UI.WidgetProps[k] = v
					}
				}
			}
		}
	}

	// 补全验证规则
	if field.Options.Validation == nil {
		field.Options.Validation = preset.Validation
	} else {
		// 合并预设的验证规则（用户配置优先）
		if preset.Validation != nil {
			if field.Options.Validation.Format == "" {
				field.Options.Validation.Format = preset.Validation.Format
			}
			if field.Options.Validation.Pattern == "" {
				field.Options.Validation.Pattern = preset.Validation.Pattern
			}
			if field.Options.Validation.Validator == "" {
				field.Options.Validation.Validator = preset.Validation.Validator
			}
			if field.Options.Validation.Max == nil && preset.Validation.Max != nil {
				field.Options.Validation.Max = preset.Validation.Max
			}
			if field.Options.Validation.Min == nil && preset.Validation.Min != nil {
				field.Options.Validation.Min = preset.Validation.Min
			}
		}
	}

	// 应用 Precision 和 Scale（仅 decimal 类型）
	if preset.Precision > 0 {
		precision := preset.Precision
		if field.Options.Validation == nil {
			field.Options.Validation = &model.FieldValidation{}
		}
		field.Options.Validation.Precision = &precision
	}
	if preset.Scale > 0 {
		scale := preset.Scale
		if field.Options.Validation == nil {
			field.Options.Validation = &model.FieldValidation{}
		}
		field.Options.Validation.Scale = &scale
	}

	return nil
}

func (s *tableFieldService) Create(c *gin.Context, tableField *model.TableField) error {
	// 检查字段code是否与系统字段冲突
	if constants.IsSystemFieldCode(tableField.Code) {
		return fmt.Errorf("字段code '%s' 与系统字段冲突,请使用其他名称", tableField.Code)
	}

	// 应用字段类型预设
	if err := s.applyPreset(tableField); err != nil {
		return err
	}

	return s.tableFieldRepository.Create(c, tableField)
}

func (s *tableFieldService) Update(c *gin.Context, tableField *model.TableField) error {
	return s.tableFieldRepository.Update(c, tableField)
}

func (s *tableFieldService) BatchUpdate(c *gin.Context, ids []uint, tableField *model.TableField) error {
	return s.tableFieldRepository.BatchUpdate(c, ids, tableField)
}

func (s *tableFieldService) Delete(c *gin.Context, id uint) (*model.TableField, error) {
	return s.tableFieldRepository.Delete(c, id)
}

func (s *tableFieldService) BatchDelete(c *gin.Context, ids []uint) error {
	return s.tableFieldRepository.BatchDelete(c, ids)
}

func (s *tableFieldService) Public(c *gin.Context, tableCode string) error {
	// 检查表的 DisplayMode,如果是 Tree 则自动创建树形字段
	tables, err := s.tableRepository.Find("", map[string]any{"code": tableCode, "status": "Normal"})
	if err == nil && len(tables) > 0 && tables[0].DisplayMode == "Tree" {
		// 检查是否已经存在树形字段
		existingFields, _ := s.tableFieldRepository.Find("", map[string]any{
			"table_code": tableCode,
			"code":       "parent_id",
			"status":     "Normal",
		})

		// 如果不存在 parent_id 字段,则创建树形字段
		if len(existingFields) == 0 {
			if err := s.createTreeFields(c, tableCode); err != nil {
				return fmt.Errorf("failed to create tree fields: %w", err)
			}
		}
	}

	where := make(map[string]any)
	// 生成 entity 结构
	tableName := fmt.Sprintf("t_%s", tableCode)
	entity := s.tableFieldRepository.BuildEntity(tableCode)

	// 生成 entity 的 draft 结构
	tableCodeDraft := fmt.Sprintf("%s_draft", tableCode)
	where["table_code"] = tableCodeDraft
	draftTableName := fmt.Sprintf("t_%s_draft", tableCode)
	draftEntity := s.tableFieldRepository.BuildEntity(tableCodeDraft)

	// // 获取 entity 的 修改log 表名
	logTableName := fmt.Sprintf("t_%s_log", tableCode)

	// 生成 entity 表
	// ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
	if err := s.tableFieldRepository.Public(tableName, entity); err != nil {
		panic(err)
	}

	// .source.Source.DB().Table(tableName).Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci").AutoMigrate(entity); err != nil {
	// 	panic(err)
	// }

	// 生成 entity 的 draft 表
	// if err := s.Dao.Base.Source.DB().Table(tableDraft).Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci").AutoMigrate(draft); err != nil {
	// 	panic(err)
	// }
	if err := s.tableFieldRepository.Public(draftTableName, draftEntity); err != nil {
		panic(err)
	}

	// // 生成 entity 的 修改日志 表
	// if err := s.Dao.Base.Source.DB().Table(tableLog).Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci").AutoMigrate(model.EntityLog{}); err != nil {
	// 	panic(err)
	// }
	if err := s.tableFieldRepository.Public(logTableName, model.EntityLog{}); err != nil {
		panic(err)
	}

	// return s.tableFieldRepository.Public(tableCode)
	return nil
}

// createTreeFields 为树形表自动创建必需的字段
func (s *tableFieldService) createTreeFields(c *gin.Context, tableCode string) error {
	// 获取表的现有字段,确定 labelField
	existingFields, err := s.tableFieldRepository.Find("", map[string]any{
		"table_code": tableCode,
		"status":     "Normal",
	})
	if err != nil {
		return err
	}

	// 动态确定 labelField: 优先 name → code → id
	labelField := "id"
	for _, f := range existingFields {
		if f.Code == "name" {
			labelField = "name"
			break
		}
		if f.Code == "code" {
			labelField = "code"
		}
	}

	// parent_id 字段 - 配置 relation 实现下拉选择本表数据
	parentIdField := &model.TableField{
		TableCode: tableCode,
		Code:      "parent_id",
		Name:      "父节点",
		Type:      "Integer",
		FieldType: "select",
		Length:    11,
		Required:  "No",
		IsShow:    "Yes",
		IsFilter:  "No",
		Options: &model.FieldOptions{
			UI: &model.FieldUI{
				Widget: "Select",
			},
			Relation: &model.FieldRelation{
				Target:     tableCode,
				LabelField: labelField,
				ValueField: "id",
			},
		},
		Sort:   1,
		Status: "Normal",
	}

	// level 字段 - 系统自动维护,不显示
	levelField := &model.TableField{
		TableCode: tableCode,
		Code:      "level",
		Name:      "层级",
		Type:      "Integer",
		FieldType: "number",
		Length:    11,
		Required:  "No",
		IsShow:    "No",
		IsFilter:  "No",
		Sort:      2,
		Status:    "Normal",
	}

	// path 字段 - 系统自动维护,不显示
	pathField := &model.TableField{
		TableCode: tableCode,
		Code:      "path",
		Name:      "路径",
		Type:      "Text",
		FieldType: "text",
		Length:    512,
		Required:  "No",
		IsShow:    "No",
		IsFilter:  "No",
		Sort:      3,
		Status:    "Normal",
	}

	// 批量创建字段
	fields := []*model.TableField{parentIdField, levelField, pathField}
	for _, field := range fields {
		err = s.tableFieldRepository.Create(c, field)
		if err != nil {
			return fmt.Errorf("failed to create field %s: %w", field.Code, err)
		}
	}

	return nil
}

// BatchCreate 批量创建字段
func (s *tableFieldService) BatchCreate(c *gin.Context, fields []*model.TableField) error {
	// 验证所有字段
	for _, field := range fields {
		if constants.IsSystemFieldCode(field.Code) {
			return fmt.Errorf("字段code '%s' 与系统字段冲突,请使用其他名称", field.Code)
		}
	}

	// 批量创建字段
	for _, field := range fields {
		// 应用字段类型预设
		if err := s.applyPreset(field); err != nil {
			return err
		}

		if err := s.tableFieldRepository.Create(c, field); err != nil {
			return err
		}
	}

	return nil
}

// GetTableFields 获取表的所有字段（包括系统字段）
func (s *tableFieldService) GetTableFields(tableCode string) ([]*model.FieldMetadata, error) {
	var fields []*model.FieldMetadata

	// 1. 从table_field读取用户定义的业务字段
	baseTableCode := strings.TrimSuffix(tableCode, "_draft")

	// 查询条件支持原表或草稿表（兼容某些手动为草稿表定义字段的情况）
	userFields, err := s.tableFieldRepository.Find("*", map[string]any{
		"status":     "Normal",
		"table_code": []string{tableCode, baseTableCode},
	})
	if err != nil {
		return nil, err
	}

	for _, field := range userFields {
		fields = append(fields, &model.FieldMetadata{
			Code:      field.Code,
			Name:      field.Name,
			FieldType: field.FieldType, // 添加 FieldType
			Type:      field.Type,
			Length:    field.Length,
			Required:  field.Required == "Yes",
			IsSystem:  false,
			IsShow:    field.IsShow == "Yes",
			IsFilter:  field.IsFilter == "Yes",
			Sort:      int(field.Sort),
			Options:   field.Options, // 添加 Options 配置
		})
	}

	// 2. 添加系统字段元数据
	systemFields := s.getSystemFieldsMetadata(tableCode)
	fields = append(fields, systemFields...)

	// 3. 按sort排序
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Sort < fields[j].Sort
	})

	return fields, nil
}

// getSystemFieldsMetadata 获取系统字段元数据
func (s *tableFieldService) getSystemFieldsMetadata(tableCode string) []*model.FieldMetadata {
	systemFields := []*model.FieldMetadata{
		{Code: "id", Name: "ID", Type: "Number", IsSystem: true, IsShow: false, Sort: 0},
		{Code: "status", Name: "状态", Type: "Text", Length: 16, IsSystem: true, IsShow: true, Sort: 9990},
		{Code: "operation", Name: "操作类型", Type: "Text", Length: 16, IsSystem: true, IsShow: true, Sort: 9991},
		{Code: "action", Name: "动作类型", Type: "Text", Length: 4, IsSystem: true, IsShow: true, Sort: 9992},
		{Code: "send_status", Name: "分发状态", Type: "Number", IsSystem: true, IsShow: true, Sort: 9993},
		{Code: "created_by", Name: "创建人", Type: "Text", Length: 64, IsSystem: true, IsShow: true, Sort: 9994},
		{Code: "updated_by", Name: "更新人", Type: "Text", Length: 64, IsSystem: true, IsShow: true, Sort: 9995},
		{Code: "created_at", Name: "创建时间", Type: "Date", IsSystem: true, IsShow: true, Sort: 9996},
		{Code: "updated_at", Name: "更新时间", Type: "Date", IsSystem: true, IsShow: true, Sort: 9997},
		{Code: "deleted_at", Name: "删除时间", Type: "Date", IsSystem: true, IsShow: true, Sort: 9998},
	}

	// 如果是draft表，添加流程字段
	if strings.HasSuffix(tableCode, "_draft") {
		draftFields := []*model.FieldMetadata{
			{Code: "entity_id", Name: "实体ID", Type: "Number", IsSystem: true, IsShow: false, Sort: 9987},
			{Code: "approval_code", Name: "审批编码", Type: "Text", Length: 64, IsSystem: true, IsShow: false, Sort: 9988},
			{Code: "draft_status", Name: "草稿状态", Type: "Text", Length: 16, IsSystem: true, IsShow: false, Sort: 9989},
		}
		systemFields = append(draftFields, systemFields...)
	}

	return systemFields
}

// GetTableOptions 获取表的选项列表（用于关联字段下拉）
func (s *tableFieldService) GetTableOptions(tableCode string, filter map[string]any) ([]map[string]any, error) {
	return s.tableFieldRepository.GetTableOptions(tableCode, filter)
}
