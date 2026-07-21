<template>
  <div class="tab-pane">
    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('module') }}</label>
      <c-input-select
        v-model="block.options.moduleID"
        :options="filterModulesByRecord"
        label="name"
        :reduce="m => m.moduleID"
        :placeholder="$t('comment.module.placeholder')"
        default-value="0"
        required
      />
    </div>

    <div v-if="selectedModule">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('recordList.record.prefilterLabel') }}</label>
        <c-input-expression
          v-model.trim="block.options.filter"
          min-height="3.688rem"
          :suggestion-params="recordAutoCompleteParams"
          :placeholder="$t('recordList.record.prefilterPlaceholder')"
        />

        <small class="text-muted">
          {{ $t('recordList.record.prefilterFootnote') }}
          <code>${record.values.fieldName}</code>
          <code>${recordID}</code>
          <code>${ownerID}</code>
          <span><code>${userID}</code>, <code>${user.name}</code></span>
        </small>
      </div>

      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('comment.titleField.label') }}</label>
            <small class="form-text d-block">{{ $t('comment.titleField.footnote') }}</small>
            <c-input-select
              v-model="block.options.titleField"
              :options="selectedModuleFieldsByType('String')"
              :get-option-label="f => `${f.label || f.name} (${f.kind})`"
              :reduce="f => f.name"
              :placeholder="$t('label.none')"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('comment.contentField.label') }}</label>
            <small class="form-text d-block">{{ $t('comment.contentField.footnote') }}</small>
            <c-input-select
              v-model="block.options.contentField"
              :options="selectedModuleFieldsByType('String')"
              :get-option-label="f => `${f.label || f.name} (${f.kind})`"
              :reduce="f => f.name"
              :placeholder="$t('label.none')"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('comment.replyField.label') }}</label>
            <small class="form-text d-block">{{ $t('comment.replyField.footnote') }}</small>
            <c-input-select
              v-model="block.options.replyField"
              :options="selectedModuleFieldsByType('Record')"
              :get-option-label="f => `${f.label || f.name} (${f.kind})`"
              :reduce="f => f.name"
              :placeholder="$t('label.none')"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('comment.referenceField.label') }}</label>
            <small class="form-text d-block">{{ $t('comment.referenceField.footnote') }}</small>
            <c-input-select
              v-model="block.options.referenceField"
              :options="selectedModuleFieldsByType('Record')"
              :get-option-label="f => `${f.label || f.name} (${f.kind})`"
              :reduce="f => f.name"
              :placeholder="$t('label.none')"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('comment.attachmentField.label') }}</label>
            <small class="form-text d-block">{{ $t('comment.attachmentField.footnote') }}</small>
            <c-input-select
              v-model="block.options.attachmentField"
              :options="selectedModuleFieldsByType('File', { includeMulti: true })"
              :get-option-label="f => `${f.label || f.name} (${f.kind})`"
              :reduce="f => f.name"
              :placeholder="$t('label.none')"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('comment.sortDirection.label') }}</label>
            <small class="form-text d-block">{{ $t('comment.sortDirection.footnote') }}</small>
            <c-input-select
              v-model="block.options.sortDirection"
              :options="sortDirections"
              label="label"
              :clearable="false"
              :reduce="o => o.value"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('comment.reactionsField.label') }}</label>
            <small class="form-text d-block">{{ $t('comment.reactionsField.footnote') }}</small>
            <c-input-select
              v-model="block.options.reactionsField"
              :options="selectedModuleFieldsByType('String')"
              :get-option-label="f => `${f.label || f.name} (${f.kind})`"
              :reduce="f => f.name"
              :placeholder="$t('label.none')"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, inject } from 'vue'
import { usePageBlockBase } from '../usePageBlockBase'
import { useStore } from '../../../store'
import { NoID } from 'corteza-lib/js/dist'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

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

const emit = defineEmits(['errors'])
const store = useStore()
const $auth = inject('$auth')

const { options } = usePageBlockBase(props, emit)

const modules = computed(() => store.module.set)
const record = computed(() => props.record)

const isRecordPage = computed(() => props.page && props.page.moduleID !== NoID)

const sortDirections = ref([
  { label: t('comment.sortDirection.asc'), value: 'asc' },
  { label: t('comment.sortDirection.desc'), value: 'desc' },
])

const filterModulesByRecord = computed(() => {
  if (record.value) {
    return modules.value.filter(module => {
      return module.fields.some(f => {
        if (f.kind === 'Record') {
          if (f.options.moduleID === options.value.moduleID) {
            return false
          }
        }
        return true
      })
    })
  }
  return modules.value
})

const selectedModule = computed(() => {
  return modules.value.find(m => m.moduleID === options.value.moduleID)
})

const selectedModuleFields = computed(() => {
  if (selectedModule.value) {
    return [...selectedModule.value.fields].sort((a, b) => a.label.localeCompare(b.label))
  }
  return []
})

const allFields = computed(() => {
  if (options.value.moduleID) {
    return [
      ...selectedModuleFields.value,
      ...selectedModule.value.systemFields().map(sf => {
        sf.label = t(`system.${sf.name}`)
        return sf
      }),
    ]
  }
  return []
})

const recordAutoCompleteParams = computed(() => {
  return processRecordAutoCompleteParams({ module: selectedModule.value, operators: true })
})

function processRecordAutoCompleteParams ({ module: mod, operators = false } = {}) {
  const { fields = [] } = mod || {}
  const moduleFields = fields.map(({ name }) => name)
  const userProperties = ($auth.user?.properties?.()) || []

  const recordSuggestions = isRecordPage.value && record.value
    ? [
        ...(['ownerID', 'recordID'].map(value => ({ interpolate: true, value }))),
        {
          interpolate: true,
          value: 'record',
          properties: [
            { value: 'values', properties: Object.keys(record.value.values) || [] },
            ...(record.value.properties || []),
          ],
        },
      ]
    : []

  return [
    ...recordSuggestions,
    ...(operators ? ['AND', 'OR'] : []),
    { interpolate: true, value: 'userID' },
    { interpolate: true, value: 'user', properties: userProperties },
    ...moduleFields,
  ]
}

function selectedModuleFieldsByType (type, { includeMulti = false } = {}) {
  return (selectedModuleFields.value || []).filter((f) => {
    if (f.kind !== type) return false
    if (!includeMulti && f.isMulti) return false
    return true
  })
}

if (!options.value.sortDirection) {
  options.value.sortDirection = 'desc'
}
</script>

<style lang="scss" scoped>
.fields {
  height: 150px;
  overflow-y: auto;
  cursor: default;
}
</style>
