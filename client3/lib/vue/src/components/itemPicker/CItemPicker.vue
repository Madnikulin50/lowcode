<template>
  <div class="overflow-hidden d-flex flex-column w-100 vh-100">
    <div class="card h-100">
      <div class="card-header p-0">
        <CInputSearch
          v-if="!hideFilter"
          v-model.trim="query"
          :disabled="_disabledFilter"
          :placeholder="labels.searchPlaceholder"
        />
      </div>

      <div class="card-body d-flex p-0">
        <div class="card h-100 col-sm-6 col-12 p-0 border-0">
          <div class="card-header py-2 ps-0 pe-2 bg-transparent">
            <div class="d-flex align-items-center">
              <label class="text-primary mb-0 py-1">
                {{ labels.availableItems }}
              </label>
              <button
                v-show="filteredAvailable.length && !disabled"
                data-test-id="link-select-all"
                class="btn btn-outline-light ms-auto border-0 text-muted"
                @click="selectAll()"
              >
                {{ labels.selectAllItems }}
              </button>
            </div>
          </div>

          <div class="card-body overflow-auto py-0 ps-0 pe-2">
            <ul class="list-group h-100">
              <draggable
                v-model="filteredAvailable"
                :sort="!_disabledSorting"
                :move="_disableDragging"
                :disabled="!!query"
                draggable=".item"
                group="items"
                item-key="value"
                class="overflow-auto h-100"
              >
                <template #item="{ element }">
                  <li
                    class="list-group-item item mb-3 border rounded"
                    :class="{
                      'handle': !isDraggable
                    }"
                    @dblclick="select(element)"
                  >
                    <CItemPickerItem
                      :item="element"
                      :disabled="disabled"
                      :disabled-dragging="isDraggable"
                      :disabled-sorting="disabledSorting"
                      :hide-icons="hideIcons"
                      @select="select(element)"
                    >
                      <template
                        v-for="(_, slot) of $slots"
                        #[slot]="scope"
                      >
                        <slot
                          :name="slot"
                          :text-field="textField"
                          :disabled="disabled"
                          :disabled-dragging="isDraggable"
                          :disabled-sorting="disabledSorting"
                          :hide-icons="hideIcons"
                          v-bind="scope"
                        />
                      </template>
                    </CItemPickerItem>
                  </li>
                </template>

                <template #footer>
                  <h6
                    v-if="!filteredAvailable.length && query"
                    class="text-center my-4"
                  >
                    {{ labels.noItemsFound }}
                  </h6>
                </template>
              </draggable>
            </ul>
          </div>
        </div>

        <div class="card h-100 ps-sm-0 col-sm-6 col-12 p-0 border-0">
          <div class="card-header py-2 ps-2 pe-0 bg-transparent">
            <div class="d-flex align-items-center">
              <label class="text-primary mb-0 py-1">
                {{ labels.selectedItems }}
              </label>
              <button
                v-show="filteredSelected.length && !disabled"
                data-test-id="link-unselect-all"
                class="btn btn-outline-light ms-auto border-0 text-muted"
                @click="unselectAll()"
              >
                {{ labels.unselectAllItems }}
              </button>
            </div>
          </div>

          <div class="card-body overflow-auto py-0 ps-2 pe-0">
            <ul class="list-group h-100">
              <draggable
                v-model="filteredSelected"
                :sort="!_disabledSorting"
                :move="_disableDragging"
                :disabled="!!query"
                draggable=".item"
                group="items"
                item-key="value"
                class="overflow-auto h-100"
              >
                <template #item="{ element }">
                  <li
                    class="list-group-item item mb-3 border rounded"
                    :class="{
                      'handle': !isDraggable
                    }"
                    @dblclick="unselect(element)"
                  >
                    <CItemPickerItem
                      :item="element"
                      :disabled="disabled"
                      :disabled-dragging="isDraggable"
                      :disabled-sorting="disabledSorting"
                      :hide-icons="hideIcons"
                      selected
                      @unselect="unselect(element)"
                    >
                      <template
                        v-for="(_, slot) of $slots"
                        #[slot]="scope"
                      >
                        <slot
                          :name="slot"
                          :text-field="textField"
                          :disabled="disabled"
                          :disabled-dragging="isDraggable"
                          :disabled-sorting="disabledSorting"
                          :hide-icons="hideIcons"
                          v-bind="scope"
                        />
                      </template>
                    </CItemPickerItem>
                  </li>
                </template>

                <template #footer>
                  <h6
                    v-if="!filteredSelected.length && query"
                    class="text-center my-4"
                  >
                    {{ labels.noItemsFound }}
                  </h6>
                </template>
              </draggable>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import draggable from 'vuedraggable'
import CItemPickerItem from './CItemPickerItem.vue'
import CInputSearch from '../input/CInputSearch.vue'
import { throttle } from 'lodash'

const props = withDefaults(defineProps<{
  options: any[]
  modelValue: any[]
  valueField?: string
  textField?: string
  labels?: Record<string, string>
  disabled?: boolean
  disabledFilter?: boolean
  disabledSorting?: boolean
  disabledDragging?: boolean
  hideIcons?: boolean
  hideFilter?: boolean
}>(), {
  valueField: 'value',
  textField: 'text',
  labels: () => ({
    searchPlaceholder: 'Filter items',
    availableItems: 'Available',
    selectAllItems: 'Select all',
    selectedItems: 'Selected',
    unselectAllItems: 'Unselect all',
    noItemsFound: 'No items found',
  }),
  disabled: false,
  disabledFilter: false,
  disabledSorting: false,
  disabledDragging: false,
  hideIcons: false,
  hideFilter: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: any[]]
}>()

const query = ref('')
const available = ref<any[]>([])
const selected = ref<any[]>([])

const _disabledFilter = computed(() => props.disabled || props.disabledFilter)
const _disabledSorting = computed(() => props.disabled || props.disabledSorting)

const filteredAvailable = computed({
  get: () => {
    const q = query.value.toLowerCase()
    return available.value.filter((i: any) => i[props.textField].toLowerCase().indexOf(q) > -1)
  },
  set: (items: any[]) => {
    available.value = items
  },
})

const filteredSelected = computed({
  get: () => {
    const q = query.value.toLowerCase()
    return selected.value.filter((i: any) => i[props.textField].toLowerCase().indexOf(q) > -1)
  },
  set: (items: any[]) => {
    selected.value = items
  },
})

const isDraggable = computed(() => props.disabledDragging || query.value.length > 0)

watch(selected, (items) => {
  const value = items.map((i: any) => i[props.valueField])
  emit('update:modelValue', value)
})

watch(() => props.options, () => {
  sync()
}, { deep: true, immediate: true })

watch(() => props.modelValue, (value = [], oldValue = []) => {
  if (value.length === oldValue.length && value.filter((v: any) => !oldValue.includes(v)).length === 0) {
    return
  }
  sync()
})

// Validate options in created
props.options.forEach((o: any) => {
  if (typeof o !== 'object') {
    throw new Error('expecting array of objects for options prop')
  }
})

onBeforeUnmount(() => {
  setDefaultValues()
})

function _disableDragging(e: any) {
  if (props.disabledDragging && e.to !== e.from) {
    return false
  }
}

function selectAll() {
  selected.value.push(...filteredAvailable.value)
  available.value = frozen().filter((i: any) => !selected.value.includes(i))
}

const select = throttle((item: any) => {
  available.value = available.value.filter((i: any) => i !== item)
  if (props.disabledSorting) {
    selected.value = frozen().filter((i: any) => !available.value.includes(i))
  } else {
    if (!selected.value.some(({ value = '' }: any) => value === item.value)) {
      selected.value.push(item)
    }
  }
}, 300)

function unselectAll() {
  filteredSelected.value.forEach((item: any) => unselect(item))
}

function unselect(item: any) {
  selected.value = selected.value.filter((i: any) => i !== item)
  available.value = frozen().filter((i: any) => !selected.value.includes(i))
}

function isPicked(item: any) {
  return item[props.valueField] && props.modelValue.includes(item[props.valueField])
}

function sync() {
  available.value = frozen().filter((opt: any) => !isPicked(opt))
  selected.value = props.modelValue.map((v: any) => {
    return frozen().find((item: any) => item[props.valueField] === v)
  }).filter((f: any) => f)
}

function frozen() {
  return props.options.map(Object.freeze)
}

function setDefaultValues() {
  query.value = ''
  available.value = []
  selected.value = []
}
</script>

<style lang="scss" scoped>
.handle {
  cursor: grab;
}

.handle:active {
  cursor: grabbing;
}
</style>
