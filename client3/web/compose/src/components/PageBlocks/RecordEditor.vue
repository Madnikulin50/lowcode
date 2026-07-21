<template>
  <Wrap card-class="position-static">
    <div v-if="isProcessing" class="d-flex align-items-center justify-content-center h-100">
      <span class="spinner-border" />
    </div>
    <div v-else-if="module" ref="fieldContainer" class="mt-3" :class="fieldLayoutClass">
      <template v-for="field in fields" :key="`${field.fieldID}-${field.name}`">
        <div v-if="canDisplay(field)" :class="`field-container ${columnWrapClass}`" :style="fieldWidth">
          <FieldEditor
            v-if="isFieldEditable(field)"
            :field="field"
            :errors="fieldErrors(field.name)"
            :extra-options="options"
            :horizontal="horizontal"
            @change="onFieldChange(field)"
          />
          <div v-else class="mb-3">
            <label class="form-label d-flex align-items-center text-primary mb-0">
              <span class="d-flex">{{ field.label || field.name }}</span>
              <c-hint :tooltip="((field.options?.hint || {}).view || '')" />
            </label>
            <div class="small text-muted" :class="{ 'mb-1': !!((field.options?.description || {}).view) }">
              {{ (field.options?.description || {}).view }}
            </div>
            <div v-if="field.canReadRecordValue" class="value align-self-center">
              <FieldViewer :field="field" value-only />
            </div>
            <i v-else class="text-muted">{{ $t('field.noPermission') }}</i>
          </div>
        </div>
      </template>
    </div>
  </Wrap>
</template>

<script setup>
import { computed, watch, onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'
import { debounce } from 'lodash'
import { usePageBlockBase } from './usePageBlockBase'
import Wrap from './Wrap/index.js'
import FieldEditor from 'corteza-webapp-compose/src/components/ModuleFields/Editor'
import FieldViewer from 'corteza-webapp-compose/src/components/ModuleFields/Viewer'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, required: false, default: '' },
  editable: { type: Boolean, required: false, default: false },
  resizing: { type: Boolean, required: false, default: false },
  magnified: { type: Boolean, required: false, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, required: false, default: false },
  errors: { type: Object, required: false, default: () => ({}) },
})

const { isProcessing, options, fieldErrors } = usePageBlockBase(props, {})
const evaluating = ref(false)

const fields = computed(() => {
  if (!props.module) return []
  if (!options.value.fields || options.value.fields.length === 0) return props.module.fields
  return props.module.filterFields(options.value.fields).map(f => ({
    ...f, label: f.isSystem ? $t(`system.${f.name}`) : f.label || f.name,
  }))
})

const fieldLayoutClass = computed(() => {
  const classes = { default: 'd-flex flex-column px-3', noWrap: 'd-flex gap-2 ps-3', wrap: 'row g-0' }
  return classes[options.value.recordFieldLayoutOption]
})

const fieldWidth = computed(() => {
  if (options.value.recordFieldLayoutOption !== 'noWrap') return {}
  return { 'min-width': '20rem' }
})

const horizontal = computed(() => props.block.options.horizontalFieldLayoutEnabled)

const isNew = computed(() => props.record && props.record.recordID === NoID)

function canDisplay(field) { return field?.canReadRecordValue !== false }
function isFieldEditable(field) { return field?.canUpdateRecordValue !== false }

const onFieldChange = debounce(function(field) {
  window.dispatchEvent(new CustomEvent('record-field-change', { detail: { fieldName: field.name } }))
}, 500)

function createEvents() {
  window.addEventListener('record-field-change', () => {})
}

function destroyEvents() {
  window.removeEventListener('record-field-change', () => {})
}

onMounted(() => { createEvents() })
onBeforeUnmount(() => { destroyEvents() })
</script>

<style scoped>
.field-col > * { margin-left: 1rem; margin-right: 1rem; }
</style>
