<template>
  <div class="mt-3" style="min-height: calc(100vh - 105px); font-size: 0.9rem">
    <div class="border-bottom fs-5 mb-2 pb-2">{{ $t('Approval Define') }} {{ $t('View') }}</div>
    <div class="card-body overlay-wrapper">
      <div id="create_wrapper">
        <ul
          class="nav nav-tabs"
          id="myTab"
          role="tablist"
        >
          <li
            class="nav-item"
            role="presentation"
          >
            <a
              class="nav-link active"
              id="baseinfo-tab"
              data-toggle="tab"
            >
              {{ $t('Basic Info') }}
            </a>
          </li>
        </ul>
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
                <div class="col-sm-12">
                  <div class="form-group row h-100">
                    <legend
                      for="ID"
                      class="col-form-label col-sm-2 my-auto"
                    >
                      {{ $t('ID') }}:
                    </legend>
                    <div class="col-sm-auto my-auto">
                      {{ dataInfo.ID }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-12">
                  <div class="form-group row h-100">
                    <legend
                      for="Code"
                      class="col-form-label col-sm-2 my-auto"
                    >
                      {{ $t('Code') }}:
                    </legend>
                    <div class="col-sm-auto my-auto">
                      {{ dataInfo.Code }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-12">
                  <div class="form-group row h-100">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-2 my-auto"
                    >
                      {{ $t('Name') }}:
                    </legend>
                    <div class="col-sm-auto my-auto">
                      {{ dataInfo.Name }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-12">
                  <div class="form-group row h-100">
                    <legend
                      for="Status"
                      class="col-form-label col-sm-2 my-auto"
                    >
                      {{ $t('Status') }}:
                    </legend>
                    <div class="col-sm-auto my-auto">
                      {{ dataInfo.Status }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-12">
                  <div class="form-group row h-100">
                    <legend
                      for="Description"
                      class="col-form-label col-sm-2 my-auto"
                    >
                      {{ $t('Description') }}:
                    </legend>
                    <div class="col-sm-auto my-auto">
                      {{ dataInfo.Description }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-12">
                  <div class="form-group row h-100">
                    <legend
                      for="FormData"
                      class="col-form-label col-sm-2 my-auto"
                    >
                      {{ $t('FormData') }}:
                    </legend>
                    <div class="col-sm-auto my-auto">
                      {{ dataInfo.FormData }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-12">
                  <div class="form-group row h-100">
                    <legend
                      for="NodeList"
                      class="col-form-label col-sm-2 my-auto"
                    >
                      {{ $t('Node List') }}:
                    </legend>
                    <div class="col-sm-auto my-auto">
                      {{ dataInfo.NodeList }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-12">
                  <div class="form-group row h-100">
                    <legend
                      for="CreatedBy"
                      class="col-form-label col-sm-2 my-auto"
                    >
                      {{ $t('Created By') }}:
                    </legend>
                    <div class="col-sm-auto my-auto">
                      {{ dataInfo.CreatedBy }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-12">
                  <div class="form-group row h-100">
                    <legend
                      for="UpdatedBy"
                      class="col-form-label col-sm-2 my-auto"
                    >
                      {{ $t('Updated By') }}:
                    </legend>
                    <div class="col-sm-auto my-auto">
                      {{ dataInfo.UpdatedBy }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-12">
                  <div class="form-group row h-100">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-2 my-auto"
                    >
                      {{ $t('Updated At') }}:
                    </legend>
                    <div class="col-sm-auto my-auto">
                      {{ dataInfo.UpdatedAt }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-12">
                  <div class="form-group row h-100">
                    <legend
                      for="Name"
                      class="col-form-label col-sm-2 my-auto"
                    >
                      {{ $t('Created At') }}:
                    </legend>
                    <div class="col-sm-auto my-auto">
                      {{ dataInfo.CreatedAt }}
                    </div>
                  </div>
                </div>
                <div class="col-sm-12 my-3">
                  <div class="form-group row">
                    <div class="col-sm-10">
                      <button
                        class="btn btn-outline-secondary btn-sm"
                        type="button"
                        @click="goIndex"
                      >
                        {{ $t('Back') }}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { getApprovalDef } from '@/api/approval_def';
  import { onMounted, ref } from 'vue';
  import { useRouter } from 'vue-router';

  const dataInfo = ref({});
  const router = useRouter();

  onMounted(() => {
    getApprovalDefInfo();
  });

  // const onReset = () => {
  //   searchInfo.value = {}
  // }

  // 查询
  const getApprovalDefInfo = async () => {
    const params = router.currentRoute.value.query;
    const res = await getApprovalDef({
      id: params.id,
    });
    dataInfo.value = res.data;
  };

  function goIndex() {
    router.push('/admin/approval_def/index');
  }
</script>

<style></style>
