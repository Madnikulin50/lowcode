<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('automation.sessions.list.title')" />
    <c-resource-list :primary-key="primaryKey" :filter="filter" :sorting="sorting" :pagination="pagination" :fields="fields" :items="items" :row-class="rowClass" :translations="{ notFound: $t('admin.general.notFound'), noItems: $t('general.resource-list.no-items'), loading: $t('automation.sessions.list.loading'), showingPagination: 'general.pagination.showing', singlePluralPagination: 'general.pagination.single', prevPagination: $t('admin.general.pagination.prev'), nextPagination: $t('admin.general.pagination.next'), resourceSingle: $t('label.session.single'), resourcePlural: $t('label.session.plural') }" clickable sticky-header hide-search :hide-total="!pagination.incTotal" class="custom-resource-list-height flex-fill" @row-clicked="handleRowClicked">
      <template #header>
        <div class="mb-0" style="min-width:200px">
          <label class="text-primary mb-1">{{ $t('automation.sessions.list.columns.sessionID') }}</label>
          <c-input-search :value="filter.sessionID" size="sm" @input="filterBySessionID" />
        </div>
        <div class="mb-0" style="min-width:200px">
          <label class="text-primary mb-1">{{ $t('automation.sessions.list.columns.workflowID') }}</label>
          <c-input-search :value="filter.workflowID" size="sm" @input="filterByWorkflowID" />
        </div>
      </template>
      <template #toolbar>
        <div class="col d-flex align-items-center">
          <div class="btn-group btn-group-sm me-2" role="group">
            <template v-for="opt in statusOptions" :key="opt.value">
              <input type="radio" class="btn-check" :id="'status-' + opt.value" :value="opt.value" v-model="filter.status" @change="filterList" />
              <label :for="'status-' + opt.value" class="btn btn-outline-primary btn-sm">{{ opt.text }}</label>
            </template>
          </div>
          <span class="ms-2 text-nowrap">{{ $t('automation.sessions.list.filterForm.sessions.label') }}</span>
        </div>
      </template>
      <template #sessionID="{ item }"><a href="javascript:;" @click="filterBySessionID(item.sessionID)">{{ item.sessionID }}</a></template>
      <template #workflowID="{ item }"><a href="javascript:;" @click="filterByWorkflowID(item.workflowID)">{{ item.workflowID }}</a></template>
      <template #actions="{ item }"><router-link class="btn btn-sm btn-link" :to="{ name: editRoute, params: { [primaryKey]: item[primaryKey] } }"><font-awesome-icon :icon="['far', 'edit']" /></router-link></template>
    </c-resource-list>
  </div>
</template>
<script setup>
import { ref, computed, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
const { CResourceList, CInputSearch } = components
const router = useRouter()
const { t } = useI18n()
const primaryKey = 'sessionID'
const editRoute = 'automation.session.edit'
const filter = reactive({ sessionID: null, workflowID: null, status: 0, completed: 1 })
const sorting = reactive({ sortBy: 'createdAt', sortDesc: true })
const pagination = reactive({ limit: 100, pageCursor: undefined, prevPage: '', nextPage: '', total: 0, page: 1, incTotal: false })
const abortableRequests = []
const tempQuery = ref(undefined)
const statusOptions = computed(() => [{ value: 0, text: t('automation.sessions.list.filterForm.started.label') }, { value: 1, text: t('automation.sessions.list.filterForm.prompted.label') }, { value: 2, text: t('automation.sessions.list.filterForm.suspended.label') }, { value: 3, text: t('automation.sessions.list.filterForm.failed.label') }, { value: 5, text: t('automation.sessions.list.filterForm.canceled.label') }, { value: 4, text: t('automation.sessions.list.filterForm.completed.label') }])
const fields = computed(() => ['sessionID', 'workflowID', 'status', { key: 'eventType', sortable: true }, { key: 'createdAt', sortable: true, formatter: (v) => new Date(v).toLocaleString('en-EN') }].map(c => ({ label: t(`columns.${c.key || c}`), ...(typeof c === 'string' ? { key: c } : c) })))
function items() { return procListResults(window.__AutomationAPI.sessionListCancellable(encodeListParams())) }
function rowClass(item) { return { 'text-primary': item && !!item.completedAt } }
function handleRowClicked(item) { router.push({ name: editRoute, params: { [primaryKey]: item[primaryKey] } }) }
function incLoader() {} function decLoader() {}
function filterList() { pagination.pageCursor = ''; pagination.page = 1; abortRequests(); window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' })) }
function encodeListParams() { const { sortBy, sortDesc } = sorting; const { limit, pageCursor, incTotal } = pagination; const sort = sortBy ? `${sortBy} ${sortDesc ? 'DESC' : 'ASC'}` : undefined; return { limit, sort: pageCursor ? undefined : sort, ...filter, pageCursor, incTotal: incTotal && (!pageCursor || tempQuery.value) } }
function procListResults(p) { const { response, cancel } = p; abortableRequests.push(cancel); incLoader(); return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))]).then(async ([{ set, filter: f }]) => { if (f.incTotal) pagination.total = f.total; if (tempQuery.value) { const query = tempQuery.value; tempQuery.value = undefined; router.replace({ query }); return [] } pagination.pageCursor = undefined; pagination.nextPage = f.nextPage; pagination.prevPage = f.prevPage; decLoader(); return set }) }
function abortRequests() { abortableRequests.forEach(c => c()); abortableRequests.length = 0 }
function filterBySessionID(sessionID) { filter.sessionID = sessionID || null; filterList() }
function filterByWorkflowID(workflowID) { filter.workflowID = workflowID || null; filterList() }
</script>
<style scoped>
.content-header { margin-bottom: 0 !important }
</style>
