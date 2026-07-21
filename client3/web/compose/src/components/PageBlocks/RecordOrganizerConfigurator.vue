<template>
  <div>
    <h5>{{ $t('recordOrganizer.label') }}</h5>
    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('module') }}</label>
      <c-input-select v-model="options.moduleID" :options="modules" label="name" :reduce="m => m.moduleID" :placeholder="$t('recordOrganizer.module.placeholder')" default-value="0" required />
    </div>

    <div v-if="selectedModule">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('field.selector.available') }}</label>
        <div class="d-flex">
          <div class="border fields w-100 p-2">
            <div v-for="field in allFields" :key="field.name" class="field">
              <span v-if="field.label">{{ field.label }} ({{ field.name }})</span>
              <span v-else>{{ field.name }}</span>
              <span class="small float-end">
                <span v-if="field.isSystem">{{ $t('field.selector.systemField') }}</span>
                <span v-else>{{ field.kind }}</span>
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('recordList.record.prefilterLabel') }}</label>
        <c-input-expression v-model.trim="options.filter" min-height="3.688rem" :suggestion-params="recordAutoCompleteParams" :placeholder="$t('recordList.record.prefilterPlaceholder')" />
        <small class="text-muted d-block mt-1">
          <code>${record.values.fieldName}</code>
          <code>${recordID}</code>
          <code>${ownerID}</code>
          <code>${userID}</code>, <code>${user.name}</code>
        </small>
      </div>

      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('recordOrganizer.labelField.label') }}</label>
            <c-input-select v-model="options.labelField" :options="allFields" :reduce="o => o.name" :get-option-label="getFieldLabel" :placeholder="$t('label.none')" />
            <small class="text-muted d-block">{{ $t('recordOrganizer.labelField.footnote') }}</small>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('recordOrganizer.descriptionField.label') }}</label>
            <c-input-select v-model="options.descriptionField" :options="allFields" :reduce="o => o.name" :get-option-label="getFieldLabel" :placeholder="$t('label.none')" />
            <small class="text-muted d-block">{{ $t('recordOrganizer.descriptionField.footnote') }}</small>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('recordOrganizer.groupField.label') }}</label>
            <c-input-select v-model="options.groupField" :options="groupFields" :reduce="o => o.name" :get-option-label="getFieldLabel" :placeholder="$t('label.none')" />
            <small class="text-muted d-block">{{ $t('recordOrganizer.groupField.footnote') }}</small>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('recordOrganizer.group.label') }}</label>
            <FieldEditor v-if="options.groupField" v-bind="mock" value-only class="mb-0" />
            <input v-else class="form-control" disabled />
            <small class="text-muted d-block">{{ $t('recordOrganizer.group.footnote') }}</small>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('recordOrganizer.positionField.label') }}</label>
            <c-input-select v-model="options.positionField" :placeholder="$t('recordOrganizer.positionField.placeholder')" :reduce="f => f.name" label="label" />
            <small class="text-muted d-block">{{ $t('recordOrganizer.positionField.footnote') }}</small>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('recordOrganizer.onRecordClick') }}</label>
            <select v-model="options.displayOption" class="form-select">
              <option v-for="opt in displayOptions" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
            </select>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { useStore } from '../../store'
import FieldEditor from '../ModuleFields/Editor'
import { compose, validator } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import autocomplete from 'corteza-webapp-compose/src/mixins/autocomplete.js'

const { CInputExpression } = components
const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  block: { type: Object, required: true },
  module: { type: Object, required: false },
})

const store = useStore()

const options = computed(() => props.block.options)

const mock = ref({
  namespace: undefined,
  module: undefined,
  field: undefined,
  record: undefined,
  errors: new validator.Validated(),
})

const modules = computed(() => store.module.set)
const selectedModule = computed(() => modules.value.find(m => m.moduleID === options.value.moduleID))
const selectedModuleFields = computed(() => {
  if (!selectedModule.value) return []
  return [...selectedModule.value.fields].sort((a, b) => a.label.localeCompare(b.label))
})

const allFields = computed(() => {
  if (!options.value.moduleID) return []
  const sm = selectedModule.value
  if (!sm) return []
  return [
    ...selectedModuleFields.value,
    ...sm.systemFields().map(sf => { sf.label = $t(`system.${sf.name}`); return sf }),
  ]
})

const positionFields = computed(() => selectedModuleFields.value.filter(({ kind, isMulti }) => kind === 'Number' && !isMulti))
const groupFields = computed(() => allFields.value.filter(({ isMulti }) => !isMulti))
const group = computed(() => allFields.value.find(f => f.name === options.value.groupField))

const displayOptions = computed(() => [
  { value: 'sameTab', text: $t('recordOrganizer.openInSameTab') },
  { value: 'newTab', text: $t('recordOrganizer.openInNewTab') },
  { value: 'modal', text: $t('recordOrganizer.openInModal') },
])

const recordAutoCompleteParams = computed(() => {
  if (typeof autocomplete.processRecordAutoCompleteParams === 'function') return autocomplete.processRecordAutoCompleteParams({ module: selectedModule.value })
  return {}
})

watch(() => options.value.moduleID, () => {
  options.value.labelField = ''
  options.value.descriptionField = ''
  options.value.positionField = ''
  options.value.groupField = ''
})

watch(() => options.value.groupField, (newGroupField, oldGroupField) => {
  if (oldGroupField) options.value.group = undefined
  if (newGroupField) {
    const gf = groupFields.value.find(f => f.name === newGroupField)
    if (gf) {
      mock.value.namespace = props.namespace
      mock.value.field = compose.ModuleFieldMaker(gf)
      mock.value.field.apply({ name: 'group' })
      mock.value.module = new compose.Module({ fields: [mock.value.field] }, props.namespace)
      mock.value.record = new compose.Record(mock.value.module, { group: options.value.group })
    }
  }
}, { immediate: true })

watch(() => mock.value.record?.values?.group, (group) => { options.value.group = group }, { deep: true })

onBeforeUnmount(() => { setDefaultValues() })

function getFieldLabel(option) { return `${option.label || option.name} (${option.kind})` }
function setDefaultValues() { mock.value = {} }
</script>
<style lang="scss" scoped>
.fields { height: 150px; overflow-y: auto; cursor: default; }
</style>
