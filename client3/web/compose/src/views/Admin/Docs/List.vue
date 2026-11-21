<template>
  <div
    v-if="namespace"
    class="container-fluid d-flex flex-column py-3"
  >
    <Teleport to="#topbar-title">
      {{ $t('title') }}
    </Teleport>

    <c-resource-list
      ref="resourceList"
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="tableFields"
      :items="items"
      :translations="{
        searchPlaceholder: $t('searchPlaceholder'),
        notFound: $t('resourceList.notFound', 'Not found'),
        noItems: $t('empty'),
        loading: $t('label.loading', 'Loading'),
        showingPagination: $t('resourceList.pagination.showing', 'Showing'),
        singlePluralPagination: 'resourceList.pagination.single',
        prevPagination: $t('resourceList.pagination.prev', 'Previous'),
        nextPagination: $t('resourceList.pagination.next', 'Next'),
        resourceSingle: $t('title'),
        resourcePlural: $t('title'),
      }"
      clickable
      sticky-header
      class="h-100 flex-fill"
      @search="handleSearch"
      @row-clicked="handleRowClicked"
    >
      <template #header>
        <div class="dropdown">
          <button
            class="btn btn-primary dropdown-toggle"
            type="button"
            data-bs-toggle="dropdown"
            aria-expanded="false"
          >
            {{ $t('add') }}
          </button>
          <ul class="dropdown-menu m-0">
            <li>
              <router-link
                :to="{ name: 'admin.documents.create', query: { kind: 'markdown' } }"
                class="dropdown-item"
              >
                {{ $t('newMarkdown') }}
              </router-link>
            </li>
            <li>
              <router-link
                :to="{ name: 'admin.documents.create', query: { kind: 'pdf' } }"
                class="dropdown-item"
              >
                {{ $t('newPdf') }}
              </router-link>
            </li>
          </ul>
        </div>
      </template>

      <template #title="{ item }">
        <font-awesome-icon
          :icon="item.kind === 'pdf' ? ['far', 'file-pdf'] : ['fas', 'file-lines']"
          class="me-2 text-secondary"
        />
        {{ item.title }}
      </template>

      <template #kind="{ item }">
        {{ $t(item.kind === 'pdf' ? 'kind.pdf' : 'kind.markdown') }}
      </template>

      <template #visible="{ item }">
        {{ item.visible ? $t('field.visible') : $t('hidden') }}
      </template>

      <template #actions="{ item }">
        <div class="d-flex justify-content-end gap-1">
          <button
            type="button"
            class="btn btn-outline-secondary btn-sm"
            :disabled="!canMove(item, -1)"
            @click.stop="move(item, -1)"
          >
            <font-awesome-icon :icon="['fas', 'chevron-up']" />
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary btn-sm"
            :disabled="!canMove(item, 1)"
            @click.stop="move(item, 1)"
          >
            <font-awesome-icon :icon="['fas', 'chevron-down']" />
          </button>
          <c-input-confirm
            :text="$t('delete')"
            show-icon
            borderless
            variant="link"
            size="md"
            button-class="dropdown-item"
            icon-class="text-danger"
            @confirmed="handleDelete(item)"
          />
        </div>
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'document' } })
import { ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { composables } from 'corteza-lib/vue/dist'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const { toastErrorHandler } = composables.useToast()
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: { type: Object, required: false, default: undefined },
})

const primaryKey = 'documentID'
const filter = ref({ query: '' })
const sorting = ref({ sortBy: 'weight', sortDesc: false })
const pagination = ref({ total: 0, limit: 25, page: 1, prevPage: '', nextPage: '' })
const documents = ref([])
const resourceList = ref(null)
let fetchedOnce = false

const tableFields = computed(() => [
  { key: 'title', label: t('document.columns.title'), sortable: true },
  { key: 'kind', label: t('document.columns.kind'), sortable: true },
  { key: 'handle', label: t('document.columns.handle') },
  { key: 'visible', label: t('document.columns.visible') },
  { key: 'actions', label: '', tdClass: 'text-end text-nowrap actions' },
])

watch(() => props.namespace?.namespaceID, () => {
  fetchedOnce = false
  resourceList.value?.refresh()
})

watch(() => route.query.page, (page) => {
  pagination.value.page = Number(page) || 1
}, { immediate: true })

function items () {
  if (!fetchedOnce) {
    fetchedOnce = true
    return $ComposeAPI.documentList({ namespaceID: props.namespace?.namespaceID })
      .then(({ set }) => {
        documents.value = set || []
        return sliceDocuments()
      })
      .catch((e) => {
        toastErrorHandler(t('document.notification.listFailed'))(e)
        return []
      })
  }
  return Promise.resolve(sliceDocuments())
}

function handleSearch () {
  pagination.value.page = 1
  router.replace({ query: { ...route.query, page: '1' } })
  resourceList.value?.refresh()
}

function ordered () {
  const q = (filter.value.query || '').toLowerCase()
  let list = documents.value
  if (q) {
    list = list.filter(doc => String(doc.title || '').toLowerCase().includes(q) || String(doc.handle || '').toLowerCase().includes(q))
  }
  return list
}

function sliceDocuments () {
  const list = ordered()
  const total = list.length
  const page = parseInt(String(route.query.page || '1'), 10) || 1
  const limit = parseInt(String(route.query.limit || '25'), 10) || 25
  const maxPage = Math.max(1, Math.ceil(total / limit))
  pagination.value.total = total
  pagination.value.limit = limit
  pagination.value.page = page
  pagination.value.prevPage = page > 1 ? String(page - 1) : ''
  pagination.value.nextPage = page < maxPage ? String(page + 1) : ''
  return list.slice((page - 1) * limit, page * limit)
}

function handleRowClicked (doc) {
  router.push({ name: 'admin.documents.edit', params: { documentID: doc.documentID }, query: null })
}

function canMove (doc, dir) {
  const list = documents.value
  const index = list.findIndex(item => item.documentID === doc.documentID)
  const target = index + dir
  return index >= 0 && target >= 0 && target < list.length
}

async function move (doc, dir) {
  const list = documents.value.slice()
  const index = list.findIndex(item => item.documentID === doc.documentID)
  const target = index + dir
  if (index < 0 || target < 0 || target >= list.length) return
  const [item] = list.splice(index, 1)
  list.splice(target, 0, item)
  documents.value = list
  try {
    await $ComposeAPI.documentReorder({
      namespaceID: props.namespace?.namespaceID,
      documentIDs: list.map(item => item.documentID),
    })
    window.dispatchEvent(new CustomEvent('compose-documents-changed'))
    resourceList.value?.refresh()
  } catch (e) {
    toastErrorHandler(t('document.notification.listFailed'))(e)
    fetchedOnce = false
    resourceList.value?.refresh()
  }
}

function handleDelete (doc) {
  $ComposeAPI.documentDelete({
    namespaceID: props.namespace?.namespaceID,
    documentID: doc.documentID,
  }).then(() => {
    documents.value = documents.value.filter(item => item.documentID !== doc.documentID)
    window.dispatchEvent(new CustomEvent('compose-documents-changed'))
    resourceList.value?.refresh()
  }).catch(toastErrorHandler(t('document.notification.deleteFailed')))
}
</script>
