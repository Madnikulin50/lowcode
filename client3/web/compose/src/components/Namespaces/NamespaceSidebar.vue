<template>
  <div>
    <portal to="sidebar-header-expanded">
      <div
        v-if="!hideNamespaceList"
        class="ns-sidebar-header mt-2"
      >
        <div class="d-flex align-items-start gap-2">
          <div class="flex-grow-1 min-w-0">
            <div class="ns-name text-truncate" :title="namespace?.name">
              {{ namespace?.name || $t('sidebar.pickNamespace') }}
            </div>
          </div>
          <button
            v-if="canManageNamespaces"
            :title="$t('editNamespace')"
            data-test-id="button-namespace-edit"
            :disabled="!canUpdateNamespace"
            class="btn btn-sm btn-outline-extra-light text-secondary border-0 ns-edit-btn flex-shrink-0"
            @click="$router.push({ name: 'namespace.edit', params: { namespaceID: namespaceID } })"
          >
            <font-awesome-icon :icon="['far', 'edit']" />
          </button>
        </div>

        <c-input-select
          data-test-id="select-namespace"
          :value="currentNamespaceID"
          :options="filteredNamespaces"
          :get-option-label="getNamespaceLabel"
          :selectable="option => option.namespaceID !== namespace.namespaceID"
          :placeholder="$t('sidebar.pickNamespace')"
          :clearable="false"
          :autoscroll="false"
          :append-to-body="false"
          class="ns-switcher mt-2"
          @input="namespaceSelected"
        >
          <template #list-header>
            <li
              v-if="showNamespaceListLink"
              class="border-bottom text-center mb-1"
            >
              <router-link
                :to="{ name: 'namespace.manage' }"
                data-test-id="button-manage-namespaces"
                class="d-block my-1 fw-bold text-decoration-none"
              >
                {{ $t('manageNamespaces') }}
              </router-link>
            </li>
          </template>
        </c-input-select>
      </div>
      <div
        v-else-if="namespace"
        class="ns-sidebar-header mt-2"
      >
        <div class="ns-name text-truncate">{{ namespace.name }}</div>
      </div>
    </portal>

    <portal to="sidebar-body-expanded">
      <div
        v-if="namespace"
        class="ns-sidebar-body"
        :class="`sidebar-density-${sidebarDensity}`"
        style="flex: 1 1 0%; min-height: 0; position: relative;"
      >
        <div class="sidebar-scroll overflow-auto position-absolute top-0 start-0 w-100 h-100 pe-2">
          <div class="sticky-top w-100 py-2 bg-white ns-sticky">
            <div class="ns-mode-switch btn-group w-100 mb-2" role="group">
              <button
                v-if="isAdminPage"
                data-test-id="button-public"
                type="button"
                class="btn btn-sm btn-outline-secondary flex-fill"
                @click="$router.push({ name: 'pages', params: { slug: namespace.slug || namespace.namespaceID } })"
              >
                {{ $t('publicPages') }}
              </button>
              <button
                v-else-if="namespace.canManageNamespace"
                data-test-id="button-admin"
                type="button"
                class="btn btn-sm btn-outline-secondary flex-fill"
                @click="$router.push({ name: 'admin.modules', params: { slug: namespace.slug || namespace.namespaceID } })"
              >
                {{ $t('adminPanel') }}
              </button>
            </div>

            <c-input-search
              v-model.trim="query"
              :disabled="loading"
              :placeholder="$t(`searchPlaceholder.${isAdminPage ? 'admin' : 'public'}`)"
              :autocomplete="'off'"
            />
          </div>

          <div v-if="!loading">
            <c-sidebar-nav-items
              :items="navItems"
              :start-expanded="!!query"
              :density="sidebarDensity"
              default-route-name="page"
            />

            <div
              v-if="!navItems.length"
              class="ns-empty text-muted text-center mt-4 px-2"
            >
              {{ $t('sidebar.noResults', 'No results') }}
            </div>
          </div>

          <div
            v-else
            class="d-flex align-items-center justify-content-center mt-5"
          >
            <span class="spinner-border spinner-border-sm" />
          </div>
        </div>
      </div>
    </portal>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'sidebar' } })
import { ref, computed, watch, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'
import { components, filter } from 'corteza-lib/vue/dist'
import { Portal } from 'portal-vue'
import { useStore } from '../../store'
const { CSidebarNavItems, CInputSearch } = components

const props = defineProps({
  namespaces: { type: Array, required: true, default: () => [] },
})

const route = useRoute()
const router = useRouter()
const store = useStore()
const { t } = useI18n()
const $Settings = inject('$Settings')
const $ComposeAPI = window.__composeAPI
const $AutomationAPI = window.__automationAPI

const namespace = ref(undefined)
const query = ref('')

const moduleLoading = computed(() => store.module.loading)
const chartLoading = computed(() => store.chart.loading)
const pageLoading = computed(() => store.page.loading)
const modules = computed(() => store.module.set)
const pages = computed(() => store.page.set)
const charts = computed(() => store.chart.set)
const can = computed(() => store.rbac.can)

const ruleChains = ref([])
const ruleChainsLoading = ref(false)

const workflows = ref([])
const workflowsLoading = ref(false)

const sidebarSettings = computed(() => $Settings.get('compose.ui.sidebar', {}) || {})
const sidebarDensity = computed(() => sidebarSettings.value.density === 'compact' ? 'compact' : 'comfortable')

const currentNamespaceID = computed(() => namespace.value ? namespace.value.namespaceID : NoID)
const loading = computed(() => moduleLoading.value || chartLoading.value || pageLoading.value || ruleChainsLoading.value || workflowsLoading.value)
const hideNamespaceList = computed(() => {
  const { hideNamespaceList: h } = sidebarSettings.value
  return h
})
const canManageNamespaces = computed(() => {
  if (can.value('compose/', 'namespace.create') || can.value('compose/', 'grant')) return true
  return props.namespaces.reduce((acc, ns) => acc || ns.canUpdateNamespace || ns.canDeleteNamespace, false)
})
const showNamespaceListLink = computed(() => {
  const { hideNamespaceListLink: h } = sidebarSettings.value
  return !h && canManageNamespaces.value
})
const isAdminPage = computed(() => route.name.includes('admin.'))
const publicRoutes = computed(() => pages.value.filter(({ moduleID: mid, visible }) => visible && mid === NoID))
const filteredPages = computed(() => {
  if (namespace.value) {
    const p = [...(isAdminPage.value ? adminRoutes() : publicPageWrap(publicRoutes.value))]
    if (!query.value) return p
    return p.filter(({ page: pg }) => !['pages', 'modules', 'charts', 'rulechains', 'workflows'].includes(pg.pageID) && filter.Assert(pg, query.value, 'title'))
  }
  return []
})
const filteredNamespaces = computed(() => props.namespaces.filter(({ enabled }) => enabled))
const navItems = computed(() => {
  const current = filteredPages.value
  const ax = pageIndex(isAdminPage.value ? adminRoutes() : publicPageWrap(pages.value))
  for (const cp of current) {
    if (cp.page.selfID && cp.page.selfID !== NoID) {
      if (!ax[cp.page.selfID]) cp.page.selfID = cp.page.rootSelfID
    }
  }
  const cx = pageIndex(current)
  for (let i = current.length - 1; i >= 0; i--) {
    const cp = current[i]
    if (!isAdminPage.value && !cp.page.visible) {
      current.splice(i, 1)
    } else if (cp.page.selfID && cp.page.selfID !== NoID) {
      let p = cx[cp.page.selfID]
      if (!p) {
        if (ax[cp.page.selfID]) {
          current.splice(i, 1, ax[cp.page.selfID])
          p = ax[cp.page.selfID]
          cx[p.page.pageID] = p
          i++
        } else {
          current.splice(i, 0, cp)
          p = cp
          cx[p.page.pageID] = p
        }
      } else {
        current.splice(i, 1)
      }
      if (cp.page.visible) p.children.unshift(cp)
    }
  }
  return current.filter(i => i?.page?.title || i?.page?.name)
})
const canUpdateNamespace = computed(() => namespace.value ? namespace.value.canUpdateNamespace : false)
const namespaceID = computed(() => namespace.value ? namespace.value.namespaceID : NoID)

watch(isAdminPage, () => { query.value = '' })

watch(() => route.params.slug, (slug = '') => {
  query.value = ''
  namespace.value = store.namespace.getByUrlPart(slug)
}, { immediate: true })

watch(() => namespace.value?.namespaceID, (nsID) => {
  if (!nsID) return
  ruleChainsLoading.value = true
  $ComposeAPI.ruleChainList({ limit: 500 })
    .then(({ chains }) => { ruleChains.value = chains || [] })
    .catch(() => { ruleChains.value = [] })
    .finally(() => { ruleChainsLoading.value = false })

  workflowsLoading.value = true
  $AutomationAPI.workflowList({ limit: 500, deleted: 0 })
    .then(({ set }) => { workflows.value = (set || []).map(i => i?.workflow || i) })
    .catch(() => { workflows.value = [] })
    .finally(() => { workflowsLoading.value = false })
}, { immediate: true })

function namespaceSelected ({ namespaceID: nid, canManageNamespace, slug = '' }) {
  let { name, params } = route
  if (name.includes('admin.modules')) name = 'admin.modules'
  else if (name.includes('admin.pages')) name = 'admin.pages'
  else if (name.includes('admin.charts')) name = 'admin.charts'
  else if (name.includes('admin.rulechains')) name = 'admin.rulechains'
  else if (name.includes('admin.workflows')) name = 'admin.workflows'

  name = !params.pageID && canManageNamespace && !name.includes('namespace.') ? name : 'pages'
  router.push({ name, params: { slug: slug || nid } })
}

function pageIndex (wraps) {
  const ix = {}
  for (const w of wraps) ix[w.page.pageID] = w
  return ix
}

function moduleIcon (module) {
  const type = module.config?.type || 'basic'
  if (type === 'datasource') return ['fas', 'cube']
  if (type === 'dbref') return ['fas', 'code-branch']
  return ['fas', 'database']
}

function moduleWrap (module, pageName) {
  return {
    page: { name: pageName, pageID: `module-${module.moduleID}`, selfID: 'modules', rootSelfID: 'modules', title: module.name || module.handle, visible: true, icon: moduleIcon(module) },
    children: [],
    params: { moduleID: module.moduleID },
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

function chartWrap (chart) {
  const icon = chartIcon(chart)
  return {
    page: { name: 'admin.charts.edit', pageID: `chart-${chart.chartID}`, selfID: 'charts', rootSelfID: 'charts', title: chart.name || chart.handle, visible: true, icon },
    children: [],
    params: { chartID: chart.chartID },
  }
}

function ruleChainWrap (chain) {
  return {
    page: {
      name: 'admin.rulechains.edit',
      pageID: `rulechain-${chain.id}`,
      selfID: 'rulechains',
      rootSelfID: 'rulechains',
      title: chain.name || chain.id,
      visible: true,
      icon: ['fas', 'random'],
    },
    children: [],
    params: { chainID: chain.id },
  }
}

function workflowWrap (workflow) {
  return {
    page: {
      name: 'admin.workflows.edit',
      pageID: `workflow-${workflow.workflowID}`,
      selfID: 'workflows',
      rootSelfID: 'workflows',
      title: workflow.meta?.name || workflow.handle || workflow.workflowID,
      visible: true,
      icon: ['fas', 'project-diagram'],
    },
    children: [],
    params: { workflowID: workflow.workflowID },
  }
}

function adminRoutes () {
  const routeName = route.name
  const pageName = routeName.startsWith('admin.modules.record') ? 'admin.modules.record.list' : 'admin.modules.edit'
  return [
    { page: { pageID: 'modules', selfID: NoID, name: 'admin.modules', title: t('navigation.module'), visible: true, section: true }, children: [] },
    ...modules.value.map((m) => moduleWrap(m, pageName)),
    { page: { pageID: 'pages', selfID: NoID, name: 'admin.pages', title: t('navigation.page'), visible: true, section: true }, children: [] },
    ...adminPageWrap(pages.value),
    { page: { pageID: 'charts', selfID: NoID, name: 'admin.charts', title: t('navigation.chart'), visible: true, section: true }, children: [] },
    ...charts.value.map(chartWrap),
    { page: { pageID: 'rulechains', selfID: NoID, name: 'admin.rulechains', title: t('navigation.rulechains'), visible: true, section: true }, children: [] },
    ...ruleChains.value.map(ruleChainWrap),
    { page: { pageID: 'workflows', selfID: NoID, name: 'admin.workflows', title: t('navigation.workflows'), visible: true, section: true }, children: [] },
    ...workflows.value.map(workflowWrap),
  ]
}

function publicPageWrap (pages) {
  return pages.map(({ pageID, selfID, title, visible, config, blocks }) => {
    const { navItem = {} } = config
    const { icon: navIcon = {}, expanded = '' } = navItem
    return { page: { pageID, selfID, title, visible, expanded, icon: resolvePageIcon(navIcon, blocks) }, children: [], params: { pageID } }
  })
}

function adminPageWrap (pages) {
  return pages.map(({ pageID, selfID, title, handle, config, blocks }) => {
    const { navItem = {} } = config
    const { icon: navIcon = {} } = navItem
    const pageName = route.name === 'admin.pages.edit' ? 'admin.pages.edit' : 'admin.pages.builder'
    return { page: { name: pageName, pageID: `page-${pageID}`, selfID: selfID !== NoID ? `page-${selfID}` : 'pages', rootSelfID: 'pages', title: title || handle, visible: true, icon: resolvePageIcon(navIcon, blocks) }, children: [], params: { pageID } }
  })
}

function resolvePageIcon (icon, blocks) {
  if (icon && icon.type && icon.src) {
    if (icon.type === 'fontawesome') {
      const parts = icon.src.split(' ')
      return parts.length >= 2 ? [parts[0], parts.slice(1).join(' ')] : ['fas', icon.src]
    }
    if (icon.type === 'attachment') return `${$ComposeAPI.baseURL}${icon.src}`
    if (icon.type === 'link') return icon.src
  }
  if (blocks && blocks.length) {
    const kinds = new Set(blocks.map(b => b.kind))
    if (kinds.has('Chart') || kinds.has('Metric')) return ['fas', 'chart-bar']
    if (kinds.has('RecordList') || kinds.has('Record')) return ['fas', 'database']
    if (kinds.has('Calendar')) return ['fas', 'calendar-alt']
    if (kinds.has('SocialFeed')) return ['fas', 'rss']
    if (kinds.has('Comment')) return ['fas', 'comments']
    if (kinds.has('Tabs') || kinds.has('Navigation')) return ['fas', 'sitemap']
  }
  return ['fas', 'file-alt']
}

function getNamespaceLabel (value) {
  if (typeof value === 'string') value = filteredNamespaces.value.find(({ namespaceID: nid }) => nid === value) || {}
  return value.name
}
</script>

<style scoped>
.ns-sidebar-header {
  padding: 0 0.15rem 0.25rem;
}

.ns-name {
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.3;
  color: var(--bs-body-color, inherit);
}

.ns-edit-btn {
  margin-top: -0.15rem;
}

.ns-switcher {
  min-width: 0;
}

.ns-sticky {
  z-index: 2;
}

.ns-mode-switch .btn {
  font-size: 0.8125rem;
}

.ns-empty {
  font-size: 0.875rem;
}

.sidebar-density-compact :deep(.c-input-search),
.sidebar-density-compact :deep(.form-control) {
  min-height: 2rem;
  font-size: 0.8125rem;
}
</style>
