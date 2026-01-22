/**
 * 节点操作服务
 * 提供节点的创建、更新、删除等核心业务逻辑
 */
import { NodeHelper } from '../utils/node-helper.js';
import { JsonHelper } from '../utils/json-helper.js';
import { NODE_TYPES } from '../constants/node-types.js';

export const NodeService = {
  /**
   * 创建新节点
   * @param {string} nodeType - 节点类型
   * @param {string} nodeName - 节点名称
   * @param {Object} extraConfig - 额外配置
   * @returns {Object} 新节点对象
   */
  createNode(nodeType, nodeName, extraConfig = {}) {
    const isAutoApprove = nodeType === 'AUTO_APPROVE';
    const isAutoReject = nodeType === 'AUTO_REJECT';

    if (isAutoApprove || isAutoReject) {
      const name = nodeName || (isAutoApprove ? '自动通过' : '自动驳回');
      const baseNode = {
        ID: null,
        NodeCode: NodeHelper.generateNodeCode(),
        NodeName: name,
        NodeType: 'APPROVAL',
        Description: '',
        SortOrder: 0,
        ApproverType: nodeType,
        ApproverConfig: JsonHelper.safeStringify({
          type: nodeType,
          mode: 'AUTO',
        }),
        ConditionConfig: '{}',
        ApprovalDefCode: '',
      };
      return { ...baseNode, ...extraConfig };
    }

    const nodeConfig = NODE_TYPES[nodeType];
    if (!nodeConfig) {
      throw new Error(`未知的节点类型: ${nodeType}`);
    }

    const baseNode = {
      ID: null, // 新添加的节点ID为空，在保存到数据库时自动生成
      NodeCode: NodeHelper.generateNodeCode(),
      NodeName: nodeName || nodeConfig.name,
      NodeType: nodeType,
      Description: '',
      SortOrder: 0,
      ApproverType: nodeConfig.approverType,
      ApproverConfig: '{}',
      ConditionConfig: '{}',
      ApprovalDefCode: '',
    };

    // 特殊配置处理
    if (nodeType === 'CONDITION') {
      baseNode.ConditionConfig = JsonHelper.safeStringify({
        branches: [
          {
            id: NodeHelper.generateNodeCode(),
            name: '分支1',
            condition: {
              conditionType: 'instance',
              fieldName: 'createdBy',
              operator: 'eq',
              fieldValue: '',
            },
            nodes: [
              {
                ID: null,
                NodeCode: NodeHelper.generateNodeCode(),
                NodeName: '审批节点',
                NodeType: 'APPROVAL',
                SortOrder: 0,
                ApproverType: 'USERS',
                ApproverConfig: JsonHelper.safeStringify({
                  type: 'USERS',
                  users: [],
                  mode: 'OR',
                }),
                ConditionConfig: '{}',
              },
            ],
          },
          {
            id: NodeHelper.generateNodeCode(),
            name: '其他',
            condition: {
              conditionType: '',
              fieldName: '',
              operator: 'eq',
              fieldValue: '',
            },
            nodes: [
              {
                ID: null,
                NodeCode: NodeHelper.generateNodeCode(),
                NodeName: '自动驳回',
                NodeType: 'APPROVAL',
                SortOrder: 0,
                ApproverType: 'AUTO_REJECT',
                ApproverConfig: JsonHelper.safeStringify({
                  type: 'AUTO_REJECT',
                  mode: 'AUTO',
                }),
                ConditionConfig: '{}',
              },
            ],
          },
        ],
      });
    }

    if (nodeType === 'APPROVAL' || nodeType === 'CC') {
      baseNode.ApproverConfig = JsonHelper.safeStringify({
        type: nodeType === 'CC' ? 'CC' : 'USERS',
        users: [],
        mode: nodeType === 'CC' ? 'CC' : 'OR',
        ccTiming: nodeType === 'CC' ? 'after_approval' : undefined,
      });
    }

    return { ...baseNode, ...extraConfig };
  },

  /**
   * 更新节点信息
   * @param {Object} node - 原始节点
   * @param {Object} updates - 更新数据
   * @returns {Object} 更新后的节点
   */
  updateNode(node, updates) {
    if (!NodeHelper.validateNode(node)) {
      throw new Error('无效的节点数据');
    }

    const updatedNode = { ...node, ...updates };

    // Ensure data consistency for auto-approval nodes
    const approverType = updatedNode.ApproverType;
    if (approverType === 'AUTO_APPROVE' || approverType === 'AUTO_REJECT') {
      updatedNode.NodeType = 'APPROVAL';
    }

    // 确保关键字段不被意外修改
    updatedNode.ID = node.ID;
    updatedNode.NodeCode = node.NodeCode;

    return updatedNode;
  },

  /**
   * 删除节点验证
   * @param {Object} node - 要删除的节点
   * @param {Array} workflowNodes - 工作流所有节点
   * @returns {Object} 验证结果 {canDelete: boolean, reason: string}
   */
  canDeleteNode(node, workflowNodes) {
    // 检查是否为系统节点
    if (NodeHelper.isStartOrEndNode(node)) {
      return { canDelete: false, reason: '开始和结束节点不能删除' };
    }

    // 检查是否为唯一节点
    if (workflowNodes.length <= 2) {
      return { canDelete: false, reason: '工作流至少需要保留开始和结束节点' };
    }

    // 检查是否有后续依赖（这里可以扩展更复杂的依赖检查）
    const nodeIndex = workflowNodes.findIndex(n => n.NodeCode === node.NodeCode);
    const nextNode = workflowNodes[nodeIndex + 1];

    if (nextNode && NodeHelper.isStartOrEndNode(nextNode) && nextNode.NodeType === 'END') {
      // 如果下一个是结束节点，可以删除
      return { canDelete: true };
    }

    return { canDelete: true };
  },

  /**
   * 复制节点
   * @param {Object} node - 要复制的节点
   * @param {string} newName - 新节点名称
   * @returns {Object} 复制的新节点
   */
  cloneNode(node, newName) {
    if (!NodeHelper.validateNode(node)) {
      throw new Error('无效的节点数据');
    }

    const clonedNode = JsonHelper.safeParse(JsonHelper.safeStringify(node));
    clonedNode.ID = NodeHelper.generateId();
    clonedNode.NodeCode = NodeHelper.generateNodeCode();
    clonedNode.NodeName = newName || `${node.NodeName} (副本)`;

    return clonedNode;
  },

  /**
   * 设置节点审批人配置
   * @param {Object} node - 节点对象
   * @param {Array} users - 审批人列表
   * @param {string} mode - 审批模式 (OR/AND)
   * @returns {Object} 更新后的节点
   */
  setNodeApprovers(node, users = [], mode = 'OR') {
    if (!NodeHelper.isApprovalNode(node) && node.NodeType !== 'CC') {
      throw new Error('只有审批节点和抄送节点可以设置审批人');
    }

    const approverConfig = {
      type: node.NodeType === 'CC' ? 'CC' : 'USERS',
      users,
      mode: node.NodeType === 'CC' ? 'CC' : mode,
    };

    if (node.NodeType === 'CC') {
      approverConfig.ccTiming = 'after_approval';
    }

    return this.updateNode(node, {
      ApproverConfig: JsonHelper.safeStringify(approverConfig),
    });
  },

  /**
   * 设置条件节点分支配置
   * @param {Object} node - 条件节点
   * @param {Array} branches - 分支配置
   * @returns {Object} 更新后的节点
   */
  setConditionBranches(node, branches) {
    if (!NodeHelper.isConditionNode(node)) {
      throw new Error('只有条件节点可以设置分支配置');
    }

    if (!Array.isArray(branches) || branches.length < 2) {
      throw new Error('条件分支至少需要2个分支');
    }

    return this.updateNode(node, {
      ConditionConfig: JsonHelper.safeStringify({ branches }),
    });
  },

  /**
   * 获取节点的显示名称
   * @param {Object} node - 节点对象
   * @returns {string} 显示名称
   */
  getDisplayName(node) {
    if (!node) return '未知节点';

    const approverText = NodeHelper.getApproverText(node);
    return `${node.NodeName} (${approverText})`;
  },
};
