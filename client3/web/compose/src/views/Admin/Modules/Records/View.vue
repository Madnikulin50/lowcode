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
import { ref, shallowRef, computed, watch, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue'
import { useRouter, useRoute, onBeforeRouteLeave, onBeforeRouteUpdate } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import { isEqual, throttle } from 'lodash'
import { composables } from 'corteza-lib/vue/dist'
import RecordToolbar from 'corteza-webapp-compose/src/components/Common/RecordToolbar'
import { compose, NoID, validator } from 'corteza-lib/js/dist'
import RecordBase from 'corteza-webapp-compose/src/components/PageBlocks/RecordBase'
import RecordEditor from 'corteza-webapp-compose/src/components/PageBlocks/RecordEditor'
import { recordCreateLocation } from 'corteza-webapp-compose/src/lib/record-create-nav'
import { useStore } from '../../../../store'

const { t } = useI18n()
const store = useStore()
const router = useRouter()
const route = useRoute()
const { proxy } = getCurrentInstance()
const { toastSuccess, toastWarning, toastErrorHandler } = composables.useToast()

const $auth = window.__auth
const $ComposeAPI = window.__composeAPI
const $EventBus = proxy.$EventBus

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

const inEditing = ref(false)
const record = ref(null)
const initialRecordState = ref(null)
const processing = ref(false)
const processingAction = ref('')
const loadingRecord = ref(false)
const errors = shallowRef(new validator.Validated())

const getNextAndPrevRecord = computed(() => store.getters['ui/getNextAndPrevRecord'])
const getUiEventResourceType = 'admin-record-page'

const isNew = computed(() => !props.recordID || props.recordID === NoID)

const isDeleted = computed(() => !!(record.value && record.value.deletedAt))

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

const validatorComp = computed(() => {
  if (!module.value) return undefined
  return new compose.RecordValidator(module.value)
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

const allRecords = computed(() => {
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

watch(processing, (val) => {
  if (!val) processingAction.value = ''
})

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

onBeforeRouteLeave(() => checkUnsavedChanges())

onBeforeRouteUpdate((to, from) => {
  if (JSON.stringify(to.params) === JSON.stringify(from.params)) return true
  return checkUnsavedChanges()
})

function createEvents () {
  window.addEventListener('refetch-records', refresh)
}

function destroyEvents () {
  window.removeEventListener('refetch-records', refresh)
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
    if (!r) return
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
  const found = store.getters['module/getByID'](moduleID)
  if (!found) return
  const mdl = Object.freeze(found.clone())

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
          toastErrorHandler(proxy.$t('notification.record.loadFailed'))(e)
          handleBack()
        }
      })
  }

  const { userID } = $auth.user
  return new compose.Record(mdl, { ownedBy: userID, values: props.values })
}

function handleBack () {
  router.push({ name: 'admin.modules.record.list', params: { moduleID: module.value.moduleID } })
}

function handleAdd () {
  router.push({ name: 'admin.modules.record.create', params: { moduleID: module.value.moduleID } })
}

function handleClone () {
  router.push(recordCreateLocation({
    name: 'admin.modules.record.create',
    moduleID: module.value.moduleID,
    values: record.value.values,
  }))
}

function handleEdit () {
  router.push({ name: 'admin.modules.record.edit', params: { moduleID: module.value.moduleID, recordID: record.value.recordID } })
}

function handleView () {
  router.push({ name: 'admin.modules.record.view', params: { moduleID: module.value.moduleID, recordID: record.value.recordID } })
}

function handleRedirectToPrevOrNext (recordID) {
  if (!recordID) return
  router.push({ params: { ...route.params, recordID } })
}

const handleFormSubmitSimple = throttle(function (routeName = 'admin.modules.record.view') {
  processingAction.value = 'submit'
  processing.value = true

  let saved
  const creating = record.value.recordID === NoID

  record.value = record.value.clone()

  return dispatchUiEvent('beforeFormSubmit')
    .then(() => validateRecordSimple())
    .then(() => {
      if (creating) {
        return $ComposeAPI.recordCreate(record.value)
      }
      return $ComposeAPI.recordUpdate(record.value)
    }).catch(err => {
      processing.value = false
      const { details = undefined } = err
      if (!!details && Array.isArray(details) && details.length > 0) {
        const e = errors.value
        e.push(...details.filter(d => !d.kind.includes('warning')))
        errors.value = e
        throw new Error(proxy.$t('notification.record.validationErrors', { fields: getValidationErrorFields() }))
      }
      throw err
    }).then(r => {
      saved = new compose.Record(module.value, r)
    }).then(() => dispatchUiEvent('afterFormSubmit', saved))
    .then(() => {
      if (saved.valueErrors?.set) {
        record.value = saved.clone()
        initialRecordState.value = saved.clone()
        setWarnings()
        toastWarning(proxy.$t('notification.record.validationWarnings', { fields: getValidationErrorFields({ includeWarnings: true }) }))
        processing.value = false
      } else {
        initialRecordState.value = record.value.clone()
        router.push({ name: routeName, params: { ...route.params, recordID: saved.recordID } })
        toastSuccess(proxy.$t(`notification.record.${creating ? 'create' : 'update'}Success`))
      }
    }).catch(e => {
      processing.value = false
      toastErrorHandler(proxy.$t(`notification.record.${creating ? 'create' : 'update'}Failed`))(e)
    })
}, 500)

const handleDelete = throttle(function () {
  processing.value = true
  processingAction.value = 'delete'

  return dispatchUiEvent('beforeDelete')
    .then(() => $ComposeAPI.recordDelete(record.value))
    .then(() => dispatchUiEvent('afterDelete'))
    .then(() => {
      window.dispatchEvent(new CustomEvent('refetch-records'))
      toastSuccess(proxy.$t('notification.record.deleteSuccess'))
    }).catch(e => {
      processing.value = false
      toastErrorHandler(proxy.$t('notification.record.deleteFailed'))(e)
    })
}, 500)

const handleUndelete = throttle(function () {
  processingAction.value = 'undelete'
  processing.value = true

  return dispatchUiEvent('beforeUndelete')
    .then(() => $ComposeAPI.recordUndelete(record.value))
    .then(() => dispatchUiEvent('afterUndelete'))
    .then(() => {
      window.dispatchEvent(new CustomEvent('refetch-records'))
      toastSuccess(proxy.$t('notification.record.restoreSuccess'))
    }).catch(e => {
      processing.value = false
      toastErrorHandler(proxy.$t('notification.record.restoreFailed'))(e)
    })
}, 500)

async function validateRecordSimple () {
  errors.value = validatorComp.value.run(record.value)
  if (errors.value.valid()) return

  await dispatchUiEvent('onFormSubmitError')

  errors.value = validatorComp.value.run(record.value)
  if (!errors.value.valid()) {
    throw new Error(proxy.$t('notification.record.validationErrors', { fields: getValidationErrorFields() }))
  }
}

function setWarnings () {
  const { set = [] } = record.value?.valueErrors || {}
  const e = errors.value
  set.forEach(s => {
    s.meta = s.meta || {}
    s.meta.isWarning = true
    e.push(s)
  })
  errors.value = e
}

function getValidationErrorFields ({ includeWarnings = false, includeErrors = true } = {}) {
  const { set = [] } = errors.value || {}
  const fields = new Set(set.filter(({ meta = {} } = {}) => {
    if (includeWarnings) return meta.isWarning
    if (includeErrors) return !meta.isWarning
    return true
  }).map(d => {
    const fieldName = d.meta.field
    const mod = (d.meta.moduleID && store.getters['module/getByID'](d.meta.moduleID)) || module.value
    if (mod) {
      const f = mod.fields.find(f => f.name === fieldName)
      if (f?.label) return f.label
    }
    return fieldName
  }))
  return Array.from(fields).join(', ')
}

function dispatchUiEvent (eventType, rec = record.value, args = {}) {
  if (!$EventBus) return Promise.resolve()
  const resourceType = `ui:compose:${getUiEventResourceType}`
  const argsBase = {
    errors: errors.value,
    validator: validatorComp.value,
    ...args,
  }
  return $EventBus.Dispatch(compose.RecordEvent(rec, { eventType, resourceType, args: argsBase }))
}

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

  let recordStateChange = true
  if (compareRecordValues()) {
    recordStateChange = window.confirm(proxy.$t('record.unsavedChanges'))
  }

  if (!recordStateChange) {
    processing.value = false
  } else if (record.value) {
    initialRecordState.value = record.value.clone()
  }

  return recordStateChange
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
