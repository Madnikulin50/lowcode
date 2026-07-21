<template>
  <li class="d-flex mt-1 mx-1">
    <button
      class="btn btn-primary flex-grow-1"
      :disabled="!hasPrevPage"
      @click="emit('prev')"
    >
      <font-awesome-icon
        :icon="['fas', 'chevron-left']"
      />
    </button>
    <button
      class="btn btn-primary flex-grow-1 ms-1"
      :disabled="!hasNextPage"
      @click="emit('next')"
    >
      <font-awesome-icon
        :icon="['fas', 'chevron-right']"
      />
    </button>
  </li>
</template>

<script setup lang="ts">
// This file is copied over from corteza-webapp-compose; src/components/ModuleFields/Common/Pagination.vue
// Refactor it out when those components get moved into corteza-vue

defineProps<{
  hasPrevPage: boolean
  hasNextPage: boolean
}>()

const emit = defineEmits<{
  (e: 'prev'): void
  (e: 'next'): void
}>()
</script>
