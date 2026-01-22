<template>
  <div class="mt-3">
    <div class="fs-5 mb-2 pb-2">{{ $t('Table') }} {{ $t('Update') }}</div>

    <ul class="nav nav-tabs">
      <li class="nav-item">
        <button class="nav-link active" id="base-info-tab" data-bs-toggle="tab" data-bs-target="#base-info-tab-pane"
          type="button">
          {{ $t('Basic Info') }}
        </button>
      </li>

      <li class="nav-item">
        <button class="nav-link" id="approval-flow-tab" data-bs-toggle="tab" data-bs-target="#approval-flow-tab-pane"
          type="button">
          {{ $t('Approval Flow') }}
        </button>
      </li>
    </ul>
    <div class="tab-content" id="myTabContent" style="min-height: calc(100vh - 200px); font-size: 0.9rem">
      <!-- data info -->
      <div class="tab-pane fade show active" id="base-info-tab-pane" tabindex="0">
        <div class="card-body mt-2">
          <div class="row">
            <app-form :data-info="dataInfo" @submit-form="submitForm" @go-index="goIndex" />
          </div>
        </div>
      </div>
      <!-- extension -->

      <!-- approval flow -->
      <div class="tab-pane fade" id="approval-flow-tab-pane" tabindex="0">
        <div class="mt-2">
          <div class="row mb-2">
            <div class="col-sm-12">
              <form class="row g-2">
                <div class="col-md-3">
                  <v-select v-model="operationApprovalDefInfo.operation" :options="operationOptions"
                    :reduce="option => option.value" label="label"
                    :placeholder="$t('Please Select') + ' ' + $t('Operation')">
                    <template #option="{ label }">{{ label }}</template>
                  </v-select>
                </div>
                <div class="col-md-4">
                  <v-select v-model="operationApprovalDefInfo.approval_def_code" :options="approvalDefs"
                    :reduce="option => option.Code" label="Code"
                    :placeholder="$t('Please Select') + ' ' + $t('Approval Flow')" :clearable="true" :searchable="true"
                    @search="loadApprovalDefs">
                    <template #option="{ Code, Name }">{{ Code }} {{ Name }}</template>
                    <template #selected-option="{ Code, Name }">
                      {{ Code }} {{ Name }}
                    </template>
                  </v-select>
                </div>
                <div class="col-md-3">
                  <input type="hidden" name="id" v-model="operationApprovalDefInfo.id" />
                  <button type="button" class="btn btn-primary btn-sm" @click="onSublimtOperationApprovalDef">
                    {{ operationApprovalDefInfo.id > 0 ? $t('Update') : $t('Create') }}
                  </button>
                  <button v-if="operationApprovalDefInfo.id > 0" type="button" class="btn btn-secondary btn-sm ms-1"
                    @click="cancelOperationApprovalDef">
                    {{ $t('Cancel') }}
                  </button>
                </div>
              </form>
            </div>
          </div>
          <!-- approval flow list -->
          <div class="row">
            <div class="col-sm-12">
              <div class="table-responsive">
                <table class="table table-sm table-bordered table-hover">
                  <thead class="table-light">
                    <tr>
                      <th>{{ $t('Entity') }}</th>
                      <th>{{ $t('Operation') }}</th>
                      <th>{{ $t('Approval Flow') }}</th>
                      <th>{{ $t('Description') }}</th>
                      <th>{{ $t('CreatedBy') }}</th>
                      <th>{{ $t('UpdatedBy') }}</th>
                      <th>{{ $t('CreatedAt') }}</th>
                      <th>{{ $t('UpdatedAt') }}</th>
                      <th>{{ $t('Status') }}</th>
                      <th class="actions text-center">{{ $t('Actions') }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-if="approvalFlowList.length === 0">
                      <td colspan="10" class="text-center text-muted py-3">
                        {{ $t('Your data is empty.') }}
                      </td>
                    </tr>
                    <tr v-for="item in approvalFlowList" :key="item.id">
                      <td>
                        {{ item.entity_code }}
                      </td>
                      <td>
                        {{ item.operation }}
                      </td>
                      <td>{{ getApprovalDefName(item.approval_def_code) }}</td>
                      <td>{{ item.description }}</td>
                      <td>{{ item.created_by }}</td>
                      <td>{{ item.updated_by }}</td>
                      <td>{{ formatDate(item.created_at) }}</td>
                      <td>{{ formatDate(item.updated_at) }}</td>
                      <td>
                        <StatusBadge :status="item.status" />
                      </td>
                      <td class="actions text-center">
                        <a href="javascript:void(0)" @click="editOperationApprovalDef(item)" :title="$t('Update')">
                          <i class="bi bi-pencil"></i>
                        </a>
                        <a href="javascript:void(0)" @click="deleteOperationApprovalDef(item.id)" :title="$t('Delete')">
                          <i class="bi bi-trash"></i>
                        </a>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
            <div class="col-sm-auto mb-5">
              <button type="button" class="btn btn-outline-secondary btn-sm" @click="goIndex">
                {{ $t('Cancel') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Extension Update Modal -->

</template>

<script setup>
import { getApprovalDefList } from '@/api/approval_def';
import { findTableList, getTable, updateTable } from '@/api/table';
import {
  createTableApprovalDef,
  deleteTableApprovalDef,
  findTableApprovalDefList,
  updateTableApprovalDef,
} from '@/api/table_approval_def';


import { AppModal } from '@/components/Modal/modal.js';
import StatusBadge from '@/components/StatusBadge.vue';
import { AppToast } from '@/components/toast.js';
import { useFormOptions } from '@/composables/useFormOptions';
import { formatDate } from '@/utils/language.js';

import { onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import vSelect from 'vue-select';
import 'vue-select/dist/vue-select.css';
import AppForm from './Form.vue';

const router = useRouter();
const { t } = useI18n();
const dataInfo = ref({});


// 审批流程管理
const approvalFlowList = ref([]);
const operationApprovalDefInfo = ref({});
const approvalDefs = ref([]);
const { operationOptions } = useFormOptions();

onMounted(async () => {
  await getDataInfo();
  await loadApprovalDefs('', null);
  await loadOpreationApprovalDefList();
});

const getApprovalDefs = async () => {
  const res = await getApprovalDefList({ pageSize: 1000, status: 'Enabled' });
  if (res && res.data) {
    approvalDefs.value = res.data.list;
  }
};

const onSublimtOperationApprovalDef = async () => {
  operationApprovalDefInfo.value.entity_code = dataInfo.value.Code;
  operationApprovalDefInfo.value.status = 'Normal';
  operationApprovalDefInfo.value.description = `${dataInfo.value.Code} - ${operationApprovalDefInfo.value.operation
    } ${t('Operation')} ${t('Approval')}`;

  if (operationApprovalDefInfo.value.id) {
    await updateTableApprovalDef(operationApprovalDefInfo.value);
    AppToast.show({ message: t('Approval flow updated successfully'), color: 'success' });
  } else {
    await createTableApprovalDef(operationApprovalDefInfo.value);
    AppToast.show({ message: t('Approval flow added successfully'), color: 'success' });
  }
  operationApprovalDefInfo.value = {};
  await loadOpreationApprovalDefList();
};

const editOperationApprovalDef = item => {
  operationApprovalDefInfo.value = item;
};

const cancelOperationApprovalDef = () => {
  operationApprovalDefInfo.value = {};
};

const deleteOperationApprovalDef = async id => {
  const confirmed = await AppModal.confirm({
    title: t('Confirm') + ' ' + t('Delete'),
    content: t('Are you sure to delete this approval flow configuration') + '?',
    okTitle: t('Delete'),
    cancelTitle: t('Cancel'),
  });

  if (!confirmed) {
    return;
  }

  await deleteTableApprovalDef(id);
  AppToast.show({ message: t('Delete successful'), color: 'success' });
  await loadOpreationApprovalDefList();
};

const loadOpreationApprovalDefList = async () => {
  const entityCode = dataInfo.value.Code || dataInfo.value.table_code;
  if (!entityCode) {
    return;
  }

  const res = await findTableApprovalDefList({
    entity_code: entityCode,
  });
  approvalFlowList.value = res.data || [];
};

const loadApprovalDefs = async (searchText, loading) => {
  try {
    // 如果有 loading 函数，则调用它
    if (loading && typeof loading === 'function') {
      loading(true);
    }

    const response = await getApprovalDefList({
      pageSize: 10,
      status: 'Normal',
    });

    approvalDefs.value = response.data || [];

    // 如果有 loading 函数，结束加载状态
    if (loading && typeof loading === 'function') {
      loading(false);
    }
  } catch (error) {
    console.error(t('Load approval definition failed') + ':', error);
    approvalDefs.value = [];

    // 如果有 loading 函数，结束加载状态
    if (loading && typeof loading === 'function') {
      loading(false);
    }
  }
};

const getApprovalDefName = defCode => {
  const def = approvalDefs.value.find(d => d.def_code === defCode);
  return def ? def.def_name : defCode;
};

const getDataInfo = async () => {
  const params = router.currentRoute.value.query;
  const res = await getTable({
    id: params.id,
  });
  dataInfo.value = res.data;
};

async function submitForm(data) {
  const res = await updateTable({
    ...data,
  });
  if (res.data) {
    AppToast.show({
      message: 'Update Table success。',
      color: 'success',
    });
    // router.push('/admin/table/index');
  }
}

function goIndex() {
  router.push('/admin/table/index');
}

// 监听表信息变化，自动加载审批流程
watch(
  () => dataInfo.value.Code,
  newTableCode => {
    if (newTableCode) {
      loadOpreationApprovalDefList();
      operationApprovalDefInfo.value = {};
    }
  }
);
</script>

<style scoped></style>
