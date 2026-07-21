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
      vertical-compact
      :is-resizable="editable"
      :is-draggable="editable"
      :cols="columnNumber"
      :margin="[0, 0]"
      :responsive="!editable"
      :use-css-transforms="false"
      class="flex-grow-1 d-flex w-100 h-100"
      @layout-updated="onLayoutUpdated"
    >
      <template
        v-for="(item, index) in layout"
      >
        <grid-item
          v-if="blocks[item.i] && !blocks[item.i].meta.hidden && (!blocks[item.i].meta.invisible || editable)"
          :key="item.i"
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
import { ref, computed, watch, nextTick, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { GridLayout, GridItem } from '../../lib/vue-grid-layout'

const { t } = useI18n()

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

watch(() => props.blocks, (blocks) => {
  const l = blocks.map(({ meta = {}, xywh = [] }, i) => {
    const [x = 0, y = 0, w = 48, h = 15] = (xywh || []).map(v => Number(v) || 0)
    if (meta.hidden || (meta.invisible && !props.editable)) return null
    return { i, x, y, w, h }
  }).filter(b => b)

  layout.value = l
  nextTick(() => {
    forceRerender()
  })
}, { immediate: true, deep: true })

onBeforeUnmount(() => {
  layout.value = []
  resizing.value = false
})

function onLayoutUpdated () {
  if (!props.editable) return

  resizing.value = false

  layout.value.forEach(({ i, x, y, w, h }) => {
    const layoutXYWH = [x, y, w, h]
    const { xywh = [] } = props.blocks[i] || {}
    if (xywh.toString() === layoutXYWH.toString()) return
    emit('item-updated', i)
    props.blocks[i].xywh = layoutXYWH
  })
}

function onGridAction () {
  if (!resizing.value) {
    resizing.value = true
  }
}

function forceRerender () {
  window.dispatchEvent(new Event('resize'))
}
</script>

<style lang="scss">
.vue-grid-item.vue-grid-placeholder {
  background: var(--primary) !important;
}
</style>
