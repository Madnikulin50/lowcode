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
      :key="gridKey"
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
import { ref, computed, watch, onBeforeUnmount } from 'vue'
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

const layout = ref([])
const resizing = ref(false)
const pendingPersist = ref(false)

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

const gridKey = computed(() => {
  return layout.value.map(b => b.i).join(',')
})

const geometryKey = computed(() => xywhSignature(props.blocks))

function rebuildLayout () {
  layout.value = (props.blocks || []).map((block, i) => {
    const meta = block.meta || {}
    if (meta.hidden || (meta.invisible && !props.editable)) return null
    const [x, y, w, h] = normalizeXYWH(block.xywh)
    return { i, x, y, w, h }
  }).filter(Boolean)
}

watch(geometryKey, () => {
  if (resizing.value || pendingPersist.value) return
  rebuildLayout()
}, { immediate: true })

onBeforeUnmount(() => {
  layout.value = []
  resizing.value = false
  pendingPersist.value = false
})

function persistLayout (newLayout) {
  if (!props.editable || !Array.isArray(newLayout)) return

  layout.value = newLayout
  newLayout.forEach(({ i, x, y, w, h }) => {
    const next = normalizeXYWH([x, y, w, h])
    const block = props.blocks[i]
    if (!block) return
    if (normalizeXYWH(block.xywh).toString() === next.toString()) return
    emit('item-updated', i)
    block.xywh = next
  })
}

function onLayoutUpdated (newLayout) {
  if (!props.editable) return
  if (!pendingPersist.value) return

  pendingPersist.value = false
  resizing.value = false
  persistLayout(newLayout)
}

function onGridAction () {
  resizing.value = true
}

function onGridSettled () {
  pendingPersist.value = true
}
</script>

<style lang="scss">
.vue-grid-item.vue-grid-placeholder {
  background: var(--primary) !important;
}

.vue-grid-item.grid-item {
  box-sizing: border-box;

  > * {
    height: 100%;
    max-height: 100%;
  }
}
</style>
