<template>
  <Wrap
    v-bind="$props"
    @refreshBlock="refresh"
  >
    <div class="p-3 h-100 overflow-auto">
      <div
        v-if="options.showSummary"
        class="row g-2 mb-3"
      >
        <div
          v-for="stat in summaryStats"
          :key="stat.key"
          class="col-6 col-md-3"
        >
          <div class="card shadow-sm h-100">
            <div class="card-body text-center py-2">
              <div
                class="anomaly-metric-value"
                :class="stat.class"
              >{{ stat.count }}</div>
              <div class="anomaly-metric-label">{{ stat.label }}</div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="options.showList">
        <div
          v-if="loading"
          class="text-center py-3"
        >
          <span class="spinner-border spinner-border-sm" />
        </div>

        <p
          v-else-if="!findings.length"
          class="text-muted small fst-italic mb-0"
        >
          {{ $t('anomaly.noFindings') }}
        </p>

        <table
          v-else
          class="table table-sm align-middle mb-0"
        >
          <thead>
            <tr>
              <th>{{ $t('anomaly.module') }}</th>
              <th>{{ $t('anomaly.field') }}</th>
              <th>{{ $t('anomaly.severity') }}</th>
              <th>{{ $t('anomaly.status') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="f in findings"
              :key="f.findingID"
            >
              <td>{{ moduleName(f.moduleID) }}</td>
              <td>{{ f.explanation?.field || '-' }}</td>
              <td>
                <span
                  class="badge"
                  :class="severityBadgeClass(f.severity)"
                >{{ $t(`anomaly.severity${capitalize(f.severity)}`) }}</span>
              </td>
              <td>{{ $t(`anomaly.status${capitalize(f.status)}`) }}</td>
              <td class="text-end">
                <button
                  v-if="f.status !== 'resolved'"
                  class="btn btn-outline-success btn-sm"
                  @click="setStatus(f, 'resolved')"
                >
                  {{ $t('anomaly.resolve') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <div class="text-end mt-2">
          <router-link
            :to="anomalyCenterRoute"
            class="small"
          >
            {{ $t('anomaly.openCenter') }}
          </router-link>
        </div>
      </div>

      <p
        v-if="!options.showSummary && !options.showList"
        class="text-muted small fst-italic mb-0"
      >
        {{ $t('anomaly.nothingToShow') }}
      </p>
    </div>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, inject, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { usePageBlockBase } from './usePageBlockBase'
import Wrap from './Wrap/index.js'

const { t: $t } = useI18n({ useScope: 'global' })
const route = useRoute()

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, required: false, default: '' },
  editable: { type: Boolean, required: false, default: false },
  resizing: { type: Boolean, required: false, default: false },
  magnified: { type: Boolean, required: false, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, required: false, default: false },
  errors: { type: Object, required: false, default: () => ({}) },
})

const { options, refreshBlock } = usePageBlockBase(props, {})

const $AnomalyAPI = inject('$AnomalyAPI')
const $ComposeAPI = inject('$ComposeAPI')

const loading = ref(false)
const findings = ref([])
const modules = ref([])

const anomalyCenterRoute = computed(() => ({
  name: route.name?.startsWith('admin.') ? 'admin.anomaly' : 'namespace.anomaly',
  params: { slug: props.namespace?.slug || props.namespace?.namespaceID },
}))

const summaryStats = computed(() => [
  { key: 'high', label: $t('anomaly.severityHigh'), class: 'text-danger', count: findings.value.filter((f) => f.severity === 'high').length },
  { key: 'medium', label: $t('anomaly.severityMedium'), class: 'text-warning', count: findings.value.filter((f) => f.severity === 'medium').length },
  { key: 'low', label: $t('anomaly.severityLow'), class: 'text-secondary', count: findings.value.filter((f) => f.severity === 'low').length },
  { key: 'new', label: $t('anomaly.statusNew'), class: 'text-primary', count: findings.value.filter((f) => f.status === 'new').length },
])

onMounted(() => {
  fetchModules()
  refresh()
})

refreshBlock(refresh)

function refresh () {
  const namespaceID = props.namespace?.namespaceID
  if (!namespaceID || !$AnomalyAPI) return

  loading.value = true
  $AnomalyAPI.findingSearch({
    namespaceID,
    moduleID: options.value.moduleID && options.value.moduleID !== '0' ? options.value.moduleID : undefined,
    severity: options.value.severity || undefined,
    status: options.value.status || undefined,
    limit: options.value.limit || 10,
    sort: 'createdAt DESC',
  })
    .then(({ set = [] }) => { findings.value = set })
    .catch(() => { findings.value = [] })
    .finally(() => { loading.value = false })
}

function fetchModules () {
  const namespaceID = props.namespace?.namespaceID
  if (!namespaceID || !$ComposeAPI) return
  $ComposeAPI.moduleList({ namespaceID, limit: 500 })
    .then(({ set = [] }) => { modules.value = set })
    .catch(() => { modules.value = [] })
}

function setStatus (finding, status) {
  $AnomalyAPI.findingUpdateStatus({ namespaceID: props.namespace?.namespaceID, findingID: finding.findingID, status })
    .then(() => refresh())
    .catch((err) => { console.error('failed to update finding status', err) })
}

function moduleName (moduleID) {
  const m = modules.value.find((mod) => String(mod.moduleID) === String(moduleID))
  return m ? (m.name || m.handle) : moduleID
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
.anomaly-metric-value {
  font-size: 1.5rem;
  font-weight: 600;
  line-height: 1.2;
}

.anomaly-metric-label {
  color: var(--bs-secondary-color, #6c757d);
  font-size: 0.75rem;
  font-weight: 500;
}
</style>
