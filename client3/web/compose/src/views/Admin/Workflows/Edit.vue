<template>
  <div
    v-if="namespace"
    class="pt-3 d-flex flex-column flex-grow-1"
    style="min-height: 0"
  >
    <Teleport to="#topbar-title">
      {{ form.name || (isEdit ? $t('workflow.edit.title') : $t('workflow.edit.createTitle')) }}
    </Teleport>

    <div
      v-if="loading"
      class="d-flex align-items-center justify-content-center h-100"
    >
      <span class="spinner-border" />
    </div>

    <div
      v-else-if="isEdit && workflowLoaded"
      class="d-flex flex-column flex-grow-1"
      style="min-height: 0"
    >
      <WorkflowEditor
        ref="editorRef"
        :workflow-object="workflowObject"
        :change-detected="false"
        :processing-save="false"
        class="flex-grow-1"
        style="height: 0"
        @save="onEditorSave"
        @delete="handleDelete"
      />

      <editor-toolbar
        :processing="processing"
        :processing-save="processingSave"
        :processing-save-and-close="processingSaveAndClose"
        :processing-clone="processingClone"
        :processing-delete="processingDelete"
        :disable-save="processing"
        @delete="handleDelete()"
        @save="handleToolbarSave()"
        @clone="handleToolbarSave({ isClone: true })"
        @saveAndClose="handleToolbarSave({ closeOnSuccess: true })"
        @back="router.push({ name: 'admin.workflows' })"
      />
    </div>

    <div
      v-else
      class="d-flex flex-column flex-grow-1"
      style="min-height: 0"
      @submit.prevent="handleSave()"
    >
      <div class="container-fluid flex-grow-1 d-flex flex-column" style="min-height: 0">
        <div class="row flex-grow-1" style="min-height: 0">
          <div class="col d-flex flex-column" style="min-height: 0">
            <div class="card shadow-sm d-flex flex-column flex-grow-1" style="min-height: 0">
              <div class="card-header d-flex py-3 align-items-center border-bottom" />

              <div class="overflow-auto p-3" style="flex: 1 1 0%; min-height: 0;">
                <div class="row pb-3">
                  <div class="col-12 col-lg-6">
                    <h5>
                      {{ $t('workflow.edit.generalSettings') }}
                    </h5>

                    <div class="mb-3">
                      <label class="form-label text-primary">
                        {{ $t('workflow.edit.name.label') }} *
                      </label>
                      <input
                        v-model="form.name"
                        class="form-control"
                        type="text"
                        required
                        :placeholder="$t('workflow.edit.name.placeholder')"
                      >
                    </div>

                    <div class="mb-3">
                      <label class="form-label text-primary">
                        {{ $t('workflow.edit.handle.label') }}
                      </label>
                      <input
                        v-model="form.handle"
                        class="form-control font-monospace"
                        type="text"
                        :placeholder="$t('workflow.edit.handle.placeholder')"
                      >
                    </div>

                    <div class="mb-3">
                      <label class="form-label text-primary">
                        {{ $t('workflow.edit.description.label') }}
                      </label>
                      <textarea
                        v-model="form.description"
                        class="form-control"
                        rows="3"
                        :placeholder="$t('workflow.edit.description.placeholder')"
                      />
                    </div>

                    <div class="mb-3">
                      <div class="form-check form-switch">
                        <input
                          id="enabled-toggle"
                          v-model="form.enabled"
                          class="form-check-input"
                          type="checkbox"
                        >
                        <label
                          class="form-check-label"
                          for="enabled-toggle"
                        >
                          {{ $t('workflow.edit.enabled.label') }}
                        </label>
                      </div>
                      <div class="form-text small text-muted">
                        {{ $t('workflow.edit.enabled.description') }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <Teleport to="#admin-toolbar">
        <editor-toolbar
          :processing="processing"
          :processing-save="processingSave"
          :processing-save-and-close="processingSaveAndClose"
          :processing-delete="processingDelete"
          :hide-delete="!isEdit"
          :hide-clone="true"
          :disable-save="!form.name"
          @delete="handleDelete()"
          @save="handleSave()"
          @saveAndClose="handleSave({ closeOnSuccess: true })"
          @back="router.push({ name: 'admin.workflows' })"
        />
      </Teleport>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted, inject } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'
import { composables } from 'corteza-lib/vue/dist'
import EditorToolbar from 'corteza-webapp-compose/src/components/Admin/EditorToolbar'
import WorkflowEditor from 'corteza-lib/vue/dist/WorkflowEditor'

const { useToast } = composables
const { t } = useI18n()
const router = useRouter()

const $AutomationAPI = inject('$AutomationAPI')
const $auth = inject('$auth')

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
  workflowID: {
    type: String,
    required: false,
    default: NoID,
  },
})

const isEdit = computed(() => props.workflowID && props.workflowID !== NoID)

const loading = ref(false)
const workflowLoaded = ref(false)
const processing = ref(false)
const processingSave = ref(false)
const processingSaveAndClose = ref(false)
const processingClone = ref(false)
const processingDelete = ref(false)

const form = reactive({
  name: '',
  handle: '',
  description: '',
  enabled: true,
})

const workflowObject = ref(null)

const { toastSuccess, toastErrorHandler } = useToast()

onMounted(() => {
  if (isEdit.value) {
    fetchWorkflow()
  } else {
    document.title = t('workflow.edit.createTitle')
  }
})

function fetchWorkflow () {
  loading.value = true
  $AutomationAPI.workflowRead({ workflowID: props.workflowID })
    .then((resp) => {
      const wf = resp?.workflow || resp
      form.name = wf.meta?.name || ''
      form.handle = wf.handle || ''
      form.description = wf.meta?.description || ''
      form.enabled = !!wf.enabled
      workflowObject.value = wf
      workflowLoaded.value = true
      document.title = wf.meta?.name || props.workflowID
    })
    .catch((e) => {
      toastErrorHandler(t('workflow.notification.loadFailed'))(e)
      router.push({ name: 'admin.workflows' })
    })
    .finally(() => { loading.value = false })
}

function buildPayload () {
  const userID = $auth?.user?.userID || '0'
  return {
    handle: form.handle || undefined,
    meta: { name: form.name, description: form.description },
    enabled: form.enabled,
    runAs: userID,
    ownedBy: userID,
  }
}

function handleSave ({ closeOnSuccess = false } = {}) {
  processing.value = true
  if (closeOnSuccess) processingSaveAndClose.value = true
  else processingSave.value = true

  const payload = buildPayload()

  const promise = isEdit.value
    ? $AutomationAPI.workflowUpdate({ workflowID: props.workflowID, ...payload })
    : $AutomationAPI.workflowCreate(payload)

  promise
    .then((resp) => {
      toastSuccess(t('workflow.notification.' + (isEdit.value ? 'updated' : 'created')))
      const wf = resp?.workflow || resp
      const id = wf?.workflowID
      if (id && !isEdit.value) {
        router.replace({ name: 'admin.workflows.edit', params: { workflowID: id } })
      }
    })
    .catch(toastErrorHandler(t('workflow.notification.' + (isEdit.value ? 'updateFailed' : 'createFailed'))))
    .finally(() => {
      setTimeout(() => {
        processing.value = false
        processingSave.value = false
        processingSaveAndClose.value = false
      }, 300)
    })
}

function onEditorSave (wf = {}) {
  persistWorkflow(wf, pendingSaveOptions.value)
}

const pendingSaveOptions = ref({})
const editorRef = ref(null)

function handleToolbarSave (options = {}) {
  pendingSaveOptions.value = options
  editorRef.value?.saveWorkflow()
}

function persistWorkflow (wf = {}, { closeOnSuccess = false, isClone = false } = {}) {
  const toggleProcessing = (value = true) => {
    processing.value = value
    if (closeOnSuccess) processingSaveAndClose.value = value
    else if (isClone) processingClone.value = value
    else processingSave.value = value
  }

  toggleProcessing(true)

  const userID = $auth?.user?.userID || '0'
  const payload = {
    handle: wf.handle || undefined,
    meta: {
      name: wf.meta?.name || form.name,
      description: wf.meta?.description || form.description,
    },
    enabled: !!wf.enabled,
    steps: wf.steps || [],
    paths: wf.paths || [],
    triggers: wf.triggers || [],
    runAs: userID,
    ownedBy: userID,
  }

  const promise = isClone
    ? $AutomationAPI.workflowCreate({
      ...payload,
      meta: {
        ...payload.meta,
        name: `${payload.meta.name || payload.handle || 'Workflow'} (copy)`,
      },
    })
    : isEdit.value
      ? $AutomationAPI.workflowUpdate({ workflowID: props.workflowID, ...payload })
      : $AutomationAPI.workflowCreate(payload)

  promise
    .then((resp) => {
      const w = resp?.workflow || resp
      toastSuccess(t('workflow.notification.' + (isClone ? 'cloned' : isEdit.value ? 'updated' : 'created')))
      const id = w?.workflowID
      if (id && (isClone || !isEdit.value)) {
        router.replace({ name: 'admin.workflows.edit', params: { workflowID: id } })
      } else if (closeOnSuccess) {
        router.push({ name: 'admin.workflows' })
      }
    })
    .catch(toastErrorHandler(t('workflow.notification.' + (isClone ? 'cloneFailed' : isEdit.value ? 'updateFailed' : 'createFailed'))))
    .finally(() => {
      setTimeout(() => {
        processing.value = false
        processingSave.value = false
        processingSaveAndClose.value = false
        processingClone.value = false
      }, 300)
    })
}

function handleDelete () {
  processingDelete.value = true
  $AutomationAPI.workflowDelete({ workflowID: props.workflowID })
    .then(() => {
      toastSuccess(t('workflow.notification.deleted'))
      router.push({ name: 'admin.workflows' })
    })
    .catch(toastErrorHandler(t('workflow.notification.deleteFailed')))
    .finally(() => {
      setTimeout(() => { processingDelete.value = false }, 300)
    })
}
</script>
