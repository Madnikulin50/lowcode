<template>
  <div
    :class="['office-preview', 'docx-preview-host', inline ? 'inline' : '', attrs.onClick ? 'clickable' : '']"
    :style="previewStyle"
    @click.stop="onPreviewClick"
  >
    <PreviewStatus
      :inline="!!inline"
      :loading="loading"
      :error="loadError"
      :loading-label="labels.loading || 'Loading'"
      @open-preview="onPreviewClick"
    >
      <div
        ref="host"
        class="docx-body"
      />
    </PreviewStatus>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, useAttrs } from 'vue'
import PreviewStatus from '../PreviewStatus.vue'
import { OFFICE_MAX_BYTES, assertPreviewSize, fetchBinary } from '../../binary.js'

defineOptions({ inheritAttrs: false })

const attrs = useAttrs()
const props = defineProps<{
  inline?: boolean
}>()

const emit = defineEmits<{
  (e: 'openPreview'): void
  (e: 'error', err: Error): void
}>()

const host = ref<HTMLElement | null>(null)
const loading = ref(true)
const loadError = ref<Error | null>(null)

const src = computed(() => (attrs as any).src)
const labels = computed(() => (attrs as any).labels || {})
const previewStyle = computed(() => (attrs as any).previewStyle || {})
const inline = computed(() => (attrs as any).inline ?? props.inline)

function onPreviewClick () {
  if (loadError.value) {
    init()
    return
  }
  if (inline.value) {
    emit('openPreview')
  }
}

async function init () {
  loading.value = true
  loadError.value = null
  try {
    assertPreviewSize((attrs as any).meta, OFFICE_MAX_BYTES, labels.value)
    const buffer = await fetchBinary(src.value)
    await nextTick()
    if (!host.value) {
      throw new Error(labels.value.loadFailed || 'Failed to load preview')
    }
    host.value.innerHTML = ''
    const { renderAsync } = await import('docx-preview')
    await renderAsync(buffer, host.value, undefined, {
      inWrapper: !inline.value,
      ignoreWidth: !!inline.value,
      ignoreHeight: !!inline.value,
      breakPages: !inline.value,
      renderHeaders: !inline.value,
      renderFooters: !inline.value,
      useBase64URL: true,
    })
    if (inline.value) {
      const extra = host.value.querySelectorAll('.docx-wrapper > section:nth-of-type(n+2)')
      extra.forEach(node => node.remove())
    }
  } catch (err: any) {
    loadError.value = err instanceof Error ? err : new Error(String(err))
    emit('error', loadError.value)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!src.value) {
    loadError.value = new Error('src.missing')
    loading.value = false
    return
  }
  init()
})

onBeforeUnmount(() => {
  if (host.value) {
    host.value.innerHTML = ''
  }
})
</script>

<style lang="scss">
.docx-preview-host {
  width: 100%;
  background: var(--white);

  &.inline {
    cursor: zoom-in;
    overflow: hidden;
    max-height: 240px;

    .docx-body {
      pointer-events: none;
      transform-origin: top center;
    }
  }

  .docx-body {
    overflow: auto;
    max-height: 100%;
  }

  &:not(.inline) .docx-wrapper {
    background: #525659;
    padding: 1rem 0;
  }
}
</style>
