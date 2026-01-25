<template>
  <div class="mt-3">
    <div class="border-bottom fs-5 mb-2 pb-2">{{ entityName }} {{ $t('Create') }}</div>
    <ul class="nav nav-tabs mh-30" id="myTab" role="tablist">
      <li class="nav-item mh-30" role="presentation">
        <a class="nav-link mh-30 active" id="baseinfo-tab" data-toggle="tab" href="#baseinfo">
          {{ $t('Basic Info') }}
        </a>
      </li>
    </ul>
    <div class="card-body mt-2">
      <div id="create_wrapper">
        <div class="tab-content" id="myTabContent">
          <div class="tab-pane fade show active" id="base-info-tab-pane" tabindex="0">
            <AppForm v-model:dataInfo="dataInfo" :fields="tableFields" :dictionarys="dictionarys"
              :readonlyFields="readonlyFields" @fetch-data="fetchData" />
            <div class="form-group row mt-3 mb-3">
              <div class="col-sm-auto">
                <button type="button" class="btn btn-outline-primary btn-sm me-2" @click="submitForm(dataInfo)">
                  {{ $t('Submit') }}
                </button>
                <button type="button" class="btn btn-outline-secondary btn-sm me-2" @click="goIndex">
                  {{ $t('Cancel') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { createEntity, getEntityList } from '@/api/entity';
import { getTableFields, findTableList } from '@/api/table_field';

import { AppModal } from '@/components/Modal/modal.js';
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import AppForm from './Form.vue';

const router = useRouter();
const { t } = useI18n();
const dataInfo = ref({});
const tableData = ref([]);
const tableFields = ref([]);
const params = ref({});
const dictionarys = ref([]);
const readonlyFields = ref([]);
const entityName = ref('');

// Mark if config warning has been shown
let hasShownConfigWarning = false;

onMounted(async () => {
  params.value = router.currentRoute.value.query;

  // Set URL query params as form initial values (for Item creation with relation fields)
  // Exclude special params: table_code, redirect_url
  const initialData = {};
  const relationFields = []; // Relation fields list
  Object.keys(params.value).forEach(key => {
    if (key !== 'table_code' && key !== 'redirect_url') {
      initialData[key] = params.value[key];
      relationFields.push(key); // Add to readonly fields list
    }
  });
  dataInfo.value = initialData;
  readonlyFields.value = relationFields; // Set readonly fields

  await getEntityName();
  await getFieldData();
});

const displayMode = ref('List');

// Get table name
const getEntityName = async () => {
  const where = {
    status: 'Normal',
  };
  if (params.value.table_code) {
    where.code = params.value.table_code;
  }
  const res = await findTableList(where);
  if (res && res.data && res.data.length > 0) {
    entityName.value = res.data[0].Name;
    displayMode.value = res.data[0].DisplayMode || res.data[0].display_mode || 'List';
  }
};

// Query all fields (using getTableFields API to get sorted fields)
const getFieldData = async () => {
  const res = await getTableFields({
    table_code: params.value.table_code,
  });

  if (res && res.data) {
    // getTableFields returns fields in correct format, use directly
    // Field names use lowercase underscore format (code, name, field_type, sort, etc.)
    // Filter out system fields (is_system === true)
    tableFields.value = res.data
      .filter(field => {
        if (field.is_system) return false;
        // If tree table, filter out level and path
        if (displayMode.value === 'Tree') {
          const code = (field.code || field.Code || '').toLowerCase();
          if (code === 'level' || code === 'path') {
            return false;
          }
        }
        return true;
      })
      .map(field => ({
        ID: field.id,
        Code: field.code,
        TableCode: params.value.table_code,
        Name: field.name,
        FieldType: field.field_type,
        Type: field.type,
        Length: field.length,
        Required: field.required ? 'Yes' : 'No',
        IsShow: field.is_show ? 'Yes' : 'No',
        IsFilter: field.is_filter ? 'Yes' : 'No',
        Sort: field.sort,
        Options: field.options,
        Status: 'Normal',
      }));

    // Initialize default values for date, time, datetime fields
    initializeDateTimeDefaults(tableFields.value);

    // Validate Select field configuration
    const misconfiguredFieldsList = [];
    res.data.forEach(field => {
      const isSelectField = field.field_type === 'select' ||
        field.field_type === 'multiselect' ||
        field.field_type === 'radio' ||
        field.field_type === 'checkboxgroup' ||
        field.options?.ui?.widget === 'Select' ||
        field.options?.ui?.widget === 'MultiSelect' ||
        field.options?.ui?.widget === 'RadioGroup' ||
        field.options?.ui?.widget === 'CheckboxGroup';

      const hasTarget = field.options?.relation?.target &&
        field.options.relation.target.trim() !== '';

      if (isSelectField && !hasTarget) {
        misconfiguredFieldsList.push({
          code: field.code,
          name: field.name,
          fieldType: field.field_type
        });
      }
    });

    // If there are unconfigured fields, show Modal (only once)
    if (misconfiguredFieldsList.length > 0 && !hasShownConfigWarning) {
      hasShownConfigWarning = true;

      // Use AppModal.alert() to show warning
      const fieldList = misconfiguredFieldsList
        .map(f => `
            <li>
              <strong>${f.name}</strong> <code>(${f.code})</code>
            </li>
          `).join('');

      AppModal.alert({
        title: t('⚠️ Incomplete Field Configuration'),
        okTitle: '',
        cancelTitle: t('Got it'),
        bodyHtml: true,
        bodyContent: `
            <p>${t('The following select fields are missing relation table configuration (relation.target):')}</p>
            <ul>${fieldList}</ul>
            <div>
              ${t('Please contact the development team to configure the relation tables for these fields in field management.')}
            </div>
          `
      });
    }

    const dicts = {};

    // Use Promise.all to wait for all async operations to complete
    const promises = res.data.map(async field => {
      // Get relation config (new way first, backward compatible with old way)
      const relationTable = field.options?.relation?.target;
      const relationFilter = field.options?.relation?.filter;

      // Load options for all fields with relation config
      // Exclude empty strings and undefined
      if (relationTable && relationTable.trim() !== '') {
        try {
          const res2 = await getEntityList({
            table_code: relationTable,
            pageSize: 1000, // Load enough options
            ...relationFilter, // Apply filter conditions
          });

          if (res2) {
            dicts['dict-' + field.code] = res2.data;
          }
        } catch (error) {
          console.error('Error loading options for', field.code, ':', error);
        }
      }
    });

    // Wait for all options to load
    await Promise.all(promises);
    dictionarys.value = dicts;
  }
};

const fetchData = async (search, loading, field) => {
  const relationTable = field.Options?.relation?.target;

  // 如果没有关联表配置,打印详细日志帮助调试
  if (!relationTable) {
    console.warn('No relation target configured for field:', field.Code);
    return;
  }

  if (loading) loading(true);
  try {
    const relationCode = field.Options?.relation?.valueField;
    const relationName = field.Options?.relation?.labelField;

    // 从 relation.filter 读取过滤条件
    const filterConditions = field.Options?.relation?.filter || {};

    // 构建查询参数
    const queryParams = {
      table_code: relationTable,
      pageSize: 100,
      // 添加所有过滤条件
      ...filterConditions
    };

    // 只有当搜索字符串非空且有 valueField/labelField 配置时才添加搜索条件
    if (search && search.length > 0 && relationCode && relationName) {
      queryParams['(' + relationCode + ' like ? or ' + relationName + ' like ? )'] =
        '%' + search + '%,%' + search + '%';
    }

    const res = await getEntityList(queryParams);

    if (res && res.data) {
      // 使用响应式更新,避免直接清空数组
      dictionarys.value = {
        ...dictionarys.value,
        ['dict-' + field.Code]: res.data
      };
    }
  } catch (error) {
    console.error('Error fetching options:', error);
  } finally {
    if (loading) loading(false);
  }
};

// Initialize default values for date, time, datetime fields
const initializeDateTimeDefaults = (fields) => {
  const now = new Date();

  fields.forEach(field => {
    // If field already has value (e.g. passed from URL query), skip
    if (dataInfo.value[field.Code] !== undefined && dataInfo.value[field.Code] !== null) {
      return;
    }

    // CheckboxGroup and MultiSelect fields: initialize as empty array
    if (field.FieldType === 'checkboxgroup' || field.FieldType === 'multiselect') {
      dataInfo.value[field.Code] = [];
    }
    // Date field: yyyy-MM-dd (match VueDatePicker model-type)
    else if (field.Type === 'Date' || field.FieldType === 'date') {
      const year = now.getFullYear();
      const month = String(now.getMonth() + 1).padStart(2, '0');
      const day = String(now.getDate()).padStart(2, '0');
      dataInfo.value[field.Code] = `${year}-${month}-${day}`;
    }
    // Time field: HH:mm:ss (match VueDatePicker model-type)
    else if (field.FieldType === 'time') {
      const hours = String(now.getHours()).padStart(2, '0');
      const minutes = String(now.getMinutes()).padStart(2, '0');
      const seconds = String(now.getSeconds()).padStart(2, '0');
      dataInfo.value[field.Code] = `${hours}:${minutes}:${seconds}`;
    }
    // DateTime field: yyyy-MM-dd HH:mm:ss (match VueDatePicker model-type)
    else if (field.Type === 'DateTime' || field.FieldType === 'datetime') {
      const year = now.getFullYear();
      const month = String(now.getMonth() + 1).padStart(2, '0');
      const day = String(now.getDate()).padStart(2, '0');
      const hours = String(now.getHours()).padStart(2, '0');
      const minutes = String(now.getMinutes()).padStart(2, '0');
      const seconds = String(now.getSeconds()).padStart(2, '0');
      dataInfo.value[field.Code] = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
    }
  });
};

const submitForm = async data => {
  try {
    // Validate reason field
    if (!data.reason || data.reason.trim() === '') {
      AppModal.alert({
        title: t('Validation Failed'),
        content: t('Modification reason is required'),
      });
      return;
    }

    // Serialize array fields (array -> JSON string)
    const serializedData = { ...data };
    tableFields.value.forEach(field => {
      // Attachment, checkbox group, multi-select fields need to be serialized to JSON string
      if ((field.FieldType === 'attachment' ||
        field.FieldType === 'checkboxgroup' ||
        field.FieldType === 'multiselect') &&
        Array.isArray(serializedData[field.Code])) {
        serializedData[field.Code] = JSON.stringify(serializedData[field.Code]);
      }
    });

    const res = await createEntity({
      table_code: params.value.table_code,
      ...serializedData,
    });

    // Only show success message on success
    if (res) {
      AppModal.alert({
        title: t('Hint'),
        content: 'Create success',
      }).then(() => {
        goIndex();
      });
    }
  } catch (error) {
    // Error already handled and displayed in axios interceptor, no need to show again here
    console.error('Create entity failed:', error);
  }
};

function goIndex() {
  // Check if there is redirect_url parameter
  if (params.value.redirect_url) {
    router.push(decodeURIComponent(params.value.redirect_url));
  } else {
    router.push('/entity/index?table_code=' + params.value.table_code);
  }
}
</script>

<style scoped></style>
