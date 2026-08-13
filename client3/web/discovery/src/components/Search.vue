<template>
  <div class="container-fluid h-100 mh-100 p-0 d-flex flex-column">
    <div
      class="d-flex h-100 overflow-hidden"
      :style="{ gap: map.show ? '12px' : '0' }"
    >
      <div
        :style="{ width: map.show ? '70%' : '100%', minWidth: '300px', flexShrink: 0, transition: 'width 0.2s' }"
        class="d-flex flex-column overflow-hidden"
      >
        <div class="px-3 flex-shrink-0">
          <div class="mb-0">
            <c-input-search
              :value="query"
              :placeholder="t('input-placeholder')"
              :autocomplete="'off'"
              :disabled="store.processing"
              submittable
              @search="onQuerySubmit"
            />
          </div>

          <div class="d-flex align-items-center px-1 mt-1 mb-2 text-muted">
            <div class="d-flex align-items-center">
              <span
                :class="{ 'discovering': store.processing }"
                class="mt-1"
              >
                {{ searchDescription }}
              </span>
            </div>

            <div class="d-flex align-items-center ms-auto gap-2">
              <button
                class="btn btn-extra-light btn-sm d-flex align-items-center gap-1 mt-2"
                @click="toggleMap"
              >
                <font-awesome-icon
                  :icon="['fas', 'map-marked-alt']"
                />
                {{ !map.show ? t('search.show-map') : t('search.hide-map') }}
              </button>

              <div class="d-flex align-items-center">
                <font-awesome-icon
                  :icon="['fas', 'grip-lines']"
                  class="mt-2 me-1 pointer"
                  :class="{ 'text-primary': viewMode === 'list' }"
                  @click="viewMode = 'list'"
                />

                <div class="form-check form-switch pointer ms-2 mt-2">
                  <input
                    v-model="viewMode"
                    :value="'grid'"
                    :false-value="'list'"
                    type="checkbox"
                    class="form-check-input-v3"
                    role="switch"
                  />
                </div>

                <font-awesome-icon
                  :icon="['fas', 'grip-horizontal']"
                  class="mt-2 ms-1 pointer"
                  :class="{ 'text-primary': viewMode === 'grid' }"
                  @click="viewMode = 'grid'"
                />
              </div>
            </div>
          </div>
        </div>

        <div class="d-flex flex-column flex-fill overflow-hidden">
          <div
            v-if="(store.processing || !total.actual) && !loadingMore"
            class="d-flex align-items-center justify-content-center w-100 my-5"
          >
            <h5 class="mb-0">
              <div
                v-if="store.processing"
                class="spinner-border text-primary p-4"
                role="status"
              />
              <span
                v-else-if="!total.actual"
              >
                {{ t('no-results') }}
              </span>
            </h5>
          </div>

          <div
            v-else
            class="results d-flex flex-wrap gap-3 px-4 py-3 overflow-auto"
            :class="{ 'list-view': viewMode === 'list' }"
          >
            <div
              v-for="(hit, i) in hits"
              :key="i"
              class="result-item w-100"
              :class="{ 'grid-view': viewMode === 'grid' }"
            >
              <result
                :id="`result-${i}`"
                :index="i"
                :hit="hit"
                :show-map="map.show"
                :class="{ 'border-primary border shadow': map.clickedMarker && map.clickedMarker === i }"
                class="border"
                @hover="onResultHover(i)"
              />
            </div>

            <div
              v-if="total.actual > 0 && total.actual < total.all"
              class="d-flex align-items-center justify-content-center py-3 w-100"
            >
              <button
                class="btn btn-primary d-flex align-items-center justify-content-center gap-1"
                :disabled="loadingMore"
                @click="getSearchData({ append: true })"
              >
                <div
                  v-if="loadingMore"
                  class="spinner-border spinner-border-sm"
                  role="status"
                />
                {{ loadingMore ? t('search.loading-more') : t('search.show-more') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="map.show"
        :style="{ width: '30%', minWidth: '300px', flexShrink: 0, transition: 'width 0.2s' }"
        class="overflow-hidden"
      >
        <discovery-map
          :markers="map.markers"
          :hover-index="map.hoverIndex"
          class="ps-2"
          @marker-clicked="markerClicked"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'search' } })
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useDiscoveryStore } from '../store'
import { components } from 'corteza-lib/vue/dist'
import Result from './Results'
import DiscoveryMap from './DiscoveryMap.vue'

const { CInputSearch } = components

const { t } = useI18n({
  useScope: 'local',
  messages: {},
})

const store = useDiscoveryStore()
const route = useRoute()
const router = useRouter()

const loadingMore = ref(false)
const query = ref('')
const hits = ref([])
const pagination = ref({ limit: 50, from: 0, size: 50 })
const total = ref({ all: 0, actual: 0 })
const initial = ref(false)
const map = ref({ show: false, markers: [], clickedMarker: undefined, hoverIndex: undefined })
const viewMode = ref('list')

const searchDescription = computed(() => {
  if (store.processing) {
    return t('discovering')
  }

  if (total.value.all > 0) {
    return t('range', { actual: total.value.actual, all: total.value.all })
  }

  return ''
})

watch(() => store.resourceTypes, () => {
  if (initial.value) return
  pagination.value.size = pagination.value.limit
  getSearchData()
})

watch(() => store.modules, () => {
  if (initial.value) return
  pagination.value.size = pagination.value.limit
  getSearchData()
})

watch(() => store.namespaces, () => {
  if (initial.value) return
  pagination.value.size = pagination.value.limit
  getSearchData()
})

onMounted(() => {
  initial.value = true

  const { query: q = '', modules, namespaces, resourceTypes, size = 50 } = route.query

  query.value = q
  pagination.value.size = size

  if (namespaces) {
    store.updateNamespaces(Array.isArray(namespaces) ? namespaces : [namespaces])
  }

  if (modules) {
    store.updateModules(Array.isArray(modules) ? modules : [modules])
  }

  if (resourceTypes) {
    store.updateResourceTypes(Array.isArray(resourceTypes) ? resourceTypes : [resourceTypes])
  }

  getSearchData()

  setTimeout(() => {
    initial.value = false
  }, 1000)
})

function getSearchData({ query: q = query.value, append = false } = {}) {
  if (append) {
    pagination.value.size += pagination.value.limit
    loadingMore.value = true
  } else {
    map.value.markers = []
    hits.value = []
    map.value.hoverIndex = undefined
  }

  const modules = store.modules
  const namespaces = store.namespaces
  const { size } = pagination.value

  updateRouteQuery({ query: q, modules, namespaces, size })

  store.fetchData({ query: q, modules, namespaces, size }).then((response = {}) => {
    if (response) {
      if (append) {
        hits.value = [...hits.value, ...(response.hits || [])]
      } else {
        hits.value = response.hits || []
      }
      total.value.all = response.total_results || 0
      total.value.actual = hits.value.length

      pagination.value = {
        ...pagination.value,
        from: response.from || 0,
        size: response.size || 0,
      }

      getMarkers()
    }
  }).catch(e => {
    toastErrorHandler(t('notification.search.failed'))(e)
    hits.value = []
  }).finally(() => {
    loadingMore.value = false
  })
}

function onQuerySubmit(q) {
  if (!store.processing) {
    query.value = q
    pagination.value.size = 50
    getSearchData()
  }
}

function getMarkers() {
  const markers = []

  hits.value.forEach(({ type, value }) => {
    if (type === 'compose:record' && Array.isArray(value.values)) {
      const id = value.recordID
      value.values.forEach(({ value = [] }) => {
        const isGeometry = value && value.find(v => {
          return (v !== null ? v : '').toString().includes('{"coordinates":[')
        })

        if (isGeometry) {
          value.forEach(coordinates => {
            coordinates = JSON.parse(coordinates || '{}').coordinates
            if (coordinates && coordinates.length) {
              markers.push({ id, coordinates })
            }
          })
        }
      })
    }
  })

  map.value.markers = markers
}

function markerClicked(ID) {
  if (ID) {
    const result = document.getElementById(`result-${ID}`)

    if (result) {
      result.scrollIntoView({
        behavior: 'smooth',
        block: 'center',
      })
    }
  }

  map.value.clickedMarker = ID
  map.value.hoverIndex = undefined
}

function onResultHover(index) {
  if (map.value.show) {
    map.value.hoverIndex = index
  }
}

function toggleMap() {
  map.value.show = !map.value.show
}

function updateRouteQuery({ query: q = undefined, modules = [], namespaces = [], size = 0 }) {
  if (JSON.stringify(route.query) !== JSON.stringify({ query: q, modules, namespaces, size })) {
    router.push({ query: { query: q || undefined, modules, namespaces, size } })
  }
}

function toastErrorHandler(msg) {
  return (e) => {
    console.error(msg, e)
  }
}
</script>

<style lang="scss" scoped>
.discovering::after {
  display: inline-block;
  animation: discovering steps(1, end) 1s infinite;
  content: '';
}

@keyframes discovering {
  0% { content: ''; }
  25% { content: '.'; }
  50% { content: '..'; }
  75% { content: '...'; }
  100% { content: ''; }
}

.result-item {
  &.grid-view {
    min-width: 30rem;
    flex: 1;
  }
}

.list-view {
  .result-item {
    max-width: 100%;
  }
}

.results {
  flex: 1;
  min-height: 0;
  position: relative;
}
</style>
