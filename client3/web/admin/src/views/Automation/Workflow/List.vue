<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('automation.workflows.list.title')" />
    <c-resource-list :primary-key="primaryKey" :filter="filter" :sorting="sorting" :pagination="pagination" :fields="fields" :items="items" :row-class="genericRowClass" :translations="{ searchPlaceholder: $t('automation.workflows.list.filterForm.query.placeholder'), notFound: $t('admin.general.notFound'), noItems: $t('general.resource-list.no-items'), loading: $t('automation.workflows.list.loading'), showingPagination: 'general.pagination.showing', singlePluralPagination: 'general.pagination.single', prevPagination: $t('admin.general.pagination.prev'), nextPagination: $t('admin.general.pagination.next'), resourceSingle: $t('label.workflow.single'), resourcePlural: $t('label.workflow.plural') }" clickable sticky-header class="custom-resource-list-height flex-fill" @search="filterList" @row-clicked="handleRowClicked">
      <template #header>
        <button v-if="canCreate" class="btn btn-primary btn-lg" @click="$router.push({ name: 'automation.workflow.new' })">{{ $t('automation.workflows.list.new') }}</button>
        <c-permissions-button v-if="canGrant" resource="corteza::automation:workflow/*" :button-label="$t('automation.workflows.list.permissions')" size="lg" />
      </template>
      <template #toolbar>
        <c-resource-list-status-filter v-model="filter.deleted" :label="$t('automation.workflows.list.filterForm.deleted.label')" :excluded-label="$t('automation.workflows.list.filterForm.excluded.label')" :inclusive-label="$t('automation.workflows.list.filterForm.inclusive.label')" :exclusive-label="$t('automation.workflows.list.filterForm.exclusive.label')" @change="filterList" />
      </template>
      <template #actions="{ item: w }">
        <div v-if="(areActionsVisible({ resource: w, conditions: ['canGrant', 'canDeleteWorkflow'] }) && w.workflowID)" class="dropdown">
          <button class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2" data-bs-toggle="dropdown" data-bs-boundary="viewport"><font-awesome-icon :icon="['fas', 'ellipsis-v']" /></button>
          <ul class="dropdown-menu m-0">
            <li><c-permissions-button v-if="canGrant" :title="w.meta.name || w.handle || w.workflowID" :target="w.meta.name || w.handle || w.workflowID" :resource="`corteza::automation:workflow/${w.workflowID}`" :button-label="$t('automation.workflows.list.permissions')" class="dropdown-item" /></li>
            <li><c-input-confirm v-if="(w.canDeleteWorkflow && !w.deletedAt) || (w.canUndeleteWorkflow && w.deletedAt)" show-icon :icon="getActionIcon(w)" :text="getActionText(w)" borderless variant="link" size="md" button-class="dropdown-item" icon-class="text-danger" class="w-100" @confirmed="handleDelete(w)" /></li>
          </ul>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>
<script setup>
import { ref, computed, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import moment from 'moment'
import { components } from 'corteza-lib/vue/dist'
const { CResourceList } = components
const router = useRouter()
const { t } = useI18n()
const primaryKey = 'workflowID'
const editRoute = 'automation.workflow.edit'
const filter = reactive({ query: '', deleted: 0, disabled: 1 })
const sorting = reactive({ sortBy: 'createdAt', sortDesc: true })
const pagination = reactive({ limit: 100, pageCursor: undefined, prevPage: '', nextPage: '', total: 0, page: 1, incTotal: true })
const abortableRequests = []
const tempQuery = ref(undefined)
const fields = computed(() => [{ key: 'handle', sortable: true }, { key: 'meta.name', label: t('columns.name'), accessor: (i) => i.meta?.name }, { key: 'enabled', sortable: true, formatter: (v) => v ? 'Yes' : 'No' }, { key: 'createdAt', sortable: true, formatter: (v) => moment(v).fromNow() }, { key: 'actions', class: 'actions', label: '' }])
const canCreate = computed(() => can('automation/', 'workflow.create'))
const canGrant = computed(() => can('automation/', 'grant'))
function can(resource, operation) { return true }
const genericRowClass = (item) => ({ 'text-secondary': item && !!item.deletedAt })
function items() { return procListResults(window.__AutomationAPI.workflowListCancellable(encodeListParams())) }
function handleRowClicked(item) { router.push({ name: editRoute, params: { [primaryKey]: item[primaryKey] } }) }
function handleDelete(workflow) { handleItemDelete({ resource: workflow, resourceName: 'workflow', api: 'automation' }) }
function incLoader() {} function decLoader() {}
function filterList() { pagination.pageCursor = ''; pagination.page = 1; abortRequests(); window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' })) }
function encodeListParams() { const { sortBy, sortDesc } = sorting; const { limit, pageCursor, incTotal } = pagination; const sort = sortBy ? `${sortBy} ${sortDesc ? 'DESC' : 'ASC'}` : undefined; return { limit, sort: pageCursor ? undefined : sort, ...filter, pageCursor, incTotal: incTotal && (!pageCursor || tempQuery.value) } }
function procListResults(p) { const { response, cancel } = p; abortableRequests.push(cancel); incLoader(); return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))]).then(async ([{ set, filter: f }]) => { if (f.incTotal) pagination.total = f.total; if (tempQuery.value) { const query = tempQuery.value; tempQuery.value = undefined; router.replace({ query }); return [] } pagination.pageCursor = undefined; pagination.nextPage = f.nextPage; pagination.prevPage = f.prevPage; decLoader(); return set }) }
function abortRequests() { abortableRequests.forEach(c => c()); abortableRequests.length = 0 }
function areActionsVisible({ resource, conditions = [] }) { return conditions.some(c => resource[c]) }
function getActionText(r) { return r.deletedAt ? t('automation.workflows.list.undelete') : t('automation.workflows.list.delete') }
function getActionIcon(r) { return r.deletedAt ? ['fas', 'trash-restore'] : ['far', 'trash-alt'] }
function handleItemDelete({ resource, resourceName, locale, api = 'system' }) { incLoader(); const { deletedAt = '' } = resource; const method = deletedAt ? `${resourceName}Undelete` : `${resourceName}Delete`; (api === 'automation' ? window.__AutomationAPI : window.__SystemAPI)[method](resource).finally(() => decLoader()) }
</script>
