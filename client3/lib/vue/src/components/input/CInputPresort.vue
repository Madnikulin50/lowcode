<template>
  <c-form-table-wrapper
    :labels="{ addButton: labels.addButton }"
    :hide-add-button="textInput"
    @add-item="items.push({ field: undefined, descending: false })"
  >
    <div
      v-if="!textInput"
      class="mb-3"
    >
      <label class="form-label text-primary">
        {{ labels.title }}
      </label>
      <table
        class="table table-borderless table-sm mb-0"
        responsive
      >
        <draggable
          v-model="items"
          group="sort"
          handle=".grab"
          tag="tbody"
        >
          <tr
            v-for="(column, index) in items"
            :key="index"
          >
            <td
              class="grab text-center align-middle"
              style="width: 40px;"
            >
              <font-awesome-icon
                :icon="['fas', 'bars']"
                class="text-secondary"
              />
            </td>
            <td
              class="align-middle"
              style="min-width: 250px;"
            >
              <c-input-select
                v-model="column.field"
                :options="availableFields"
                :reduce="(o: Record<string, unknown>) => o.name"
                :placeholder="labels.none"
                class="rounded"
              />
            </td>
            <td
              class="text-center align-middle"
              style="min-width: 200px;"
            >
              <div
                class="btn-group btn-group-sm bg-white"
                role="group"
              >
                <input
                  v-for="dir in sortDirections"
                  :key="String(dir.value)"
                  type="radio"
                  class="btn-check"
                  :name="`sort-${index}`"
                  :id="`sort-${index}-${String(dir.value)}`"
                  :value="dir.value"
                  :checked="column.descending === dir.value"
                  @change="column.descending = dir.value"
                />
                <label
                  v-for="dir in sortDirections"
                  :key="`label-${String(dir.value)}`"
                  class="btn btn-outline-primary"
                  :class="{ active: column.descending === dir.value }"
                  :for="`sort-${index}-${String(dir.value)}`"
                >
                  {{ dir.text }}
                </label>
              </div>
            </td>
            <td
              class="align-middle text-end"
              style="min-width: 80px; width: 80px;"
            >
              <c-input-confirm
                show-icon
                @confirmed="items.splice(index, 1)"
              />
            </td>
          </tr>
        </draggable>
      </table>
    </div>

    <div v-else>
      <textarea
        v-model="presortValue"
        class="form-control"
        :placeholder="labels.placeholder"
      />
      <div class="form-text">
        {{ labels.footnote }}
      </div>
    </div>

    <div
      v-if="allowTextInput"
      class="d-flex align-items-center mt-1"
    >
      <button
        class="btn btn-link btn-sm text-decoration-none ms-auto"
        @click="textInput = !textInput"
      >
        {{ labels.toggleInput }}
      </button>
    </div>
  </c-form-table-wrapper>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import Draggable from 'vuedraggable'
import CInputSelect from './CInputSelect.vue'
import CFormTableWrapper from '../wrapper/CFormTableWrapper.vue'

interface SortField {
  field: string | undefined
  descending: boolean
}

interface FieldOption {
  name: string
  label: string
}

interface PresortLabels {
  addButton?: string
  title?: string
  none?: string
  ascending?: string
  descending?: string
  placeholder?: string
  footnote?: string
  toggleInput?: string
}

interface Props {
  modelValue: string
  fields: FieldOption[]
  labels: PresortLabels
  allowTextInput?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  allowTextInput: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const items = ref<SortField[]>([])
const textInput = ref(false)

const sortDirections = computed(() => [
  { value: false, text: props.labels.ascending },
  { value: true, text: props.labels.descending },
])

const availableFields = computed(() =>
  props.fields.map(f => ({ ...f, label: `${f.label} (${f.name})` }))
)

const presortValue = computed({
  get: () => props.modelValue,
  set: (value: string) => emit('update:modelValue', value),
})

watch(() => props.modelValue, (value) => {
  if (value) {
    const sort = value.includes(',') ? value.split(',') : [value]
    items.value = sort.map((field: string) => {
      let descending = false
      if (field.includes(' ')) {
        field = field.split(' ')[0]
        descending = true
      }
      return { field, descending: !!descending }
    })
  } else {
    items.value = [{ field: undefined, descending: false }]
  }
}, { immediate: true })

watch(items, (newItems: SortField[] = [], oldItems: SortField[] | undefined) => {
  if (oldItems) {
    emit('update:modelValue', newItems
      .filter(({ field }) => field)
      .map(({ field, descending }) => descending ? `${field} DESC` : field)
      .join(','))
  }
}, { deep: true })
</script>
