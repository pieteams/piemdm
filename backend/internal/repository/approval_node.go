//go:generate mockgen -source=approval_node_repository.go -destination=../../test/mocks/repository/approval_node.go

package repository

import (
	"piemdm/internal/model"
)

type ApprovalNodeRepository interface {
	// 基础查询
	FindOne(id uint) (*model.ApprovalNode, error)
	FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.ApprovalNode, error)

	// Base CRUD
	Create(approvalNode *model.ApprovalNode) error
	Update(approvalNode *model.ApprovalNode) error
	Delete(id uint) (*model.ApprovalNode, error)

	// Batch operations
	BatchUpdate(ids []uint, approvalNode *model.ApprovalNode) error
	BatchDelete(ids []uint) error

	// 业务查询方法
	FirstByCode(code string) (*model.ApprovalNode, error)
	First(where map[string]any) (*model.ApprovalNode, error)
	FindByApprovalDefCode(approvalDefCode string) ([]*model.ApprovalNode, error)
	FindByNodeType(nodeType string) ([]*model.ApprovalNode, error)
	FindActiveByApprovalDefCode(approvalDefCode string) ([]*model.ApprovalNode, error)
	FindStartNode(approvalDefCode string) (*model.ApprovalNode, error)
	FindEndNodes(approvalDefCode string) ([]*model.ApprovalNode, error)
	FindNextNodes(approvalDefCode, currentNodeCode string) ([]*model.ApprovalNode, error)

	// 状态管理
	UpdateStatus(id uint, status string) error
	UpdateStatusByIds(ids []uint, status string) error

	// 统计查询
	CountByApprovalDefCode(approvalDefCode string) (int64, error)
	CountByNodeType(nodeType string) (int64, error)
}

type approvalNodeRepository struct {
	*Repository
	source Base
}

func NewApprovalNodeRepository(repository *Repository, source Base) ApprovalNodeRepository {
	return &approvalNodeRepository{
		Repository: repository,
		source:     source,
	}
}

func (r *approvalNodeRepository) FindOne(id uint) (*model.ApprovalNode, error) {
	var approvalNode model.ApprovalNode
	if err := r.source.FirstById(&approvalNode, id); err != nil {
		return nil, err
	}
	return &approvalNode, nil
}

func (r *approvalNodeRepository) FirstByCode(code string) (*model.ApprovalNode, error) {
	var approvalNode model.ApprovalNode
	if err := r.db.Where("node_code = ? AND deleted_at IS NULL", code).First(&approvalNode).Error; err != nil {
		return nil, err
	}
	return &approvalNode, nil
}

func (r *approvalNodeRepository) First(where map[string]any) (*model.ApprovalNode, error) {
	var approvalNode model.ApprovalNode
	if err := r.db.Where(where).First(&approvalNode).Error; err != nil {
		return nil, err
	}
	return &approvalNode, nil
}

func (r *approvalNodeRepository) FindPage(page, pageSize int, total *int64, where map[string]any) ([]*model.ApprovalNode, error) {
	var approvalNodes []*model.ApprovalNode
	var approvalNode model.ApprovalNode

	preloads := []string{}
	err := r.source.FindPage(approvalNode, &approvalNodes, page, pageSize, total, where, preloads, "sort_order ASC, created_at DESC")
	if err != nil {
		r.logger.Error("获取审批节点分页数据失败", "err", err)
	}
	return approvalNodes, nil
}

func (r *approvalNodeRepository) Create(approvalNode *model.ApprovalNode) error {
	if err := r.source.Create(approvalNode); err != nil {
		return err
	}
	return nil
}

func (r *approvalNodeRepository) Update(approvalNode *model.ApprovalNode) error {
	if err := r.source.Updates(&approvalNode, approvalNode); err != nil {
		return err
	}
	return nil
}

func (r *approvalNodeRepository) BatchUpdate(ids []uint, approvalNode *model.ApprovalNode) error {
	if err := r.db.Model(&approvalNode).Where("id in ?", ids).Updates(approvalNode).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalNodeRepository) Delete(id uint) (*model.ApprovalNode, error) {
	var approvalNode model.ApprovalNode
	if err := r.db.Where("id = ?", id).First(&approvalNode).Error; err != nil {
		return nil, err
	}

	// 使用软删除，不物理删除数据
	if err := r.db.Delete(&approvalNode).Error; err != nil {
		return nil, err
	}

	return &approvalNode, nil
}

func (r *approvalNodeRepository) BatchDelete(ids []uint) error {
	// 使用软删除，不物理删除数据
	if err := r.db.Where("id in ?", ids).Delete(&model.ApprovalNode{}).Error; err != nil {
		return err
	}
	return nil
}

// 业务查询方法
func (r *approvalNodeRepository) FindByApprovalDefCode(approvalDefCode string) ([]*model.ApprovalNode, error) {
	var approvalNodes []*model.ApprovalNode
	if err := r.db.Where("approval_def_code = ? AND deleted_at IS NULL", approvalDefCode).
		Order("sort_order ASC, created_at ASC").Find(&approvalNodes).Error; err != nil {
		return nil, err
	}
	return approvalNodes, nil
}

func (r *approvalNodeRepository) FindByNodeType(nodeType string) ([]*model.ApprovalNode, error) {
	var approvalNodes []*model.ApprovalNode
	if err := r.db.Where("node_type = ? AND deleted_at IS NULL", nodeType).
		Order("sort_order ASC, created_at ASC").Find(&approvalNodes).Error; err != nil {
		return nil, err
	}
	return approvalNodes, nil
}

func (r *approvalNodeRepository) FindActiveByApprovalDefCode(approvalDefCode string) ([]*model.ApprovalNode, error) {
	var approvalNodes []*model.ApprovalNode
	if err := r.db.Where("approval_def_code = ? AND status = ? AND deleted_at IS NULL",
		approvalDefCode, model.ApprovalDefStatusNormal).
		Order("sort_order ASC, created_at ASC").Find(&approvalNodes).Error; err != nil {
		return nil, err
	}
	return approvalNodes, nil
}

func (r *approvalNodeRepository) FindStartNode(approvalDefCode string) (*model.ApprovalNode, error) {
	var approvalNode model.ApprovalNode
	if err := r.db.Where("approval_def_code = ? AND node_type = ? AND status = ? AND deleted_at IS NULL",
		approvalDefCode, model.NodeTypeStart, model.ApprovalDefStatusNormal).
		First(&approvalNode).Error; err != nil {
		return nil, err
	}
	return &approvalNode, nil
}

func (r *approvalNodeRepository) FindEndNodes(approvalDefCode string) ([]*model.ApprovalNode, error) {
	var approvalNodes []*model.ApprovalNode
	if err := r.db.Where("approval_def_code = ? AND node_type = ? AND status = ? AND deleted_at IS NULL",
		approvalDefCode, model.NodeTypeEnd, model.ApprovalDefStatusNormal).
		Order("sort_order ASC").Find(&approvalNodes).Error; err != nil {
		return nil, err
	}
	return approvalNodes, nil
}

func (r *approvalNodeRepository) FindNextNodes(approvalDefCode, currentNodeCode string) ([]*model.ApprovalNode, error) {
	// 首先获取当前节点
	var currentNode model.ApprovalNode
	if err := r.db.Where("approval_def_code = ? AND node_code = ? AND deleted_at IS NULL",
		approvalDefCode, currentNodeCode).First(&currentNode).Error; err != nil {
		return nil, err
	}

	var nextNodes []*model.ApprovalNode
	// 这里需要根据当前节点的next_nodes字段来查找下一个节点
	// 简化实现：按sort_order查找下一个节点
	if err := r.db.Where("approval_def_code = ? AND sort_order > ? AND status = ? AND deleted_at IS NULL",
		approvalDefCode, currentNode.SortOrder, model.ApprovalDefStatusNormal).
		Order("sort_order ASC").Find(&nextNodes).Error; err != nil {
		return nil, err
	}
	return nextNodes, nil
}

// 状态管理
func (r *approvalNodeRepository) UpdateStatus(id uint, status string) error {
	if err := r.db.Model(&model.ApprovalNode{}).Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

func (r *approvalNodeRepository) UpdateStatusByIds(ids []uint, status string) error {
	if err := r.db.Model(&model.ApprovalNode{}).Where("id in ?", ids).
		Update("status", status).Error; err != nil {
		return err
	}
	return nil
}

// 统计查询
func (r *approvalNodeRepository) CountByApprovalDefCode(approvalDefCode string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.ApprovalNode{}).
		Where("approval_def_code = ? AND deleted_at IS NULL", approvalDefCode).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *approvalNodeRepository) CountByNodeType(nodeType string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.ApprovalNode{}).
		Where("node_type = ? AND deleted_at IS NULL", nodeType).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
