//go:generate mockgen -source=approval_task_repository.go -destination=../../test/mocks/repository/approval_task.go -package=mock_repository

package repository

import (
	"time"

	"piemdm/internal/model"
)

type ApprovalTaskRepository interface {
	// 基础查询
	FindOne(id uint) (*model.ApprovalTask, error)
	FindPage(page, pageSize int, total *int64, where map[string]any, order string) ([]*model.ApprovalTask, error)

	// Base CRUD
	Create(approvalTask *model.ApprovalTask) error
	Update(approvalTask *model.ApprovalTask) error
	Delete(id uint) (*model.ApprovalTask, error)

	// Batch operations
	BatchUpdate(ids []uint, approvalTask *model.ApprovalTask) error
	BatchDelete(ids []uint) error

	// 业务查询方法
	FirstByCode(code string) (*model.ApprovalTask, error)
	First(where map[string]any) (*model.ApprovalTask, error)
	Find(where map[string]any) ([]*model.ApprovalTask, error)
	FindByApprovalCode(approvalCode string) ([]*model.ApprovalTask, error)
	FindByAssigneeID(assigneeID string) ([]*model.ApprovalTask, error)
	FindPendingByAssignee(assigneeID string) ([]*model.ApprovalTask, error)
	FindByNodeCode(nodeCode string) ([]*model.ApprovalTask, error)
	FindByStatus(status string) ([]*model.ApprovalTask, error)
	FindOverdueTasks() ([]*model.ApprovalTask, error)
	FindExpiredTasks() ([]*model.ApprovalTask, error)

	// 状态管理
	UpdateStatus(id uint, status string) error
	UpdateResult(id uint, result, comment, reason string) error
	CompleteTask(id uint, result, comment, reason string) error
	TransferTask(id uint, fromUserID, toUserID, fromUserName, toUserName, reason string) error

	// 时间管理
	UpdateAssignedAt(id uint, assignedAt time.Time) error
	UpdateStartedAt(id uint, startedAt time.Time) error
	UpdateCompletedAt(id uint, completedAt time.Time) error
	UpdateExpiredAt(id uint, expiredAt time.Time) error
	UpdateLastRemindAt(id uint, lastRemindAt time.Time) error

	// 统计查询
	CountByStatus(status string) (int64, error)
	CountByAssignee(assigneeID string) (int64, error)
	CountPendingByAssignee(assigneeID string) (int64, error)
	CountByNodeCode(nodeCode string) (int64, error)
	CountByApprovalCode(approvalCode string) (int64, error)

	// 催办管理
	UpdateRemindCount(id uint, count int) error
	UpdateRemindInfo(id uint) error
	BatchUpdateRemindInfo(ids []uint) error
	FindTasksNeedRemind() ([]*model.ApprovalTask, error)
}

type approvalTaskRepository struct {
	*Repository
	source Base
}

func NewApprovalTaskRepository(repository *Repository, source Base) ApprovalTaskRepository {
	return &approvalTaskRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *approvalTaskRepository) FindOne(id uint) (*model.ApprovalTask, error) {
	var approvalTask model.ApprovalTask
	if err := r.source.FirstById(&approvalTask, id); err != nil {
		return nil, err
	}
	return &approvalTask, nil
}

func (r *approvalTaskRepository) FirstByCode(code string) (*model.ApprovalTask, error) {
	var approvalTask model.ApprovalTask
	if err := r.db.Where("task_code = ? AND deleted_at IS NULL", code).First(&approvalTask).Error; err != nil {
		return nil, err
	}
	return &approvalTask, nil
}

func (r *approvalTaskRepository) First(where map[string]any) (*model.ApprovalTask, error) {
	var approvalTask model.ApprovalTask
	if err := r.db.Where(where).First(&approvalTask).Error; err != nil {
		return nil, err
	}
	return &approvalTask, nil
}

func (r *approvalTaskRepository) Find(where map[string]any) ([]*model.ApprovalTask, error) {
	var approvalTasks []*model.ApprovalTask
	if err := r.db.Where(where).Find(&approvalTasks).Error; err != nil {
		return nil, err
	}
	return approvalTasks, nil
}

func (r *approvalTaskRepository) FindPage(page, pageSize int, total *int64, where map[string]any, order string) ([]*model.ApprovalTask, error) {
	var approvalTasks []*model.ApprovalTask
	var approvalTask model.ApprovalTask

	preloads := []string{}
	if order == "" {
		order = "created_at DESC"
	}
	err := r.source.FindPage(approvalTask, &approvalTasks, page, pageSize, total, where, preloads, order)
	if err != nil {
		r.logger.Error("获取审批任务分页数据失败", "err", err)
	}
	return approvalTasks, nil
}

func (r *approvalTaskRepository) Create(approvalTask *model.ApprovalTask) error {
	if err := r.source.Create(approvalTask); err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) Update(approvalTask *model.ApprovalTask) error {
	if err := r.source.Updates(&approvalTask, approvalTask); err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) BatchUpdate(ids []uint, approvalTask *model.ApprovalTask) error {
	if err := r.db.Model(&approvalTask).Where("id in ?", ids).Updates(approvalTask).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) Delete(id uint) (*model.ApprovalTask, error) {
	var approvalTask model.ApprovalTask
	if err := r.db.Where("id = ?", id).First(&approvalTask).Error; err != nil {
		return nil, err
	}

	// 使用软删除，不物理删除数据
	if err := r.db.Delete(&approvalTask).Error; err != nil {
		return nil, err
	}

	return &approvalTask, nil
}

func (r *approvalTaskRepository) BatchDelete(ids []uint) error {
	// 使用软删除，不物理删除数据
	if err := r.db.Where("id in ?", ids).Delete(&model.ApprovalTask{}).Error; err != nil {
		return err
	}
	return nil
}

// 业务查询方法
func (r *approvalTaskRepository) FindByApprovalCode(approvalCode string) ([]*model.ApprovalTask, error) {
	var approvalTasks []*model.ApprovalTask
	if err := r.db.Where("approval_code = ? AND deleted_at IS NULL", approvalCode).
		Order("created_at ASC").Find(&approvalTasks).Error; err != nil {
		return nil, err
	}
	return approvalTasks, nil
}

func (r *approvalTaskRepository) FindByAssigneeID(assigneeID string) ([]*model.ApprovalTask, error) {
	var approvalTasks []*model.ApprovalTask
	if err := r.db.Where("assignee_id = ? AND deleted_at IS NULL", assigneeID).
		Order("created_at DESC").Find(&approvalTasks).Error; err != nil {
		return nil, err
	}
	return approvalTasks, nil
}

func (r *approvalTaskRepository) FindPendingByAssignee(assigneeID string) ([]*model.ApprovalTask, error) {
	var approvalTasks []*model.ApprovalTask
	if err := r.db.Where("assignee_id = ? AND status = ? AND deleted_at IS NULL",
		assigneeID, model.TaskStatusPending).
		Order("created_at DESC").Find(&approvalTasks).Error; err != nil {
		return nil, err
	}
	return approvalTasks, nil
}

func (r *approvalTaskRepository) FindByNodeCode(nodeCode string) ([]*model.ApprovalTask, error) {
	var approvalTasks []*model.ApprovalTask
	if err := r.db.Where("node_code = ? AND deleted_at IS NULL", nodeCode).
		Order("created_at DESC").Find(&approvalTasks).Error; err != nil {
		return nil, err
	}
	return approvalTasks, nil
}

func (r *approvalTaskRepository) FindByStatus(status string) ([]*model.ApprovalTask, error) {
	var approvalTasks []*model.ApprovalTask
	if err := r.db.Where("status = ? AND deleted_at IS NULL", status).
		Order("created_at DESC").Find(&approvalTasks).Error; err != nil {
		return nil, err
	}
	return approvalTasks, nil
}

func (r *approvalTaskRepository) FindOverdueTasks() ([]*model.ApprovalTask, error) {
	var approvalTasks []*model.ApprovalTask
	if err := r.db.Where("expired_at IS NOT NULL AND expired_at < ? AND status = ? AND deleted_at IS NULL",
		time.Now(), model.TaskStatusPending).Find(&approvalTasks).Error; err != nil {
		return nil, err
	}
	return approvalTasks, nil
}

func (r *approvalTaskRepository) FindExpiredTasks() ([]*model.ApprovalTask, error) {
	var approvalTasks []*model.ApprovalTask
	if err := r.db.Where("expired_at IS NOT NULL AND expired_at < ? AND status IN ? AND deleted_at IS NULL",
		time.Now(), []string{model.TaskStatusPending}).Find(&approvalTasks).Error; err != nil {
		return nil, err
	}
	return approvalTasks, nil
}

// 状态管理
func (r *approvalTaskRepository) UpdateStatus(id uint, status string) error {
	if err := r.db.Model(&model.ApprovalTask{}).Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) UpdateResult(id uint, result, comment, reason string) error {
	if err := r.db.Model(&model.ApprovalTask{}).Where("id = ?", id).
		Updates(map[string]any{
			"result":  result,
			"comment": comment,
			"reason":  reason,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) CompleteTask(id uint, result, comment, reason string) error {
	now := time.Now()
	if err := r.db.Model(&model.ApprovalTask{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":       model.TaskStatusDone,
			"result":       result,
			"comment":      comment,
			"reason":       reason,
			"completed_at": now,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) TransferTask(id uint, fromUserID, toUserID, fromUserName, toUserName, reason string) error {
	now := time.Now()
	if err := r.db.Model(&model.ApprovalTask{}).Where("id = ?", id).
		Updates(map[string]any{
			"assignee_id":        toUserID,
			"assignee_name":      toUserName,
			"transfer_from_id":   fromUserID,
			"transfer_from_name": fromUserName,
			"transfer_reason":    reason,
			"transfer_at":        now,
			"status":             model.TaskStatusTransferred,
		}).Error; err != nil {
		return err
	}
	return nil
}

// 时间管理
func (r *approvalTaskRepository) UpdateAssignedAt(id uint, assignedAt time.Time) error {
	if err := r.db.Model(&model.ApprovalTask{}).Where("id = ?", id).
		Update("assigned_at", assignedAt).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) UpdateStartedAt(id uint, startedAt time.Time) error {
	if err := r.db.Model(&model.ApprovalTask{}).Where("id = ?", id).
		Update("started_at", startedAt).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) UpdateCompletedAt(id uint, completedAt time.Time) error {
	if err := r.db.Model(&model.ApprovalTask{}).Where("id = ?", id).
		Update("completed_at", completedAt).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) UpdateExpiredAt(id uint, expiredAt time.Time) error {
	if err := r.db.Model(&model.ApprovalTask{}).Where("id = ?", id).
		Update("expired_at", expiredAt).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) UpdateLastRemindAt(id uint, lastRemindAt time.Time) error {
	if err := r.db.Model(&model.ApprovalTask{}).Where("id = ?", id).
		Update("last_remind_at", lastRemindAt).Error; err != nil {
		return err
	}
	return nil
}

// 统计查询
func (r *approvalTaskRepository) CountByStatus(status string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.ApprovalTask{}).
		Where("status = ? AND deleted_at IS NULL", status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *approvalTaskRepository) CountByAssignee(assigneeID string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.ApprovalTask{}).
		Where("assignee_id = ? AND deleted_at IS NULL", assigneeID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *approvalTaskRepository) CountPendingByAssignee(assigneeID string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.ApprovalTask{}).
		Where("assignee_id = ? AND status = ? AND deleted_at IS NULL",
			assigneeID, model.TaskStatusPending).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *approvalTaskRepository) CountByApprovalCode(approvalCode string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.ApprovalTask{}).
		Where("approval_code = ? AND deleted_at IS NULL", approvalCode).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *approvalTaskRepository) CountByNodeCode(nodeCode string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.ApprovalTask{}).
		Where("node_code = ? AND deleted_at IS NULL", nodeCode).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// 催办管理
func (r *approvalTaskRepository) UpdateRemindCount(id uint, count int) error {
	now := time.Now()
	if err := r.db.Model(&model.ApprovalTask{}).Where("id = ?", id).
		Updates(map[string]any{
			"remind_count":   count,
			"last_remind_at": now,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) UpdateRemindInfo(id uint) error {
	now := time.Now()
	updates := map[string]any{
		"last_remind_at": now,
		"remind_count":   r.db.Raw("remind_count + 1"),
		"updated_at":     now,
	}

	if err := r.db.Model(&model.ApprovalTask{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) BatchUpdateRemindInfo(ids []uint) error {
	now := time.Now()
	updates := map[string]any{
		"last_remind_at": now,
		"remind_count":   r.db.Raw("remind_count + 1"),
		"updated_at":     now,
	}

	if err := r.db.Model(&model.ApprovalTask{}).Where("id IN ?", ids).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalTaskRepository) FindTasksNeedRemind() ([]*model.ApprovalTask, error) {
	var approvalTasks []*model.ApprovalTask
	// 查找需要催办的任务：状态为待处理，且距离上次催办时间超过催办间隔
	if err := r.db.Where(`status = ? AND deleted_at IS NULL AND
		(last_remind_at IS NULL OR last_remind_at < DATE_SUB(NOW(), INTERVAL 24 HOUR))`,
		model.TaskStatusPending).Find(&approvalTasks).Error; err != nil {
		return nil, err
	}
	return approvalTasks, nil
}
