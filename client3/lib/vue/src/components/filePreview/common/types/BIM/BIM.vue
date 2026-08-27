<template>
  <div
    :class="['bim-preview-host', inline ? 'inline' : '', attrs.onClick ? 'clickable' : '']"
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
        class="bim-canvas"
      />
    </PreviewStatus>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, useAttrs } from 'vue'
import PreviewStatus from '../PreviewStatus.vue'
import { BIM_MAX_BYTES, assertPreviewSize, fetchBinary } from '../../binary.js'
import { getFileExt } from '../../index.js'

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

let components: any = null
let world: any = null

const src = computed(() => (attrs as any).src)
const labels = computed(() => (attrs as any).labels || {})
const previewStyle = computed(() => (attrs as any).previewStyle || {})
const inline = computed(() => (attrs as any).inline ?? props.inline)
const ext = computed(() => getFileExt({
  src: src.value,
  name: (attrs as any).name,
  meta: (attrs as any).meta,
}))

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
  await destroyViewer()
  try {
    assertPreviewSize((attrs as any).meta, BIM_MAX_BYTES, labels.value)
    await nextTick()
    if (!host.value) {
      throw new Error(labels.value.loadFailed || 'Failed to load preview')
    }

    const OBC = await import('@thatopen/components')
    components = new OBC.Components()
    const worlds = components.get(OBC.Worlds)
    world = worlds.create()
    world.scene = new OBC.SimpleScene(components)
    world.renderer = new OBC.SimpleRenderer(components, host.value)
    world.camera = new OBC.SimpleCamera(components)
    components.init()
    world.scene.setup?.()
    await world.camera.controls?.setLookAt?.(12, 6, 8, 0, 0, 0)

    if (inline.value && world.camera.controls) {
      world.camera.controls.enabled = false
    }

    const fragments = components.get(OBC.FragmentsManager)
    fragments.init(await OBC.FragmentsManager.getWorker())

    const buffer = new Uint8Array(await fetchBinary(src.value))

    if (ext.value === 'frag') {
      const model = await fragments.core.load(buffer, { modelId: 'preview' })
      if (model?.object) {
        world.scene.three.add(model.object)
      }
    } else {
      const ifcLoader = components.get(OBC.IfcLoader)
      await ifcLoader.setup({
        autoSetWasm: true,
      })
      const model = await ifcLoader.load(buffer, true, 'preview')
      if (model?.object) {
        world.scene.three.add(model.object)
      } else if (model?.three) {
        world.scene.three.add(model.three)
      }
    }

    try {
      await world.camera.controls?.fitToSphere?.(world.scene.three, true)
    } catch {}
  } catch (err: any) {
    loadError.value = err instanceof Error ? err : new Error(String(err?.message || err))
    emit('error', loadError.value)
  } finally {
    loading.value = false
  }
}

async function destroyViewer () {
  try {
    world?.renderer?.dispose?.()
    components?.dispose?.()
  } catch {}
  components = null
  world = null
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
.bim-preview-host {
  width: 100%;
  height: 100%;
  min-height: 280px;
  background: #111;

  &.inline {
    cursor: zoom-in;
    min-height: 160px;
    max-height: 240px;
    .bim-canvas {
      pointer-events: none;
    }
  }
}

.bim-canvas {
  width: 100%;
  height: 100%;
  min-height: 280px;
  position: relative;
}
</style>
