<template>
  <div
    v-if="namespace"
    class="container-fluid d-flex flex-column py-3"
  >
    <Teleport to="#topbar-title">
      {{ $t('navigation.rulechains') }}
    </Teleport>

    <c-resource-list
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="tableFields"
      :items="items"
      :translations="{
        searchPlaceholder: $t('rulechain.searchPlaceholder'),
        notFound: $t('resourceList.notFound', 'Not found'),
        noItems: $t('rulechain.noItems', 'No rule chains found'),
        loading: $t('label.loading', 'Loading'),
        showingPagination: $t('resourceList.pagination.showing', 'Showing'),
        singlePluralPagination: 'resourceList.pagination.single',
        prevPagination: $t('resourceList.pagination.prev', 'Previous'),
        nextPagination: $t('resourceList.pagination.next', 'Next'),
        resourceSingle: $t('rulechain.label.single'),
        resourcePlural: $t('rulechain.label.plural'),
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
          :to="{ name: 'admin.rulechains.create' }"
          class="btn btn-primary"
        >
          <font-awesome-icon :icon="['fas', 'plus']" />
          <span class="ms-1">
            {{ $t('rulechain.add') }}
          </span>
        </router-link>
      </template>

      <template #name="{ item }">
        <font-awesome-icon
          :icon="['fas', 'random']"
          class="me-2"
          style="height: 1rem; width: 1rem; opacity: 0.65;"
        />
        {{ item.name }}
      </template>

      <template #nodeCount="{ item }">
        <span class="badge bg-secondary">
          {{ item.nodeCount }}
        </span>
      </template>

      <template #edgeCount="{ item }">
        <span class="badge bg-secondary">
          {{ item.edgeCount }}
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
                :text="$t('rulechain.edit.delete', 'Delete rule chain')"
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
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { composables } from 'corteza-lib/vue/dist'

const { useToast } = composables
const { t } = useI18n()
const router = useRouter()
const route = useRoute()

const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
})

const primaryKey = 'id'

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

const chains = ref([])
const resourceList = ref(null)

const tableFields = computed(() => [
  {
    key: 'name',
    label: t('rulechain.columns.name'),
    sortable: true,
    tdClass: 'text-nowrap',
  },
  {
    key: 'description',
    label: t('rulechain.columns.description'),
  },
  {
    key: 'nodeCount',
    label: t('rulechain.columns.nodes'),
    sortable: true,
    class: 'text-end text-nowrap',
  },
  {
    key: 'edgeCount',
    label: t('rulechain.columns.edges'),
    sortable: true,
    class: 'text-end text-nowrap',
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
  document.title = t('label.app-name.rulechains', { label: props.namespace?.name, interpolation: { escapeValue: false } })
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
    return $ComposeAPI.ruleChainList({ limit: 500 })
      .then(({ chains: set }) => {
        chains.value = set || []
        return sliceChains()
      })
      .catch((e) => {
        toastErrorHandler(t('rulechain.notification.listFailed'))(e)
        return []
      })
  }
  return Promise.resolve(sliceChains())
}

function handleSearch () {
  pagination.value.page = 1
  router.replace({ query: { ...route.query, page: '1' } })
  resourceList.value?.refresh()
}

function sliceChains () {
  const q = (filter.value.query || '').toLowerCase()
  let list = chains.value
  if (q) {
    list = list.filter((c) => (c.name || '').toLowerCase().includes(q) || (c.description || '').toLowerCase().includes(q))
  }
  const { sortBy, sortDesc } = sorting.value
  if (sortBy === 'name' || sortBy === 'nodeCount' || sortBy === 'edgeCount') {
    list = [...list].sort((a, b) => {
      let r = 0
      if (sortBy === 'name') r = String(a.name || '').localeCompare(String(b.name || ''))
      else r = (Number(a[sortBy]) || 0) - (Number(b[sortBy]) || 0)
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

function handleRowClicked (chain) {
  router.push({ name: 'admin.rulechains.edit', params: { chainID: chain.id }, query: null })
}

function handleDelete (chain) {
  $ComposeAPI.ruleChainDelete({ chainID: chain.id })
    .then(() => {
      toastSuccess(t('rulechain.notification.deleted'))
      items()
    })
    .catch(toastErrorHandler(t('rulechain.notification.deleteFailed')))
}
</script>
