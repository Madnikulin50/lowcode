<template>
  <div>
    <div class="mb-3">
      <label class="form-label text-primary">{{ t('kind.record.moduleLabel') }}</label>
      <c-input-select
        v-model="f.options.moduleID"
        :options="moduleOptions"
        label="name"
        :placeholder="t('kind.record.modulePlaceholder')"
        default-value="0"
        :reduce="m => m.moduleID"
      />
    </div>

    <template v-if="selectedModule">
      <div class="mb-3">
        <label class="d-flex align-items-center text-primary p-0">
          {{ t('kind.record.moduleField') }}
          <c-hint :tooltip="t('kind.record.tooltip.moduleField')" icon-class="text-warning" />
        </label>
        <c-input-select
          v-model="f.options.labelField"
          :options="fieldOptions"
          label="text"
          :placeholder="t('kind.record.pickField')"
          :reduce="field => field.value"
        />
      </div>

      <div v-if="labelField && labelField.kind === 'Record'">
        <div class="mb-3">
          <label class="form-label text-primary">{{ t('kind.record.fieldFromModuleField') }}</label>
          <c-input-select
            v-model="f.options.recordLabelField"
            :options="labelFieldOptions"
            :disabled="!labelFieldModule"
            label="text"
            :placeholder="t('kind.record.pickField')"
            :reduce="field => field.value"
          />
        </div>
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ t('kind.record.queryFieldsLabel') }}</label>
        <c-input-select
          v-model="f.options.queryFields"
          :options="queryFieldOptions"
          label="text"
          :reduce="field => field.value"
          :placeholder="t('kind.record.queryFieldsPlaceholder')"
          multiple
        />
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ t('kind.record.prefilterLabel') }}</label>
        <textarea v-model="f.options.prefilter" class="form-control" :placeholder="t('kind.record.prefilterPlaceholder')"></textarea>
        <small class="text-muted">{{ t('kind.record.prefilterFootnote', { interpolation: { escapeValue: false } }) }}</small>
      </div>

      <template v-if="field.isMulti">
        <div class="mb-3">
          <label class="form-label text-primary pr-2">{{ t('kind.select.optionType.label') }}</label>
          <div class="btn-group" data-bs-toggle="buttons">
            <label
              v-for="opt in selectOptions"
              :key="opt.value"
              class="btn btn-outline-primary btn-sm"
              :class="{ active: f.options.selectType === opt.value }"
            >
              <input
                type="radio"
                class="btn-check"
                :value="opt.value"
                :checked="f.options.selectType === opt.value"
                autocomplete="off"
                @change="onSelectTypeChange(opt.value)"
              />
              {{ opt.text }}
            </label>
          </div>
        </div>

        <div v-if="shouldAllowDuplicates" class="form-check mb-3">
          <input
            id="allowDuplicates"
            v-model="f.options.isUniqueMultiValue"
            :true-value="false"
            :false-value="true"
            type="checkbox"
            class="form-check-input"
          />
          <label class="form-check-label" for="allowDuplicates">{{ t('kind.select.allow-duplicates') }}</label>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'
import { nonQueryableFieldKinds } from 'corteza-webapp-compose/src/lib/record-filter'
import { useConfiguratorBase } from './base'
import { useModuleStore } from 'corteza-webapp-compose/src/stores/module'

const props = defineProps({
  namespace: { type: Object, required: true },
  module: { type: Object, required: true },
  field: { type: Object, required: true },
  hasRecords: { type: Boolean, default: false },
})

const emit = defineEmits(['update:field'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { f, isNew, hasData } = useConfiguratorBase(props, emit)

const moduleStore = useModuleStore()

const selectOptions = computed(() => [
  { text: t('kind.select.optionType.default'), value: 'default', allowDuplicates: true },
  { text: t('kind.select.optionType.multiple'), value: 'multiple' },
  { text: t('kind.select.optionType.each'), value: 'each', allowDuplicates: true },
])

const moduleOptions = computed(() => {
  let modules = moduleStore.set || []
  if (props.module.moduleID === NoID) {
    modules = [
      ({ moduleID: '-1', name: props.module.name || t('kind.record.currentUnnamedModule') }),
      ...modules,
    ]
  }
  return modules
})

const selectedModule = computed(() => {
  if (f.value.options.moduleID === '-1') {
    return props.module
  } else if (f.value.options.moduleID !== NoID) {
    return moduleStore.getByID(f.value.options.moduleID)
  }
  return undefined
})

const fieldOptions = computed(() => {
  const fields = selectedModule.value
    ? selectedModule.value.fields.map(({ label, name, kind }) => ({ value: name, text: label || name, kind }))
    : []
  return [...fields.sort((a, b) => a.text.localeCompare(b.text))]
})

const queryFieldOptions = computed(() => {
  return fieldOptions.value.filter(({ kind }) => !nonQueryableFieldKinds.includes(kind))
})

const labelField = computed(() => {
  if (f.value.options.labelField && selectedModule.value) {
    return selectedModule.value.fields.find(({ name }) => name === f.value.options.labelField)
  }
  return undefined
})

const labelFieldModule = computed(() => {
  if (labelField.value) {
    return moduleStore.getByID(labelField.value.options.moduleID)
  }
  return undefined
})

const labelFieldOptions = computed(() => {
  let fields = []
  if (labelField.value && labelFieldModule.value) {
    fields = labelFieldModule.value.fields.map(({ label, name }) => ({ value: name, text: label || name }))
    return [...fields.sort((a, b) => a.text.localeCompare(b.text))]
  }
  return fields
})

const shouldAllowDuplicates = computed(() => {
  if (!f.value.isMulti) return false
  const { allowDuplicates } = selectOptions.value.find(({ value }) => value === f.value.options.selectType) || {}
  return !!allowDuplicates
})

watch(() => f.value.options.moduleID, () => {
  f.value.options.labelField = ''
  f.value.options.selectType = 'default'
})

watch(() => f.value.options.labelField, () => {
  f.value.options.queryFields = []
  f.value.options.prefilter = ''
  f.value.options.recordLabelField = ''
})

function onSelectTypeChange (value) {
  f.value.options.selectType = value
  updateIsUniqueMultiValue(value)
}

function updateIsUniqueMultiValue (value) {
  const { allowDuplicates = false } = selectOptions.value.find(({ value: v }) => v === value) || {}
  if (!allowDuplicates) {
    f.value.options.isUniqueMultiValue = true
  }
}
</script>
