<template>
  <div class="container-fluid d-flex flex-column flex-fill pt-2 pb-3">
    <c-content-header :title="$t('federation.nodes.list.title')" />
    <c-resource-list :primary-key="primaryKey" :fields="fields" :filter="filter" :pagination="pagination" :sorting="sorting" :items="items" :translations="{ searchPlaceholder: $t('federation.nodes.list.filterForm.query.placeholder'), notFound: $t('admin.general.notFound'), noItems: $t('general.resource-list.no-items'), loading: $t('federation.nodes.list.loading'), showingPagination: 'general.pagination.showing', singlePluralPagination: 'general.pagination.single', prevPagination: $t('admin.general.pagination.prev'), nextPagination: $t('admin.general.pagination.next') }" clickable sticky-header hide-total class="custom-resource-list-height flex-fill" @row-clicked="handleRowClicked" @search="filterList">
      <template #header>
        <button v-if="canCreate" class="btn btn-primary btn-lg" @click="$router.push({ name: 'federation.nodes.new' })">{{ $t('federation.nodes.list.new') }}</button>
        <button v-if="canCreate" class="btn btn-outline-secondary btn-lg" @click="openPairModal()">{{ $t('federation.nodes.list.pair.label') }}</button>
      </template>
      <template #actions="{ item: n }">
        <div v-if="n.nodeID === n.sharedNodeID && (n.status || '').toLowerCase() === 'pair_requested'" class="dropdown">
          <button class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2" data-bs-toggle="dropdown"><font-awesome-icon :icon="['fas', 'ellipsis-v']" /></button>
          <ul class="dropdown-menu m-0">
            <li><button class="dropdown-item" @click="openConfirmPending(n)"><font-awesome-icon :icon="['fas', 'exclamation-triangle']" class="text-danger" /> {{ $t('federation.nodes.list.pair.confirm') }}</button></li>
          </ul>
        </div>
      </template>
    </c-resource-list>
    <div v-if="pair.modal" class="modal fade show d-block" tabindex="-1" style="background:rgba(0,0,0,0.5)">
      <div class="modal-dialog modal-lg modal-dialog-centered">
        <div class="modal-content px-5">
          <div class="modal-body">
            <div v-if="!pair.status" class="text-center px-5">
              <font-awesome-icon size="7x" :icon="['fas', 'share-alt']" class="text-light mb-2" />
              <h2>{{ $t('federation.nodes.list.pair.status.none.description') }}</h2>
              <div class="input-group input-group-lg mt-5">
                <input v-model="pair.url" type="url" class="form-control" placeholder="" />
                <button class="btn btn-primary" :disabled="!pair.url" @click="pairNode()">{{ pair.processing ? $t('federation.nodes.list.loading') : pair.success ? '✓' : $t('federation.nodes.list.pair.confirm') }}</button>
              </div>
              <p class="mt-4"><strong>{{ $t('federation.nodes.list.pair.note') }}</strong> {{ $t('federation.nodes.list.pair.networkEstablished') }}</p>
            </div>
            <div v-else-if="pair.status === 'pair-successful'" class="text-center px-5">
              <font-awesome-icon size="7x" :icon="['far', 'check-circle']" class="text-light mb-4" />
              <h2>{{ $t('federation.nodes.list.pair.status.pending.description') }}</h2>
            </div>
            <div v-else-if="pair.status === 'confirm-pending'" class="text-center px-5">
              <font-awesome-icon size="7x" :icon="['fas', 'share-alt']" class="text-light mb-4" />
              <h2>{{ $t(pair.node.email ? 'pair.status.confirmPending.description' : 'pair.status.confirmPending.descriptionNoMail', pair.node) }}</h2>
              <button class="btn btn-primary" :disabled="pair.processing" @click="confirmPending()">{{ pair.success ? '✓' : $t('federation.nodes.list.pair.confirm') }}</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, computed, reactive, inject } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import moment from 'moment'
import { components } from 'corteza-lib/vue/dist'
const { CResourceList } = components
const router = useRouter()
const { t } = useI18n()
const primaryKey = 'nodeID'
const editRoute = 'federation.nodes.edit'
const filter = reactive({ query: '', suspended: 0, deleted: 0 })
const sorting = reactive({ sortBy: 'createdAt', sortDesc: true })
const pagination = reactive({ limit: 100, pageCursor: undefined, prevPage: '', nextPage: '', total: 0, page: 1, incTotal: false })
const abortableRequests = []
const tempQuery = ref(undefined)
const pair = reactive({ modal: false, processing: false, success: false, url: '', status: undefined, node: undefined })
const fields = computed(() => [{ key: 'name', sortable: true }, { key: 'status', sortable: true }, { key: 'createdAt', label: 'Created', sortable: true, formatter: (v) => moment(v).fromNow() }, { key: 'actions', class: 'actions', label: '' }].map(c => ({ label: t(`columns.${c.key}`), ...c })))
const canCreate = computed(() => can('federation/', 'node.create'))
function can(resource, operation) { return true }
function items() { return procListResults(window.__FederationAPI.nodeSearchCancellable(encodeListParams())) }
function handleRowClicked(item) { router.push({ name: editRoute, params: { [primaryKey]: item[primaryKey] } }) }
function incLoader() {} function decLoader() {}
function filterList() { pagination.pageCursor = ''; pagination.page = 1; abortRequests(); window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' })) }
function encodeListParams() { const { sortBy, sortDesc } = sorting; const { limit, pageCursor, incTotal } = pagination; const sort = sortBy ? `${sortBy} ${sortDesc ? 'DESC' : 'ASC'}` : undefined; return { limit, sort: pageCursor ? undefined : sort, ...filter, pageCursor, incTotal: incTotal && (!pageCursor || tempQuery.value) } }
function procListResults(p) { const { response, cancel } = p; abortableRequests.push(cancel); incLoader(); return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))]).then(async ([{ set, filter: f }]) => { if (f.incTotal) pagination.total = f.total; if (tempQuery.value) { const query = tempQuery.value; tempQuery.value = undefined; router.replace({ query }); return [] } pagination.pageCursor = undefined; pagination.nextPage = f.nextPage; pagination.prevPage = f.prevPage; decLoader(); return set }) }
function abortRequests() { abortableRequests.forEach(c => c()); abortableRequests.length = 0 }
function openPairModal() { pair.status = undefined; pair.modal = true }
async function pairNode() { pair.processing = true; await window.__FederationAPI.nodeCreate({ pairingURI: pair.url }).then(async node => { await window.__FederationAPI.nodePair(node); pair.url = ''; pair.status = 'pair-successful'; pair.node = node; window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' })) }).finally(() => { pair.processing = false }) }
function openConfirmPending(node) { pair.node = node; pair.status = 'confirm-pending'; pair.modal = true }
async function confirmPending() { pair.processing = true; await window.__FederationAPI.nodeHandshakeConfirm({ nodeID: pair.node.nodeID }).then(() => { pair.success = true; setTimeout(() => { pair.success = false }, 2000); setTimeout(() => { pair.node = undefined; pair.status = undefined; pair.modal = false }, 1000); window.dispatchEvent(new CustomEvent('bv::refresh::table', { detail: 'resource-list' })) }).finally(() => { pair.processing = false }) }
</script>
<style>
.pointer { cursor: pointer }
</style>
