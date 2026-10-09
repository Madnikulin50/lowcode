<template>
  <div class="container-fluid p-0">
    <div class="row">
      <div class="col-3">
        <div class="list-group">
          <draggable
            item-key="elementID"
            :list="localItems"
            :disabled="!draggable"
            group="items"
            handle=".grab"
            @change="emitUpdate"
          >
            <template #item="{ element: item, index }">
              <button
                class="list-group-item list-group-item-action d-flex align-items-center justify-content-between"
                :class="{ active: currentIndex !== undefined ? currentIndex === index : index === 0 }"
                @click="$emit('select', index)"
              >
                <slot name="label" :item="item" />
              </button>
            </template>
          </draggable>
          <button
            class="list-group-item list-group-item-action text-primary rounded-top"
            :class="{ 'border-top-0': localItems.length }"
            @click="$emit('add')"
          >
            <font-awesome-icon :icon="['fas', 'plus']" size="sm" class="me-1" />
            {{ t('label.add') }}
          </button>
        </div>
      </div>
      <div v-if="currentIndex !== undefined" class="col-9">
        <slot name="configurator" />
        <c-input-confirm
          variant="danger"
          size="lg"
          size-confirm="lg"
          :borderless="false"
          :text="t('label.delete')"
          class="d-flex"
          @confirmed="$emit('delete')"
        />
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Draggable from 'vuedraggable'

const props = defineProps({
  items: { type: Array, required: true },
  currentIndex: { type: Number, default: undefined },
  draggable: { type: Boolean, default: false },
})
defineEmits(['select', 'add', 'delete', 'update:items'])

const { t } = useI18n()
const localItems = ref([...props.items])

watch(() => props.items, (val) => { localItems.value = [...val] }, { deep: true })

function emitUpdate() {
  // vuedraggable v4 with v-model:list pattern
}
</script>
<style lang="scss">
.item {
  .grab { opacity: 0; }
  &:hover .grab { opacity: 1; }
}
</style>