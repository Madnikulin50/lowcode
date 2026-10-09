<template>
  <div class="d-flex flex-column">
    <h5 class="fw-bold text-break">
      {{ title }}
    </h5>

    <div class="text-secondary mb-1 text-break">
      {{ description }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  notification: Record<string, any>
}>()

const title = computed(() => {
  if (!props.notification || !props.notification.config) {
    return ''
  }

  return props.notification.config.title
})

const description = computed(() => {
  if (!props.notification || !props.notification.config) {
    return ''
  }

  return props.notification.config.description || ''
})
</script>
