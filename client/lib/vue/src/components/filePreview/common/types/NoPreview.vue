<template>
  <div class="d-flex align-items-center justify-content-center h-100 mb-2">
    <font-awesome-icon
      v-if="inline"
      :title="name"
      :icon="['far', `file-${icon}`]"
      :style="previewStyle"
      class="inline-icon d-block text-secondary"
      @click.stop="$emit('openPreview')"
    />

    <p
      v-else
      class="text-secondary"
    >
      {{ labels.previewUnavailable || 'Preview unavailable' }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, useAttrs } from 'vue'
import { getExtensionIconType } from '../index.js'

defineOptions({ inheritAttrs: false })

const attrs = useAttrs()

defineProps<{
  inline?: boolean
  previewStyle?: Record<string, any>
  labels?: Record<string, any>
}>()

defineEmits<{
  (e: 'openPreview'): void
}>()

const meta = computed(() => (attrs as any).meta || {})
const name = computed(() => (attrs as any).name || '')

const icon = computed(() => {
  const original = meta.value.original || {}
  const ext = original.ext
  return getExtensionIconType(ext)
})
</script>

<style lang="scss" scoped>
.inline-icon {
  cursor: zoom-in;
}
</style>
