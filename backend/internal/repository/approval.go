//go:generate mockgen -source=approval_repository.go -destination=../../test/mocks/repository/approval.go

package repository

import (
	"context"
	"encoding/json"
	"time"

	"piemdm/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ApprovalRepository interface {
	// 基础CRUD操作
	FindOne(id uint) (*model.Approval, error)
	FirstByCode(code string) (*model.Approval, error)
	FirstByExternalInstanceID(externalID string) (*model.Approval, error)
	First(where map[string]any) (*model.Approval, error)
	FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.Approval, error)
	Create(c *gin.Context, approval *model.Approval) error
	Update(c *gin.Context, approval *model.Approval) error
	BatchUpdate(c *gin.Context, ids []uint, approval *model.Approval) error
	Delete(c *gin.Context, id uint) (*model.Approval, error)
	BatchDelete(c *gin.Context, ids []uint) error
	SaveToQueue(webhookReq model.WebhookReq) error

	// 业务查询方法
	FindByApplicantID(applicantID string) ([]*model.Approval, error)
	FindByStatus(status string) ([]*model.Approval, error)
	FindByApprovalDefCode(approvalDefCode string) ([]*model.Approval, error)
	FindPendingByApplicant(applicantID string) ([]*model.Approval, error)
	FindByEntityCode(entityCode string) ([]*model.Approval, error)
	FindByEntityID(entityCode, entityID string) ([]*model.Approval, error)
	FindHistory(id uint) ([]*model.Approval, error)

	// 基于审批任务的查询方法
	FindPendingByAssignee(assigneeName string, page, pageSize int, total *int64, timeRange string) ([]*model.Approval, error)
	FindProcessedByAssignee(assigneeName string, page, pageSize int, total *int64, timeRange string) ([]*model.Approval, error)

	// 状态管理
	UpdateStatus(c *gin.Context, id uint, status string) error
	UpdateStatusByCode(c *gin.Context, code, status string) error
	UpdateCurrentNode(c *gin.Context, id uint, nodeID, nodeName string) error

	// 时间管理
	UpdateSubmittedAt(c *gin.Context, id uint, submittedAt time.Time) error
	UpdateStartedAt(c *gin.Context, id uint, startedAt time.Time) error
	UpdateCompletedAt(c *gin.Context, id uint, completedAt time.Time) error
	UpdateExpiredAt(c *gin.Context, id uint, expiredAt time.Time) error

	// 统计查询
	CountByStatus(status string) (int64, error)
	CountByApplicant(applicantID string) (int64, error)
	CountByApprovalDef(approvalDefCode string) (int64, error)
	FindExpiredApprovals() ([]*model.Approval, error)

	// 任务统计更新
	UpdateTaskCount(c *gin.Context, id uint, taskCount, completedTasks, pendingTasks int) error
}

type approvalRepository struct {
	*Repository
	source Base
}

func NewApprovalRepository(repository *Repository, source Base) ApprovalRepository {
	return &approvalRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *approvalRepository) First(where map[string]any) (*model.Approval, error) {
	var approval model.Approval
	if err := r.db.Where(where).First(&approval).Error; err != nil {
		return nil, err
	}
	return &approval, nil
}

func (r *approvalRepository) FindOne(id uint) (*model.Approval, error) {
	var approval model.Approval
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&approval).Error; err != nil {
		return nil, err
	}
	return &approval, nil
}

func (r *approvalRepository) FirstByCode(code string) (*model.Approval, error) {
	var approval model.Approval
	if err := r.db.Where("code = ? AND deleted_at IS NULL", code).First(&approval).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &approval, nil
}

func (r *approvalRepository) FirstByExternalInstanceID(externalID string) (*model.Approval, error) {
	var approval model.Approval
	if err := r.db.Where("external_instance_id = ? AND deleted_at IS NULL", externalID).First(&approval).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &approval, nil
}

func (r *approvalRepository) FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.Approval, error) {
	var approvals []*model.Approval
	var approval model.Approval

	// if err := r.db.Where("deleted_at is null").Offset((page - 1) * pageSize).Limit(pageSize).Find(&approvals).Error; err != nil {
	// 	return nil, err
	// }

	// if err := r.db.Model(approval).Where("deleted_at is null").Count(total).Error; err != nil {
	// 	r.logger.Error("apptoval repository count err", "err", err)
	// 	return nil, err
	// }
	preloads := []string{}
	err := r.source.FindPage(approval, &approvals, page, pageSize, total, where, preloads, "created_at DESC")
	if err != nil {
		r.logger.Error("获取审批实例分页数据失败", "err", err)
	}
	return approvals, nil
}

func (r *approvalRepository) Create(c *gin.Context, approval *model.Approval) error {
	if err := r.db.WithContext(c).Create(approval).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalRepository) Update(c *gin.Context, approval *model.Approval) error {
	if err := r.db.WithContext(c).Model(&approval).Updates(approval).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalRepository) BatchUpdate(c *gin.Context, ids []uint, approval *model.Approval) error {
	if err := r.db.WithContext(c).Model(&approval).Where("id in ?", ids).Updates(approval).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalRepository) Delete(c *gin.Context, id uint) (*model.Approval, error) {
	var approval model.Approval
	if err := r.db.WithContext(c).Where("id = ?", id).First(&approval).Error; err != nil {
		return nil, err
	}

	// 使用软删除，不物理删除数据
	if err := r.db.WithContext(c).Delete(&approval).Error; err != nil {
		return nil, err
	}

	return &approval, nil
}

func (r *approvalRepository) BatchDelete(c *gin.Context, ids []uint) error {
	// 使用软删除，不物理删除数据
	if err := r.db.WithContext(c).Where("id in ?", ids).Delete(&model.Approval{}).Error; err != nil {
		return err
	}
	return nil
}

// 业务查询方法
func (r *approvalRepository) FindByApplicantID(applicantID string) ([]*model.Approval, error) {
	var approvals []*model.Approval
	if err := r.db.Where("created_by = ? AND deleted_at IS NULL", applicantID).
		Order("created_at DESC").Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *approvalRepository) FindByStatus(status string) ([]*model.Approval, error) {
	var approvals []*model.Approval
	if err := r.db.Where("status = ? AND deleted_at IS NULL", status).
		Order("created_at DESC").Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *approvalRepository) FindByApprovalDefCode(approvalDefCode string) ([]*model.Approval, error) {
	var approvals []*model.Approval
	if err := r.db.Where("approval_def_code = ? AND deleted_at IS NULL", approvalDefCode).
		Order("created_at DESC").Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *approvalRepository) FindPendingByApplicant(applicantID string) ([]*model.Approval, error) {
	var approvals []*model.Approval
	if err := r.db.Where("created_by = ? AND status = ? AND deleted_at IS NULL",
		applicantID, model.ApprovalStatusPending).
		Order("created_at DESC").Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *approvalRepository) FindByEntityCode(entityCode string) ([]*model.Approval, error) {
	var approvals []*model.Approval
	if err := r.db.Where("entity_code = ? AND deleted_at IS NULL", entityCode).
		Order("created_at DESC").Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *approvalRepository) FindByEntityID(entityCode, entityID string) ([]*model.Approval, error) {
	var approvals []*model.Approval
	if err := r.db.Where("entity_code = ? AND entity_id = ? AND deleted_at IS NULL",
		entityCode, entityID).
		Order("created_at DESC").Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

// 状态管理
func (r *approvalRepository) UpdateStatus(c *gin.Context, id uint, status string) error {
	if err := r.db.WithContext(c).Model(&model.Approval{}).Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalRepository) UpdateStatusByCode(c *gin.Context, code, status string) error {
	if err := r.db.WithContext(c).Model(&model.Approval{}).Where("code = ?", code).
		Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalRepository) UpdateCurrentNode(c *gin.Context, id uint, nodeID, nodeName string) error {
	if err := r.db.WithContext(c).Model(&model.Approval{}).Where("id = ?", id).
		Updates(map[string]any{
			"current_task_id":   nodeID,
			"current_task_name": nodeName,
		}).Error; err != nil {
		return err
	}
	return nil
}

// 时间管理
func (r *approvalRepository) UpdateSubmittedAt(c *gin.Context, id uint, submittedAt time.Time) error {
	if err := r.db.WithContext(c).Model(&model.Approval{}).Where("id = ?", id).
		Update("submitted_at", submittedAt).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalRepository) UpdateStartedAt(c *gin.Context, id uint, startedAt time.Time) error {
	if err := r.db.WithContext(c).Model(&model.Approval{}).Where("id = ?", id).
		Update("started_at", startedAt).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalRepository) UpdateCompletedAt(c *gin.Context, id uint, completedAt time.Time) error {
	if err := r.db.WithContext(c).Model(&model.Approval{}).Where("id = ?", id).
		Update("completed_at", completedAt).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalRepository) UpdateExpiredAt(c *gin.Context, id uint, expiredAt time.Time) error {
	if err := r.db.WithContext(c).Model(&model.Approval{}).Where("id = ?", id).
		Update("expired_at", expiredAt).Error; err != nil {
		return err
	}
	return nil
}

// 统计查询
func (r *approvalRepository) CountByStatus(status string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Approval{}).
		Where("status = ? AND deleted_at IS NULL", status).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *approvalRepository) CountByApplicant(applicantID string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Approval{}).
		Where("created_by = ? AND deleted_at IS NULL", applicantID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *approvalRepository) CountByApprovalDef(approvalDefCode string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Approval{}).
		Where("approval_def_code = ? AND deleted_at IS NULL", approvalDefCode).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *approvalRepository) FindExpiredApprovals() ([]*model.Approval, error) {
	var approvals []*model.Approval
	if err := r.db.Where("expired_at IS NOT NULL AND expired_at < ? AND status = ? AND deleted_at IS NULL",
		time.Now(), model.ApprovalStatusPending).Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

// 任务统计更新
func (r *approvalRepository) UpdateTaskCount(c *gin.Context, id uint, taskCount, completedTasks, pendingTasks int) error {
	if err := r.db.WithContext(c).Model(&model.Approval{}).Where("id = ?", id).
		Updates(map[string]any{
			"task_count":      taskCount,
			"completed_tasks": completedTasks,
			"pending_tasks":   pendingTasks,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalRepository) FindHistory(id uint) ([]*model.Approval, error) {
	var approvals []*model.Approval
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).
		Order("created_at DESC").Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *approvalRepository) SaveToQueue(webhookReq model.WebhookReq) error {
	// 将 map[string]any 转换为 JSON 字符串
	dataJson, err := json.Marshal(webhookReq)
	if err != nil {
		return nil
	}

	// 存储消息
	// rds := redis.RedisPool.Get()
	// defer rds.Close()
	ctx := context.Background()
	res, err := r.Repository.rdb.LPush(ctx, "WebhookQueue", dataJson).Result()
	if err != nil {
		r.logger.Error("Workflow redis.LPush Error", "err", err)
	}
	r.logger.Info("Workflow redis.LPush res", "res", res)

	return nil
}

// FindPendingByAssignee 查询待指定用户审批的实例(分页)
// 通过关联approval_task表,查询assignee_name = 指定用户 且 status = 'Pending' 的任务对应的审批实例
func (r *approvalRepository) FindPendingByAssignee(assigneeName string, page, pageSize int, total *int64, timeRange string) ([]*model.Approval, error) {
	var approvals []*model.Approval

	// 构建基础查询
	query := r.db.Table("approvals a").
		Select("DISTINCT a.*").
		Joins("INNER JOIN approval_tasks t ON a.code = t.approval_code").
		Where("t.assignee_name = ?", assigneeName).
		Where("t.status = ?", model.TaskStatusPending).
		Where("a.status = ?", model.ApprovalStatusPending). // 添加审批实例状态过滤
		Where("a.deleted_at IS NULL")

	// 应用时间范围过滤
	query = r.applyTimeRangeFilter(query, timeRange)

	// 计算总数
	if total != nil {
		var count int64
		countQuery := r.db.Table("approvals a").
			Select("COUNT(DISTINCT a.id)").
			Joins("INNER JOIN approval_tasks t ON a.code = t.approval_code").
			Where("t.assignee_name = ?", assigneeName).
			Where("t.status = ?", model.TaskStatusPending).
			Where("a.status = ?", model.ApprovalStatusPending). // 添加审批实例状态过滤
			Where("a.deleted_at IS NULL")
		countQuery = r.applyTimeRangeFilter(countQuery, timeRange)
		if err := countQuery.Count(&count).Error; err != nil {
			r.logger.Error("统计待审批实例数量失败", "err", err)
			return nil, err
		}
		*total = count
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("a.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&approvals).Error

	if err != nil {
		r.logger.Error("查询待审批实例失败", "err", err)
		return nil, err
	}

	return approvals, nil
}

// FindProcessedByAssignee 查询指定用户已审批过的实例(分页)
// 通过关联approval_task表,查询assignee_name = 指定用户 且 status IN ('Approved', 'Rejected') 的任务对应的审批实例
func (r *approvalRepository) FindProcessedByAssignee(assigneeName string, page, pageSize int, total *int64, timeRange string) ([]*model.Approval, error) {
	var approvals []*model.Approval

	// 构建基础查询
	query := r.db.Table("approvals a").
		Select("DISTINCT a.*").
		Joins("INNER JOIN approval_tasks t ON a.code = t.approval_code").
		Where("t.assignee_name = ?", assigneeName).
		Where("t.status IN ?", []string{model.TaskStatusApproved, model.TaskStatusRejected}).
		Where("a.deleted_at IS NULL")

	// 应用时间范围过滤
	query = r.applyTimeRangeFilter(query, timeRange)

	// 计算总数
	if total != nil {
		var count int64
		countQuery := r.db.Table("approvals a").
			Select("COUNT(DISTINCT a.id)").
			Joins("INNER JOIN approval_tasks t ON a.code = t.approval_code").
			Where("t.assignee_name = ?", assigneeName).
			Where("t.status IN ?", []string{model.TaskStatusApproved, model.TaskStatusRejected}).
			Where("a.deleted_at IS NULL")
		countQuery = r.applyTimeRangeFilter(countQuery, timeRange)
		if err := countQuery.Count(&count).Error; err != nil {
			r.logger.Error("统计已审批实例数量失败", "err", err)
			return nil, err
		}
		*total = count
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("a.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&approvals).Error

	if err != nil {
		r.logger.Error("查询已审批实例失败", "err", err)
		return nil, err
	}

	return approvals, nil
}

// applyTimeRangeFilter 应用时间范围过滤
func (r *approvalRepository) applyTimeRangeFilter(query *gorm.DB, timeRange string) *gorm.DB {
	if timeRange == "" || timeRange == "All" {
		return query
	}

	now := time.Now()
	switch timeRange {
	case "Today":
		return query.Where("a.created_at >= ?", now.Format("2006-01-02 00:00:00"))
	case "LastWeek":
		return query.Where("a.created_at >= ?", now.AddDate(0, 0, -7).Format("2006-01-02 00:00:00"))
	case "LastMonth":
		return query.Where("a.created_at >= ?", now.AddDate(0, 0, -30).Format("2006-01-02 00:00:00"))
	}
	return query
}
