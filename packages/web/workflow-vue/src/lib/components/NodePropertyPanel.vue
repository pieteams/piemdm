<template>
  <div
    class="node-property-panel"
    :class="{ show: visible }"
  >
    <div class="panel-header">
      <h5 class="panel-title">
        节点配置
      </h5>
      <button
        type="button"
        class="btn-close"
        @click="hidePanel"
      />
    </div>

    <div class="panel-body">
      <div
        v-if="!currentNode"
        class="text-center py-4"
      >
        <div class="text-muted">
          请选择一个节点进行配置
        </div>
      </div>

      <div
        v-else
        class="node-property-form"
      >
        <!-- 通用配置 -->
        <div class="mb-4">
          <div class="mb-3">
            <label class="form-label">
              节点名称
              <span class="text-danger">*</span>
            </label>
            <input
              v-model="formData.NodeName"
              type="text"
              class="form-control form-control-sm"
              placeholder="请输入节点名称"
            >
          </div>

          <!-- 审批节点类型选择 -->
          <div
            v-if="currentNode.NodeType === 'APPROVAL'"
            class="mb-3"
          >
            <label class="form-label">审批类型</label>
            <div class="d-flex flex-wrap gap-2">
              <div class="form-check">
                <input
                  id="approval-type-users"
                  v-model="approverConfig.type"
                  class="form-check-input"
                  type="radio"
                  value="USERS"
                >
                <label
                  class="form-check-label"
                  for="approval-type-users"
                >人工审批</label>
              </div>
              <div class="form-check">
                <input
                  id="approval-type-auto-approve"
                  v-model="approverConfig.type"
                  class="form-check-input"
                  type="radio"
                  value="AUTO_APPROVE"
                >
                <label
                  class="form-check-label"
                  for="approval-type-auto-approve"
                >自动通过</label>
              </div>
              <div class="form-check">
                <input
                  id="approval-type-auto-reject"
                  v-model="approverConfig.type"
                  class="form-check-input"
                  type="radio"
                  value="AUTO_REJECT"
                >
                <label
                  class="form-check-label"
                  for="approval-type-auto-reject"
                >自动拒绝</label>
              </div>
            </div>
          </div>

          <div class="mb-3">
            <label class="form-label">节点描述</label>
            <textarea
              v-model="formData.Description"
              class="form-control form-control-sm"
              rows="3"
              placeholder="请输入节点描述（可选）"
            />
          </div>
        </div>

        <!-- 人工审批配置 -->
        <div
          v-if="currentNode.NodeType === 'APPROVAL' && approverConfig.type === 'USERS'"
          class="mb-4"
        >
          <div class="mb-3">
            <label class="form-label">审批人列表</label>
            <v-select
              v-model="approverConfig.users"
              :options="userOptions"
              :reduce="user => user.Username"
              label="Username"
              multiple
              placeholder="请选择审批人"
              @search="handleUserSearch"
            >
              <template #no-options>
                无匹配用户
              </template>
            </v-select>
          </div>

          <div class="mb-3">
            <label class="form-label">审批模式</label>
            <select
              v-model="approverConfig.mode"
              class="form-select form-select-sm"
            >
              <option value="OR">
                任意一人审批即可
              </option>
              <option value="AND">
                需要所有人审批
              </option>
            </select>
          </div>
        </div>

        <!-- 自动审批提示 -->
        <div
          v-if="currentNode.NodeType === 'APPROVAL' && isAutoApproval"
          class="mb-4"
        >
          <div class="alert alert-info">
            <small>
              <i class="bi bi-info-circle me-1" />
              此节点将自动{{
                approverConfig.type === 'AUTO_APPROVE' ? '通过' : '拒绝'
              }}，无需人工干预。
            </small>
          </div>
        </div>

        <!-- 条件节点配置 -->
        <div
          v-if="currentNode.NodeType === 'CONDITION'"
          class="mb-4"
        >
          <div class="mb-3">
            <label class="form-label">条件分支</label>
            <div
              v-for="(branch, index) in conditionConfig.branches"
              :key="branch.id"
              class="border p-3 mb-3"
            >
              <div class="d-flex justify-content-between align-items-center mb-2">
                <h6 class="mb-0">
                  分支 {{ index + 1 }}
                </h6>
                <button
                  v-if="conditionConfig.branches.length > 1"
                  type="button"
                  class="btn-close"
                  @click="removeBranch(index)"
                />
              </div>
              <div class="row g-2">
                <div class="col-12">
                  <label class="form-label">分支名称</label>
                  <input
                    v-model="branch.name"
                    type="text"
                    class="form-control form-control-sm"
                    placeholder="请输入分支名称"
                  >
                </div>
                <div class="col-12">
                  <label class="form-label">条件类型</label>
                  <select
                    v-model="branch.condition.conditionType"
                    class="form-select form-select-sm"
                  >
                    <option
                      v-for="type in conditionTypes"
                      :key="type.value"
                      :value="type.value"
                    >
                      {{ type.label }}
                    </option>
                  </select>
                </div>
                <div class="col-12">
                  <label class="form-label">字段名</label>
                  <input
                    v-model="branch.condition.fieldName"
                    type="text"
                    class="form-control form-control-sm"
                    placeholder="请输入字段名"
                    @input="onConditionFieldChange(branch.condition)"
                  >
                </div>
                <div class="col-12">
                  <label class="form-label">操作符</label>
                  <select
                    v-model="branch.condition.operator"
                    class="form-select form-control-sm"
                  >
                    <option value="eq">
                      等于
                    </option>
                    <option value="ne">
                      不等于
                    </option>
                    <option value="gt">
                      大于
                    </option>
                    <option value="lt">
                      小于
                    </option>
                    <option value="gte">
                      大于等于
                    </option>
                    <option value="lte">
                      小于等于
                    </option>
                    <option value="contains">
                      包含
                    </option>
                    <option value="not_contains">
                      不包含
                    </option>
                  </select>
                </div>
                <div class="col-12">
                  <label class="form-label">字段值</label>
                  <div v-if="branch.condition.fieldName === 'createdBy'">
                    <div class="input-group input-group-sm">
                      <input
                        v-model="userInput"
                        type="text"
                        class="form-control"
                        placeholder="输入用户名，按回车添加"
                        @keyup.enter="addConditionUser(branch.condition)"
                      >
                      <button
                        class="btn btn-outline-primary"
                        @click="addConditionUser(branch.condition)"
                      >
                        添加
                      </button>
                    </div>
                    <div
                      v-if="
                        branch.condition.fieldValue &&
                          branch.condition.fieldValue.split(',').length > 0
                      "
                      class="mt-2"
                    >
                      <span
                        v-for="(user, index) in branch.condition.fieldValue.split(',')"
                        :key="index"
                        class="badge bg-primary me-2 mb-2"
                      >
                        {{ getUserDisplayName(user) }}
                        <button
                          type="button"
                          class="btn-close btn-close-white ms-1"
                          style="font-size: 10px"
                          @click="removeConditionUser(branch.condition, index)"
                        />
                      </span>
                    </div>
                  </div>
                  <input
                    v-else
                    v-model="branch.condition.fieldValue"
                    type="text"
                    class="form-control form-control-sm"
                    placeholder="请输入字段值"
                  >
                </div>
              </div>
            </div>
            <button
              type="button"
              class="btn btn-outline-primary btn-sm"
              @click="addBranch"
            >
              <i class="bi bi-plus" />
              添加分支
            </button>
          </div>
        </div>

        <!-- 抄送节点配置 -->
        <div
          v-if="currentNode.NodeType === 'CC'"
          class="mb-4"
        >
          <div class="mb-3">
            <label class="form-label">抄送人列表</label>
            <v-select
              v-model="approverConfig.users"
              :options="userOptions"
              :reduce="user => user.Username"
              label="Username"
              multiple
              placeholder="请选择抄送人"
              @search="handleUserSearch"
            >
              <template #no-options>
                无匹配用户
              </template>
            </v-select>
          </div>

          <div class="mb-3">
            <label class="form-label">抄送时机</label>
            <select
              v-model="approverConfig.ccTiming"
              class="form-select form-select-sm"
            >
              <option value="after_approval">
                审批通过后抄送
              </option>
              <option value="before_approval">
                审批前抄送
              </option>
              <option value="on_submit">
                提交时抄送
              </option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <div class="panel-footer">
      <div class="d-flex gap-2">
        <button
          type="button"
          class="btn btn-secondary btn-sm"
          @click="hidePanel"
        >
          取消
        </button>
        <button
          type="button"
          class="btn btn-primary btn-sm"
          @click="saveConfiguration"
        >
          保存
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref, reactive, computed, watch } from 'vue';
  import { JsonHelper } from '../utils/json-helper.js';
  import vSelect from 'vue-select';
  import 'vue-select/dist/vue-select.css';

  const emit = defineEmits(['save', 'close', 'requestUserList']);

  const visible = ref(false);
  const currentNode = ref(null);

  const formData = reactive({
    NodeName: '',
    Description: '',
  });

  const approverConfig = reactive({
    type: 'USERS',
    users: [],
    mode: 'OR',
    ccTiming: 'after_approval',
  });

  const conditionConfig = reactive({
    branches: [],
  });

  const userInput = ref('');

  const handleUserSearch = query => {
    if (query.length > 2) {
      loadUserList(query);
    }
  };

  // 监听用户输入变化，实时搜索用户 (用于条件节点的 createdBy 搜索)
  watch(userInput, newValue => {
    if (newValue.trim().length > 2) {
      loadUserList(newValue);
    } else if (newValue.trim().length === 0) {
      loadUserList();
    }
  });
  const conditionTypes = [
    { value: 'instance', label: '审批实例字段' },
    { value: 'data', label: '数据字段' },
  ];

  const userList = ref([]);

  // 完整的用户选项列表（包含已选用户和搜索结果）
  const userOptions = computed(() => {
    const selected = approverConfig.users || [];
    const existing = userList.value || [];

    // 创建一个 Map 来去重
    const userMap = new Map();

    // 先添加已选用户（作为占位符，如果在 existing 中找不到）
    selected.forEach(username => {
      if (username && username !== 'system') {
        userMap.set(username, { ID: null, Username: username });
      }
    });

    // 用实际的用户对象覆盖占位符
    existing.forEach(user => {
      if (user && user.Username) {
        userMap.set(user.Username, user);
      }
    });

    return Array.from(userMap.values());
  });

  // 加载用户列表（通过 emit 事件从主程序获取数据）
  const loadUserList = async (search = '') => {
    try {
      // 发出请求用户列表事件，主程序需要监听并返回数据
      emit(
        'requestUserList',
        {
          username: search || userInput.value,
          pageSize: 10,
        },
        users => {
          if (Array.isArray(users)) {
            userList.value = users;
          } else if (users && typeof users === 'object') {
            // 尝试从常见的包装字段中获取数组
            if (Array.isArray(users.data)) {
              userList.value = users.data;
            } else if (Array.isArray(users.list)) {
              userList.value = users.list;
            } else if (Array.isArray(users.users)) {
              userList.value = users.users;
            } else {
              console.warn('requestUserList 返回的数据格式不正确，期望数组:', users);
              userList.value = [];
            }
          } else {
            console.warn('requestUserList 返回的数据格式不正确:', users);
            userList.value = [];
          }
        },
      );
    } catch (error) {
      console.error('加载用户列表失败:', error);
      // 如果主程序未处理事件，使用模拟数据作为降级方案
      userList.value = [
        { ID: 6, Username: 'jasen8888' },
        { ID: 7, Username: 'jasen' },
      ];
    }
  };

  // 计算属性，用于控制UI显示
  const isAutoApproval = computed(
    () => approverConfig.type === 'AUTO_APPROVE' || approverConfig.type === 'AUTO_REJECT',
  );

  // 监听审批类型的变化，自动更新节点名称和审批人列表
  watch(
    () => approverConfig.type,
    newType => {
      if (currentNode.value && currentNode.value.NodeType === 'APPROVAL') {
        if (newType === 'AUTO_APPROVE') {
          formData.NodeName = '自动通过';
          approverConfig.users = ['system'];
        } else if (newType === 'AUTO_REJECT') {
          formData.NodeName = '自动拒绝';
          approverConfig.users = ['system'];
        } else if (newType === 'USERS') {
          formData.NodeName = '人工审批';
          approverConfig.users = [];
        }
      }
    },
  );

  // 显示配置面板
  const show = async node => {
    currentNode.value = node;
    visible.value = true;

    // 重置表单数据
    Object.assign(formData, {
      NodeName: node.NodeName || '',
      Description: node.Description || '',
    });

    // 根据节点类型加载配置
    if (node.NodeType === 'APPROVAL') {
      const parsedApproverConfig = JsonHelper.safeParse(node.ApproverConfig);

      Object.assign(approverConfig, {
        type: node.ApproverType || 'USERS',
        users: (parsedApproverConfig.users || []).filter(u => u != null && u !== ''),
        mode: parsedApproverConfig.mode || 'OR',
      });
    } else if (node.NodeType === 'CC') {
      const parsedApproverConfig = JsonHelper.safeParse(node.ApproverConfig);

      Object.assign(approverConfig, {
        type: 'CC',
        users: (parsedApproverConfig.users || []).filter(u => u != null && u !== ''),
        mode: 'CC',
        ccTiming: parsedApproverConfig.ccTiming || 'after_approval',
      });
    } else if (node.NodeType === 'CONDITION') {
      const parsedConditionConfig = JsonHelper.safeParse(node.ConditionConfig);
      conditionConfig.branches = parsedConditionConfig.branches || [];
    }

    await loadUserList();
  };

  // 隐藏面板
  const hidePanel = () => {
    visible.value = false;
    emit('close');
  };

  // 获取节点图标
  const getNodeIcon = node => {
    if (!node) return 'bi-circle-fill';
    if (node.NodeType === 'APPROVAL') {
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
  };

  // 为条件节点添加用户
  const addConditionUser = condition => {
    if (userInput.value.trim()) {
      const currentUsers = condition.fieldValue ? condition.fieldValue.split(',') : [];
      if (!currentUsers.includes(userInput.value.trim())) {
        currentUsers.push(userInput.value.trim());
        condition.fieldValue = currentUsers.join(',');
        userInput.value = '';
      }
    }
  };

  // 从条件节点移除用户
  const removeConditionUser = (condition, index) => {
    const currentUsers = condition.fieldValue ? condition.fieldValue.split(',') : [];
    currentUsers.splice(index, 1);
    condition.fieldValue = currentUsers.join(',');
  };

  // 条件字段变化处理
  const onConditionFieldChange = condition => {
    if (condition.fieldName === 'createdBy') {
      condition.conditionType = 'instance';
      condition.operator = 'eq';
    } else {
      condition.conditionType = 'data';
      condition.operator = 'eq';
    }
  };

  // 获取用户显示名称
  const getUserDisplayName = userId => {
    const user = userList.value.find(u => u.ID === userId);
    return user ? user.Username : userId;
  };

  // 添加分支
  const addBranch = () => {
    conditionConfig.branches.push({
      id: Date.now().toString(),
      name: `分支 ${conditionConfig.branches.length + 1}`,
      condition: {
        conditionType: 'instance',
        fieldName: '',
        operator: 'eq',
        fieldValue: '',
      },
      nodes: [],
    });
  };

  // 移除分支
  const removeBranch = index => {
    conditionConfig.branches.splice(index, 1);
  };

  // 保存配置
  const saveConfiguration = () => {
    if (!formData.NodeName.trim()) {
      alert('请输入节点名称');
      return;
    }

    const updateData = {
      ...formData,
    };

    if (currentNode.value.NodeType === 'APPROVAL') {
      updateData.ApproverType = approverConfig.type;

      let configToSave;
      if (isAutoApproval.value) {
        configToSave = {
          type: approverConfig.type,
          users: ['system'],
          mode: 'OR',
        };
      } else if (approverConfig.type === 'USERS') {
        configToSave = {
          ...approverConfig,
          users: approverConfig.users.filter(u => u != null && u !== ''),
        };
        if (configToSave.users.length === 0) {
          alert('请至少添加一个审批人');
          return;
        }
      } else {
        configToSave = {
          type: approverConfig.type,
          users: ['system'],
          mode: 'OR',
        };
      }

      updateData.ApproverConfig = JSON.stringify(configToSave);
    } else if (currentNode.value.NodeType === 'CC') {
      const configToSave = {
        type: 'CC',
        users: approverConfig.users.filter(u => u != null && u !== ''),
        mode: 'CC',
        ccTiming: approverConfig.ccTiming,
      };
      updateData.ApproverConfig = JSON.stringify(configToSave);
    } else if (currentNode.value.NodeType === 'CONDITION') {
      // 验证条件分支 - 所有分支都必须填写完整
      const invalidBranch = conditionConfig.branches.find(branch => {
        // 检查分支名称
        if (!branch.name.trim()) {
          return true;
        }

        // 所有分支都需要检查字段名、操作符和字段值
        if (
          !branch.condition.fieldName.trim() ||
          !branch.condition.operator ||
          !branch.condition.fieldValue.trim()
        ) {
          return true;
        }

        return false;
      });

      if (invalidBranch) {
        if (!invalidBranch.name.trim()) {
          alert('请填写分支名称');
        } else {
          alert('请填写字段名、操作符和字段值');
        }
        return;
      }

      const processedConfig = {
        branches: conditionConfig.branches.map(branch => ({
          ...branch,
          condition: {
            ...branch.condition,
            operator: branch.condition.operator,
          },
        })),
      };

      updateData.ConditionConfig = JSON.stringify(processedConfig);
    }

    emit('save', currentNode.value, updateData);

    hidePanel();
  };

  defineExpose({
    show,
    hide: hidePanel,
  });
</script>

<style scoped>
  .node-property-panel {
    position: fixed;
    top: 0;
    right: -400px;
    width: 400px;
    height: 100vh;
    background: white;
    border-left: 1px solid #dee2e6;
    box-shadow: -2px 0 10px rgba(0, 0, 0, 0.1);
    transition: right 0.3s ease;
    z-index: 1000;
    display: flex;
    flex-direction: column;
  }

  .node-property-panel.show {
    right: 0;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem;
    border-bottom: 1px solid #dee2e6;
    background: #f8f9fa;
  }

  .panel-title {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
  }

  .panel-body {
    flex: 1;
    padding: 1rem;
    overflow-y: auto;
  }

  .panel-footer {
    padding: 1rem;
    border-top: 1px solid #dee2e6;
    background: #f8f9fa;
  }

  .node-property-form {
    height: 100%;
  }

  .badge {
    display: inline-flex;
    align-items: center;
    font-size: 0.75rem;
  }

  .badge .btn-close {
    font-size: 0.6rem;
    padding: 0.25rem;
  }

  .border {
    border: 1px solid #dee2e6 !important;
  }

  @media (max-width: 768px) {
    .node-property-panel {
      width: 100%;
      right: -100%;
    }
  }
</style>
