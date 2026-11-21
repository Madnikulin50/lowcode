<template>
  <div
    v-if="namespace"
    class="container-fluid d-flex flex-column py-3"
  >
    <Teleport to="#topbar-title">
      {{ $t('navigation.workflows') }}
    </Teleport>

    <c-resource-list
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="tableFields"
      :items="items"
      :translations="{
        searchPlaceholder: $t('workflow.searchPlaceholder'),
        notFound: $t('resourceList.notFound', 'Not found'),
        noItems: $t('workflow.noItems', 'No workflows found'),
        loading: $t('label.loading', 'Loading'),
        showingPagination: $t('resourceList.pagination.showing', 'Showing'),
        singlePluralPagination: 'resourceList.pagination.single',
        prevPagination: $t('resourceList.pagination.prev', 'Previous'),
        nextPagination: $t('resourceList.pagination.next', 'Next'),
        resourceSingle: $t('workflow.label.single'),
        resourcePlural: $t('workflow.label.plural'),
      }"
      clickable
      sticky-header
      class="h-100 flex-fill"
      ref="resourceList"
      @search="handleSearch"
      @row-clicked="handleRowClicked"
    >
      <template #header>
        <router-link
          :to="{ name: 'admin.workflows.create' }"
          class="btn btn-primary"
        >
          <font-awesome-icon :icon="['fas', 'plus']" />
          <span class="ms-1">
            {{ $t('workflow.add') }}
          </span>
        </router-link>
      </template>

      <template #name="{ item }">
        <font-awesome-icon
          :icon="['fas', 'project-diagram']"
          class="me-2"
          style="height: 1rem; width: 1rem; opacity: 0.65;"
        />
        {{ item.name }}
      </template>

      <template #enabled="{ item }">
        <span
          class="badge"
          :class="item.enabled ? 'bg-success' : 'bg-secondary'"
        >
          {{ item.enabled ? $t('workflow.enabled') : $t('workflow.disabled') }}
        </span>
      </template>

      <template #actions="{ item }">
        <div class="dropdown">
          <button
            class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2"
            type="button"
            data-bs-toggle="dropdown"
            aria-expanded="false"
          >
            <font-awesome-icon :icon="['fas', 'ellipsis-v']" />
          </button>
          <ul class="dropdown-menu m-0">
            <li>
              <c-input-confirm
                :text="$t('workflow.edit.delete', 'Delete workflow')"
                show-icon
                borderless
                variant="link"
                size="md"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="handleDelete(item)"
              />
            </li>
          </ul>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, inject } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { composables } from 'corteza-lib/vue/dist'

const { useToast } = composables
const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const $AutomationAPI = inject('$AutomationAPI')

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
})

const primaryKey = 'workflowID'

const filter = ref({ query: '' })
const sorting = ref({
  sortBy: 'name',
  sortDesc: false,
})
const pagination = ref({
  total: 0,
  limit: 25,
  page: 1,
  prevPage: '',
  nextPage: '',
})

const workflows = ref([])
const resourceList = ref(null)

const tableFields = computed(() => [
  {
    key: 'name',
    label: t('workflow.columns.name'),
    sortable: true,
    tdClass: 'text-nowrap',
  },
  {
    key: 'handle',
    label: t('workflow.columns.handle'),
    tdClass: 'text-nowrap font-monospace',
  },
  {
    key: 'description',
    label: t('workflow.columns.description'),
  },
  {
    key: 'enabled',
    label: t('workflow.columns.enabled'),
    class: 'text-nowrap',
  },
  {
    key: 'actions',
    label: '',
    tdClass: 'text-end text-nowrap actions',
  },
])

const { toastSuccess, toastErrorHandler } = useToast()

let fetchedOnce = false

onMounted(() => {
  document.title = t('label.app-name.workflows', { label: props.namespace?.name, interpolation: { escapeValue: false } })
})

watch(() => route.query.page, (page) => {
  pagination.value.page = Number(page) || 1
}, { immediate: true })

watch(() => route.query.limit, (limit) => {
  pagination.value.limit = Number(limit) || 25
}, { immediate: true })

function items () {
  if (!fetchedOnce) {
    fetchedOnce = true
    return $AutomationAPI.workflowList({ limit: 500 })
      .then(({ set }) => {
        workflows.value = (set || []).filter(Boolean).map((item) => {
          const w = item?.workflow || item
          return {
            workflowID: w.workflowID || '',
            name: w.meta?.name || w.handle || w.workflowID || '',
            handle: w.handle || '',
            description: w.meta?.description || '',
            enabled: !!w.enabled,
          }
        })
        return sliceWorkflows()
      })
      .catch((e) => {
        toastErrorHandler(t('workflow.notification.listFailed'))(e)
        return []
      })
  }
  return Promise.resolve(sliceWorkflows())
}

function handleSearch () {
  pagination.value.page = 1
  router.replace({ query: { ...route.query, page: '1' } })
  resourceList.value?.refresh()
}

function sliceWorkflows () {
  const q = (filter.value.query || '').toLowerCase()
  let list = workflows.value
  if (q) {
    list = list.filter((w) => (w.name || '').toLowerCase().includes(q) || (w.handle || '').toLowerCase().includes(q) || (w.description || '').toLowerCase().includes(q))
  }
  const { sortBy, sortDesc } = sorting.value
  if (sortBy === 'name') {
    list = [...list].sort((a, b) => {
      const r = String(a.name || '').localeCompare(String(b.name || ''))
      return sortDesc ? -r : r
    })
  }

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

function handleRowClicked (wf) {
  router.push({ name: 'admin.workflows.edit', params: { workflowID: wf.workflowID }, query: null })
}

function handleDelete (wf) {
  $AutomationAPI.workflowDelete({ workflowID: wf.workflowID })
    .then(() => {
      toastSuccess(t('workflow.notification.deleted'))
      items()
    })
    .catch(toastErrorHandler(t('workflow.notification.deleteFailed')))
}
</script>