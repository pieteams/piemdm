package service

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"piemdm/internal/model"
	"piemdm/internal/repository"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
)

type EntityService interface {
	// Base CRUD
	Get(c *gin.Context, tableCode string, id uint) (map[string]any, error)
	List(c *gin.Context, tableCode string, page, pageSize int, total *int64, where map[string]any) ([]map[string]any, error)
	Create(c *gin.Context, tableCode string, entity any) error
	Update(c *gin.Context, tableCode string, entity any, where map[string]any) error
	Delete(c *gin.Context, tableCode string, reason string, id uint) error

	// Batch operations
	BatchUpdate(c *gin.Context, tableCode, reason string, ids []uint, entityMap map[string]any) error
	BatchDelete(c *gin.Context, tableCode, reason string, ids []uint) error

	// 历史与日志
	FindLogPage(c *gin.Context, tableCode string, page, pageSize int, total *int64, where map[string]any) ([]map[string]any, error)
	Find(c *gin.Context, tableCode, selectString string, where map[string]any) ([]map[string]any, error)

	// 草稿功能
	CreateDraft(c *gin.Context, tableCode, reason string, entityMap map[string]any) error
	UpdateDraft(c *gin.Context, tableCode, reason string, entityMap map[string]any) error

	// 导入导出
	Import(c *gin.Context, tableCode, reason, operation string, r io.Reader) error
	Export(c *gin.Context, tableCode, filter string, where map[string]any) (string, error)
	Template(c *gin.Context, tableCode, operation string) (string, error)

	// 其他
	BuildEntity(c *gin.Context, tableCode string) map[string]any
	GetEntitiesStatistics(c *gin.Context) ([]map[string]any, error)
}

type entityService struct {
	*Service
	entityRepository                  repository.EntityRepository
	tableFieldService                 TableFieldService
	tableFieldRepository              repository.TableFieldRepository
	tableApprovalDefinitionRepository repository.TableApprovalDefinitionRepository
	approvalService                   ApprovalService
	globalIdService                   GlobalIdService
	entityLogService                  EntityLogService
	autocodeService                   AutocodeService
	tablePermissionService            TablePermissionService // 新增
	tableRepository                   repository.TableRepository
	conf                              *viper.Viper
}

func NewEntityService(service *Service,
	entityRepository repository.EntityRepository,
	tableFieldService TableFieldService,
	tableFieldRepository repository.TableFieldRepository,
	tableApprovalDefinitionRepository repository.TableApprovalDefinitionRepository,
	approvalService ApprovalService,
	globalIdService GlobalIdService,
	entityLogService EntityLogService,
	autocodeService AutocodeService,
	tablePermissionService TablePermissionService, // 新增
	tableRepository repository.TableRepository,
	conf *viper.Viper) EntityService {
	return &entityService{
		Service:                           service,
		entityRepository:                  entityRepository,
		tableFieldService:                 tableFieldService,
		tableFieldRepository:              tableFieldRepository,
		tableApprovalDefinitionRepository: tableApprovalDefinitionRepository,
		approvalService:                   approvalService,
		globalIdService:                   globalIdService,
		entityLogService:                  entityLogService,
		autocodeService:                   autocodeService,
		tablePermissionService:            tablePermissionService, // 新增
		tableRepository:                   tableRepository,
		conf:                              conf,
	}
}

func (s *entityService) checkPermission(c *gin.Context, tableCode string) error {
	// 获取当前登录用户ID (JWT中间件设置的是string类型)
	userIdStr, exists := c.Get("user_id")
	if !exists {
		return fmt.Errorf("user not logged in")
	}

	// 将string转换为uint
	var userId uint
	switch v := userIdStr.(type) {
	case string:
		// 如果是字符串,转换为uint
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid user id")
		}
		userId = uint(id)
	case uint:
		// 如果已经是uint,直接使用
		userId = v
	default:
		return fmt.Errorf("invalid user id type")
	}

	if userId == 0 {
		return fmt.Errorf("user not logged in")
	}

	has, err := s.tablePermissionService.CheckTablePermission(c, userId, tableCode)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("permission denied for table: %s", tableCode)
	}
	return nil
}

func (s *entityService) Get(c *gin.Context, tableCode string, id uint) (map[string]any, error) {

	entityMap, err := s.entityRepository.FindOne(tableCode, id)
	if err != nil {
		return nil, err
	}

	// TODO 获取其他视图数据
	// // 获取关联数据
	// entityRelations, err2 := s.entityRepository.Find("entity_253", "*", map[string]any{
	// 	"VECode": entityMap["VECode"],
	// })
	// if err2 != nil {
	// 	s.logger.Error("EntityService-Get", "err2", err2)
	// }
	// entityMap["LFM1"] = entityRelations

	// entityRelations2, err3 := s.entityRepository.Find("entity_233", "*", map[string]any{
	// 	"VECode": entityMap["VECode"],
	// })
	// if err3 != nil {
	// 	s.logger.Error("EntityService-Get", "err3", err3)
	// }
	// entityMap["LFB1"] = entityRelations2

	return entityMap, err
}

func (s *entityService) List(c *gin.Context, tableCode string, page, pageSize int, total *int64, where map[string]any) ([]map[string]any, error) {
	if err := s.checkPermission(c, tableCode); err != nil {
		return nil, err
	}
	return s.entityRepository.FindPage(tableCode, page, pageSize, total, where)
}

func (s *entityService) FindLogPage(c *gin.Context, tableCode string, page, pageSize int, total *int64, where map[string]any) ([]map[string]any, error) {
	return s.entityRepository.FindLogPage(tableCode, page, pageSize, total, where)
}

func (s *entityService) Find(c *gin.Context, tableCode, selectString string, where map[string]any) ([]map[string]any, error) {
	return s.entityRepository.Find(tableCode, selectString, where)
}

func (s *entityService) CreateDraft(c *gin.Context, tableCode, reason string, entityMap map[string]any) error {
	if err := s.checkPermission(c, tableCode); err != nil {
		return err
	}

	// 设置操作类型,供 generateAutocodes 判断
	if _, ok := entityMap["operation"]; !ok {
		entityMap["operation"] = "Create"
	}

	// 生成自动编码字段
	if err := s.generateAutocodes(c, tableCode, entityMap); err != nil {
		return err
	}
	return s.approvalService.CreateDraftWithApproval(c, tableCode, reason, entityMap)
}

func (s *entityService) UpdateDraft(c *gin.Context, tableCode, reason string, entityMap map[string]any) error {
	if err := s.checkPermission(c, tableCode); err != nil {
		return err
	}
	// TODO 需要查询是否绑定了流程
	// 1. 如果没有绑定流程，直接更新数据表
	// 2. 如果不绑定了流程，则走审批流程
	// 判断是否关联审批流程，如果有审批流程则提交审批, 没有审批流程, 则直接保存数据
	operation := "Update"
	tableApprovalDefs, err := s.tableApprovalDefinitionRepository.List(tableCode, operation)
	if err != nil {
		s.logger.Error("err", "err", err)
	}

	// 如果有审批流程则提交审批, 没有审批流程, 则直接保存数据
	if len(tableApprovalDefs) <= 0 {
		// 不走审批流程，直接修改原数据

		// 验证唯一索引约束 (无审批流程)
		if err := s.validateUniqueConstraints(c, "Update", tableCode, entityMap, false); err != nil {
			return err
		}

		// 计算树形字段 (如果 parent_id 存在且变更)
		if _, ok := entityMap["parent_id"]; ok {
			if err := s.calculateTreeFields(c, tableCode, entityMap); err != nil {
				return err
			}
		}

		// 1. 获取原始数据用于比较变更
		id, ok := entityMap["id"].(uint)
		if !ok {
			return fmt.Errorf("invalid entity id type")
		}
		origin, err := s.entityRepository.FindOne(tableCode, id)
		if err != nil {
			return fmt.Errorf("获取原始数据失败: %v", err)
		}

		// 2. 获取表字段定义
		fieldWhere := map[string]any{}
		fieldWhere["table_code"] = tableCode
		tableFields, err := s.tableFieldService.Find("", fieldWhere)
		if err != nil {
			return fmt.Errorf("获取表字段失败: %v", err)
		}

		// 3. 构建字段Code到Name的映射
		fieldMap := make(map[string]string)
		for _, field := range tableFields {
			fieldMap[field.Code] = field.Name
		}

		// 4. 记录变更日志
		exclude := []string{
			"id",
			"table_code",
			"reason",
			"updated_at",
			"updated_by",
			"created_at",
			"created_by",
			"deleted_at",
		}

		for key := range entityMap {
			// 检查是否是排除字段
			excluded := false
			for _, v := range exclude {
				if key == v {
					excluded = true
					break
				}
			}

			if excluded {
				continue
			}

			// 比较原值和新值
			originValue := fmt.Sprintf("%v", origin[key])
			newValue := fmt.Sprintf("%v", entityMap[key])

			if originValue != newValue {
				// 创建变更日志
				entityLog := model.EntityLog{
					EntityID:     id,
					FieldCode:    key,
					FieldName:    fieldMap[key],
					BeforeUpdate: originValue,
					AfterUpdate:  newValue,
					Reason:       reason,                   // 修改原因
					UpdateBy:     c.GetString("user_name"), // 修改人
				}

				if err := s.entityLogService.Create(c, tableCode, &entityLog); err != nil {
					s.logger.Error("创建变更日志失败", "error", err, "field", key)
					// 日志记录失败不应该阻断更新流程
				}
			}
		}

		// 5. 更新 entity
		// 创建一个新的map,只包含应该更新的字段
		updateMap := make(map[string]any)
		for key, value := range entityMap {
			// 排除不应该更新到数据库的字段
			if key != "table_code" && key != "reason" && key != "id" {
				updateMap[key] = value
			}
		}

		whereMap := make(map[string]any)
		whereMap["id"] = entityMap["id"]
		return s.entityRepository.Update(c, tableCode, updateMap, whereMap)
	}

	// 验证唯一索引约束 (有审批流程)
	if err := s.validateUniqueConstraints(c, "Update", tableCode, entityMap, true); err != nil {
		return err
	}

	// 走审批流程，生成草稿
	return s.approvalService.UpdateDraftWithApproval(c, tableCode, reason, entityMap)

	// return nil
}

func (s *entityService) BatchUpdate(c *gin.Context, tableCode, reason string, ids []uint, entityMap map[string]any) error {
	if err := s.checkPermission(c, tableCode); err != nil {
		return err
	}
	// 获取操作类型
	operation, ok := entityMap["operation"].(string)
	if !ok {
		return fmt.Errorf("operation is required")
	}

	// 检查是否包含已删除的数据
	if operation == "BatchFreeze" || operation == "BatchUnfreeze" {
		where := map[string]any{
			"id in": ids,
		}
		// 只查询 status 字段
		entities, err := s.entityRepository.Find(tableCode, "status", where)
		if err != nil {
			return fmt.Errorf("查询数据状态失败: %v", err)
		}
		for _, ent := range entities {
			if status, ok := ent["status"].(string); ok && status == "Deleted" {
				return fmt.Errorf("包含已删除的数据，无法进行操作")
			}
		}
	}

	// 检查是否有审批流程定义
	tableApprovalDefs, err := s.tableApprovalDefinitionRepository.List(tableCode, operation)
	if err != nil {
		s.logger.Error("查询审批流程定义失败", "err", err)
		return err
	}

	// 如果没有审批流程,直接更新主表数据
	if len(tableApprovalDefs) == 0 {
		// 调用 GetOperationInfo 获取操作信息
		operationInfo := make(map[string]string)
		if err := s.approvalService.GetOperationInfo(operation, &operationInfo); err != nil {
			return err
		}

		// 设置必要字段 - 使用小写下划线,直接更新数据库
		entityMap["action"] = operationInfo["action"]
		entityMap["status"] = operationInfo["status"]
		entityMap["updated_by"] = c.GetString("user_name")

		// 获取表字段定义用于变更日志
		fieldWhere := map[string]any{}
		fieldWhere["table_code"] = tableCode
		tableFields, err := s.tableFieldService.Find("", fieldWhere)
		if err != nil {
			s.logger.Error("获取表字段失败", "error", err)
		}

		// 构建字段Code到Name的映射
		fieldMap := make(map[string]string)
		for _, field := range tableFields {
			fieldMap[field.Code] = field.Name
		}
		// 添加系统字段映射
		fieldMap["status"] = "状态"
		fieldMap["action"] = "操作"
		fieldMap["operation"] = "操作类型"

		// 为每个ID记录变更日志
		for _, id := range ids {
			// 获取原始数据
			origin, err := s.entityRepository.FindOne(tableCode, id)
			if err != nil {
				s.logger.Error("获取原始数据失败", "error", err, "id", id)
				continue
			}

			// 记录变更日志
			exclude := []string{
				"id",
				"table_code",
				"reason",
				"updated_at",
				"updated_by",
				"created_at",
				"created_by",
				"deleted_at",
			}

			for key := range entityMap {
				// 检查是否是排除字段
				excluded := false
				for _, v := range exclude {
					if key == v {
						excluded = true
						break
					}
				}

				if excluded {
					continue
				}

				// 比较原值和新值
				originValue := fmt.Sprintf("%v", origin[key])
				newValue := fmt.Sprintf("%v", entityMap[key])

				if originValue != newValue {
					// 创建变更日志
					entityLog := model.EntityLog{
						EntityID:     id,
						FieldCode:    key,
						FieldName:    fieldMap[key],
						BeforeUpdate: originValue,
						AfterUpdate:  newValue,
						Reason:       reason,
						UpdateBy:     c.GetString("user_name"),
					}

					if err := s.entityLogService.Create(c, tableCode, &entityLog); err != nil {
						s.logger.Error("创建变更日志失败", "error", err, "field", key, "id", id)
						// 日志记录失败不应该阻断更新流程
					}
				}
			}
		}

		// 直接更新主表
		return s.entityRepository.BatchUpdate(c, tableCode, ids, entityMap)
	}

	// 有审批流程,走审批流程
	return s.approvalService.UpdateByIdsWithApproval(c, tableCode, reason, ids, entityMap)
}

func (s *entityService) Delete(c *gin.Context, tableCode, reason string, id uint) error {
	if err := s.checkPermission(c, tableCode); err != nil {
		return err
	}
	return s.entityRepository.Delete(c, tableCode, id)
}

func (s *entityService) BatchDelete(c *gin.Context, tableCode, reason string, ids []uint) error {
	if err := s.checkPermission(c, tableCode); err != nil {
		return err
	}
	return s.entityRepository.BatchDelete(c, tableCode, ids)
}

func (s *entityService) Import(c *gin.Context, tableCode, reason, operation string, r io.Reader) error {
	if err := s.checkPermission(c, tableCode); err != nil {
		return err
	}
	// 检查是否有审批流程定义
	tableApprovalDefs, err := s.tableApprovalDefinitionRepository.List(tableCode, operation)
	if err != nil {
		s.logger.Error("查询审批流程定义失败", "err", err)
		return err
	}

	// 如果没有审批流程,直接导入到主表
	if len(tableApprovalDefs) == 0 {
		return s.importDirect(c, tableCode, reason, operation, r)
	}

	// 有审批流程,走审批流程
	return s.approvalService.ImportWithApproval(c, tableCode, reason, operation, r)
}

// importDirect 直接导入数据到主表(无审批流程)
func (s *entityService) importDirect(c *gin.Context, tableCode, reason, operation string, r io.Reader) error {
	// 获取操作信息
	operationInfo := make(map[string]string)
	if err := s.approvalService.GetOperationInfo(operation, &operationInfo); err != nil {
		return err
	}

	// Fetch field definitions for type mapping
	fieldWhere := map[string]any{"table_code": tableCode}
	tFields, err := s.tableFieldService.Find("code,field_type,type", fieldWhere)
	if err != nil {
		s.logger.Error("importDirect: fetch fields failed", "err", err)
		return err
	}
	fieldTypeMap := make(map[string]struct {
		uiType string
		dbType string
	})

	for _, f := range tFields {
		fieldTypeMap[f.Code] = struct{ uiType, dbType string }{
			uiType: f.FieldType,
			dbType: f.Type,
		}
	}

	// 打开 Excel 文件
	xlsx, err := excelize.OpenReader(r)
	if err != nil {
		return err
	}

	rows := xlsx.GetRows("Sheet1")
	var fields []string

	// 遍历所有行
	for irow, row := range rows {
		if irow == 0 {
			// 第一行是字段名
			fields = append(fields, row...)
		} else {
			// 构建实体数据
			entityMap := make(map[string]any)
			for index, cell := range row {
				if index >= len(fields) {
					break
				}
				fieldCode := fields[index]

				// 处理 ID 字段
				if fieldCode == "ID" || fieldCode == "id" {
					if num, err := strconv.Atoi(cell); err == nil {
						entityMap[fieldCode] = num
						continue
					}
				}

				// Handle empty strings based on field type
				if cell == "" {
					if types, ok := fieldTypeMap[fieldCode]; ok {
						uiType := strings.ToLower(types.uiType)
						dbType := strings.ToLower(types.dbType)

						isDateOrNum := false
						if strings.Contains(uiType, "date") ||
							strings.Contains(uiType, "time") ||
							strings.Contains(uiType, "int") ||
							strings.Contains(uiType, "number") ||
							strings.Contains(uiType, "decimal") ||
							strings.Contains(uiType, "float") ||
							strings.Contains(uiType, "double") {
							isDateOrNum = true
						}

						if !isDateOrNum {
							if strings.Contains(dbType, "date") ||
								strings.Contains(dbType, "time") ||
								strings.Contains(dbType, "int") ||
								strings.Contains(dbType, "decimal") ||
								strings.Contains(dbType, "float") ||
								strings.Contains(dbType, "double") {
								isDateOrNum = true
							}
						}

						if isDateOrNum {
							entityMap[fieldCode] = nil
							continue
						}
					}
				}

				entityMap[fieldCode] = cell
			}

			// 根据操作类型选择创建或更新
			if operation == "BatchCreate" {
				// 新增操作 - 使用snake_case以匹配数据库列名
				gid := s.globalIdService.GetNewID("entity")
				entityMap["id"] = gid
				entityMap["operation"] = operation
				entityMap["action"] = operationInfo["action"]
				entityMap["status"] = operationInfo["status"]
				// 生成 entity global id
				entityMap["entity_id"] = gid
				entityMap["created_by"] = c.GetString("user_name") // Assuming user_name is the creator
				entityMap["send_status"] = 0
				entityMap["updated_by"] = c.GetString("user_name")

				// 生成自动编码
				if err := s.generateAutocodes(c, tableCode, entityMap); err != nil {
					return err
				}

				if err := s.entityRepository.Create(c, tableCode, entityMap); err != nil {
					return err
				}
			} else if operation == "BatchUpdate" {
				// 更新操作 - 使用snake_case以匹配数据库列名
				if entityMap["id"] == nil {
					return fmt.Errorf("BatchUpdate requires id field")
				}

				// 获取ID
				var id uint
				switch v := entityMap["id"].(type) {
				case string:
					// 从Excel读取的可能是字符串,需要转换
					idInt, err := strconv.ParseUint(v, 10, 64)
					if err != nil {
						return fmt.Errorf("invalid id value: %s", v)
					}
					id = uint(idInt)
				case int:
					id = uint(v)
				case uint:
					id = v
				case float64:
					// Excel数字可能被解析为float64
					id = uint(v)
				default:
					return fmt.Errorf("invalid id type: %T", v)
				}

				// 1. 获取原始数据用于比较变更
				origin, err := s.entityRepository.FindOne(tableCode, id)
				if err != nil {
					s.logger.Error("获取原始数据失败", "error", err, "id", id)
					return err
				}

				// 2. 获取表字段定义
				fieldWhere := map[string]any{}
				fieldWhere["table_code"] = tableCode
				tableFields, err := s.tableFieldService.Find("", fieldWhere)
				if err != nil {
					s.logger.Error("获取表字段失败", "error", err)
					return fmt.Errorf("获取表字段失败: %v", err)
				}

				// 3. 构建字段Code到Name的映射
				fieldMap := make(map[string]string)
				for _, field := range tableFields {
					fieldMap[field.Code] = field.Name
				}

				// 4. 记录变更日志
				exclude := []string{
					"id",
					"table_code",
					"reason",
					"updated_at",
					"updated_by",
					"created_at",
					"created_by",
					"deleted_at",
					"action",
				}

				for key := range entityMap {
					// 检查是否是排除字段
					excluded := false
					for _, v := range exclude {
						if key == v {
							excluded = true
							break
						}
					}

					if excluded {
						continue
					}

					// 比较原值和新值
					originValue := fmt.Sprintf("%v", origin[key])
					newValue := fmt.Sprintf("%v", entityMap[key])

					if originValue != newValue {
						// 创建变更日志
						entityLog := model.EntityLog{
							EntityID:     id,
							FieldCode:    key,
							FieldName:    fieldMap[key],
							BeforeUpdate: originValue,
							AfterUpdate:  newValue,
							Reason:       reason,                   // 修改原因
							UpdateBy:     c.GetString("user_name"), // 修改人
						}

						if err := s.entityLogService.Create(c, tableCode, &entityLog); err != nil {
							s.logger.Error("创建变更日志失败", "error", err, "field", key)
							// 日志记录失败不应该阻断更新流程
						}
					}
				}

				// 5. 设置必要字段
				entityMap["action"] = operationInfo["action"]
				entityMap["status"] = operationInfo["status"]
				entityMap["updated_by"] = c.GetString("user_name")

				// 6. 构建 where 条件
				where := map[string]any{
					"id": id,
				}

				// 7. 删除 id 字段,避免更新 id
				delete(entityMap, "id")

				// 8. 更新数据
				if err := s.entityRepository.Update(c, tableCode, entityMap, where); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *entityService) Export(c *gin.Context, tableCode, filter string, where map[string]any) (string, error) {
	// get table fields
	stfWhere := make(map[string]any)
	stfWhere["status"] = "Normal"
	stfWhere["table_code"] = tableCode
	tableFields, err := s.tableFieldService.Find("*", stfWhere)
	if err != nil {
		s.logger.Error("tableFieldService.Find", "err", err)
	}

	sel := "*"
	// Get all data for export
	entities, _ := s.entityRepository.Find(tableCode, sel, where)

	// create xlsx file
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		return "", err
	}
	row := sheet.AddRow()

	var cell *xlsx.Cell
	cell = row.AddCell()
	// Style of Cell 00D5DAE9
	fill := xlsx.NewFill("solid", "00D0D0D0", "00000000")
	border := *xlsx.NewBorder("thin", "thin", "thin", "thin")
	border.LeftColor = "00000000"
	border.RightColor = "00000000"
	border.TopColor = "00000000"
	border.BottomColor = "00000000"
	// Style of Table Header
	styleHeader := xlsx.NewStyle()
	styleHeader.Fill = *fill
	styleHeader.Border = border
	// Style of Table body
	styleBody := xlsx.NewStyle()
	styleBody.Border = border
	// Set cell for id
	cell.SetStyle(styleHeader)
	cell.Value = "id"
	for _, field := range tableFields {
		cell = row.AddCell()
		cell.SetStyle(styleHeader)
		cell.Value = field.Code
	}

	for _, v := range entities {
		var values []string
		values = append(values, strconv.FormatUint(v["id"].(uint64), 10))
		for _, field := range tableFields {
			if v[field.Code] != nil {
				// 根据实际类型转换为字符串
				switch val := v[field.Code].(type) {
				case string:
					values = append(values, val)
				case int, int8, int16, int32, int64:
					values = append(values, fmt.Sprintf("%d", val))
				case uint, uint8, uint16, uint32, uint64:
					values = append(values, fmt.Sprintf("%d", val))
				case float32, float64:
					values = append(values, fmt.Sprintf("%v", val))
				case bool:
					values = append(values, fmt.Sprintf("%t", val))
				default:
					values = append(values, fmt.Sprintf("%v", val))
				}
			} else {
				values = append(values, "")
			}
		}

		row = sheet.AddRow()
		for _, value := range values {
			cell = row.AddCell()
			cell.SetStyle(styleBody)
			cell.Value = value
		}
	}
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.SetStyle(styleBody)
	cell.Value = "Export " + strconv.Itoa(len(entities)) + " pieces of data."

	time := strconv.Itoa(int(time.Now().Unix()))
	filename := tableCode + "-" + time + ".xlsx"

	fullPath := s.conf.GetString("app.runtime-root-path") + s.conf.GetString("app.export-save-path") + filename
	err = file.Save(fullPath)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (s *entityService) Template(c *gin.Context, tableCode, operation string) (string, error) {
	// get table fields
	stfWhere := make(map[string]any)
	stfWhere["status"] = "Normal"
	stfWhere["table_code"] = tableCode
	tableFields, err := s.tableFieldService.Find("*", stfWhere)
	if err != nil {
		s.logger.Error("tableFieldService.Find", "err", err)
	}

	// create xlsx file
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		return "", err
	}
	row := sheet.AddRow()

	var cell *xlsx.Cell
	// Style of Cell 00D5DAE9
	fill := xlsx.NewFill("solid", "00D0D0D0", "00000000")
	border := *xlsx.NewBorder("thin", "thin", "thin", "thin")
	border.LeftColor = "00000000"
	border.RightColor = "00000000"
	border.TopColor = "00000000"
	border.BottomColor = "00000000"
	// Style of Table Header
	styleHeader := xlsx.NewStyle()
	styleHeader.Fill = *fill
	styleHeader.Border = border
	// Style of Table body
	styleBody := xlsx.NewStyle()
	styleBody.Border = border
	if operation == "BatchUpdate" {
		// Set cell for id
		cell = row.AddCell()
		cell.SetStyle(styleHeader)
		cell.Value = "id"
	}
	for _, field := range tableFields {
		cell = row.AddCell()
		cell.SetStyle(styleHeader)
		cell.Value = field.Code
	}
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.SetStyle(styleBody)

	time := strconv.Itoa(int(time.Now().Unix()))
	filename := tableCode + "-create-template-" + time + ".xlsx"
	if operation == "BatchUpdate" {
		filename = tableCode + "-update-template-" + time + ".xlsx"
	}

	fullPath := s.conf.GetString("app.runtime-root-path") + s.conf.GetString("app.export-save-path") + filename
	err = file.Save(fullPath)
	if err != nil {
		return "", err
	}

	return filename, nil
}

// calculateTreeFields 计算树形结构的 level 和 path
func (s *entityService) calculateTreeFields(c *gin.Context, tableCode string, entityMap map[string]any) error {
	// 1. 检查表是否为树形结构
	tables, err := s.tableRepository.Find("", map[string]any{"code": tableCode})
	if err != nil || len(tables) == 0 || tables[0].DisplayMode != "Tree" {
		return nil
	}

	// 2. 获取 parent_id及其数据类型处理
	var parentId uint

	// Helper to extract uint from interface
	parseUint := func(val any) uint {
		if val == nil {
			return 0
		}
		// Try direct numeric assumptions first for performance
		switch v := val.(type) {
		case uint:
			return v
		case int:
			return uint(v)
		case float64:
			return uint(v)
		case uint64:
			return uint(v)
		case int64:
			return uint(v)
		}

		// Fallback: convert to string then parse
		// This handles string, json.Number, and other types
		s := fmt.Sprintf("%v", val)
		if s != "" {
			// Trim potential whitespace just in case
			s = strings.TrimSpace(s)
			if id, err := strconv.ParseUint(s, 10, 64); err == nil {
				return uint(id)
			}
			// Try float parsing if int parsing fails (e.g. "123.0")
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return uint(f)
			}
		}
		return 0
	}

	// Check various possible keys
	if v, ok := entityMap["parent_id"]; ok && v != nil {
		parentId = parseUint(v)
	} else if v, ok := entityMap["ParentId"]; ok && v != nil {
		parentId = parseUint(v)
	} else if v, ok := entityMap["ParentID"]; ok && v != nil {
		parentId = parseUint(v)
	}

	s.logger.Info("calculateTreeFields", "tableCode", tableCode, "parentId_extracted", parentId, "map_keys", fmt.Sprintf("%v", entityMap))

	// Get current ID
	var currentId uint
	if v, ok := entityMap["id"]; ok && v != nil {
		currentId = parseUint(v)
	}

	// 3. 计算 level 和 path
	if parentId == 0 {
		entityMap["level"] = 1
		entityMap["path"] = fmt.Sprintf("/%v/", currentId)
	} else {
		// 查询父节点
		parent, err := s.entityRepository.FindOne(tableCode, parentId)
		if err != nil {
			return fmt.Errorf("invalid parent_id: %v", err)
		}

		var pLevel int
		switch v := parent["level"].(type) {
		case int:
			pLevel = v
		case float64:
			pLevel = int(v)
		case string:
			pLevel, _ = strconv.Atoi(v)
		case int64:
			pLevel = int(v)
		}

		pPath := "/"
		if v, ok := parent["path"].(string); ok && v != "" {
			pPath = v
		} else if v, ok := parent["path"]; ok && v != nil {
			pPath = fmt.Sprintf("%v", v)
		}

		// Ensure pPath ends with / if not empty, or handled by logic
		// If parent path is like /1/2/, and we append 3/, result /1/2/3/
		// If parent path is empty (root created with old logic /), we expect /<pid>/

		// Fix for legacy data or root node logic assumption
		if pPath == "" || pPath == "/" {
			// If parent has no valid path, reconstruct it (best effort)
			pPath = fmt.Sprintf("/%v/", parentId)
		} else {
			// Ensure it ends with /
			if pPath[len(pPath)-1] != '/' {
				pPath += "/"
			}
		}

		entityMap["level"] = pLevel + 1
		entityMap["path"] = fmt.Sprintf("%s%v/", pPath, currentId)
	}
	s.logger.Info("calculateTreeFields_Result", "level", entityMap["level"], "path", entityMap["path"])
	return nil
}

// generateAutocodes 为 autocode 类型的字段生成编码
func (s *entityService) generateAutocodes(c *gin.Context, tableCode string, entityMap map[string]any) error {
	return s.autocodeService.GenerateOrRestoreAutocodes(c, tableCode, entityMap, s.tableFieldService, s.entityRepository, s.logger)
}

func (s *entityService) Create(c *gin.Context, tableCode string, entity any) error {
	if err := s.checkPermission(c, tableCode); err != nil {
		return err
	}
	// 如果传入的是 map,需要设置必要的字段
	if entityMap, ok := entity.(map[string]any); ok {
		// 设置操作类型
		operation := "Create"

		// 调用 GetOperationInfo 获取操作信息
		operationInfo := make(map[string]string)
		if err := s.approvalService.GetOperationInfo(operation, &operationInfo); err != nil {
			return err
		}

		// 生成全局唯一 ID
		gid := s.globalIdService.GetNewID("entity")

		// 设置必要字段 - 使用snake_case以匹配数据库列名
		entityMap["id"] = gid
		entityMap["operation"] = operation
		entityMap["action"] = operationInfo["action"]
		entityMap["status"] = operationInfo["status"]
		entityMap["send_status"] = 0
		entityMap["created_by"] = c.GetString("user_name")
		entityMap["updated_by"] = c.GetString("user_name")

		// 计算树形字段
		if err := s.calculateTreeFields(c, tableCode, entityMap); err != nil {
			return err
		}

		// 生成自动编码字段
		if err := s.generateAutocodes(c, tableCode, entityMap); err != nil {
			return err
		}

		// 验证唯一索引约束 (无审批流程)
		if err := s.validateUniqueConstraints(c, "Create", tableCode, entityMap, false); err != nil {
			return err
		}
	}

	return s.entityRepository.Create(c, tableCode, entity)
}

func (s *entityService) Update(c *gin.Context, tableCode string, entity any, where map[string]any) error {
	return s.entityRepository.Update(c, tableCode, entity, where)
}

func (s *entityService) BuildEntity(c *gin.Context, tableCode string) map[string]any {
	return s.tableFieldRepository.BuildEntity(tableCode)
}

// validateUniqueConstraints 验证唯一索引约束
// operation: "Create" 或 "Update"
// tableCode: 表代码
// entityMap: 实体数据
// hasWorkflow: 是否有审批流程
func (s *entityService) validateUniqueConstraints(
	c *gin.Context,
	operation string,
	tableCode string,
	entityMap map[string]any,
	hasWorkflow bool,
) error {
	// 1. 获取唯一索引字段
	uniqueFields, err := s.tableFieldService.Find("code,is_unique,index_name", map[string]any{
		"table_code": tableCode,
		"is_unique":  "Yes",
	})
	if err != nil {
		s.logger.Error("查询唯一索引字段失败", "error", err)
		return fmt.Errorf("查询唯一索引字段失败: %v", err)
	}

	// 如果没有唯一索引字段,直接返回
	if len(uniqueFields) == 0 {
		return nil
	}

	// 2. 构建唯一索引分组 map[indexName]map[fieldCode]fieldValue
	uniqueFieldsMap := make(map[string]map[string]any)
	for _, field := range uniqueFields {
		indexName := field.IndexName
		if indexName == "" {
			// 如果没有 index_name,使用字段 code 作为索引名
			indexName = field.Code
		}

		if _, exists := uniqueFieldsMap[indexName]; !exists {
			uniqueFieldsMap[indexName] = make(map[string]any)
		}
		uniqueFieldsMap[indexName][field.Code] = entityMap[field.Code]
	}

	// 3. 根据操作类型和是否有审批流程,选择查询的表
	targetTableCode := tableCode
	if hasWorkflow {
		targetTableCode = fmt.Sprintf("%s_draft", tableCode)
	}

	// 4. 遍历每个唯一索引组,检查是否存在冲突
	for _, fieldMap := range uniqueFieldsMap {
		where := make(map[string]any)

		// 添加唯一索引字段条件
		for fieldCode, fieldValue := range fieldMap {
			where[fieldCode] = fieldValue
		}

		// 如果是修改操作,需要排除当前记录
		if operation == "Update" {
			if entityID, ok := entityMap["id"]; ok {
				if hasWorkflow {
					// 有审批流程时,使用统一的检查方法
					// 检查是否存在 Pending 状态的草稿(通过 entity_id)
					if err := s.approvalService.CheckExistingActiveDraft(c, targetTableCode, entityID); err != nil {
						return err
					}
				} else {
					// 无审批流程时,排除当前记录
					where["id !="] = entityID
					exists, err := s.entityRepository.Find(targetTableCode, "id", where)
					if err != nil {
						s.logger.Error("查询唯一索引冲突失败", "error", err)
						return fmt.Errorf("查询唯一索引冲突失败: %v", err)
					}
					if len(exists) > 0 {
						// 构建友好的错误提示
						var fieldNames []string
						for fieldCode := range fieldMap {
							fieldNames = append(fieldNames, fieldCode)
						}
						return fmt.Errorf("唯一索引字段 [%s] 的值已存在,请修改后重试", fmt.Sprintf("%v", fieldNames))
					}
				}
			}
		} else {
			// 创建操作
			exists, err := s.entityRepository.Find(targetTableCode, "id", where)
			if err != nil {
				s.logger.Error("查询唯一索引冲突失败", "error", err)
				return fmt.Errorf("查询唯一索引冲突失败: %v", err)
			}
			if len(exists) > 0 {
				if hasWorkflow {
					return fmt.Errorf("该数据的唯一索引字段已在审批流程中,请等待审批完成后再提交")
				}
				// 构建友好的错误提示
				var fieldNames []string
				for fieldCode := range fieldMap {
					fieldNames = append(fieldNames, fieldCode)
				}
				return fmt.Errorf("唯一索引字段 [%s] 的值已存在,请修改后重试", fmt.Sprintf("%v", fieldNames))
			}
		}
	}

	return nil
}
func (s *entityService) GetEntitiesStatistics(c *gin.Context) ([]map[string]any, error) {
	// 注意：这是管理员统计接口，不需要检查单个表的权限

	// 1. 获取所有正常的动态表主体
	where := map[string]any{
		"status":     "Normal",
		"table_type": "Entity",
	}
	tables, err := s.tableRepository.Find("", where)
	if err != nil {
		return nil, err
	}

	var statsList []map[string]any
	for _, table := range tables {
		// 跳过编码为空的表，防止构造出非法的表名 (如 t_)
		if table.Code == "" {
			s.logger.Warn("跳过编码为空的表统计", "table_id", table.ID)
			continue
		}

		// 2. 获取该表的状态统计
		counts, err := s.entityRepository.GetStatisticsByStatus(table.Code)
		if err != nil {
			s.logger.Error("获取表统计失败", "table", table.Code, "err", err)
			continue
		}

		statsList = append(statsList, map[string]any{
			"code": table.Code,
			"name": table.Name,
			"statistics": map[string]int64{
				"Normal":  counts["Normal"],
				"Frozen":  counts["Frozen"],
				"Deleted": counts["Deleted"],
			},
		})
	}

	return statsList, nil
}
