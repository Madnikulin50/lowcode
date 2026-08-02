<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('system.queues.list.title')" />
    <c-resource-list :primary-key="primaryKey" :filter="filter" :sorting="sorting" :pagination="pagination" :fields="fields" :items="items" :row-class="genericRowClass" :translations="{ searchPlaceholder: $t('system.queues.list.filterForm.handle.placeholder'), notFound: $t('admin.general.notFound'), noItems: $t('general.resource-list.no-items'), loading: $t('system.queues.list.loading'), showingPagination: 'general.pagination.showing', singlePluralPagination: 'general.pagination.single', prevPagination: $t('admin.general.pagination.prev'), nextPagination: $t('admin.general.pagination.next'), resourceSingle: $t('label.queue.single'), resourcePlural: $t('label.queue.plural') }" clickable sticky-header class="custom-resource-list-height flex-fill" @search="filterList" @row-clicked="handleRowClicked">
      <template #header>
        <button v-if="canCreate" class="btn btn-primary btn-lg" @click="$router.push({ name: 'system.queue.new' })">{{ $t('system.queues.list.new') }}</button>
      </template>
      <template #toolbar>
        <c-resource-list-status-filter v-model="filter.deleted" data-test-id="filter-deleted-queues" :label="$t('system.queues.list.filterForm.deleted.label')" :excluded-label="$t('system.queues.list.filterForm.excluded.label')" :inclusive-label="$t('system.queues.list.filterForm.inclusive.label')" :exclusive-label="$t('system.queues.list.filterForm.exclusive.label')" @change="filterList" />
      </template>
      <template #actions="{ item: q }">
        <div v-if="q.canDeleteQueue" class="dropdown">
          <button class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2" data-bs-toggle="dropdown"><font-awesome-icon :icon="['fas', 'ellipsis-v']" /></button>
          <ul class="dropdown-menu m-0"><li><c-input-confirm :text="getActionText(q)" show-icon :icon="getActionIcon(q)" borderless variant="link" size="md" button-class="dropdown-item" icon-class="text-danger" class="w-100" @confirmed="handleDelete(q)" /></li></ul>
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
const primaryKey = 'queueID'
const editRoute = 'system.queue.edit'
const filter = reactive({ query: '', archived: 0, deleted: 0 })
const sorting = reactive({ sortBy: 'createdAt', sortDesc: true })
const pagination = reactive({ limit: 100, pageCursor: undefined, prevPage: '', nextPage: '', total: 0, page: 1, incTotal: true })
const abortableRequests = []
const tempQuery = ref(undefined)
const fields = computed(() => [{ key: 'queue', sortable: true, label: t('columns.queue') }, { key: 'consumer', sortable: false, label: t('columns.consumer') }, { key: 'createdAt', sortable: true, label: t('columns.createdAt'), formatter: (v) => moment(v).fromNow() }, { key: 'actions', label: '', class: 'actions' }])
const canCreate = computed(() => can('system/', 'queue.create'))
function can(resource, operation) { return true }
function items() { return procListResults(window.__systemAPI.queuesListCancellable(encodeListParams())) }
function handleRowClicked(item) { router.push({ name: editRoute, params: { [primaryKey]: item[primaryKey] } }) }
function handleDelete(queue) { handleItemDelete({ resource: queue, resourceName: 'queues', locale: 'queue' }) }
function genericRowClass(item) { return { 'text-secondary': item && !!item.deletedAt } }
function incLoader() {}
function decLoader() {}
function filterList() { pagination.pageCursor = ''; pagination.page = 1; abortRequests(); window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' })) }
function encodeListParams() { const { sortBy, sortDesc } = sorting; const { limit, pageCursor, incTotal } = pagination; const sort = sortBy ? `${sortBy} ${sortDesc ? 'DESC' : 'ASC'}` : undefined; return { limit, sort: pageCursor ? undefined : sort, ...filter, pageCursor, incTotal: incTotal && (!pageCursor || tempQuery.value) } }
function procListResults(p) { const { response, cancel } = p; abortableRequests.push(cancel); incLoader(); return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))]).then(async ([{ set, filter: f }]) => { if (f.incTotal) pagination.total = f.total; if (tempQuery.value) { const query = tempQuery.value; tempQuery.value = undefined; router.replace({ query }); return [] } pagination.pageCursor = undefined; pagination.nextPage = f.nextPage; pagination.prevPage = f.prevPage; decLoader(); return set }) }
function abortRequests() { abortableRequests.forEach(c => c()); abortableRequests.length = 0 }
function getActionText(r) { return r.deletedAt ? t('system.queues.list.undelete') : t('system.queues.list.delete') }
function getActionIcon(r) { return r.deletedAt ? ['fas', 'trash-restore'] : ['far', 'trash-alt'] }
function handleItemDelete({ resource, resourceName, locale, api = 'system' }) { incLoader(); const { deletedAt = '' } = resource; const method = deletedAt ? `${resourceName}Undelete` : `${resourceName}Delete`; window.__systemAPI[method](resource).finally(() => decLoader()) }
</script>
