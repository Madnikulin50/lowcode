<template>
  <div
    v-if="namespace"
    class="d-flex flex-column w-100 h-100"
  >
    <Teleport to="#topbar-title">
      {{ pageTitle }}
    </Teleport>

    <Teleport to="#topbar-tools">
      <div>
        <div
          v-if="isEdit"
          class="btn-group btn-group-sm"
          role="group"
        >
          <b-button
            data-test-id="button-visit-namespace"
            variant="primary"
            size="sm"
            :class="{ disabled: !namespaceEnabled }"
            :to="namespaceEnabled ? openNamespace : null"
          >
            {{ $t('visit') }}
            <font-awesome-icon
              :icon="['far', 'eye']"
              class="ms-2"
            />
          </b-button>
          <b-button
            v-if="namespace.canManageNamespace"
            :title="$t('configure')"
            data-test-id="button-visit-admin-panel"
            variant="primary"
            size="sm"
            style="margin-left:2px;"
            @click="goToAdmin"
          >
            <font-awesome-icon
              :icon="['far', 'edit']"
            />
          </b-button>
          <namespace-translator
            v-if="namespace"
            :namespace="namespace"
            :disabled="isNew"
            button-variant="primary"
            style="margin-left:2px;"
          />
        </div>
      </div>
    </Teleport>

    <div class="flex-grow-1 overflow-auto py-3">
      <div class="container-fluid flex-grow-1">
        <div class="card">
          <div
            v-if="isEdit"
            class="card-header d-flex align-items-center gap-1 border-bottom"
          >
            <button
              v-if="namespace.canExportNamespace"
              data-test-id="button-export-namespace"
              class="btn btn-outline-secondary btn-lg"
              @click="exportNamespace"
            >
              {{ $t('export') }}
            </button>

            <c-permissions-button
              v-if="namespace.canGrant"
              data-test-id="button-permissions"
              :title="namespace.name || namespace.slug || namespace.namespaceID"
              :target="namespace.name || namespace.slug || namespace.namespaceID"
              :resource="`corteza::compose:namespace/${namespace.namespaceID}`"
              :button-label="$t('label.permissions')"
              button-variant="outline-secondary"
              class="btn-lg"
            />
          </div>

          <div class="card-body p-3">
            <form>
              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('name.label') }}</label>
                <div class="input-group">
                  <input
                    id="ns-nm"
                    v-model="namespace.name"
                    data-test-id="input-name"
                    class="form-control"
                    :class="{ 'is-invalid': nameState === false }"
                    type="text"
                    required
                    :placeholder="$t('name.placeholder')"
                  />
                  <namespace-translator
                    :namespace="namespace"
                    highlight-key="name"
                    :disabled="isNew"
                  />
                </div>
              </div>

              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('slug.label') }}</label>
                <div class="form-text">{{ $t('slug.description') }}</div>
                <input
                  v-model="namespace.slug"
                  data-test-id="input-slug"
                  class="form-control"
                  :class="{ 'is-invalid': slugState === false }"
                  type="text"
                  required
                  :placeholder="$t('slug.placeholder')"
                />
                <div class="invalid-feedback" :class="{ 'd-block': slugState === false }">
                  {{ $t('slug.invalid-handle-characters') }}
                </div>
              </div>

              <div class="mb-3">
                <div class="form-check mb-3">
                  <input
                    id="ns-enabled"
                    v-model="namespace.enabled"
                    data-test-id="checkbox-enable-namespace"
                    class="form-check-input"
                    type="checkbox"
                  />
                  <label class="form-check-label" for="ns-enabled">{{ $t('enabled.label') }}</label>
                </div>
                <div class="form-check">
                  <input
                    id="ns-application"
                    v-model="isApplication"
                    data-test-id="checkbox-toggle-application"
                    class="form-check-input"
                    type="checkbox"
                    :disabled="!canToggleApplication"
                  />
                  <label class="form-check-label" for="ns-application">{{ $t('application.label') }}</label>
                </div>
              </div>

              <hr>

              <div class="mb-3">
                <div class="form-check">
                  <input
                    id="ns-logo-enabled"
                    v-model="namespace.meta.logoEnabled"
                    data-test-id="checkbox-show-logo"
                    class="form-check-input"
                    type="checkbox"
                  />
                  <label class="form-check-label" for="ns-logo-enabled">{{ $t('logo.show') }}</label>
                </div>
              </div>

              <div
                v-if="namespace.meta.logoEnabled && isEdit"
                class="mb-3"
              >
                <div class="d-flex align-items-center mb-1">
                  <label class="form-label text-primary mb-0 me-1">{{ $t('logo.label') }}</label>
                  <button
                    v-if="logoPreview"
                    data-test-id="button-logo-preview"
                    class="btn btn-link btn-sm d-flex align-items-center border-0 p-0 ms-2"
                    type="button"
                    @click="showLogoPreview"
                  >
                    <font-awesome-icon
                      :icon="['far', 'eye']"
                    />
                  </button>
                  <button
                    v-if="!!namespace.meta.logo"
                    data-test-id="button-logo-reset"
                    class="btn btn-outline-secondary btn-sm py-0 ms-2"
                    type="button"
                    @click="resetLogo()"
                  >
                    {{ $t('logo.reset') }}
                  </button>
                </div>

                <input
                  type="file"
                  class="form-control"
                  accept="image/*"
                  @change="e => { namespaceAssets.logo = e.target.files[0] }"
                />
              </div>

              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('subtitle.label') }}</label>
                <div class="input-group">
                  <input
                    v-model="namespace.meta.subtitle"
                    data-test-id="input-subtitle"
                    class="form-control"
                    type="text"
                    :placeholder="$t('subtitle.placeholder')"
                  />
                  <namespace-translator
                    :namespace="namespace"
                    highlight-key="meta.subtitle"
                    :disabled="isNew"
                  />
                </div>
              </div>

              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('description.label') }}</label>
                <div class="input-group">
                  <textarea
                    v-model="namespace.meta.description"
                    data-test-id="input-description"
                    class="form-control"
                    rows="1"
                    :placeholder="$t('description.placeholder')"
                  ></textarea>
                  <namespace-translator
                    :namespace="namespace"
                    highlight-key="meta.description"
                    :disabled="isNew"
                  />
                </div>
              </div>

              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('prompt.label') }}</label>
                <div class="input-group">
                  <c-rich-text-input
                    v-model="namespace.meta.prompt"
                    :placeholder="$t('prompt.placeholder')"
                    body-class="form-control"
                    min-body-height="10rem"
                    :labels="{
                      urlPlaceholder: $t('content.urlPlaceholder'),
                      ok: $t('content.ok'),
                    }"
                  />
                  <namespace-translator
                    :namespace="namespace"
                    highlight-key="meta.prompt"
                    :disabled="isNew"
                  />
                </div>
              </div>

              <hr>

              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('rag.documents.label', 'AI Knowledge Base (RAG)') }}</label>
                <div class="d-flex gap-2 mb-2">
                  <input
                    ref="ragFileInput"
                    type="file"
                    accept=".txt,.html,.htm,.docx,.pdf"
                    class="form-control form-control-sm"
                    style="max-width: 300px"
                    @change="onRagFileSelected"
                  />
                  <button
                    class="btn btn-outline-primary btn-sm"
                    :disabled="!ragUploadFile || ragUploading"
                    @click="uploadRagDocument"
                  >
                    <span v-if="ragUploading" class="spinner-border spinner-border-sm me-1" />
                    {{ $t('rag.documents.upload', 'Upload') }}
                  </button>
                </div>
                <p class="text-muted small mb-2">{{ $t('rag.documents.formats', 'Supported: TXT, HTML, DOCX, PDF') }}</p>

                <div v-if="ragDocs.length" class="list-group list-group-flush border rounded" style="max-height: 200px; overflow-y: auto;">
                  <div
                    v-for="doc in ragDocs"
                    :key="doc.id"
                    class="list-group-item d-flex justify-content-between align-items-center py-2 px-3"
                  >
                    <div class="text-truncate me-2">
                      <span class="fw-medium">{{ doc.name }}</span>
                      <span class="text-muted ms-2 small">{{ formatRagSize(doc.size) }}</span>
                    </div>
                    <button
                      class="btn btn-outline-danger btn-sm"
                      :disabled="ragDeleting === doc.id"
                      @click="deleteRagDocument(doc)"
                    >
                      <span v-if="ragDeleting === doc.id" class="spinner-border spinner-border-sm" />
                      <font-awesome-icon v-else :icon="['far', 'trash-alt']" />
                    </button>
                  </div>
                </div>
                <div v-else-if="ragLoaded" class="text-muted small">
                  {{ $t('rag.documents.empty', 'No documents uploaded.') }}
                </div>
              </div>

              <hr>

              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('sidebar.configure') }}</label>
                <div class="form-check">
                  <input
                    id="ns-hide-sidebar"
                    v-model="namespace.meta.hideSidebar"
                    data-test-id="checkbox-show-sidebar"
                    class="form-check-input"
                    type="checkbox"
                  />
                  <label class="form-check-label" for="ns-hide-sidebar">{{ $t('sidebar.hide') }}</label>
                </div>
              </div>
            </form>
          </div>

          <div
            v-if="isClone"
            class="card-footer bg-warning"
          >
            {{ $t('cloneWarning.wfInclusion') }}
          </div>
        </div>
      </div>
    </div>

    <editor-toolbar
      :processing="processing"
      :processing-save="processingSave"
      :processing-save-and-close="processingSaveAndClose"
      :processing-clone="processingClone"
      :processing-delete="processingDelete"
      :hide-delete="hideDelete"
      :hide-clone="!isEdit"
      :hide-save="hideSave"
      :disable-save="disableSave"
      @back="router.go(-1)"
      @delete="handleDelete"
      @save="handleSave()"
      @clone="handleClone()"
      @saveAndClose="handleSave({ closeOnSuccess: true })"
    />

    <div
      ref="logoModalEl"
      class="modal fade"
      tabindex="-1"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-body p-1">
            <img
              v-if="logoPreview"
              :src="logoPreview"
              class="img-fluid w-100"
            />
          </div>
        </div>
      </div>
    </div>

    <div
      ref="iconModalEl"
      class="modal fade"
      tabindex="-1"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-body p-1">
            <img
              :src="iconPreview"
              class="img-fluid w-100"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'namespace' } })
import { ref, computed, reactive, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter, onBeforeRouteLeave, onBeforeRouteUpdate } from 'vue-router'

import { composables, useNsI18n } from 'corteza-lib/vue/dist'
import { isEqual } from 'lodash'
import { compose, NoID } from 'corteza-lib/js/dist'
import { url, handle, components } from 'corteza-lib/vue/dist'
import { Modal } from 'bootstrap'
import { useNamespaceStore } from '../../store/namespace'
import { useUiStore } from '../../store/ui'
import { useRBACStore } from 'corteza-lib/vue/dist'
import { useResourceTranslations } from '../../mixins/resource-translations'
import EditorToolbar from 'corteza-webapp-compose/src/components/Admin/EditorToolbar'
import NamespaceTranslator from 'corteza-webapp-compose/src/components/Namespaces/NamespaceTranslator'
import {BButton} from "bootstrap-vue-next";

const { CRichTextInput } = components

const t = useNsI18n()
const { toastSuccess, toastErrorHandler } = composables.useToast()
const route = useRoute()
const router = useRouter()
const nsStore = useNamespaceStore()
const uiStore = useUiStore()
const rbac = useRBACStore()

const { currentLanguage } = useResourceTranslations()

const $auth = window.__auth
const $Settings = window.__settings
const $ComposeAPI = window.__composeAPI
const $SystemAPI = window.__systemAPI

const props = defineProps({
  isClone: {
    type: Boolean,
    default: false,
  },
})

const logoModalEl = ref(null)
const iconModalEl = ref(null)
let logoModal = null
let iconModal = null

const processing = ref(false)
const processingSave = ref(false)
const processingSaveAndClose = ref(false)
const processingClone = ref(false)
const processingDelete = ref(false)

const namespace = ref(undefined)
const initialNamespaceState = ref(undefined)

const namespaceAssets = reactive({
  logo: undefined,
  icon: undefined,
})

const namespaceAssetsInitialState = reactive({
  logo: undefined,
  icon: undefined,
})

const namespaceEnabled = ref(false)

const application = ref(undefined)
const isApplication = ref(false)
const isApplicationInitialState = ref(false)

const can = (resource, action) => rbac.can(resource, action)

const canCreateApplication = computed(() => can('system/', 'application.create'))

const isNew = computed(() => namespace.value?.namespaceID === NoID)

const pageTitle = computed(() => {
  switch (true) {
    case isEdit.value:
      return t('edit')
    case route.name === 'namespace.clone' || route.params.isClone:
      return t('clone')
    default:
      return t('create')
  }
})

const watchKey = computed(() => `${route.params.namespaceID}|${route.name}`)

const openNamespace = computed(() => ({ name: 'pages', params: { slug: (namespace.value?.slug || namespace.value?.namespaceID) } }))

const isEdit = computed(() => route.name === 'namespace.edit' || route.name === 'namespace.clone' || props.isClone)

const logoPreview = computed(() => namespace.value?.meta?.logo || $Settings.attachment('ui.mainLogo'))

const iconPreview = computed(() => namespace.value?.meta?.icon || '')

const nameState = computed(() => namespace.value?.name?.length > 0 ? null : false)

const slugState = computed(() => handle.handleState(namespace.value?.slug || ''))

const canToggleApplication = computed(() => canCreateApplication.value)

const disableSave = computed(() => [nameState.value, slugState.value].includes(false))

const hideSave = computed(() => isEdit.value && !namespace.value?.canUpdateNamespace)

const hideDelete = computed(() => !isEdit.value || !!namespace.value?.deletedAt || (isEdit.value && !namespace.value?.canDeleteNamespace))

watch(watchKey, () => {
  fetchNamespace()
}, { immediate: true })

onBeforeUnmount(() => {
  setDefaultValues()
})

onBeforeRouteUpdate(async () => {
  return checkUnsavedNamespace()
})

onBeforeRouteLeave(async () => {
  return checkUnsavedNamespace()
})

onMounted(() => {
  if (logoModalEl.value) {
    logoModal = new Modal(logoModalEl.value)
  }
  if (iconModalEl.value) {
    iconModal = new Modal(iconModalEl.value)
  }
})

watch(() => namespace.value?.namespaceID, (id) => {
  if (id) fetchRagDocuments()
})

function goToAdmin() {
  router.push({ name: 'admin.modules', params: { slug: namespace.value?.slug || namespace.value?.namespaceID } })
}

function showLogoPreview() {
  logoModal?.show()
}

function showIconPreview() {
  iconModal?.show()
}

async function fetchNamespace() {
  processing.value = true

  const namespaceID = route.params.namespaceID

  namespace.value = undefined
  initialNamespaceState.value = undefined
  application.value = undefined
  isApplication.value = false
  isApplicationInitialState.value = isApplication.value

  if (namespaceID) {
    await nsStore.findByID({ namespaceID })
      .then(ns => {
        namespaceEnabled.value = ns.enabled
        namespace.value = new compose.Namespace(ns)
        fetchApplication()
      })
  } else {
    namespace.value = new compose.Namespace({ enabled: true })
  }

  namespace.value.meta = {
    subtitle: '',
    prompt: '',
    description: '',
    hideSidebar: false,
    logoEnabled: null,
    ...namespace.value.meta,
  }

  initialNamespaceState.value = namespace.value.clone()
  namespaceAssetsInitialState.logo = namespaceAssets.logo
  namespaceAssetsInitialState.icon = namespaceAssets.icon

  document.title = isEdit.value
    ? t('label.app-name.namespace.edit', { label: namespace.value.name || namespace.value.slug })
    : t('label.app-name.namespace.create')

  processing.value = false
}

function exportNamespace() {
  const params = {
    namespaceID: namespace.value.namespaceID,
    filename: encodeURIComponent(namespace.value.name.replace(/\./g, '-')),
  }

  const exportUrl = url.Make({
    url: `${$ComposeAPI.baseURL}${$ComposeAPI.namespaceExportEndpoint(params)}`,
    query: {
      jwt: $auth.accessToken,
    },
  })

  window.open(exportUrl)
}

function fetchApplication() {
  const { namespaceID, slug } = namespace.value
  $SystemAPI.applicationList({ name: slug || namespaceID })
    .then(({ set = [] }) => {
      if (set.length) {
        application.value = set[0]
        isApplication.value = application.value.enabled
        isApplicationInitialState.value = isApplication.value
      }
    })
    .catch(toastErrorHandler(t('notification.namespace.application.fetchFailed')))
}

async function handleSave({ closeOnSuccess = false } = {}) {
  const toggleProcessing = () => {
    processing.value = !processing.value
    if (closeOnSuccess) {
      processingSaveAndClose.value = !processingSaveAndClose.value
    } else {
      processingSave.value = !processingSave.value
    }
  }

  toggleProcessing()

  const resourceTranslationLanguage = currentLanguage.value
  let { namespaceID, name, slug, enabled, meta } = namespace.value
  let assets

  if (namespaceAssets.logo || namespaceAssets.icon) {
    try {
      assets = await uploadAssets()
      meta = { ...meta, ...assets }
      namespaceAssetsInitialState.logo = namespaceAssets.logo
      namespaceAssetsInitialState.icon = namespaceAssets.icon
    } catch (e) {
      toastErrorHandler(t('notification.namespace.assetUploadFailed'))(e)
      toggleProcessing()
      return
    }
  }

  const payload = {
    name,
    slug,
    enabled,
    meta,
    resourceTranslationLanguage,
  }

  if (isEdit.value) {
    try {
      await nsStore.update({ ...payload, namespaceID }).then((ns) => {
        namespaceEnabled.value = ns.enabled
        namespace.value = new compose.Namespace(ns)
        toastSuccess(t('notification.namespace.saved'))
      })
    } catch (e) {
      toastErrorHandler(t('notification.namespace.saveFailed'))(e)
      toggleProcessing()
      return
    }
  } else {
    try {
      await nsStore.create(payload).then((ns) => {
        namespaceEnabled.value = ns.enabled
        namespace.value = new compose.Namespace(ns)
        toastSuccess(t('notification.namespace.saved'))
      })
    } catch (e) {
      toastErrorHandler(t('notification.namespace.createFailed'))(e)
      toggleProcessing()
      return
    }
  }

  await handleApplicationSave()
    .catch(() => toastErrorHandler(t('notification.namespace.createAppFailed')))

  initialNamespaceState.value = namespace.value.clone()
  isApplicationInitialState.value = isApplication.value

  setTimeout(() => {
    toggleProcessing()
  }, 300)

  if (closeOnSuccess) {
    router.push(uiStore.previousPage || { name: 'namespace.manage' })
  } else if (props.isClone) {
    router.push({
      name: 'namespace.edit',
      params: { namespaceID: namespace.value.namespaceID },
    })
  } else if (!isEdit.value && namespace.value.enabled) {
    router.push({
      name: 'pages',
      params: {
        slug: namespace.value.slug || namespace.value.namespaceID,
      },
    })
  }
}

function handleClone() {
  processingClone.value = true

  let { name, slug } = namespace.value

  name = `${name} (${t('cloneSuffix')})`
  slug = ''

  return nsStore.clone({ ...namespace.value, name, slug }).then(({ namespaceID }) => {
    toastSuccess(t('notification.namespace.cloned'))
    router.push({ name: 'namespace.edit', params: { namespaceID, isClone: true } })
  }).catch(e => {
    toastErrorHandler(t('notification.namespace.cloneFailed'))(e)
  }).finally(() => {
    processingClone.value = false
  })
}

function handleDelete() {
  processing.value = true
  processingDelete.value = true

  const { namespaceID } = namespace.value
  const { applicationID } = application.value || {}

  nsStore.delete({ namespaceID })
    .catch(toastErrorHandler(t('notification.namespace.deleteFailed')))
    .then(() => {
      namespace.value.deletedAt = new Date()
      if (applicationID) {
        return $SystemAPI.applicationDelete({ applicationID })
      }
    })
    .then(() => {
      router.push({ name: 'namespace.manage' })
      toastSuccess(t('notification.namespace.deleted'))
    })
    .finally(() => {
      processing.value = false
      processingDelete.value = false
    })
}

async function handleApplicationSave() {
  if (application.value) {
    application.value.name = namespace.value.slug || namespace.value.namespaceID
    application.value.unify.name = namespace.value.name
    application.value.unify.url = `/compose/ns/${application.value.name}/pages`

    let enabled = application.value.enabled
    if (isApplication.value && !application.value.enabled) {
      enabled = true
    } else if (!isApplication.value && application.value.enabled) {
      enabled = false
    }

    application.value.unify.listed = enabled

    application.value.unify.icon = application.value.unify.icon || namespace.value.meta.icon
    application.value.unify.logo = application.value.unify.logo || namespace.value.meta.logo

    return $SystemAPI.applicationUpdate({ ...application.value, enabled })
      .then(app => {
        application.value = app
        isApplication.value = application.value.enabled
      })
      .catch(toastErrorHandler(t('notification.namespace.application.saveFailed')))
  } else if (isApplication.value) {
    const app = {
      name: namespace.value.slug || namespace.value.namespaceID,
      enabled: true,
      unify: {
        name: namespace.value.name,
        listed: true,
        url: `compose/ns/${namespace.value.slug || namespace.value.namespaceID}/pages`,
        icon: namespace.value.meta.icon || $Settings.attachment('ui.iconLogo'),
        logo: namespace.value.meta.logo || $Settings.attachment('ui.mainLogo'),
      },
    }
    return $SystemAPI.applicationCreate({ ...app })
      .then(app => {
        application.value = app
        isApplication.value = application.value.enabled
      })
      .catch(toastErrorHandler(t('notification.namespace.application.createFailed')))
  }
}

async function uploadAssets() {
  const rr = {}

  const rq = async (file) => {
    const formData = new FormData()
    formData.append('upload', file)

    const rsp = await $ComposeAPI.api().request({
      method: 'post',
      url: $ComposeAPI.namespaceUploadEndpoint(),
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    if (rsp.data.error) {
      throw new Error(rsp.data.error.message)
    }
    return rsp.data.response
  }

  const baseURL = $ComposeAPI.baseURL

  if (namespaceAssets.logo) {
    const rsp = await rq(namespaceAssets.logo)
    rr.logo = baseURL + rsp.url
    rr.logoID = rsp.attachmentID
    namespaceAssets.logo = undefined
  }

  if (namespaceAssets.icon) {
    const rsp = await rq(namespaceAssets.icon)
    rr.icon = baseURL + rsp.url
    rr.iconID = rsp.attachmentID
    namespaceAssets.icon = undefined
  }

  return rr
}

function resetLogo() {
  namespace.value.meta.logo = undefined
  namespace.value.meta.logoID = undefined
}

function checkUnsavedNamespace() {
  if (!namespace.value) return true

  const isNewRecord = namespace.value.namespaceID === NoID
  if (isNewRecord || processingClone.value || namespace.value.deletedAt) {
    return true
  }

  const namespaceState = initialNamespaceState.value && !isEqual(namespace.value.clone(), initialNamespaceState.value.clone())
  const appState = isApplication.value !== isApplicationInitialState.value
  const assetsState = !isEqual(namespaceAssets, namespaceAssetsInitialState)

  if (namespaceState || appState || assetsState) {
    return window.confirm(t('editor.unsavedChanges'))
  }
  return true
}

function setDefaultValues() {
  processing.value = false
  processingSaveAndClose.value = false
  processingSave.value = false
  processingClone.value = false
  namespace.value = undefined
  initialNamespaceState.value = undefined
  namespaceAssets.logo = undefined
  namespaceAssets.icon = undefined
  namespaceAssetsInitialState.logo = undefined
  namespaceAssetsInitialState.icon = undefined
  namespaceEnabled.value = false
  application.value = undefined
  isApplication.value = false
  isApplicationInitialState.value = false
}

const ragDocs = ref([])
const ragLoaded = ref(false)
const ragUploadFile = ref(null)
const ragUploading = ref(false)
const ragDeleting = ref(null)
const ragFileInput = ref(null)

function onRagFileSelected(e) {
  ragUploadFile.value = e.target.files?.[0] || null
}

async function fetchRagDocuments() {
  try {
    const { set } = await $ComposeAPI.ragDocumentList({ namespaceID: namespace.value.namespaceID })
    ragDocs.value = set || []
  } finally {
    ragLoaded.value = true
  }
}

async function uploadRagDocument() {
  if (!ragUploadFile.value) return
  ragUploading.value = true
  try {
    await $ComposeAPI.ragDocumentUpload({ namespaceID: namespace.value.namespaceID, file: ragUploadFile.value })
    ragUploadFile.value = null
    if (ragFileInput.value) ragFileInput.value.value = ''
    await fetchRagDocuments()
    toastSuccess(t('rag.documents.uploaded', 'Document uploaded and indexed'))
  } catch (e) {
    toastErrorHandler(t('rag.documents.uploadFailed', 'Upload failed'))(e)
  } finally {
    ragUploading.value = false
  }
}

async function deleteRagDocument(doc) {
  ragDeleting.value = doc.id
  try {
    await $ComposeAPI.ragDocumentDelete({ namespaceID: namespace.value.namespaceID, docID: doc.id })
    await fetchRagDocuments()
  } catch (e) {
    toastErrorHandler(t('rag.documents.deleteFailed', 'Delete failed'))(e)
  } finally {
    ragDeleting.value = null
  }
}

function formatRagSize(bytes) {
  if (!bytes) return ''
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1048576).toFixed(1) + ' MB'
}
</script>
