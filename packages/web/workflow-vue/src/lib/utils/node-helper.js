/**
 * 节点操作辅助函数
 * 提供节点相关的通用操作方法
 */
import { v4 as uuidv4 } from 'uuid';
import { JsonHelper } from './json-helper.js';

export const NodeHelper = {
  /**
   * 生成唯一节点编码
   * @returns {string} 唯一节点编码
   */
  generateNodeCode() {
    return uuidv4();
  },

  /**
   * 生成唯一ID
   * @returns {number} 唯一ID
   */
  generateId() {
    return Date.now();
  },

  /**
   * 检查节点是否为条件分支类型
   * @param {Object} node - 节点对象
   * @returns {boolean} 是否为条件分支
   */
  isConditionNode(node) {
    return node.NodeType === 'CONDITION';
  },

  /**
   * 检查节点是否为开始或结束节点
   * @param {Object} node - 节点对象
   * @returns {boolean} 是否为开始或结束节点
   */
  isStartOrEndNode(node) {
    return ['START', 'END'].includes(node.NodeType);
  },

  /**
   * 检查节点是否为审批节点
   * @param {Object} node - 节点对象
   * @returns {boolean} 是否为审批节点
   */
  isApprovalNode(node) {
    return node.NodeType === 'APPROVAL';
  },

  /**
   * 获取节点的审批人文本
   * @param {Object} node - 节点对象
   * @returns {string} 审批人文本
   */
  getApproverText(node) {
    if (this.isStartOrEndNode(node)) {
      const textMap = {
        START: '流程发起人',
        END: '流程结束',
      };
      return textMap[node.NodeType] || '未知节点';
    }

    if (node.NodeType === 'APPROVAL') {
      if (node.ApproverType === 'AUTO_APPROVE') return '系统自动通过';
      if (node.ApproverType === 'AUTO_REJECT') return '系统自动驳回';
      if (node.ApproverType === 'USERS' && node.ApproverConfig) {
        const config = JsonHelper.safeParse(node.ApproverConfig);
        if (config.users && config.users.length > 0) {
          return config.users.join(', ');
        }
      }
      return '未指定审批人';
    }

    if (node.NodeType === 'CC') {
      return '抄送相关人员';
    }

    return '未指定';
  },

  /**
   * 获取节点图标类名
   * @param {Object} node - 节点对象
   * @returns {string} Bootstrap图标类名
   */
  getNodeIcon(node) {
    if (node && node.NodeType === 'APPROVAL') {
      if (node.ApproverType === 'AUTO_APPROVE') return 'bi-check-circle-fill';
      if (node.ApproverType === 'AUTO_REJECT') return 'bi-x-circle-fill';
    }

    const iconMap = {
      START: 'bi-play-circle-fill',
      END: 'bi-stop-circle-fill',
      APPROVAL: 'bi-person-check-fill',
      CONDITION: 'bi-diagram-3-fill',
      CC: 'bi-send-fill',
    };
    return iconMap[node.NodeType] || 'bi-circle-fill';
  },

  /**
   * 获取节点样式类名
   * @param {Object} node - 节点对象
   * @returns {string} CSS类名
   */
  getNodeClass(node) {
    if (node && node.NodeType === 'APPROVAL') {
      if (node.ApproverType === 'AUTO_APPROVE') return 'node-auto-approve';
      if (node.ApproverType === 'AUTO_REJECT') return 'node-auto-reject';
    }
    const classMap = {
      START: 'node-start',
      APPROVAL: 'node-approval',
      CONDITION: 'node-condition',
      CC: 'node-cc',
      END: 'node-end',
    };
    return classMap[node.NodeType] || 'node-approval';
  },

  /**
   * 验证节点数据完整性
   * @param {Object} node - 节点对象
   * @returns {boolean} 验证结果
   */
  validateNode(node) {
    if (!node) return false;

    const requiredFields = ['NodeCode', 'NodeName', 'NodeType'];
    return requiredFields.every(field => node[field]);
  },
};
