package transaction_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"piemdm/internal/transaction"
	"piemdm/pkg/log"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func TestTransactionManager_BeginTransaction(t *testing.T) {
	db, mock := setupTestDB(t)
	defer mock.ExpectationsWereMet()

	logger := &log.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	tm := transaction.NewTransactionManager(db, logger)

	// Mock开始事务
	mock.ExpectBegin()

	ctx := context.Background()
	tx, err := tm.BeginTransaction(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.True(t, tx.IsActive())
	assert.NotEmpty(t, tx.GetID())
}

func TestTransactionManager_ExecuteInTransaction_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	defer mock.ExpectationsWereMet()

	logger := &log.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	tm := transaction.NewTransactionManager(db, logger)

	// Mock事务操作
	mock.ExpectBegin()
	mock.ExpectCommit()

	ctx := context.Background()
	executed := false

	err := tm.ExecuteInTransaction(ctx, func(tx transaction.Transaction) error {
		executed = true
		assert.NotNil(t, tx)
		assert.True(t, tx.IsActive())
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestTransactionManager_ExecuteInTransaction_Rollback(t *testing.T) {
	db, mock := setupTestDB(t)
	defer mock.ExpectationsWereMet()

	logger := &log.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	tm := transaction.NewTransactionManager(db, logger)

	// Mock事务操作
	mock.ExpectBegin()
	mock.ExpectRollback()

	ctx := context.Background()
	testError := errors.New("test error")

	err := tm.ExecuteInTransaction(ctx, func(tx transaction.Transaction) error {
		return testError
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "事务执行失败")
}

func TestTransactionManager_GetActiveTransactions(t *testing.T) {
	db, mock := setupTestDB(t)
	defer mock.ExpectationsWereMet()

	logger := &log.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	tm := transaction.NewTransactionManager(db, logger)

	// 初始状态应该没有活跃事务
	assert.Equal(t, 0, tm.GetActiveTransactions())

	// Mock开始事务
	mock.ExpectBegin()

	ctx := context.Background()
	tx, err := tm.BeginTransaction(ctx)
	require.NoError(t, err)

	// 应该有一个活跃事务
	assert.Equal(t, 1, tm.GetActiveTransactions())

	// Mock提交事务
	mock.ExpectCommit()
	err = tx.Commit()
	require.NoError(t, err)
}

func TestGormTransaction_Commit(t *testing.T) {
	db, mock := setupTestDB(t)
	defer mock.ExpectationsWereMet()

	logger := &log.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}

	// Mock开始和提交事务
	mock.ExpectBegin()
	mock.ExpectCommit()

	tx := transaction.NewGormTransaction("test-tx", db, logger)

	// 初始化事务
	_ = tx.GetDB()

	assert.True(t, tx.IsActive())

	err := tx.Commit()
	assert.NoError(t, err)
	assert.False(t, tx.IsActive())
}

func TestGormTransaction_Rollback(t *testing.T) {
	db, mock := setupTestDB(t)
	defer mock.ExpectationsWereMet()

	logger := &log.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}

	// Mock开始和回滚事务
	mock.ExpectBegin()
	mock.ExpectRollback()

	tx := transaction.NewGormTransaction("test-tx", db, logger)

	// 初始化事务
	_ = tx.GetDB()

	assert.True(t, tx.IsActive())

	err := tx.Rollback()
	assert.NoError(t, err)
	assert.False(t, tx.IsActive())
}

func TestTransactionContext(t *testing.T) {
	db, mock := setupTestDB(t)
	defer mock.ExpectationsWereMet()

	logger := &log.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}

	// Mock开始事务
	mock.ExpectBegin()

	tx := transaction.NewGormTransaction("test-tx", db, logger)
	_ = tx.GetDB() // 初始化事务

	ctx := context.Background()

	// 测试WithTransaction
	txCtx := transaction.WithTransaction(ctx, tx)
	assert.NotNil(t, txCtx)

	// 测试GetTransaction
	retrievedTx, ok := transaction.GetTransaction(txCtx)
	assert.True(t, ok)
	assert.Equal(t, tx, retrievedTx)

	// 测试从普通上下文获取事务
	_, ok = transaction.GetTransaction(ctx)
	assert.False(t, ok)
}

func TestTransactionManager_CleanupExpiredTransactions(t *testing.T) {
	db, mock := setupTestDB(t)
	defer mock.ExpectationsWereMet()

	logger := &log.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	tm := transaction.NewTransactionManager(db, logger)

	// 这个测试主要验证方法不会panic
	err := tm.CleanupExpiredTransactions()
	assert.NoError(t, err)
}
