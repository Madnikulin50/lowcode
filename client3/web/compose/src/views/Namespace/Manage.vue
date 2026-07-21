<template>
  <div class="container-fluid d-flex flex-column py-3">
    <Teleport to="#topbar-title">
      {{ $t('title') }}
    </Teleport>

    <Teleport to="#topbar-tools">
      <div>
        <b-button
          data-test-id="button-namespace-list"
          variant="primary"
          size="sm"
          :to="{ name: 'namespace.list' }"
        >
          {{ $t('list-view') }}
          <font-awesome-icon
            :icon="['fas', 'columns']"
            class="ms-2"
          />
        </b-button>
      </div>

    </Teleport>

    <c-resource-list
      data-test-id="table-namespaces-list"
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="namespacesFields"
      :items="namespaceList"
      :translations="{
        searchPlaceholder: $t('namespace.searchPlaceholder'),
        notFound: $t('resourceList.notFound', 'Not found'),
        noItems: $t('resourceList.noItems', 'No items'),
        loading: $t('label.loading', 'Loading'),
        showingPagination: $t('resourceList.pagination.showing', 'Showing'),
        singlePluralPagination: $t('resourceList.pagination.single', 'resource'),
        prevPagination: $t('resourceList.pagination.prev', 'Previous'),
        nextPagination: $t('resourceList.pagination.next', 'Next'),
        resourceSingle: $t('label.namespace.single', 'Namespace'),
        resourcePlural: $t('label.namespace.plural', 'Namespaces'),
      }"
      clickable
      sticky-header
      class="h-100 flex-fill"
      @search="filterList"
      @row-clicked="handleRowClicked"
    >
      <template #header>
        <router-link
          v-if="canCreate"
          data-test-id="button-create"
          class="btn btn-primary btn-lg"
          :to="{ name: 'namespace.create' }"
        >
          {{ $t('toolbar.buttons.create') }}
        </router-link>

        <importer-modal
          v-if="canImport"
          @imported="onImported"
          @failed="onFailed"
        />

        <c-permissions-button
          v-if="canGrant"
          resource="corteza::compose:namespace/*"
          :button-label="$t('toolbar.buttons.permissions')"
          size="lg"
        />
      </template>

      <template #enabled="{ item }">
        <font-awesome-icon
          :icon="['fas', item.enabled ? 'check' : 'times']"
        />
      </template>

      <template #changedAt="{ item }">
        {{ $locFullDateTime(item.deletedAt || item.updatedAt || item.createdAt) }}
      </template>

      <template #actions="{ item: n }">
        <div
          v-if="n.canDeleteNamespace || n.canGrant"
          class="dropdown dropstart"
        >
          <button
            class="btn btn-outline-extra-light d-flex align-items-center justify-content-center text-primary border-0 py-2"
            data-bs-toggle="dropdown"
            aria-expanded="false"
          >
            <font-awesome-icon
              :icon="['fas', 'ellipsis-v']"
            />
          </button>
          <div class="dropdown-menu m-0">
            <c-permissions-button
              v-if="n.canGrant"
              :title="n.name || n.slug || n.namespaceID"
              :target="n.name || n.slug || n.namespaceID"
              :resource="`corteza::compose:namespace/${n.namespaceID}`"
              :tooltip="$t('permissions.resources.compose.namespace.tooltip')"
              :button-label="$t('permissions.ui.label')"
              class="dropdown-item"
            />

            <c-input-confirm
              v-if="n.canDeleteNamespace"
              :text="$t('delete')"
              show-icon
              borderless
              variant="link"
              size="md"
              button-class="dropdown-item"
              icon-class="text-danger"
              class="w-100"
              @confirmed="handleDelete(n)"
            />
          </div>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
import { ref, computed, reactive, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { composables } from 'corteza-lib/vue/dist'
import { useNamespaceStore } from '../../store/namespace'
import { useRBACStore } from 'corteza-lib/vue/dist'
import ImporterModal from 'corteza-webapp-compose/src/components/Namespaces/Importer'
import axios from 'axios'
import {Portal} from "portal-vue";
import {BButton} from "bootstrap-vue-next";

const { t } = useI18n()
const { toastSuccess, toastErrorHandler } = composables.useToast()
const router = useRouter()
const route = useRoute()
const nsStore = useNamespaceStore()
const rbac = useRBACStore()
const $ComposeAPI = window.__composeAPI
const $SystemAPI = window.__systemAPI

const primaryKey = 'namespaceID'
const application = ref(undefined)
const isApplication = ref(false)

const filter = reactive({
  query: '',
})

const sorting = reactive({
  sortBy: 'name',
  sortDesc: false,
})

const pagination = reactive({
  limit: 100,
  pageCursor: undefined,
  prevPage: '',
  nextPage: '',
  total: 0,
  page: 1,
})

const tempQuery = ref(undefined)
const abortableRequests = ref([])
const cancelled = ref(false)

const can = (resource, action) => rbac.can(resource, action)

const canGrant = computed(() => can('compose/', 'grant'))

const canCreate = computed(() => can('compose/', 'namespace.create'))

const canImport = computed(() => can('compose/', 'namespace.create'))

const importNamespaceEndpoint = computed(() => $ComposeAPI.namespaceImportEndpoint({}))

const namespaces = computed(() => nsStore.set)

const namespacesFields = computed(() => [
  {
    key: 'name',
    sortable: true,
    label: t('table.columns.name'),
  },
  {
    key: 'slug',
    sortable: true,
    label: t('table.columns.slug'),
    class: 'text-nowrap',
  },
  {
    key: 'enabled',
    label: t('table.columns.enabled'),
    class: 'text-center',
  },
  {
    key: 'changedAt',
    sortable: true,
    label: t('table.columns.changedAt'),
    class: 'text-end text-nowrap',
  },
  {
    key: 'actions',
    label: '',
    tdClass: 'text-end text-nowrap actions',
  },
])

watch(() => route.fullPath, () => {
  handleQueryParams()
})

handleQueryParams(true)

onMounted(() => {
  document.title = t('label.app-name.namespace.list')
})

onBeforeUnmount(() => {
  abortRequests()
})

// ---- listHelpers mixin methods ----

function handleQueryParams(initial = false) {
  let {
    limit = pagination.limit,
    pageCursor = pagination.pageCursor,
    prevPage = pagination.prevCursor,
    nextPage = pagination.nextCursor,
    total = pagination.total,
    page = pagination.page,
    ...r1
  } = route.query

  limit = parseInt(limit)
  total = parseInt(total)
  page = parseInt(page)

  if (initial && pageCursor) {
    tempQuery.value = route.query
    router.replace({ query: { ...route.query, limit: 1, pageCursor: undefined } })
    return
  }

  const refresh = route.query.pageCursor !== pagination.pageCursor
  pagination.limit = limit
  pagination.pageCursor = pageCursor
  pagination.prevPage = prevPage
  pagination.nextPage = nextPage
  pagination.total = total
  pagination.page = page

  let { sortBy = sorting.sortBy, sortDesc = sorting.sortDesc, ...r2 } = r1

  sortDesc = sortDesc === true || sortDesc === 'true'

  if (!initial && (sortBy !== sorting.sortBy || sortDesc !== sorting.sortDesc)) {
    pagination.pageCursor = ''
    pagination.page = 1
  }
  sorting.sortBy = sortBy
  sorting.sortDesc = sortDesc

  for (const key in r2) {
    if (typeof filter[key] === 'boolean') {
      r2[key] = r2[key] === 'true'
    }
  }
  Object.assign(filter, { ...filter, ...r2 })

  if (refresh) {
    window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' }))
  }
}

function filterList() {
  pagination.pageCursor = ''
  pagination.page = 1
  abortRequests()
  window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' }))
}

function encodeListParams() {
  let { sortBy, sortDesc } = sorting
  const { limit, pageCursor } = pagination

  if (sortBy === 'changedAt') {
    sortBy = 'coalesce(deletedAt, updatedAt, createdAt)'
  }

  const sort = sortBy ? `${sortBy} ${sortDesc ? 'DESC' : 'ASC'}` : undefined

  return {
    limit,
    sort: pageCursor ? undefined : sort,
    ...filter,
    pageCursor,
    incTotal: !pageCursor || tempQuery.value,
  }
}

function encodeRouteParams() {
  const { limit, pageCursor, page } = pagination
  return {
    query: {
      limit,
      ...sorting,
      ...filter,
      page,
      pageCursor,
    },
  }
}

function procListResults(p, updateQuery = true) {
  abortRequests()

  const { response, cancel } = p
  abortableRequests.value.push(cancel)

  if (updateQuery && !tempQuery.value) {
    router.replace(encodeRouteParams())
  }

  return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
    .then(async ([{ set, filter: f }]) => {
      if (f.incTotal) {
        pagination.total = f.total
      }

      if (tempQuery.value) {
        const query = tempQuery.value
        tempQuery.value = undefined
        router.replace({ query })
        return []
      }

      pagination.pageCursor = undefined
      pagination.nextPage = f.nextPage
      pagination.prevPage = f.prevPage

      return set
    }).catch(error => {
      if (!axios.isCancel(error)) {
        toastErrorHandler(t('notification.list.load.error'))(error)
      } else {
        cancelled.value = true
      }
    }).finally(() => {
      cancelled.value = false
    })
}

function genericRowClass(item) {
  return { 'text-secondary': item && !!item.deletedAt }
}

function abortRequests() {
  abortableRequests.value.forEach((cancel) => {
    cancel()
  })
  abortableRequests.value = []
}

// ---- component methods ----

function onImported() {
  nsStore.load({ force: true })
    .then(() => {
      filterList()
      toastSuccess(t('notification.namespace.imported'))
    })
    .catch(toastErrorHandler(t('notification.namespace.importFailed')))
}

function onFailed(err) {
  toastErrorHandler(t('notification.namespace.importFailed'))(err)
}

function handleRowClicked({ namespaceID }) {
  router.push({
    name: 'namespace.edit',
    params: { namespaceID },
  })
}

function namespaceList() {
  return procListResults($ComposeAPI.namespaceListCancellable(encodeListParams()))
}

function fetchApplication(namespace) {
  const { namespaceID, slug } = namespace
  return $SystemAPI.applicationList({ name: slug || namespaceID })
    .then(({ set = [] }) => {
      if (set.length) {
        application.value = set[0]
        isApplication.value = application.value.enabled
      }
    })
    .catch(toastErrorHandler(t('notification.namespace.deleteFailed')))
}

async function handleDelete(namespace) {
  fetchApplication(namespace).then(() => {
    const { namespaceID } = namespace
    const { applicationID } = application.value || {}
    nsStore.delete({ namespaceID })
      .catch(toastErrorHandler(t('notification.namespace.deleteFailed')))
      .then(() => {
        if (applicationID) {
          return $SystemAPI.applicationDelete({ applicationID })
        }
      })
      .then(() => {
        toastSuccess(t('notification.namespace.deleted'))
        filterList()
      })
  })
}
</script>
