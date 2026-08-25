<template>
  <div
    class="overflow-auto p-2"
  >
    <Teleport to="#topbar-title">
      {{ title }}
    </Teleport>

    <Teleport to="#topbar-tools">
      <div
        v-if="modulePage || allRecords"
        class="btn-group"
      >
        <router-link
          :disabled="!modulePage"
          :to="modulePage"
          class="btn btn-primary d-flex align-items-center"
        >
          {{ $t('edit.edit') }}
          <font-awesome-icon
            :icon="['far', 'edit']"
            size="sm"
            class="ms-2"
          />
        </router-link>

        <router-link
          v-if="allRecords"
          :disabled="!allRecords"
          :to="allRecords"
          class="btn btn-primary d-flex align-items-center"
          style="margin-left:2px;"
        >
          <font-awesome-icon
            :icon="['fas', 'columns']"
          />
        </router-link>
      </div>
    </Teleport>

    <div
      v-if="isDeleted"
      class="alert alert-warning m-2"
      role="alert"
    >
      {{ $t('block.record.recordDeleted') }}
    </div>

    <div
      v-if="module"
      class="row"
    >
      <div
        v-for="(b, index) in blocks"
        :key="index"
        class="col-12 col-md-6 col-lg-4"
        style="max-height: 650px; height: 650px;"
      >
        <component
          :is="getRecordComponent"
          :errors="errors"
          v-bind="{ namespace, page, module, block: b, record }"
          :loading-record="loadingRecord"
          class="p-2"
        />
      </div>
    </div>

    <Teleport to="#admin-toolbar">
      <record-toolbar
        :module="module"
        :record="record"
        :processing="processing"
        :processing-action="processingAction"
        :in-editing="inEditing"
        :is-created="!isNew"
        :record-navigation="recordNavigation"
        :hide-back="false"
        :hide-delete="false"
        :hide-new="false"
        :hide-clone="false"
        :hide-edit="false"
        :hide-submit="false"
        @add="handleAdd()"
        @clone="handleClone()"
        @edit="handleEdit()"
        @view="handleView()"
        @delete="handleDelete()"
        @undelete="handleUndelete()"
        @back="handleBack()"
        @submit="handleFormSubmitSimple('admin.modules.record.view')"
        @update-navigation="handleRedirectToPrevOrNext"
      />
    </Teleport>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'module' } })
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useStore } from '../../../../store'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import { isEqual } from 'lodash'
import RecordToolbar from 'corteza-webapp-compose/src/components/Common/RecordToolbar'
import { compose, NoID } from 'corteza-lib/js/dist'
import RecordBase from 'corteza-webapp-compose/src/components/PageBlocks/RecordBase'
import RecordEditor from 'corteza-webapp-compose/src/components/PageBlocks/RecordEditor'
import recordMixin from 'corteza-webapp-compose/src/mixins/record'

const { t } = useI18n()
const store = useStore()
const router = useRouter()
const route = useRoute()

const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
  moduleID: {
    type: String,
    required: false,
    default: '',
  },
  recordID: {
    type: String,
    required: false,
    default: '',
  },
  edit: {
    type: Boolean,
    default: false,
  },
  values: {
    type: Object,
    required: false,
    default: () => ({}),
  },
})

const blocks = ref([])
const page = ref(new compose.Page())
const recordNavigation = ref({ prev: undefined, next: undefined })
const abortableRequests = ref([])

// Record mixin state
const inEditing = ref(false)
const record = ref(null)
const initialRecordState = ref(null)
const processing = ref(false)
const processingAction = ref(false)
const loadingRecord = ref(false)
const isDeleted = ref(false)
const errors = ref({})

const getNextAndPrevRecord = computed(() => store.getters['ui/getNextAndPrevRecord'])
const getUiEventResourceType = 'admin-record-page'

const isNew = computed(() => !props.recordID || props.recordID === NoID)

const title = computed(() => {
  const { name, handle } = module.value || {}
  const prefix = isNew.value ? 'create' : inEditing.value ? 'edit' : 'view'
  return t(`allRecords.${prefix}.title`, { name: name || handle, interpolation: { escapeValue: false } })
})

const module = computed(() => {
  if (props.moduleID) {
    return store.getters['module/getByID'](props.moduleID)
  }
  return undefined
})

const fieldsList = computed(() => {
  if (!module.value) return []
  const result = []
  const fieldSetSize = 8
  for (let i = 0, j = module.value.fields.length; i < j; i += fieldSetSize) {
    result.push(module.value.fields.slice(i, i + fieldSetSize))
  }
  result.push(module.value.systemFields())
  return result
})

const getRecordComponent = computed(() => inEditing.value ? RecordEditor : RecordBase)

const modulePage = computed(() => {
  if (module.value) {
    return { name: 'admin.modules.edit', params: { moduleID: module.value.moduleID }, query: null }
  }
  return undefined
})

const allRecordsLink = computed(() => {
  if (module.value && module.value.moduleID) {
    return { name: 'admin.modules.record.list', params: { moduleID: module.value.moduleID } }
  }
  return undefined
})

const currentRecordNavigation = computed(() => {
  const { recordID } = record.value || {}
  return getNextAndPrevRecord.value(recordID)
})

const uniqueID = computed(() => [props.moduleID, props.recordID, props.edit])

watch(uniqueID, () => {
  refresh()
}, { immediate: true })

watch(() => props.edit, (val) => {
  inEditing.value = val
}, { immediate: true })

watch(currentRecordNavigation, (rn, oldRn) => {
  if (rn && (rn.prev || rn.next)) {
    recordNavigation.value = rn
  } else if (props.recordID !== NoID && oldRn && (oldRn.prev || oldRn.next)) {
    recordNavigation.value = oldRn
  } else {
    recordNavigation.value = { prev: undefined, next: undefined }
  }
})

onMounted(() => {
  createBlocks()
  createEvents()
})

onBeforeUnmount(() => {
  abortRequests()
  destroyEvents()
  setDefaultValues()
})

function createEvents () {
  // Assuming $root events - needs root access
}

function destroyEvents () {
}

function createBlocks () {
  fieldsList.value.forEach(f => {
    const options = {
      moduleID: props.moduleID,
      fields: f,
      inlineRecordEditEnabled: true,
    }
    blocks.value.push(new compose.PageBlockRecord({ options }))
  })
}

function refresh () {
  processing.value = true
  loadingRecord.value = true
  return loadRecord().then(r => {
    record.value = r
    initialRecordState.value = record.value.clone()
  }).finally(() => {
    loadingRecord.value = false
    processing.value = false
  })
}

async function loadRecord () {
  const { moduleID = NoID, recordID = NoID } = props
  if (!moduleID || moduleID === NoID) return
  const mdl = Object.freeze(store.getters['module/getByID'](moduleID).clone())

  if (recordID && recordID !== NoID) {
    const { namespaceID } = props.namespace
    const { response, cancel } = $ComposeAPI.recordReadCancellable({ namespaceID, moduleID, recordID })
    abortableRequests.value.push(cancel)
    return response()
      .then(r => {
        r = new compose.Record(mdl, r)
        store.dispatch('record/updateRecords', r)
        return new Promise(resolve => setTimeout(resolve, 300)).then(() => r)
      })
      .catch(e => {
        if (!axios.isCancel(e)) {
          toastErrorHandler(t('notification.record.loadFailed'))(e)
          handleBack()
        }
      })
  } else {
    const { userID } = $auth.user
    return new compose.Record(mdl, { ownedBy: userID, values: props.values })
  }
}

function handleBack () {
  router.push({ name: 'admin.modules.record.list', params: { moduleID: module.value.moduleID } })
}

function handleAdd () {
  router.push({ name: 'admin.modules.record.create', params: { moduleID: module.value.moduleID, edit: true } })
}

function handleClone () {
  router.push({ name: 'admin.modules.record.create', params: { moduleID: module.value.moduleID, values: record.value.values, edit: true } })
}

function handleEdit () {
  router.push({ name: 'admin.modules.record.edit', params: { moduleID: module.value.moduleID, edit: true } })
}

function handleView () {
  router.push({ name: 'admin.modules.record.view', params: { moduleID: module.value.moduleID, edit: false } })
}

function handleRedirectToPrevOrNext (recordID) {
  if (!recordID) return
  router.push({ params: { ...route.params, recordID } })
}

function handleFormSubmitSimple (name) {}
function handleUndelete () {}
function handleDelete () {}
function toastErrorHandler (msg) { return (e) => {} }

function abortRequests () {
  abortableRequests.value.forEach((cancel) => cancel())
}

function compareRecordValues () {
  const rv = JSON.parse(JSON.stringify(record.value ? record.value.values : {}))
  const iv = JSON.parse(JSON.stringify(initialRecordState.value ? initialRecordState.value.values : {}))
  return !isEqual(rv, iv)
}

function checkUnsavedChanges () {
  if (!inEditing.value) return true
  const changed = compareRecordValues()
  if (!changed) {
    const result = window.confirm(t('record.unsavedChanges'))
    if (!result) processing.value = false
    else initialRecordState.value = record.value.clone()
    return result
  }
  return true
}

function setDefaultValues () {
  blocks.value = []
  page.value = undefined
  abortableRequests.value = []
  recordNavigation.value = { prev: undefined, next: undefined }
}
</script>

<style lang="scss">
.value {
  min-height: 1.2rem;
}
</style>
