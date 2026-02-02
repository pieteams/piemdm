# workflow-vue

一个功能强大的 Vue 3 工作流构建器组件库,提供可视化的工作流设计和管理能力。

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Vue 3](https://img.shields.io/badge/Vue-3.4+-green.svg)](https://vuejs.org/)
[![Bootstrap 5](https://img.shields.io/badge/Bootstrap-5.3+-purple.svg)](https://getbootstrap.com/)

[English](./README.md) | [简体中文](./README_CN.md)

## ✨ 特性

- 🎨 **可视化设计** - 拖拽式工作流构建器,直观易用
- 🔧 **灵活配置** - 支持多种节点类型:审批、条件分支、抄送等
- 🚀 **开箱即用** - 提供完整的 Vue 插件和按需导入两种使用方式
- 📦 **轻量级** - 核心功能精简,依赖最小化
- 🎯 **TypeScript 支持** - 完整的类型定义
- 🧪 **测试覆盖** - 完善的单元测试

## 📦 安装

```bash
# 使用 pnpm (推荐)
pnpm add workflow-vue

# 使用 npm
npm install workflow-vue

# 使用 yarn
yarn add workflow-vue
```

## 🚀 快速开始

### 方式一:作为 Vue 插件使用(推荐)

```javascript
// main.js
import { createApp } from 'vue';
import App from './App.vue';
import { WorkflowVue } from 'workflow-vue';
import 'workflow-vue/style.css';

const app = createApp(App);
app.use(WorkflowVue);
app.mount('#app');
```

使用插件后,所有组件将自动注册为全局组件,可直接在模板中使用:

```vue
<template>
  <WorkflowBuilder v-model="workflowData" @save="handleSave" />
</template>
```

### 方式二:按需导入组件

```vue
<template>
  <WorkflowBuilder 
    v-model="workflowData"
    @save="handleSave"
  />
</template>

<script setup>
import { ref } from 'vue';
import { WorkflowBuilder } from 'workflow-vue';

const workflowData = ref(null);

const handleSave = (workflow) => {
  console.log('保存工作流:', workflow);
};
</script>
```

## 📚 核心 API

### 组件

#### WorkflowBuilder - 工作流构建器

主要的工作流可视化构建器组件。

```vue
<WorkflowBuilder 
  v-model="workflowData"
  :readonly="false"
  @save="handleSave"
  @cancel="handleCancel"
/>
```

**属性 (Props):**
- `modelValue` - 工作流数据对象
- `readonly` - 是否只读模式(默认: `false`)

**事件 (Events):**
- `update:modelValue` - 工作流数据更新时触发
- `save` - 点击保存按钮时触发
- `cancel` - 点击取消按钮时触发

#### WorkflowNode - 工作流节点

单个工作流节点组件,用于展示和编辑节点。

```vue
<WorkflowNode 
  :node="nodeData"
  :readonly="false"
  @edit="handleEdit"
  @delete="handleDelete"
/>
```

**属性 (Props):**
- `node` - 节点数据对象
- `readonly` - 是否只读模式

**事件 (Events):**
- `edit` - 编辑节点时触发
- `delete` - 删除节点时触发

### 服务层

#### WorkflowService - 工作流管理服务

提供工作流的创建、验证、序列化等功能。

```javascript
import { WorkflowService } from 'workflow-vue';

// 创建新工作流
const workflow = WorkflowService.createWorkflow('审批流程');

// 验证工作流是否有效
const isValid = WorkflowService.validateWorkflow(workflow);

// 序列化工作流为 JSON
const json = WorkflowService.serializeWorkflow(workflow);

// 从 JSON 反序列化工作流
const workflow = WorkflowService.deserializeWorkflow(json);
```

**主要方法:**
- `createWorkflow(name)` - 创建新工作流
- `validateWorkflow(workflow)` - 验证工作流
- `serializeWorkflow(workflow)` - 序列化为 JSON
- `deserializeWorkflow(json)` - 从 JSON 反序列化

#### NodeService - 节点管理服务

提供节点的创建、验证、克隆等功能。

```javascript
import { NodeService } from 'workflow-vue';

// 创建审批节点
const approvalNode = NodeService.createNode('APPROVAL', '部门审批');

// 创建抄送节点
const ccNode = NodeService.createNode('CC', '抄送HR');

// 验证节点配置
const isValid = NodeService.validateNode(node);

// 克隆节点
const clonedNode = NodeService.cloneNode(node);
```

**主要方法:**
- `createNode(type, name)` - 创建节点
- `validateNode(node)` - 验证节点
- `cloneNode(node)` - 克隆节点
- `updateNode(node, updates)` - 更新节点

### 常量定义

#### NODE_TYPES - 节点类型配置

所有支持的节点类型及其配置。

```javascript
import { NODE_TYPES } from 'workflow-vue';

console.log(NODE_TYPES.START);      // 开始节点
console.log(NODE_TYPES.APPROVAL);   // 审批节点
console.log(NODE_TYPES.CONDITION);  // 条件分支
console.log(NODE_TYPES.CC);         // 抄送节点
console.log(NODE_TYPES.END);        // 结束节点
```

每个节点类型包含以下配置:
- `type` - 节点类型标识
- `name` - 节点显示名称
- `description` - 节点描述
- `icon` - Bootstrap 图标类名
- `class` - CSS 样式类
- `deletable` - 是否可删除
- `editable` - 是否可编辑

#### ADDABLE_NODE_TYPES - 可添加节点列表

用于 UI 展示的可添加节点类型列表。

```javascript
import { ADDABLE_NODE_TYPES } from 'workflow-vue';

// 包含:审批、条件分支、抄送、自动通过、自动驳回
ADDABLE_NODE_TYPES.forEach(nodeType => {
  console.log(nodeType.name, nodeType.description);
});
```

### 工具函数

#### WorkflowUtils - 工作流工具

工作流相关的实用工具函数。

```javascript
import { WorkflowUtils } from 'workflow-vue';

// 根据 ID 查找节点
const node = WorkflowUtils.findNodeById(workflow, 'node-123');

// 获取工作流中的所有节点
const allNodes = WorkflowUtils.getAllNodes(workflow);

// 检测是否存在循环依赖
const hasCycle = WorkflowUtils.detectCycle(workflow);

// 获取节点的所有后继节点
const nextNodes = WorkflowUtils.getNextNodes(workflow, nodeId);
```

#### NodeHelper - 节点辅助函数

节点相关的辅助函数。

```javascript
import { NodeHelper } from 'workflow-vue';

// 获取节点图标
const icon = NodeHelper.getNodeIcon('APPROVAL');

// 获取节点样式类
const className = NodeHelper.getNodeClass('APPROVAL');

// 检查节点是否可删除
const deletable = NodeHelper.isNodeDeletable(node);

// 检查节点是否可编辑
const editable = NodeHelper.isNodeEditable(node);
```

#### JsonHelper - JSON 工具

JSON 序列化和反序列化工具。

```javascript
import { JsonHelper } from 'workflow-vue';

// 深度克隆对象
const cloned = JsonHelper.deepClone(obj);

// 安全的 JSON 解析
const data = JsonHelper.safeParse(jsonString, defaultValue);

// 格式化 JSON
const formatted = JsonHelper.stringify(obj, { pretty: true });
```

### 工作流引擎

#### WorkflowEngine - 执行引擎

工作流运行时执行引擎。

```javascript
import { createWorkflowEngine } from 'workflow-vue';

// 创建引擎实例
const engine = createWorkflowEngine(workflowData);

// 启动工作流执行
const result = await engine.execute({
  userId: 'user-123',
  formData: { amount: 5000, reason: '采购申请' }
});

// 获取当前执行节点
const currentNode = engine.getCurrentNode();

// 推进到下一节点
await engine.moveToNext({
  approved: true,
  comment: '同意'
});

// 获取执行历史
const history = engine.getHistory();
```

**主要方法:**
- `execute(context)` - 启动工作流
- `getCurrentNode()` - 获取当前节点
- `moveToNext(result)` - 推进到下一节点
- `getHistory()` - 获取执行历史
- `rollback()` - 回退到上一节点

## 🎯 节点类型说明

| 节点类型 | 说明 | 可删除 | 可编辑 | 图标 |
|---------|------|--------|--------|------|
| `START` | 开始节点,工作流的起点 | ❌ | ❌ | `bi-play-circle-fill` |
| `APPROVAL` | 审批节点,需要指定审批人 | ✅ | ✅ | `bi-person-check-fill` |
| `CONDITION` | 条件分支,根据条件走不同路径 | ✅ | ✅ | `bi-diagram-3-fill` |
| `CC` | 抄送节点,通知相关人员 | ✅ | ✅ | `bi-send-fill` |
| `END` | 结束节点,工作流的终点 | ❌ | ❌ | `bi-stop-circle-fill` |
| `AUTO_APPROVE` | 自动通过,系统自动批准 | ✅ | ✅ | `bi-check-circle-fill` |
| `AUTO_REJECT` | 自动驳回,系统自动拒绝 | ✅ | ✅ | `bi-x-circle-fill` |

## 💡 使用示例

### 示例 1: 简单审批流程

```vue
<template>
  <div class="workflow-container">
    <WorkflowBuilder 
      v-model="workflow"
      @save="saveWorkflow"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { WorkflowBuilder, WorkflowService } from 'workflow-vue';

// 创建初始工作流
const workflow = ref(
  WorkflowService.createWorkflow('请假审批流程')
);

// 保存工作流
const saveWorkflow = async (data) => {
  try {
    const json = WorkflowService.serializeWorkflow(data);
    await api.saveWorkflow(json);
    console.log('工作流保存成功');
  } catch (error) {
    console.error('保存失败:', error);
  }
};
</script>
```

### 示例 2: 只读模式展示

```vue
<template>
  <WorkflowBuilder 
    v-model="workflow"
    :readonly="true"
  />
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { WorkflowService } from 'workflow-vue';

const workflow = ref(null);

onMounted(async () => {
  const json = await api.getWorkflow(workflowId);
  workflow.value = WorkflowService.deserializeWorkflow(json);
});
</script>
```

### 示例 3: 编程式创建工作流

```javascript
import { WorkflowService, NodeService } from 'workflow-vue';

// 创建工作流
const workflow = WorkflowService.createWorkflow('采购审批');

// 添加审批节点
const deptApproval = NodeService.createNode('APPROVAL', '部门审批');
deptApproval.approvers = ['user-001', 'user-002'];

const financeApproval = NodeService.createNode('APPROVAL', '财务审批');
financeApproval.approvers = ['user-003'];

// 添加抄送节点
const ccNode = NodeService.createNode('CC', '抄送HR');
ccNode.ccUsers = ['user-004'];

// 构建工作流
workflow.nodes = [
  workflow.startNode,
  deptApproval,
  financeApproval,
  ccNode,
  workflow.endNode
];

// 保存
const json = WorkflowService.serializeWorkflow(workflow);
```

## 🔧 本地开发

### Monorepo 内部引用

在 PieMDM monorepo 中,`frontend` 通过 workspace 引用本地包:

```json
{
  "dependencies": {
    "workflow-vue": "workspace:*"
  }
}
```

**开发体验优势:**
- ✅ **无需预构建** - Vite 直接处理源码,无需先执行 `pnpm build`
- ✅ **热更新** - 修改 `workflow-vue` 源码后,`frontend` 自动刷新
- ✅ **类型提示** - TypeScript 完整支持,IDE 智能提示
- ✅ **调试友好** - 可直接在源码中设置断点调试

**为什么不需要预构建?**

在 pnpm workspace 模式下:
1. Vite 会直接处理 `src/lib/index.js` 源码
2. 不会使用 `package.json` 中的 `exports` 配置
3. 源码修改立即生效,无需重新构建

`package.json` 中的 `main`、`module`、`exports` 字段主要用于:
- 发布到 npm 后,外部项目安装使用
- TypeScript 类型定义文件引用

### 开发命令

```bash
# 安装依赖
pnpm install

# 启动开发服务器(带热更新)
pnpm dev

# 构建生产版本
pnpm build

# 运行单元测试
pnpm test

# 运行测试并生成覆盖率报告
pnpm test:coverage

# 监听模式运行测试
pnpm test:watch

# 测试 UI 界面
pnpm test:ui

# TypeScript 类型检查
pnpm type-check

# ESLint 代码检查
pnpm lint

# Prettier 代码格式化
pnpm format
```

### 项目结构

```
workflow-vue/
├── src/
│   ├── lib/                    # 📚 库源码
│   │   ├── components/         # 🎨 Vue 组件
│   │   │   ├── WorkflowBuilder.vue    # 工作流构建器
│   │   │   ├── WorkflowNode.vue       # 工作流节点
│   │   │   └── AddNodeModal.vue       # 添加节点弹窗
│   │   ├── services/           # 🔧 业务服务
│   │   │   ├── workflow-service.js    # 工作流服务
│   │   │   ├── node-service.js        # 节点服务
│   │   │   └── user-service.js        # 用户服务
│   │   ├── utils/              # 🛠️ 工具函数
│   │   │   ├── workflow-utils.js      # 工作流工具
│   │   │   ├── node-helper.js         # 节点辅助
│   │   │   ├── json-helper.js         # JSON 工具
│   │   │   └── validator.js           # 验证工具
│   │   ├── constants/          # 📋 常量定义
│   │   │   └── node-types.js          # 节点类型
│   │   ├── engine/             # ⚙️ 工作流引擎
│   │   │   └── workflow-engine.js     # 执行引擎
│   │   └── index.js            # 📦 入口文件
│   ├── App.vue                 # 🎯 开发预览应用
│   └── main.js                 # 🚀 开发入口
├── tests/                      # 🧪 测试文件
│   ├── unit/                   # 单元测试
│   └── integration/            # 集成测试
├── dist/                       # 📦 构建产物(自动生成)
├── package.json                # 📄 包配置
├── vite.config.js              # ⚡ Vite 配置
├── vitest.config.js            # 🧪 Vitest 配置
├── tsconfig.json               # 📘 TypeScript 配置
├── eslint.config.js            # 🔍 ESLint 配置
├── README.md                   # 📖 英文文档
└── README_CN.md                # 📖 中文文档
```

## 📦 发布到 npm

### 发布前检查清单

- [ ] 更新 `package.json` 中的 `version` 版本号
- [ ] 确保所有测试通过 (`pnpm test`)
- [ ] 确保构建成功 (`pnpm build`)
- [ ] 检查 TypeScript 类型 (`pnpm type-check`)
- [ ] 运行代码检查 (`pnpm lint`)
- [ ] 更新 `CHANGELOG.md`(如果有)
- [ ] 提交所有代码更改

### 版本管理

遵循 [语义化版本](https://semver.org/lang/zh-CN/) 规范:

```bash
# 在 packages/web/workflow-vue 目录下

# 补丁版本(bug 修复)
pnpm version patch  # 1.0.0 -> 1.0.1

# 次版本(新功能,向后兼容)
pnpm version minor  # 1.0.0 -> 1.1.0

# 主版本(破坏性变更)
pnpm version major  # 1.0.0 -> 2.0.0
```

或直接编辑 `package.json` 中的 `version` 字段。

### 发布到公开 npm

在 monorepo 根目录执行:

```bash
# 方式一:使用 -C 指定目录
pnpm -C packages/web/workflow-vue publish --access public

# 方式二:使用 filter (推荐,更符合 monorepo 习惯)
pnpm -r --filter workflow-vue publish --access public

# 发布 beta 版本
pnpm -r --filter workflow-vue publish --tag beta

# 发布指定版本
pnpm -r --filter workflow-vue publish --tag next
```

> **注意:** `package.json` 中已配置 `prepublishOnly` 脚本,发布前会自动执行 `pnpm build`。

### 发布到私有 Registry

如果需要发布到私有 npm registry:

```bash
# 方式一:临时指定 registry
pnpm -r --filter workflow-vue publish --registry https://your-registry.com

# 方式二:在 package.json 中配置
{
  "publishConfig": {
    "registry": "https://your-registry.com",
    "access": "restricted"
  }
}
```

### 发布 Scoped 包

如果需要发布为 scoped 包(如 `@pieteams/workflow-vue`):

1. 修改 `package.json` 中的 `name`:
```json
{
  "name": "@pieteams/workflow-vue"
}
```

2. 发布时指定 access:
```bash
pnpm publish --access public
```

### 何时考虑独立仓库

在 monorepo 子目录直接发布是最省事且可维护的方案。只有在以下情况下才考虑拆分为独立仓库:

- ❌ 需要完全独立的权限/可见性控制
- ❌ 需要完全独立的发布流程和版本管理
- ❌ 不希望消费者项目获取到 monorepo 相关元信息
- ❌ 需要独立的 CI/CD 流程

否则,保持在 monorepo 中更有优势:
- ✅ 统一的依赖管理
- ✅ 代码共享更方便
- ✅ 重构影响范围可控
- ✅ 本地开发体验更好

## 🔗 依赖要求

### Peer Dependencies (需要在使用项目中安装)

这些依赖需要在使用 `workflow-vue` 的项目中安装:

- **Vue** `^3.4.0` - Vue 3 框架
- **Bootstrap** `^5.3.0` - UI 样式框架
- **Bootstrap Icons** `^1.11.0` - 图标库

安装命令:
```bash
pnpm add vue@^3.4.0 bootstrap@^5.3.0 bootstrap-icons@^1.11.0
```

### Runtime Dependencies (自动安装)

这些依赖会在安装 `workflow-vue` 时自动安装:

- `uuid` `^9.0.1` - UUID 生成工具
- `vue-select` `4.0.0-beta.6` - 下拉选择组件

## 🤝 贡献指南

欢迎贡献代码!请遵循以下步骤:

1. **Fork 本仓库**
2. **创建特性分支** (`git checkout -b feature/amazing-feature`)
3. **提交更改** (`git commit -m 'feat: add amazing feature'`)
4. **推送到分支** (`git push origin feature/amazing-feature`)
5. **创建 Pull Request**

### Commit 规范

遵循 [Conventional Commits](https://www.conventionalcommits.org/zh-hans/) 规范:

- `feat:` - 新功能
- `fix:` - Bug 修复
- `docs:` - 文档更新
- `style:` - 代码格式调整(不影响功能)
- `refactor:` - 代码重构(既不是新功能也不是 Bug 修复)
- `perf:` - 性能优化
- `test:` - 测试相关
- `chore:` - 构建/工具链更新

示例:
```bash
git commit -m "feat(workflow): 添加工作流导出功能"
git commit -m "fix(node): 修复节点删除时的内存泄漏"
git commit -m "docs: 更新 API 文档"
```

### 代码规范

- 使用 ESLint 进行代码检查
- 使用 Prettier 进行代码格式化
- 编写单元测试覆盖新功能
- 更新相关文档

## 📄 许可证

[MIT](./LICENSE) © PieTeams

## 🔗 相关链接

- [GitHub 仓库](https://github.com/pieteams/piemdm)
- [问题反馈](https://github.com/pieteams/piemdm/issues)
- [PieMDM 文档](https://github.com/pieteams/piemdm/tree/main/docs)
- [更新日志](https://github.com/pieteams/piemdm/releases)

## 💬 获取帮助

如有问题或建议,请:

- 📝 提交 [Issue](https://github.com/pieteams/piemdm/issues)
- 📖 查看 [文档](https://github.com/pieteams/piemdm/tree/main/packages/web/workflow-vue)
- 💬 联系维护团队

## 🙏 致谢

感谢所有为本项目做出贡献的开发者!

---

**Made with ❤️ by PieTeams**