<template>
  <Wrap :block="block" :record="record" :loading-record="loadingRecord" :magnified="magnified" body-class="pt-3 px-3">
    <div v-if="isProcessing" class="d-flex align-items-center justify-content-center h-100">
      <span class="spinner-border" />
    </div>
    <div
      v-else-if="fieldModule"
      ref="fieldContainer"
      class="fixed-corner-container rb"
      :class="[fieldLayoutClass, densityClass]"
    >
      <button
        v-if="!block.options?.hideBrainButton"
        class="btn btn-outline-extra-light text-secondary border-0 fixed-corner-button btn-sm"
        title="Ask AI about this record"
        :disabled="editable"
        @click="promptAiChat"
      >
        <font-awesome-icon :icon="['fas', 'brain']" />
      </button>

      <!-- Title / subtitle / badges -->
      <div
        v-if="header.title || header.subtitle || header.badges.length"
        class="rb-header mb-3 pe-4"
      >
        <div v-if="header.title" class="rb-title-wrap d-flex align-items-start gap-1">
          <div class="rb-title flex-grow-1 min-w-0">
            <template v-if="header.title.canReadRecordValue">
              <template v-if="isFieldEmpty(header.title)">
                <span v-if="showEmptyPlaceholder" class="rb-empty">{{ $t('record.emptyPlaceholder') }}</span>
              </template>
              <FieldViewer
                v-else
                :field="header.title"
                :extra-options="options"
                :record="fieldRecord"
                :namespace="namespace"
              />
            </template>
            <i v-else class="text-muted">{{ $t('field.noPermission') }}</i>
          </div>
          <div
            v-if="!record.deletedAt && options.inlineRecordEditEnabled && isFieldEditable(header.title)"
            class="inline-actions"
          >
            <button
              class="btn btn-outline-extra-light text-secondary border-0 btn-sm"
              :title="$t('record.inlineEdit.button.title')"
              :disabled="editable"
              @click="editInlineField(fieldRecord, header.title)"
            >
              <font-awesome-icon :icon="['fas', 'pen']" />
            </button>
          </div>
        </div>
        <div v-if="header.subtitle" class="rb-subtitle text-muted mt-1">
          <template v-if="header.subtitle.canReadRecordValue">
            <template v-if="isFieldEmpty(header.subtitle)">
              <span v-if="showEmptyPlaceholder" class="rb-empty">{{ $t('record.emptyPlaceholder') }}</span>
            </template>
            <FieldViewer
              v-else
              :field="header.subtitle"
              :extra-options="options"
              :record="fieldRecord"
              :namespace="namespace"
            />
          </template>
          <i v-else class="text-muted">{{ $t('field.noPermission') }}</i>
        </div>
        <div v-if="header.badges.length" class="rb-badges d-flex flex-wrap gap-2 mt-2">
          <div
            v-for="field in header.badges"
            :key="`badge-${field.name}`"
            class="rb-badge"
          >
            <span v-if="field.label" class="rb-badge-label">{{ field.label }}</span>
            <span class="rb-badge-value">
              <template v-if="field.canReadRecordValue">
                <template v-if="isFieldEmpty(field)">
                  <span v-if="showEmptyPlaceholder" class="rb-empty">{{ $t('record.emptyPlaceholder') }}</span>
                </template>
                <FieldViewer
                  v-else
                  :field="field"
                  :extra-options="options"
                  :record="fieldRecord"
                  :namespace="namespace"
                />
              </template>
              <i v-else class="text-muted">{{ $t('field.noPermission') }}</i>
            </span>
          </div>
        </div>
      </div>

      <!-- Meta strip -->
      <div v-if="metaFields.length" class="rb-meta mb-3">
        <div
          v-for="field in metaFields"
          :key="`meta-${field.name}`"
          class="rb-meta-item"
          :data-test-id="getFieldCypressId(field.label || field.name)"
        >
          <span class="rb-meta-label">
            {{ field.label || field.name }}
            <c-hint :tooltip="((field.options?.hint || {}).view || '')" />
          </span>
          <span class="rb-meta-value">
            <template v-if="field.canReadRecordValue">
              <template v-if="isFieldEmpty(field)">
                <span v-if="showEmptyPlaceholder" class="rb-empty">{{ $t('record.emptyPlaceholder') }}</span>
              </template>
              <FieldViewer
                v-else
                :field="field"
                :extra-options="options"
                :record="fieldRecord"
                :namespace="namespace"
              />
            </template>
            <i v-else class="text-muted">{{ $t('field.noPermission') }}</i>
          </span>
          <button
            v-if="!record.deletedAt && options.inlineRecordEditEnabled && isFieldEditable(field)"
            class="btn btn-outline-extra-light text-secondary border-0 btn-sm rb-meta-edit"
            :title="$t('record.inlineEdit.button.title')"
            :disabled="editable"
            @click="editInlineField(fieldRecord, field)"
          >
            <font-awesome-icon :icon="['fas', 'pen']" />
          </button>
        </div>
      </div>

      <!-- Sections / body fields -->
      <template v-for="(section, sIdx) in displaySections" :key="`section-${sIdx}`">
        <div v-if="section.fields.length" class="rb-section" :class="{ 'mb-3': sIdx < displaySections.length - 1 }">
          <h6 v-if="section.title" class="rb-section-title text-muted text-uppercase">
            {{ section.title }}
          </h6>
          <div :class="sectionLayoutClass">
            <div
              v-for="field in section.fields"
              :key="`${field.fieldID}-${field.name}`"
              :data-test-id="getFieldCypressId(field.label || field.name)"
              class="field-container"
              :class="[columnWrapClass, fieldRoleClass(field), { 'mb-3': options.density !== 'compact', 'mb-2': options.density === 'compact' }]"
              :style="genStyle(options.viewStyle, { addStyle: fieldWidth })"
            >
              <div v-if="horizontal" class="row g-2">
                <div class="col-md-6 col-xl-5">
                  <label class="d-flex align-items-center mb-0 form-label rb-label">
                    <span class="d-flex" style="margin-top: 0.1rem;">{{ field.label || field.name }}</span>
                    <c-hint :tooltip="((field.options?.hint || {}).view || '')" />
                    <div
                      v-if="!record.deletedAt && options.inlineRecordEditEnabled && isFieldEditable(field)"
                      class="inline-actions ms-1"
                    >
                      <button
                        class="btn btn-outline-extra-light text-secondary border-0 btn-sm"
                        :title="$t('record.inlineEdit.button.title')"
                        :disabled="editable"
                        @click="editInlineField(fieldRecord, field)"
                      >
                        <font-awesome-icon :icon="['fas', 'pen']" />
                      </button>
                    </div>
                  </label>
                  <div
                    class="small text-muted"
                    :class="{ 'mb-1': !!((field.options?.description || {}).view) }"
                  >
                    {{ (field.options?.description || {}).view }}
                  </div>
                </div>
                <div class="col-md-6 col-xl-7 d-flex align-items-center">
                  <div v-if="field.canReadRecordValue" class="value w-100">
                    <template v-if="isFieldEmpty(field)">
                      <span v-if="showEmptyPlaceholder" class="rb-empty">{{ $t('record.emptyPlaceholder') }}</span>
                    </template>
                    <FieldViewer
                      v-else
                      :field="field"
                      :extra-options="options"
                      :record="fieldRecord"
                      :namespace="namespace"
                    />
                  </div>
                  <i v-else class="text-muted">{{ $t('field.noPermission') }}</i>
                </div>
              </div>
              <template v-else>
                <label class="form-label d-flex align-items-center mb-0 rb-label">
                  <span class="d-flex" style="margin-top: 0.1rem;">{{ field.label || field.name }}</span>
                  <c-hint :tooltip="((field.options?.hint || {}).view || '')" />
                  <div
                    v-if="!record.deletedAt && options.inlineRecordEditEnabled && isFieldEditable(field)"
                    class="inline-actions ms-1"
                  >
                    <button
                      class="btn btn-outline-extra-light text-secondary border-0 btn-sm"
                      :title="$t('record.inlineEdit.button.title')"
                      :disabled="editable"
                      @click="editInlineField(fieldRecord, field)"
                    >
                      <font-awesome-icon :icon="['fas', 'pen']" />
                    </button>
                  </div>
                </label>
                <div
                  class="small text-muted"
                  :class="{ 'mb-1': !!((field.options?.description || {}).view) }"
                >
                  {{ (field.options?.description || {}).view }}
                </div>
                <div v-if="field.canReadRecordValue" class="value align-self-center">
                  <template v-if="isFieldEmpty(field)">
                    <span v-if="showEmptyPlaceholder" class="rb-empty">{{ $t('record.emptyPlaceholder') }}</span>
                  </template>
                  <FieldViewer
                    v-else
                    :field="field"
                    :extra-options="options"
                    :record="fieldRecord"
                    :namespace="namespace"
                  />
                </div>
                <i v-else class="text-muted">{{ $t('field.noPermission') }}</i>
              </template>
            </div>
          </div>
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
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { composables } from 'corteza-lib/vue/dist'
import { compose } from 'corteza-lib/js/dist'
import axios from 'axios'
import { usePageBlockBase } from './usePageBlockBase'
import { useModuleStore } from 'corteza-webapp-compose/src/store/module'
import Wrap from './Wrap/index.js'
import FieldViewer from 'corteza-webapp-compose/src/components/ModuleFields/Viewer'
import BulkEditModal from 'corteza-webapp-compose/src/components/Public/Record/BulkEdit/index.vue'
import { isUserWritableField } from 'corteza-webapp-compose/src/lib/field-editable'

const { t: $t } = useI18n({ useScope: 'global' })
const { toastErrorHandler } = composables.useToast()
const $ComposeAPI = window.__composeAPI
const moduleStore = useModuleStore()

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

const horizontal = computed(() => options.value.horizontalFieldLayoutEnabled && options.value.recordFieldLayoutOption !== 'noWrap')

const densityClass = computed(() =>
  options.value.density === 'compact' ? 'rb-density-compact' : 'rb-density-comfortable',
)

const showEmptyPlaceholder = computed(() => options.value.showEmptyPlaceholder !== false)

const fields = computed(() => {
  if (!fieldModule.value) return []
  let ff = fieldModule.value.fields
  if (options.value.fields && options.value.fields.length > 0) {
    ff = fieldModule.value.filterFields(options.value.fields)
  }
  return ff.map(f => {
    const label = f.isSystem ? $t(`system.${f.name}`) : f.label || f.name
    return Object.assign(Object.create(Object.getPrototypeOf(f)), f, { label })
  })
})

const visibleFields = computed(() => {
  return fields.value.filter(f => {
    if (!canDisplay(f)) return false
    if (options.value.hideEmptyFields && isFieldEmpty(f)) return false
    return true
  })
})

function fieldRole (field) {
  const roles = options.value.fieldRoles || {}
  return roles[field.name] || 'default'
}

function fieldRoleClass (field) {
  const role = fieldRole(field)
  if (role === 'body') return 'rb-field-body col-12'
  if (options.value.recordFieldLayoutOption === 'wrap' && role === 'default') return 'col-md-6'
  return ''
}

const header = computed(() => {
  const list = visibleFields.value
  const title = list.find(f => fieldRole(f) === 'title') || null
  const subtitle = list.find(f => fieldRole(f) === 'subtitle') || null
  const badges = list.filter(f => fieldRole(f) === 'badge')
  return { title, subtitle, badges }
})

const metaFields = computed(() => visibleFields.value.filter(f => fieldRole(f) === 'meta'))

const bodyFields = computed(() => {
  const headerRoles = new Set(['title', 'subtitle', 'badge', 'meta'])
  return visibleFields.value.filter(f => !headerRoles.has(fieldRole(f)))
})

const displaySections = computed(() => {
  const body = bodyFields.value
  const sections = (options.value.sections || []).filter(s => s && (s.title || (s.fields && s.fields.length)))
  if (!sections.length) {
    return [{ title: '', fields: body }]
  }

  const used = new Set()
  const result = []
  for (const section of sections) {
    const names = new Set(section.fields || [])
    const sectionFields = body.filter(f => names.has(f.name))
    sectionFields.forEach(f => used.add(f.name))
    if (sectionFields.length || section.title) {
      result.push({ title: section.title || '', fields: sectionFields })
    }
  }

  const rest = body.filter(f => !used.has(f.name))
  if (rest.length) {
    result.push({ title: '', fields: rest })
  }
  return result.length ? result : [{ title: '', fields: body }]
})

const fieldLayoutClass = computed(() => {
  // Outer container stays column; section inner uses layout
  return 'd-flex flex-column'
})

const sectionLayoutClass = computed(() => {
  const classes = { default: 'd-flex flex-column', noWrap: 'd-flex gap-2', wrap: 'row g-2' }
  return classes[options.value.recordFieldLayoutOption] || classes.default
})

const columnWrapClass = computed(() => {
  if (options.value.recordFieldLayoutOption === 'noWrap') return 'field-col'
  return ''
})

const fieldModule = computed(() => options.value.referenceField ? referenceModule.value : props.module)

const fieldRecord = computed(() => options.value.referenceField ? referenceRecord.value : props.record)

const isProcessing = computed(() => props.loadingRecord || !fieldRecord.value || evaluating.value)

function canDisplay (field) {
  return field?.canReadRecordValue !== false
}

function isFieldEditable (field) {
  return isUserWritableField(field)
}

function getFieldValue (field) {
  const rec = fieldRecord.value
  if (!rec || !field) return undefined
  if (field.isSystem) return rec[field.name]
  return rec.values?.[field.name]
}

function isFieldEmpty (field) {
  const v = getFieldValue(field)
  if (v === null || v === undefined) return true
  if (typeof v === 'string' && v.trim() === '') return true
  if (Array.isArray(v)) {
    return v.length === 0 || v.every(x => x === null || x === undefined || x === '')
  }
  // Bool '0' / false and number 0 are meaningful values
  return false
}

function getFieldCypressId (field) {
  return `field-${field.toLowerCase().split(' ').join('-')}`
}

function promptAiChat () {
  const record = props.record
  const page = props.page
  const locale = navigator.language || navigator.languages?.[0] || 'en-US'
  let prompt = record.prompt || page.prompt || props.namespace?.meta?.prompt || ''
  if (!prompt) {
    prompt = locale === 'ru-RU' ? 'Что это? Зачем это?' : 'What is this? '
  }
  prompt += '\r\n'
  for (const field of fields.value) {
    if (field.isSystem) continue
    const val = record.values[field.name]
    if (val) prompt += `${field.label}=${val}\r\n`
  }
  window.dispatchEvent(new CustomEvent('show-chat-modal', { detail: { namespace: page.namespaceID, module: page.moduleID, prompt } }))
}

function editInlineField (record, field) {
  inlineEdit.value = { fields: [field.name], record: record.clone(), recordIDs: [record.recordID], query: `recordID = ${record.recordID}` }
}

function onInlineEdit () { inlineEdit.value = { fields: [], recordIDs: [], record: {} } }
function onInlineEditClose () { inlineEdit.value = { fields: [], record: {}, recordIDs: [] } }

function fetchReferenceModule (moduleID) {
  if (!moduleID) { referenceModule.value = undefined; return }
  moduleStore.findByID({ namespace: props.namespace.namespaceID, moduleID })
    .then(mod => {
      referenceModule.value = new compose.Module({ ...mod })
      if (options.value.referenceField) loadRecord(referenceModule.value)
    })
}

function loadRecord (mod) {
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

.rb-label {
  color: var(--bs-secondary-color, #6c757d);
  font-size: 0.8rem;
  font-weight: 500;
}

.rb-title {
  font-size: 1.5rem;
  font-weight: 600;
  line-height: 1.3;
  color: var(--bs-body-color, inherit);
}

.rb-subtitle {
  font-size: 0.95rem;
  line-height: 1.4;
}

.rb-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.65rem;
  border-radius: 0.375rem;
  background: var(--bs-tertiary-bg, #f8f9fa);
  border: 1px solid var(--bs-border-color, #dee2e6);
  font-size: 0.8125rem;
}

.rb-badge-label {
  color: var(--bs-secondary-color, #6c757d);
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.rb-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1.5rem;
  padding: 0.65rem 0;
  border-top: 1px solid var(--bs-border-color-translucent, rgba(0, 0, 0, 0.08));
  border-bottom: 1px solid var(--bs-border-color-translucent, rgba(0, 0, 0, 0.08));
}

.rb-meta-item {
  display: inline-flex;
  align-items: baseline;
  gap: 0.4rem;
  min-width: 0;
}

.rb-meta-label {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--bs-secondary-color, #6c757d);
  white-space: nowrap;
}

.rb-meta-value {
  font-size: 0.875rem;
  min-width: 0;
}

.rb-meta-edit {
  padding: 0 0.25rem !important;
  opacity: 0.5;
}
.rb-meta-item:hover .rb-meta-edit { opacity: 1; }

.rb-section-title {
  font-size: 0.7rem;
  letter-spacing: 0.06em;
  font-weight: 600;
  margin-bottom: 0.75rem;
}

.rb-field-body {
  width: 100%;
}

.rb-empty {
  color: var(--bs-secondary-color, #adb5bd);
}

.rb-density-compact .rb-title { font-size: 1.25rem; }
.rb-density-compact .rb-section-title { margin-bottom: 0.5rem; }
.rb-density-compact .rb-meta { gap: 0.5rem 1rem; padding: 0.4rem 0; }
.rb-density-compact .rb-label { font-size: 0.75rem; }
</style>
