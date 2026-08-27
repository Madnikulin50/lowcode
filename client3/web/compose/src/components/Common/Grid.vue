<template>
  <div
    v-if="layout.length"
    class="w-100"
    :class="{
      'editable': editable,
      'flex-grow-1 d-flex': isStretchable,
    }"
  >
    <grid-layout
      :layout="layout"
      :col-num="48"
      :row-height="10"
      :vertical-compact="true"
      :is-resizable="editable"
      :is-draggable="editable"
      :cols="columnNumber"
      :margin="[0, 0]"
      :responsive="false"
      :use-css-transforms="false"
      class="flex-grow-1 d-flex w-100 h-100"
      @layout-updated="onLayoutUpdated"
    >
      <template
        v-for="(item, index) in layout"
        :key="item.i"
      >
        <grid-item
          v-if="blocks[item.i] && !blocks[item.i].meta.hidden && (!blocks[item.i].meta.invisible || editable)"
          :i="item.i"
          :h="item.h"
          :w="item.w"
          :x="item.x"
          :y="item.y"
          :min-w="6"
          :min-h="5"
          :class="{ 'h-100': isStretchable }"
          :style="{ 'touch-action': editable ? 'none' : 'auto' }"
          class="grid-item"
          @move="onGridAction"
          @resize="onGridAction"
          @moved="onGridSettled"
          @resized="onGridSettled"
        >
          <slot
            v-if="!blocks[item.i].meta.invisible"
            :block="blocks[item.i]"
            :index="index"
            :block-index="item.i"
            :resizing="resizing"
            :loading-record="loadingRecord"
          />
        </grid-item>
      </template>
    </grid-layout>
  </div>

  <div
    v-else
    class="no-builder-grid h-100 pt-5 container text-center"
  >
    <h4>
      {{ $t('noBlock') }}
    </h4>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'page' } })
import { shallowRef, ref, computed, watch, onBeforeUnmount, nextTick } from 'vue'
import { GridLayout, GridItem } from '../../lib/vue-grid-layout'
import { normalizeXYWH, xywhSignature } from '../../lib/block-layout'

const props = defineProps({
  blocks: {
    type: Array,
    default: () => ([]),
  },
  editable: {
    type: Boolean,
  },
  loadingRecord: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['item-updated'])

const layout = shallowRef([])
const resizing = ref(false)
let ignoreGeometryRebuild = false
let compactSynced = false
let lastEmittedLayout = []

const oneBlockLayout = computed(() => {
  return props.blocks.filter(({ meta }) => !meta.hidden && (!meta.invisible || props.editable)).length === 1
})

const isStretchable = computed(() => {
  return !props.editable && oneBlockLayout.value
})

const columnNumber = computed(() => {
  if (oneBlockLayout.value) {
    return { lg: 1, md: 1, sm: 1, xs: 1, xxs: 1 }
  }
  return { lg: 48, md: 48, sm: 1, xs: 1, xxs: 1 }
})

const geometryKey = computed(() => xywhSignature(props.blocks))

function rebuildLayout () {
  compactSynced = false
  lastEmittedLayout = []
  layout.value = (props.blocks || []).map((block, i) => {
    const meta = block.meta || {}
    if (meta.hidden || (meta.invisible && !props.editable)) return null
    const [x, y, w, h] = normalizeXYWH(block.xywh)
    return { i, x, y, w, h }
  }).filter(Boolean)
}

watch(geometryKey, () => {
  if (resizing.value || ignoreGeometryRebuild) return
  rebuildLayout()
}, { immediate: true })

onBeforeUnmount(() => {
  window.removeEventListener('pointerup', onGridSettled)
  layout.value = []
  resizing.value = false
})

function layoutItemsEqual (a, b) {
  if (!Array.isArray(a) || !Array.isArray(b) || a.length !== b.length) return false
  return a.every((item, idx) => {
    const o = b[idx]
    return !!item && !!o && item.i === o.i && item.x === o.x && item.y === o.y && item.w === o.w && item.h === o.h
  })
}

function toLayoutItems (src) {
  return (src || []).map(({ i, x, y, w, h }) => ({
    i,
    x: Number(x) || 0,
    y: Number(y) || 0,
    w: Number(w) || 1,
    h: Number(h) || 1,
  }))
}

function persistLayout (newLayout) {
  if (!props.editable || !Array.isArray(newLayout)) return

  ignoreGeometryRebuild = true
  newLayout.forEach(({ i, x, y, w, h }) => {
    const next = normalizeXYWH([x, y, w, h])
    const block = props.blocks[i]
    if (!block) return
    if (normalizeXYWH(block.xywh).toString() === next.toString()) return
    emit('item-updated', i)
    block.xywh = next
  })
  nextTick(() => { ignoreGeometryRebuild = false })
}

function onLayoutUpdated (newLayout) {
  if (!Array.isArray(newLayout)) return
  const next = toLayoutItems(newLayout)
  lastEmittedLayout = next
  if (resizing.value) return
  if (layoutItemsEqual(layout.value, next)) {
    compactSynced = true
    return
  }
  // GridLayout compact/emits on every prop change. Writing back re-enters its
  // deep layout watcher and remounts page-block cards until Vue aborts.
  if (compactSynced) return
  compactSynced = true
  layout.value = next
}

function onGridAction () {
  if (resizing.value) return
  resizing.value = true
  window.addEventListener('pointerup', onGridSettled, { once: true })
}

function onGridSettled () {
  window.removeEventListener('pointerup', onGridSettled)
  const next = lastEmittedLayout.length ? lastEmittedLayout : layout.value
  if (!layoutItemsEqual(layout.value, next)) layout.value = next
  persistLayout(next)
  resizing.value = false
}
</script>

<style lang="scss">
.vue-grid-item.vue-grid-placeholder {
  background: var(--primary) !important;
}

.vue-grid-item.grid-item {
  box-sizing: border-box;

  > *:not(.vue-resizable-handle) {
    height: 100%;
    max-height: 100%;
  }
}

.vue-grid-item > .vue-resizable-handle {
  z-index: 20;
}
</style>
