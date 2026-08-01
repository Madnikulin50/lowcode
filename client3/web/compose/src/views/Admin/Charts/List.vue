<template>
  <div
    v-if="namespace"
    class="container-fluid d-flex flex-column py-3"
  >
    <Teleport to="#topbar-title">
      {{ $t('navigation.chart') }}
    </Teleport>

    <c-resource-list
      :primary-key="primaryKey"
      :filter="filter"
      :sorting="sorting"
      :pagination="pagination"
      :fields="tableFields"
      :items="chartList"
      :translations="{
        searchPlaceholder: $t('chart.searchPlaceholder'),
        notFound: $t('resourceList.notFound', 'Not found'),
        noItems: $t('resourceList.noItems', 'No items'),
        loading: $t('label.loading', 'Loading'),
        showingPagination: $t('resourceList.pagination.showing', 'Showing'),
        singlePluralPagination: $t('resourceList.pagination.single', 'resource'),
        prevPagination: $t('resourceList.pagination.prev', 'Previous'),
        nextPagination: $t('resourceList.pagination.next', 'Next'),
        resourceSingle: $t('label.chart.single', 'Chart'),
        resourcePlural: $t('label.chart.plural', 'Charts'),
      }"
      clickable
      sticky-header
      class="h-100 flex-fill"
      @search="filterList"
      @row-clicked="handleRowClicked"
    >
      <template #header>
        <div
          v-if="namespace.canCreateChart"
          class="dropdown"
        >
          <button
            class="btn btn-primary dropdown-toggle"
            type="button"
            data-bs-toggle="dropdown"
            aria-expanded="false"
          >
            {{ $t('chart.add') }}
          </button>
          <ul class="dropdown-menu m-0">
            <li>
              <router-link
                :to="{ name: 'admin.charts.create', params: { category: 'generic' } }"
                class="dropdown-item"
              >
                {{ $t('chart.addGeneric') }}
              </router-link>
            </li>
            <li>
              <router-link
                :to="{ name: 'admin.charts.create', params: { category: 'funnel' } }"
                class="dropdown-item"
              >
                {{ $t('chart.addFunnel') }}
              </router-link>
            </li>
            <li>
              <router-link
                :to="{ name: 'admin.charts.create', params: { category: 'gauge' } }"
                class="dropdown-item"
              >
                {{ $t('chart.addGauge') }}
              </router-link>
            </li>
            <li>
              <router-link
                :to="{ name: 'admin.charts.create', params: { category: 'radar' } }"
                class="dropdown-item"
              >
                {{ $t('chart.addRadar') }}
              </router-link>
            </li>
          </ul>
        </div>

        <import
          v-if="namespace.canCreateChart"
          :namespace="namespace"
          type="chart"
          @importSuccessful="onImportSuccessful"
        />

        <export
          v-if="namespace.canExportCharts"
          :list="charts"
          type="chart"
        />

        <c-permissions-button
          v-if="namespace.canGrant"
          :resource="`corteza::compose:chart/${namespace.namespaceID}/*`"
          :button-label="$t('label.permissions')"
          class="btn-lg"
        />
      </template>

      <template #actions="{ item: c }">
        <div
          v-if="c.canGrant || c.canDeleteChart"
          class="dropdown"
        >
          <button
            class="btn btn-outline-extra-light dropdown-toggle d-flex align-items-center justify-content-center text-primary border-0 py-2"
            type="button"
            data-bs-toggle="dropdown"
            aria-expanded="false"
          >
            <font-awesome-icon
              :icon="['fas', 'ellipsis-v']"
            />
          </button>
          <ul class="dropdown-menu m-0">
            <li>
              <c-permissions-button
                v-if="c.canGrant"
                :title="c.name || c.handle || c.chartID"
                :target="c.name || c.handle || c.chartID"
                :resource="`corteza::compose:chart/${namespace.namespaceID}/${c.chartID}`"
                :tooltip="$t('permissions.resources.compose.chart.tooltip')"
                :button-label="$t('permissions.ui.label')"
                class="dropdown-item"
              />
            </li>
            <li>
              <c-input-confirm
                v-if="c.canDeleteChart"
                :text="$t('chart.delete')"
                show-icon
                borderless
                variant="link"
                size="md"
                button-class="dropdown-item"
                icon-class="text-danger"
                class="w-100"
                @confirmed="handleDelete(c)"
              />
            </li>
          </ul>
        </div>
      </template>

      <template #name="{ item }">
        <font-awesome-icon
          :icon="chartIcon(item)"
          class="me-2"
          style="height: 1rem; width: 1rem; opacity: 0.65;"
        />
        {{ item.name }}
      </template>

      <template #changedAt="{ item }">
        {{ $locFullDateTime(item.deletedAt || item.updatedAt || item.createdAt) }}
      </template>
    </c-resource-list>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useStore } from '../../../store'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import Import from 'corteza-webapp-compose/src/components/Admin/Import'
import Export from 'corteza-webapp-compose/src/components/Admin/Export'
import listHelpers from 'corteza-webapp-compose/src/mixins/listHelpers'

const { t } = useI18n()
const store = useStore()
const router = useRouter()
const route = useRoute()

const $ComposeAPI = window.__composeAPI

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
})

const primaryKey = 'chartID'

const { procListResults, encodeListParams, pagination, filterList, filter, toastSuccess, toastErrorHandler } = listHelpers.setup({
  store,
  router,
  route,
})

watch(() => props.namespace?.namespaceID, (nsID) => { if (nsID) filter.value.namespaceID = nsID }, { immediate: true })

const sorting = ref({
  sortBy: 'name',
  sortDesc: false,
})

const newChart = ref(new compose.Chart({}))

const charts = computed(() => store.chart.set)

const tableFields = computed(() => [
  {
    key: 'name',
    label: t('chart.columns.name'),
    sortable: true,
    tdClass: 'text-nowrap',
  },
  {
    key: 'handle',
    label: t('chart.columns.handle'),
    sortable: true,
  },
  {
    key: 'changedAt',
    label: t('chart.columns.changedAt'),
    sortable: true,
    class: 'text-end text-nowrap',
  },
  {
    key: 'actions',
    label: '',
    tdClass: 'text-end text-nowrap actions',
  },
])

onMounted(() => {
  document.title = t('label.app-name.chart.list', { label: props.namespace?.name, interpolation: { escapeValue: false } })
})

function chartList () {
  if (!filter.value.namespaceID) return Promise.resolve([])
  return procListResults($ComposeAPI.chartListCancellable(encodeListParams()))
}

function create (subType) {
  let c = new compose.Chart({ ...newChart.value, namespaceID: props.namespace?.namespaceID })
  switch (subType) {
    case 'gauge':
      c = new compose.GaugeChart(c)
      break
    case 'funnel':
      c = new compose.FunnelChart(c)
      break
    case 'radar':
      c = new compose.RadarChart(c)
      break
  }
  store.chart.create(c).then((chart) => {
    router.push({ name: 'admin.charts.edit', params: { chartID: chart.chartID } })
  }).catch(toastErrorHandler(t('notification.chart.createFailed')))
}

function encodeRouteParams () {
  const { query } = filter.value
  const { limit, pageCursor, page } = pagination.value
  return {
    query: { limit, ...sorting.value, query, page, pageCursor },
  }
}

const chartIconMap = {
  pie: ['fas', 'chart-pie'],
  bar: ['fas', 'chart-bar'],
  line: ['fas', 'chart-line'],
  doughnut: ['fas', 'chart-pie'],
  funnel: ['fas', 'filter'],
  gauge: ['fas', 'gauge'],
  radar: ['fas', 'compass'],
  scatter: ['fas', 'chart-line'],
}

function chartIcon (chart) {
  const type = chart.config?.reports?.[0]?.metrics?.[0]?.type
  if (type && chartIconMap[type]) return chartIconMap[type]
  return chartIconMap.bar
}

function handleRowClicked ({ chartID, canUpdateChart, canDeleteChart }) {
  if (!(canUpdateChart || canDeleteChart)) return
  router.push({ name: 'admin.charts.edit', params: { chartID }, query: null })
}

function onImportSuccessful () {
  filterList()
  toastSuccess(t('notification.general.import.successful'))
}

function handleDelete (chart) {
  store.chart.delete(chart).then(() => {
    toastSuccess(t('notification.chart.deleted'))
    filterList()
  }).catch(toastErrorHandler(t('notification.chart.deleteFailed')))
}
</script>

<style lang="scss">
$input-height: 42px;

.chart-name-input {
  height: $input-height;
}
</style>
