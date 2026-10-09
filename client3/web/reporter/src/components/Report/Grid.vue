<template>
  <div v-if="layout.length" class="w-100">
    <grid-layout
      :layout="layout"
      :col-num="48"
      :row-height="10"
      vertical-compact
      :is-resizable="editable"
      :is-draggable="editable"
      :cols="{ lg: 48, md: 48, sm: 1, xs: 1, xxs: 1 }"
      :margin="[0, 0]"
      :responsive="!editable"
      :use-css-transforms="false"
      @layout-updated="layout = $event"
    >
      <grid-item
        v-for="(item, index) in grid"
        :key="item.i"
        :min-w="3"
        :min-h="3"
        :i="item.i"
        :h="item.h"
        :w="item.w"
        :x="item.x"
        :y="item.y"
        :class="{ 'editable-grid-item': editable }"
        drag-ignore-from=".gutter"
        @moved="onBlockUpdated(index)"
        @resized="onBlockUpdated(index)"
      >
        <slot :block="blocks[item.i]" :index="index" :block-index="item.i" />
      </grid-item>
    </grid-layout>
  </div>
  <div v-else class="d-flex align-items-center justify-content-center h-50 w-100">
    <h4>{{ t('builder.no-blocks-added') }}</h4>
  </div>
</template>
<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { GridLayout, GridItem } from 'vue-grid-layout-v3'

const props = defineProps({
  blocks: { type: Array, default: () => [] },
  editable: { type: Boolean, default: false },
})
const emit = defineEmits(['update:blocks', 'item-updated'])

const { t } = useI18n()
const grid = ref(undefined)

const layout = computed({
  get: () => grid.value || props.blocks,
  set: (val) => {
    if (props.editable) {
      emit('update.blocks', val)
    } else {
      grid.value = val
    }
  },
})

watch(() => props.blocks, (blocks) => {
  if (props.editable) grid.value = blocks
}, { immediate: true, deep: true })

function onBlockUpdated(index) {
  emit('item-updated', index)
}
</script>
<style lang="scss">
.vue-grid-item.vue-grid-placeholder { background: var(--primary) !important; }
</style>