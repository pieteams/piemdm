<template>
  <div
    ref="modal"
    class="modal fade"
    tabindex="-1"
  >
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">
            选择要添加的节点
          </h5>
          <button
            type="button"
            class="btn-close"
            data-bs-dismiss="modal"
            aria-label="Close"
          />
        </div>
        <div class="modal-body">
          <div class="list-group">
            <button
              v-for="item in nodeTypes"
              :key="item.type"
              type="button"
              class="list-group-item list-group-item-action d-flex align-items-center"
              @click="selectNodeType(item.type, item.name)"
            >
              <i :class="`bi ${item.icon} fs-4 me-3`" />
              <div>
                <h6 class="mb-0">
                  {{ item.name }}
                </h6>
                <small>{{ item.desc }}</small>
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref, onMounted } from 'vue';
  import { Modal } from 'bootstrap';
  import { ADDABLE_NODE_TYPES } from '../constants/node-types.js';

  const emit = defineEmits(['select']);
  let modalInstance = null;
  const modal = ref(null);

  // 使用常量定义可添加的节点类型
  const nodeTypes = ADDABLE_NODE_TYPES;

  onMounted(() => {
    modalInstance = new Modal(modal.value);
  });

  const show = () => modalInstance.show();
  const hide = () => modalInstance.hide();

  const selectNodeType = (type, name) => {
    emit('select', type, name);
    hide();
  };

  // 暴露 show 方法给父组件
  defineExpose({ show });
</script>
