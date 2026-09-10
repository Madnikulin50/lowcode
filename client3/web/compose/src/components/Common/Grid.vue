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
          :class="{ 'h-100': isStretchable, 'grid-item--compact': isCompactBlock(item.i) }"
          :style="{ 'touch-action': editable ? 'none' : 'auto', '--order': item.y * 1000 + item.x }"
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
let resyncTimer = 0

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

// Blocks small enough to sit two-up (or more, on a wider phone/tablet) on a
// row once the mobile flex-wrap override below takes over from the library's
// px grid (see the `<768px` block in <style>). Everything else defaults to
// full row width — a table, chart or map squeezed to a sliver is unusable —
// so an unlisted/new block kind fails safe toward "wide" rather than
// "compact". Whether a given compact block ends up sharing a row or getting
// the full width to itself (e.g. an odd one out) is left entirely to the
// browser's own flex-wrap + flex-grow — see the CSS, no JS prediction needed.
const compactKinds = new Set(['Metric', 'Progress'])

function isCompactBlock (i) {
  return compactKinds.has(props.blocks[i]?.kind)
}

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
  if (resyncTimer) window.clearTimeout(resyncTimer)
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

function cloneLayoutItems (src) {
  return toLayoutItems(src).map((item) => ({ ...item }))
}

function applyLayout (next) {
  const items = cloneLayoutItems(next)
  if (layoutItemsEqual(layout.value, items)) return items
  layout.value = items
  return items
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
  // Only keep user drag/resize geometry. resizeend recalculates from stale
  // GridItem innerH and would poison this buffer with the pre-resize height.
  if (resizing.value) {
    lastEmittedLayout = next
    return
  }
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
  applyLayout(next)
  persistLayout(next)
  lastEmittedLayout = []
  compactSynced = true
  resizing.value = false
  // vue-grid-layout-v3 resizeend writes stale innerH back into its state after
  // @resized. Re-push the settled geometry on the next tick so GridItem :h stays.
  if (resyncTimer) window.clearTimeout(resyncTimer)
  resyncTimer = window.setTimeout(() => {
    resyncTimer = 0
    compactSynced = true
    // New array so GridLayout's layout watcher re-imports after resizeend rollback.
    layout.value = cloneLayoutItems(next)
  }, 0)
}
</script>

<style lang="scss">
.vue-grid-item.vue-grid-placeholder {
  background: var(--primary) !important;
}

.vue-grid-item {
  box-sizing: border-box;

  > *:not(.vue-resizable-handle) {
    height: 100%;
    max-height: 100%;
  }
}

.vue-grid-item > .vue-resizable-handle {
  z-index: 20;
}

// Below md, the runtime (non-editable) view abandons the library's px grid —
// designed for desktop widths — for a flex-wrap flow: the browser decides how
// many tiles fit per row instead of every block shrinking proportionally to a
// sliver. Flexbox rather than CSS grid specifically because of flex-grow: a
// block left alone on the last line of its row (an odd one out, or simply the
// last block on the page) stretches to fill it instead of sitting at
// half-width with empty space next to it — grid's `auto-fit` tracks don't do
// that without predicting column counts in JS, which breaks the moment a
// wider viewport (tablet) fits 3+ compact blocks per row instead of 2.
// The Page Builder canvas (.editable) is excluded: dragging and resizing need
// the real desktop geometry regardless of how narrow the window you're
// editing from happens to be.
@media (max-width: 767.98px) {
  .w-100:not(.editable) .vue-grid-layout {
    display: flex !important;
    flex-wrap: wrap;
    gap: 12px;
    height: auto !important;
  }

  .w-100:not(.editable) .vue-grid-item {
    position: static !important;
    left: auto !important;
    top: auto !important;
    width: auto !important;
    // Height stays as authored (block.xywh's h, already converted to px by
    // the library) — it's a reasonable signal of how much room a block
    // wants. Only horizontal placement (x/y/w) is handed to the browser.
    flex: 1 1 100%;
    // Keeps desktop reading order (top-to-bottom, left-to-right) now that
    // position is no longer doing that job.
    order: var(--order, 0);

    // A smaller basis so 2+ compact blocks share a row on a phone-width
    // screen; flex-grow (from the shorthand above) still applies, so one
    // left alone on its line — no room for a second, or simply the last
    // block — grows to fill the row rather than sitting at partial width.
    &.grid-item--compact {
      flex-basis: 150px;
    }
  }
}
</style>
