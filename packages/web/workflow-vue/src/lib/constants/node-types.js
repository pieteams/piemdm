/**
 * 节点类型常量定义
 * 统一管理所有节点类型的相关配置
 */

export const NODE_TYPES = {
  START: {
    type: 'START',
    name: '开始节点',
    description: '工作流的起始点',
    icon: 'bi-play-circle-fill',
    class: 'node-start',
    approverType: 'SYSTEM',
    deletable: false,
    editable: false,
  },

  APPROVAL: {
    type: 'APPROVAL',
    name: '审批节点',
    description: '需要指定人员进行审批',
    icon: 'bi-person-check-fill',
    class: 'node-approval',
    approverType: 'USERS',
    deletable: true,
    editable: true,
  },

  CONDITION: {
    type: 'CONDITION',
    name: '条件分支',
    description: '根据不同条件，走向不同流程',
    icon: 'bi-diagram-3-fill',
    class: 'node-condition',
    approverType: 'USERS',
    deletable: true,
    editable: true,
  },

  CC: {
    type: 'CC',
    name: '抄送节点',
    description: '抄送给相关人员',
    icon: 'bi-send-fill',
    class: 'node-cc',
    approverType: 'USERS',
    deletable: true,
    editable: true,
  },

  END: {
    type: 'END',
    name: '结束节点',
    description: '工作流的终止点',
    icon: 'bi-stop-circle-fill',
    class: 'node-end',
    approverType: 'SYSTEM',
    deletable: false,
    editable: false,
  },
};

/**
 * 可添加的节点类型列表（用于添加节点弹窗）
 */
export const ADDABLE_NODE_TYPES = [
  NODE_TYPES.APPROVAL,
  NODE_TYPES.CONDITION,
  NODE_TYPES.CC,
  {
    type: 'AUTO_APPROVE',
    name: '自动通过',
    description: '系统自动通过流程',
    icon: 'bi-check-circle-fill',
    class: 'node-auto-approve',
  },
  {
    type: 'AUTO_REJECT',
    name: '自动驳回',
    description: '系统自动驳回流程',
    icon: 'bi-x-circle-fill',
    class: 'node-auto-reject',
  },
];

/**
 * 系统节点类型（不可删除和编辑）
 */
export const SYSTEM_NODE_TYPES = [NODE_TYPES.START, NODE_TYPES.END];

/**
 * 审批节点类型（需要配置审批人）
 */
export const APPROVAL_NODE_TYPES = [NODE_TYPES.APPROVAL, NODE_TYPES.CC];

/**
 * 根据类型获取节点配置
 * @param {string} type - 节点类型
 * @returns {Object|null} 节点配置对象
 */
export function getNodeConfig(type) {
  return NODE_TYPES[type] || null;
}

/**
 * 检查节点是否可删除
 * @param {string} type - 节点类型
 * @returns {boolean} 是否可删除
 */
export function isNodeDeletable(type) {
  const config = getNodeConfig(type);
  return config ? config.deletable : true;
}

/**
 * 检查节点是否可编辑
 * @param {string} type - 节点类型
 * @returns {boolean} 是否可编辑
 */
export function isNodeEditable(type) {
  const config = getNodeConfig(type);
  return config ? config.editable : true;
}

/**
 * 检查节点是否为系统节点
 * @param {string} type - 节点类型
 * @returns {boolean} 是否为系统节点
 */
export function isSystemNode(type) {
  return SYSTEM_NODE_TYPES.some(node => node.type === type);
}

/**
 * 获取所有节点类型的名称映射
 * @returns {Object} 类型到名称的映射对象
 */
export function getNodeTypeNameMap() {
  const map = {};
  Object.values(NODE_TYPES).forEach(config => {
    map[config.type] = config.name;
  });
  // Manually add the auto types for name resolution
  map['AUTO_APPROVE'] = '自动通过';
  map['AUTO_REJECT'] = '自动驳回';
  return map;
}

/**
 * 根据节点类型获取节点名称
 * @param {string} type - 节点类型
 * @returns {string} 节点名称
 */
export function getNodeTypeName(type) {
  const map = getNodeTypeNameMap();
  return map[type] || '未知节点';
}
