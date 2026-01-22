// src/lib/index.js

// 1. 导入组件
import WorkflowBuilder from './components/WorkflowBuilder.vue';
import WorkflowNode from './components/WorkflowNode.vue';
import AddNodeModal from './components/AddNodeModal.vue';

// 2. 导入服务层
import { WorkflowService } from './services/workflow-service.js';
import { NodeService } from './services/node-service.js';

// 3. 导入工具函数
import { WorkflowUtils } from './utils/workflow-utils.js';
import { NodeHelper } from './utils/node-helper.js';
import { JsonHelper } from './utils/json-helper.js';

// 4. 导入常量
import { NODE_TYPES, ADDABLE_NODE_TYPES } from './constants/node-types.js';

// 5. 导入工作流引擎
import { WorkflowEngine, createWorkflowEngine } from './engine/workflow-engine.js';

// 6. Vue 插件
const WorkflowVue = {
  install(app, options) {
    // 在这里将你的组件注册为全局组件
    app.component('WorkflowBuilder', WorkflowBuilder);
    app.component('WorkflowNode', WorkflowNode);
    app.component('AddNodeModal', AddNodeModal);

    // 提供全局服务和工具
    app.config.globalProperties.$workflowService = WorkflowService;
    app.config.globalProperties.$nodeService = NodeService;
    app.config.globalProperties.$workflowUtils = WorkflowUtils;
    app.config.globalProperties.$nodeHelper = NodeHelper;
    app.config.globalProperties.$jsonHelper = JsonHelper;
    app.config.globalProperties.$createWorkflowEngine = createWorkflowEngine;
  },
};

// 7. 命名导出 - 所有组件、服务、工具、常量等
export {
  // 组件
  WorkflowBuilder,
  WorkflowNode,
  AddNodeModal,

  // 服务
  WorkflowService,
  NodeService,

  // 工具
  WorkflowUtils,
  NodeHelper,
  JsonHelper,

  // 常量
  NODE_TYPES,
  ADDABLE_NODE_TYPES,

  // 工作流引擎
  WorkflowEngine,
  createWorkflowEngine,

  // Vue 插件（作为命名导出）
  WorkflowVue,
};

// 注意：已移除默认导出，只使用具名导出
// 其他项目应该使用具名导入，例如：
// import { WorkflowBuilder, WorkflowService } from 'workflow-vue';
// 或者使用插件：
// import { WorkflowVue } from 'workflow-vue';
// app.use(WorkflowVue);
