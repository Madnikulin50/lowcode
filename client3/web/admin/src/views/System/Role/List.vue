<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('system.roles.list.title')" />
    <c-resource-list
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="fields"
      :items="items"
      :row-class="rowClass"
      :translations="{ searchPlaceholder: $t('system.roles.list.filterForm.query.placeholder'), notFound: $t('admin.general.notFound'), noItems: $t('general.resource-list.no-items'), loading: $t('system.roles.list.loading'), showingPagination: 'general.pagination.showing', singlePluralPagination: 'general.pagination.single', prevPagination: $t('admin.general.pagination.prev'), nextPagination: $t('admin.general.pagination.next'), resourceSingle: $t('label.role.single'), resourcePlural: $t('label.role.plural') }"
      clickable sticky-header
      class="custom-resource-list-height flex-fill"
      @search="filterList"
      @row-clicked="handleRowClicked"
    >
      <template #header>
        <button v-if="canCreate" class="btn btn-primary btn-lg" @click="$router.push({ name: 'system.role.new' })">{{ $t('system.roles.list.new') }}</button>
        <c-permissions-button v-if="canGrant" resource="corteza::system:role/*" :button-label="$t('system.roles.list.permissions')" size="lg" />
        <c-corredor-manual-buttons ui-page="role/list" ui-slot="toolbar" resource-type="system" default-variant="link" size="lg" @click="dispatchCortezaSystemEvent($event)" />
      </template>
      <template #toolbar>
        <div class="d-flex align-items-center flex-wrap gap-1">
          <c-resource-list-status-filter v-model="filter.deleted" data-test-id="filter-deleted-roles" :label="$t('system.roles.list.filterForm.deleted.label')" :excluded-label="$t('system.roles.list.filterForm.excluded.label')" :inclusive-label="$t('system.roles.list.filterForm.inclusive.label')" :exclusive-label="$t('system.roles.list.filterForm.exclusive.label')" @change="filterList" />
          <div class="ml-4"></div>
          <c-resource-list-status-filter v-model="filter.archived" data-test-id="filter-archived-roles" :label="$t('system.roles.list.filterForm.archived.label')" :excluded-label="$t('system.roles.list.filterForm.excluded.label')" :inclusive-label="$t('system.roles.list.filterForm.inclusive.label')" :exclusive-label="$t('system.roles.list.filterForm.exclusive.label')" @change="filterList" />
        </div>
        <div class="col" />
      </template>
      <template #actions="{ item: r }">
        <div v-if="(areActionsVisible({ resource: r, conditions: ['canDeleteRole', 'canGrant'] }) && r.roleID)" class="dropdown">
          <button class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2" data-bs-toggle="dropdown"><font-awesome-icon :icon="['fas', 'ellipsis-v']" /></button>
          <ul class="dropdown-menu m-0">
            <li><c-permissions-button v-if="canGrant" :title="r.name || r.handle || r.roleID" :target="r.name || r.handle || r.roleID" :resource="`corteza::system:role/${r.roleID}`" :button-label="$t('system.roles.list.permissions')" class="dropdown-item" /></li>
            <li><c-input-confirm v-if="r.canDeleteRole" :text="getActionText(r)" show-icon :icon="getActionIcon(r)" borderless variant="link" size="md" button-class="dropdown-item" icon-class="text-danger" class="w-100" @confirmed="handleDelete(r)" /></li>
          </ul>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>
<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.roles', keyPrefix: 'list' } })
import { ref, computed, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import moment from 'moment'
import { components } from 'corteza-lib/vue/dist'
const { CResourceList } = components
const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const primaryKey = 'roleID'
const editRoute = 'system.role.edit'
const filter = reactive({ query: '', archived: 0, deleted: 0 })
const sorting = reactive({ sortBy: 'createdAt', sortDesc: true })
const pagination = reactive({ limit: 100, pageCursor: undefined, prevPage: '', nextPage: '', total: 0, page: 1, incTotal: true })
const abortableRequests = []
const tempQuery = ref(undefined)
const fields = computed(() => [
  { key: 'name', sortable: true, label: t('columns.name') },
  { key: 'handle', sortable: true, label: t('columns.handle') },
  { key: 'createdAt', sortable: true, label: t('columns.createdAt'), formatter: (v) => moment(v).fromNow() },
  { key: 'actions', class: 'actions', label: '' },
])
const canCreate = computed(() => can('system/', 'role.create'))
const canGrant = computed(() => can('system/', 'grant'))
function can(resource, operation) { return true }
function items() { return procListResults(window.__systemAPI.roleListCancellable(encodeListParams())) }
function rowClass(item) { return { 'text-secondary': item && (!!item.deletedAt || !!item.archivedAt) } }
function handleRowClicked(item) { router.push({ name: editRoute, params: { [primaryKey]: item[primaryKey] } }) }
function handleDelete(role) { handleItemDelete({ resource: role, resourceName: 'role' }) }
function incLoader() {}
function decLoader() {}
function filterList() { pagination.pageCursor = ''; pagination.page = 1; abortRequests(); window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' })) }
function encodeListParams() { const { sortBy, sortDesc } = sorting; const { limit, pageCursor, incTotal } = pagination; const sort = sortBy ? `${sortBy} ${sortDesc ? 'DESC' : 'ASC'}` : undefined; return { limit, sort: pageCursor ? undefined : sort, ...filter, pageCursor, incTotal: incTotal && (!pageCursor || tempQuery.value) } }
function procListResults(p) { const { response, cancel } = p; abortableRequests.push(cancel); incLoader(); return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))]).then(async ([{ set, filter: f }]) => { if (f.incTotal) pagination.total = f.total; if (tempQuery.value) { const query = tempQuery.value; tempQuery.value = undefined; router.replace({ query }); return [] } pagination.pageCursor = undefined; pagination.nextPage = f.nextPage; pagination.prevPage = f.prevPage; decLoader(); return set }) }
function abortRequests() { abortableRequests.forEach(c => c()); abortableRequests.length = 0 }
function areActionsVisible({ resource, conditions = [] }) { return conditions.some(c => resource[c]) }
function getActionText(r) { return r.deletedAt ? t('system.roles.list.undelete') : t('system.roles.list.delete') }
function getActionIcon(r) { return r.deletedAt ? ['fas', 'trash-restore'] : ['far', 'trash-alt'] }
function handleItemDelete({ resource, resourceName, locale, api = 'system' }) { incLoader(); const { deletedAt = '' } = resource; const method = deletedAt ? `${resourceName}Undelete` : `${resourceName}Delete`; const API = api === 'system' ? window.__systemAPI : window.__automationAPI; API[method](resource).finally(() => decLoader()) }
function dispatchCortezaSystemEvent($event) {}
</script>
