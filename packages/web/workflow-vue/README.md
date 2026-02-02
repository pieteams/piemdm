# workflow-vue

A powerful Vue 3 workflow builder component library that provides visual workflow design and management capabilities.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Vue 3](https://img.shields.io/badge/Vue-3.4+-green.svg)](https://vuejs.org/)
[![Bootstrap 5](https://img.shields.io/badge/Bootstrap-5.3+-purple.svg)](https://getbootstrap.com/)

[English](./README.md) | [简体中文](./README_CN.md)

## ✨ Features

- 🎨 **Visual Design** - Intuitive drag-and-drop workflow builder
- 🔧 **Flexible Configuration** - Support for multiple node types: approval, conditional branches, CC, etc.
- 🚀 **Ready to Use** - Available as both Vue plugin and on-demand imports
- 📦 **Lightweight** - Minimal core functionality with optimized dependencies
- 🎯 **TypeScript Support** - Complete type definitions
- 🧪 **Test Coverage** - Comprehensive unit tests

## 📦 Installation

```bash
# Using pnpm (recommended)
pnpm add workflow-vue

# Using npm
npm install workflow-vue

# Using yarn
yarn add workflow-vue
```

## 🚀 Quick Start

### Option 1: Use as Vue Plugin (Recommended)

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

After installing the plugin, all components are automatically registered globally and can be used directly in templates:

```vue
<template>
  <WorkflowBuilder v-model="workflowData" @save="handleSave" />
</template>
```

### Option 2: Import Components On-Demand

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
  console.log('Save workflow:', workflow);
};
</script>
```

## 📚 Core API

### Components

#### WorkflowBuilder - Workflow Builder

The main visual workflow builder component.

```vue
<WorkflowBuilder 
  v-model="workflowData"
  :readonly="false"
  @save="handleSave"
  @cancel="handleCancel"
/>
```

**Props:**
- `modelValue` - Workflow data object
- `readonly` - Read-only mode (default: `false`)

**Events:**
- `update:modelValue` - Emitted when workflow data is updated
- `save` - Emitted when save button is clicked
- `cancel` - Emitted when cancel button is clicked

#### WorkflowNode - Workflow Node

Individual workflow node component for displaying and editing nodes.

```vue
<WorkflowNode 
  :node="nodeData"
  :readonly="false"
  @edit="handleEdit"
  @delete="handleDelete"
/>
```

**Props:**
- `node` - Node data object
- `readonly` - Read-only mode

**Events:**
- `edit` - Emitted when editing a node
- `delete` - Emitted when deleting a node

### Services

#### WorkflowService - Workflow Management Service

Provides workflow creation, validation, serialization, and other functionalities.

```javascript
import { WorkflowService } from 'workflow-vue';

// Create a new workflow
const workflow = WorkflowService.createWorkflow('Approval Process');

// Validate workflow
const isValid = WorkflowService.validateWorkflow(workflow);

// Serialize workflow to JSON
const json = WorkflowService.serializeWorkflow(workflow);

// Deserialize workflow from JSON
const workflow = WorkflowService.deserializeWorkflow(json);
```

**Main Methods:**
- `createWorkflow(name)` - Create a new workflow
- `validateWorkflow(workflow)` - Validate workflow
- `serializeWorkflow(workflow)` - Serialize to JSON
- `deserializeWorkflow(json)` - Deserialize from JSON

#### NodeService - Node Management Service

Provides node creation, validation, cloning, and other functionalities.

```javascript
import { NodeService } from 'workflow-vue';

// Create approval node
const approvalNode = NodeService.createNode('APPROVAL', 'Department Approval');

// Create CC node
const ccNode = NodeService.createNode('CC', 'CC to HR');

// Validate node configuration
const isValid = NodeService.validateNode(node);

// Clone node
const clonedNode = NodeService.cloneNode(node);
```

**Main Methods:**
- `createNode(type, name)` - Create a node
- `validateNode(node)` - Validate a node
- `cloneNode(node)` - Clone a node
- `updateNode(node, updates)` - Update a node

### Constants

#### NODE_TYPES - Node Type Configuration

All supported node types and their configurations.

```javascript
import { NODE_TYPES } from 'workflow-vue';

console.log(NODE_TYPES.START);      // Start node
console.log(NODE_TYPES.APPROVAL);   // Approval node
console.log(NODE_TYPES.CONDITION);  // Conditional branch
console.log(NODE_TYPES.CC);         // CC node
console.log(NODE_TYPES.END);        // End node
```

Each node type includes the following configuration:
- `type` - Node type identifier
- `name` - Node display name
- `description` - Node description
- `icon` - Bootstrap icon class name
- `class` - CSS style class
- `deletable` - Whether it can be deleted
- `editable` - Whether it can be edited

#### ADDABLE_NODE_TYPES - Addable Node List

List of addable node types for UI display.

```javascript
import { ADDABLE_NODE_TYPES } from 'workflow-vue';

// Includes: Approval, Conditional, CC, Auto Approve, Auto Reject
ADDABLE_NODE_TYPES.forEach(nodeType => {
  console.log(nodeType.name, nodeType.description);
});
```

### Utilities

#### WorkflowUtils - Workflow Utilities

Utility functions related to workflows.

```javascript
import { WorkflowUtils } from 'workflow-vue';

// Find node by ID
const node = WorkflowUtils.findNodeById(workflow, 'node-123');

// Get all nodes in the workflow
const allNodes = WorkflowUtils.getAllNodes(workflow);

// Detect circular dependencies
const hasCycle = WorkflowUtils.detectCycle(workflow);

// Get all successor nodes of a node
const nextNodes = WorkflowUtils.getNextNodes(workflow, nodeId);
```

#### NodeHelper - Node Helper Functions

Helper functions related to nodes.

```javascript
import { NodeHelper } from 'workflow-vue';

// Get node icon
const icon = NodeHelper.getNodeIcon('APPROVAL');

// Get node style class
const className = NodeHelper.getNodeClass('APPROVAL');

// Check if node is deletable
const deletable = NodeHelper.isNodeDeletable(node);

// Check if node is editable
const editable = NodeHelper.isNodeEditable(node);
```

#### JsonHelper - JSON Utilities

JSON serialization and deserialization utilities.

```javascript
import { JsonHelper } from 'workflow-vue';

// Deep clone object
const cloned = JsonHelper.deepClone(obj);

// Safe JSON parsing
const data = JsonHelper.safeParse(jsonString, defaultValue);

// Format JSON
const formatted = JsonHelper.stringify(obj, { pretty: true });
```

### Workflow Engine

#### WorkflowEngine - Execution Engine

Workflow runtime execution engine.

```javascript
import { createWorkflowEngine } from 'workflow-vue';

// Create engine instance
const engine = createWorkflowEngine(workflowData);

// Start workflow execution
const result = await engine.execute({
  userId: 'user-123',
  formData: { amount: 5000, reason: 'Purchase Request' }
});

// Get current execution node
const currentNode = engine.getCurrentNode();

// Move to next node
await engine.moveToNext({
  approved: true,
  comment: 'Approved'
});

// Get execution history
const history = engine.getHistory();
```

**Main Methods:**
- `execute(context)` - Start workflow
- `getCurrentNode()` - Get current node
- `moveToNext(result)` - Move to next node
- `getHistory()` - Get execution history
- `rollback()` - Rollback to previous node

## 🎯 Node Types

| Node Type | Description | Deletable | Editable | Icon |
|-----------|-------------|-----------|----------|------|
| `START` | Start node, workflow entry point | ❌ | ❌ | `bi-play-circle-fill` |
| `APPROVAL` | Approval node, requires approvers | ✅ | ✅ | `bi-person-check-fill` |
| `CONDITION` | Conditional branch, different paths based on conditions | ✅ | ✅ | `bi-diagram-3-fill` |
| `CC` | CC node, notify relevant personnel | ✅ | ✅ | `bi-send-fill` |
| `END` | End node, workflow termination point | ❌ | ❌ | `bi-stop-circle-fill` |
| `AUTO_APPROVE` | Auto approve, system automatically approves | ✅ | ✅ | `bi-check-circle-fill` |
| `AUTO_REJECT` | Auto reject, system automatically rejects | ✅ | ✅ | `bi-x-circle-fill` |

## 💡 Usage Examples

### Example 1: Simple Approval Process

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

// Create initial workflow
const workflow = ref(
  WorkflowService.createWorkflow('Leave Approval Process')
);

// Save workflow
const saveWorkflow = async (data) => {
  try {
    const json = WorkflowService.serializeWorkflow(data);
    await api.saveWorkflow(json);
    console.log('Workflow saved successfully');
  } catch (error) {
    console.error('Save failed:', error);
  }
};
</script>
```

### Example 2: Read-Only Display

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

### Example 3: Programmatic Workflow Creation

```javascript
import { WorkflowService, NodeService } from 'workflow-vue';

// Create workflow
const workflow = WorkflowService.createWorkflow('Purchase Approval');

// Add approval nodes
const deptApproval = NodeService.createNode('APPROVAL', 'Department Approval');
deptApproval.approvers = ['user-001', 'user-002'];

const financeApproval = NodeService.createNode('APPROVAL', 'Finance Approval');
financeApproval.approvers = ['user-003'];

// Add CC node
const ccNode = NodeService.createNode('CC', 'CC to HR');
ccNode.ccUsers = ['user-004'];

// Build workflow
workflow.nodes = [
  workflow.startNode,
  deptApproval,
  financeApproval,
  ccNode,
  workflow.endNode
];

// Save
const json = WorkflowService.serializeWorkflow(workflow);
```

## 🔧 Local Development

### Monorepo Internal Reference

In the PieMDM monorepo, `frontend` references the local package via workspace:

```json
{
  "dependencies": {
    "workflow-vue": "workspace:*"
  }
}
```

**Development Experience Advantages:**
- ✅ **No Pre-build Required** - Vite processes source code directly, no need to run `pnpm build` first
- ✅ **Hot Module Replacement** - Modifying `workflow-vue` source code automatically refreshes `frontend`
- ✅ **Type Hints** - Full TypeScript support with IDE intellisense
- ✅ **Debug Friendly** - Can set breakpoints directly in source code

**Why No Pre-build Required?**

In pnpm workspace mode:
1. Vite processes `src/lib/index.js` source code directly
2. Does not use the `exports` configuration in `package.json`
3. Source code changes take effect immediately without rebuilding

The `main`, `module`, and `exports` fields in `package.json` are mainly used for:
- External projects after publishing to npm
- TypeScript type definition file references

### Development Commands

```bash
# Install dependencies
pnpm install

# Start development server (with HMR)
pnpm dev

# Build production version
pnpm build

# Run unit tests
pnpm test

# Run tests with coverage report
pnpm test:coverage

# Run tests in watch mode
pnpm test:watch

# Test UI interface
pnpm test:ui

# TypeScript type checking
pnpm type-check

# ESLint code checking
pnpm lint

# Prettier code formatting
pnpm format
```

### Project Structure

```
workflow-vue/
├── src/
│   ├── lib/                    # 📚 Library source code
│   │   ├── components/         # 🎨 Vue components
│   │   │   ├── WorkflowBuilder.vue    # Workflow builder
│   │   │   ├── WorkflowNode.vue       # Workflow node
│   │   │   └── AddNodeModal.vue       # Add node modal
│   │   ├── services/           # 🔧 Business services
│   │   │   ├── workflow-service.js    # Workflow service
│   │   │   ├── node-service.js        # Node service
│   │   │   └── user-service.js        # User service
│   │   ├── utils/              # 🛠️ Utility functions
│   │   │   ├── workflow-utils.js      # Workflow utilities
│   │   │   ├── node-helper.js         # Node helpers
│   │   │   ├── json-helper.js         # JSON utilities
│   │   │   └── validator.js           # Validation utilities
│   │   ├── constants/          # 📋 Constants
│   │   │   └── node-types.js          # Node types
│   │   ├── engine/             # ⚙️ Workflow engine
│   │   │   └── workflow-engine.js     # Execution engine
│   │   └── index.js            # 📦 Entry file
│   ├── App.vue                 # 🎯 Development preview app
│   └── main.js                 # 🚀 Development entry
├── tests/                      # 🧪 Test files
│   ├── unit/                   # Unit tests
│   └── integration/            # Integration tests
├── dist/                       # 📦 Build output (auto-generated)
├── package.json                # 📄 Package configuration
├── vite.config.js              # ⚡ Vite configuration
├── vitest.config.js            # 🧪 Vitest configuration
├── tsconfig.json               # 📘 TypeScript configuration
├── eslint.config.js            # 🔍 ESLint configuration
├── README.md                   # 📖 English documentation
└── README_CN.md                # 📖 Chinese documentation
```

## 📦 Publishing to npm

### Pre-publish Checklist

- [ ] Update `version` in `package.json`
- [ ] Ensure all tests pass (`pnpm test`)
- [ ] Ensure build succeeds (`pnpm build`)
- [ ] Check TypeScript types (`pnpm type-check`)
- [ ] Run code linting (`pnpm lint`)
- [ ] Update `CHANGELOG.md` (if applicable)
- [ ] Commit all code changes

### Version Management

Follow [Semantic Versioning](https://semver.org/) specification:

```bash
# In packages/web/workflow-vue directory

# Patch version (bug fixes)
pnpm version patch  # 1.0.0 -> 1.0.1

# Minor version (new features, backward compatible)
pnpm version minor  # 1.0.0 -> 1.1.0

# Major version (breaking changes)
pnpm version major  # 1.0.0 -> 2.0.0
```

Or directly edit the `version` field in `package.json`.

### Publishing to Public npm

Execute in monorepo root directory:

```bash
# Method 1: Use -C to specify directory
pnpm -C packages/web/workflow-vue publish --access public

# Method 2: Use filter (recommended, more suitable for monorepo)
pnpm -r --filter workflow-vue publish --access public

# Publish beta version
pnpm -r --filter workflow-vue publish --tag beta

# Publish with specific tag
pnpm -r --filter workflow-vue publish --tag next
```

> **Note:** The `prepublishOnly` script is configured in `package.json` and will automatically run `pnpm build` before publishing.

### Publishing to Private Registry

If you need to publish to a private npm registry:

```bash
# Method 1: Temporarily specify registry
pnpm -r --filter workflow-vue publish --registry https://your-registry.com

# Method 2: Configure in package.json
{
  "publishConfig": {
    "registry": "https://your-registry.com",
    "access": "restricted"
  }
}
```

### Publishing Scoped Package

If you need to publish as a scoped package (e.g., `@pieteams/workflow-vue`):

1. Modify `name` in `package.json`:
```json
{
  "name": "@pieteams/workflow-vue"
}
```

2. Specify access when publishing:
```bash
pnpm publish --access public
```

### When to Consider Separate Repository

Publishing directly from a monorepo subdirectory is the most convenient and maintainable approach. Only consider splitting into a separate repository if:

- ❌ Need completely independent permission/visibility control
- ❌ Need completely independent release process and version management
- ❌ Don't want consumers to get monorepo-related metadata
- ❌ Need independent CI/CD pipeline

Otherwise, keeping it in the monorepo has more advantages:
- ✅ Unified dependency management
- ✅ Easier code sharing
- ✅ Controllable refactoring impact
- ✅ Better local development experience

## 🔗 Dependencies

### Peer Dependencies (Required in consuming projects)

These dependencies need to be installed in projects using `workflow-vue`:

- **Vue** `^3.4.0` - Vue 3 framework
- **Bootstrap** `^5.3.0` - UI styling framework
- **Bootstrap Icons** `^1.11.0` - Icon library

Installation command:
```bash
pnpm add vue@^3.4.0 bootstrap@^5.3.0 bootstrap-icons@^1.11.0
```

### Runtime Dependencies (Auto-installed)

These dependencies are automatically installed when installing `workflow-vue`:

- `uuid` `^9.0.1` - UUID generation utility
- `vue-select` `4.0.0-beta.6` - Dropdown select component

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Commit your changes** (`git commit -m 'feat: add amazing feature'`)
4. **Push to the branch** (`git push origin feature/amazing-feature`)
5. **Create a Pull Request**

### Commit Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation updates
- `style:` - Code formatting (no functional changes)
- `refactor:` - Code refactoring (neither new feature nor bug fix)
- `perf:` - Performance optimization
- `test:` - Test-related changes
- `chore:` - Build/toolchain updates

Examples:
```bash
git commit -m "feat(workflow): add workflow export functionality"
git commit -m "fix(node): fix memory leak when deleting nodes"
git commit -m "docs: update API documentation"
```

### Code Standards

- Use ESLint for code linting
- Use Prettier for code formatting
- Write unit tests for new features
- Update relevant documentation

## 📄 License

[MIT](./LICENSE) © PieTeams

## 🔗 Related Links

- [GitHub Repository](https://github.com/pieteams/piemdm)
- [Issue Tracker](https://github.com/pieteams/piemdm/issues)
- [PieMDM Documentation](https://github.com/pieteams/piemdm/tree/main/docs)
- [Changelog](https://github.com/pieteams/piemdm/releases)

## 💬 Support

For questions or suggestions:

- 📝 Submit an [Issue](https://github.com/pieteams/piemdm/issues)
- 📖 Check the [Documentation](https://github.com/pieteams/piemdm/tree/main/packages/web/workflow-vue)
- 💬 Contact the maintenance team

## 🙏 Acknowledgments

Thanks to all developers who have contributed to this project!

---

**Made with ❤️ by PieTeams**
