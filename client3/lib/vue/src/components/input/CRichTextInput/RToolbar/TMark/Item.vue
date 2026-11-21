<template>
  <button
    class="btn btn-link text-dark fw-bold text-decoration-none"
    @click="onClick(format.type, format.attrs)"
  >
    <span :class="activeClasses(format.attrs)">
      <font-awesome-icon
        v-if="format.icon"
        :icon="format.icon"
      />

      <span v-else>
        {{ format.label }}
      </span>
    </span>
  </button>
</template>

<script setup lang="ts">
defineOptions({ inheritAttrs: false })

const props = defineProps<{
  editor: any
  format: any
  isActive?: any
  getMarkAttrs?: (...args: any[]) => any
  currentValue?: string
}>()

const emit = defineEmits<{
  (e: 'click', payload: { type: string; attrs: Record<string, any> }): void
}>()

function onClick(type: string, attrs: Record<string, any>) {
  emit('click', { type, attrs })
}

function activeClasses(attrs?: Record<string, any>) {
  const isActive = props.editor.isActive(props.format.type, attrs)
  if (isActive) {
    return ['text-primary']
  }
  return undefined
}
</script>
