<template>
  <Wrap v-bind="$props" @refreshBlock="refresh">
    <template v-if="canAddRecord" #toolbar>
      <div class="p-3 border-bottom">
        <button class="btn btn-primary" @click.prevent="createNewRecord">
          {{ $t('recordOrganizer.addNewRecord') }}
        </button>
      </div>
    </template>

    <template #default>
      <label v-if="!isConfigured" class="text-primary p-3">{{ $t('recordOrganizer.notConfigured') }}</label>
      <div v-else class="h-100">
        <div v-if="isProcessing" class="d-flex align-items-center justify-content-center h-100">
          <div class="spinner-border" />
        </div>

        <draggable item-key="id" v-else :id="draggableID" v-model="records" handle=".record-item" :group="{ name: moduleID, pull: canPull, put: canPut }" :move="checkMove" class="h-100 pt-3 px-3 overflow-auto" @change="onChange">
          <template #header>
            <div v-if="!records.length" class="small text-secondary">{{ $t('recordOrganizer.noRecords') }}</div>
          </template>
          <template #item="{ element }">
            <div :key="`${element.recordID}`" class="card record-item border border-light mb-3 grab shadow-sm" @click="handleRecordClick(element)">
              <div class="card-body rounded p-3">
                <h6 v-if="labelField" class="d-flex overflow-hidden">
                  <FieldViewer v-if="labelField.canReadRecordValue" :field="labelField" :record="element" :namespace="namespace" value-only />
                  <i v-else class="text-secondary">{{ $t('field.noPermission') }}</i>
                </h6>
                <div v-if="descriptionField" class="d-flex overflow-hidden">
                  <FieldViewer v-if="descriptionField.canReadRecordValue" :field="descriptionField" :record="element" :namespace="namespace" value-only />
                  <i v-else class="text-secondary">{{ $t('field.noPermission') }}</i>
                </div>
              </div>
            </div>
          </template>
        </draggable>
      </div>
    </template>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch, onMounted, onBeforeUnmount, inject } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import { usePageBlockBase } from './usePageBlockBase'
import { compose, NoID } from 'corteza-lib/js/dist'
import { useStore } from '../../store'
import draggable from 'vuedraggable'
import FieldViewer from 'corteza-webapp-compose/src/components/ModuleFields/Viewer'
import { evalPrefilterOrSkip, getFieldFilter, isFieldInFilter } from 'corteza-webapp-compose/src/lib/record-filter'
import Wrap from './Wrap/index.js'

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

const emit = defineEmits(['errors', 'edit-block', 'clone-block', 'copy-block', 'delete-tab'])

const store = useStore()
const $auth = inject('$auth')
const $ComposeAPI = inject('$ComposeAPI')
const $router = useRouter()

const { processing, isProcessing, options, refreshBlock } = usePageBlockBase(props, emit)
const records = ref([])
const abortableRequests = ref([])

const getModuleByID = computed(() => store.module.getByID)
const pages = computed(() => store.page.set)

const draggableID = computed(() => `recordOrganizer-${props.blockIndex}`)
const roModule = computed(() => getModuleByID.value(moduleID.value))
const roRecordPage = computed(() => pages.value.find(p => p.moduleID === moduleID.value))
const moduleID = computed(() => options.value.moduleID)

const allFields = computed(() => {
  if (!options.value.moduleID) return []
  const mod = roModule.value
  if (!mod) return []
  return [
    ...mod.fields,
    ...mod.systemFields().map(sf => { sf.label = $t(`system.${sf.name}`); return sf }),
  ]
})

const labelField = computed(() => {
  const fn = options.value.labelField
  return fn ? allFields.value.find(f => f.name === fn) || {} : undefined
})

const descriptionField = computed(() => {
  const fn = options.value.descriptionField
  return fn ? allFields.value.find(f => f.name === fn) || {} : undefined
})

const positionField = computed(() => {
  const fn = options.value.positionField
  return fn ? allFields.value.find(f => f.name === fn) || {} : undefined
})

const groupField = computed(() => {
  const fn = options.value.groupField
  return fn ? allFields.value.find(f => f.name === fn) || {} : undefined
})

const canPull = computed(() => positionField.value ? positionField.value.canUpdateRecordValue : true)
const canPut = computed(() => canPull.value && (groupField.value ? groupField.value.canUpdateRecordValue : true))
const canAddRecord = computed(() => roModule.value?.canCreateRecord && roRecordPage.value)
const isConfigured = computed(() => !!(labelField.value || descriptionField.value))

const inModal = computed(() => props.mode === 'modal')

watch(() => options.value, () => { refresh() }, { deep: true })
watch(() => props.record?.recordID, () => { refresh() }, { immediate: true })

onMounted(() => {
  refreshBlock(refresh)
  createEvents()
})

onBeforeUnmount(() => {
  abortRequests()
  destroyEvents()
  setDefaultValues()
})

function createEvents() {
  window.addEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.addEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.addEventListener('refetch-records', refresh)
}

function refetchOnPrefilterValueChange({ detail: { fieldName } } = {}) {
  if (isFieldInFilter(fieldName, options.value.filter)) refresh()
}

function checkMove({ draggedContext = {}, relatedContext = {} }) {
  const { moduleID: mid, recordID } = draggedContext.element || {}
  const { $attrs = {}, $el = {}, $options = {} } = relatedContext.component || {}
  const relatedRecords = ($options.propsData || {}).value || []
  if (mid !== $attrs.group.name) return false
  return draggableID.value === $el.id || !relatedRecords.some(r => r.recordID === recordID)
}

function onChange({ added, moved }) {
  if (added) reorganize(added)
  else if (moved) reposition(moved)
}

function reorganize({ element: record, newIndex }) {
  moveRecord(record, calcNewPosition(record, newIndex), options.value.group)
}

function reposition({ element: record, newIndex }) {
  moveRecord(record, calcNewPosition(record, newIndex))
}

function calcNewPosition(record, newPosition = 0) {
  if (newPosition <= 0) return 0
  const total = records.value.length
  if (newPosition > total) return total
  return parseInt(records.value[newPosition - 1].values[options.value.positionField] || 0) + 1
}

function createNewRecord() {
  const { groupField: gf, group } = options.value
  if (!roRecordPage.value) return
  const { pageID } = roRecordPage.value
  const values = {}
  if (gf) values[gf] = group
  const route = {
    name: 'page.record.create',
    params: { pageID, values, refRecord: props.record },
    query: null,
    edit: true,
  }
  if (inModal.value || options.value.displayOption === 'modal') {
    window.dispatchEvent(new CustomEvent('show-record-modal', { detail: { recordID: NoID, recordPageID: pageID, values, refRecord: props.record, edit: true } }))
  } else if (options.value.displayOption === 'newTab') {
    window.open($router.resolve(route).href)
  } else {
    $router.push(route)
  }
}

function refreshOnRelatedRecordsUpdate({ detail: { moduleID: mid } = {} } = {}) {
  if (options.value.moduleID === mid) refresh()
}

function expandFilter() {
  const filter = []
  if (!props.record) {
    if ((options.value.filter || '').includes('${record')) throw new Error($t('notification.record.invalidRecordVar'))
    if ((options.value.filter || '').includes('${ownerID}')) throw new Error($t('notification.record.invalidOwnerVar'))
  }
  if (options.value.filter) {
    const { skip, filter: evaluated } = evalPrefilterOrSkip(options.value.filter, {
      record: props.record,
      user: $auth?.user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: ($auth?.user || {}).userID || NoID,
      loadingRecord: !!props.loadingRecord,
    })
    filter.push(`(${skip ? 'false' : evaluated})`)
  }
  if (groupField.value && options.value.group !== undefined) {
    filter.push(`(${getFieldFilter(groupField.value.name, groupField.value.kind, options.value.group, '=')})`)
  }
  return filter.join(' AND ')
}

async function moveRecord(record, position, group) {
  const { namespaceID, moduleID: mid, recordID } = record
  if (mid !== options.value.moduleID) throw new Error($t('record.moduleMismatch'))
  const { positionField: pf, groupField: gf } = options.value
  const args = { recordID, filter: expandFilter(), positionField: pf, position }
  if (group !== undefined) { args.groupField = gf; args.group = group || '' }
  const params = {
    procedure: 'organize',
    namespaceID,
    moduleID: mid,
    args: Object.keys(args).map(name => ({ name, value: String(args[name]) })),
  }
  return $ComposeAPI.recordExec(params).then(pullRecords)
}

async function pullRecords() {
  if (!roModule.value || roModule.value.moduleID !== options.value.moduleID) return
  const query = expandFilter()
  const pf = options.value.positionField || 'updatedAt'
  const { moduleID: mid, namespaceID: nid } = roModule.value
  processing.value = true
  const { response, cancel } = $ComposeAPI.recordListCancellable({ namespaceID: nid, moduleID: mid, query, sort: pf })
  abortableRequests.value.push(cancel)
  return response().then(({ set }) => {
    records.value = set.map(r => Object.freeze(new compose.Record(roModule.value, r)))
    store.record.updateRecords(records.value)
    return Promise.all([
      fetchUsers([labelField.value, descriptionField.value].filter(Boolean), records.value),
      fetchRecords(nid, [labelField.value, descriptionField.value].filter(Boolean), records.value),
    ])
  }).catch(e => { if (!axios.isCancel(e)) console.error(e) })
    .finally(() => { setTimeout(() => { processing.value = false }, 300) })
}

function handleRecordClick(record) {
  if (!roRecordPage.value) return
  const page = pages.value.find(p => p.moduleID === moduleID.value)
  if (!page) return
  const route = { name: 'page.record', params: { pageID: roRecordPage.value.pageID, recordID: record.recordID }, query: null }
  if (options.value.displayOption === 'modal' || inModal.value) {
    window.dispatchEvent(new CustomEvent('show-record-modal', { detail: { recordID: record.recordID, recordPageID: roRecordPage.value.pageID } }))
  } else if (options.value.displayOption === 'newTab') {
    window.open($router.resolve(route).href)
  } else {
    $router.push(route)
  }
}

function refresh() { pullRecords() }
function setDefaultValues() { processing.value = false; records.value = []; abortableRequests.value = [] }
function abortRequests() { abortableRequests.value.forEach(c => c()) }
function destroyEvents() {
  window.removeEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.removeEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.removeEventListener('refetch-records', refresh)
}

function fetchUsers(fields, recs) {
  const list = [...new Set(recs.map(r => fields.filter(c => c.kind === 'User').map(f => f.isSystem ? [r[f.name]] : f.isMulti ? r.values[f.name] : [r.values[f.name]])).flat(Infinity))].filter(uID => uID !== NoID)
  if (list.length) return store.user.resolveUsers(list)
}

function fetchRecords(nid, fields, recs) {
  const moduleRecords = {}
  fields.filter(c => c.kind === 'Record').forEach(f => {
    const { moduleID: fmid } = f.options || {}
    if (!fmid) return
    if (!moduleRecords[fmid]) moduleRecords[fmid] = new Set()
    recs.forEach(r => {
      const ids = f.isSystem ? [r[f.name]] : f.isMulti ? r.values[f.name] : [r.values[f.name]]
      ids.forEach(id => { if (id) moduleRecords[fmid].add(id) })
    })
  })
  return Promise.all(Object.entries(moduleRecords).map(([fmid, ids]) => {
    ids = [...ids]
    return ids.length ? store.record.resolveRecords({ namespaceID: nid, moduleID: fmid, recordIDs: ids }) : Promise.resolve([])
  }))
}
</script>
<style lang="scss" scoped>
.grab { cursor: grab !important; }
.record-item:hover { background-color: var(--light) !important; }
</style>
