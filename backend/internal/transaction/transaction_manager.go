package transaction

import (
	"context"
	"fmt"
	"sync"
	"time"

	"piemdm/pkg/log"

	"gorm.io/gorm"
)

// Transaction 事务接口
type Transaction interface {
	// GetDB 获取事务数据库连接
	GetDB() *gorm.DB

	// Commit 提交事务
	Commit() error

	// Rollback 回滚事务
	Rollback() error

	// IsActive 检查事务是否活跃
	IsActive() bool

	// GetID 获取事务ID
	GetID() string
}

// TransactionManager 事务管理器接口
type TransactionManager interface {
	// BeginTransaction 开始事务
	BeginTransaction(ctx context.Context) (Transaction, error)

	// ExecuteInTransaction 在事务中执行操作
	ExecuteInTransaction(ctx context.Context, fn func(tx Transaction) error) error

	// GetActiveTransactions 获取活跃事务数量
	GetActiveTransactions() int

	// CleanupExpiredTransactions 清理过期事务
	CleanupExpiredTransactions() error
}

// TransactionStatus 事务状态
type TransactionStatus string

const (
	TransactionStatusActive     TransactionStatus = "active"      // 活跃
	TransactionStatusCommitted  TransactionStatus = "committed"   // 已提交
	TransactionStatusRolledBack TransactionStatus = "rolled_back" // 已回滚
	TransactionStatusExpired    TransactionStatus = "expired"     // 已过期
)

// gormTransaction GORM事务实现
type gormTransaction struct {
	id        string
	db        *gorm.DB
	tx        *gorm.DB
	status    TransactionStatus
	startTime time.Time
	logger    *log.Logger
	mutex     sync.RWMutex
}

// NewGormTransaction 创建GORM事务
func NewGormTransaction(id string, db *gorm.DB, logger *log.Logger) *gormTransaction {
	return &gormTransaction{
		id:        id,
		db:        db,
		status:    TransactionStatusActive,
		startTime: time.Now(),
		logger:    logger,
	}
}

// GetDB 获取事务数据库连接
func (t *gormTransaction) GetDB() *gorm.DB {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if t.tx == nil {
		t.tx = t.db.Begin()
	}
	return t.tx
}

// Commit 提交事务
func (t *gormTransaction) Commit() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.status != TransactionStatusActive {
		return fmt.Errorf("事务状态无效: %s", t.status)
	}

	if t.tx == nil {
		return fmt.Errorf("事务未初始化")
	}

	err := t.tx.Commit().Error
	if err != nil {
		t.status = TransactionStatusRolledBack
		t.logger.Error("事务提交失败",
			"transaction_id", t.id,
			"error", err)
		return fmt.Errorf("事务提交失败: %w", err)
	}

	t.status = TransactionStatusCommitted
	duration := time.Since(t.startTime)

	t.logger.Info("事务提交成功",
		"transaction_id", t.id,
		"duration", duration)

	return nil
}

// Rollback 回滚事务
func (t *gormTransaction) Rollback() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.status != TransactionStatusActive {
		return nil // 已经不是活跃状态，无需回滚
	}

	if t.tx == nil {
		t.status = TransactionStatusRolledBack
		return nil
	}

	err := t.tx.Rollback().Error
	t.status = TransactionStatusRolledBack
	duration := time.Since(t.startTime)

	if err != nil {
		t.logger.Error("事务回滚失败",
			"transaction_id", t.id,
			"duration", duration,
			"error", err)
		return fmt.Errorf("事务回滚失败: %w", err)
	}

	t.logger.Info("事务回滚成功",
		"transaction_id", t.id,
		"duration", duration)

	return nil
}

// IsActive 检查事务是否活跃
func (t *gormTransaction) IsActive() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.status == TransactionStatusActive
}

// GetID 获取事务ID
func (t *gormTransaction) GetID() string {
	return t.id
}

// defaultTransactionManager 默认事务管理器实现
type defaultTransactionManager struct {
	db                 *gorm.DB
	logger             *log.Logger
	activeTransactions map[string]*gormTransaction
	mutex              sync.RWMutex
	maxTransactionAge  time.Duration
	cleanupInterval    time.Duration
	stopCleanup        chan bool
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(db *gorm.DB, logger *log.Logger) TransactionManager {
	tm := &defaultTransactionManager{
		db:                 db,
		logger:             logger,
		activeTransactions: make(map[string]*gormTransaction),
		maxTransactionAge:  30 * time.Minute, // 最大事务存活时间
		cleanupInterval:    5 * time.Minute,  // 清理间隔
		stopCleanup:        make(chan bool),
	}

	// 启动清理协程
	go tm.startCleanupRoutine()

	return tm
}

// BeginTransaction 开始事务
func (tm *defaultTransactionManager) BeginTransaction(ctx context.Context) (Transaction, error) {
	transactionID := tm.generateTransactionID()

	tx := NewGormTransaction(transactionID, tm.db, tm.logger)

	// 初始化事务
	_ = tx.GetDB()

	tm.mutex.Lock()
	tm.activeTransactions[transactionID] = tx
	tm.mutex.Unlock()

	tm.logger.Debug("开始事务",
		"transaction_id", transactionID)

	return tx, nil
}

// ExecuteInTransaction 在事务中执行操作
func (tm *defaultTransactionManager) ExecuteInTransaction(ctx context.Context, fn func(tx Transaction) error) error {
	tx, err := tm.BeginTransaction(ctx)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}

	defer func() {
		// 从活跃事务列表中移除
		tm.mutex.Lock()
		delete(tm.activeTransactions, tx.GetID())
		tm.mutex.Unlock()
	}()

	// 执行业务逻辑
	err = fn(tx)
	if err != nil {
		// 回滚事务
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			tm.logger.Error("事务回滚失败",
				"transaction_id", tx.GetID(),
				"error", rollbackErr)
		}
		return fmt.Errorf("事务执行失败: %w", err)
	}

	// 提交事务
	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("事务提交失败: %w", commitErr)
	}

	return nil
}

// GetActiveTransactions 获取活跃事务数量
func (tm *defaultTransactionManager) GetActiveTransactions() int {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return len(tm.activeTransactions)
}

// CleanupExpiredTransactions 清理过期事务
func (tm *defaultTransactionManager) CleanupExpiredTransactions() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	now := time.Now()
	expiredTransactions := make([]*gormTransaction, 0)

	for id, tx := range tm.activeTransactions {
		if now.Sub(tx.startTime) > tm.maxTransactionAge {
			expiredTransactions = append(expiredTransactions, tx)
			delete(tm.activeTransactions, id)
		}
	}

	// 回滚过期事务
	for _, tx := range expiredTransactions {
		if tx.IsActive() {
			tx.status = TransactionStatusExpired
			if err := tx.Rollback(); err != nil {
				tm.logger.Error("清理过期事务失败",
					"transaction_id", tx.GetID(),
					"error", err)
			} else {
				tm.logger.Warn("清理过期事务",
					"transaction_id", tx.GetID(),
					"age", now.Sub(tx.startTime))
			}
		}
	}

	if len(expiredTransactions) > 0 {
		tm.logger.Info("清理过期事务完成",
			"cleaned_count", len(expiredTransactions),
			"active_count", len(tm.activeTransactions))
	}

	return nil
}

// generateTransactionID 生成事务ID
func (tm *defaultTransactionManager) generateTransactionID() string {
	return fmt.Sprintf("tx_%d_%d", time.Now().UnixNano(), len(tm.activeTransactions))
}

// startCleanupRoutine 启动清理协程
func (tm *defaultTransactionManager) startCleanupRoutine() {
	ticker := time.NewTicker(tm.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := tm.CleanupExpiredTransactions(); err != nil {
				tm.logger.Error("清理过期事务失败", "error", err)
			}
		case <-tm.stopCleanup:
			return
		}
	}
}

// Stop 停止事务管理器
func (tm *defaultTransactionManager) Stop() {
	close(tm.stopCleanup)

	// 回滚所有活跃事务
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	for _, tx := range tm.activeTransactions {
		if tx.IsActive() {
			if err := tx.Rollback(); err != nil {
				tm.logger.Error("停止时回滚事务失败",
					"transaction_id", tx.GetID(),
					"error", err)
			}
		}
	}

	tm.logger.Info("事务管理器已停止")
}

// TransactionContext 事务上下文
type TransactionContext struct {
	context.Context
	Transaction Transaction
}

// NewTransactionContext 创建事务上下文
func NewTransactionContext(ctx context.Context, tx Transaction) *TransactionContext {
	return &TransactionContext{
		Context:     ctx,
		Transaction: tx,
	}
}

// GetTransaction 从上下文获取事务
func GetTransaction(ctx context.Context) (Transaction, bool) {
	if txCtx, ok := ctx.(*TransactionContext); ok {
		return txCtx.Transaction, true
	}
	return nil, false
}

// WithTransaction 在上下文中添加事务
func WithTransaction(ctx context.Context, tx Transaction) context.Context {
	return NewTransactionContext(ctx, tx)
}
