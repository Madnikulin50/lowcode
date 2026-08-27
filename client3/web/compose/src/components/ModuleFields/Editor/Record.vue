<template>
  <div class="mb-3" :class="formGroupStyleClasses">
    <div v-if="!valueOnly" :class="labelColClass">
      <div class="d-flex align-items-center text-primary p-0">
        <span :title="label" class="d-inline-block mw-100">{{ label }}</span>
        <c-hint :tooltip="hint" />
        <slot name="tools" />
      </div>
      <div class="small text-muted" :class="{ 'mb-1': description }">{{ description }}</div>
    </div>
    <div :class="contentColClass">
    <multi
      v-if="field.isMulti"
      v-model:value="value"
      :errors="errors"
      :single-input="field.options.selectType !== 'each'"
      :show-list="field.options.selectType !== 'multiple'"
    >
      <template #single>
        <div class="input-group input-group-sm d-flex w-100">
          <c-input-select
            v-if="field.options.selectType === 'multiple'"
            v-model="multipleSelected"
            :options="options"
            :get-option-key="getOptionKey"
            :get-option-label="getOptionLabel"
            :disabled="!module"
            :loading="processing"
            :clearable="false"
            :filterable="false"
            :searchable="searchable"
            :selectable="isSelectable"
            :placeholder="placeholder"
            multiple
            @search="search"
          >
            <template #option="option">
              <FieldViewer v-if="labelField && option.values[labelField.name]" :field="labelField" :record="optionAsRecord(option)" :namespace="namespace" disable-click value-only />
              <template v-else>{{ option.recordID }}</template>
            </template>
            <template #selected-option="option">
              <FieldViewer v-if="labelField && getRecordByID(option.label).values[labelField.name]" :field="labelField" :record="getRecordByID(option.label)" :namespace="namespace" disable-click value-only />
              <template v-else>{{ option.label }}</template>
            </template>
            <template #list-footer>
              <Pagination v-if="showPagination" :has-prev-page="hasPrevPage" :has-next-page="hasNextPage" @prev="goToPage(false)" @next="goToPage(true)" />
            </template>
          </c-input-select>

          <c-input-select
            v-else
            ref="singleSelect"
            :options="options"
            :get-option-key="getOptionKey"
            :get-option-label="getOptionLabel"
            :disabled="!module"
            :loading="processing"
            :clearable="false"
            :filterable="false"
            :searchable="searchable"
            :selectable="isSelectable"
            :placeholder="placeholder"
            @input="selectChange($event)"
            @search="search"
          >
            <template #option="option">
              <FieldViewer v-if="labelField && option.values[labelField.name]" :field="labelField" :record="optionAsRecord(option)" :namespace="namespace" disable-click value-only />
              <template v-else>{{ option.recordID }}</template>
            </template>
            <template #list-footer>
              <Pagination v-if="showPagination" :has-prev-page="hasPrevPage" :has-next-page="hasNextPage" @prev="goToPage(false)" @next="goToPage(true)" />
            </template>
          </c-input-select>

          <button v-if="canAddRecordThroughSelectField" class="btn btn-outline-secondary d-flex align-items-center" :title="t('kind.record.tooltip.addRecord')" @click="addRecordThroughRecordSelectField()">
            <font-awesome-icon :icon="['fas', 'plus']" class="text-primary" />
          </button>
        </div>
      </template>

      <template #default="ctx">
        <div v-if="field.options.selectType === 'each'" class="input-group input-group-sm d-flex w-100">
          <c-input-select
            :options="options"
            :get-option-key="getOptionKey"
            :get-option-label="getOptionLabel"
            :disabled="!module"
            :loading="processing"
            :clearable="false"
            :filterable="false"
            :searchable="searchable"
            :selectable="isSelectable"
            :placeholder="placeholder"
            :value="getRecord(ctx.index)"
            @input="setRecord($event, ctx.index)"
            @search="search"
          >
            <template #option="option">
              <FieldViewer v-if="labelField && option.values[labelField.name]" :field="labelField" :record="optionAsRecord(option)" :namespace="namespace" disable-click value-only />
              <template v-else>{{ option.recordID }}</template>
            </template>
            <template #selected-option="option">
              <FieldViewer v-if="labelField && getRecordByID(option.label).values[labelField.name]" :field="labelField" :record="getRecordByID(option.label)" :namespace="namespace" disable-click value-only />
              <template v-else>{{ option.label }}</template>
            </template>
            <template #list-footer>
              <Pagination v-if="showPagination" :has-prev-page="hasPrevPage" :has-next-page="hasNextPage" @prev="goToPage(false)" @next="goToPage(true)" />
            </template>
          </c-input-select>

          <button v-if="canAddRecordThroughSelectField" class="btn btn-outline-secondary d-flex align-items-center" :title="t('kind.record.tooltip.addRecord')" @click="addRecordThroughRecordSelectField()">
            <font-awesome-icon :icon="['fas', 'plus']" class="text-primary" />
          </button>
        </div>

        <FieldViewer v-else :field="labelField" :record="getRecordByID(multipleSelected[ctx.index])" :namespace="namespace" disable-click value-only />
      </template>
    </multi>

    <template v-else>
      <div class="input-group input-group-sm">
        <c-input-select
          v-model="selected"
          :options="options"
          :get-option-key="getOptionKey"
          :get-option-label="getOptionLabel"
          :disabled="!module"
          :loading="processing"
          :placeholder="placeholder"
          :filterable="false"
          :searchable="searchable"
          :selectable="isSelectable"
          @search="search"
        >
          <template #option="option">
            <FieldViewer v-if="labelField && option.values[labelField.name]" :field="labelField" :record="optionAsRecord(option)" :namespace="namespace" disable-click value-only />
            <template v-else>{{ option.recordID }}</template>
          </template>
          <template #selected-option>
            <FieldViewer v-if="labelField && getRecordByID(selected).values[labelField.name]" :field="labelField" :record="getRecordByID(selected)" :namespace="namespace" disable-click value-only />
            <template v-else>{{ selected }}</template>
          </template>
          <template #list-footer>
            <Pagination v-if="showPagination" :has-prev-page="hasPrevPage" :has-next-page="hasNextPage" @prev="goToPage(false)" @next="goToPage(true)" />
          </template>
        </c-input-select>

        <button v-if="canAddRecordThroughSelectField" class="btn btn-outline-secondary d-flex align-items-center" :title="t('kind.record.tooltip.addRecord')" @click="addRecordThroughRecordSelectField()">
          <font-awesome-icon :icon="['fas', 'plus']" class="text-primary" />
        </button>
      </div>
      <FieldErrors :errors="errors" />
    </template>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { ref, shallowRef, computed, inject, watch, onBeforeUnmount, onMounted, markRaw } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { debounce } from 'lodash'
import axios from 'axios'
import { compose, NoID } from 'corteza-lib/js/dist'
import { queryToFilter, evalPrefilterOrSkip, isFieldInFilter } from 'corteza-webapp-compose/src/lib/record-filter'
import { useModuleStore } from 'corteza-webapp-compose/src/stores/module'
import { useRecordStore } from 'corteza-webapp-compose/src/stores/record'
import { usePageStore } from 'corteza-webapp-compose/src/stores/page'
import { recordCreateLocation } from 'corteza-webapp-compose/src/lib/record-create-nav'
import { useEditorBase } from './base'
import Pagination from '../Common/Pagination.vue'
import FieldViewer from '../Viewer'
import FieldErrors from '../errors'
import multi from './multi'

const props = defineProps({
  namespace: { type: Object, required: true },
  field: { type: Object, required: true },
  record: { type: Object, required: true },
  errors: { type: Object, required: true },
  valueOnly: { type: Boolean, default: false },
  horizontal: { type: Boolean, default: false },
  extraOptions: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['change', 'update:preventPopoverClose'])

const router = useRouter()
const { t } = useI18n({ useScope: 'global', messages: {} })
const { value, formGroupStyleClasses, labelColClass, contentColClass, label, hint, description } = useEditorBase(props, emit)

const $ComposeAPI = inject('$ComposeAPI')
const $auth = inject('$auth')

const moduleStore = useModuleStore()
const recordStore = useRecordStore()
const pageStore = usePageStore()

const processing = ref(false)
const cancelRequest = ref(null)
const query = ref('')
const records = shallowRef([])
const singleSelect = ref(null)
const filter = ref({ query: '', sort: '', limit: 10, pageCursor: '', prevPage: '', nextPage: '' })

const labelField = computed(() => {
  if (!props.field.options.labelField || !module.value) return undefined
  return module.value.fields.find(({ name }) => name === props.field.options.labelField)
})

const options = computed(() => records.value)

const module = computed(() => {
  if (props.field.options.moduleID !== NoID) return moduleStore.getByID(props.field.options.moduleID)
  return undefined
})

const searchable = computed(() => !props.field.options.recordLabelField)

const placeholder = computed(() => searchable.value ? t('kind.record.suggestionPlaceholder') : t('kind.select.placeholder'))

const multipleSelected = computed({
  get () { return value.value },
  set (v) { value.value = v.map(v => typeof v === 'string' ? v : v.recordID) },
})

const selected = computed({
  get () { return getRecord() },
  set (v) { setRecord(v) },
})

const showPagination = computed(() => hasPrevPage.value || hasNextPage.value)
const hasPrevPage = computed(() => !!filter.value.prevPage)
const hasNextPage = computed(() => !!filter.value.nextPage)

const canAddRecordThroughSelectField = computed(() => {
  if (!props.extraOptions.recordSelectorShowAddRecordButton || module.value === undefined) return false
  return !!getRecordSelectorPage().page?.pageID && module.value.canCreateRecord
})

watch(() => filter.value.pageCursor, (pageCursor) => {
  if (pageCursor) fetchPrefiltered(filter.value)
})

onBeforeUnmount(() => {
  setDefaultValues()
  destroyEvents()
})

onMounted(() => {
  createEvents()
  loadLatest()
  formatRecordValues(value.value)
})

function createEvents () {
  window.addEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.addEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.addEventListener('refetch-records', loadLatest)
}

function destroyEvents () {
  window.removeEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.removeEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.removeEventListener('refetch-records', loadLatest)
}

function refreshOnRelatedRecordsUpdate ({ detail: { moduleID } = {} } = {}) {
  if (props.field.options.moduleID === moduleID) loadLatest()
}

function refetchOnPrefilterValueChange ({ detail: { fieldName } = {} } = {}) {
  const { prefilter } = props.field.options
  if (isFieldInFilter(fieldName, prefilter)) {
    fetchPrefiltered({ namespaceID: props.namespace.namespaceID, moduleID: props.field.options.moduleID })
  }
}

function optionAsRecord (option) {
  if (!option) return option
  if (option instanceof compose.Record) return option
  if (!module.value) return option
  const found = records.value.find(r => r.recordID === option.recordID)
  if (found) return found
  return markRaw(new compose.Record(module.value, option))
}

function getRecordByID (recordID) {
  const id = typeof recordID === 'object' && recordID ? recordID.recordID : recordID
  if (!module.value) return { values: {} }
  if (!id) return markRaw(new compose.Record(module.value))
  const found = recordStore.findByID(id)
  if (!found) return markRaw(new compose.Record(module.value))
  return markRaw(new compose.Record(module.value, found))
}

function getRecord (index = undefined) {
  return index !== undefined ? (value.value || [])[index] : value.value
}

function setRecord ({ recordID } = {}, index = undefined) {
  const crtValue = index !== undefined ? (value.value || [])[index] : value.value

  if (recordID !== crtValue) {
    if (recordID) {
      if (index !== undefined) {
        const arr = [...(value.value || [])]
        arr.splice(index, 1, recordID)
        value.value = arr
      } else {
        value.value = recordID
      }
    } else {
      if (index !== undefined) {
        const arr = [...(value.value || [])]
        arr.splice(index, 1)
        value.value = arr
      } else {
        value.value = undefined
      }
    }
  }
  emit('change', value.value)
}

function isSelectable ({ recordID } = {}) {
  if (!recordID) return false
  if (props.field.isMulti) return !props.field.options.isUniqueMultiValue || !((value.value || [])).includes(recordID)
  return value.value !== recordID
}

const search = debounce(function (q = '') {
  if (q !== query.value) {
    query.value = q
    filter.value.pageCursor = undefined
  }
  const { limit, pageCursor } = filter.value
  const namespaceID = props.namespace.namespaceID
  const moduleID = props.field.options.moduleID

  if (moduleID && moduleID !== NoID) {
    let qf = props.field.options.queryFields
    if (!qf || qf.length === 0) qf = [props.field.options.labelField]
    let searchQuery = q
    if (searchQuery.length > 0) {
      const fields = qf.map(f => module.value.fields.find(({ name }) => name === f))
      searchQuery = queryToFilter(searchQuery, '', fields)
    }
    fetchPrefiltered({ namespaceID, moduleID, query: searchQuery, sort: sortString(), limit, pageCursor })
  }
}, 600)

function loadLatest () {
  const namespaceID = props.namespace.namespaceID
  const moduleID = props.field.options.moduleID
  const { limit } = filter.value
  if (moduleID && moduleID !== NoID) {
    fetchPrefiltered({ namespaceID, moduleID, limit })
  }
}

function fetchPrefiltered (q = filter.value) {
  processing.value = true
  let { query: searchQuery = '' } = q
  if (props.field.options.prefilter) {
    const { skip, filter: pf } = evalPrefilterOrSkip(props.field.options.prefilter, {
      record: props.record,
      user: ($auth || {}).user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: (($auth || {}).user || {}).userID || NoID,
    })
    if (skip) {
      processing.value = false
      return
    }
    searchQuery = searchQuery ? `(${pf}) AND (${searchQuery})` : pf
  }

  if (q.pageCursor) q.sort = ''

  if (cancelRequest.value) {
    cancelRequest.value()
    cancelRequest.value = null
  }

  const api = $ComposeAPI || {}
  if (!api.recordListCancellable) {
    processing.value = false
    return
  }

  const { response, cancel } = api.recordListCancellable({ ...q, query: searchQuery })
  cancelRequest.value = cancel

  return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
    .then(([{ filter: f, set }]) => {
      filter.value = { ...filter.value, ...f }
      filter.value.nextPage = f.nextPage
      filter.value.prevPage = f.prevPage
      recordStore.updateRecords(set)
      records.value = set.map(r => markRaw(new compose.Record(module.value, r)))
      formatRecordValues(set.map(({ recordID }) => recordID))
      processing.value = false
    })
    .catch((e) => {
      if (axios.isCancel(e)) return
      processing.value = false
    })
}

function sortString () {
  return [props.field.options.labelField].filter(f => !!f).join(', ')
}

function formatRecordValues (recordIDs) {
  recordIDs = Array.isArray(recordIDs) ? recordIDs : [recordIDs].filter(v => v) || []
  const { namespaceID = NoID } = props.namespace
  const { moduleID = NoID, recordLabelField } = props.field.options

  if (!recordIDs.length || [moduleID, namespaceID].includes(NoID) || !labelField.value || !module.value) return

  if (labelField.value.kind === 'Record' && recordLabelField) {
    processing.value = true
    // Simplified - inline the fetchRecords logic
    setTimeout(() => { processing.value = false }, 300)
  } else if (labelField.value.kind === 'User') {
    processing.value = true
    setTimeout(() => { processing.value = false }, 300)
  }
}

function addRecordThroughRecordSelectField () {
  const { page } = getRecordSelectorPage()
  if (page === undefined) return
  const { pageID } = page

  const route = recordCreateLocation({ pageID })

  const { recordSelectorAddRecordDisplayOption } = props.extraOptions

  if (recordSelectorAddRecordDisplayOption === 'modal') {
    window.dispatchEvent(new CustomEvent('show-record-modal', { detail: { recordID: NoID, recordPageID: pageID, edit: true } }))
  } else if (recordSelectorAddRecordDisplayOption === 'newTab') {
    window.open(router.resolve(route).href)
  } else {
    router.push(route)
  }
}

function getRecordSelectorPage () {
  const recordFieldModuleID = props.field.options.moduleID
  if (!recordFieldModuleID) return {}
  const recordFieldPage = pageStore.set.find(p => p.moduleID === recordFieldModuleID)
  return { page: recordFieldPage }
}

function selectChange ({ recordID } = {}) {
  if (!recordID) return
  value.value = [...(value.value || []), recordID]
  if (singleSelect.value) {
    singleSelect.value._data._value = undefined
  }
  emit('change', value.value)
}

function goToPage (next = true) {
  filter.value.pageCursor = next ? filter.value.nextPage : filter.value.prevPage
}

function getOptionKey (val) {
  if (!val) return
  return typeof val === 'string' ? val : val.recordID
}

function getOptionLabel (val) {
  if (!val) return ''
  return typeof val === 'string' ? val : val.recordID
}

function setDefaultValues () {
  processing.value = false
  query.value = ''
  records.value = []
  filter.value = {}
}
</script>
