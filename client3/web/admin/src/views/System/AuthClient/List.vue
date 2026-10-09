<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('system.authclients.list.title')" />
    <c-resource-list :primary-key="primaryKey" :filter="filter" :sorting="sorting" :pagination="pagination" :fields="fields" :items="items" :row-class="genericRowClass" :translations="{ searchPlaceholder: $t('system.authclients.list.filterForm.query.placeholder'), notFound: $t('admin.general.notFound'), noItems: $t('general.resource-list.no-items'), loading: $t('system.authclients.list.loading'), showingPagination: 'general.pagination.showing', singlePluralPagination: 'general.pagination.single', prevPagination: $t('admin.general.pagination.prev'), nextPagination: $t('admin.general.pagination.next'), resourceSingle: $t('label.auth-client.single'), resourcePlural: $t('label.auth-client.plural') }" clickable sticky-header hide-search class="custom-resource-list-height flex-fill" @search="filterList" @row-clicked="handleRowClicked">
      <template #header>
        <button v-if="canCreate" class="btn btn-primary btn-lg" @click="$router.push({ name: 'system.authClient.new' })">{{ $t('system.authclients.list.new') }}</button>
        <c-permissions-button v-if="canGrant" resource="corteza::system:auth-client/*" :button-label="$t('system.authclients.list.permissions')" size="lg" />
      </template>
      <template #toolbar>
        <c-resource-list-status-filter v-model="filter.deleted" data-test-id="filter-deleted-auth-clients" :label="$t('system.authclients.list.filterForm.deleted.label')" :excluded-label="$t('system.authclients.list.filterForm.excluded.label')" :inclusive-label="$t('system.authclients.list.filterForm.inclusive.label')" :exclusive-label="$t('system.authclients.list.filterForm.exclusive.label')" @change="filterList" />
      </template>
      <template #actions="{ item: a }">
        <div v-if="(areActionsVisible({ resource: a, conditions: ['canDeleteAuthClient', 'canGrant'] }) && a.authClientID)" class="dropdown">
          <button class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2" data-bs-toggle="dropdown"><font-awesome-icon :icon="['fas', 'ellipsis-v']" /></button>
          <ul class="dropdown-menu m-0">
            <li><c-permissions-button v-if="canGrant" :title="a.meta.name || a.handle || a.authClientID" :target="a.meta.name || a.handle || a.authClientID" :resource="`corteza::system:auth-client/${a.authClientID}`" :button-label="$t('system.authclients.list.permissions')" class="dropdown-item" /></li>
            <li><c-input-confirm v-if="!a.isDefault && a.canDeleteAuthClient" :text="getActionText(a)" show-icon :icon="getActionIcon(a)" borderless variant="link" size="md" button-class="dropdown-item" icon-class="text-danger" class="w-100" @confirmed="handleDelete(a)" /></li>
          </ul>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>
<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.authclients', keyPrefix: 'list' } })
import { ref, computed, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import moment from 'moment'
import { components } from 'corteza-lib/vue/dist'
const { CResourceList } = components
const router = useRouter()
const { t } = useI18n()
const primaryKey = 'authClientID'
const editRoute = 'system.authClient.edit'
const filter = reactive({ query: '', deleted: 0 })
const sorting = reactive({ sortBy: 'createdAt', sortDesc: true })
const pagination = reactive({ limit: 100, pageCursor: undefined, prevPage: '', nextPage: '', total: 0, page: 1, incTotal: true })
const abortableRequests = []
const tempQuery = ref(undefined)
const fields = computed(() => [{ key: 'meta.name', sortable: false, label: t('columns.meta.name'), accessor: (i) => i.meta?.name }, { key: 'handle', sortable: true, label: t('columns.handle') }, { key: 'enabled', label: t('columns.enabled'), formatter: (v) => v ? 'Yes' : 'No' }, { key: 'createdAt', sortable: true, label: t('columns.createdAt'), formatter: (v) => moment(v).fromNow() }, { key: 'actions', class: 'actions', label: '' }])
const canCreate = computed(() => can('system/', 'auth-client.create'))
const canGrant = computed(() => can('system/', 'grant'))
function can(resource, operation) { return true }
function items() { return procListResults(window.__systemAPI.authClientListCancellable(encodeListParams())) }
function handleRowClicked(item) { router.push({ name: editRoute, params: { [primaryKey]: item[primaryKey] } }) }
function handleDelete(authclient) { handleItemDelete({ resource: { clientID: authclient.authClientID }, resourceName: 'authClient', locale: 'authclient' }) }
function genericRowClass(item) { return { 'text-secondary': item && !!item.deletedAt } }
function incLoader() {}
function decLoader() {}
function filterList() { pagination.pageCursor = ''; pagination.page = 1; abortRequests(); window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' })) }
function encodeListParams() { const { sortBy, sortDesc } = sorting; const { limit, pageCursor, incTotal } = pagination; const sort = sortBy ? `${sortBy} ${sortDesc ? 'DESC' : 'ASC'}` : undefined; return { limit, sort: pageCursor ? undefined : sort, ...filter, pageCursor, incTotal: incTotal && (!pageCursor || tempQuery.value) } }
function procListResults(p) { const { response, cancel } = p; abortableRequests.push(cancel); incLoader(); return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))]).then(async ([{ set, filter: f }]) => { if (f.incTotal) pagination.total = f.total; if (tempQuery.value) { const query = tempQuery.value; tempQuery.value = undefined; router.replace({ query }); return [] } pagination.pageCursor = undefined; pagination.nextPage = f.nextPage; pagination.prevPage = f.prevPage; decLoader(); return set }) }
function abortRequests() { abortableRequests.forEach(c => c()); abortableRequests.length = 0 }
function areActionsVisible({ resource, conditions = [] }) { return conditions.some(c => resource[c]) }
function getActionText(r) { return r.deletedAt ? t('system.authclients.list.undelete') : t('system.authclients.list.delete') }
function getActionIcon(r) { return r.deletedAt ? ['fas', 'trash-restore'] : ['far', 'trash-alt'] }
function handleItemDelete({ resource, resourceName, locale, api = 'system' }) { incLoader(); const { deletedAt = '' } = resource; const method = deletedAt ? `${resourceName}Undelete` : `${resourceName}Delete`; window.__systemAPI[method](resource).finally(() => decLoader()) }
</script>
