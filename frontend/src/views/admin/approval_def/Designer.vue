<template>
  <div class="workflow-designer-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="px-md-4">
        <div class="row align-items-center py-3">
          <div class="col">
            <!-- 面包屑导航 -->
            <nav aria-label="breadcrumb">
              <ol class="breadcrumb mb-0">
                <li class="breadcrumb-item">
                  <router-link
                    to="/admin/approval_def"
                    class="text-decoration-none"
                  >
                    <i class="fas fa-sitemap me-1"></i>
                    审批定义
                  </router-link>
                </li>
                <li
                  class="breadcrumb-item active"
                  aria-current="page"
                >
                  <strong>工作流设计器</strong>
                </li>
              </ol>
            </nav>
          </div>
          <div class="col-auto">
            <button
              type="button"
              class="btn btn-outline-secondary btn-sm me-2"
              @click="goBack"
            >
              <i class="fas fa-arrow-left me-1"></i>
              返回
            </button>
            <button
              type="button"
              class="btn btn-outline-primary btn-sm"
              @click="saveWorkflow"
            >
              <i class="fas fa-save me-1"></i>
              保存工作流
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 工作流设计器 -->
    <!--
            :initial-data="workflowData"
        @save="handleSave"
        @node-select="handleNodeSelect"
        @workflow-change="handleWorkflowChange"
     -->
    <div class="designer-container">
      <WorkflowBuilder
        ref="workflowBuilder"
        v-model="workflowData"
        @requestUserList="handleUserListRequest"
      />
    </div>

    <!-- 消息提示 -->
    <div
      v-if="messageVisible"
      class="toast-container position-fixed top-0 end-0 p-3"
      style="z-index: 9999"
    >
      <div
        class="toast show"
        :class="{
          'text-bg-success': messageType === 'success',
          'text-bg-danger': messageType === 'error',
          'text-bg-warning': messageType === 'warning',
          'text-bg-info': messageType === 'info',
        }"
      >
        <div class="toast-header">
          <i
            class="me-2"
            :class="{
              'fas fa-check-circle text-success': messageType === 'success',
              'fas fa-exclamation-circle text-danger': messageType === 'error',
              'fas fa-exclamation-triangle text-warning': messageType === 'warning',
              'fas fa-info-circle text-info': messageType === 'info',
            }"
          ></i>
          <strong class="me-auto">
            {{
              messageType === 'success'
                ? '成功'
                : messageType === 'error'
                  ? '错误'
                  : messageType === 'warning'
                    ? '警告'
                    : '提示'
            }}
          </strong>
          <button
            type="button"
            class="btn-close"
            @click="messageVisible = false"
          ></button>
        </div>
        <div class="toast-body">
          {{ messageText }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { getApprovalDef } from '@/api/approval_def'; // Changed from getApprovalDefList
  import { batchSyncApprovalNodes, getApprovalNodeList } from '@/api/approval_node';
  import { getUserList } from '@/api/user';
  import { onMounted, ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router'; // added useRoute

  // 路由相关
  const router = useRouter();
  const route = useRoute(); // Use route for params
  const workflowData = ref({});
  const messageVisible = ref(false);
  const messageText = ref('');
  const messageType = ref('info');
  const workflowBuilder = ref(null);

  // 生命周期
  onMounted(async () => {
    await loadWorkflowData();
  });

  // 方法定义
  const loadWorkflowData = async () => {
    const params = route.params; // Get ID from path params
    const query = route.query;   // Get Code from query params

    try {
      const approvalDefId = params.id;
      // 1. 检查审批定义详情，确认 Platform
      if (approvalDefId) {
        const defRes = await getApprovalDef({ id: approvalDefId });
        if (defRes.data) {
          const def = defRes.data;
          if (def.Platform && def.Platform !== 'Builtin') {
            showMessage(`当前审批流由外部平台 (${def.Platform}) 管理，无法在此设计`, 'warning');
            setTimeout(() => {
              goBack();
            }, 1500);
            return;
          }
        }
      }

      const approvalDefCode = query.code;
      if (approvalDefCode && approvalDefCode !== 'new') {
        // 2. 加载现有工作流数据
        const response = await getApprovalNodeList({ approvalDefCode: approvalDefCode });
        await workflowBuilder.value.loadFromBackendData(response.data);
        // workflowData.value = nodes
      }
    } catch (error) {
      console.error(error);
      showMessage('加载工作流数据失败', 'error');
    }
  };

  const goBack = () => {
    router.push('/admin/approval_def');
  };

  const saveWorkflow = async () => {
    const params = router.currentRoute.value.query;
    const data = Array();
    try {
      const nodes = workflowBuilder.value.getBackendData(params.code);
      data.approvalDefCode = params.code;
      data.nodes = nodes;
      console.log('workflowBuilder.value.getBackendData======: ', data);
      await batchSyncApprovalNodes(data);
      showMessage('工作流保存成功', 'success');
    } catch (error) {
      console.error('保存工作流失败:', error);
      showMessage('保存工作流失败', 'error');
    }
  };

  const handleSave = data => {
    workflowData.value = { ...workflowData.value, ...data };
    saveWorkflow();
  };

  const handleNodeSelect = node => {};

  const handleWorkflowChange = data => {
    workflowData.value = { ...workflowData.value, ...data };
  };

  // 处理用户数据请求
  const handleUserListRequest = async (params, callback) => {
    if (params.username.length > 0) {
      params.username = 'like ' + params.username;
    }
    const users = await getUserList(params);
    // 通过回调返回数据
    callback(users.data);
  };

  // 消息提示函数
  const showMessage = (text, type = 'info') => {
    messageText.value = text;
    messageType.value = type;
    messageVisible.value = true;

    setTimeout(() => {
      messageVisible.value = false;
    }, 3000);
  };
</script>

<style scoped>
  .workflow-designer-page {
    height: 100vh;
    display: flex;
    flex-direction: column;
    background-color: #f8f9fa;
  }

  .page-header {
    background: white;
    border-bottom: 1px solid #dee2e6;
    flex-shrink: 0;
  }

  .designer-container {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
  }

  /* Toast 消息样式 */
  .toast {
    min-width: 300px;
    box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
  }

  .toast.show {
    opacity: 1;
  }

  .toast-header {
    border-bottom: 1px solid rgba(0, 0, 0, 0.125);
  }

  .toast-body {
    padding: 0.75rem;
  }
</style>
