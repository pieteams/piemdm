<template>
  <div class="mt-3">
    <!-- search criteria -->
    <div class="border-bottom fs-5 mb-2 pb-2">{{ $t('Approval Define') }} {{ $t('List') }}</div>

    <AppSearch @search="onSearch" />

    <AppResult :page="page" :page-size="pageSize" :total="total"></AppResult>

    <!-- operation list-->
    <div class="form-group row py-2 px-1">
      <div class="col-sm-12">
        <a class="btn btn-outline-primary btn-sm me-1" href="/admin/approval_def/create" role="button">
          <i class="bi bi-file-earmark-plus"></i>
          {{ $t('New') }}
        </a>
        <button type="button" class="btn btn-outline-primary btn-sm me-1" @click="changeStatus('Frozen')">
          {{ $t('Freeze') }}
        </button>
        <button type="button" class="btn btn-outline-primary btn-sm me-1" @click="changeStatus('Normal')">
          {{ $t('Unfreeze') }}
        </button>
        <button type="button" class="btn btn-outline-danger btn-sm me-1" @click="confirmDelete">
          <i class="bi bi-trash3"></i>
          {{ selected.length ? '(' + selected.length + ')' : '' }}
        </button>

        <!-- <button type="button" class="btn btn-outline-primary btn-sm me-1" @click="massOperate('Copy')">
        <i class="bi bi-files"></i>
        {{ $t('Copy') }}
      </button> -->
      </div>
    </div>

    <!-- data table -->
    <!-- margin-top: 115px;min-height: 658px; -->
    <div class="table-responsive text-nowrap p-1"
      style="min-height: calc(100vh - 310px); overflow-y: auto; font-size: 0.9rem" v-if="total">
      <table class="table table-sm table-bordered table-striped table-hover w-100 mb-0 sticky-table">
        <thead class="table-light">
          <tr>
            <th class="text-center align-middle sticky-col sticky-col-checkbox">
              <input type="checkbox" @click="selectAll" v-model="checked" />
            </th>
            <th class="text-center">{{ $t('ID') }}</th>
            <th>{{ $t('Code') }}</th>
            <th>{{ $t('Name') }}</th>
            <th>{{ $t('Description') }}</th>
            <th>{{ $t('FormData') }}</th>
            <th>{{ $t('NodeList') }}</th>
            <th>{{ $t('Platform') }}</th>
            <th>{{ $t('Status') }}</th>
            <th>{{ $t('UpdatedAt') }}</th>
            <th>{{ $t('CreatedAt') }}</th>
            <th class="sticky-col sticky-col-actions text-center">{{ $t('Actions') }}</th>
          </tr>
        </thead>
        <tbody id="tabletext">
          <tr v-for="item in tableData">
            <td class="text-center align-middle sticky-col sticky-col-checkbox">
              <input type="checkbox" v-model="selected" :value="item.ID" number />
            </td>
            <td class="text-center">
              {{ item.ID }}
            </td>
            <td>
              <router-link :to="'/admin/approval_def/view?id=' + item.ID">
                {{ item.Code }}
              </router-link>
            </td>
            <td>{{ item.Name }}</td>
            <td>{{ item.Description }}</td>
            <td>{{ item.FormData }}</td>
            <td>{{ item.NodeList }}</td>
            <td>{{ item.Platform }}</td>
            <td>
              <StatusBadge :status="item.Status" />
            </td>
            <td>{{ formatDate(item.UpdatedAt) }}</td>
            <td>{{ formatDate(item.CreatedAt) }}</td>
            <td class="sticky-col sticky-col-actions text-center">
              <a v-if="!item.Platform || item.Platform === 'Builtin'" :href="'/admin/approval_def/designer/' + item.ID + '?code=' + item.Code" title="流程设计">
                <i class="bi bi-diagram-3 text-primary"></i>
              </a>
              <span v-else class="text-muted p-2" title="外部流程无法设计">
                <i class="bi bi-diagram-3"></i>
              </span>
              <a :href="'/admin/approval_def/update?id=' + item.ID" title="编辑">
                <i class="bi bi-pencil"></i>
              </a>
              <a :href="'/admin/approval_def/view?id=' + item.ID" title="查看">
                <i class="bi bi-file-text"></i>
              </a>
              <a @click="testApprovalDef(item)" href="javascript:void(0)" title="测试"
                :disabled="item.Status !== 'Normal'">
                <i class="bi bi-flask"></i>
              </a>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div class="table-responsive text-nowrap p-1" style="min-height: 75vh" v-else>
      {{ $t('Your approval definition is empty.') }}
    </div>
    <app-pagination :page="page" :page-size="pageSize" :total="total" @page-change="pageChange" />

    <!-- <Teleport to="body">
      <AppModal :show="showModal" @close="showModal = false" />
    </Teleport> -->
  </div>
</template>

<script setup>
import {
  batchDeleteApprovalDef,
  createApprovalDef,
  getApprovalDefList,
  updateApprovalDefStatus,
} from '@/api/approval_def';
import { AppModal } from '@/components/Modal/modal.js';
import AppPagination from '@/components/Pagination.vue';
import AppResult from '@/components/Result.vue';
import StatusBadge from '@/components/StatusBadge.vue';
import { AppToast } from '@/components/toast.js';
import { formatDate } from '@/utils/language.js';
import httpLinkHeader from 'http-link-header';
import { onMounted, ref } from 'vue';
import AppSearch from './Search.vue';

const page = ref(1);
const pageSize = ref(15);
const total = ref(0);
const tableData = ref([]);
const selected = ref([]);
const checked = ref(false);
const formData = ref({});

onMounted(() => {
  getApprovalDefData();
});

// 条件搜索前端看此方法
const onSearch = searchDate => {
  formData.value = searchDate;
  getApprovalDefData();
};

// 查询
const getApprovalDefData = async () => {
  const res = await getApprovalDefList({
    page: page.value,
    pageSize: pageSize.value,
    ...formData.value,
  });
  if (res) {
    tableData.value = res.data;
    const links = httpLinkHeader.parse(res.headers.link).refs;
    links.forEach(link => {
      if (['last'].includes(link.rel)) {
        const url = new URL(link.uri);
        total.value = parseInt(url.searchParams.get('page')) || 1;
      }
    });
  } else {
    AppToast.show({
      message: 'Get approval definition list failed',
      color: 'danger',
    });
  }
};

// 获取审批流任务
// TODO 如果code为空，增加提示
// function getProcessInfo(code) {
//   showModal.value = true
//   // const spinner = document.querySelector('.spinner')
//   // spinner.style.display = "block"

//   // if (code === "") return;
//   // let url = '/admin/approval_def/task?resultType=json';
//   // url = url + "&code=" + code;

//   // const statusToast = document.getElementById('statusToast')
//   // const toastBody = statusToast.querySelector('.toast-body')
//   // var toast = new bootstrap.Toast(statusToast)

//   // const response = fetch(url);
//   // if (response.ok) {
//   //   const res = response.json()
//   //   if (res.code == 200) {
//   //     this.tasks = res.data.data
//   //     const tm = new bootstrap.Modal(document.getElementById('taskModal'))
//   //     tm.show()
//   //   } else {
//   //     toast.show()
//   //   }
//   // } else {
//   //   toastBody.innerHTML = response.statusText
//   //   toast.show()
//   // }
//   // spinner.style.display = "none"
// }

const selectAll = () => {
  var ids = [];
  if (!checked.value) {
    tableData.value.forEach(function (val) {
      ids.push(val.ID);
    });
    selected.value = ids;
  } else {
    selected.value = [];
  }
};

// const toDetail = (row) => {
//   router.push({
//     name: 'dictionaryDetail',
//     params: {
//       id: row.ID,
//     },
//   })
// }

// const dialogFormVisible = ref(false)
// const type = ref('')
// const updateApprovalDefFunc = async (row) => {
//   const res = await findApprovalDef({ ID: row.ID, status: row.status })
//   type.value = 'update'
//   if (res.code === 0) {
//     formData.value = res.data.resysDictionary
//     dialogFormVisible.value = true
//   }
// }
// const closeDialog = () => {
//   dialogFormVisible.value = false
//   formData.value = {
//     name: null,
//     type: null,
//     status: true,
//     desc: null,
//   }
// }

const changeStatus = async status => {
  const res = await updateApprovalDefStatus({
    ids: selected.value,
    status: status,
  });
  if (res) {
    selected.value = [];
    checked.value = false;
    AppToast.show({
      message: '更新成功',
      color: 'success',
    });
    if (tableData.value.length === 1 && page.value > 1) {
      page.value--;
    }
    getApprovalDefData();
  }
};

const confirmDelete = async () => {
  if (selected.value.length === 0) {
    AppModal.alert({
      title: '提示',
      content: '请先选择要操作的记录',
    });
    return;
  }

  const ok = await AppModal.confirm({
    title: '删除确认',
    content: `确定要删除选中的 ${selected.value.length} 个审批定义吗？此操作不可恢复。`,
  });
  if (ok) {
    handlerDelete();
  }
};

const handlerDelete = async () => {
  const res = await batchDeleteApprovalDef({
    ids: selected.value,
    action: 'Delete',
  });
  if (res) {
    AppToast.show({
      message: `成功删除 ${selected.value.length} 个审批定义`,
      color: 'success',
    });
    if (tableData.value.length === 1 && page.value > 1) {
      page.value--;
    }
    selected.value = [];
    checked.value = false;
    getApprovalDefData();
  }
};

// const dialogForm = ref(null)
// const enterDialog = async () => {
//   dialogForm.value.validate(async (valid) => {
//     if (!valid) return
//     let res
//     switch (type.value) {
//       case 'create':
//         res = await createApprovalDef(formData.value)
//         break
//       case 'update':
//         res = await updateApprovalDef(formData.value)
//         break
//       default:
//         res = await createApprovalDef(formData.value)
//         break
//     }
//     if (res.code === 0) {
//       ElMessage.success('操作成功')
//       closeDialog()
//       getApprovalDefData()
//     }
//   })
// }
// const openDialog = () => {
//   type.value = 'create'
//   dialogFormVisible.value = true
// }

const pageChange = p => {
  page.value = p;
  getApprovalDefData();
};

// 添加新的方法
const copyApprovalDef = async item => {
  const ok = await AppModal.confirm({
    title: '复制确认',
    content: `确定要复制审批定义「${item.Name}」吗？`,
  });
  if (!ok) return;

  try {
    const res = await createApprovalDef({
      ...item,
      Code: item.Code + '_copy',
      Name: item.Name + ' (副本)',
      Status: 'Draft',
    });
    if (res) {
      AppToast.show({
        message: '复制成功',
        color: 'success',
      });
      getApprovalDefData();
    }
  } catch (error) {
    AppToast.show({
      message: '复制失败',
      color: 'danger',
    });
  }
};

const previewApprovalDef = item => {
  // 打开预览模态框
  window.open(`/admin/approval_def/view?id=${item.ID}&preview=true`, '_blank');
};

const testApprovalDef = item => {
  // 测试审批定义
  AppToast.show({
    message: '测试功能开发中...',
    color: 'info',
  });
};
</script>

<style scoped></style>
