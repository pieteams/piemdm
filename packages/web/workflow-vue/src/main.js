import { createApp } from 'vue';
import App from './App.vue';

// 引入 Bootstrap CSS
import 'bootstrap/dist/css/bootstrap.min.css';
// 引入 Bootstrap Icons CSS
import 'bootstrap-icons/font/bootstrap-icons.css';
// 引入 Bootstrap JS (用于 Modal 等组件)
import 'bootstrap';

// 导入修复后的工作流组件
import WorkflowVue from './lib/index.js';
import 'workflow-vue/style.css';

createApp(App).use(WorkflowVue).mount('#app');
