<template>
  <div class="test-app">
    <h1>Workflow-Vue 兼容性测试</h1>
    <WorkflowBuilder />
  </div>
</template>

<script setup>
  import WorkflowBuilder from './components/WorkflowBuilder.vue';
</script>

<style scoped>
  .test-app {
    padding: 20px;
  }
</style>
