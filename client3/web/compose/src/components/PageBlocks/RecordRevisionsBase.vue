<template>
  <Wrap v-bind="$props" scrollable-body @refreshBlock="refresh">
    <div class="d-flex flex-column align-items-center h-100 overflow-hidden">
      <span v-if="revisionsDisabledOnModule" class="my-auto">{{ $t('errors.disabled-on-module') }}</span>
      <span v-else-if="isProcessing" class="spinner-border my-auto" />
      <button v-else-if="!preloadRevisions && !loadedRevisions" class="btn btn-outline-secondary my-auto" @click="refresh()">
        {{ $t('show-revisions', { revision: record.revision }) }}
      </button>
      <template v-else>
        <table class="table table-hover table-responsive flex-fill mh-100 mb-0 w-100 rounded">
          <thead>
            <tr>
              <th class="border-top-0 text-start">#</th>
              <th class="border-top-0">{{ $t('revisions.columns.operation.label') }}</th>
              <th class="border-top-0 text-end">{{ $t('revisions.columns.user.label') }}</th>
              <th class="border-top-0 text-end">{{ $t('revisions.columns.timestamp.label') }}</th>
              <th class="border-top-0 text-nowrap text-end"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(rev, idx) in revisions" :key="idx">
              <td>{{ rev.values.revision }}</td>
              <td>{{ $t(`operations.${rev.values.operation}`) }}</td>
              <td>
                <FieldViewer value-only :field="mockRevisionModule.findField('userID')" :record="rev" :module="mockRevisionModule" :namespace="namespace" />
              </td>
              <td>
                <FieldViewer value-only :field="mockRevisionModule.findField('timestamp')" :record="rev" :module="mockRevisionModule" :namespace="namespace" />
              </td>
              <td>
                <button v-if="rev.meta.changes.length > 0" class="btn btn-outline-extra-light d-flex align-items-center m-0 border-0 text-primary ms-auto" @click="toggleDetails(idx)">
                  {{ $t('show-changes', { count: rev.meta.changes.length }) }}
                  <font-awesome-icon :icon="expandedRows[idx] ? 'chevron-up' : 'chevron-down'" class="ms-2" />
                </button>
              </td>
            </tr>
            <tr v-for="(rev, idx) in revisions" :key="'details-'+idx" v-if="expandedRows[idx]">
              <td colspan="5" class="p-0">
                <table class="table mb-0">
                  <thead class="text-primary">
                    <tr>
                      <th>{{ $t('changes.columns.field.label') }}</th>
                      <th>{{ $t('changes.columns.old-value.label') }}</th>
                      <th>{{ $t('changes.columns.new-value.label') }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(change) in rev.meta.changes" :key="change.key">
                      <td>{{ change.label }}</td>
                      <td>
                        <FieldViewer v-if="rev.meta.oldRecord" value-only :field="module.findField(change.key)" :record="rev.meta.oldRecord" :module="module" :namespace="namespace" />
                        <span v-else>-</span>
                      </td>
                      <td>
                        <FieldViewer v-if="rev.meta.newRecord" value-only :field="module.findField(change.key)" :record="rev.meta.newRecord" :module="module" :namespace="namespace" />
                        <span v-else>-</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="!revisions.length" class="position-absolute text-center mt-5" style="left:0;right:0;bottom:calc(50% - 33px);">
          <p class="mb-0 mx-2">{{ $t('errors.no-revisions') }}</p>
        </div>
      </template>
    </div>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block', keyPrefix: 'recordRevisions.viewer' } })
import { ref, computed, watch, onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose, NoID } from 'corteza-lib/js/dist'
import { usePageBlockBase } from './usePageBlockBase'
import Wrap from './Wrap/index.js'
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

const { processing, isProcessing, options, refreshBlock } = usePageBlockBase(props, {})

const error = ref(null)
const loadedRevisions = ref(false)
const mockRevisionModule = ref(undefined)
const revisions = ref([])
const expandedRows = ref({})

const revisionsDisabledOnModule = computed(() => props.module ? !props.module.config?.recordRevisions?.enabled : false)
const preloadRevisions = computed(() => options.value.preload)

function toggleDetails(idx) {
  expandedRows.value[idx] = !expandedRows.value[idx]
}

watch(() => props.record?.recordID, () => { if (preloadRevisions.value) refresh() }, { immediate: true })
watch(options, () => { refresh() }, { deep: true })

onMounted(() => {
  refreshBlock(refresh)
  window.addEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.addEventListener('refetch-records', refresh)
})

onBeforeUnmount(() => {
  window.removeEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.removeEventListener('refetch-records', refresh)
})

function refreshOnRelatedRecordsUpdate({ detail: { moduleID } } = {}) {
  if (props.module?.moduleID === moduleID) refresh()
}

async function loadRevisions() {
  revisions.value = []
  if (revisionsDisabledOnModule.value || !props.record || props.record.recordID === NoID) return
  processing.value = true
  const fields = [
    { name: 'revision', kind: 'Number' }, { name: 'changeID', kind: 'String' },
    { name: 'userID', kind: 'User' }, { name: 'timestamp', kind: 'DateTime' }, { name: 'operation', kind: 'String' },
  ]
  mockRevisionModule.value = new compose.Module({ ...props.module, fields })
  return props.block.fetch(window.__composeAPI, props.record, options.value.sortDirection).then(set => {
    revisions.value = set.map(r => {
      let oldOwnedBy = NoID, newOwnedBy = NoID
      const oldValues = {}, newValues = {}
      r.changes.forEach(c => {
        if (c.old !== undefined) { if (c.key === 'ownedBy') oldOwnedBy = c.old[0]; else oldValues[c.key] = c.old }
        if (c.new !== undefined) { if (c.key === 'ownedBy') newOwnedBy = c.new[0]; else newValues[c.key] = c.new }
      })
      const oldRecord = Object.keys(oldValues).length > 0 ? new compose.Record(props.module, { ownedBy: oldOwnedBy, values: oldValues }) : null
      const newRecord = Object.keys(newValues).length > 0 ? new compose.Record(props.module, { ownedBy: newOwnedBy, values: newValues }) : null
      const changes = r.changes.map(c => {
        const field = props.module.findField(c.key)
        return { key: c.key, label: field ? field.label || field.name : c.key }
      })
      return new compose.Record(mockRevisionModule.value, {
        recordID: r.resource, values: { revision: r.revision, changeID: r.changeID, operation: r.operation, timestamp: r.timestamp, userID: r.userID },
        meta: { changes, oldRecord, newRecord },
      })
    })
  }).finally(() => { setTimeout(() => { processing.value = false }, 300) })
}

function refresh() { loadedRevisions.value = true; loadRevisions() }
</script>
