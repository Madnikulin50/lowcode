<template>
  <div class="d-flex flex-column flex-grow-1 w-100 h-100">
    <Teleport :to="portalTopbarTitle">
      {{ title }}
    </Teleport>

    <Teleport to="#topbar-tools">
      <c-input-search
        v-if="enableAI"
        v-model.trim="aiPrompt"
        class="me-2"
        :ai="true"
        :aria-label="$t('AI')"
        :placeholder="$t('aiChat.startPrompt')"
        :autocomplete="'off'"
        submittable
        @search="handleAiSearch"
      />

      <router-link
        v-if="page && isRecordPage && page.canUpdatePage"
        :to="moduleEditor"
        class="btn btn-primary btn-sm d-flex align-items-center me-2"
        :class="{ disabled: !moduleEditor }"
      >
        {{ $t('navigation.editModule') }}
        <font-awesome-icon
          :icon="['far', 'edit']"
          class="ms-2"
        />
      </router-link>

      <div
        v-if="page && page.canUpdatePage"
        class="btn-group btn-group-sm text-nowrap"
      >
        <router-link
          data-test-id="button-page-builder"
          :to="pageBuilder"
          class="btn btn-primary d-flex align-items-center"
        >
          {{ $t('label.pageBuilder') }}
          <font-awesome-icon
            :icon="['fas', 'tools']"
            class="ms-2"
          />
        </router-link>

        <router-link
          data-test-id="button-page-edit"
          :to="pageEditor"
          class="btn btn-primary d-flex align-items-center"
          :title="$t('tooltip.edit.page')"
          style="margin-left:2px;"
          data-bs-toggle="tooltip"
          data-bs-boundary="body"
        >
          <font-awesome-icon
            :icon="['far', 'edit']"
          />
        </router-link>

        <page-translator
          v-if="trPage"
          data-test-id="button-page-translations"
          :page="trPage"
          :page-layout="layout"
          button-variant="primary"
          style="margin-left:2px;"
        />
      </div>
    </Teleport>

    <div
      v-if="isDeleted"
      class="alert alert-warning mx-1 my-2"
      role="alert"
    >
      {{ $t('block.record.recordDeleted') }}
    </div>
    <ai-chat-modal
      :namespace="page.namespaceID"
      :page="page.pageID"
      :module="page.moduleID"
    />
    <div
      v-if="isDraft"
      class="alert alert-secondary d-flex align-items-center mx-1 my-2 border"
      role="alert"
    >
      {{ $t('drafts.activeDraft') }}
    </div>

    <div
      v-if="isLoading"
      class="d-flex align-items-center justify-content-center w-100 h-100"
    >
      <span class="spinner-border" />
    </div>

    <grid
      v-else
      v-bind="$props"
      :errors="errors"
      :record="record"
      :loading-record="loadingRecord"
      :blocks="blocks"
      :mode="inEditing ? 'editor' : 'base'"
      class="h-100"
    />

    <Teleport :to="portalRecordToolbar">
      <record-toolbar
        :module="module"
        :record="record"
        :labels="recordToolbarLabels"
        :processing="processing"
        :processing-action="processingAction"
        :in-editing="inEditing"
        :in-modal="inModal"
        :is-created="!isNew"
        :record-navigation="recordNavigation"
        :is-draft="isDraft"
        :is-new="isNew"
        :hide-back="!layoutButtons.has('back')"
        :hide-delete="!layoutButtons.has('delete')"
        :hide-new="!layoutButtons.has('new')"
        :hide-clone="!layoutButtons.has('clone')"
        :hide-edit="!layoutButtons.has('edit')"
        :hide-submit="!layoutButtons.has('submit')"
        :has-back="viewHasBack"
        @add="handleAdd()"
        @clone="handleClone()"
        @edit="handleEdit()"
        @view="handleView()"
        @delete="handleDelete()"
        @undelete="handleUndelete()"
        @back="handleBack()"
        @submit="handleFormSubmit('page.record')"
        @update-navigation="handleRedirectToPrevOrNext"
      >
        <template #start-actions>
          <template v-for="(action, index) in layoutActions.filter(a => a.placement === 'start')" :key="index">
            <router-link
              v-if="generateActionLink(action)"
              :to="generateActionLink(action)"
              class="btn text-nowrap"
              :class="'btn-' + (action.meta.style.variant || 'secondary')"
            >
              {{ action.meta.label }}
            </router-link>
            <a
              v-else-if="generateActionHref(action)"
              :href="generateActionHref(action)"
              :target="generateActionTarget(action)"
              class="btn text-nowrap"
              :class="'btn-' + (action.meta.style.variant || 'secondary')"
            >
              {{ action.meta.label }}
            </a>
            <button
              v-else
              :disabled="processing"
              class="btn text-nowrap"
              :class="'btn-' + (action.meta.style.variant || 'secondary')"
            >
              {{ action.meta.label }}
            </button>
          </template>
        </template>

        <template #center-actions>
          <template v-for="(action, index) in layoutActions.filter(a => a.placement === 'center')" :key="index">
            <router-link
              v-if="generateActionLink(action)"
              :to="generateActionLink(action)"
              class="btn text-nowrap"
              :class="'btn-' + (action.meta.style.variant || 'secondary')"
            >
              {{ action.meta.label }}
            </router-link>
            <a
              v-else-if="generateActionHref(action)"
              :href="generateActionHref(action)"
              :target="generateActionTarget(action)"
              class="btn text-nowrap"
              :class="'btn-' + (action.meta.style.variant || 'secondary')"
            >
              {{ action.meta.label }}
            </a>
            <button
              v-else
              :disabled="processing"
              class="btn text-nowrap"
              :class="'btn-' + (action.meta.style.variant || 'secondary')"
            >
              {{ action.meta.label }}
            </button>
          </template>
        </template>

        <template #end-actions>
          <template v-for="(action, index) in layoutActions.filter(a => a.placement === 'end')" :key="index">
            <router-link
              v-if="generateActionLink(action)"
              :to="generateActionLink(action)"
              class="btn text-nowrap"
              :class="'btn-' + (action.meta.style.variant || 'secondary')"
            >
              {{ action.meta.label }}
            </router-link>
            <a
              v-else-if="generateActionHref(action)"
              :href="generateActionHref(action)"
              :target="generateActionTarget(action)"
              class="btn text-nowrap"
              :class="'btn-' + (action.meta.style.variant || 'secondary')"
            >
              {{ action.meta.label }}
            </a>
            <button
              v-else
              :disabled="processing"
              class="btn text-nowrap"
              :class="'btn-' + (action.meta.style.variant || 'secondary')"
            >
              {{ action.meta.label }}
            </button>
          </template>
        </template>
      </record-toolbar>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue'
import { useRouter, useRoute, onBeforeRouteLeave, onBeforeRouteUpdate } from 'vue-router'
import axios from 'axios'
import { isEqual, throttle } from 'lodash'
import { compose, system, NoID, validator } from 'corteza-lib/js/dist'
import { components, composables, useDraftsStore } from 'corteza-lib/vue/dist'
import { useUiStore } from '../../../../store/ui'
import { useModuleStore } from '../../../../store/module'
import { usePageStore } from '../../../../store/page'
import { usePageLayoutStore } from '../../../../store/page-layout'
import { useRecordStore } from '../../../../store/record'
import { evaluatePrefilter } from 'corteza-webapp-compose/src/lib/record-filter'
import { fetchID } from 'corteza-webapp-compose/src/lib/block'
import bus from '../../../../lib/bus'
import Grid from 'corteza-webapp-compose/src/components/Public/Page/Grid'
import RecordToolbar from 'corteza-webapp-compose/src/components/Common/RecordToolbar'
import PageTranslator from 'corteza-webapp-compose/src/components/Admin/Page/PageTranslator'
import AiChatModal from 'corteza-webapp-compose/src/components/Public/Page/AiChat/Modal.vue'

const { CInputSearch } = components
const { toastSuccess, toastWarning, toastErrorHandler } = composables.useToast()
const { proxy } = getCurrentInstance()

const $auth = window.__auth
const $ComposeAPI = window.__composeAPI
const $SystemAPI = window.__systemAPI
const $Settings = window.__settings
const draftsStore = useDraftsStore()
const $EventBus = proxy.$EventBus

const props = defineProps({
  page: { type: compose.Page, required: true },
  namespace: { type: compose.Namespace, required: true },
  module: { type: compose.Module, required: false, default: () => ({}) },
  refRecord: { type: compose.Record, required: false, default: undefined },
  values: { type: Object, required: false, default: () => ({}) },
  inModal: { type: Boolean, default: false },
  edit: { type: Boolean, default: false },
})

const emit = defineEmits(['on-modal-back', 'handle-record-redirect'])

const router = useRouter()
const route = useRoute()

const uiStore = useUiStore()
const moduleStore = useModuleStore()
const pageStore = usePageStore()
const pageLayoutStore = usePageLayoutStore()
const recordStore = useRecordStore()

const layouts = ref([])
const layout = ref(undefined)
const blocks = ref(undefined)
const tempRecord = ref(undefined)

const inEditing = ref(props.edit)
const processing = ref(false)
const processingAction = ref('')
const record = ref(undefined)
const initialRecordState = ref(undefined)
const errors = ref(new validator.Validated())

const aiPrompt = ref('')
const loading = ref(false)
const layoutButtons = ref(new Set())
const recordNavigation = ref({ prev: undefined, next: undefined })
const abortableRequests = ref([])
const loadingRecord = ref(false)
const activeDraftKey = ref(null)

const isRecordPage = computed(() => record.value?.recordID || route.name === 'page.record.create')

const validatorComp = computed(() => {
  if (!props.module?.moduleID) {
    throw new Error('can not initialize record validator without module')
  }
  return new compose.RecordValidator(props.module)
})

const isValid = computed(() => errors.value.valid())

const isDeleted = computed(() => record.value && record.value.deletedAt)

const isNew = computed(() => !recordID.value || recordID.value === NoID)

const recordID = computed(() => {
  if (record.value?.recordID && record.value.recordID !== NoID) return record.value.recordID
  return route.params.recordID || ''
})

const isLoading = computed(() => loading.value || !layout.value || !blocks.value)

const portalTopbarTitle = computed(() => props.inModal ? '#record-modal-header' : '#topbar-title')

const portalRecordToolbar = computed(() => props.inModal ? '#record-modal-footer' : '#toolbar')

const getUiEventResourceType = computed(() => 'record-page')

const recordToolbarLabels = computed(() => {
  const aux = {}
  const { config = {} } = layout.value || {}
  const { buttons = {} } = config
  Object.entries(buttons).forEach(([key, { label = '' }]) => {
    aux[key] = label
  })
  if (isDraft.value) {
    aux.delete = proxy.$t('drafts.deleteDraft')
    aux.clone = proxy.$t('drafts.saveAsNewDraft')
    aux.new = proxy.$t('drafts.saveAsNewDraft')
    aux.submit = isNew.value ? proxy.$t('label.createRecordFromDraft') : proxy.$t('drafts.applyDraft')
    aux.edit = proxy.$t('drafts.viewRecord')
  }
  return aux
})

const layoutActions = computed(() => {
  const { config = {} } = layout.value || {}
  const { actions = [] } = config
  return actions.filter(({ enabled }) => enabled)
})

const title = computed(() => {
  if (!layout.value) return ''
  const { config = {}, meta = {} } = layout.value || {}
  const { useTitle = false } = config
  if (useTitle) {
    try {
      return evaluatePrefilter(meta.title, {
        record: record.value,
        user: $auth.user || {},
        recordID: (record.value || {}).recordID || NoID,
        ownerID: (record.value || {}).ownedBy || NoID,
        userID: ($auth.user || {}).userID || NoID,
      })
    } catch (e) {
      return e
    }
  }
  const { name, handle } = props.module || {}
  const titlePrefix = isNew.value ? 'create' : inEditing.value ? 'edit' : 'view'
  return proxy.$t(`page.public.record.${titlePrefix}.title`, { name: name || handle, interpolation: { escapeValue: false } })
})

const currentRecordNavigation = computed(() => {
  const { recordID: rid } = record.value || {}
  return uiStore.getNextAndPrevRecord(rid)
})

const viewHasBack = computed(() => {
  if (props.inModal) return uiStore.modalPreviousPages.length > 1
  return uiStore.previousPages.length > 0
})

const uniqueID = computed(() => [
  (props.page || {}).pageID,
  route.query.layoutID,
  route.query.modalLayoutID,
  recordID.value,
  props.edit,
  route.query.draftID,
])

const showDrafts = computed(() => $Settings.get('ui.topbar.showDrafts', false))

const isDraft = computed(() => !!route.query.draftID)

const enableAI = computed(() => true)

const pageEditor = computed(() => ({
  name: 'admin.pages.edit',
  params: { pageID: props.page.pageID },
}))

const pageBuilder = computed(() => {
  const { pageLayoutID: plID } = layout.value || {}
  return { name: 'admin.pages.builder', params: { pageID: props.page.pageID }, query: { layoutID: plID } }
})

const moduleEditor = computed(() => {
  if (!props.module?.moduleID) return undefined
  return { name: 'admin.modules.edit', params: { moduleID: props.module.moduleID } }
})

const trPage = computed({
  get: () => props.page.clone(),
  set: (v) => pageStore.updateSet([v]),
})

const getPageLayouts = (pageID) => pageLayoutStore.getByPageID(pageID)

watch(uniqueID, (value, oldValue) => {
  const [pageID = '', pageLayoutID = '', modalPageLayoutID = '', recID = '', edit = '', draftID = ''] = value || []
  const [oldPageID = '', oldPageLayoutID = '', oldModalPageLayoutID = '', oldRecordID = '', oldEdit = '', oldDraftID = ''] = oldValue || []

  if (!pageID || pageID === NoID) return

  if (pageID !== oldPageID) {
    loading.value = true
    layouts.value = getPageLayouts(props.page.pageID)
  }

  if ((recID === NoID && recID !== oldRecordID) || recID !== oldRecordID || edit !== oldEdit || pageID !== oldPageID || draftID !== oldDraftID) {
    refresh()
    return
  }

  const currentLayoutID = props.inModal ? modalPageLayoutID : pageLayoutID
  const oldLayoutID = props.inModal ? oldModalPageLayoutID : oldPageLayoutID

  if (currentLayoutID !== oldLayoutID) {
    determineLayout({ pageLayoutID: currentLayoutID })
      .then(b => { if (b) blocks.value = b })
      .finally(() => { processing.value = false })
  }
}, { immediate: true })

watch(() => layout.value?.handle, (handle, oldHandle) => {
  if (handle !== oldHandle) {
    props.inModal ? uiStore.setModalLayoutHandle(handle) : uiStore.setLayoutHandle(handle)
  }
}, { immediate: true })

watch(currentRecordNavigation, (rn, oldRn) => {
  if (rn?.prev || rn?.next) {
    recordNavigation.value = rn
  } else if (recordID.value !== NoID && (oldRn?.prev || oldRn?.next)) {
    recordNavigation.value = oldRn
  } else {
    recordNavigation.value = { prev: undefined, next: undefined }
  }
})

watch(title, (t) => {
  if (t && !props.inModal) {
    document.title = t
  }
}, { immediate: true })

onMounted(() => {
  createEvents()
})

onBeforeUnmount(() => {
  abortRequests()
  destroyEvents()
  cleanupDraftSync()
  setDefaultValues()
})

onBeforeRouteLeave(() => {
  return checkUnsavedChanges()
})

onBeforeRouteUpdate((to, from) => {
  const areParamsChanged = JSON.stringify(to.params) !== JSON.stringify(from.params)
  if (!areParamsChanged) return true
  return checkUnsavedChanges()
})

function createEvents() {
  bus.$on('refetch-records', refetchRecords)
  bus.$on('record-field-change', evaluateLayoutConditions)
  bus.$on('record-field-change', saveDraftRevision)

  if (props.inModal) {
    bus.$on('bv::modal::hide', checkUnsavedChanges)
  }
}

function evaluateLayoutConditions() {
  evaluateBlocks()
}

async function loadRecord() {
  if (!props.page) return

  const { namespaceID, moduleID } = props.page

  if (moduleID !== NoID) {
    const mod = Object.freeze(getModuleByID(moduleID).clone())
    const rid = recordID.value

    if (rid && rid !== NoID) {
      const { response, cancel } = $ComposeAPI.recordReadCancellable({ namespaceID, moduleID, recordID: rid })
      abortableRequests.value.push(cancel)

      return response()
        .then(r => {
          r = new compose.Record(mod, r)
          recordStore.updateRecords([r])
          return new Promise(resolve => setTimeout(resolve, 300)).then(() => r)
        })
        .catch(e => {
          if (!axios.isCancel(e)) {
            toastErrorHandler(proxy.$t('notification.record.loadFailed'))(e)
            handleBack()
          }
        })
    } else {
      if (props.refRecord?.recordID && props.refRecord.recordID !== NoID) {
        recordStore.updateRecords([props.refRecord])

        mod.fields.filter(f => f.kind === 'Record' && f.options.moduleID === props.refRecord.moduleID).forEach(f => {
          if (f.isMulti) {
            props.values[f.name] = [props.refRecord.recordID]
          } else {
            props.values[f.name] = props.refRecord.recordID
          }
        })
      }

      const { userID } = $auth.user
      await new Promise(resolve => setTimeout(resolve, 300))
      return new compose.Record(mod, { ownedBy: userID, values: props.values })
    }
  }
}

async function handleBack() {
  if (props.inModal) {
    if (checkUnsavedChanges()) {
      uiStore.popModalPreviousPage().then(({ recordID, recordPageID, edit }) => {
        emit('on-modal-back', { recordID, recordPageID, pushModalPreviousPage: false, edit })
      })
    }
  } else {
    const previousPage = await uiStore.popPreviousPages()
    const extraPop = !isNew.value
    router.push(previousPage || { name: 'pages', params: { slug: props.namespace.slug || props.namespace.namespaceID } })
    if (extraPop) {
      uiStore.popPreviousPages()
    }
  }
}

function handleAdd() {
  processing.value = true

  if (isDraft.value) {
    const currentDraftID = activeDraftKey.value
    activeDraftKey.value = generateChangeID()
    saveDraft()
    activeDraftKey.value = currentDraftID
    toastSuccess(proxy.$t('drafts.saved'))
    processing.value = false
    return
  }

  if (props.inModal) {
    if (checkUnsavedChanges()) {
      emit('handle-record-redirect', { recordID: NoID, recordPageID: props.page.pageID, edit: true })
    }
  } else {
    router.push({ name: 'page.record.create', params: { pageID: props.page.pageID, edit: true } })
  }
}

function handleClone() {
  processing.value = true

  if (isDraft.value) {
    const currentDraftID = activeDraftKey.value
    activeDraftKey.value = generateChangeID()
    saveDraft()
    activeDraftKey.value = currentDraftID
    toastSuccess(proxy.$t('drafts.saved'))
    processing.value = false
    return
  }

  if (props.inModal) {
    if (checkUnsavedChanges()) {
      emit('handle-record-redirect', { recordID: NoID, recordPageID: props.page.pageID, values: record.value?.values, edit: true })
    }
  } else {
    router.push({ name: 'page.record.create', params: { pageID: props.page.pageID, values: record.value?.values, edit: true } })
  }
}

function handleEdit() {
  processing.value = true

  if (props.inModal) {
    emit('handle-record-redirect', { recordID: recordID.value, recordPageID: props.page.pageID, edit: true })
  } else {
    router.push({ name: 'page.record.edit', params: { recordID: recordID.value, pageID: props.page.pageID, edit: true } })
  }
}

function handleView() {
  processing.value = true

  if (isDraft.value) {
    openDraftRecord()
    processing.value = false
    return
  }

  if (props.inModal) {
    if (checkUnsavedChanges()) {
      emit('handle-record-redirect', { recordID: recordID.value, recordPageID: props.page.pageID, edit: false })
    }
  } else {
    router.push({ name: 'page.record', params: { recordID: recordID.value, pageID: props.page.pageID, edit: false } })
  }
  processing.value = false
}

const handleDelete = throttle(function () {
  processing.value = true
  processingAction.value = 'delete'

  if (isDraft.value) {
    draftsStore.removeDraft({ changeID: activeDraftKey.value })
    activeDraftKey.value = null
    toastSuccess(proxy.$t('drafts.deleted'))
    openDraftRecord()
    processing.value = false
    return
  }

  return dispatchUiEvent('beforeDelete')
    .then(() => $ComposeAPI.recordDelete(record.value))
    .then(() => dispatchUiEvent('afterDelete'))
    .then(() => {
      if (activeDraftKey.value) {
        draftsStore.removeDraft({ changeID: activeDraftKey.value })
        activeDraftKey.value = null
      }
      bus.$emit('refetch-records')
      toastSuccess(proxy.$t('notification.record.deleteSuccess'))
    }).catch(e => {
      processing.value = false
      toastErrorHandler(proxy.$t('notification.record.deleteFailed'))(e)
    })
}, 500)

function handleRedirectToPrevOrNext(recID) {
  if (!recID) return
  processing.value = true

  if (props.inModal) {
    if (checkUnsavedChanges()) {
      emit('handle-record-redirect', { recordID: recID, recordPageID: props.page.pageID })
    }
  } else {
    router.push({ params: { ...route.params, recordID: recID } })
    uiStore.popPreviousPages()
  }
}

function openDraftRecord() {
  const { draftID, ...query } = route.query
  activeDraftKey.value = null

  if (props.inModal) {
    emit('handle-record-redirect', { recordID: recordID.value, recordPageID: props.page.pageID, edit: false })
  } else if (!isNew.value) {
    router.push({ name: 'page.record', params: { ...route.params, recordID: recordID.value }, query })
  } else {
    router.push({ name: 'page.record.create', params: { pageID: props.page.pageID, edit: true } })
  }
}

function handleRecordButtons() {
  const { config = {} } = layout.value || {}
  const { buttons = [] } = config
  layoutButtons.value = new Set(
    Object.entries(buttons).reduce((acc, [key, value]) => {
      if (value.enabled) acc.push(key)
      return acc
    }, [])
  )
}

function refetchRecords({ recordID: rid } = {}) {
  if (rid && rid === recordID.value && props.inModal) return

  if (isNew.value || (inEditing.value && compareRecordValues() && !window.confirm(proxy.$t('notification.record.staleDataRefresh')))) {
    return
  }

  refresh()
}

async function refresh() {
  processing.value = true
  loadingRecord.value = true

  if (isRecordPage.value) {
    inEditing.value = props.edit
  }

  return loadRecord().then(r => {
    tempRecord.value = r
    const pageLayoutID = route.query.layoutID
    return determineLayout({ pageLayoutID }).then(b => {
      if (b) blocks.value = b
      record.value = tempRecord.value
      initialRecordState.value = record.value.clone()
      getRecordDraft()
    })
  }).finally(() => {
    tempRecord.value = undefined
    processing.value = false
    loading.value = false
    loadingRecord.value = false
  })
}

function generateActionLink(action) {
  const { kind, params = {} } = action
  if (kind === 'toLayout') {
    const pageLayoutID = params.pageLayoutID
    if (pageLayoutID) {
      if (props.inModal) {
        return { ...route, query: { ...route.query, modalLayoutID: pageLayoutID } }
      } else {
        return { ...route, query: { ...route.query, layoutID: pageLayoutID } }
      }
    }
  }
  return undefined
}

function generateActionHref(action) {
  const { kind, params = {} } = action
  return kind === 'toURL' ? params.url : undefined
}

function generateActionTarget(action) {
  const { kind, params = {} } = action
  return kind === 'toURL' ? (params.openIn === 'newTab' ? '_blank' : '_self') : undefined
}

function setDefaultValues() {
  inEditing.value = false
  layoutButtons.value = new Set()
  blocks.value = undefined
  recordNavigation.value = { prev: undefined, next: undefined }
  abortableRequests.value = []
  loadingRecord.value = false
}

function abortRequests() {
  abortableRequests.value.forEach(cancel => cancel())
}

function destroyEvents() {
  bus.$off('refetch-records', refetchRecords)
  bus.$off('record-field-change', evaluateLayoutConditions)
  bus.$off('record-field-change', saveDraftRevision)
  if (props.inModal) {
    bus.$off('bv::modal::hide', checkUnsavedChanges)
  }
}

function compareRecordValues() {
  const rv = JSON.parse(JSON.stringify(record.value ? record.value.values : {}))
  const iv = JSON.parse(JSON.stringify(initialRecordState.value ? initialRecordState.value.values : {}))
  return !isEqual(rv, iv)
}

function checkUnsavedChanges(bvEvent, modalId) {
  if ((bvEvent && modalId !== 'record-modal') || !inEditing.value || isDraft.value) return true

  let recordStateChange = true

  if (compareRecordValues()) {
    const message = showDrafts.value ? proxy.$t('record.unsavedChangesDraft') : proxy.$t('record.unsavedChanges')
    recordStateChange = window.confirm(message)
  }

  if (!recordStateChange) {
    processing.value = false
    if (bvEvent) bvEvent.preventDefault()
  } else if (record.value) {
    initialRecordState.value = record.value.clone()
  }

  return recordStateChange
}

function getRecordDraft() {
  activeDraftKey.value = null
  if (!record.value || !inEditing.value) return

  const { draftID } = route.query
  if (!draftID) return

  const draft = draftsStore.getDraft(draftID)
  if (!draft) return

  const { revision } = draft
  if (revision.record) {
    const revRecord = new compose.Record(props.module, revision.record)
    record.value.values = revRecord.values
  } else {
    revision.changes.forEach(({ key, new: values }) => {
      if (props.module.fields.find(f => f.name === key)) {
        record.value.values[key] = props.module.makeValue(key, values)
      }
    })
  }
  activeDraftKey.value = revision.changeID
}

function cleanupDraftSync() {
  activeDraftKey.value = null
}

const saveDraftRevision = throttle(function () {
  saveDraft()
}, 2000)

function saveDraft() {
  if (!record.value || !inEditing.value || !showDrafts.value) return

  const changes = computeChanges()
  if (changes.length === 0) return

  if (!activeDraftKey.value) {
    activeDraftKey.value = generateChangeID()
  }

  const resourceID = isNew.value ? '0' : String(record.value.recordID)
  const resource = `compose:record/${props.namespace.namespaceID}/${props.module.moduleID}/${resourceID}`

  const revision = new system.Revision({
    changeID: activeDraftKey.value,
    timestamp: new Date().toISOString(),
    resource,
    revision: record.value.revision || 0,
    operation: isNew.value ? 'created' : 'updated',
    status: 'draft',
    userID: String($auth.user?.userID || '0'),
    changes,
    comment: '',
    record: record.value.clone(),
  })

  draftsStore.saveDraft({ revision })
}

function computeChanges() {
  const changes = []
  const current = record.value?.values || {}
  const initial = initialRecordState.value?.values || {}
  const allKeys = new Set([...Object.keys(current), ...Object.keys(initial)])

  for (const key of allKeys) {
    const oldVal = initial[key]
    const newVal = current[key]
    if (JSON.stringify(oldVal) === JSON.stringify(newVal)) continue
    changes.push({
      key,
      old: valueToArray(oldVal),
      new: valueToArray(newVal),
    })
  }
  return changes
}

function handleAiSearch(query) {
  const { moduleID, namespaceID, pageID } = props.page
  window.dispatchEvent(new CustomEvent('show-chat-modal', { detail: {
    namespace: namespaceID,
    module: moduleID,
    page: pageID,
    prompt: query,
  } }))
}

function valueToArray(value) {
  if (value === undefined || value === null) return []
  if (Array.isArray(value)) return value.map(v => v ?? '')
  return [value]
}

function generateChangeID() {
  return `local-${Date.now()}-${Math.random().toString(36).substring(2, 10)}`
}

const handleFormSubmit = throttle(async function (route = 'page.record') {
  processingAction.value = 'submit'
  processing.value = true

  let rec
  const isNewRecord = record.value?.recordID === NoID
  const queue = []

  ;(blocks.value || []).forEach((b, index) => {
    if (b.kind === 'RecordList' && b.options.editable) {
      const p = new Promise((resolve) => {
        const recordListUniqueID = [props.page.pageID, (record.value || {}).recordID, b.blockID, false].map(v => v || NoID).join('-')
        bus.$emit(`record-line:collect:${recordListUniqueID}`, resolve)
      })
      queue.push(p)
    }
  })

  const pairs = await Promise.all(queue)

  for (const p of pairs) {
    if (p.positionField) {
      let i = 0
      for (const item of p.items) {
        if (!item.r.deletedAt) {
          item.r.values[p.positionField] = i++
        }
      }
    }
  }

  const records = pairs.reduce((acc, cur) => {
    if (cur.idPrefix) {
      const existingIndex = acc.findIndex(({ module }) => module.moduleID === cur.module.moduleID)
      if (existingIndex !== -1) {
        acc[existingIndex].set = cur.items.map(({ r }) => r).filter(({ deletedAt, recordID: rid }) => rid !== NoID || !deletedAt)
      } else {
        acc.push({
          refField: cur.refField,
          set: cur.items.map(({ r }) => r).filter(({ deletedAt, recordID: rid }) => rid !== NoID || !deletedAt),
          module: cur.module,
          idPrefix: cur.idPrefix,
        })
      }
    }
    return acc
  }, [])

  const { recordID: rid = NoID } = record.value || {}
  pairs.push({
    module: props.module,
    items: [{ r: record.value, id: rid === NoID ? 'parent:0' : rid }],
  })

  record.value = record.value.clone()

  return dispatchUiEvent('beforeFormSubmit', record.value, { $records: records })
    .then(() => evaluateLayoutRequiredFields())
    .then(() => validateRecord(pairs))
    .then(() => {
      if (isNewRecord) {
        return $ComposeAPI.recordCreate({ ...record.value, records })
      } else {
        return $ComposeAPI.recordUpdate({ ...record.value, records })
      }
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
      rec = new compose.Record(props.module, r)
    }).then(() => dispatchUiEvent('afterFormSubmit', rec, { $records: records }))
    .then(() => {
      if (activeDraftKey.value) {
        draftsStore.removeDraft({ changeID: activeDraftKey.value })
        activeDraftKey.value = null
      }

      if (rec.valueErrors?.set) {
        record.value = rec.clone()
        initialRecordState.value = rec.clone()
        setWarnings()
        bus.$emit('refetch-records', { recordID: rec.recordID })
        toastWarning(proxy.$t('notification.record.validationWarnings', { errors: errors.value, fields: getValidationErrorFields({ includeWarnings: true }) }))
        processing.value = false
      } else {
        initialRecordState.value = record.value.clone()

        if (props.inModal) {
          emit('handle-record-redirect', { recordID: rec.recordID, recordPageID: props.page.pageID, edit: false })
          bus.$emit('refetch-records', { recordID: rec.recordID })
        } else {
          const relatedRecords = [
            props.module.moduleID,
            ...new Set(records.filter(r => r.module.moduleID !== props.module.moduleID).map(r => r.module.moduleID)),
          ]
          relatedRecords.forEach(moduleID => bus.$emit('module-records-updated', { moduleID }))
          router.push({ name: route, params: { ...route.params, recordID: rec.recordID, edit: false } })
        }

        if (props.page.meta?.notifications?.enabled) {
          toastSuccess(proxy.$t(`notification.record.${isNewRecord ? 'create' : 'update'}Success`))
        }
      }
    }).catch(e => {
      processing.value = false
      toastErrorHandler(proxy.$t(`notification.record.${isNewRecord ? 'create' : 'update'}Failed`))(e)
    })
}, 500)

const handleUndelete = throttle(function () {
  processingAction.value = 'undelete'
  processing.value = true

  return dispatchUiEvent('beforeUndelete')
    .then(() => $ComposeAPI.recordUndelete(record.value))
    .then(() => dispatchUiEvent('afterUndelete'))
    .then(() => {
      bus.$emit('refetch-records')
      toastSuccess(proxy.$t('notification.record.restoreSuccess'))
    }).catch(e => {
      processing.value = false
      toastErrorHandler(proxy.$t('notification.record.restoreFailed'))(e)
    })
}, 500)

function getModuleByID(id) {
  return moduleStore.getByID(id)
}

async function validateRecord(pairs) {
  const layoutRequiredFields = uiStore.layoutRequiredFields || []

  const validators = {}
  for (const p of pairs) {
    let moduleForValidator = p.module
    if (layoutRequiredFields.length > 0) {
      const modifiedFields = p.module.fields.map(field => {
        const isLayoutRequired = layoutRequiredFields.includes(field.name) || layoutRequiredFields.includes(field.fieldID)
        if (isLayoutRequired && !field.isRequired) {
          return new compose.ModuleField({ ...field, isRequired: true })
        }
        return field
      })
      moduleForValidator = new compose.Module({ ...p.module, fields: modifiedFields })
    }
    validators[p.module.resourceID] = validators[p.module.resourceID] || new compose.RecordValidator(moduleForValidator)
  }

  const vRunner = () => {
    const newErrors = new validator.Validated()
    for (const p of pairs) {
      const v = validators[p.module.resourceID]
      const errs = new validator.Validated()
      p.items.forEach(({ r, id }) => {
        if (r.deletedAt) return
        const fields = p.module.fields
          .filter(({ canReadRecordValue, canUpdateRecordValue }) => canReadRecordValue && canUpdateRecordValue)
          .map(({ name }) => name)
        if (fields.length) {
          const err = v.run(r, ...fields)
          if (!err.valid()) {
            err.applyMeta({ id })
            errs.push(...err.set)
          }
        }
      })
      newErrors.push(...errs.set)
    }
    errors.value = newErrors
  }

  vRunner()
  if (errors.value.valid()) return

  await dispatchUiEvent('onFormSubmitError')
  vRunner()
  if (!errors.value.valid()) {
    throw new Error(proxy.$t('notification.record.validationErrors', { fields: getValidationErrorFields() }))
  }
}

function setWarnings() {
  const { set = [] } = record.value?.valueErrors || {}
  const e = errors.value
  set.forEach(s => {
    s.meta = s.meta || {}
    s.meta.isWarning = true
    e.push(s)
  })
  errors.value = e
}

function resetErrors() {
  errors.value = new validator.Validated()
}

function fieldErrors(name) {
  if (!errors.value) return new validator.Validated()
  return errors.value.filterByMeta('field', name)
}

function getValidationErrorFields({ includeWarnings = false, includeErrors = true } = {}) {
  const { set = [] } = errors.value || {}
  const fields = new Set(set.filter(({ meta = {} } = {}) => {
    if (includeWarnings) return meta.isWarning
    if (includeErrors) return !meta.isWarning
    return true
  }).map(d => {
    const fieldName = d.meta.field
    const mod = (d.meta.moduleID && getModuleByID(d.meta.moduleID)) || props.module
    if (mod) {
      const f = mod.fields.find(f => f.name === fieldName)
      if (f?.label) return f.label
    }
    return fieldName
  }))
  return Array.from(fields).join(', ')
}

function dispatchUiEvent(eventType, rec, args = {}) {
  const resourceType = `ui:compose:${getUiEventResourceType.value || 'record-page'}`
  const argsBase = { errors: errors.value, validator: validatorComp.value, ...args }
  const event = compose.RecordEvent(rec || record.value, { eventType, resourceType, args: argsBase })
  return $EventBus.Dispatch(event)
}

async function determineLayout({ pageLayoutID, redirectOnFail = true } = {}) {
  if (isRecordPage.value) resetErrors()

  let expressions = {}
  if (layouts.value.some(({ config = {} }) => config.visibility?.expression)) {
    expressions = await evaluateLayoutExpressions()
  }

  const matchedLayout = layouts.value.find(l => {
    if (pageLayoutID && l.pageLayoutID !== pageLayoutID) return false
    const { expression, roles = [] } = l.config.visibility || {}
    if (expression && !expressions[l.pageLayoutID]) return false
    if (!roles.length) return true
    return $auth.user.roles.some(roleID => roles.includes(roleID))
  })

  if (!matchedLayout) {
    toastWarning(proxy.$t('notification.page.page-layout.notFound.view'))
    if (redirectOnFail) router.go(-1)
    return
  }

  inEditing.value = props.edit
  layout.value = matchedLayout

  if (isRecordPage.value) {
    handleRecordButtons()
  } else {
    const { handle, meta = {} } = matchedLayout
    document.title = (meta.title || props.page.title) || handle || proxy.$t('navigation.noPageTitle')
  }

  return prepareBlocks()
}

async function evaluateLayoutExpressions() {
  const expressions = {}
  const variables = expressionVariables()
  layouts.value.forEach(l => {
    const { config = {} } = l
    if (!config.visibility?.expression) return
    variables.layout = l
    expressions[l.pageLayoutID] = config.visibility.expression
  })
  return $SystemAPI.expressionEvaluate({ variables, expressions }).catch(e => {
    toastErrorHandler(proxy.$t('notification.evaluate.failed'))(e)
    Object.keys(expressions).forEach(key => (expressions[key] = false))
    return expressions
  })
}

function expressionVariables() {
  const r = tempRecord.value || record.value
  return {
    user: $auth.user,
    record: r ? r.serialize() : {},
    screen: {
      width: window.innerWidth,
      height: window.innerHeight,
      userAgent: navigator.userAgent,
      breakpoint: getBreakpoint(),
    },
    oldLayout: layout.value,
    layout: undefined,
    ...(isRecordPage.value && {
      isView: !inEditing.value && !isNew.value,
      isCreate: isNew.value,
      isEdit: inEditing.value && !isNew.value,
    }),
  }
}

function getBreakpoint() {
  const w = window.innerWidth
  if (w < 576) return 'xs'
  if (w < 768) return 'sm'
  if (w < 992) return 'md'
  if (w < 1200) return 'lg'
  if (w < 1400) return 'xl'
  return 'xxl'
}

async function prepareBlocks() {
  blocks.value = undefined
  const tempBlocks = []
  const layoutBlocks = layout.value?.blocks || []
  const tabbedIDs = new Set()

  layoutBlocks.forEach(({ blockID, xywh }) => {
    const block = props.page.blocks.find(b => b.blockID === blockID)
    if (block) {
      block.xywh = xywh
      tempBlocks.push(block)
      if (block.kind === 'Tabs') {
        const { tabs = [] } = block.options
        tabs.forEach(t => {
          if (t.blockID && !layoutBlocks.some(b => b.blockID === t.blockID)) {
            tabbedIDs.add(t.blockID)
          }
        })
      }
    }
  })

  props.page.blocks.forEach(block => {
    if (tabbedIDs.has(block.blockID)) tempBlocks.push(block)
  })

  return evaluateBlocks(tempBlocks)
}

async function evaluateBlocks(pageBlocks) {
  let layoutBlocksExpressions = {}
  if (pageBlocks.some(({ meta = {} }) => (meta.visibility || {}).expression)) {
    layoutBlocksExpressions = await evaluateBlocksExpressions(pageBlocks)
  }

  pageBlocks.forEach(block => {
    const { meta = {} } = block
    const blockID = fetchID(block)
    const visibility = meta.visibility || {}
    const { roles = [] } = visibility
    const validExpression = !visibility.expression || layoutBlocksExpressions[blockID]
    const validRole = !roles.length || $auth.user.roles.some(roleID => roles.includes(roleID))
    const showBlock = block && validExpression && validRole
    if (!block.meta) block.meta = {}
    block.meta.invisible = !showBlock
  })

  pageBlocks.forEach(block => {
    if (block.kind === 'Tabs' && !block.meta.invisible) {
      const { tabs = [] } = block.options
      const hasVisibleTab = tabs.some(t => {
        const b = pageBlocks.find(b2 => fetchID(b2) === t.blockID)
        return b ? !b.meta.invisible : !!t.title
      })
      if (!hasVisibleTab) block.meta.invisible = true
    }
  })

  return pageBlocks
}

async function evaluateBlocksExpressions(pageBlocks) {
  const expressions = {}
  const variables = expressionVariables()
  pageBlocks.forEach(block => {
    const { visibility } = block.meta
    if (!(visibility || {}).expression) return
    expressions[fetchID(block)] = visibility.expression
  })
  return $SystemAPI.expressionEvaluate({ variables, expressions }).catch(e => {
    toastErrorHandler(proxy.$t('notification.evaluate.failed'))(e)
    Object.keys(expressions).forEach(key => (expressions[key] = false))
    return expressions
  })
}

async function evaluateLayoutRequiredFields() {
  if (!layout.value) {
    uiStore.clearLayoutRequiredFields()
    return
  }

  const { config = {} } = layout.value || {}
  const { validation = {} } = config || {}
  const { requiredFields = [] } = validation || {}

  if (requiredFields.length === 0) {
    uiStore.clearLayoutRequiredFields()
    return
  }

  await new Promise(resolve => setTimeout(resolve, 300))

  const { expressions, variables: vars } = prepareLayoutRequiredFieldsData()

  if (Object.keys(expressions).length === 0) {
    const fields = requiredFields.filter(rf => !rf.condition || rf.condition.trim() === '').map(rf => rf.field)
    uiStore.setLayoutRequiredFields(fields)
    return
  }

  return $SystemAPI.expressionEvaluate({ variables: vars, expressions })
    .then(res => {
      const fields = []
      requiredFields.forEach(({ field, condition }) => {
        if (!condition || condition.trim() === '') {
          fields.push(field)
        } else if (res[field]) {
          fields.push(field)
        }
      })
      uiStore.setLayoutRequiredFields(fields)
    }).catch(toastErrorHandler(proxy.$t('notification.record.requiredFields.failed')))
}

function prepareLayoutRequiredFieldsData() {
  const expressions = {}
  const variables = expressionVariables()
  const { requiredFields = [] } = layout.value?.config?.validation || {}
  requiredFields.forEach(({ field, condition }) => {
    if (field && condition && condition.trim() !== '') {
      expressions[field] = condition
    }
  })
  return { expressions, variables }
}
</script>
