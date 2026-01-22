/**
 * 工作流操作服务
 * 提供工作流的创建、更新、验证等核心业务逻辑
 */
import { WorkflowUtils } from '../utils/workflow-utils.js';
import { NodeService } from './node-service.js';
import { JsonHelper } from '../utils/json-helper.js';
import { NodeHelper } from '../utils/node-helper.js';

export const WorkflowService = {
  /**
   * 创建新工作流
   * @param {string} name - 工作流名称
   * @param {string} description - 工作流描述
   * @returns {Object} 新工作流对象
   */
  createWorkflow(name, description = '') {
    const startNode = NodeService.createNode('START', '开始');
    const endNode = NodeService.createNode('END', '结束');

    return {
      name,
      description,
      nodes: [startNode, endNode],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      status: 'draft',
    };
  },

  /**
   * 添加节点到工作流
   * @param {Object} workflow - 工作流对象
   * @param {Array} parentList - 父级节点列表
   * @param {number} index - 插入位置
   * @param {string} nodeType - 节点类型
   * @param {string} nodeName - 节点名称
   * @returns {Object} 更新后的工作流
   */
  addNode(workflow, parentList, index, nodeType, nodeName) {
    const newNode = NodeService.createNode(nodeType, nodeName);

    // 设置排序顺序
    newNode.SortOrder = index;

    // 更新后续节点的排序
    for (let i = index; i < parentList.length; i++) {
      parentList[i].SortOrder = i + 1;
    }

    // 插入新节点
    parentList.splice(index, 0, newNode);

    // 更新工作流时间戳
    workflow.updatedAt = new Date().toISOString();

    return workflow;
  },

  /**
   * 从工作流删除节点
   * @param {Object} workflow - 工作流对象
   * @param {Array} parentList - 父级节点列表
   * @param {number} index - 删除位置
   * @returns {Object} 删除结果 {success: boolean, workflow: Object, error: string}
   */
  deleteNode(workflow, parentList, index) {
    const nodeToDelete = parentList[index];

    if (!nodeToDelete) {
      return {
        success: false,
        workflow,
        error: '节点不存在',
      };
    }

    // 检查是否可以删除
    const canDelete = NodeService.canDeleteNode(nodeToDelete, workflow.nodes);
    if (!canDelete.canDelete) {
      return {
        success: false,
        workflow,
        error: canDelete.reason,
      };
    }

    // 删除节点
    parentList.splice(index, 1);

    // 重新设置排序
    parentList.forEach((node, i) => {
      node.SortOrder = i;
    });

    // 更新工作流时间戳
    workflow.updatedAt = new Date().toISOString();

    return {
      success: true,
      workflow,
      error: null,
    };
  },

  /**
   * 更新工作流节点
   * @param {Object} workflow - 工作流对象
   * @param {string} nodeCode - 节点编码
   * @param {Object} updates - 更新数据
   * @returns {Object} 更新结果 {success: boolean, workflow: Object, error: string}
   */
  updateNode(workflow, nodeCode, updates) {
    const nodePath = WorkflowUtils.findNodePath(workflow.nodes, nodeCode);

    if (nodePath.length === 0) {
      return {
        success: false,
        workflow,
        error: '节点不存在',
      };
    }

    try {
      const { array, index } = nodePath[nodePath.length - 1];
      const updatedNode = NodeService.updateNode(array[index], updates);
      array[index] = updatedNode;

      // 特殊处理：如果更新的节点在条件分支中，需要更新原始的ConditionConfig
      if (nodePath.length > 1) {
        const conditionNodePath = nodePath[0];
        const conditionNode = conditionNodePath.array[conditionNodePath.index];
        if (conditionNode.NodeType === 'CONDITION') {
          // 重新构建ConditionConfig以包含更新的节点数据
          const config = JsonHelper.safeParse(conditionNode.ConditionConfig);
          if (config.branches && Array.isArray(config.branches)) {
            // 找到包含更新节点的分支并更新其配置
            let foundInBranch = false;
            for (const branch of config.branches) {
              if (branch.nodes && Array.isArray(branch.nodes)) {
                const branchNodeIndex = branch.nodes.findIndex(node => node.NodeCode === nodeCode);
                if (branchNodeIndex !== -1) {
                  // 深度复制更新后的节点数据，确保所有字段都被正确更新
                  const branchNodeUpdate = { ...updatedNode };
                  branch.nodes[branchNodeIndex] = branchNodeUpdate;
                  foundInBranch = true;
                  break;
                }
              }
            }

            if (foundInBranch) {
              // 强制更新条件节点的ConditionConfig，确保响应式系统能够检测到变化
              conditionNode.ConditionConfig = JsonHelper.safeStringify(config);

              // 额外触发条件节点的更新，确保workflowData能够响应变化
              const conditionNodeUpdates = {
                ConditionConfig: conditionNode.ConditionConfig,
              };
              const updatedConditionNode = NodeService.updateNode(
                conditionNode,
                conditionNodeUpdates,
              );
              conditionNodePath.array[conditionNodePath.index] = updatedConditionNode;
            }
          }
        }
      }

      // 更新工作流时间戳
      workflow.updatedAt = new Date().toISOString();

      return {
        success: true,
        workflow,
        error: null,
      };
    } catch (error) {
      return {
        success: false,
        workflow,
        error: error.message,
      };
    }
  },

  /**
   * 验证工作流数据
   * @param {Object} workflow - 工作流对象
   * @returns {Object} 验证结果 {isValid: boolean, errors: Array}
   */
  validateWorkflow(workflow) {
    if (!workflow || typeof workflow !== 'object') {
      return {
        isValid: false,
        errors: ['工作流数据格式错误'],
      };
    }

    if (!workflow.nodes || !Array.isArray(workflow.nodes)) {
      return {
        isValid: false,
        errors: ['工作流必须包含节点数组'],
      };
    }

    return WorkflowUtils.validateWorkflow(workflow.nodes);
  },

  /**
   * 克隆工作流
   * @param {Object} workflow - 原始工作流
   * @param {string} newName - 新工作流名称
   * @returns {Object} 克隆的工作流
   */
  cloneWorkflow(workflow, newName) {
    const clonedWorkflow = JsonHelper.safeParse(JsonHelper.safeStringify(workflow));

    // 生成新的节点编码
    const regenerateNodeCodes = nodes => {
      nodes.forEach(node => {
        node.ID = NodeService.generateId();
        node.NodeCode = NodeService.generateNodeCode();

        // 处理条件分支中的节点
        if (node.NodeType === 'CONDITION') {
          const config = JsonHelper.safeParse(node.ConditionConfig);
          if (config.branches) {
            config.branches.forEach(branch => {
              if (branch.nodes) {
                regenerateNodeCodes(branch.nodes);
              }
            });
            node.ConditionConfig = JsonHelper.safeStringify(config);
          }
        }
      });
    };

    regenerateNodeCodes(clonedWorkflow.nodes);

    // 更新工作流信息
    clonedWorkflow.name = newName || `${workflow.name} (副本)`;
    clonedWorkflow.createdAt = new Date().toISOString();
    clonedWorkflow.updatedAt = new Date().toISOString();
    clonedWorkflow.status = 'draft';

    return clonedWorkflow;
  },

  /**
   * 获取工作流统计信息
   * @param {Object} workflow - 工作流对象
   * @returns {Object} 统计信息
   */
  getWorkflowStats(workflow) {
    if (!workflow || !workflow.nodes) {
      return {
        totalNodes: 0,
        nodeTypes: {},
        maxDepth: 0,
      };
    }

    const stats = {
      totalNodes: workflow.nodes.length,
      nodeTypes: {},
      maxDepth: 0,
    };

    // 统计节点类型
    workflow.nodes.forEach(node => {
      stats.nodeTypes[node.NodeType] = (stats.nodeTypes[node.NodeType] || 0) + 1;
    });

    // 计算最大深度（简化版本）
    const calculateDepth = (nodes, currentDepth = 1) => {
      let maxDepth = currentDepth;

      nodes.forEach(node => {
        if (node.NodeType === 'CONDITION') {
          const branches = WorkflowUtils.parseConditionConfig(node.ConditionConfig);
          branches.forEach(branch => {
            if (branch.nodes && branch.nodes.length > 0) {
              const branchDepth = calculateDepth(branch.nodes, currentDepth + 1);
              maxDepth = Math.max(maxDepth, branchDepth);
            }
          });
        }
      });

      return maxDepth;
    };

    stats.maxDepth = calculateDepth(workflow.nodes);

    return stats;
  },

  /**
   * 导出工作流为JSON字符串
   * @param {Object} workflow - 工作流对象
   * @param {boolean} formatted - 是否格式化输出
   * @returns {string} JSON字符串
   */
  exportWorkflow(workflow, formatted = true) {
    const exportData = {
      ...workflow,
      exportedAt: new Date().toISOString(),
      version: '1.0',
    };

    return formatted ? JSON.stringify(exportData, null, 2) : JSON.stringify(exportData);
  },

  /**
   * 从JSON字符串导入工作流
   * @param {string} jsonString - JSON字符串
   * @returns {Object} 导入结果 {success: boolean, workflow: Object, error: string}
   */
  importWorkflow(jsonString) {
    try {
      const workflow = JsonHelper.safeParse(jsonString);

      if (!workflow.nodes) {
        return {
          success: false,
          workflow: null,
          error: '无效的工作流数据格式',
        };
      }

      // 验证导入的数据
      const validation = this.validateWorkflow(workflow);
      if (!validation.isValid) {
        return {
          success: false,
          workflow: null,
          error: `工作流验证失败: ${validation.errors.join(', ')}`,
        };
      }

      return {
        success: true,
        workflow,
        error: null,
      };
    } catch (error) {
      return {
        success: false,
        workflow: null,
        error: `JSON解析失败: ${error.message}`,
      };
    }
  },
};
