<template>
  <div
    class="d-flex align-items-center justify-content-center h-100 mb-2 file-hint"
    :class="{ inline: !!inline }"
    @click.stop="onClick"
  >
    <font-awesome-icon
      v-if="inline"
      :title="name"
      :icon="['far', `file-${icon}`]"
      :style="previewStyle"
      class="inline-icon d-block text-secondary"
    />

    <div
      v-else
      class="text-secondary text-center px-4"
      style="max-width: 36rem;"
    >
      <p class="mb-2">
        {{ message }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, useAttrs } from 'vue'
import { getExtensionIconType, getFileExt, hintKind } from '../index.js'

defineOptions({ inheritAttrs: false })

const attrs = useAttrs()

const props = defineProps<{
  inline?: boolean
  previewStyle?: Record<string, any>
  labels?: Record<string, any>
}>()

const emit = defineEmits<{
  (e: 'openPreview'): void
}>()

const labels = computed(() => props.labels || (attrs as any).labels || {})
const meta = computed(() => (attrs as any).meta || {})
const name = computed(() => (attrs as any).name || '')
const ext = computed(() => getFileExt({
  src: (attrs as any).src,
  name: name.value,
  meta: meta.value,
}))
const icon = computed(() => getExtensionIconType(ext.value))
const kind = computed(() => hintKind(ext.value))

const message = computed(() => {
  const l = labels.value
  if (kind.value === 'dwg') {
    return l.hintDwg || 'Native DWG drawings cannot be previewed in the browser yet. Download the file, or export DXF from CAD software and attach that instead.'
  }
  if (kind.value === 'archicad') {
    return l.hintArchicad || 'Native ArchiCAD projects cannot be opened in the browser. In ArchiCAD use File → Save as → IFC and attach that file.'
  }
  if (kind.value === 'bimx') {
    return l.hintBimx || 'Open this Hyper-model in the BIMx app, or export IFC from ArchiCAD for in-app 3D preview.'
  }
  return l.previewUnavailable || 'Preview unavailable'
})

function onClick () {
  if (props.inline || (attrs as any).inline) {
    emit('openPreview')
  }
}
</script>

<style lang="scss" scoped>
.inline-icon {
  cursor: zoom-in;
}
.file-hint.inline {
  cursor: zoom-in;
}
</style>
