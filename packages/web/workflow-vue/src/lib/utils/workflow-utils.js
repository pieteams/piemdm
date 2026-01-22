/**
 * 工作流工具函数
 * 提供工作流相关的通用操作方法
 */
import { JsonHelper } from './json-helper.js';
import { NodeHelper } from './node-helper.js';

export const WorkflowUtils = {
  /**
   * 收集条件分支中的所有节点编码
   * @param {Array} nodes - 节点列表
   * @returns {Set} 条件分支节点编码集合
   */
  collectConditionBranchNodes(nodes) {
    const conditionBranchNodeCodes = new Set();

    nodes.forEach(node => {
      if (!NodeHelper.isConditionNode(node)) return;

      const config = JsonHelper.safeParse(node.ConditionConfig);
      if (config.branches && Array.isArray(config.branches)) {
        config.branches.forEach(branch => {
          if (branch.nodes && Array.isArray(branch.nodes)) {
            branch.nodes.forEach(branchNode => {
              const nodeCode = branchNode.nodeCode || branchNode.NodeCode;
              if (nodeCode) {
                conditionBranchNodeCodes.add(nodeCode);
              }
            });
          }
        });
      }
    });

    return conditionBranchNodeCodes;
  },

  /**
   * 处理工作流数据，过滤重复的条件分支节点
   * @param {Array} nodes - 原始节点列表
   * @returns {Array} 处理后的节点列表
   */
  processConditionConfig(nodes) {
    const processedNodes = [];
    const conditionBranchNodeCodes = this.collectConditionBranchNodes(nodes);

    nodes.forEach(node => {
      if (!conditionBranchNodeCodes.has(node.NodeCode)) {
        processedNodes.push(node);
      }
    });

    return processedNodes;
  },

  /**
   * 解析条件配置为标准化格式
   * @param {string} conditionConfig - 条件配置JSON字符串
   * @returns {Array} 标准化的分支数组
   */
  parseConditionConfig(conditionConfig) {
    if (!conditionConfig || conditionConfig === '{}') {
      return [];
    }

    const config = JsonHelper.safeParse(conditionConfig);
    if (!config.branches || !Array.isArray(config.branches)) {
      return [];
    }

    return config.branches.map((branch, branchIndex) => {
      // 处理条件配置，确保向后兼容
      let condition = branch.condition || {};

      // 如果使用旧格式，转换为新格式
      if (condition.fieldName && !condition.conditionType) {
        // 判断是实例字段还是数据字段
        if (condition.fieldName === 'createdBy') {
          condition = {
            conditionType: 'instance',
            fieldName: 'createdBy',
            operator: condition.operator || 'eq',
            fieldValue: condition.fieldValue || '',
          };
        } else {
          condition = {
            conditionType: 'data',
            fieldName: condition.fieldName,
            operator: condition.operator || 'eq',
            fieldValue: condition.fieldValue || '',
          };
        }
      }

      const formattedNodes = (branch.nodes || []).map(node => {
        return {
          // 主字段：大写字段（与主流程节点格式保持一致）
          ID: node.id || node.ID || 0,
          NodeCode: node.nodeCode || node.NodeCode,
          NodeName: node.nodeName || node.NodeName,
          NodeType: node.nodeType || node.NodeType,
          Description: node.description || node.Description || '',
          SortOrder: node.sortOrder || node.SortOrder || 0,
          ApproverType: node.approverType || node.ApproverType || 'USERS',
          ApproverConfig: node.approverConfig || node.ApproverConfig || '{}',
          ConditionConfig: node.conditionConfig || node.ConditionConfig || '{}',
          ApprovalDefCode: node.approvalDefCode || node.ApprovalDefCode || '',
          // 保留原始小写字段作为备用
          nodeCode: node.nodeCode || node.NodeCode,
          nodeName: node.nodeName || node.NodeName,
          nodeType: node.nodeType || node.NodeType,
        };
      });

      return {
        name: branch.name || `条件分支${branchIndex + 1}`,
        nodes: formattedNodes,
        condition: condition,
      };
    });
  },

  /**
   * 验证工作流数据的完整性
   * @param {Array} nodes - 工作流节点列表
   * @returns {Object} 验证结果 {isValid: boolean, errors: Array}
   */
  validateWorkflow(nodes) {
    const errors = [];

    if (!Array.isArray(nodes)) {
      errors.push('工作流数据必须为数组');
      return { isValid: false, errors };
    }

    // 检查是否有开始和结束节点
    const hasStart = nodes.some(node => node.NodeType === 'START');
    const hasEnd = nodes.some(node => node.NodeType === 'END');

    if (!hasStart) errors.push('缺少开始节点');
    if (!hasEnd) errors.push('缺少结束节点');

    // 检查节点完整性
    nodes.forEach((node, index) => {
      if (!NodeHelper.validateNode(node)) {
        errors.push(`第${index + 1}个节点数据不完整`);
      }

      // 检查条件节点的分支配置
      if (NodeHelper.isConditionNode(node)) {
        const branches = this.parseConditionConfig(node.ConditionConfig);
        if (branches.length === 0) {
          errors.push(`条件节点"${node.NodeName}"缺少有效分支配置`);
        } else {
          // 验证条件分支的配置
          branches.forEach((branch, branchIndex) => {
            if (!branch.condition) {
              errors.push(`条件节点"${node.NodeName}"的分支"${branch.name}"缺少条件配置`);
            } else {
              const condition = branch.condition;
              if (!condition.conditionType) {
                errors.push(`条件节点"${node.NodeName}"的分支"${branch.name}"缺少条件类型配置`);
              }
              if (!condition.operator) {
                errors.push(`条件节点"${node.NodeName}"的分支"${branch.name}"缺少操作符配置`);
              }
              if (
                (condition.conditionType === 'instance' &&
                  condition.fieldName === 'createdBy' &&
                  !condition.fieldValue) ||
                condition.fieldValue.trim() === ''
              ) {
                errors.push(`条件节点"${node.NodeName}"的分支"${branch.name}"需要选择至少一个人员`);
              }
              if (
                condition.conditionType === 'data' &&
                (!condition.fieldName || !condition.fieldValue)
              ) {
                errors.push(
                  `条件节点"${node.NodeName}"的分支"${branch.name}"需要填写数据字段名和值`,
                );
              }
            }
          });
        }
      }
    });

    return {
      isValid: errors.length === 0,
      errors,
    };
  },

  /**
   * 克隆工作流数据
   * @param {Array} nodes - 原始节点列表
   * @returns {Array} 克隆后的节点列表
   */
  cloneWorkflow(nodes) {
    return JsonHelper.safeParse(JsonHelper.safeStringify(nodes), []);
  },

  /**
   * 查找节点在工作流中的路径
   * @param {Array} nodes - 工作流节点列表
   * @param {string} nodeCode - 节点编码
   * @returns {Array} 节点路径数组
   */
  findNodePath(nodes, nodeCode) {
    const path = [];

    const self = this; // 保存 this 引用
    function findInArray(arr, targetCode, currentPath) {
      for (let i = 0; i < arr.length; i++) {
        const node = arr[i];
        const newPath = [...currentPath, { array: arr, index: i }];

        if (node.NodeCode === targetCode) {
          return newPath;
        }

        // 检查条件分支中的节点
        if (NodeHelper.isConditionNode(node)) {
          const config = JsonHelper.safeParse(node.ConditionConfig);
          if (config.branches && Array.isArray(config.branches)) {
            for (const branch of config.branches) {
              // 直接使用原始的分支节点数组，不进行解析
              if (branch.nodes && Array.isArray(branch.nodes)) {
                const result = findInArray(branch.nodes, targetCode, newPath);
                if (result) return result;
              }
            }
          }
        }
      }
      return null;
    }

    return findInArray(nodes, nodeCode, []) || [];
  },
};
