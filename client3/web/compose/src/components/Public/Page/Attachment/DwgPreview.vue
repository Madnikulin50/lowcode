<template>
  <div
    class="dwg-lightbox"
    @click.self="emit('close')"
  >
    <header class="dwg-lightbox-bar">
      <span class="dwg-lightbox-name">{{ name }}</span>
      <a
        v-if="download"
        :href="download"
      >Скачать</a>
      <button
        type="button"
        class="btn btn-sm btn-light"
        @click="emit('close')"
      >
        Закрыть
      </button>
    </header>
    <div
      ref="host"
      class="dwg-lightbox-canvas"
    />
    <p
      v-if="error"
      class="dwg-lightbox-error"
    >
      {{ error }}
    </p>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'

const props = defineProps({
  src: { type: String, default: '' },
  name: { type: String, default: '' },
  download: { type: String, default: '' },
  meta: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['close'])

const host = ref(null)
const error = ref('')
let viewer = null

const previewSrc = computed(() => {
  const raw = String(props.src || '')
  if (!raw || raw.includes('as=dxf')) return raw
  return raw + (raw.includes('?') ? '&' : '?') + 'as=dxf'
})

function destroyViewer () {
  try {
    viewer?.Destroy?.()
  } catch {
    // viewer may already be gone
  }
  viewer = null
  if (host.value) host.value.innerHTML = ''
}

onMounted(async () => {
  if (!previewSrc.value) {
    error.value = 'Нет адреса файла'
    return
  }
  try {
    await nextTick()
    if (!host.value) throw new Error('Не удалось открыть область просмотра')
    const { DxfViewer } = await import('dxf-viewer')
    viewer = new DxfViewer(host.value, {
      autoResize: true,
      colorCorrection: true,
      canvasWidth: host.value.clientWidth || 800,
      canvasHeight: host.value.clientHeight || 600,
    })
    await viewer.Load?.({ url: previewSrc.value, fonts: null })
    const bounds = viewer.GetBounds?.()
    if (bounds) viewer.FitView?.(bounds.minX, bounds.maxX, bounds.minY, bounds.maxY)
  } catch (err) {
    error.value = err?.message || 'Не удалось показать чертёж'
  }
})

onBeforeUnmount(destroyViewer)
</script>

<style scoped>
.dwg-lightbox {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  flex-direction: column;
  background: #111;
  color: #fff;
}

.dwg-lightbox-bar {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.6rem 1rem;
  background: #1c1c1c;
}

.dwg-lightbox-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dwg-lightbox-canvas {
  flex: 1;
  min-height: 0;
}

.dwg-lightbox-error {
  margin: 0;
  padding: 0.75rem 1rem;
  background: #3a1d1d;
}
</style>
