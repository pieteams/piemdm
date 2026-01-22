# workflow-vue

A structured workflow builder component for Vue 3.

## 安装

```sh
pnpm add workflow-vue
# 或
npm install workflow-vue
```

## 使用方式

### 1. 作为Vue插件使用

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

### 2. 按需导入组件

```javascript
// 在Vue组件中
import { WorkflowBuilder } from 'workflow-vue';

export default {
  components: {
    WorkflowBuilder
  }
};
```

### 3. 使用服务和工具

```javascript
import { WorkflowService, NodeService } from 'workflow-vue';

// 创建工作流
const workflow = WorkflowService.createWorkflow('工作流名称');

// 创建节点
const node = NodeService.createNode('APPROVAL', '审批节点');
```

### 4. 使用常量

```javascript
import { NODE_TYPES } from 'workflow-vue';

console.log(NODE_TYPES.START); // 开始节点配置
```

## 开发

```sh
# 安装依赖
pnpm install

# 开发模式
pnpm dev

# 构建
pnpm build

# 测试
pnpm test
```

## 依赖要求

- Vue 3.4+
- Bootstrap 5.3+
- Bootstrap Icons 1.11+

## 许可证

MIT