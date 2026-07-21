<template>
  <Wrap body-class="pt-3 px-3">
    <div v-if="isProcessing" class="d-flex align-items-center justify-content-center h-100">
      <span class="spinner-border" />
    </div>
    <div v-else-if="fieldModule" ref="fieldContainer" class="fixed-corner-container" :class="fieldLayoutClass">
      <button
        class="btn btn-outline-extra-light text-secondary border-0 fixed-corner-button btn-sm"
        title="Ask AI about this record"
        :disabled="editable"
        @click="promptAiChat"
      >
        <font-awesome-icon :icon="['fas', 'brain']" />
      </button>
      <template v-for="field in fields" :key="`${field.fieldID}-${field.name}`">
        <div v-if="canDisplay(field)" :data-test-id="getFieldCypressId(field.label || field.name)" class="field-container mb-3" :class="columnWrapClass" :style="genStyle(options.viewStyle, { addStyle: fieldWidth})">
          <label class="form-label d-flex align-items-center text-primary mb-0">
            <span class="d-flex" style="margin-top: 0.1rem;">{{ field.label || field.name }}</span>
            <c-hint :tooltip="((field.options?.hint || {}).view || '')" />
            <div v-if="!record.deletedAt && options.inlineRecordEditEnabled && isFieldEditable(field)" class="inline-actions ms-1">
              <button
                class="btn btn-outline-extra-light text-secondary border-0 btn-sm"
                title="Edit inline"
                :disabled="editable"
                @click="editInlineField(fieldRecord, field)"
              >
                <font-awesome-icon :icon="['fas', 'pen']" />
              </button>
            </div>
          </label>
          <div class="small text-muted" :class="{ 'mb-1': !!((field.options?.description || {}).view) }">
            {{ (field.options?.description || {}).view }}
          </div>
          <div v-if="field.canReadRecordValue" class="value align-self-center">
            <FieldViewer :field="field" :extra-options="options" :record="fieldRecord" />
          </div>
          <i v-else class="text-muted">{{ $t('field.noPermission') }}</i>
        </div>
      </template>
    </div>
    <BulkEditModal
      v-if="options.inlineRecordEditEnabled && fieldModule"
      :modal-title="$t('record.inlineEdit.modal.title')"
      :namespace="namespace"
      :module="fieldModule"
      :selected-records="inlineEdit.recordIDs"
      :selected-fields="inlineEdit.fields"
      :initial-record="inlineEdit.record"
      :query="inlineEdit.query"
      :allow-add-field="options.inlineRecordEditAllowAddField"
      open-on-select
      @save="onInlineEdit()"
      @close="onInlineEditClose()"
    />
  </Wrap>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { composables } from 'corteza-lib/vue/dist'
import { compose, NoID } from 'corteza-lib/js/dist'
import axios from 'axios'
import { usePageBlockBase } from './usePageBlockBase'
import { useModuleStore } from 'corteza-webapp-compose/src/store/module'
import { useRecordStore } from 'corteza-webapp-compose/src/store/record'
import Wrap from './Wrap/index.js'
import FieldViewer from 'corteza-webapp-compose/src/components/ModuleFields/Viewer'
import BulkEditModal from 'corteza-webapp-compose/src/components/Public/Record/BulkEdit/index.vue'

const { t: $t } = useI18n({ useScope: 'global' })
const { toastErrorHandler } = composables.useToast()
const $ComposeAPI = window.__composeAPI
const $auth = window.__auth
const moduleStore = useModuleStore()
const recordStore = useRecordStore()

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

const { options, genStyle } = usePageBlockBase(props, {})

const referenceRecord = ref(undefined)
const referenceModule = ref(undefined)
const inlineEdit = ref({ fields: [], recordIDs: [], record: {} })
const abortableRequests = ref([])
const evaluating = ref(false)

const fieldWidth = computed(() => {
  if (options.value.recordFieldLayoutOption !== 'noWrap') return {}
  return { 'min-width': '13rem' }
})

const fields = computed(() => {
  if (!fieldModule.value) return []
  if (!options.value.fields || options.value.fields.length === 0) return fieldModule.value.fields
  return fieldModule.value.filterFields(options.value.fields).map(f => ({
    ...f, label: f.isSystem ? $t(`system.${f.name}`) : f.label || f.name,
  }))
})

const fieldLayoutClass = computed(() => {
  const classes = { default: 'd-flex flex-column', noWrap: 'd-flex gap-2', wrap: 'row g-0' }
  return classes[options.value.recordFieldLayoutOption]
})

const fieldModule = computed(() => options.value.referenceField ? referenceModule.value : props.module)

const fieldRecord = computed(() => options.value.referenceField ? referenceRecord.value : props.record)

const isProcessing = computed(() => props.loadingRecord || !fieldRecord.value || evaluating.value)

function canDisplay(field) {
  return field?.canReadRecordValue !== false
}

function isFieldEditable(field) {
  return field?.canUpdateRecordValue !== false
}

function getFieldCypressId(field) {
  return `field-${field.toLowerCase().split(' ').join('-')}`
}

function promptAiChat() {
  const record = props.record
  const page = props.page
  const locale = navigator.language || navigator.languages?.[0] || 'en-US'
  let prompt = record.prompt || page.prompt || props.namespace?.meta?.prompt || ''
  if (!prompt) {
    prompt = locale === 'ru-RU' ? '\u0427\u0442\u043e \u044d\u0442\u043e? \u0417\u0430\u0447\u0435\u043c \u044d\u0442\u043e?' : 'What is this? '
  }
  prompt += '\r\n'
  for (const field of fields.value) {
    if (field.isSystem) continue
    const val = record.values[field.name]
    if (val) prompt += `${field.label}=${val}\r\n`
  }
  window.dispatchEvent(new CustomEvent('show-chat-modal', { detail: { namespace: page.namespaceID, module: page.moduleID, prompt } }))
}

function editInlineField(record, field) {
  inlineEdit.value = { fields: [field.name], record: record.clone(), recordIDs: [record.recordID], query: `recordID = ${record.recordID}` }
}

function onInlineEdit() { inlineEdit.value = { fields: [], recordIDs: [], record: {} } }
function onInlineEditClose() { inlineEdit.value = { fields: [], record: {}, recordIDs: [] } }

function fetchReferenceModule(moduleID) {
  if (!moduleID) { referenceModule.value = undefined; return }
  moduleStore.findByID({ namespace: props.namespace.namespaceID, moduleID })
    .then(mod => {
      referenceModule.value = new compose.Module({ ...mod })
      if (options.value.referenceField) loadRecord(referenceModule.value)
    })
}

function loadRecord(mod) {
  if (!mod) return
  const { namespaceID, moduleID } = mod
  const field = props.module.fields.find(({ fieldID }) => fieldID === options.value.referenceField)
  const recordID = props.record.values[field?.name]
  if (!recordID || !field || field.isMulti) {
    referenceRecord.value = new compose.Record(fieldModule.value, {})
    return
  }
  const { response, cancel } = $ComposeAPI.recordReadCancellable({ namespaceID, moduleID, recordID })
  abortableRequests.value.push(cancel)
  response()
    .then(rec => { referenceRecord.value = new compose.Record(fieldModule.value, { ...rec }) })
    .catch(e => {
      if (!axios.isCancel(e)) {
        referenceRecord.value = new compose.Record(fieldModule.value, {})
        toastErrorHandler($t('notification.record.loadFailed'))(e)
      }
    })
}

watch(() => props.loadingRecord, (loadingRecord) => {
  const { recordID } = props.record || {}
  if (!recordID || loadingRecord) return
  evaluating.value = true
  Promise.all([]).finally(() => { evaluating.value = false })
  if (options.value.referenceModuleID) fetchReferenceModule(options.value.referenceModuleID)
}, { immediate: true })

watch(() => props.record?.values, (newValues = {}) => {
  if (options.value.referenceField) {
    const oldValue = referenceRecord.value?.recordID
    const newValue = newValues[options.value.referenceField]
    if (oldValue !== newValue) loadRecord(referenceModule.value)
  }
}, { deep: true })

onBeforeUnmount(() => {
  abortableRequests.value.forEach(c => c())
})
</script>

<style scoped>
.field-col > * { margin-left: 1rem; margin-right: 1rem; }
.fixed-corner-container { position: relative; }
.fixed-corner-button {
  position: absolute; top: 3px; right: 0px;
  transform: translateY(-50%); z-index: 10;
}
</style>
