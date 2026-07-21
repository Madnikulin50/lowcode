<template>
  <div class="card shadow-sm" data-test-id="card-requests">
    <div class="card-header border-bottom">
      <h4 class="mb-0">{{ $t('label.requests') }}</h4>
    </div>

    <div class="d-flex align-items-center flex-wrap flex-fill p-3 gap-1">
      <div class="flex-fill">
        <button
          data-test-id="button-refresh"
          class="btn btn-primary"
          :disabled="loading"
          @click="loadItems()"
        >
          {{ $t('label.refresh') }}
        </button>
        <span
          class="ms-1"
          :class="{ 'loading': loading }"
        >
          {{ autoRefreshLabel }}
        </span>
      </div>

      <c-input-confirm
        :disabled="!items.length"
        :processing="processingPurge"
        :text="$t('purge.this')"
        variant="danger"
        @confirmed="purgeRequests"
      />
    </div>

    <div class="card-body p-0">
      <table class="table table-hover mb-0">
        <thead class="table-light">
          <tr>
            <th v-for="f in fields" :key="f.key" :class="f.class" :style="f.thStyle">
              {{ f.label }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading">
            <td :colspan="fields.length" class="text-center p-4">
              <div class="spinner-border" />
            </td>
          </tr>
          <tr v-for="row in items" :key="row.hitID">
            <td>{{ fields[0].formatter ? fields[0].formatter(row.time_start) : row.time_start }}</td>
            <td>{{ fields[1].formatter ? fields[1].formatter(row.time_finish) : row.time_finish }}</td>
            <td class="text-center">{{ row.request.Method }}</td>
            <td class="text-center">
              <h6 class="mb-0">
                <span :class="'badge bg-' + getStatusCodeVariant(row.http_status_code)">{{ row.http_status_code }}</span>
              </h6>
            </td>
            <td class="text-end">{{ ((row.request.ContentLength || 0) / 1000).toFixed(3) }} kB</td>
            <td class="text-end">{{ row.time_duration ? `${row.time_duration.toFixed(2)} ms` : '' }}</td>
            <td class="text-end">
              <button
                data-test-id="button-edit-route"
                class="btn btn-link p-0"
                @click="$router.push({ name: 'system.apigw.profiler.hit', params: { hitID: row.hitID } })"
              >
                <font-awesome-icon
                  :icon="['fas', 'info-circle']"
                  class="text-primary"
                />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="items.length" class="card-footer border-top d-flex align-items-center justify-content-center">
      <button
        class="btn btn-outline-secondary"
        :disabled="!hasNextPage || loading"
        @click="loadMore()"
      >
        {{ $t('label.loadMore') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { fmt } from 'corteza-lib/js/dist'
import { useListHelpers } from 'corteza-webapp-admin/src/mixins/listHelpers'

const { t } = useI18n()
const $router = useRouter()
const { pagination, incLoader, decLoader, encodeListParams } = useListHelpers()

const props = defineProps({
  route: { type: String, default: '' },
})

const processingPurge = ref(false)
const filter = ref({
  routeID: '',
  next: '',
  before: '',
  query: '',
})
const sorting = ref({ sortBy: 'time_start', sortDesc: true })
const totalItems = ref(0)
const items = ref([])
const refresh = ref({ timer: undefined, countdown: 0 })

const fields = computed(() => [
  { key: 'time_start', sortable: true, formatter: v => v ? fmt.fullDateTime(v) : '' },
  { key: 'time_finish', sortable: true, formatter: v => v ? fmt.fullDateTime(v) : '' },
  { key: 'http_method', sortable: true, class: 'text-center' },
  { key: 'http_status_code', sortable: true, class: 'text-center' },
  { key: 'content_length', sortable: true, formatter: v => `${v || 0} bytes`, class: 'text-end' },
  { key: 'time_duration', sortable: true, formatter: v => v ? `${v.toFixed(2)} ms` : '', class: 'text-end' },
  { key: 'actions', label: '', class: 'text-end' },
].map(c => ({
  ...c,
  label: t(`columns.${c.key}`),
})))

const loading = computed(() => !refresh.value.countdown)
const autoRefreshLabel = computed(() => !loading.value ? t('refreshingIn', { seconds: refresh.value.countdown }) : t('label.loading'))
const hasNextPage = computed(() => filter.value.next)

watch(() => props.route, () => {
  if (props.route) {
    resetItems()
  }
}, { immediate: true })

onBeforeUnmount(() => {
  clearRefresh()
})

function loadItems({ append = false } = {}) {
  clearRefresh()

  const oldBeforeID = filter.value.before
  filter.value.before = append ? filter.value.before : ''
  filter.value.routeID = props.route
  pagination.limit = append ? 10 : totalItems.value

  const { response } = window.__systemAPI.apigwProfilerRouteCancellable(encodeListParams(filter.value, sorting.value, pagination))
  response().then(({ filter: f = {}, set = [] }) => {
    const { next } = f
    filter.value = { ...filter.value, next }
    items.value = [
      ...(append ? items.value : []),
      ...set.map(i => ({ ...i, hitID: i.ID })),
    ]
    totalItems.value = append ? totalItems.value + set.length : totalItems.value

    return { filter: f, set }
  })
    .catch(window.__toastError(t('notification.gateway.profiler.fetch.error')))
    .finally(() => {
      if (!append) {
        filter.value.before = oldBeforeID
      }
      startRefresh()
    })
}

function purgeRequests() {
  processingPurge.value = false

  const { routeID } = filter.value
  window.__systemAPI.apigwProfilerPurge({ routeID })
    .then(() => {
      loadItems()
      window.__toastSuccess(t('notification.gateway.profiler.purge.success'))
    })
    .catch(window.__toastError(t('notification.gateway.profiler.purge.error')))
    .finally(() => {
      processingPurge.value = true
    })
}

function resetItems(s = sorting.value) {
  sorting.value = s
  filter.value.before = ''
  totalItems.value = 10
  loadItems()
}

function loadMore() {
  filter.value.before = filter.value.next
  loadItems({ append: true })
}

function startRefresh() {
  refresh.value.countdown = 10
  resetRefresh()
}

function clearRefresh() {
  refresh.value.timer = clearTimeout(refresh.value.timer)
  refresh.value.countdown = 0
}

function resetRefresh() {
  clearTimeout(refresh.value.timer)
  refresh.value.timer = setTimeout(() => {
    refresh.value.countdown--
    if (refresh.value.countdown) {
      resetRefresh()
    } else {
      loadItems()
    }
  }, 1000)
}

function getStatusCodeVariant(statusCode = '') {
  const codeVariants = {
    2: 'success',
    3: 'info',
    4: 'danger',
    5: 'warning',
  }
  return codeVariants[statusCode[0]]
}
</script>

<style>
</style>
