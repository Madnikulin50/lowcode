<template>
  <div
    v-if="namespace"
    class="container-fluid d-flex flex-column py-3"
  >
    <Teleport to="#topbar-title">
      {{ $t('navigation.page') }}
    </Teleport>

    <div class="card shadow-sm h-100">
      <div class="card-header d-flex flex-column border-bottom gap-1">
        <div class="d-flex align-items-stretch align-items-sm-center justify-content-between flex-column flex-sm-row gap-1">
          <div class="flex-grow-1">
            <div
              v-if="namespace.canCreatePage"
              class="input-group h-100"
              style="min-width: 300px;"
            >
              <input
                id="name"
                v-model="page.title"
                data-test-id="input-name"
                required
                type="text"
                class="form-control h-100"
                :placeholder="$t('placeholder.title')"
              />
              <button
                data-test-id="button-create-page"
                type="submit"
                class="btn btn-primary btn-sm d-flex align-items-center gap-1"
                :disabled="creatingPage"
                @click="createNewPage"
              >
                {{ $t('label.createPage') }}
                <span
                  v-if="creatingPage"
                  class="spinner-border spinner-border-sm"
                />
              </button>
            </div>
          </div>

          <div class="d-flex justify-content-sm-end flex-fill flex-grow-1">
            <router-link
              :to="{ name: 'admin.pages.rag' }"
              class="btn btn-outline-secondary btn-sm me-2 d-flex align-items-center gap-1"
            >
              <font-awesome-icon :icon="['fas', 'database']" />
              <span>RAG</span>
            </router-link>
            <div
              v-if="namespace.canGrant"
              class="dropdown d-flex align-items-center flex-sm-grow-0 flex-sm-shrink-0 flex-fill"
            >
              <button
                data-test-id="dropdown-permissions"
                class="btn btn-outline-secondary btn-sm dropdown-toggle"
                type="button"
                data-bs-toggle="dropdown"
                aria-expanded="false"
              >
                <font-awesome-icon :icon="['fas', 'lock']" />
                <span>
                  {{ $t('label.permissions') }}
                </span>
              </button>
              <ul class="dropdown-menu m-0">
                <li>
                  <c-permissions-button
                    :resource="`corteza::compose:page/${namespace.namespaceID}/*`"
                    :button-label="$t('label.page')"
                    :show-button-icon="false"
                    class="dropdown-item"
                  />
                </li>
                <li>
                  <c-permissions-button
                    :resource="`corteza::compose:page-layout/${namespace.namespaceID}/*/*`"
                    :button-label="$t('label.pageLayout')"
                    :show-button-icon="false"
                    class="dropdown-item"
                  />
                </li>
              </ul>
            </div>
          </div>
        </div>

        <span class="text-muted font-italic">
          {{ $t('instructions') }}
        </span>
      </div>

      <div
        v-if="processing"
        class="text-center text-muted m-5"
      >
        <div>
          <span class="spinner-border align-middle m-2" />
        </div>
        {{ $t('loading') }}
      </div>

      <page-tree
        v-else
        :value="tree"
        @input="tree = $event"
        :namespace="namespace"
        class="card overflow-auto h-100"
        @reorder="handleReorder"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useStore } from '../../../store'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import PageTree from 'corteza-webapp-compose/src/components/Admin/Page/Tree'
import { compose } from 'corteza-lib/js/dist'

const { t } = useI18n()
const store = useStore()
const router = useRouter()
const route = useRoute()

const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: {
    type: compose.Namespace,
    required: false,
    default: undefined,
  },
})

const tree = ref([])
const page = ref(new compose.Page({ visible: true }))
const processing = ref(false)
const abortableRequests = ref([])
const creatingPage = ref(false)

function loadTree (toggleProcessing = true) {
  if (toggleProcessing) {
    processing.value = true
  }

  const namespaceID = props.namespace?.namespaceID

  const { response, cancel } = $ComposeAPI.pageTreeCancellable({ namespaceID })
  abortableRequests.value.push(cancel)

  response()
    .then((tr) => {
      tree.value = tr.map(p => new compose.Page(p))
    }).catch((e) => {
      if (!axios.isCancel(e)) {
        toastErrorHandler(t('notification.page.listFailed'))(e)
      }
    })
    .finally(() => {
      if (toggleProcessing) {
        processing.value = false
      }
    })
}

function createNewPage () {
  creatingPage.value = true

  const namespaceID = props.namespace?.namespaceID
  page.value.weight = tree.value.length
  store.dispatch('page/create', { ...page.value, namespaceID }).then(({ pageID, title }) => {
    const pageLayout = new compose.PageLayout({ namespaceID, pageID, handle: 'primary', meta: { title } })
    return store.dispatch('pageLayout/create', pageLayout).then(() => {
      router.push({ name: 'admin.pages.builder', params: { pageID } })
    })
  }).catch((e) => {
    if (!axios.isCancel(e)) {
      toastErrorHandler(t('notification.page.saveFailed'))(e)
    }
  }).finally(() => {
    setTimeout(() => {
      creatingPage.value = false
    }, 400)
  })
}

function handleReorder () {
  loadTree(false)
}

function toastErrorHandler (msg) {
  return (e) => {}
}

loadTree()

onMounted(() => {
  document.title = t('label.app-name.page.list', { label: props.namespace?.name, interpolation: { escapeValue: false } })
})

onBeforeUnmount(() => {
  abortableRequests.value.forEach((cancel) => cancel())
  setDefaultValues()
})

function setDefaultValues () {
  tree.value = []
  page.value = {}
  processing.value = false
  abortableRequests.value = []
}
</script>
