<template>
  <div class="mt-3">
    <!-- search criteria -->
    <div class="border-bottom fs-5 mb-1 pb-2">{{ entityName }} {{ $t('List') }}</div>

    <AppSearch @search="handleSearch" :fieldData="fieldData" :criterias="criterias" />

    <!-- <AppResult
      :page="page"
      :page-size="pageSize"
      :total="total"
    ></AppResult> -->

    <div class="row g-0 h-100">
      <!-- Category Tree Sidebar -->
      <div v-if="treeTable" class="col-md-3 col-lg-2 border-end pe-2 h-100"
        style="height: calc(100vh - 120px) !important; overflow-y: auto;"> <!-- Adjust height as needed -->
        <CategoryTree :table-code="treeTable" :selected-id="selectedCategoryId" @select="handleCategorySelect" />
      </div>

      <!-- Main Content -->
      <div :class="mainContentClass">
        <!-- operation list-->
        <div class="form-group row py-2 px-1">
          <div class="col-sm-10">
            <a class="btn btn-outline-primary btn-sm me-1" :href="'/entity/create?table_code=' + params.table_code"
              role="button">
              <i class="bi bi-file-earmark-plus"></i>
              {{ $t('New') }}
            </a>
            <button type="button" class="btn btn-outline-primary btn-sm me-1" @click="changeStatus('BatchFreeze')">
              {{ $t('Freeze') }}
            </button>
            <button type="button" class="btn btn-outline-primary btn-sm me-1" @click="changeStatus('BatchUnfreeze')">
              {{ $t('Unfreeze') }}
            </button>
            <button type="button" class="btn btn-outline-danger btn-sm me-1" @click="changeStatus('BatchDelete')">
              <i class="bi bi-trash3"></i>
              {{ selected.length ? '(' + selected.length + ')' : '' }}
            </button>
            <button type="button" class="btn btn-outline-primary btn-sm me-1" @click="openImportModal">
              {{ $t('Import') }}
            </button>
            <div class="btn-group">
              <button type="button" class="btn btn-outline-primary btn-sm dropdown-toggle" data-bs-toggle="dropdown"
                aria-expanded="false">
                {{ $t('Export') }}
              </button>
              <ul class="dropdown-menu dropdown-menu-end">
                <li>
                  <button type="button" class="dropdown-item" @click="handleExport('selected')">
                    {{ $t('Selected') }}
                  </button>
                </li>
                <li>
                  <button type="button" class="dropdown-item" @click="handleExport('filtered')">
                    {{ $t('Filtered') }}
                  </button>
                </li>
                <li>
                  <button type="button" class="dropdown-item" @click="handleExport('all')">
                    {{ $t('All') }}
                  </button>
                </li>
              </ul>
            </div>
          </div>
          <div class="col-sm-2 text-end">
            <button type="button" class="btn btn-outline-primary btn-sm ms-1" @click="openColumnSettings"
              :title="$t('Column Settings')">
              <i class="bi bi-gear"></i>
            </button>
          </div>
        </div>

        <!-- data table -->
        <!-- margin-top: 115px;min-height: 658px; -->
        <!-- data table -->
        <!-- margin-top: 115px;min-height: 658px; -->
        <div class="table-responsive text-nowrap"
          style="min-height: calc(100vh - 325px); overflow-y: auto; font-size: 0.9rem" v-if="total">
          <!-- Tree Table Mode -->
          <TreeTable v-if="displayMode === 'Tree'" :items="tableData" :fields="fields" :table-code="params.table_code"
            :frozen-column-code="frozenColumnCode" v-model:selected="selected">
            <template #actions="{ item }">
              <!-- 审批记录 -->
              <a href="javascript:;" @click="getApprovalHistory(item.id)" :title="$t('Approval History')">
                <i class="bi bi-clipboard2-check"></i>
              </a>
              <!-- 修改日志 -->
              <a href="javascript:;" @click="getChangeHistory(item.id)" :title="$t('Change History')">
                <i class="bi bi-file-ruled"></i>
              </a>
              <!-- 分发记录 -->
              <a href="javascript:;" @click="getDeliveryHistory(item.id)" :title="$t('Delivery History')">
                <i class="bi bi-send"></i>
              </a>
              <!-- 历史版本 -->
              <a href="javascript:;" @click="getVersionHistory(item.id)" :title="$t('Version History')">
                <i class="bi bi-clock"></i>
              </a>
              <!-- 修改 -->
              <a :href="'/entity/update?table_code=' + params.table_code + '&id=' + item.id" :title="$t('Update')">
                <i class="bi bi-pencil"></i>
              </a>
              <!-- 查看 -->
              <a :href="'/entity/view?table_code=' + params.table_code + '&id=' + item.id" :title="$t('View')">
                <i class="bi bi-file-text"></i>
              </a>
            </template>
          </TreeTable>

          <!-- Standard List Mode -->
          <table v-else class="table table-sm table-bordered table-striped table-hover w-100 mb-0 sticky-table">
            <thead class="table-light">
              <tr>
                <th class="text-center align-middle sticky-col sticky-col-checkbox">
                  <input type="checkbox" @click="selectAll" v-model="checked" />
                </th>
                <th v-for="field in fields" :key="field.Code" :class="{
                  'sticky-col sticky-col-data': field.code === frozenColumnCode,
                }">
                  {{ field.Name }}
                </th>
                <th v-if="tableFields.length > displayNumber">
                  <a href="javascript:;" @click="selectField">
                    ...
                  </a>
                </th>
                <th class="sticky-col sticky-col-actions text-center" style="width: 100px">
                  {{ $t('Actions') }}
                </th>
              </tr>
            </thead>
            <tbody id="tabletext">
              <tr v-for="item in tableData" :key="item.id">
                <td class="text-center align-middle sticky-col sticky-col-checkbox">
                  <input type="checkbox" v-model="selected" :value="item.id" number />
                </td>
                <td v-for="field in fields" :key="field.Code" :class="{
                  'sticky-col sticky-col-data': field.code === frozenColumnCode,
                }">
                  <!-- status字段使用颜色标识 -->
                  <template v-if="field.code === 'status'">
                    <StatusBadge :status="item[field.code]" />
                  </template>
                  <!-- 其他字段:使用 {field}_display 显示值 -->
                  <template v-else>
                    {{ item[`${field.code}_display`] !== null && item[`${field.code}_display`] !== undefined ?
                      item[`${field.code}_display`] : '' }}
                  </template>
                </td>
                <td v-if="tableFields.length > displayNumber">...</td>
                <td class="sticky-col sticky-col-actions text-center">
                  <!-- 审批记录 -->
                  <a href="javascript:;" @click="getApprovalHistory(item.id)" :title="$t('Approval History')">
                    <i class="bi bi-clipboard2-check"></i>
                  </a>
                  <!-- 修改日志 -->
                  <a href="javascript:;" @click="getChangeHistory(item.id)" :title="$t('Change History')">
                    <i class="bi bi-file-ruled"></i>
                  </a>
                  <!-- 分发记录 -->
                  <a href="javascript:;" @click="getDeliveryHistory(item.id)" :title="$t('Delivery History')">
                    <i class="bi bi-send"></i>
                  </a>
                  <!-- 历史版本 -->
                  <a href="javascript:;" @click="getVersionHistory(item.id)" :title="$t('Version History')">
                    <i class="bi bi-clock"></i>
                  </a>
                  <!-- 修改 -->
                  <a :href="'/entity/update?table_code=' + params.table_code + '&id=' + item.id" :title="$t('Update')">
                    <i class="bi bi-pencil"></i>
                  </a>
                  <!-- 查看 -->
                  <a :href="'/entity/view?table_code=' + params.table_code + '&id=' + item.id" :title="$t('View')">
                    <i class="bi bi-file-text"></i>
                  </a>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="table-responsive text-nowrap p-1" style="min-height: 75vh" v-else>
          {{ $t('Your entity is empty.') }}
        </div>

        <AppPagination :page="page" :page-size="pageSize" :total="total" @page-change="pageChange" />
      </div> <!-- End Main Content Col -->
    </div> <!-- End Row -->

    <!-- Modal Import -->
    <Modal v-model:show="isImportModalVisible" :title="$t('Import')" @confirm="importSubmit">
      <form id="importForm" enctype="multipart/form-data">
        <div class="mb-3">
          <div class="form-check form-check-inline">
            <input class="form-check-input" type="radio" value="BatchCreate" v-model="importData.operation" />
            <label class="form-check-label" for="create">
              {{ $t('For Create') }}
            </label>
          </div>
          <div class="form-check form-check-inline">
            <input class="form-check-input" type="radio" value="BatchUpdate" v-model="importData.operation" />
            <label class="form-check-label" for="update">
              {{ $t('For Update') }}
            </label>
          </div>
        </div>
        <div class="mb-3">
          <div class="form-text text-danger">
            {{ $t('Please') }}
            <a href="javascript:;" @click="downloadTemplate">
              {{ $t('Download') }}
            </a>
            {{ $t('and') }}{{ $t('check the last template.') }}
          </div>
        </div>
        <div class="mb-3">
          <label for="file" class="form-label">
            {{ $t('File') }}:
          </label>
          <input ref="fileInput" class="form-control form-control-sm" id="file" type="file" name="file"
            @change="uploadFile" />
        </div>
        <div class="mb-3">
          <label for="reason" class="col-form-label">
            {{ $t('Reason') }}:
          </label>
          <textarea class="form-control" id="reason" name="reason" v-model="importData.reason"></textarea>
        </div>
      </form>
    </Modal>

    <!-- Column Settings Modal -->
    <Modal v-model:show="isColumnSettingVisible" :title="$t('Column Settings')" @confirm="saveColumnSettings">
      <div class="mb-2">
        <button type="button" class="btn btn-sm btn-link p-0 me-2" @click="checkAllColumns">
          {{ $t('Select All') }}
        </button>
        <button type="button" class="btn btn-sm btn-link p-0" @click="uncheckAllColumns">
          {{ $t('Deselect All') }}
        </button>
      </div>
      <div class="row g-2">
        <div class="col-6 d-flex align-items-center" v-for="field in tableFields" :key="field.code">
          <div class="form-check d-flex align-items-center">
            <input class="form-check-input" type="checkbox" :value="field.code" :id="'col_' + field.code"
              v-model="tempSelectedColumnCodes" />
            <i class="bi ms-2 me-1"
              :class="tempFrozenColumnCode === field.code ? 'bi-lock-fill text-primary' : 'bi-unlock text-muted'"
              @click="toggleFreeze(field.code)" style="cursor: pointer" :title="$t('Freeze Column')"></i>
            <label class="form-check-label text-truncate" :for="'col_' + field.code" :title="field.Name">
              {{ field.Name }}
            </label>
          </div>
        </div>
      </div>
    </Modal>
  </div>
</template>

<script setup>
import {
  getEntityHistoryList,
  getEntityList,
  getEntityLogList,
  getExportFile,
  getTemplate,
  importFile,
  updateEntityStatus,
} from '@/api/entity';
import { findTableList } from '@/api/table';
import { getTableFields as getTableFieldsAPI } from '@/api/table_field';
import { getWebhookDeliveryList } from '@/api/webhook_delivery';
import { AppModal } from '@/components/Modal/modal.js';
import Modal from '@/components/Modal/Modal.vue';
import AppPagination from '@/components/Pagination.vue';
import { AppToast } from '@/components/toast.js';
import httpLinkHeader from 'http-link-header';
import { computed, h, onMounted, ref, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import 'vue-select/dist/vue-select.css';
import AppSearch from './Search.vue';
import TreeTable from '@/components/TreeTable.vue';
import CategoryTree from '@/components/CategoryTree.vue';
import StatusBadge from '@/components/StatusBadge.vue';
import { formatFieldValue, preloadFieldDictionaries } from '@/utils/fieldFormatter';

const isImportModalVisible = ref(false);
const fileInput = ref(null);

const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const searchData = ref({});
const page = ref(1);
const pageSize = ref(15);
const total = ref(0);
const tableData = ref([]);
const selected = ref([]);
const checked = ref(false);
const tableFields = ref([]);
const displayNumber = ref(0);
const params = ref({ ...route.query });

// Watch route params change
watch(
  () => route.query,
  (newQuery) => {
    params.value = { ...newQuery };
    // Optionally reload data if needed, but onMounted/initPage handles initial load.
    // Ideally we should reload if table_code changes.
    // For now, let's just ensure params has data.
    if (newQuery.table_code) {
      initPage();
    }
  }
);
const reason = ref('');
const fieldData = ref([]);
const importData = ref({
  operation: 'BatchCreate', // Default select create
});
const entityName = ref('');
const isColumnSettingVisible = ref(false);
const selectedColumnCodes = ref([]); // Final effective column config
const tempSelectedColumnCodes = ref([]); // Temporary config in modal
const frozenColumnCode = ref(''); // Current frozen column Code
const tempFrozenColumnCode = ref(''); // Temporary frozen config in modal

const criterias = ref([
  {
    field: 'ID',
    symbol: '=',
    value: '',
  },
]);

const displayMode = ref('List');
const treeTable = ref('');
const selectedCategoryId = ref(null);

const mainContentClass = computed(() =>
  treeTable.value ? 'col-md-9 col-lg-10 ps-2 h-100' : 'col-12 h-100'
);

const detectParentTreeTable = async (parentTableCode) => {
  if (!parentTableCode) return;
  try {
    const res = await findTableList({ code: parentTableCode });
    const parentTable = res?.data?.[0];
    if (parentTable?.DisplayMode === 'Tree' || parentTable?.display_mode === 'Tree') {
      treeTable.value = parentTableCode;
    }
  } catch (error) {
    console.error('Failed to fetch parent table info:', error);
  }
};

// Get table name and metadata
const getEntityName = async () => {
  const where = { status: 'Normal' };
  if (params.value.table_code) {
    where.code = params.value.table_code;
  }

  const res = await findTableList(where);
  if (!res?.data?.length) return;

  const table = res.data[0];
  // Compatible with CamelCase and snake_case
  const displayModeVal = table.DisplayMode || table.display_mode || 'List';
  const parentTableVal = table.ParentTable || table.parent_table || '';

  entityName.value = table.Name || table.name;
  displayMode.value = displayModeVal;

  await detectParentTreeTable(parentTableVal);

  // If it is tree mode, adjust page size to load more data (temporarily 10000, can be optimized later for backend full amount interface)
  if (displayMode.value === 'Tree') {
    pageSize.value = 10000;
  }
};

onMounted(async () => {
  params.value = router.currentRoute.value.query;
  displayNumber.value = 700;
  await getEntityName();
  // Load field definition first, then load data (data formatting requires field information)
  await getTableFields();
  loadColumnSettings(); // Load user column settings
  getEntityData();
});

// Handle category selection
const handleCategorySelect = (node) => {
  // Update selected category ID for highlighting (use code for string-based IDs)
  selectedCategoryId.value = node?.code || node?.id || null;

  // Find field linked to treeTable
  const categoryField = tableFields.value.find(f =>
    f.Options?.relation?.target === treeTable.value
  );

  if (!categoryField) {
    console.warn(`No field found linking to tree table: ${treeTable.value}`);
    return;
  }

  const filterValue = node?.code || node?.id;
  // Use 'like' query for hierarchical filtering (starts with filterValue)
  const likeKey = `${categoryField.Code} like`;
  const exactKey = categoryField.Code;

  if (filterValue) {
    searchData.value[likeKey] = `${filterValue}%`;
  } else {
    delete searchData.value[likeKey];
  }

  // Clear exact match to prevent conflicts
  delete searchData.value[exactKey];

  // Reset pagination and query
  page.value = 1;
  getEntityData();
};



// Get entity list
const getEntityData = async () => {
  try {
    // showSpinner.value = true
    const queryParams = {
      table_code: params.value.table_code,
      page: page.value,
      pageSize: pageSize.value,
      ...searchData.value,
    };

    const res = await getEntityList(queryParams);
    if (res.status === 200) {
      // Check if data exists
      if (!res.data || res.data.length === 0) {
        tableData.value = [];
        total.value = 0;
        AppToast.show({
          message: t('No data found'),
          color: 'info',
        });
        return;
      }

      // 1. Preload all required dictionary data
      await preloadFieldDictionaries(tableFields.value);

      // 2. Preprocess each row, add {field}_display
      tableData.value = await Promise.all(
        res.data.map(async (item) => {
          const processedItem = { ...item };  // Keep original data

          // Generate display value for each field
          for (const field of tableFields.value) {
            const value = item[field.code];
            const formattedValue = await formatFieldValue(value, field);
            // Add {field}_display, null shows as empty string
            processedItem[`${field.code}_display`] = formattedValue ?? '';
          }

          return processedItem;
        })
      );

      if (res.headers.link) {
        try {
          const links = httpLinkHeader.parse(res.headers.link).refs;
          links.forEach(link => {
            if (['last'].includes(link.rel)) {
              const url = new URL(link.uri);
              total.value = parseInt(url.searchParams.get('page')) || 1;
            }
          });
        } catch (e) {
          console.warn("Failed to parse Link header", e);
        }
      }

      // Fallback: if total is still 0 but we have data, assume at least 1 page
      if (total.value === 0 && tableData.value.length > 0) {
        total.value = 1;
      }
    }
  } catch (error) {
    // Check for permission error
    if (error.response?.status === 403) {
      AppModal.alert(
        t('You do not have permission to access this table, please contact the administrator.'),
        () => {
          // Return to home or other page
          window.location.href = '/';
        }
      );
    } else {
      // Other errors, show generic error message
      console.error('Failed to fetch data:', error);
      AppToast.show({
        message: t('Failed to fetch data: ') + (error.response?.data?.message || error.message),
        color: 'danger',
      });
    }
  }
};

// find all fields for a table
const getTableFields = async () => {
  try {
    const res = await getTableFieldsAPI({
      table_code: params.value.table_code,
    });
    if (res && res.data) {
      // Only show fields where is_show is true
      const visibleFields = res.data.filter(f => f.is_show);

      // Convert to field structure, including Options
      tableFields.value = visibleFields.map(f => ({
        Code: f.code,
        code: f.code,  // Lowercase code for data access
        Name: f.name,
        Type: f.type,
        FieldType: f.field_type,  // Field type
        field_type: f.field_type,  // Lowercase field type
        IsSystem: f.is_system,
        Options: f.options,  // Full Options config
        relation: f.options?.relation,  // Relation config (for formatter)
      }));

      // fieldData for search component
      fieldData.value = [...tableFields.value];
    }
  } catch (error) {
    console.error('Failed to fetch field list:', error);
    // If failed, keep original logic
    tableFields.value = [];
    fieldData.value = [];
  }
};

// fliter display field, top 10 or selected
const fields = computed(() => {
  let visibleFields = [];
  // If no config (empty array), default show top N
  if (selectedColumnCodes.value.length === 0) {
    visibleFields = tableFields.value.filter((n, index) => {
      return index < displayNumber.value;
    });
  } else {
    // Filter by config, keep tableFields order
    // Filter out fields not in tableFields (prevent errors if field deleted)
    visibleFields = tableFields.value.filter(f =>
      selectedColumnCodes.value.includes(f.code)
    );
  }

  // If frozen column exists, move it to the front
  if (frozenColumnCode.value) {
    const frozenIndex = visibleFields.findIndex(
      f => f.code === frozenColumnCode.value
    );
    if (frozenIndex > -1) {
      // Create new array to avoid side effects, and move frozen column
      const frozenField = visibleFields[frozenIndex];
      const otherFields = visibleFields.filter(
        f => f.code !== frozenColumnCode.value
      );
      return [frozenField, ...otherFields];
    }
  }
  return visibleFields;
});

// select all/ unselect all
const selectAll = () => {
  var ids = [];
  if (!checked.value) {
    tableData.value.forEach(function (val) {
      ids.push(val.id);
    });
    selected.value = ids;
  } else {
    selected.value = [];
  }
};

// Note: formatFieldValue imported from @/utils/fieldFormatter.js
// Data preprocessing done in getEntityData, template uses {field}_display


const getApprovalHistory = async id => {
  try {
    const res = await getEntityList({
      table_code: params.value.table_code,
      entity_id: id,
      is_draft: 'Yes',
    });

    let bodyContent = `<table class="table table-sm table-hover"><thead><tr><th>${t('Code')}</th><th>${t('Process Code')}</th><th>${t('Created At')}</th></tr></thead><tbody>`;

    // Check if data exists
    if (res && res.data && res.data.length > 0) {
      res.data.forEach(item => {
        bodyContent += '<tr>' +
          '<td>' + (item.id || '') + '</td>' +
          '<td>' + (item.approval_code || '-') + '</td>' +
          '<td>' + (item.created_at || '') + '</td>' +
          '</tr>';
      });
    } else {
      // Show hint when no data
      bodyContent += `<tr><td colspan="3" class="text-center text-muted">${t('No approval history')}</td></tr>`;
    }

    bodyContent += '</tbody></table>';

    AppModal.alert({
      title: t('Approval History'),
      bodyHtml: true,
      bodyContent: bodyContent,
      size: 'lg',
    });
  } catch (error) {
    console.error('Failed to fetch approval history:', error);
    AppToast.show({
      message: t('Failed to fetch approval history'),
      color: 'danger',
    });
  }
};

const getChangeHistory = async id => {
  try {
    const res = await getEntityLogList({
      table_code: params.value.table_code,
      entity_id: id,  // 关联的实体ID
    });

    let bodyContent = `<table class="table table-sm table-hover"><thead><tr><th>${t('Code')}</th><th>${t('Field Name')}</th><th>${t('Before Update')}</th><th>${t('After Update')}</th><th>${t('Reason')}</th><th>${t('Updated By')}</th><th>${t('Updated At')}</th></tr></thead><tbody>`;

    // Check if data exists
    if (res && res.data && res.data.length > 0) {
      res.data.forEach(item => {
        bodyContent += '<tr>' +
          '<th>' + (item.id || '') + '</th>' +
          '<td>' + (item.field_name || '') + '</td>' +
          '<td>' + (item.before_update || '') + '</td>' +
          '<td>' + (item.after_update || '') + '</td>' +
          '<td>' + (item.reason || '') + '</td>' +
          '<td>' + (item.update_by || '') + '</td>' +
          '<td>' + (item.updated_at || '') + '</td>' +
          '</tr>';
      });
    } else {
      // Show hint when no data
      bodyContent += `<tr><td colspan="7" class="text-center text-muted">${t('No change history')}</td></tr>`;
    }

    bodyContent += '</tbody></table>';

    AppModal.alert({
      title: t('Change History'),
      bodyHtml: true,
      bodyContent: bodyContent,
      size: 'xl',
    });
  } catch (error) {
    console.error('Failed to fetch change history:', error);
    AppToast.show({
      message: t('Failed to fetch change history'),
      color: 'danger',
    });
  }
};

const getDeliveryHistory = async id => {
  try {
    const res = await getWebhookDeliveryList({
      table_code: params.value.table_code,
      entity_id: id,
    });

    let bodyContent = `<table class="table table-sm table-hover"><thead><tr><th>${t('Code')}</th><th>${t('Delivery Code')}</th><th>${t('Status')}</th><th>${t('Completed At')}</th></tr></thead><tbody>`;

    // Check if data exists
    if (res && res.data && res.data.length > 0) {
      res.data.forEach(item => {
        bodyContent += '<tr>' +
          '<td>' + (item.ID || '') + '</td>' +
          '<td>' + (item.DeliveryCode || '') + '</td>' +
          '<td>' + (item.Status || '') + '</td>' +
          '<td>' + (item.CompletedAt || '') + '</td>' +
          '</tr>';
      });
    } else {
      // Show hint when no data
      bodyContent += `<tr><td colspan="4" class="text-center text-muted">${t('No delivery history')}</td></tr>`;
    }

    bodyContent += '</tbody></table>';

    AppModal.alert({
      title: t('Delivery History'),
      bodyHtml: true,
      bodyContent: bodyContent,
      size: 'lg',
    });
  } catch (error) {
    console.error('Failed to fetch delivery history:', error);
    AppToast.show({
      message: t('Failed to fetch delivery history'),
      color: 'danger',
    });
  }
};

const getVersionHistory = async id => {
  try {
    const res = await getEntityHistoryList({
      table_code: params.value.table_code,
      entity_id: id,
    });

    let bodyContent = `<table class="table table-sm table-hover"><thead><tr><th>ID</th><th>${t('Code')}</th><th>${t('Name')}</th><th>${t('Entity ID')}</th><th>${t('Updated At')}</th></tr></thead><tbody>`;

    // Check if data exists
    if (res && res.data && res.data.length > 0) {
      res.data.forEach(item => {
        bodyContent += '<tr>' +
          '<td>' + (item.id || '') + '</td>' +
          '<td>' + (item.code || '') + '</td>' +
          '<td>' + (item.name || '') + '</td>' +
          '<td>' + (item.entity_id || '') + '</td>' +
          '<td>' + (item.updated_at || '') + '</td>' +
          '</tr>';
      });
    } else {
      // Show hint when no data
      bodyContent += `<tr><td colspan="5" class="text-center text-muted">${t('No version history')}</td></tr>`;
    }

    bodyContent += '</tbody></table>';

    AppModal.alert({
      title: t('Version History'),
      bodyHtml: true,
      bodyContent: bodyContent,
      size: 'lg',
    });
  } catch (error) {
    console.error('Failed to fetch version history:', error);
    AppToast.show({
      message: t('Failed to fetch version history'),
      color: 'danger',
    });
  }
};

// Modal for freeze, unfreeze, delete
const changeStatus = async operation => {
  // Check if records are selected
  if (selected.value.length === 0) {
    AppModal.alert({
      content: t('Please select records first'),
      title: t('Hint'),
    });
    return;
  }

  // Clear reason input
  reason.value = '';

  // Set title based on operation
  let title = '';
  let actionText = '';

  switch (operation) {
    case 'BatchFreeze':
      // Check if deleted data is included
      for (const id of selected.value) {
        const item = tableData.value.find(i => i.id === id);
        if (item && item.status === 'Deleted') {
          AppModal.alert({
            content: t('Selected data contains deleted items, operation not allowed'),
            title: t('Hint'),
          });
          return;
        }
      }
      title = t('Confirm Freeze');
      actionText = t('Freeze data');
      break;
    case 'BatchUnfreeze':
      // Check if deleted data included
      for (const id of selected.value) {
        const item = tableData.value.find(i => i.id === id);
        if (item && item.status === 'Deleted') {
          AppModal.alert({
            content: t('Selected data contains deleted items, operation not allowed'),
            title: t('Hint'),
          });
          return;
        }
      }
      title = t('Confirm Unfreeze');
      actionText = t('Unfreeze data');
      break;
    case 'BatchDelete':
      title = t('Confirm Delete');
      actionText = t('Delete data');
      break;
    default:
      title = t('Confirm Operation');
      actionText = t('Operate');
  }

  // Create custom content function with reason input
  const content = () => {
    return h('div', [
      h('div', { class: 'mb-3' }, [
        h('p', {}, t('Are you sure you want to {actionText} for {count} records?', { actionText, count: selected.value.length })),
      ]),
      h('div', { class: 'mb-3' }, [
        h('label', { class: 'form-label', htmlFor: 'reasonInput' }, t('Reason') + '：'),
        h('textarea', {
          id: 'reasonInput',
          class: 'form-control',
          rows: 3,
          placeholder: t('Please enter operation reason...'),
          value: reason.value || '',
          onInput: e => {
            reason.value = e.target.value;
          },
        }),
      ]),
    ]);
  };

  const result = await AppModal.confirm({
    title: title,
    content: content,
    size: 'md',
  });

  if (result) {
    // User clicked confirm, call changeStatusConfirm function
    await changeStatusConfirm(operation);
  }
};

// Confirm Freeze, Unfreeze, Delete
const changeStatusConfirm = async operation => {
  const res = await updateEntityStatus({
    table_code: params.value.table_code,
    ids: selected.value,
    operation: operation,
    reason: reason.value,
  });
  if (res.status === 200) {
    selected.value = [];
    checked.value = false;
    AppToast.show({
      message: '更新成功',
      color: 'success',
    });
    if (tableData.value.length === 1 && page.value > 1) {
      page.value--;
    }
    getEntityData();
  }
};

// Open import modal and clear file input
const openImportModal = () => {
  // Clear importData
  importData.value = {
    operation: 'BatchCreate',
    reason: '',
    file: null,
  };
  // Clear file input
  if (fileInput.value) {
    fileInput.value.value = '';
  }
  // Open modal
  isImportModalVisible.value = true;
};

const downloadTemplate = async () => {
  if (!importData.value.operation) {
    AppToast.show({
      message: t('Please select operation type'),
      color: 'warning',
    });
  }
  const res = await getTemplate({
    table_code: params.value.table_code,
    operation: importData.value.operation,
  });
  if (res) {
    let baseUrl = import.meta.env.VITE_BASE_API;
    window.location.href = baseUrl + res.data.export_url;
  }
};

const uploadFile = event => {
  importData.value.file = event.target.files[0];
};

const importSubmit = async () => {
  // Validate required fields
  if (!importData.value.file || !importData.value.operation) {
    AppToast.show({
      message: t('Please complete parameters: file and operation type are required'),
      color: 'warning',
    });
    return;
  }

  const formData = new FormData();

  formData.append('file', importData.value.file);
  formData.append('operation', importData.value.operation);
  formData.append('reason', importData.value.reason);
  formData.append('table_code', params.value.table_code);

  try {
    const res = await importFile(formData);
    if (res) {
      AppToast.show({
        message: t('Submission successful'),
        color: 'success',
      });
      // Clear form data, reset to default
      importData.value = {
        operation: 'BatchCreate',
      };
      // Close import dialog
      isImportModalVisible.value = false;
      // Refresh page data
      getEntityData();
    }
  } catch (error) {
    console.error('Import failed:', error);
    AppToast.show({
      message: t('Import failed, please check network or file format'),
      color: 'danger',
    });
  }
};

const handleExport = async filter => {
  if (filter === 'selected' && selected.value.length === 0) {
    AppModal.alert({
      content: t('Please select data to export'),
      title: t('Hint'),
    });
    return;
  }
  const res = await getExportFile({
    table_code: params.value.table_code,
    filter: filter,
    ids: selected.value.join(','),
    ...searchData.value,
  });
  if (res.status === 200) {
    let baseUrl = import.meta.env.VITE_BASE_API;
    let url = baseUrl + res.data.export_url;
    window.location.replace(url);
    AppToast.show({
      message: t('Export task started, please check download'),
      color: 'success',
    });
  }
};

const pageChange = p => {
  page.value = p;
  getEntityData();
};

const selectField = () => { };
// Search condition method
const handleSearch = () => {
  // Save category filter params
  let categoryParam = {};
  if (treeTable.value) {
    const categoryField = tableFields.value.find(f =>
      f.Options?.relation?.target === treeTable.value
    );
    if (categoryField && searchData.value[categoryField.Code]) {
      categoryParam[categoryField.Code] = searchData.value[categoryField.Code];
    }
  }

  // Reset search data
  // Note: Original code used [], changed to {}
  searchData.value = { ...categoryParam };

  for (const criteria of criterias.value) {
    if (criteria.value !== undefined && criteria.value !== '') {
      searchData.value[criteria.field + ' ' + criteria.symbol] = criteria.value;
    }
  }
  page.value = 1; // Reset page number on search
  getEntityData();
};
// --- Column Settings Logic ---

// Load config
const loadColumnSettings = () => {
  const key = `piemdm_column_config_${params.value.table_code}`;
  const stored = localStorage.getItem(key);
  if (stored) {
    try {
      selectedColumnCodes.value = JSON.parse(stored);
    } catch (e) {
      console.error('Failed to parse column settings', e);
      selectedColumnCodes.value = [];
    }
  }

  // Load frozen column config
  const frozenKey = `piemdm_column_frozen_${params.value.table_code}`;
  frozenColumnCode.value = localStorage.getItem(frozenKey) || '';
};

// Open modal
const openColumnSettings = () => {
  // 1. If active config exists, use it
  // 2. If no config (first open), select all visible columns (or columns in fields computed property)
  if (selectedColumnCodes.value.length > 0) {
    tempSelectedColumnCodes.value = [...selectedColumnCodes.value];
  } else {
    // Default select all visible columns (based on displayNumber logic)
    tempSelectedColumnCodes.value = tableFields.value
      .filter((_, index) => index < displayNumber.value)
      .map(f => f.code);
  }
  // Initialize frozen column temp state
  tempFrozenColumnCode.value = frozenColumnCode.value;

  isColumnSettingVisible.value = true;
};

// Save config
const saveColumnSettings = () => {
  selectedColumnCodes.value = [...tempSelectedColumnCodes.value];
  const key = `piemdm_column_config_${params.value.table_code}`;
  localStorage.setItem(key, JSON.stringify(selectedColumnCodes.value));

  // Save frozen column config
  frozenColumnCode.value = tempFrozenColumnCode.value;
  const frozenKey = `piemdm_column_frozen_${params.value.table_code}`;
  localStorage.setItem(frozenKey, frozenColumnCode.value);

  isColumnSettingVisible.value = false;
};

// Toggle freeze status
const toggleFreeze = (code) => {
  if (tempFrozenColumnCode.value === code) {
    // If selected, cancel
    tempFrozenColumnCode.value = '';
  } else {
    // Otherwise select current column (exclusive)
    tempFrozenColumnCode.value = code;
  }
};

// Select all columns
const checkAllColumns = () => {
  tempSelectedColumnCodes.value = tableFields.value.map(f => f.code);
};

// Deselect all columns
const uncheckAllColumns = () => {
  const allCodes = tableFields.value.map(f => f.code);
  tempSelectedColumnCodes.value = allCodes.filter(
    code => !tempSelectedColumnCodes.value.includes(code)
  );
};
</script>

<style scoped></style>
