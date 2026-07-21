<template>
  <div
    v-if="tooltip"
    data-bs-toggle="tooltip"
    data-bs-placement="right"
    data-bs-trigger="hover"
    :title="tooltip"
    class="d-flex align-items-center ms-1"
  >
    <font-awesome-icon
      :icon="icon"
      :class="iconClass"
    />
  </div>
</template>

<script setup lang="ts">
defineProps({
  icon: {
    type: Array,
    default: () => ['far', 'question-circle'],
  },
  iconClass: {
    type: String,
    default: 'text-primary',
  },
  tooltip: {
    type: String,
    default: '',
  },
  tooltipVariant: {
    type: String,
    default: 'dark',
  },
})
</script>
