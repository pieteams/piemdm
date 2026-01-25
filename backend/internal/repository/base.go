// Package repository provides the data access layer for the PieMDM system.
// It contains repository interfaces and implementations that handle database
// operations, data queries, and persistence logic using GORM.
package repository

import (
	"gorm.io/gorm"
)

// ====================
// insert(array $data)
// update($data)
// save($id = null)
// put($key, $value)
// count($column_name = '*', $is_distinct = false)
// select($fields = ”)
// find($id = null)
// query($sql, $fetchAll = true, $fetchMode = PDO::FETCH_ASSOC)
// ====================
// execute($sql, array $sqldata = array())
// fetchAll($sql, array $sqldata = array(), $mode = PDO::FETCH_ASSOC)
// fetchRow($sql, array $sqldata = array(), $mode = PDO::FETCH_ASSOC)
// fetchCol($sql, array $sqldata = array())
// fetchOne($sql, array $sqldata = array(), $column_number = 0)
// fetchMap($sql, array $sqldata = array(), array $key_fields = array(), array $val_fields = array(), $mode = PDO::FETCH_ASSOC)
// insert($table, array $data, $is_return_id = false)
// replace($table, array $data)
// insertIgnore($table, array $data, $is_return_id = false)
// ignoreInsert($table, array $data, $is_return_id = false)
// insertUpdate($table, array $in_data, array $up_data)
// insertMulti($table, array $data, array $fields = array(), $type = 'insert')
// multiInsert($table, array $data, array $fields = array(), $type = 'insert')
// insertUpdateMulti($table, array $in_data, array $fields, array $up_data)
// multiInsertUpadate($table, array $in_data, array $fields, array $up_data)
// update($table, array $data, $condition, array $condata = array())
// updateMulti($table, array $data, $index_field, $condition = ”, array $condata = array())
// delete($table, $condition, array $condata = array())
// lockTable($table_name, $type = 1)
// unlockTable()
// lastInsertId()
// rowCount()
// begin()
// commit()
// rollback()
// ====================
// TODO 基础模型，传实例就行，动态模型，传模型和实例数据(是否可以简化)
type Base interface {
	Create(instance any) error
	Save(instance any) error
	Updates(model any, instance any) error
	DeleteByWhere(model, where any) (count int64, err error)
	DeleteById(model any, id uint) error
	DeleteByIds(model any, ids []uint) (count int64, err error)
	First(instance any, preloads []string, selects ...string) error
	FirstById(out any, id uint) error
	Find(model any, out any, sel string, where map[string]any, orders ...string) error
	FindPage(model any, out any, pageIndex, pageSize int, totalCount *int64, where map[string]any, preloads []string, orders ...string) error
	FindPageWithScopes(model any, out any, pageIndex, pageSize int, totalCount *int64, where map[string]any, preloads []string, scopes []func(*gorm.DB) *gorm.DB, orders ...string) error
	GetEntityWithPage(table string, model any, out any, pageIndex, pageSize int, totalCount *int64, where any, preloads []string, orders ...string) error
	PluckList(model, where any, out any, fieldName string) error
	GetTransaction() *gorm.DB
}

type base struct {
	*Repository
}

func NewBaseRepository(repository *Repository) Base {
	return &base{
		Repository: repository,
	}
}

// Create 创建数据
func (r *base) Create(instance any) error {
	return r.db.Create(instance).Error
}

// Save 保存数据
func (r *base) Save(instance any) error {
	return r.db.Save(instance).Error
}

// // Updates 更新数据
func (r *base) Updates(model any, instance any) error {
	return r.db.Model(model).Updates(instance).Error
}

// DeleteByWhere 根据条件删除数据
func (r *base) DeleteByWhere(model, where any) (count int64, err error) {
	db := r.db.Where(where).Delete(model)
	err = db.Error
	if err != nil {
		r.logger.Error("删除数据出错", "err", err)
		return
	}
	count = db.RowsAffected
	return
}

// DeleteById 根据id删除数据
func (r *base) DeleteById(model any, id uint) error {
	return r.db.Where("id=?", id).Delete(model).Error
}

// DeleteByIds 根据多个id删除多个数据
func (r *base) DeleteByIds(model any, ids []uint) (count int64, err error) {
	db := r.db.Where("id in (?)", ids).Delete(model)
	err = db.Error
	if err != nil {
		r.logger.Error("删除多个数据出错", "err", err)
		return
	}
	count = db.RowsAffected
	return
}

// First 根据条件获取一个数据
func (r *base) First(instance any, preloads []string, selects ...string) error {
	db := r.db.Where(instance)

	if len(preloads) > 0 {
		for _, pre := range preloads {
			db = db.Preload(pre)
		}
	}
	if len(selects) > 0 {
		for _, sel := range selects {
			db = db.Select(sel)
		}
	}
	return db.First(instance).Error
}

// FirstByID 根据条件获取一个数据
func (r *base) FirstById(out any, id uint) error {
	return r.db.Unscoped().First(out, id).Error
}

// GetPages 分页返回数据
func (r *base) Find(model any, out any, sel string, where map[string]any, orders ...string) error {

	db := r.db.Model(model).Where(model)
	db = db.Unscoped().Where(where)
	if sel != "" {
		db = db.Select(sel)
	}
	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	}
	err := db.Find(out).Error

	return err
}

// GetPages 分页返回数据
func (r *base) FindPage(model any, out any, pageIndex, pageSize int, totalCount *int64, where map[string]any, preloads []string, orders ...string) error {
	// build Condition string
	conditionString, values, _ := BuildCondition(where)

	db := r.db.Model(model).Where(model)
	// db = db.Where(map[string]any{"ID": 294})
	// db = db.Where(where)
	db = db.Unscoped().Where(conditionString, values...)

	if len(preloads) > 0 {
		for _, pre := range preloads {
			db = db.Preload(pre)
		}
	}
	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	}
	err := db.Count(totalCount).Error
	if err != nil {
		r.logger.Error("查询总数出错", "err", err)
		return err
	}
	if *totalCount == 0 {
		return nil
	}

	data := db.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(out).Error
	// r.logger.Debug("FindPage777", "data", data)

	return data
}

// FindPageWithScopes 分页返回数据，支持 GORM Scopes
func (r *base) FindPageWithScopes(model any, out any, pageIndex, pageSize int, totalCount *int64, where map[string]any, preloads []string, scopes []func(*gorm.DB) *gorm.DB, orders ...string) error {
	// build Condition string
	conditionString, values, _ := BuildCondition(where)

	db := r.db.Model(model).Where(model)
	// Apply Scopes
	if len(scopes) > 0 {
		db = db.Scopes(scopes...)
	}

	db = db.Unscoped().Where(conditionString, values...)

	if len(preloads) > 0 {
		for _, pre := range preloads {
			db = db.Preload(pre)
		}
	}
	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	}
	err := db.Count(totalCount).Error
	if err != nil {
		r.logger.Error("查询总数出错", "err", err)
		return err
	}
	if *totalCount == 0 {
		return nil
	}

	data := db.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(out).Error

	return data
}

// GetPages 分页返回数据
// TODO 弃用
func (r *base) GetEntityWithPage(table string, model any, out any, pageIndex, pageSize int, totalCount *int64, where any, preloads []string, orders ...string) error {
	// r.logger.Info("Base where", "where", where)
	// b.db.AutoMigrate(model)
	// db := r.db.Model(model).Where(where)

	db := r.db.Table(table).Model(model).Model(where)
	db = db.Where(where)

	if len(preloads) > 0 {
		for _, pre := range preloads {
			db = db.Preload(pre)
		}
	}
	if len(orders) > 0 {
		for _, order := range orders {
			db = db.Order(order)
		}
	}
	err := db.Count(totalCount).Error
	if err != nil {
		r.logger.Info("查询总数出错", "err", err)
		return err
	}
	if *totalCount == 0 {
		return nil
	}

	// data := db.Offset((pageIndex - 1) * pageSize).Limit(pageSize).Find(&out).Error
	data := db.Find(&out).Error
	r.logger.Info("Base data", "out", out)
	// for _, entity := range entities {
	// 	r.logger.Info("entity dao", "entity", entity)
	// }

	return data
}

// PluckList 查询 model 中的一个列作为切片
func (r *base) PluckList(model, where any, out any, fieldName string) error {
	return r.db.Model(model).Where(where).Pluck(fieldName, out).Error
}

// GetTransaction 获取事务
func (r *base) GetTransaction() *gorm.DB {
	return r.db.Begin()
}
