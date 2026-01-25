<template>
  <div class="mt-3">
    <div class="border-bottom fs-5 mb-2 pb-2">{{ $t('Approval Define') }} {{ $t('Update') }}</div>
    <ul class="nav nav-tabs mh-30" id="myTab" role="tablist">
      <li class="nav-item mh-30" role="presentation">
        <a class="nav-link mh-30 active" id="baseinfo-tab" data-toggle="tab" href="#baseinfo">
          {{ $t('Basic Info') }}
        </a>
      </li>
    </ul>
    <div class="card-body mt-2" style="min-height: calc(100vh - 200px); font-size: 0.9rem">
      <div id="create_wrapper">
        <div class="row">
          <div class="col-sm-12">
            <app-form :data-info="dataInfo" @submit-form="submitForm" @go-index="goIndex" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import {
  getApprovalDef,
  // createApprovalDef,
  // deleteApprovalDef,
  updateApprovalDef,
} from '@/api/approval_def';

import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
// import AppSpinner from "@/components/Spinner.vue";
import { AppToast } from '@/components/toast.js';
import AppForm from './Form.vue';

// const open = ref(false)
// const showSpinner = ref(false)
const router = useRouter();
const dataInfo = ref({});

const formData = ref({
  showType: 'All',
  startDate: '2023-01-01',
  endDate: '2023-07-30',
});
const rules = ref({
  name: [
    {
      required: true,
      message: '请输入字典名（中）',
      trigger: 'blur',
    },
  ],
  type: [
    {
      required: true,
      message: '请输入字典名（英）',
      trigger: 'blur',
    },
  ],
  desc: [
    {
      required: true,
      message: '请输入描述',
      trigger: 'blur',
    },
  ],
});

onMounted(() => {
  getApprovalDefInfo();
});

const getApprovalDefInfo = async () => {
  const params = router.currentRoute.value.query;
  const res = await getApprovalDef({
    id: params.id,
  });
  dataInfo.value = res.data;
};

async function submitForm(data) {
  const res = await updateApprovalDef({
    ...data,
  });
  if (res) {
    AppToast.show({
      message: 'Update approval definition success。',
      color: 'success',
    });
    router.push('/admin/approval_def/index');
  }
}

function goIndex() {
  router.push('/admin/approval_def/index');
}
</script>

<style scoped></style>
