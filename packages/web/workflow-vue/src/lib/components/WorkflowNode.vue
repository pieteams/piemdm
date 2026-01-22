<template>
  <div
    v-for="(node, index) in nodes"
    :key="node.NodeCode || index"
    class="node-wrapper"
  >
    <div
      v-if="node && node.NodeType !== 'START'"
      class="connector"
    >
      <button
        class="btn btn-primary rounded-circle add-node-btn"
        @click="$emit('add-node', nodes, index)"
      >
        <i class="bi bi-plus-lg" />
      </button>
    </div>

    <div
      v-if="node"
      class="node"
    >
      <div
        v-if="node && node.NodeType === 'CONDITION'"
        class="node-content bg-light"
      >
        <div
          class="node-header condition-header d-flex justify-content-between align-items-center"
          style="cursor: pointer"
          @click="$emit('update-node', node)"
        >
          <span>{{ node.NodeName }}</span>
          <div>
            <button
              class="btn btn-sm btn-outline-light border-0 me-2"
              title="配置节点"
              @click.stop="$emit('update-node', node)"
            >
              <i class="bi bi-gear" />
            </button>
            <button
              class="btn btn-sm btn-outline-light border-0"
              @click.stop="$emit('delete-node', nodes, index)"
            >
              <i class="bi bi-trash" />
            </button>
          </div>
        </div>
        <div class="condition-branches d-flex px-2 pb-2">
          <div
            v-for="(branch, branchIndex) in getBranches(node)"
            :key="branchIndex"
            class="branch flex-grow-1"
          >
            <div
              class="branch-header text-center p-2 border rounded-top d-flex justify-content-center align-items-center"
            >
              <div>
                <div>{{ branch.name }}</div>
                <small
                  v-if="branch.condition"
                  class="text-muted"
                >
                  {{ getConditionDisplayText(branch.condition) }}
                </small>
              </div>
            </div>
            <div class="branch-content border-start border-end border-bottom rounded-bottom px-2">
              <!-- 分支节点详细显示 -->
              <WorkflowNode
                :nodes="branch.nodes || []"
                @add-node="(parentList, i) => $emit('add-node', parentList, i)"
                @delete-node="(parentList, i) => $emit('delete-node', parentList, i)"
                @update-node="(node, newData) => $emit('update-node', node, newData)"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 开始/结束节点：简单方框结构 -->
      <div
        v-else-if="node.NodeType === 'START' || node.NodeType === 'END'"
        :class="getNodeStyleClass(node)"
      >
        <div class="node-content">
          <i
            :class="getNodeStyleIcon(node)"
            class="me-2"
          />
          <span>{{ node.NodeName || '节点' }}</span>
        </div>
      </div>

      <!-- 其他节点：标准结构（标题+内容） -->
      <div
        v-else
        :class="getNodeStyleClass(node)"
        style="cursor: pointer"
        @click="$emit('update-node', node)"
      >
        <div
          class="node-header d-flex justify-content-between align-items-center"
          :class="getNodeHeaderClass(node)"
        >
          <span>
            <i
              :class="getNodeStyleIcon(node)"
              class="me-2"
            />
            {{ node.NodeName }}
          </span>
          <button
            v-if="canDeleteNode(node)"
            class="btn btn-sm btn-light border-0"
            @click.stop="$emit('delete-node', nodes, index)"
          >
            <i class="bi bi-trash" />
          </button>
        </div>
        <div class="node-body d-flex align-items-center">
          <p class="ms-2 text-muted">
            <span v-if="node.NodeType === 'APPROVAL'">审批人：</span>
            <span v-else-if="node.NodeType === 'CC'">接收人：</span>
            {{ getNodeApproverText(node) }}
          </p>
        </div>
      </div>
    </div>

    <div
      v-if="node && index === nodes.length - 1 && node.NodeType !== 'END'"
      class="connector"
    >
      <button
        class="btn btn-primary rounded-circle add-node-btn"
        @click="$emit('add-node', nodes, nodes.length)"
      >
        <i class="bi bi-plus-lg" />
      </button>
    </div>
  </div>
</template>

<script setup>
  import { isSystemNode } from '../constants/node-types.js';

  const props = defineProps({
    nodes: Array,
  });
  const emit = defineEmits(['add-node', 'delete-node', 'update-node']);

  // 根据节点类型获取对应的样式类
  const getNodeStyleClass = node => {
    if (!node) return 'node-approval';
    if (node.NodeType === 'APPROVAL') {
      if (node.ApproverType === 'AUTO_REJECT') return 'node-auto-reject';
      if (node.ApproverType === 'AUTO_APPROVE') return 'node-auto-approve';
    }
    const typeMap = {
      START: 'node-start',
      APPROVAL: 'node-approval',
      CONDITION: 'node-condition',
      CC: 'node-cc',
      END: 'node-end',
    };
    return typeMap[node.NodeType] || 'node-approval';
  };

  // 获取审批人文本
  const getNodeApproverText = node => {
    if (!node) return '未知节点';

    if (node.NodeType === 'START') return '流程发起人';
    if (node.NodeType === 'END') return '流程结束';

    if (node.NodeType === 'CC') {
      if (node.ApproverType === 'USERS' && node.ApproverConfig) {
        try {
          const config = JSON.parse(node.ApproverConfig);
          if (config.users && config.users.length > 0) {
            return config.users.join(', ');
          }
        } catch (e) {
          /* 静默处理 */
        }
      }
      return '未指定抄送人';
    }

    if (node.NodeType === 'APPROVAL') {
      if (node.ApproverType === 'AUTO_REJECT') return '系统自动驳回';
      if (node.ApproverType === 'AUTO_APPROVE') return '系统自动通过';

      if (node.ApproverType === 'USERS' && node.ApproverConfig) {
        try {
          const config = JSON.parse(node.ApproverConfig);
          if (config.users && config.users.length > 0) {
            return config.users.join(', ');
          }
        } catch (e) {
          /* 静默处理 */
        }
      }
      return '未指定审批人';
    }

    return '未指定';
  };

  // 获取节点图标类
  const getNodeStyleIcon = node => {
    if (!node) return 'bi-circle-fill';
    if (node.NodeType === 'APPROVAL') {
      if (node.ApproverType === 'AUTO_REJECT') return 'bi-x-circle-fill';
      if (node.ApproverType === 'AUTO_APPROVE') return 'bi-check-circle-fill';
    }
    const iconMap = {
      START: 'bi-play-circle-fill',
      END: 'bi-stop-circle-fill',
      APPROVAL: 'bi-person-check-fill',
      CONDITION: 'bi-diagram-3-fill',
      CC: 'bi-send-fill',
    };
    return iconMap[node.NodeType] || 'bi-circle-fill';
  };

  // 获取分支数据
  const getBranches = node => {
    if (!node || node.NodeType !== 'CONDITION') {
      return [];
    }

    try {
      if (!node.ConditionConfig || node.ConditionConfig === '{}') {
        return [];
      }

      const config = JSON.parse(node.ConditionConfig);
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

        const formattedNodes = (branch.nodes || []).map(nodeItem => {
          const formattedNode = {
            ID: nodeItem.id || nodeItem.ID,
            NodeCode: nodeItem.nodeCode || nodeItem.NodeCode,
            NodeName: nodeItem.nodeName || nodeItem.NodeName,
            NodeType: nodeItem.nodeType || nodeItem.NodeType,
            Description: nodeItem.description || nodeItem.Description || '',
            SortOrder: nodeItem.sortOrder || nodeItem.SortOrder || 0,
            ApproverType: nodeItem.approverType || nodeItem.ApproverType || 'USERS',
            ApproverConfig: nodeItem.approverConfig || nodeItem.ApproverConfig || '{}',
            ConditionConfig: nodeItem.conditionConfig || nodeItem.ConditionConfig || '{}',
            ApprovalDefCode: nodeItem.approvalDefCode || nodeItem.ApprovalDefCode || '',
            nodeCode: nodeItem.nodeCode || nodeItem.NodeCode,
            nodeName: nodeItem.nodeName || nodeItem.NodeName,
            nodeType: nodeItem.nodeType || nodeItem.NodeType,
          };

          // Migration for old data structure
          if (
            formattedNode.NodeType === 'AUTO_APPROVE' ||
            formattedNode.NodeType === 'AUTO_REJECT'
          ) {
            formattedNode.ApproverType = formattedNode.NodeType;
            formattedNode.NodeType = 'APPROVAL';
          }
          if (
            formattedNode.nodeType === 'AUTO_APPROVE' ||
            formattedNode.nodeType === 'AUTO_REJECT'
          ) {
            formattedNode.approverType = formattedNode.nodeType;
            formattedNode.nodeType = 'APPROVAL';
          }

          return formattedNode;
        });

        return {
          name: branch.name || `条件分支${branchIndex + 1}`,
          nodes: formattedNodes,
          condition: condition,
        };
      });
    } catch (error) {
      console.warn('获取分支数据失败:', error);
      return [];
    }
  };

  // 检查是否可以删除节点
  const canDeleteNode = node => {
    if (!node) return false;
    return !isSystemNode(node.NodeType);
  };

  // 获取节点header样式类
  const getNodeHeaderClass = node => {
    if (!node) return '';
    const classMap = {
      APPROVAL: 'approval-header',
      CC: 'approval-header',
      CONDITION: 'condition-header',
    };
    return classMap[node.NodeType] || '';
  };

  // 分支header样式类统一
  const getBranchHeaderClass = () => {
    return ''; // 所有分支使用统一样式
  };

  // 获取条件显示文本
  const getConditionDisplayText = condition => {
    if (!condition) return '';

    const operatorMap = {
      eq: '等于',
      ne: '不等于',
      gt: '大于',
      gte: '大于等于',
      lt: '小于',
      lte: '小于等于',
      contains: '包含',
      not_contains: '不包含',
      else: '其他情况',
    };

    const operatorText = operatorMap[condition.operator] || condition.operator;

    if (condition.conditionType === 'instance') {
      if (condition.fieldName === 'createdBy') {
        if (condition.fieldValue && condition.fieldValue.split(',').length > 0) {
          return `创建人 ${operatorText} ${condition.fieldValue.split(',').length}人`;
        }
        return `创建人 ${operatorText}`;
      }
      if (condition.operator === 'else') {
        return '其他情况';
      }
      return `${condition.fieldName} ${operatorText} ${condition.fieldValue}`;
    } else if (condition.conditionType === 'data') {
      return `${condition.fieldName} ${operatorText} ${condition.fieldValue}`;
    }

    // 向后兼容旧格式
    if (condition.fieldName && condition.operator) {
      return `${condition.fieldName} ${operatorText} ${condition.fieldValue || ''}`;
    }

    return '';
  };
</script>

<style scoped>
  .condition-branches {
    gap: 10px; /* 分支之间的间距 */
    display: flex;
    flex-direction: row;
    flex-wrap: nowrap;
    width: 100%;
    min-height: fit-content;
  }
  .branch {
    background-color: #f8f9fa;
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: fit-content;
  }
  .branch-content {
    padding: 10px 0;
    display: flex;
    flex: 1 1 auto;
    min-height: fit-content;
    flex-direction: column;
    align-items: center;
  }

  /* 节点类型样式 */
  .node-start .node-content {
    background-color: #198754;
    color: white;
  }
  .node-end .node-content {
    background-color: #6c757d;
    color: white;
  }
  .node-condition .node-content {
    background-color: #ffc107;
    color: #212529;
  }
  .node-approval .node-content {
    background-color: #0d6efd;
    color: white;
  }
  .node-cc .node-content {
    background-color: #6f42c1;
    color: white;
  }
  .node-auto-reject .node-content {
    background-color: #dc3545;
    color: white;
  }

  .node-auto-approve .node-content {
    background-color: #198754;
    color: white;
  }

  /* 节点Header样式 */
  .approval-header {
    background-color: var(--bs-primary) !important;
    color: var(--bs-white) !important;
    border-bottom: 1px solid var(--bs-primary);
    padding: 10px 15px;
    border-radius: 8px 8px 0 0;
  }

  .condition-header {
    background-color: var(--bs-info) !important;
    color: var(--bs-white) !important;
    border-bottom: 1px solid var(--bs-info);
    padding: 10px 15px;
    border-radius: 8px 8px 0 0;
  }

  /* 统一的分支Header样式 */
  .branch-header {
    background-color: var(--bs-light) !important;
    color: var(--bs-black) !important;
  }

  /* 删除按钮样式优化 */
  .approval-header .btn-light,
  .condition-header .btn-light {
    background-color: rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 255, 255, 0.2);
    color: white;
  }

  .approval-header .btn-light:hover,
  .condition-header .btn-light:hover {
    background-color: rgba(255, 255, 255, 0.2);
    border-color: rgba(255, 255, 255, 0.3);
    color: white;
  }

  .branch-first-header .btn-light,
  .branch-default-header .btn-light {
    background-color: rgba(0, 0, 0, 0.1);
    border-color: rgba(0, 0, 0, 0.2);
  }

  .branch-first-header .btn-light:hover,
  .branch-default-header .btn-light:hover {
    background-color: rgba(0, 0, 0, 0.2);
    border-color: rgba(0, 0, 0, 0.3);
  }
</style>
