<template>
  <div
    :data-test-id="title || `${operation} on ${resource}`"
  >
    <p
      :title="title || `${operation} on ${resource}`"
      class="mb-1 text-truncate"
    >
      {{ title || `${operation} on ${resource}` }}
    </p>

    <Access
      :access="access"
      :current="current"
      :enabled="enabled"
      class="w-100"
      @update="onUpdate"
    />
  </div>
</template>

<script setup lang="ts">
import Access from './Access.vue'

const props = defineProps<{
  resource: string
  operation: string
  title?: string
  description?: string
  enabled?: boolean
  access?: string
  current?: string
}>()

const emit = defineEmits<{
  (e: 'update', value: { resource: string; operation: string; access: string }): void
}>()

function onUpdate(access: string) {
  emitValue(access)
}

function emitValue(access: string) {
  emit('update', {
    resource: props.resource,
    operation: props.operation,
    access,
  })
}
</script>
