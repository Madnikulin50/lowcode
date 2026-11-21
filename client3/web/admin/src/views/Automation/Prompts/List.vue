<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('automation.prompts.list.title')" />
    <c-resource-list
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="fields"
      :items="items"
      :translations="{
        searchPlaceholder: $t('automation.prompts.list.filterForm.query.placeholder'),
        notFound: $t('admin.general.notFound'),
        noItems: $t('automation.prompts.list.empty'),
        loading: $t('automation.prompts.list.loading'),
        showingPagination: 'general.pagination.showing',
        singlePluralPagination: 'general.pagination.single',
        prevPagination: $t('admin.general.pagination.prev'),
        nextPagination: $t('admin.general.pagination.next'),
        resourceSingle: $t('automation.prompts.list.single'),
        resourcePlural: $t('automation.prompts.list.plural'),
      }"
      clickable
      sticky-header
      class="custom-resource-list-height flex-fill"
      @search="filterList"
      @row-clicked="handleRowClicked"
    >
      <template #header>
        <button
          class="btn btn-primary btn-lg"
          @click="$router.push({ name: 'automation.prompts.new' })"
        >
          {{ $t('automation.prompts.list.new') }}
        </button>
      </template>

      <template #actions="{ item: p }">
        <div class="dropdown">
          <button
            class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2"
            data-bs-toggle="dropdown"
            data-bs-boundary="viewport"
          >
            <font-awesome-icon :icon="['fas', 'ellipsis-v']" />
          </button>
          <ul class="dropdown-menu m-0">
            <li>
              <c-input-confirm
                show-icon
                :icon="['far', 'trash-alt']"
                :text="$t('automation.prompts.list.delete')"
                borderless
                variant="link"
                size="md"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="handleDelete(p)"
              />
            </li>
          </ul>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'automation.prompts' } })
import { computed, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import moment from 'moment'
import { components } from 'corteza-lib/vue/dist'

const { CResourceList } = components
const router = useRouter()
const { t } = useI18n()

const primaryKey = 'handle'
const editRoute = 'automation.prompts.edit'

const filter = reactive({ query: '' })
const sorting = reactive({ sortBy: 'handle', sortDesc: false })
// the library is small and arrives in one piece: no paging
const pagination = reactive({ limit: 1000, pageCursor: undefined, prevPage: '', nextPage: '', total: 0, page: 1, incTotal: true })

const fields = computed(() => [
  { key: 'handle', label: t('automation.prompts.list.columns.handle'), sortable: true },
  { key: 'description', label: t('automation.prompts.list.columns.description'), sortable: true },
  { key: 'activeVersion', label: t('automation.prompts.list.columns.active'), sortable: true, formatter: (v) => (v ? `v${v}` : '') },
  { key: 'versions', label: t('automation.prompts.list.columns.versions'), sortable: true },
  { key: 'cases', label: t('automation.prompts.list.columns.cases'), sortable: true },
  { key: 'updatedAt', label: t('automation.prompts.list.columns.updatedAt'), sortable: true, formatter: (v) => (v ? moment(v).fromNow() : '') },
  { key: 'actions', class: 'actions', label: '' },
])

async function items () {
  const { set = [] } = await window.__AutomationAPI.promptList()

  const q = filter.query.trim().toLowerCase()
  const rows = set.filter(p => !q || p.handle.toLowerCase().includes(q) || (p.description || '').toLowerCase().includes(q))

  const { sortBy, sortDesc } = sorting
  rows.sort((a, b) => {
    const x = a[sortBy] ?? ''
    const y = b[sortBy] ?? ''
    const order = typeof x === 'number' && typeof y === 'number' ? x - y : String(x).localeCompare(String(y))
    return sortDesc ? -order : order
  })

  pagination.total = rows.length
  return rows
}

function filterList () {
  pagination.pageCursor = ''
  pagination.page = 1
  window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' }))
}

function handleRowClicked (item) {
  router.push({ name: editRoute, params: { [primaryKey]: item[primaryKey] } })
}

async function handleDelete (prompt) {
  await window.__AutomationAPI.promptDelete({ handle: prompt.handle })
  filterList()
}
</script>
