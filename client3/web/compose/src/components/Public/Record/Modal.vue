<template>
  <div
    v-if="showModal"
    class="modal d-block"
    tabindex="-1"
    role="dialog"
    @click.self="onHidden"
  >
    <div class="modal-dialog modal-xl" :class="dialogClass" role="document">
      <div class="modal-content position-initial">
        <div class="modal-header">
          <h5 class="modal-title">
            <div id="record-modal-header" />
          </h5>
        </div>
        <div class="modal-body p-0">
          <ViewRecord
            v-if="showModal"
            :namespace="namespace"
            :page="page"
            :module="module"
            :record-i-d="recordID"
            :values="values"
            :ref-record="refRecord"
            :edit="edit"
            in-modal
            @handle-record-redirect="loadRecord"
            @on-modal-back="loadRecord"
          />
        </div>
        <div class="modal-footer overflow-hidden">
          <div id="record-modal-footer" class="w-100 m-0" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NoID, compose } from 'corteza-lib/js/dist'
import { usePageStore } from 'corteza-webapp-compose/src/store/page'
import { useModuleStore } from 'corteza-webapp-compose/src/store/module'
import { useUiStore } from 'corteza-webapp-compose/src/store/ui'
import ViewRecord from 'corteza-webapp-compose/src/views/Public/Pages/Records/View.vue'

const { t: $t } = useI18n({ useScope: 'global' })
const route = useRoute()
const router = useRouter()
const pageStore = usePageStore()
const moduleStore = useModuleStore()
const uiStore = useUiStore()

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
})

const showModal = ref(false)
const recordID = ref(undefined)
const module = ref(undefined)
const page = ref(undefined)
const values = ref(undefined)
const refRecord = ref(undefined)
const edit = ref(false)

const uniqueID = computed(() => {
  const { recordPageID, recordID: rID, edit: e = false } = route.query
  const isEdit = typeof e === 'string' ? e === 'true' : Boolean(e)
  return [recordPageID, rID, isEdit]
})

const dialogClass = computed(() => {
  const classes = ['h-100', 'modal-max-width']
  if (uiStore.modalPageHandle) {
    classes.push(`page-${uiStore.modalPageHandle}-modal`)
  }
  if (uiStore.modalLayoutHandle) {
    classes.push(`layout-${uiStore.modalLayoutHandle}-modal`)
  }
  return classes.join(' ')
})

watch(uniqueID, (value = [], oldValue = []) => {
  const [recordPageID, rID, isEdit] = value
  const [oldRecordPageID] = oldValue

  if (!recordPageID) {
    setDefaultValues()
    uiStore.clearModalPreviousPage()
  }

  if (recordPageID !== oldRecordPageID) {
    if (uiStore.recordPaginationUsable) {
      uiStore.setRecordPaginationUsable(false)
    } else {
      uiStore.clearRecordPagination()
    }
  }

  if (!recordPageID) return

  if (rID !== recordID.value || recordPageID !== (page.value || {}).pageID || isEdit !== edit.value) {
    loadRecord({ recordID: rID, recordPageID: recordPageID, edit: isEdit })
  }
}, { immediate: true })

function loadRecord({ recordID: rID, recordPageID, values: vals = route.params.values, refRecord: refR, edit: e = edit.value, pushModalPreviousPage = true }) {
  if (!rID && !recordPageID) {
    onHidden()
    return
  }

  recordID.value = rID
  values.value = vals
  refRecord.value = refR
  edit.value = e || !rID || rID === NoID

  loadModal({ recordID: rID, recordPageID })

  if (pushModalPreviousPage) {
    uiStore.pushModalPreviousPage({ recordID: rID, recordPageID, edit: e })
  }

  router.push({
    params: route.params,
    query: {
      ...route.query,
      recordID: rID,
      recordPageID,
      edit: e,
    },
  })
}

function loadModal({ recordID: rID, recordPageID }) {
  if (rID && recordPageID) {
    recordID.value = rID

    if (!page.value || page.value.pageID !== recordPageID) {
      page.value = pageStore.getByID(recordPageID)
      if (page.value) {
        uiStore.setModalPageHandle(page.value.handle)
      }
    }

    if (page.value && (!module.value || module.value.moduleID !== page.value.moduleID)) {
      module.value = moduleStore.getByID(page.value.moduleID)
    }

    if (page.value && module.value) {
      showModal.value = true
    }
  }
}

function onHidden() {
  const { recordID: rID, recordPageID } = route.query
  setDefaultValues()

  if (rID !== undefined || recordPageID !== undefined) {
    router.push({
      params: route.params,
      query: {
        ...route.query,
        recordID: undefined,
        moduleID: undefined,
        recordPageID: undefined,
        edit: undefined,
        modalLayoutID: undefined,
      },
    })
  }

  uiStore.setModalPageHandle(undefined)
  uiStore.setModalLayoutHandle(undefined)
}

function setDefaultValues() {
  showModal.value = false
  edit.value = false
  recordID.value = undefined
  module.value = undefined
  page.value = undefined
  values.value = undefined
  refRecord.value = undefined
  uiStore.clearModalPreviousPage()
}

function destroyEvents() {
  window.removeEventListener('show-record-modal', loadRecord)
}

onMounted(() => {
  window.addEventListener('show-record-modal', (e) => loadRecord(e.detail))
})

onBeforeUnmount(() => {
  destroyEvents()
  setDefaultValues()
})
</script>

<style lang="scss">
.modal-max-width {
  max-width: 97vw;
}

.position-initial {
  position: initial;
}

.modal-header h5 {
  min-height: 27px;
}

.modal-body {
  background-color: var(--body-bg);
}
</style>
