<template>
  <div class="d-flex flex-column h-100 overflow-hidden">
    <div class="mb-0 flex-fill overflow-auto">
      <label class="form-label text-primary p-3 mb-0">{{ $t('recordList.import.matchFields') }}</label>

      <div v-if="hasRequiredFileFields" class="small px-3 mb-3">
        {{ $t('recordList.import.hasRequiredFileFields') }}: {{ showRequiredFields }}
      </div>

      <table class="table field-table mb-0" style="max-height: 70vh;">
        <thead class="table-outline-secondary">
          <tr>
            <th style="width: 30px">
              <div class="form-check">
                <input
                  class="form-check-input"
                  type="checkbox"
                  :checked="selectAll"
                  @change="onSelectAll"
                />
              </div>
            </th>
            <th>{{ $t('recordList.import.fileColumns') }}</th>
            <th style="width: 25rem">{{ $t('recordList.import.moduleFields') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, idx) in rows" :key="idx">
            <td class="align-middle">
              <div class="form-check">
                <input
                  v-model="row.selected"
                  class="form-check-input"
                  type="checkbox"
                />
              </div>
            </td>
            <td class="align-middle">{{ row.fileColumn }}</td>
            <td>
              <c-input-select
                v-model="row.moduleField"
                :options="moduleFields"
                :reduce="o => o.name"
                :placeholder="$t('recordList.import.pickModuleField')"
                :class="{ 'border border-danger': row.selected && !row.moduleField }"
                @input="moduleChanged(row)"
              />
              <span
                v-if="row.fileColumn === 'id'"
                class="small text-muted"
              >
                {{ $t('recordList.import.idFieldDescription') }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="mt-auto p-3">
      <button
        class="btn btn-outline-secondary float-start"
        @click="$emit('back')"
      >
        {{ $t('label.back') }}
      </button>
      <button
        class="btn btn-primary float-end"
        :disabled="!canContinue"
        @click="nextStep"
      >
        {{ $t('label.import') }}
      </button>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  session: { type: Object, required: true, default: () => ({}) },
  module: { type: Object, required: true, default: () => ({}) },
})

const emit = defineEmits(['fieldsMatched', 'back'])

const rows = ref([])
const unsupportedFields = ref([])

// Init rows
const { fields = {} } = props.session
const moduleFieldsMap = { id: 'recordID' }
props.module.fields.forEach(({ name }) => { moduleFieldsMap[name] = name })
props.module.systemFields().forEach(({ name }) => { moduleFieldsMap[name] = name })

rows.value = Object.entries(fields).map(([fileColumn, moduleField]) => {
  moduleField = moduleField || moduleFieldsMap[fileColumn]
  return { selected: false, fileColumn, moduleField }
})

const selectAll = computed({
  get: () => rows.value.reduce((acc, { selected }) => acc && selected, true),
  set: (v) => { rows.value.forEach((r) => { r.selected = v }) },
})

const canContinue = computed(() => {
  const selected = rows.value.filter(({ selected }) => selected)
  const named = selected.filter(({ moduleField }) => !!moduleField)
  return !!selected.length && selected.length === named.length && !hasRequiredFileFields.value
})

const moduleFields = computed(() => {
  return [
    ...props.module.fields,
    ...props.module.systemFields().map(({ name }) => ({
      label: $t(`system.${name}`),
      name,
    })),
  ].filter(({ kind }) => !['File'].includes(kind))
    .map(field => field.isRequired === true ? { ...field, label: field.label + '*' } : field)
})

const requiredFields = computed(() => props.module.fields.filter(field => field.isRequired === true))

const filteredRows = computed(() => {
  const result = rows.value.filter(row => {
    return requiredFields.value.some(field => row.moduleField === field.name)
  })
  return result.filter((value, index, self) => self.findIndex(v => v.moduleField === value.moduleField) === index)
})

const hasRequiredFileFields = computed(() => {
  return !(requiredFields.value.length === filteredRows.value.length)
})

const showRequiredFields = computed(() => {
  let result = requiredFields.value.filter(field => {
    return filteredRows.value.some(row => field.name !== row.moduleField)
  })
  if (result.length === 0) result = requiredFields.value
  return result.map(field => field.label).join(', ')
})

function moduleChanged(row) {
  if (row.moduleField) {
    const result = rows.value.find(r => r.moduleField === row.moduleField)
    if (result) result.selected = true
  }
}

function nextStep() {
  if (!canContinue.value) return
  const rtr = {}
  rows.value.forEach(({ selected, fileColumn, moduleField }) => {
    if (selected) rtr[fileColumn] = moduleField
  })
  emit('fieldsMatched', rtr)
}

function onSelectAll(e) {
  selectAll.value = e
}
</script>
