<template>
  <div class="container-fluid d-flex flex-column py-3 h-100">
    <Teleport to="#topbar-title">
      {{ $t('anomaly.title') }}
    </Teleport>

    <div class="row mb-3 g-3 flex-shrink-0">
      <div
        v-for="stat in summaryStats"
        :key="stat.key"
        class="col-6 col-md-3"
      >
        <div class="card shadow-sm h-100">
          <div class="card-body text-center py-3">
            <div
              class="anomaly-metric-value"
              :class="stat.class"
            >{{ stat.count }}</div>
            <div class="anomaly-metric-label">{{ stat.label }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="card shadow-sm mb-3 flex-shrink-0">
      <div class="card-header border-bottom bg-transparent">
        <h6 class="mb-0">{{ $t('anomaly.watchedFields') }}</h6>
      </div>
      <div class="card-body p-0">
        <div
          v-if="rulesLoading"
          class="text-muted small p-3"
        >
          <span class="spinner-border spinner-border-sm me-1" />
          {{ $t('anomaly.loading') }}
        </div>

        <p
          v-else-if="!rules.length"
          class="text-muted small fst-italic p-3 mb-0"
        >
          {{ $t('anomaly.noRules') }}
        </p>

        <div
          v-else
          class="table-responsive"
        >
          <table class="table table-sm align-middle mb-0">
            <thead>
              <tr>
                <th class="ps-3">{{ $t('anomaly.module') }}</th>
                <th>{{ $t('anomaly.field') }}</th>
                <th>{{ $t('anomaly.detector') }}</th>
                <th>{{ $t('anomaly.threshold') }}</th>
                <th>{{ $t('anomaly.enabled') }}</th>
                <th class="pe-3">{{ $t('anomaly.lastScanned') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="r in rules"
                :key="r.ruleID"
              >
                <td class="ps-3">{{ moduleName(r.moduleID) }}</td>
                <td>{{ r.field }}</td>
                <td>{{ r.detector }}</td>
                <td>{{ r.threshold }}</td>
                <td>
                  <span
                    class="badge"
                    :class="r.enabled ? 'bg-success' : 'bg-secondary'"
                  >{{ $t(r.enabled ? 'label.yes' : 'label.no') }}</span>
                </td>
                <td class="pe-3">{{ r.lastScannedAt ? new Date(r.lastScannedAt).toLocaleString() : $t('anomaly.neverScanned') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <c-resource-list
      ref="resourceList"
      primary-key="findingID"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="tableFields"
      :items="items"
      :translations="{
        searchPlaceholder: $t('anomaly.searchPlaceholder'),
        notFound: $t('resourceList.notFound', 'Not found'),
        noItems: $t('anomaly.noFindings'),
        loading: $t('anomaly.loading'),
        showingPagination: $t('resourceList.pagination.showing', 'Showing'),
        singlePluralPagination: 'resourceList.pagination.single',
        prevPagination: $t('resourceList.pagination.prev', 'Previous'),
        nextPagination: $t('resourceList.pagination.next', 'Next'),
        resourceSingle: $t('anomaly.findingSingle'),
        resourcePlural: $t('anomaly.findingPlural'),
      }"
      sticky-header
      class="flex-fill"
      @search="handleSearch"
    >
      <template #header>
        <div class="d-flex gap-2">
          <c-input-select
            v-model="filters.moduleID"
            :options="moduleOptions"
            label="label"
            :reduce="o => o.value"
            :clearable="false"
            style="min-width: 10rem;"
          />
          <c-input-select
            v-model="filters.severity"
            :options="severityOptions"
            label="label"
            :reduce="o => o.value"
            :clearable="false"
            style="min-width: 9rem;"
          />
          <c-input-select
            v-model="filters.status"
            :options="statusOptions"
            label="label"
            :reduce="o => o.value"
            :clearable="false"
            style="min-width: 9rem;"
          />
        </div>
      </template>

      <template #module="{ item }">
        {{ moduleName(item.moduleID) }}
      </template>

      <template #field="{ item }">
        {{ item.explanation?.field || '-' }}
      </template>

      <template #score="{ item }">
        {{ formatScore(item.score) }}
      </template>

      <template #severity="{ item }">
        <span
          class="badge"
          :class="severityBadgeClass(item.severity)"
        >{{ $t(`anomaly.severity${capitalize(item.severity)}`) }}</span>
      </template>

      <template #status="{ item }">
        {{ $t(`anomaly.status${capitalize(item.status)}`) }}
      </template>

      <template #detected="{ item }">
        {{ item.createdAt ? new Date(item.createdAt).toLocaleString() : '-' }}
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
            <li v-if="isAdminContext">
              <router-link
                :to="{ name: 'admin.modules.record.view', params: { moduleID: item.moduleID, recordID: item.recordID } }"
                class="dropdown-item"
              >
                {{ $t('anomaly.viewRecord') }}
              </router-link>
            </li>
            <li v-if="item.status !== 'acknowledged'">
              <button
                class="dropdown-item"
                @click="setStatus(item, 'acknowledged')"
              >
                {{ $t('anomaly.acknowledge') }}
              </button>
            </li>
            <li v-if="item.status !== 'resolved'">
              <button
                class="dropdown-item"
                @click="setStatus(item, 'resolved')"
              >
                {{ $t('anomaly.resolve') }}
              </button>
            </li>
            <li v-if="item.status !== 'false_positive'">
              <button
                class="dropdown-item"
                @click="setStatus(item, 'false_positive')"
              >
                {{ $t('anomaly.falsePositive') }}
              </button>
            </li>
          </ul>
        </div>
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'

const { t: $t } = useI18n({ useScope: 'global' })
const route = useRoute()

// This view is shared between the admin panel (admin.anomaly) and the
// published-pages navigation (namespace.anomaly) - only the admin route can
// deep-link into the module editor's record view.
const isAdminContext = computed(() => route.name?.startsWith('admin.'))

// Registered in main.js via CortezaAPI('anomaly', ...).
const $AnomalyAPI = window.__anomalyAPI
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
})

const findings = ref([])
const modules = ref([])
const rules = ref([])
const rulesLoading = ref(false)
const resourceList = ref(null)

const filter = ref({ query: '' })
const sorting = ref({ sortBy: 'detected', sortDesc: true })
const pagination = ref({ total: 0, limit: 25, page: 1, prevPage: '', nextPage: '' })

const filters = reactive({
  moduleID: '',
  severity: '',
  status: '',
})

const moduleOptions = computed(() => [
  { value: '', label: $t('anomaly.allModules') },
  ...modules.value.map((m) => ({ value: m.moduleID, label: m.name || m.handle })),
])

const severityOptions = computed(() => [
  { value: '', label: $t('anomaly.allSeverities') },
  { value: 'low', label: $t('anomaly.severityLow') },
  { value: 'medium', label: $t('anomaly.severityMedium') },
  { value: 'high', label: $t('anomaly.severityHigh') },
])

const statusOptions = computed(() => [
  { value: '', label: $t('anomaly.allStatuses') },
  { value: 'new', label: $t('anomaly.statusNew') },
  { value: 'acknowledged', label: $t('anomaly.statusAcknowledged') },
  { value: 'resolved', label: $t('anomaly.statusResolved') },
  { value: 'false_positive', label: $t('anomaly.statusFalsePositive') },
])

const tableFields = computed(() => [
  { key: 'module', label: $t('anomaly.module') },
  { key: 'field', label: $t('anomaly.field') },
  { key: 'score', label: $t('anomaly.score'), sortable: true, class: 'text-end' },
  { key: 'severity', label: $t('anomaly.severity') },
  { key: 'status', label: $t('anomaly.status') },
  { key: 'detected', label: $t('anomaly.detected'), sortable: true, tdClass: 'text-nowrap' },
  { key: 'actions', label: '', tdClass: 'text-end text-nowrap actions' },
])

const summaryStats = computed(() => [
  { key: 'high', label: $t('anomaly.severityHigh'), class: 'text-danger', count: findings.value.filter((f) => f.severity === 'high').length },
  { key: 'medium', label: $t('anomaly.severityMedium'), class: 'text-warning', count: findings.value.filter((f) => f.severity === 'medium').length },
  { key: 'low', label: $t('anomaly.severityLow'), class: 'text-secondary', count: findings.value.filter((f) => f.severity === 'low').length },
  { key: 'new', label: $t('anomaly.statusNew'), class: 'text-primary', count: findings.value.filter((f) => f.status === 'new').length },
])

watch(() => [filters.moduleID, filters.severity, filters.status], () => {
  pagination.value.page = 1
  resourceList.value?.refresh()
})

onMounted(() => {
  fetchModules()
  fetchFindings()
  fetchRules()
})

watch(() => props.namespace?.namespaceID, () => {
  fetchModules()
  fetchFindings()
  fetchRules()
})

let fetchedOnce = false

function fetchModules () {
  const namespaceID = props.namespace?.namespaceID
  if (!namespaceID || !$ComposeAPI) return
  $ComposeAPI.moduleList({ namespaceID, limit: 500 })
    .then(({ set = [] }) => { modules.value = set })
    .catch(() => { modules.value = [] })
}

function fetchFindings () {
  fetchedOnce = false
  resourceList.value?.refresh()
}

function fetchRules () {
  const namespaceID = props.namespace?.namespaceID
  if (!namespaceID || !$AnomalyAPI) return
  rulesLoading.value = true
  $AnomalyAPI.ruleSearch({ namespaceID, limit: 500 })
    .then(({ set = [] }) => { rules.value = set })
    .catch(() => { rules.value = [] })
    .finally(() => { rulesLoading.value = false })
}

function items () {
  if (!fetchedOnce) {
    fetchedOnce = true
    const namespaceID = props.namespace?.namespaceID
    if (!namespaceID || !$AnomalyAPI) return Promise.resolve([])
    return $AnomalyAPI.findingSearch({ namespaceID, limit: 500 })
      .then(({ set = [] }) => {
        findings.value = set
        return sliceFindings()
      })
      .catch(() => {
        findings.value = []
        return []
      })
  }
  return Promise.resolve(sliceFindings())
}

function handleSearch () {
  pagination.value.page = 1
  resourceList.value?.refresh()
}

function sliceFindings () {
  const q = (filter.value.query || '').toLowerCase()
  let list = findings.value.filter((f) => {
    if (filters.moduleID && String(f.moduleID) !== String(filters.moduleID)) return false
    if (filters.severity && f.severity !== filters.severity) return false
    if (filters.status && f.status !== filters.status) return false
    if (q) {
      const field = String(f.explanation?.field || '').toLowerCase()
      if (!field.includes(q) && !moduleName(f.moduleID).toLowerCase().includes(q)) return false
    }
    return true
  })

  const { sortBy, sortDesc } = sorting.value
  list = [...list].sort((a, b) => {
    let r = 0
    if (sortBy === 'score') r = (Number(a.score) || 0) - (Number(b.score) || 0)
    else r = new Date(a.createdAt || 0) - new Date(b.createdAt || 0)
    return sortDesc ? -r : r
  })

  const total = list.length
  const page = pagination.value.page || 1
  const limit = pagination.value.limit || 25
  const maxPage = Math.max(1, Math.ceil(total / limit))

  pagination.value.total = total
  pagination.value.page = page
  pagination.value.prevPage = page > 1 ? String(page - 1) : ''
  pagination.value.nextPage = page < maxPage ? String(page + 1) : ''

  return list.slice((page - 1) * limit, page * limit)
}

function setStatus (finding, status) {
  const namespaceID = props.namespace?.namespaceID
  $AnomalyAPI.findingUpdateStatus({ namespaceID, findingID: finding.findingID, status })
    .then((updated) => {
      const i = findings.value.findIndex((f) => f.findingID === finding.findingID)
      if (i >= 0) findings.value[i] = { ...findings.value[i], ...updated }
      resourceList.value?.refresh()
    })
    .catch((err) => { console.error('failed to update finding status', err) })
}

function moduleName (moduleID) {
  const m = modules.value.find((m) => String(m.moduleID) === String(moduleID))
  return m ? (m.name || m.handle) : moduleID
}

function formatScore (score) {
  const n = Number(score)
  return Number.isFinite(n) ? n.toFixed(2) : score
}

function capitalize (s) {
  if (!s) return ''
  return s.split('_').map((p) => p.charAt(0).toUpperCase() + p.slice(1)).join('')
}

function severityBadgeClass (severity) {
  if (severity === 'high') return 'bg-danger'
  if (severity === 'medium') return 'bg-warning text-dark'
  return 'bg-secondary'
}
</script>

<style scoped lang="scss">
// Mirrors the "hero" role of PageBlocks/Metric/Item.vue (big centered
// number + small muted label) - not the component itself, since that one
// is wired to a page-block's own data/threshold config.
.anomaly-metric-value {
  font-size: 2rem;
  font-weight: 600;
  line-height: 1.2;
}

.anomaly-metric-label {
  color: var(--bs-secondary-color, #6c757d);
  font-size: 0.8rem;
  font-weight: 500;
  margin-top: 0.25rem;
}
</style>
