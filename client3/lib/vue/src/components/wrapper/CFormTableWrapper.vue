<template>
  <div class="list-background w-100 p-3 rounded border border-light">
    <div
      v-if="loading"
      class="d-flex justify-content-center align-items-center w-100"
    >
      <div class="spinner-border" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
    </div>

    <slot
      v-else
    />

    <button
      v-if="!hideAddButton && !loading"
      class="btn btn-primary btn-sm"
      :data-test-id="testID"
      :class="addButtonClass"
      :disabled="disableAddButton"
      @click="$emit('add-item')"
    >
      <font-awesome-icon
        :icon="['fas', 'plus']"
        class="mr-1"
      />

      {{ labels.addButton || '' }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  labels?: Record<string, string>
  hideAddButton?: boolean
  disableAddButton?: boolean
  addButtonClass?: string
  buttonTestId?: string
  loading?: boolean
}>(), {
  labels: () => ({}),
  hideAddButton: false,
  disableAddButton: false,
  addButtonClass: '',
  buttonTestId: '',
  loading: false,
})

defineEmits<{
  (e: 'add-item'): void
}>()

const testID = computed(() => props.buttonTestId)
</script>

<style scoped>
.list-background {
  background-color: var(--body-bg);
}
</style>
