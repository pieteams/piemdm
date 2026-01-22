<template>
  <div
    class="workflow-builder bg-light"
    style="min-height: 100vh"
  >
    <div
      class="d-flex"
      style="height: 100vh"
    >
      <!-- 左侧工作流区域 -->
      <div class="workflow-area flex-grow-1 p-4">
        <div
          class="container"
          style="max-width: 900px"
        >
          <div class="workflow-tree">
            <WorkflowNode
              :nodes="workflowData"
              @add-node="onAddNode"
              @delete-node="onDeleteNode"
              @update-node="onUpdateNode"
            />
          </div>
        </div>

        <AddNodeModal
          ref="addNodeModal"
          @select="confirmAddNode"
        />
      </div>

      <!-- 右侧配置面板 -->
      <div class="config-panel-area">
        <NodePropertyPanel
          ref="nodePropertyPanel"
          @save="onSaveNodeProperty"
          @request-user-list="handleUserListRequest"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref, reactive } from 'vue';
  import WorkflowNode from './WorkflowNode.vue';
  import AddNodeModal from './AddNodeModal.vue';
  import NodePropertyPanel from './NodePropertyPanel.vue';
  import { WorkflowService } from '../services/workflow-service.js';
  import { WorkflowUtils } from '../utils/workflow-utils.js';
  import { JsonHelper } from '../utils/json-helper.js';
  import { v4 as uuidv4 } from 'uuid';
  import {
    defaultWorkflowData as importedDefaultWorkflowData,
    loadFromBackendData as importedLoadFromBackendData,
    getBackendData as importedGetBackendData,
  } from '../utils/workflow-builder-methods.js';

  // 定义事件类型
  const emit = defineEmits(['requestUserList']);

  // 使用共享的默认工作流数据
  const defaultWorkflowData = importedDefaultWorkflowData;

  // 工作流数据
  const workflowData = reactive(WorkflowUtils.processConditionConfig(defaultWorkflowData));
  const addNodeModal = ref(null);
  const nodePropertyPanel = ref(null);
  let pendingInsertionInfo = null;

  // 处理用户数据请求事件
  const handleUserListRequest = (options, callback) => {
    // 转发给主程序
    emit('requestUserList', options, callback);
  };

  // 工作流对象（用于服务层操作）
  const workflow = reactive({
    name: '示例工作流',
    nodes: workflowData,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    status: 'draft',
  });

  // 处理添加节点事件
  const onAddNode = (parentList, index) => {
    pendingInsertionInfo = { parentList, index };
    addNodeModal.value.show();
  };

  // 确认添加节点
  const confirmAddNode = (nodeType, nodeName) => {
    if (!pendingInsertionInfo) return;

    const { parentList, index } = pendingInsertionInfo;

    try {
      WorkflowService.addNode(workflow, parentList, index, nodeType, nodeName);
    } catch (error) {
      console.error('添加节点失败:', error.message);
    }

    pendingInsertionInfo = null;
  };

  // 处理删除节点事件
  const onDeleteNode = (parentList, index) => {
    const result = WorkflowService.deleteNode(workflow, parentList, index);
    if (!result.success) {
      console.error('删除节点失败:', result.error);
    }
  };

  // 处理更新节点事件（包含显示配置面板）
  const onUpdateNode = (node, newData) => {
    // 如果没有 newData 参数，说明是点击节点要显示配置面板
    if (!newData) {
      nodePropertyPanel.value.show(node);
      return;
    }

    const result = WorkflowService.updateNode(workflow, node.NodeCode, newData);
    if (!result.success) {
      console.error('更新节点失败:', result.error);
    }
  };

  // 处理保存节点属性
  const onSaveNodeProperty = (node, updateData) => {
    const result = WorkflowService.updateNode(workflow, node.NodeCode, updateData);
    if (!result.success) {
      console.error('保存节点属性失败:', result.error);
      alert('保存失败: ' + result.error);
    }
  };

  // 从后端数据加载工作流
  const loadFromBackendData = backendData => {
    // 调用共享的方法
    importedLoadFromBackendData(workflowData, backendData);

    // 更新工作流对象
    workflow.nodes.splice(0, workflow.nodes.length, ...workflowData);
    workflow.updatedAt = new Date().toISOString();
  };

  // 获取处理后端数据
  const getBackendData = () => {
    // 调用共享的方法
    return importedGetBackendData(workflowData);
  };

  // 暴露方法给父组件
  defineExpose({
    loadFromBackendData,
    getBackendData,
  });
</script>

<style>
  /* 可以在这里添加一些全局工作流样式 */
  .workflow-tree {
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .node-wrapper {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    margin: 1px 0; /* 上下1px间距 */
  }

  /* 连接线 */
  .connector {
    width: 2px;
    background-color: transparent;
    height: 34px;
    min-height: 34px;
    position: relative;
    margin: 1px 0; /* 上下1px间距 */
  }

  .add-node-btn {
    position: absolute;
    top: calc(50% - 1px);
    left: 50%;
    transform: translate(-50%, -50%);
    z-index: 10;
    width: 32px !important;
    height: 32px !important;
    border-radius: 50% !important;
    padding: 0 !important;
    display: flex !important;
    align-items: center !important;
    justify-content: center !important;
    font-size: 16px;
    line-height: 1;
    min-width: auto !important;
    min-height: auto !important;
  }

  .node {
    width: 280px;
    border-radius: 8px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    background-color: white;
    cursor: pointer;
    transition: box-shadow 0.2s ease;
    min-height: fit-content;
  }

  .node:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  }

  .node:has(.condition-branches) {
    width: auto;
    min-width: 320px;
    min-height: fit-content;
  }

  .node-start .node-content {
    background-color: #198754;
    color: white;
    border-radius: 8px; /* 改为与普通节点相同的圆角值 */
    display: flex;
    align-items: center;
    justify-content: flex-start; /* 改为左对齐 */
    padding: 12px 20px;
    min-height: 40px;
  }

  .node-end .node-content {
    background-color: #6c757d;
    color: white;
    border-radius: 8px; /* 改为与普通节点相同的圆角值 */
    display: flex;
    align-items: center;
    justify-content: flex-start; /* 改为左对齐 */
    padding: 12px 20px;
    min-height: 40px;
  }

  .node {
    width: 280px;
    border-radius: 8px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    background-color: white;
    cursor: pointer;
    transition: box-shadow 0.2s ease;
    min-height: fit-content;
    min-width: fit-content; /* 允许节点根据内容自适应宽度 */
  }

  .node-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid #eee;
    padding-bottom: 10px;
    margin-bottom: 10px;
  }

  /* 响应式布局 */
  .workflow-area {
    overflow-y: auto;
    border-right: 1px solid #dee2e6;
  }

  .config-panel-area {
    flex-shrink: 0;
  }

  @media (max-width: 768px) {
    .workflow-area {
      width: 100%;
    }

    .config-panel-area {
      position: fixed;
      top: 0;
      right: 0;
      height: 100vh;
      z-index: 1050;
    }
  }
</style>
