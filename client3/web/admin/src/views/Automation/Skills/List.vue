<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('automation.skills.list.title')" />
    <c-resource-list
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="fields"
      :items="items"
      :translations="{
        searchPlaceholder: $t('automation.skills.list.filterForm.query.placeholder'),
        notFound: $t('admin.general.notFound'),
        noItems: $t('automation.skills.list.empty'),
        loading: $t('automation.skills.list.loading'),
        showingPagination: 'general.pagination.showing',
        singlePluralPagination: 'general.pagination.single',
        prevPagination: $t('admin.general.pagination.prev'),
        nextPagination: $t('admin.general.pagination.next'),
        resourceSingle: $t('automation.skills.list.single'),
        resourcePlural: $t('automation.skills.list.plural'),
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
          @click="$router.push({ name: 'automation.skills.new' })"
        >
          {{ $t('automation.skills.list.new') }}
        </button>
        <button
          class="btn btn-outline-primary btn-lg ms-2"
          :disabled="importing"
          @click="fileInput.click()"
        >
          {{ $t('automation.skills.list.import') }}
        </button>
        <input
          ref="fileInput"
          type="file"
          class="d-none"
          accept=".md,.zip,text/markdown,application/zip"
          @change="handleImport"
        >
        <div
          v-if="importError"
          class="alert alert-danger py-2 mt-2 mb-0"
        >
          {{ importError }}
        </div>
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
                :text="$t('automation.skills.list.delete')"
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
defineOptions({ i18nOptions: { namespaces: 'automation.skills' } })
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import moment from 'moment'
import { components } from 'corteza-lib/vue/dist'

const { CResourceList } = components
const router = useRouter()
const { t } = useI18n()

const fileInput = ref(null)
const importing = ref(false)
const importError = ref('')

const primaryKey = 'handle'
const editRoute = 'automation.skills.edit'

const filter = reactive({ query: '' })
const sorting = reactive({ sortBy: 'handle', sortDesc: false })
// the library is small and arrives in one piece: no paging
const pagination = reactive({ limit: 1000, pageCursor: undefined, prevPage: '', nextPage: '', total: 0, page: 1, incTotal: true })

const fields = computed(() => [
  { key: 'handle', label: t('automation.skills.list.columns.handle'), sortable: true },
  { key: 'description', label: t('automation.skills.list.columns.description'), sortable: true },
  { key: 'activeVersion', label: t('automation.skills.list.columns.active'), sortable: true, formatter: (v) => (v ? `v${v}` : '') },
  { key: 'versions', label: t('automation.skills.list.columns.versions'), sortable: true },
  { key: 'resources', label: t('automation.skills.list.columns.files'), sortable: true },
  { key: 'updatedAt', label: t('automation.skills.list.columns.updatedAt'), sortable: true, formatter: (v) => (v ? moment(v).fromNow() : '') },
  { key: 'actions', class: 'actions', label: '' },
])

async function items () {
  const { set = [] } = await window.__AutomationAPI.skillList()

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

async function handleDelete (skill) {
  await window.__AutomationAPI.skillDelete({ handle: skill.handle })
  filterList()
}

// a SKILL.md or a zip with SKILL.md and files; the imported version goes live
async function handleImport (ev) {
  const file = ev.target.files && ev.target.files[0]
  ev.target.value = ''
  if (!file) return

  importError.value = ''
  importing.value = true
  try {
    const content = /\.zip$/i.test(file.name) ? await file.arrayBuffer() : await file.text()
    const skill = await window.__AutomationAPI.skillImport({ content, activate: true })
    router.push({ name: editRoute, params: { [primaryKey]: skill.handle } })
  } catch (e) {
    importError.value = (e && e.response && e.response.data && e.response.data.error && e.response.data.error.message) || String((e && e.message) || e)
  } finally {
    importing.value = false
  }
}
</script>
