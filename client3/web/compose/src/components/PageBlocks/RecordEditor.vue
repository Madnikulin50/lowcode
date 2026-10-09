<template>
  <Wrap :block="block" :record="record" :loading-record="loadingRecord" :magnified="magnified" card-class="position-static" body-class="pt-3 px-3">
    <div v-if="busy" class="d-flex align-items-center justify-content-center h-100">
      <span class="spinner-border" />
    </div>
    <div v-else-if="module" ref="fieldContainer" class="fixed-corner-container" :class="fieldLayoutClass">
      <block-help-button
        :block="block"
        variant="metric"
      />
      <template v-for="field in fields" :key="`${field.fieldID}-${field.name}`">
        <div v-if="canDisplay(field)" :class="`field-container ${columnWrapClass}`" :style="fieldWidth">
          <FieldEditor
            v-if="isFieldEditable(field)"
            :field="field"
            :record="record"
            :namespace="namespace"
            :module="module"
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
              <FieldViewer :field="field" :record="record" :namespace="namespace" value-only />
            </div>
            <i v-else class="text-muted">{{ $t('field.noPermission') }}</i>
          </div>
        </div>
      </template>
    </div>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { computed, onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { debounce } from 'lodash'
import { usePageBlockBase } from './usePageBlockBase'
import Wrap from './Wrap/index.js'
import BlockHelpButton from './Shared/BlockHelpButton.vue'
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
  errors: { type: Object, required: false, default: undefined },
})

const emit = defineEmits(['errors'])
const { isProcessing, options, fieldErrors } = usePageBlockBase(props, emit)
const busy = computed(() => isProcessing.value || !props.record)

const fields = computed(() => {
  if (!props.module) return []
  let ff = props.module.fields
  if (options.value.fields && options.value.fields.length > 0) {
    ff = props.module.filterFields(options.value.fields)
  }
  return ff.map(f => {
    const label = f.isSystem ? $t(`system.${f.name}`) : f.label || f.name
    return Object.assign(Object.create(Object.getPrototypeOf(f)), f, { label })
  })
})

const fieldLayoutClass = computed(() => {
  const classes = { default: 'd-flex flex-column', noWrap: 'd-flex gap-2', wrap: 'row g-2' }
  return classes[options.value.recordFieldLayoutOption] || classes.default
})

const columnWrapClass = computed(() => {
  if (options.value.recordFieldLayoutOption === 'noWrap') return 'field-col'
  if (options.value.recordFieldLayoutOption === 'wrap') return 'col-md-6'
  return ''
})

const fieldWidth = computed(() => {
  if (options.value.recordFieldLayoutOption !== 'noWrap') return {}
  return { 'min-width': '20rem' }
})

const horizontal = computed(() => props.block.options.horizontalFieldLayoutEnabled && props.block.options.recordFieldLayoutOption !== 'noWrap')

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
.fixed-corner-container { position: relative; }
</style>
