<template>
  <div
    data-tets-id="profiler"
    class="container-fluid d-flex flex-column flex-fill pt-2 pb-3"
  >
    <c-content-header :title="$t('system.apigw.title')" />

    <div class="card shadow-sm flex-fill" data-test-id="card-profiler">
      <div class="card-header border-bottom">
        <h4>{{ $t('label.routes') }}</h4>
        <em>{{ description }}</em>
      </div>

      <div class="d-flex align-items-center flex-wrap p-3 gap-1">
        <div class="flex-fill">
          <button
            data-test-id="button-refresh"
            class="btn btn-primary btn-lg"
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
          :processing="processingPurgeRequests"
          :text="$t('purge.all')"
          variant="danger"
          size="lg"
          button-class="flex-fill"
          class="d-flex justify-content-end ms-auto"
          @confirmed="purgeRequests"
        />
      </div>

      <table class="table table-hover mb-0">
        <thead class="table-light">
          <tr>
            <th v-for="f in fields" :key="f.key" :class="f.class">
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
          <tr v-for="item in items" :key="item.routeID">
            <td>{{ item.path }}</td>
            <td class="text-end">{{ item.count }}</td>
            <td class="text-end">{{ item.size_min }}</td>
            <td class="text-end">{{ (item.size_max / 1000).toFixed(3) }} kB</td>
            <td class="text-end">{{ (item.size_avg / 1000).toFixed(3) }} kB</td>
            <td class="text-end">{{ item.time_min.toFixed(2) }} ms</td>
            <td class="text-end">{{ item.time_max.toFixed(2) }} ms</td>
            <td class="text-end">{{ item.time_avg.toFixed(2) }} ms</td>
            <td class="text-end">
              <button
                class="btn btn-link p-0"
                @click="$router.push({ name: 'system.apigw.profiler.route.list', params: { routeID: item.routeID } })"
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
  </div>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useListHelpers } from 'corteza-webapp-admin/src/mixins/listHelpers'

const { t } = useI18n()
const $route = useRoute()
const $router = useRouter()
const $Settings = inject('$Settings', {})
const { pagination, incLoader, decLoader, encodeListParams } = useListHelpers()

const processingPurgeRequests = ref(false)
const totalItems = ref(0)
const items = ref([])
const refresh = ref({ timer: undefined, countdown: 0 })
const filter = ref({ next: '', before: '', query: '', deleted: 0 })
const sorting = ref({ sortBy: 'path', sortDesc: false })

const fields = computed(() => [
  { key: 'path', sortable: true },
  { key: 'count', sortable: true, class: 'text-end' },
  { key: 'size_min', sortable: true, class: 'text-end' },
  { key: 'size_max', sortable: true, class: 'text-end', formatter: v => `${(v / 1000).toFixed(3)} kB` },
  { key: 'size_avg', sortable: true, class: 'text-end', formatter: v => `${(v / 1000).toFixed(3)} kB` },
  { key: 'time_min', sortable: true, class: 'text-end', formatter: v => `${v.toFixed(2)} ms` },
  { key: 'time_max', sortable: true, class: 'text-end', formatter: v => `${v.toFixed(2)} ms` },
  { key: 'time_avg', sortable: true, class: 'text-end', formatter: v => `${v.toFixed(2)} ms` },
  { key: 'actions', label: '', class: 'text-end' },
].map(c => ({
  ...c,
  label: t(`columns.${c.key}`),
})))

const loading = computed(() => !refresh.value.countdown)
const autoRefreshLabel = computed(() => !loading.value ? t('refreshingIn', { seconds: refresh.value.countdown }) : t('label.loading'))
const description = computed(() => $Settings.get('apigw.profiler.global', false) ? t('system.apigw.profiler.description.globalEnabled') : t('system.apigw.profiler.description.globalDisabled'))
const hasNextPage = computed(() => filter.value.next)

watch(() => $route.params, () => {
  resetItems()
}, { immediate: true })

onBeforeUnmount(() => {
  clearRefresh()
})

function loadItems({ append = false } = {}) {
  clearRefresh()

  const oldBeforeID = filter.value.before
  filter.value.before = append ? filter.value.before : ''
  filter.value.routeID = $route.params.routeID
  pagination.limit = append ? 10 : totalItems.value

  const { response } = window.__systemAPI.apigwProfilerAggregationCancellable({ ...filter.value, ...encodeListParams(filter.value, sorting.value, pagination) })
  response().then(({ filter: f = {}, set = [] }) => {
    const { next } = f
    filter.value = { ...filter.value, next }
    items.value = [
      ...(append ? items.value : []),
      ...set.map(i => ({ ...i, routeID: encodeRouteID(i.path) })),
    ]
    totalItems.value = append ? totalItems.value + set.length : totalItems.value

    return { filter: f, set }
  }).finally(() => {
    if (!append) {
      filter.value.before = oldBeforeID
    }
    startRefresh()
  })
}

function purgeRequests() {
  processingPurgeRequests.value = true
  window.__systemAPI.apigwProfilerPurgeAll()
    .then(() => {
      loadItems()
      window.__toastSuccess(t('notification.gateway.profiler.purge.success'))
    })
    .catch(window.__toastError(t('notification.gateway.profiler.purge.error')))
    .finally(() => {
      processingPurgeRequests.value = false
    })
}

function resetItems(s = sorting.value) {
  sorting.value = s
  filter.value.before = ''
  totalItems.value = 10
  loadItems()
}

function encodeRouteID(routeID) {
  return btoa(routeID)
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

</script>
