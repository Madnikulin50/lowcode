<template>
  <div
    :class="['preview-status', inline ? 'inline' : '', error ? 'doc-err' : '']"
    :style="previewStyle"
    @click.stop="onClick"
  >
    <div
      v-if="error"
      class="preview-overlay"
    >
      <p class="err-message mb-0">
        {{ error.message || error }}
      </p>
    </div>
    <div
      v-else-if="loading"
      class="preview-overlay"
    >
      <p class="d-flex align-items-center gap-1 mb-0">
        <span class="spinner-border spinner-border-sm" />
        {{ loadingLabel }}
      </p>
    </div>
    <div
      class="preview-slot"
      :class="{ 'is-busy': loading || error }"
    >
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  inline?: boolean
  loading?: boolean
  error?: Error | string | null
  loadingLabel?: string
  previewStyle?: Record<string, any>
}>()

const emit = defineEmits<{
  (e: 'retry'): void
  (e: 'openPreview'): void
}>()

function onClick () {
  emit('openPreview')
}
</script>

<style lang="scss" scoped>
.preview-status {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 80px;
  width: 100%;
  height: 100%;
  background-color: var(--white);

  &.inline {
    cursor: zoom-in;
    min-height: 120px;
  }

  &.doc-err .err-message {
    color: var(--danger);
    text-align: center;
    padding: 1rem;
  }
}

.preview-overlay {
  position: absolute;
  inset: 0;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--white);
}

.preview-slot {
  width: 100%;
  height: 100%;
  min-height: inherit;

  &.is-busy {
    visibility: hidden;
    pointer-events: none;
  }
}
</style>
