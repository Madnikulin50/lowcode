<template>
  <table
    v-if="mockModule"
    class="table table-sm table-borderless mb-0"
    style="border-collapse: separate; border-spacing: 0;"
  >
    <template v-for="(filterGroup, groupIndex) in internalFilter">
      <template v-if="filterGroup.filter.length">
        <tr
          v-for="(filter, index) in filterGroup.filter"
          :key="`${groupIndex}-${index}`"
          class="pb-2"
          style="border: none;"
        >
          <td style="width: 250px; border: none; padding: 0 0 0.5rem 0;">
            <c-input-select
              v-model="filter.name"
              :options="fields"
              :get-option-key="getOptionKey"
              :clearable="false"
              :placeholder="$t('recordList.filter.fieldPlaceholder')"
              :reduce="(f) => f.name"
              :class="{ 'filter-field-picker': !!filter.name }"
              @update:model-value="onChange($event, groupIndex, index)"
            />
          </td>

          <td
            v-if="getPreparedField(filter.name)"
            style="width: 250px; border: none; padding: 0 0 0.5rem 0.5rem;"
          >
            <select
              v-if="getPreparedField(filter.name)"
              v-model="filter.operator"
              class="form-select form-control form-select-sm d-flex field-operator w-100"
              @change="onValueChange"
            >
              <option
                v-for="opt in getOperators(filter.kind, getPreparedField(filter.name))"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.text }}
              </option>
            </select>
          </td>

          <td
            v-if="getPreparedField(filter.name)"
            :key="`${getPreparedField(filter.name)?.fieldID}-${filter.name}`"
            style="border: none; padding: 0 0 0.5rem 0.5rem;"
          >
            <template v-if="isBetweenOperator(filter.operator)">
              <template v-if="getPreparedField(`${filter.name}-start`)">
                <field-editor
                  :field="getPreparedField(`${filter.name}-start`)"
                  :record="filter.record"
                  :module="mockModule"
                  :namespace="namespace"
                  :errors="errors"
                  value-only
                  class="mb-0 field-editor"
                  @change="onValueChange"
                />
                <div class="my-1 text-center w-100">
                  {{ $t('label.and') }}
                </div>
                <field-editor
                  :field="getPreparedField(`${filter.name}-end`)"
                  :record="filter.record"
                  :module="mockModule"
                  :namespace="namespace"
                  :errors="errors"
                  value-only
                  class="mb-0 field-editor"
                  @change="onValueChange"
                />
              </template>
            </template>

            <template v-else>
              <field-editor
                :field="getPreparedField(filter.name)"
                :errors="errors"
                :record="filter.record"
                :module="mockModule"
                :namespace="namespace"
                value-only
                class="mb-0 field-editor"
                @change="onValueChange"
              />
            </template>
          </td>

          <td
            v-if="getPreparedField(filter.name)"
            style="width: 1%; border: none; padding: 0 0 0.5rem 0.5rem;"
          >
            <button
              type="button"
              class="btn btn-outline-extra-light d-block text-dark border-0 h-full px-2 mt-1"
              style="padding-top: 0; padding-bottom: 0;"
              @click="deleteFilter(groupIndex, index)"
            >
              <font-awesome-icon
                :icon="['far', 'trash-alt']"
                size="sm"
              />
            </button>
          </td>
        </tr>

        <tr
          v-if="showAddCondition"
          :key="`addFilter-${groupIndex}`"
          style="border: none;"
        >
          <td style="border: none; padding: 0 0 0.5rem 0;">
            <button
              type="button"
              class="btn btn-primary btn-sm d-block me-auto"
              @click="addFilter(groupIndex)"
            >
              <font-awesome-icon
                :icon="['fas', 'plus']"
                class="me-1"
              />
              {{ $t('label.add') }}
            </button>
          </td>
        </tr>

        <tr :key="`groupCondtion-${groupIndex}`" style="border: none;">
          <td
            colspan="100%"
            class="p-0 justify-content-center"
            :class="{ 'pb-2': groupIndex !== (internalFilter.length - 1) }"
            style="border: none;"
          >
            <div class="group-separator">
              <button
                v-if="groupIndex === (internalFilter.length - 1)"
                type="button"
                class="btn btn-outline-primary btn-add-group d-block py-2 px-3 m-auto bg-white"
                @click="addGroup()"
              >
                <font-awesome-icon
                  :icon="['fas', 'plus']"
                  class="mb-0 h6"
                />
              </button>

              <select
                v-else
                v-model="internalFilter[groupIndex + 1].groupCondition"
                class="form-select form-control form-select-sm group-condition-select"
                @change="onGroupConditionChange"
              >
                <option
                  v-for="opt in groupConditionOptions"
                  :key="opt.value"
                  :value="opt.value"
                >
                  {{ opt.text }}
                </option>
              </select>
            </div>
          </td>
        </tr>
      </template>
    </template>
  </table>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose, validator } from 'corteza-lib/js/dist'
import FieldEditor from 'corteza-webapp-compose/src/components/ModuleFields/Editor'
import { isBetweenOperator } from 'corteza-webapp-compose/src/lib/record-filter.js'

const { t } = useI18n()

const props = defineProps({
  modelValue: {
    type: Array,
    default: undefined,
  },
  value: {
    type: Array,
    default: undefined,
  },
  module: {
    type: compose.Module,
    required: true,
  },
  namespace: {
    type: compose.Namespace,
    required: true,
  },
  selectedField: {
    type: Object,
    default: undefined,
  },
  startEmpty: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:modelValue', 'input', 'value-change'])

const conditions = ref([
  { value: 'AND', text: t('recordList.filter.conditions.and') },
  { value: 'OR', text: t('recordList.filter.conditions.or') },
])

const errors = ref(new validator.Validated())

const mockModule = ref(undefined)
const preparedFields = ref([])
const internalFilter = ref([])
const isLoadingExternalData = ref(false)

const fields = computed(() => {
  if (!mockModule.value) return []
  return [
    ...[...mockModule.value.fields].sort((a, b) => (a.label || a.name).localeCompare(b.label || b.name)),
    ...mockModule.value.systemFields().map((sf) => {
      sf.label = t(`system.${sf.name}`)
      return sf
    }),
  ].filter(({ isFilterable, canReadRecordValue }) => {
    return isFilterable && canReadRecordValue !== false
  })
})

const showAddCondition = computed(() => {
  return internalFilter.value.length >= 1 && internalFilter.value[0].filter[0].name
})

const groupConditionOptions = computed(() => {
  return [
    { value: 'OR', text: t('recordList.filter.conditions.or') },
    { value: 'AND', text: t('recordList.filter.conditions.and') },
  ]
})

const resolvedSelectedField = computed(() => {
  if (props.selectedField) {
    return props.selectedField
  } else if (fields.value.length) {
    return fields.value[0]
  }
  return {}
})

watch(() => props.module, (newModule) => {
  mockModule.value = new compose.Module(newModule)
  if (mockModule.value) {
    prepareFields()
  }
}, { immediate: true })

watch(() => props.modelValue ?? props.value, (rawFilter) => {
  let internal = []

  if (!rawFilter || !rawFilter.length) {
    internal = [createDefaultFilterGroup(props.startEmpty ? undefined : resolvedSelectedField.value)]
  } else {
    internal = rawFilterToInternal(rawFilter)
    if (!internal.length) {
      internal = [createDefaultFilterGroup(props.startEmpty ? undefined : resolvedSelectedField.value)]
    } else {
      isLoadingExternalData.value = true
    }
  }

  internalFilter.value = internal

  nextTick(() => {
    isLoadingExternalData.value = false
  })
}, { immediate: true, deep: true })

watch(internalFilter, (val) => {
  if (isLoadingExternalData.value) return
  const processed = processInternalFilter(val)
  emit('update:modelValue', processed)
  emit('input', processed)
}, { deep: true })

function rawFilterToInternal (recordListFilter = []) {
  if (!recordListFilter.length || !mockModule.value) return []

  return recordListFilter.map(({ filter = [], name, groupCondition = 'OR' }) => {
    filter = filter.map(({ value, ...f } = {}) => {
      f.record = new compose.Record(mockModule.value, {})
      if (isBetweenOperator(f.operator)) {
        const field = getPreparedField(f.name)
        if (field && field.isSystem) {
          f.record[`${f.name}-start`] = value?.start
          f.record[`${f.name}-end`] = value?.end
        } else {
          f.record.values[`${f.name}-start`] = value?.start
          f.record.values[`${f.name}-end`] = value?.end
        }
      } else if (Object.keys(f.record.values).includes(f.name)) {
        f.record.values[f.name] = value
      } else if (Object.keys(f.record).includes(f.name)) {
        f.record[f.name] = value
      }
      return f
    })
    return { filter, name, groupCondition }
  })
}

function processInternalFilter (filter = []) {
  if (!filter.length || !mockModule.value) return []

  return filter.map(({ filter = [], name, groupCondition }) => {
    filter = filter.map(({ record, ...f }) => {
      if (!f.name || !record) return undefined
      if (isBetweenOperator(f.operator)) {
        const field = getPreparedField(f.name)
        if (field) {
          f.value = {
            start: field.isSystem ? record[`${f.name}-start`] : record.values[`${f.name}-start`],
            end: field.isSystem ? record[`${f.name}-end`] : record.values[`${f.name}-end`],
          }
        }
      } else if (Object.keys(record.values).includes(f.name)) {
        f.value = record.values[f.name]
      } else if (Object.keys(record).includes(f.name)) {
        f.value = record[f.name]
      }
      return f
    })
    return { filter, name, groupCondition }
  })
}

function prepareFields () {
  const flds = []
  fields.value.forEach(f => {
    if (f.isMulti) {
      f.isMulti = false
      f.multi = true
    }
    if (f.kind === 'Record') {
      f.options.prefilter = ''
    }
    if (f.kind === 'DateTime') {
      f.options.onlyFutureValues = false
      f.options.onlyPastValues = false
    }
    if (f.kind === 'Number') {
      f.options.min = undefined
      f.options.max = undefined
    }
    flds.push(f)
    if (f.kind === 'DateTime' || f.kind === 'Number') {
      flds.push({ ...f, name: `${f.name}-start` })
      flds.push({ ...f, name: `${f.name}-end` })
    }
  })
  preparedFields.value = flds
}

function onChange (fieldName, groupIndex, index) {
  const field = getPreparedField(fieldName)
  const filterExists = !!((internalFilter.value[groupIndex] || { filter: [] }).filter[index])
  if (field && filterExists) {
    internalFilter.value[groupIndex].filter[index].kind = field.kind
    internalFilter.value[groupIndex].filter[index].name = field.name
    internalFilter.value[groupIndex].filter[index].value = undefined
    internalFilter.value[groupIndex].filter[index].operator = field.multi ? 'IN' : '='
  }
  emit('value-change')
}

function onValueChange () {
  emit('value-change')
}

function getOperators (kind, field) {
  const operators = [
    { value: '=', text: t('recordList.filter.operators.equal') },
    { value: '!=', text: t('recordList.filter.operators.notEqual') },
  ]
  const inOperators = [
    { value: 'IN', text: t('recordList.filter.operators.contains') },
    { value: 'NOT IN', text: t('recordList.filter.operators.notContains') },
  ]
  const lgOperators = [
    { value: '>', text: t('recordList.filter.operators.greaterThan') },
    { value: '<', text: t('recordList.filter.operators.lessThan') },
  ]
  const matchOperators = [
    { value: 'LIKE', text: t('recordList.filter.operators.like') },
    { value: 'NOT LIKE', text: t('recordList.filter.operators.notLike') },
  ]
  const betweenOperators = [
    { value: 'BETWEEN', text: t('recordList.filter.operators.between') },
    { value: 'NOT BETWEEN', text: t('recordList.filter.operators.notBetween') },
  ]

  if (field.multi) return inOperators

  switch (kind) {
    case 'Number':
    case 'DateTime':
      return [...operators, ...lgOperators, ...betweenOperators]
    case 'String':
    case 'Url':
    case 'Email':
      return [...operators, ...matchOperators]
    default:
      return operators
  }
}

function deleteFilter (groupIndex, index) {
  const filterExists = !!((internalFilter.value[groupIndex] || { filter: [] }).filter[index])
  if (filterExists) {
    internalFilter.value[groupIndex].filter.splice(index, 1)
    if (!internalFilter.value[groupIndex].filter.length) {
      internalFilter.value.splice(groupIndex, 1)
      if (!internalFilter.value.length) {
        internalFilter.value = [createDefaultFilterGroup()]
      }
    }
  }
  emit('value-change')
}

function getOptionKey ({ name }) {
  return name
}

function getPreparedField (name = '') {
  if (!preparedFields.value.length) return undefined
  return preparedFields.value.find(f => f.name === name)
}

function addFilter (groupIndex) {
  if ((internalFilter.value[groupIndex] || {}).filter) {
    internalFilter.value[groupIndex].filter.push(createDefaultFilter(resolvedSelectedField.value))
  }
}

function createDefaultFilter (field = {}) {
  return {
    name: field.name,
    operator: field.isMulti ? 'IN' : '=',
    value: undefined,
    kind: field.kind,
    record: new compose.Record(mockModule.value, {}),
  }
}

function createDefaultFilterGroup (field, groupCondition = 'OR') {
  return {
    filter: [createDefaultFilter(field)],
    groupCondition,
  }
}

function onGroupConditionChange () {
  emit('value-change')
}

function addGroup () {
  internalFilter.value.push(createDefaultFilterGroup(resolvedSelectedField.value))
  emit('value-change')
}
</script>

<style lang="scss" scoped>
.group-separator {
  display: flex;
  align-items: center;
  justify-content: center;
  background-image: linear-gradient(to left, lightgray, lightgray);
  background-repeat: no-repeat;
  background-size: 100% 1px;
  background-position: center;
}

td {
  padding: 0;
  padding-bottom: 0.5rem;
}

.btn-add-group {
  &:hover,
  &:active {
    background-color: var(--primary) !important;
    color: var(--white) !important;
  }
}

.group-condition-select {
  width: auto;
  min-width: 80px;
  border: 1px solid var(--light);
}
</style>
