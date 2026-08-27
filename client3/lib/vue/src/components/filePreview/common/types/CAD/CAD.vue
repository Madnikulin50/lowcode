<template>
  <div
    :class="['cad-preview-host', inline ? 'inline' : '', attrs.onClick ? 'clickable' : '']"
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
        class="cad-canvas"
      />
    </PreviewStatus>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, useAttrs } from 'vue'
import PreviewStatus from '../PreviewStatus.vue'
import { CAD_MAX_BYTES, assertPreviewSize } from '../../binary.js'

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
let viewer: { Load?: (opts: any) => Promise<unknown>, Destroy?: () => void, GetBounds?: () => any, FitView?: (...args: number[]) => void } | null = null

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
  destroyViewer()
  try {
    assertPreviewSize((attrs as any).meta, CAD_MAX_BYTES, labels.value)
    await nextTick()
    if (!host.value) {
      throw new Error(labels.value.loadFailed || 'Failed to load preview')
    }
    const { DxfViewer } = await import('dxf-viewer')
    viewer = new DxfViewer(host.value, {
      autoResize: true,
      colorCorrection: true,
      canvasWidth: host.value.clientWidth || 640,
      canvasHeight: inline.value ? 200 : (host.value.clientHeight || 480),
    })
    await viewer.Load?.({
      url: src.value,
      fonts: null,
    })
    const bounds = viewer.GetBounds?.()
    if (bounds) {
      viewer.FitView?.(bounds.minX, bounds.maxX, bounds.minY, bounds.maxY)
    }
  } catch (err: any) {
    loadError.value = err instanceof Error ? err : new Error(String(err))
    emit('error', loadError.value)
  } finally {
    loading.value = false
  }
}

function destroyViewer () {
  try {
    viewer?.Destroy?.()
  } catch {}
  viewer = null
  if (host.value) {
    host.value.innerHTML = ''
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
  destroyViewer()
})
</script>

<style lang="scss" scoped>
.cad-preview-host {
  width: 100%;
  height: 100%;
  min-height: 240px;
  background: #1b1b1b;

  &.inline {
    cursor: zoom-in;
    min-height: 160px;
    max-height: 240px;
    .cad-canvas {
      pointer-events: none;
    }
  }
}

.cad-canvas {
  width: 100%;
  height: 100%;
  min-height: 240px;
}
</style>
