<template>
  <div class="container-xl d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('system.users.list.title')" />

    <c-resource-list
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="fields"
      :items="items"
      :row-class="rowClass"
      :translations="{
        searchPlaceholder: $t('system.users.list.filterForm.query.placeholder'),
        notFound: $t('admin.general.notFound'),
        noItems: $t('admin.general.resource-list.no-items'),
        loading: $t('system.users.list.loading'),
        showingPagination: 'admin.general.pagination.showing',
        singlePluralPagination: 'admin.general.pagination.single',
        prevPagination: $t('admin.general.pagination.prev'),
        nextPagination: $t('admin.general.pagination.next'),
        resourceSingle: $t('general.label.user.single'),
        resourcePlural: $t('general.label.user.plural'),
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
          data-test-id="button-new-user"
          @click="$router.push({ name: 'system.user.new' })"
        >
          {{ $t('system.users.list.new') }}
        </button>

        <c-user-import-modal
          @imported="onImported"
        />

        <c-user-export-modal
          @export="onExport"
        />

        <c-permissions-button
          v-if="canGrant"
          resource="corteza::system:user/*"
                :button-label="$t('system.users.list.permissions')"
          size="lg"
        />

        <c-corredor-manual-buttons
          ui-page="user/list"
          ui-slot="toolbar"
          resource-type="system"
          size="lg"
          @click="dispatchCortezaSystemEvent($event)"
        />
      </template>

      <template #toolbar>
        <c-resource-list-status-filter
          v-model="filter.deleted"
          data-test-id="filter-deleted-users"
          :label="$t('system.users.list.filterForm.deleted.label')"
          :excluded-label="$t('system.users.list.filterForm.excluded.label')"
          :inclusive-label="$t('system.users.list.filterForm.inclusive.label')"
          :exclusive-label="$t('system.users.list.filterForm.exclusive.label')"
          @change="filterList"
        />

        <c-resource-list-status-filter
          v-model="filter.suspended"
          data-test-id="filter-suspended-users"
          :label="$t('system.users.list.filterForm.suspended.label')"
          :excluded-label="$t('system.users.list.filterForm.excluded.label')"
          :inclusive-label="$t('system.users.list.filterForm.inclusive.label')"
          :exclusive-label="$t('system.users.list.filterForm.exclusive.label')"
          @change="filterList"
        />

        <div class="col" />
      </template>

      <template #actions="{ item: u }">
        <div
          v-if="(areActionsVisible({ resource: u, conditions: ['canDeleteUser', 'canGrant'] }))"
          class="dropdown dropstart"
        >
          <button
            class="btn btn-outline-extra-light d-flex align-items-center justify-content-center text-primary border-0 py-2"
            data-bs-toggle="dropdown"
          >
            <font-awesome-icon :icon="['fas', 'ellipsis-v']" />
          </button>
          <ul class="dropdown-menu m-0">
            <li>
              <c-permissions-button
                v-if="canGrant"
                :title="u.name || u.handle || u.email || u.userID"
                :target="u.name || u.handle || u.email || u.userID"
                :resource="`corteza::system:user/${u.userID}`"
          :button-label="$t('system.users.list.permissions')"
                class="dropdown-item"
              />
            </li>
            <li>
              <c-input-confirm
                v-if="u.canDeleteUser"
                :text="getActionText(u)"
                show-icon
                :icon="getActionIcon(u)"
                borderless
                variant="link"
                size="md"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="handleDelete(u)"
              />
            </li>
          </ul>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import moment from 'moment'
import { components, url } from 'corteza-lib/vue/dist'
import CUserExportModal from '../../../components/User/CUserExportModal/index.vue'
import CUserImportModal from '../../../components/User/CUserImportModal/index.vue'

const { CResourceList } = components
const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const id = 'users'
const primaryKey = 'userID'
const editRoute = 'system.user.edit'

const filter = reactive({
  query: '',
  suspended: 0,
  deleted: 0,
})

const sorting = reactive({
  sortBy: 'createdAt',
  sortDesc: true,
})

const pagination = reactive({
  limit: 100,
  pageCursor: undefined,
  prevPage: '',
  nextPage: '',
  total: 0,
  page: 1,
  incTotal: true,
})

const abortableRequests = []
const tempQuery = ref(undefined)

const fields = computed(() => [
  { key: 'name', sortable: true, label: t('columns.name') },
  { key: 'email', sortable: true, label: t('columns.email') },
  { key: 'handle', sortable: true, label: t('columns.handle') },
  { key: 'createdAt', sortable: true, label: t('columns.createdAt'), formatter: (v) => moment(v).fromNow() },
  { key: 'actions', class: 'actions', label: '' },
])

const canGrant = computed(() => can('system/', 'grant'))

function can(resource, operation) {
  return true
}

function onExport(e) {
  const params = { filename: 'export', ...e }
  const exportUrl = url.Make({
    url: `${window.__systemAPI.baseURL}${window.__systemAPI.userExportEndpoint(params)}`,
    query: { jwt: '', inclRoleMembership: e.inclRoleMembership || false, inclRoles: e.inclRoles || false },
  })
  window.open(exportUrl)
}

function onImported() {
  filterList()
}

function items() {
  return procListResults(window.__systemAPI.userListCancellable(encodeListParams()))
}

function rowClass(item) {
  return { 'text-secondary': item && (!!item.deletedAt || !!item.suspendedAt) }
}

function handleRowClicked(item) {
  router.push({ name: editRoute, params: { [primaryKey]: item[primaryKey] } })
}

function handleDelete(user) {
  handleItemDelete({ resource: user, resourceName: 'user' })
}

// List helpers
function incLoader() {}
function decLoader() {}
function filterList() {
  pagination.pageCursor = ''
  pagination.page = 1
  abortRequests()
  window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' }))
}
function encodeListParams() {
  const { sortBy, sortDesc } = sorting
  const { limit, pageCursor, incTotal } = pagination
  const sort = sortBy ? `${sortBy} ${sortDesc ? 'DESC' : 'ASC'}` : undefined
  return { limit, sort: pageCursor ? undefined : sort, ...filter, pageCursor, incTotal: incTotal && (!pageCursor || tempQuery.value) }
}
function procListResults(p) {
  const { response, cancel } = p
  abortableRequests.push(cancel)
  incLoader()
  return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
    .then(async ([{ set, filter: f }]) => {
      if (f.incTotal) pagination.total = f.total
      if (tempQuery.value) {
        const query = tempQuery.value
        tempQuery.value = undefined
        router.replace({ query })
        return []
      }
      pagination.pageCursor = undefined
      pagination.nextPage = f.nextPage
      pagination.prevPage = f.prevPage
      decLoader()
      return set
    })
}
function abortRequests() {
  abortableRequests.forEach(c => c())
  abortableRequests.length = 0
}
function areActionsVisible({ resource, conditions = [] }) {
  return conditions.some(c => resource[c])
}
function getActionText(r) {
  return r.deletedAt ? t('system.users.list.undelete') : t('system.users.list.delete')
}
function getActionIcon(r) {
  return r.deletedAt ? ['fas', 'trash-restore'] : ['far', 'trash-alt']
}
function handleItemDelete({ resource, resourceName, locale, api = 'system' }) {
  incLoader()
  const { deletedAt = '' } = resource
  const method = deletedAt ? `${resourceName}Undelete` : `${resourceName}Delete`
  const API = api === 'system' ? window.__systemAPI : window.__automationAPI
  API[method](resource).finally(() => decLoader())
}
function dispatchCortezaSystemEvent($event) {}
</script>
