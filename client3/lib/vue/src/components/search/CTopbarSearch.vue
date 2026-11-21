<template>
  <div>
    <div
      ref="modalRef"
      class="modal fade"
      tabindex="-1"
    >
      <div class="modal-dialog modal-lg search-modal-dialog">
        <div class="modal-content overflow-hidden border-0 shadow-lg">
          <div
            class="search-container d-flex flex-column"
            style="max-height: 70vh;"
          >
            <div class="border-bottom">
              <c-input-search
                ref="searchInput"
                v-model="query"
                :placeholder="labels.placeholder"
                :loading="loading"
                submittable
                class="topbar-search-input"
                @search="submitSearch"
              />
            </div>

            <template v-if="query.length < 2 && recentSearches.length > 0">
              <div class="flex-grow-1 overflow-auto">
                <div class="px-3 py-2 bg-extra-light border-bottom d-flex align-items-center justify-content-between text-muted">
                  <span class="small font-weight-bold text-uppercase">{{ labels.recentSearches }}</span>
                  <button
                    type="button"
                    class="btn btn-outline-light px-1 py-0 text-muted small border-0 shadow-none"
                    @click="clearRecentSearches"
                  >
                    {{ labels.clearHistory }}
                  </button>
                </div>
                <div
                  v-for="(s, index) in recentSearches"
                  :key="index"
                  class="recent-search-item d-flex align-items-center justify-content-between px-3 py-2 cursor-pointer"
                  @click="useRecentSearch(s)"
                >
                  <div class="d-flex align-items-center">
                    <font-awesome-icon
                      :icon="['fas', 'history']"
                      class="text-muted mr-3 small"
                    />
                    <span class="text-dark">{{ s }}</span>
                  </div>
                  <button
                    type="button"
                    class="btn btn-outline-extra-light remove-btn px-2 text-muted small border-0 shadow-none"
                    @click.stop="removeRecentSearch(index)"
                  >
                    <font-awesome-icon :icon="['fas', 'times']" />
                  </button>
                </div>
              </div>
            </template>

            <template v-if="query.length >= 2 && (hasResults || hasSearched || !loading)">
              <div
                v-if="!hasResults && !loading && hasSearched"
                class="flex-grow-1 overflow-auto p-5 text-center text-muted"
              >
                <p>{{ labels.noResults() }}</p>
              </div>

              <div
                v-else-if="hasResults"
                ref="resultsList"
                class="search-results-list d-flex flex-column flex-grow-1 overflow-auto"
              >
                <div
                  v-for="ns in sortedGroups"
                  :key="ns.id"
                  class="border-bottom"
                >
                  <item-group
                    :title="ns.name"
                    :items="ns.items"
                    :collapse-id="`collapse-${ns.id}`"
                    :expanded="ns.expanded"
                    :labels="labels"
                    @update:expanded="(val: boolean) => expandedGroups[ns.id] = val"
                  >
                    <item-group
                      v-for="mod in ns.sortedModules"
                      :key="mod.id"
                      :title="mod.name"
                      :items="mod.items"
                      :collapse-id="`collapse-${ns.id}-${mod.id}`"
                      :expanded="mod.expanded"
                      :labels="labels"
                      subgroup
                      @update:expanded="(val: boolean) => expandedGroups[`${ns.id}-${mod.id}`] = val"
                    >
                      <record-item
                        v-for="hit in mod.items"
                        :key="hit.id"
                        :hit="hit"
                        :labels="labels"
                        @click="onResultClick(hit)"
                        @open-new-tab="onOpenNewTab(hit)"
                      />
                    </item-group>
                  </item-group>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, getCurrentInstance, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Modal } from 'bootstrap'
import axios from 'axios'
import { library } from '@fortawesome/fontawesome-svg-core'
import { faTimes, faHistory } from '@fortawesome/free-solid-svg-icons'
import RecordItem from './items/RecordItem.vue'
import ItemGroup from './items/ItemGroup.vue'
import { CInputSearch } from '../input'
import { useToast } from '../../composables/useToast'

library.add(faTimes, faHistory)

interface Hit {
  id: string
  type: string
  value: {
    recordID: string
    namespace: { namespaceID: string; slug?: string; name?: string }
    module: { moduleID: string; name?: string }
    matching_fields?: Record<string, any>
    values?: Array<{ name: string; label?: string; value?: string[] }>
  }
}

interface NamespaceGroup {
  id: string
  name: string
  slug: string
  modules: Record<string, ModuleGroup>
  moduleOrder: string[]
  expanded: boolean
  items: Hit[]
  sortedModules: ModuleGroup[]
}

interface ModuleGroup {
  id: string
  name: string
  items: Hit[]
  expanded: boolean
}

const props = withDefaults(defineProps<{
  labels?: Record<string, any>
}>(), {
  labels: () => ({
    numberOfResults: (count: number) => `${count} results`,
  }),
})

const instance = getCurrentInstance()
const $t = instance!.appContext.config.globalProperties.$t
const route = (() => { try { return useRoute() } catch(e) { return {} } })() || {}
const router = (() => { try { return useRouter() } catch(e) { return {} } })() || {}
const { toastDanger, toastErrorHandler } = useToast()

const modalRef = ref<HTMLDivElement | null>(null)
const searchInput = ref<InstanceType<typeof CInputSearch> | null>(null)
const resultsList = ref<HTMLDivElement | null>(null)

let modalInstance: Modal | null = null

const showModal = ref(false)
const query = ref('')
const loading = ref(false)
const results = ref<Hit[]>([])
const hasSearched = ref(false)
const recentSearches = ref<string[]>([])
let cancelRequest: (() => void) | null = null
const expandedGroups = ref<Record<string, boolean>>({})

const currentNamespaceSlug = computed(() => route.params.slug as string || '')

const allResults = computed(() => {
  return results.value
    .map(hit => {
      const highlight = getHitHighlight(hit)
      return { ...hit, highlight }
    })
    .filter(hit => hit.highlight.label)
})

const sortedGroups = computed(() => {
  const currentNs = currentNamespaceSlug.value
  const nsOrder: string[] = []
  const namespaces: Record<string, NamespaceGroup> = {}

  allResults.value.forEach(hit => {
    const ns = hit.value.namespace || {} as any
    const nsID = ns.namespaceID || 'unknown'
    const nsSlug = ns.slug || nsID
    const mod = hit.value.module || {} as any
    const modID = mod.moduleID || 'unknown'

    if (!namespaces[nsID]) {
      namespaces[nsID] = {
        id: nsID,
        name: ns.name || nsSlug,
        slug: nsSlug,
        modules: {},
        moduleOrder: [],
        expanded: expandedGroups.value[nsID] !== false,
        items: [],
        sortedModules: [],
      }
      nsOrder.push(nsID)
    }

    const nsObj = namespaces[nsID]
    if (!nsObj.modules[modID]) {
      nsObj.modules[modID] = {
        id: modID,
        name: mod.name || modID,
        items: [],
        expanded: expandedGroups.value[`${nsID}-${modID}`] !== false,
      }
      nsObj.moduleOrder.push(modID)
    }

    nsObj.modules[modID].items.push(hit)
  })

  return nsOrder
    .map(nsID => {
      const ns = namespaces[nsID]
      ns.sortedModules = ns.moduleOrder.map(modID => ns.modules[modID])
      ns.items = ns.sortedModules.reduce((acc, mod) => acc.concat(mod.items), [] as Hit[])
      return ns
    })
    .sort((a, b) => {
      if (a.slug === currentNs || a.id === currentNs) return -1
      if (b.slug === currentNs || b.id === currentNs) return 1
      return 0
    })
})

const hasResults = computed(() => allResults.value.length > 0)

watch(showModal, (val) => {
  if (val) {
    query.value = ''
    results.value = []
    hasSearched.value = false
  }
})

watch(query, (newVal) => {
  if (newVal.length < 2) {
    loading.value = false
    results.value = []
    hasSearched.value = false
  }
})

function openSearch() {
  showModal.value = true
  nextTick(() => {
    if (modalRef.value) {
      if (!modalInstance) {
        modalInstance = new Modal(modalRef.value)
      }
      modalInstance.show()
    }
  })
}

function onShown() {
  if (searchInput.value) {
    (searchInput.value as any).focus()
  }
}

function closeSearch() {
  showModal.value = false
  modalInstance?.hide()
}

function handleKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    openSearch()
    return
  }
  if (showModal.value && e.key === 'Escape') {
    closeSearch()
  }
}

function loadRecentSearches() {
  const saved = localStorage.getItem('discovery-recent-searches')
  if (saved) {
    try {
      recentSearches.value = JSON.parse(saved)
    } catch {
      recentSearches.value = []
    }
  }
}

function addToRecent(q: string) {
  if (!q || q.length < 2) return
  const list = [q, ...recentSearches.value.filter(s => s !== q)].slice(0, 5)
  recentSearches.value = list
  localStorage.setItem('discovery-recent-searches', JSON.stringify(list))
}

function clearRecentSearches() {
  recentSearches.value = []
  localStorage.removeItem('discovery-recent-searches')
}

function removeRecentSearch(index: number) {
  recentSearches.value.splice(index, 1)
  localStorage.setItem('discovery-recent-searches', JSON.stringify(recentSearches.value))
}

function useRecentSearch(s: string) {
  query.value = s
  submitSearch()
}

function submitSearch() {
  if (query.value.length >= 2) {
    addToRecent(query.value)
    loading.value = true
    performSearch(query.value)
  }
}

async function performSearch(searchQuery: string) {
  if (searchQuery.length < 2) return

  if (cancelRequest) {
    cancelRequest()
  }

  loading.value = true
  const discoveryApi = (instance!.appContext.config.globalProperties as Record<string, any>)['$DiscoveryAPI']
  const { response, cancel } = discoveryApi.queryCancellable({
    query: searchQuery,
    resourceTypes: ['compose:record'],
    size: 20,
  })
  cancelRequest = cancel

  try {
    const { hits = [] } = await response()
    results.value = hits || []
  } catch (e) {
    if (axios.isCancel(e)) return
    console.error('Search failed', e)
    results.value = []
  } finally {
    loading.value = false
    hasSearched.value = true
  }
}

function getHitHighlight(hit: Hit) {
  const matchingFields = hit.value.matching_fields || {}
  const fieldName = Object.keys(matchingFields)[0]
  const values = hit.value.values || []
  const field = fieldName ? values.find(v => v.name === fieldName) : values[0]
  return {
    label: field ? (field.label || field.name) : '',
    value: field && field.value ? field.value[0] : '',
  }
}

async function resolveRecordRoute(hit: Hit) {
  const { recordID, module, namespace } = hit.value
  const { namespaceID } = namespace
  const { moduleID } = module

  const composeApi = (instance!.appContext.config.globalProperties as Record<string, any>)['$ComposeAPI']
  const ns = await composeApi.namespaceRead({ namespaceID })
  if (!ns) {
    toastDanger(props.labels.notFoundNamespace)
    return null
  }

  const slug = ns.slug || ns.namespaceID
  const pages = await composeApi.pageList({ namespaceID, moduleID })
  const recordPage = (pages.set || []).find((p: Record<string, any>) => p.moduleID === moduleID)

  if (!recordPage) {
    toastDanger(props.labels.notFoundPage)
    return null
  }

  return {
    name: 'page.record',
    params: { slug, pageID: recordPage.pageID, recordID },
  }
}

async function onResultClick(hit: Hit) {
  if (hit.type !== 'compose:record') return
  closeSearch()
  addToRecent(query.value)

  try {
    const routeObj = await resolveRecordRoute(hit)
    if (!routeObj) return

    const isPagesRoute = route.name && ['pages', 'page', 'page.record', 'page.record.edit', 'page.record.create'].includes(route.name as string)
    const sameNamespace = routeObj.params.slug === route.params.slug

    if (isPagesRoute && sameNamespace) {
      window.dispatchEvent(new CustomEvent('show-record-modal', {
        detail: { recordID: routeObj.params.recordID, recordPageID: routeObj.params.pageID },
      }))
    } else {
      router.push(routeObj)
    }
  } catch (e) {
    toastErrorHandler(props.labels.recordRedirectError)(e as Error)
  }
}

async function onOpenNewTab(hit: Hit) {
  if (hit.type !== 'compose:record') return
  addToRecent(query.value)

  try {
    const routeObj = await resolveRecordRoute(hit)
    if (!routeObj) return

    const url = router.resolve(routeObj).href
    window.open(url, '_blank')
  } catch (e) {
    toastErrorHandler(props.labels.recordRedirectError)(e as Error)
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
  window.addEventListener('search:open', openSearch)
  loadRecentSearches()

  if (modalRef.value) {
    modalRef.value.addEventListener('shown.bs.modal', onShown)
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown)
  window.removeEventListener('search:open', openSearch)

  if (modalRef.value) {
    modalRef.value.removeEventListener('shown.bs.modal', onShown)
  }
})
</script>

<style lang="scss" scoped>
.text-truncate {
  min-width: 0;
}

.cursor-pointer {
  cursor: pointer;
}

.recent-search-item {
  transition: background-color 0.2s ease;

  .remove-btn {
    opacity: 0;
    transition: opacity 0.2s ease, color 0.2s ease;
  }

  &:hover {
    background-color: var(--light);

    .remove-btn {
      opacity: 1;
    }
  }
}
</style>

<style lang="scss">
.search-modal-dialog {
  margin-top: calc(var(--topbar-height) + 1rem) !important;
  max-width: 700px;
}

.topbar-search-input {
  input {
    border: none !important;
    background: transparent !important;
  }

  .search-button {
    top: 0 !important;
    bottom: 0 !important;
    right: 0 !important;
    border-top: none !important;
    border-right: none !important;
    border-bottom: none !important;
  }
}
</style>
