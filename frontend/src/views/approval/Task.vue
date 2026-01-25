<template>
  <div class="mt-3">
    <div class="border-bottom fs-5 mb-2 pb-2 d-flex justify-content-between align-items-center">
      <span>{{ $t('Approval Task') }}</span>

      <div>
        <button
          type="button"
          class="btn btn-outline-secondary btn-sm me-2"
          @click="goBack"
        >
          <i class="fas fa-arrow-left me-1"></i>
          {{ $t('Back') }}
        </button>
      </div>
    </div>

    <div class="card-body overlay-wrapper p-1">
      <div id="create_wrapper">
        <div
          class="tab-content"
          id="myTabContent"
        >
          <div
            class="tab-pane fade show active"
            id="baseinfo"
          >
            <div id="create_wrapper container">
              <div class="row">
                <div class="col-sm-6">
                  <div class="form-group row">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-3"
                    >
                      {{ $t('Title') }}:
                    </legend>
                    <div class="col-sm-9 my-auto">
                      {{ dataInfo.Title }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-6">
                  <div class="form-group row">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-3"
                    >
                      {{ $t('Status') }}:
                    </legend>
                    <div class="col-sm-9 my-auto">
                      <span
                        class="badge"
                        :class="taskStatusClass"
                      >
                        {{ dataInfo.Status }}
                      </span>
                    </div>
                  </div>
                </div>
                <div class="col-sm-6">
                  <div class="form-group row">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-3"
                    >
                      {{ $t('Code') }}:
                    </legend>
                    <div class="col-sm-9 my-auto">
                      {{ dataInfo.Code }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-6">
                  <div class="form-group row">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-3"
                    >
                      {{ $t('Current WorkItem') }}:
                    </legend>
                    <div class="col-sm-9 my-auto">
                      {{ dataInfo.NodeName }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-6">
                  <div class="form-group row">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-3"
                    >
                      {{ $t('Priority') }}:
                    </legend>
                    <div class="col-sm-9 my-auto">within 2 days</div>
                  </div>
                </div>
                <div class="col-sm-6">
                  <div class="form-group row">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-3"
                    >
                      {{ $t('Created At/By') }}:
                    </legend>
                    <div class="col-sm-9 my-auto">
                      {{ formatDate(dataInfo.CreatedAt) }}
                      {{ dataInfo.CreatedBy ? ' / ' + dataInfo.CreatedBy : '' }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-6">
                  <div class="form-group row">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-3"
                    >
                      {{ $t('Due Date') }}:
                    </legend>
                    <div class="col-sm-9 my-auto">
                      {{ formatDateDistance(dataInfo.CreatedAt, getDateFnsLocale(locale)) }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-6">
                  <div class="form-group row">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-3"
                    >
                      {{ $t('Updated At/By') }}:
                    </legend>
                    <div class="col-sm-9 my-auto">
                      {{ formatDate(dataInfo.UpdatedAt) }}
                      {{ dataInfo.UpdatedBy ? ' / ' + dataInfo.UpdatedBy : '' }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-6">
                  <div class="form-group row">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-3"
                    >
                      {{ $t('Description') }}:
                    </legend>
                    <div class="col-sm-9 my-auto">
                      {{ dataInfo.Desc }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-6">
                  <div class="form-group row">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-3"
                    >
                      {{ $t('Entity') }}:
                    </legend>
                    <div class="col-sm-9 my-auto">
                      {{ dataInfo.EntityCode }}
                    </div>
                  </div>
                </div>
                  <div class="col-sm-12">
                    <div class="row" v-if="platform === 'Builtin'">
                      <div class="col-sm-3">
                        <input
                          type="text"
                          class="form-control form-control-sm"
                          v-model="comment"
                          :placeholder="$t('Comment')"
                        />
                      </div>
                      <div class="col-sm-9">
                        <button
                          type="button"
                          class="btn btn-sm btn-outline-primary me-1"
                          @click="confirmApproval()"
                        >
                          {{ $t('Approve') }}
                        </button>
                        <button
                          type="button"
                          class="btn btn-sm btn-outline-secondary me-1"
                          @click="cancelApproval()"
                        >
                          {{ $t('Reject') }}
                        </button>
                      </div>
                    </div>
                    <div class="alert alert-warning mt-2" role="alert" v-else>
                      <i class="bi bi-info-circle me-2"></i>
                      此审批关联外部平台 ({{ platform }})，请前往对应平台进行审批。
                    </div>
                  </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Approval Workflow -->
    <AppWorkflow :approvalTasks="approvalTasks"></AppWorkflow>

    <!-- data table -->
    <div
      class="table-responsive text-nowrap p-1"
      style="min-height: 41.5vh"
    >
      <table class="table table-sm table-bordered table-hover w-auto mb-0">
        <thead class="thead-light">
          <tr>
            <th v-for="field in tableFields" :key="field.code">
              {{ field.name }}
            </th>
          </tr>
        </thead>
        <tbody id="tabletext">
          <tr v-for="item in tableData" :key="item.id">
            <td v-for="field in tableFields" :key="field.code">
              <template v-if="field.code === 'created_at' || field.code === 'updated_at' || field.code === 'deleted_at'">
                {{ formatDate(item[field.code]) }}
              </template>
              <template v-else-if="field.code === 'status'">
                <StatusBadge :status="item[field.code]" />
              </template>
              <template v-else>
                {{ item[`${field.code}_display`] !== undefined ? item[`${field.code}_display`] : item[field.code] }}
              </template>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <!-- pagination -->
    <AppPagination
      :page="page"
      :page-size="pageSize"
      :total="total"
      @page-change="pageChange"
    />
  </div>
</template>

<script setup>
  import { approveTask, findApproval, rejectTask } from '@/api/approval';
  import { getApprovalDefByCode } from '@/api/approval_def'; // Import added
  import { getApprovalNodeList } from '@/api/approval_node';
  import { getApprovalTaskList } from '@/api/approval_task';
  import { getEntityList } from '@/api/entity';
  import { findTableFieldList, getTableFields as getTableFieldsAPI } from '@/api/table_field';
  import { AppModal } from '@/components/Modal/modal';
  import AppPagination from '@/components/Pagination.vue';
  import { AppToast } from '@/components/toast.js';
  import AppWorkflow from '@/components/Workflow.vue';
  import StatusBadge from '@/components/StatusBadge.vue';
  import { formatDate, formatDateDistance, getDateFnsLocale } from '@/utils/language.js';
  import { formatFieldValue, preloadFieldDictionaries } from '@/utils/fieldFormatter';
  import httpLinkHeader from 'http-link-header';
  import { computed, onMounted, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';

  const router = useRouter();
  const { locale } = useI18n();
  const page = ref(1);
  const pageSize = ref(15);
  const total = ref(0);
  const tableData = ref([]);
  const dataInfo = ref({});
  const tableFields = ref([]);
  const params = ref({});
  const displayNumer = ref(0);
  const comment = ref('test111');
  const loading = ref(false);
  const showApprovalModal = ref(false);
  const approvalAction = ref(''); // 'APPROVE' or 'REJECT'
  const approvalNodes = ref([]);
  const approvalTasks = ref([]);
  const platform = ref('Builtin'); // Default to Builtin

  // Calculate task status style
  const taskStatusClass = computed(() => {
    const status = dataInfo.value.Status;
    switch (status) {
      case 'Pending':
        return 'text-bg-warning';
      case 'Approved':
        return 'text-bg-success';
      case 'Rejected':
        return 'text-bg-danger';
      case 'Cancelled':
        return 'text-bg-secondary';
      default:
        return 'text-bg-primary';
    }
  });

  onMounted(() => {
    params.value = router.currentRoute.value.query;
    getApprovalInfo();
  });

  // get approval info
  const getApprovalInfo = async () => {
    try {
      loading.value = true;
      const res = await findApproval({
        id: params.value.id,
      });
      if (res) {
        console.log('res.data: ', res.data);
        dataInfo.value = res.data;
        console.log('dataInfo.value: ', dataInfo.value);
        getInstanceData(res.data);
        getTableFields(res.data.EntityCode);
        getApprovalNodes(res.data.ApprovalDefCode);
        getApprovalTasks(res.data.Code);

        // Fetch platform info
        if (res.data.ApprovalDefCode) {
          const defRes = await getApprovalDefByCode(res.data.ApprovalDefCode);
          if (defRes.data) {
            platform.value = defRes.data.Platform || 'Builtin';
          }
        }
      }
    } catch (error) {
      AppToast.show({
        message: 'Failed to get approval information',
        color: 'danger',
      });
    } finally {
      loading.value = false;
    }
  };

  // get instance data
  const getInstanceData = async instance => {
    console.log('instance: ', instance);
    const res = await getEntityList({
      'table_code': instance.EntityCode,
      'is_draft': 'Yes',
      'entity_id': instance.EntityID,
      'approval_code': instance.Code,
    });
    if (res && res.data) {
      // Preload dictionaries
      await preloadFieldDictionaries(tableFields.value);

      // Process display values
      tableData.value = await Promise.all(
        res.data.map(async (item) => {
          const processedItem = { ...item };
          for (const field of tableFields.value) {
            const val = item[field.code];
            processedItem[`${field.code}_display`] = await formatFieldValue(val, field);
          }
          return processedItem;
        })
      );

      const links = httpLinkHeader.parse(res.headers.link).refs;
      links.forEach(link => {
        if (['last'].includes(link.rel)) {
          const url = new URL(link.uri);
          total.value = parseInt(url.searchParams.get('page')) || 1;
        }
      });
    }
  };

  // get table fields
  const getTableFields = async tableCode => {
    try {
      // Use _draft suffix to get all draft related fields including system fields
      const res = await getTableFieldsAPI({
        table_code: tableCode + '_draft',
      });
      if (res && res.data) {
        // Map to uniform structure, including lowercase code for data access
        tableFields.value = res.data.map(f => ({
          code: f.code,
          name: f.name,
          Name: f.name, // Keep Name for compatibility
          Code: f.code, // Keep Code for compatibility
          is_system: f.is_system,
          field_type: f.field_type,
          options: f.options,
        }));
      }
    } catch (error) {
      console.error('Failed to fetch field list:', error);
      // Fallback
      const res = await findTableFieldList({
        table_code: tableCode,
      });
      if (res) {
        tableFields.value = res.data;
      }
    }
  };

  const getApprovalNodes = async approvalDefCode => {
    const res = await getApprovalNodeList({
      approval_def_code: approvalDefCode,
    });
    if (res) {
      approvalNodes.value = res.data;
    }
  };

  const getApprovalTasks = async approvalCode => {
    const res = await getApprovalTaskList({
      approval_code: approvalCode,
    });
    if (res) {
      approvalTasks.value = res.data;
    }
  };
  // Confirm approval operation
  const confirmApproval = async () => {
    const result = await AppModal.confirm({
      title: 'Confirm Operation',
      bodyContent: 'Are you sure you want to approve?',
    });

    if (!result) {
      return;
    }

    // Agree, comment can be empty, default to agree
    if (!comment.value.trim()) {
      comment.value = 'Agree';
    }

    // try {
    loading.value = true;
    let res;
    res = await approveTask(dataInfo.value.CurrentTaskID, {
      'comment': comment.value,
      'action': 'APPROVE',
    });

    if (res) {
      AppToast.show({
        message: 'Approval successful',
        color: 'success',
      });
      showApprovalModal.value = false;
      comment.value = '';
      // Refresh page data
      await getApprovalInfo();
    } else {
      AppToast.show({
        message: 'Operation failed, please try again',
        color: 'danger',
      });
    }
  };

  const cancelApproval = async () => {
    const result = await AppModal.confirm({
      title: 'Confirm Operation',
      bodyContent: 'Are you sure you want to reject?',
    });

    if (!result) {
      return;
    }

    // Reject, comment cannot be empty
    if (!comment.value.trim()) {
      AppToast.show({
        message: 'Please fill in the comment',
        color: 'warning',
      });
      return;
    }

    try {
      loading.value = true;
      let res;

      res = await rejectTask(dataInfo.value.CurrentTaskID, {
        'comment': comment.value,
        'action': 'REJECT',
      });

      if (res) {
        AppToast.show({
          message: 'Approval rejected successfully',
          color: 'success',
        });
        showApprovalModal.value = false;
        comment.value = '';
        // Refresh page data
        await getApprovalInfo();
      }
    } catch (error) {
      AppToast.show({
        message: 'Operation failed, please try again',
        color: 'danger',
      });
    } finally {
      loading.value = false;
    }
  };

  const pageChange = p => {
    page.value = p;
    getApprovalInfo();
  };

  // goto index page
  function goIndex() {
    router.push('/approval/index');
  }

  // Go back to approval list
  function goBack() {
    router.go(-1);
  }
</script>

<style scoped>
  .timeline {
    position: relative;
    padding-left: 30px;
  }

  .timeline::before {
    content: '';
    position: absolute;
    left: 15px;
    top: 0;
    bottom: 0;
    width: 2px;
    background: #dee2e6;
  }

  .timeline-item {
    position: relative;
    margin-bottom: 20px;
  }

  .timeline-marker {
    position: absolute;
    left: -23px;
    top: 5px;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    border: 3px solid #fff;
    box-shadow: 0 0 0 2px #dee2e6;
  }

  .timeline-content {
    background: #f8f9fa;
    padding: 15px;
    border-radius: 8px;
    border-left: 3px solid #007bff;
  }
</style>
