<template>
  <form name="entityForm" id="entityForm" method="post" enctype="multipart/form-data">
    <div class="col-12 col-sm-12">
      <div class="row">
        <div class="col-sm-12">
          <div class="form-group row">
            <legend :class="['col-form-label', 'col-sm-2', { required: requiredFields.name }]">
              {{ $t('Name') }}:
            </legend>
            <div class="col-sm-auto">
              <input type="text" class="form-control form-control-sm" v-model="name" v-bind="nameAttrs" name="name"
                :placeholder="$t('Name')" maxlength="64" size="64" />
              <div v-if="errors.Name" class="text-danger small mt-1">
                {{ errors.Name }}
              </div>
            </div>
          </div>
        </div>
        <div class="col-sm-12" v-if="platform !== 'Builtin'">
          <div class="form-group row">
            <legend :class="['col-form-label', 'col-sm-2', { required: requiredFields.code }]">
              {{ $t('Code') }}:
            </legend>
            <div class="col-sm-auto">
              <input type="text" class="form-control form-control-sm" v-model="code" v-bind="codeAttrs" name="code"
                :placeholder="$t('Code')" maxlength="128" size="64" />
              <div v-if="errors.Code" class="text-danger small mt-1">
                {{ errors.Code }}
              </div>
            </div>
          </div>
        </div>
        <div class="col-sm-12">
          <div class="form-group row">
            <legend :class="['col-form-label', 'col-sm-2', { required: requiredFields.description }]">
              {{ $t('Description') }}:
            </legend>
            <div class="col-sm-auto">
              <input type="text" class="form-control form-control-sm" v-model="description" v-bind="descriptionAttrs"
                name="description" :placeholder="$t('Description')" maxlength="64" size="64" />
              <div v-if="errors.Description" class="text-danger small mt-1">
                {{ errors.Description }}
              </div>
            </div>
          </div>
        </div>
        <div class="col-sm-12 mb-1">
          <div class="form-group row">
            <legend :class="['col-form-label', 'col-sm-2', { required: requiredFields.formData }]">
              {{ $t('FormData') }}:
              <button v-if="platform !== 'Builtin'" type="button" class="btn btn-outline-info btn-sm ms-2" @click="handleSyncFeishu" :disabled="!code">
                <i class="bi bi-arrow-repeat"></i> {{ $t('Sync') }}
              </button>
            </legend>
            <div class="col-sm-6">
              <textarea class="form-control form-control-sm" v-model="formData" v-bind="formDataAttrs" name="formData"
                :placeholder="$t('FormData')" maxlength="20000" rows="5"></textarea>
              <div v-if="errors.FormData" class="text-danger small mt-1">
                {{ errors.FormData }}
              </div>
            </div>
          </div>
        </div>
        <div class="col-sm-12 mb-1">
          <div class="form-group row">
            <legend :class="['col-form-label', 'col-sm-2', { required: requiredFields.nodelist }]">
              {{ $t('NodeList') }}:
            </legend>
            <div class="col-sm-6">
              <textarea class="form-control form-control-sm" v-model="nodeList" v-bind="nodeListAttrs" name="nodeList"
                :placeholder="$t('NodeList')" maxlength="255" rows="3"></textarea>
              <div v-if="errors.NodeList" class="text-danger small mt-1">
                {{ errors.NodeList }}
              </div>
            </div>
          </div>
        </div>
        <div class="col-sm-12">
          <div class="form-group row">
            <legend :class="['col-form-label', 'col-sm-2', { required: requiredFields.platform }]">
              {{ $t('Platform') }}:
            </legend>
            <div class="col-sm-3">
              <VSelect v-model="platform" v-bind="platformAttrs" name="platform"
                :reduce="option => option.value" :placeholder="$t('Please Select')" :options="platformOptions">
              </VSelect>
              <div v-if="errors.Platform" class="text-danger small mt-1">
                {{ errors.Platform }}
              </div>
            </div>
          </div>
        </div>
        <div class="col-sm-12">
          <div class="form-group row">
            <legend :class="['col-form-label', 'col-sm-2', { required: requiredFields.status }]">
              {{ $t('Status') }}:
            </legend>
            <div class="col-sm-3">
              <VSelect v-model="status" v-bind="statusAttrs" name="status" :reduce="option => option.value"
                :placeholder="$t('Please Select')" :options="statusOptions"></VSelect>
              <div v-if="errors.Status" class="text-danger small mt-1">
                {{ errors.Status }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="form-group row mt-3 mb-3">
      <div class="col-sm-auto">
        <input v-if="props.dataInfo && props.dataInfo.ID" type="hidden" name="id" :value="props.dataInfo.ID" />
        <button type="button" class="btn btn-outline-primary btn-sm me-2" @click="onSubmit()">
          {{ $t('Submit') }}
        </button>
        <button type="button" class="btn btn-outline-secondary btn-sm" @click="$emit('goIndex')">
          {{ $t('Cancel') }}
        </button>
      </div>
    </div>
  </form>
</template>

<script setup>
import { useFormOptions } from '@/composables/useFormOptions';
import { useForm } from 'vee-validate';
import { computed, watch } from 'vue';
import VSelect from 'vue-select';
import 'vue-select/dist/vue-select.css';
import * as yup from 'yup';

const { statusOptions, platformOptions } = useFormOptions();

// Define props to receive dataInfo from parent component
const props = defineProps({
  dataInfo: {
    type: Object,
    default: () => ({}),
  },
});

// 简化的验证模式 - 直接使用平面结构
const validationSchema = yup.object({
  Name: yup.string().required(),
  Description: yup.string(),
  FormData: yup.string(),
  NodeList: yup.string(),
  Platform: yup.string().required(),
  Status: yup.string().required(),
  Code: yup.string().test('required-if-not-builtin', 'Code is required for external platforms', function (value) {
    const { Platform } = this.parent;
    if (Platform !== 'Builtin') {
      return !!value;
    }
    return true;
  }),
});

// required字段映射
const requiredFields = computed(() => {
  const requiredMap = {};
  Object.keys(validationSchema.fields).forEach(key => {
    const field = validationSchema.fields[key];
    requiredMap[key.toLowerCase()] = field.tests.some(test => test.OPTIONS?.name === 'required');
  });
  return requiredMap;
});

// 表单初始化
const { values, errors, defineField, handleSubmit, setValues } = useForm({
  validationSchema,
  initialValues: props.dataInfo,
});

// 字段定义
const [name, nameAttrs] = defineField('Name');
const [code, codeAttrs] = defineField('Code');
const [description, descriptionAttrs] = defineField('Description');
const [formData, formDataAttrs] = defineField('FormData');
const [nodeList, nodeListAttrs] = defineField('NodeList');
const [platform, platformAttrs] = defineField('Platform');
const [status, statusAttrs] = defineField('Status');

const emit = defineEmits(['submitForm', 'goIndex']);

const onSubmit = handleSubmit(values => {
  emit('submitForm', values);
});

// 监听 dataInfo 变化并更新表单值
watch(
  () => props.dataInfo,
  newDataInfo => {
    if (newDataInfo && Object.keys(newDataInfo).length > 0) {
      setValues(newDataInfo);
    }
  },
  { immediate: true, deep: true }
);

import { syncFeishuDefinition } from '@/api/approval_def';
import { AppToast } from '@/components/toast.js';

const handleSyncFeishu = async () => {
  if (!code.value) {
    AppToast.show({
      message: '请先填写 Code',
      color: 'warning',
    });
    return;
  }

  try {
    const res = await syncFeishuDefinition(code.value);

    // 标准化响应解析: 后端 HandleSuccess 直接返回数据对象，Axios 放在 res.data 中
    if (res.data && res.data.form_data) {
      formData.value = res.data.form_data;
      AppToast.show({
        message: '同步成功',
        color: 'success',
      });
    } else {
       AppToast.show({
        message: '未获取到表单内容',
        color: 'warning',
      });
    }
  } catch (err) {
    console.error(err);
    // request.js usually handles error toast, but we can add specific one
    // AppToast.show({ message: '同步失败', color: 'danger' });
  }
};
</script>

<style scoped>
.required:after {
  content: ' *';
  color: #dc3545;
  font-weight: bold;
}
</style>
