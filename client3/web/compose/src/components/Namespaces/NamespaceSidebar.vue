<template>
  <div>
    <portal to="sidebar-header-expanded">
      <div
        v-if="!hideNamespaceList"
        class="d-flex w-100 mt-2"
        style="gap: 0.25rem;"
      >
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
          class="flex-grow-1"
          style="min-width: 0"
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

        <button
          v-if="canManageNamespaces"
          :title="$t('editNamespace')"
          data-test-id="button-namespace-edit"
          :disabled="!canUpdateNamespace"
          class="btn btn-outline-primary d-flex align-items-center flex-shrink-0"
          style="height: 38px;"
          @click="$router.push({ name: 'namespace.edit', params: { namespaceID: namespaceID } })"
        >
          <font-awesome-icon :icon="['far', 'edit']" />
        </button>
      </div>
    </portal>

    <portal to="sidebar-body-expanded">
      <div
        v-if="namespace"
        class="d-flex flex-column flex-grow-1"
      >
        <div class="sticky-top w-100 py-2">
          <button
            v-if="isAdminPage"
            data-test-id="button-public"
            class="btn btn-outline-secondary w-100 mb-2"
            @click="$router.push({ name: 'pages', params: { slug: namespace.slug || namespace.namespaceID } })"
          >
            {{ $t('publicPages') }}
          </button>

          <button
            v-else-if="namespace.canManageNamespace"
            data-test-id="button-admin"
            class="btn btn-outline-secondary w-100 mb-2"
            @click="$router.push({ name: 'admin.modules', params: { slug: namespace.slug || namespace.namespaceID } })"
          >
            {{ $t('adminPanel') }}
          </button>

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
            default-route-name="page"
            class="overflow-auto h-100"
          />

          <div
            v-if="!navItems.length"
            class="d-flex justify-content-center mt-5"
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
    </portal>
  </div>
</template>

<script setup>
import { ref, computed, watch, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router'
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
const $Settings = inject('$Settings')
const $ComposeAPI = window.__composeAPI

const namespace = ref(undefined)
const query = ref('')

const moduleLoading = computed(() => store.module.loading)
const chartLoading = computed(() => store.chart.loading)
const pageLoading = computed(() => store.page.loading)
const modules = computed(() => store.module.set)
const pages = computed(() => store.page.set)
const charts = computed(() => store.chart.set)
const can = computed(() => store.rbac.can)

const currentNamespaceID = computed(() => namespace.value ? namespace.value.namespaceID : NoID)
const loading = computed(() => moduleLoading.value || chartLoading.value || pageLoading.value)
const hideNamespaceList = computed(() => {
  const { hideNamespaceList: h } = $Settings.get('compose.ui.sidebar', {})
  return h
})
const canManageNamespaces = computed(() => {
  if (can.value('compose/', 'namespace.create') || can.value('compose/', 'grant')) return true
  return props.namespaces.reduce((acc, ns) => acc || ns.canUpdateNamespace || ns.canDeleteNamespace, false)
})
const showNamespaceListLink = computed(() => {
  const { hideNamespaceListLink: h } = $Settings.get('compose.ui.sidebar', {})
  return !h && canManageNamespaces.value
})
const isAdminPage = computed(() => route.name.includes('admin.'))
const publicRoutes = computed(() => pages.value.filter(({ moduleID: mid, visible }) => visible && mid === NoID))
const filteredPages = computed(() => {
  if (namespace.value) {
    const p = [...(isAdminPage.value ? adminRoutes() : publicPageWrap(publicRoutes.value))]
    if (!query.value) return p
    return p.filter(({ page: pg }) => !['pages', 'modules', 'charts'].includes(pg.pageID) && filter.Assert(pg, query.value, 'title'))
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

function namespaceSelected ({ namespaceID: nid, canManageNamespace, slug = '' }) {
  let { name, params } = route
  if (name.includes('admin.modules')) name = 'admin.modules'
  else if (name.includes('admin.pages')) name = 'admin.pages'
  else if (name.includes('admin.charts')) name = 'admin.charts'
  name = !params.pageID && canManageNamespace && !name.includes('namespace.') ? name : 'pages'
  router.push({ name, params: { slug: slug || nid } })
}

function pageIndex (wraps) {
  const ix = {}
  for (const w of wraps) ix[w.page.pageID] = w
  return ix
}

function moduleWrap (module, pageName) {
  return {
    page: { name: pageName, pageID: `module-${module.moduleID}`, selfID: 'modules', rootSelfID: 'modules', title: module.name || module.handle, visible: true },
    children: [],
    params: { moduleID: module.moduleID },
  }
}

function chartWrap (chart) {
  return {
    page: { name: 'admin.charts.edit', pageID: `chart-${chart.chartID}`, selfID: 'charts', rootSelfID: 'charts', title: chart.name || chart.handle, visible: true },
    children: [],
    params: { chartID: chart.chartID },
  }
}

function adminRoutes () {
  const routeName = route.name
  const pageName = routeName.startsWith('admin.modules.record') ? 'admin.modules.record.list' : 'admin.modules.edit'
  return [
    { page: { pageID: 'modules', selfID: NoID, name: 'admin.modules', title: 'Modules', visible: true }, children: [] },
    ...modules.value.map((m) => moduleWrap(m, pageName)),
    { page: { pageID: 'pages', selfID: NoID, name: 'admin.pages', title: 'Pages', visible: true }, children: [] },
    ...adminPageWrap(pages.value),
    { page: { pageID: 'charts', selfID: NoID, name: 'admin.charts', title: 'Charts', visible: true }, children: [] },
    ...charts.value.map(chartWrap),
  ]
}

function publicPageWrap (pages) {
  return pages.map(({ pageID, selfID, title, visible, config }) => {
    const { navItem = {} } = config
    const { icon = {}, expanded = '' } = navItem
    const { type = '', src = '' } = icon
    const iconType = 'attachment'
    let iconSrc = src
    if (type === iconType) iconSrc = `${$ComposeAPI.baseURL}${src}`
    return { page: { pageID, selfID, title, visible, expanded, icon: iconSrc }, children: [], params: { pageID } }
  })
}

function adminPageWrap (pages) {
  return pages.map(({ pageID, selfID, title, handle, config }) => {
    const { navItem = {} } = config
    const { icon = {} } = navItem
    const { type = '', src = '' } = icon
    const iconType = 'attachment'
    let iconSrc = src
    if (type === iconType) iconSrc = `${$ComposeAPI.baseURL}${src}`
    const pageName = route.name === 'admin.pages.edit' ? 'admin.pages.edit' : 'admin.pages.builder'
    return { page: { name: pageName, pageID: `page-${pageID}`, selfID: selfID !== NoID ? `page-${selfID}` : 'pages', rootSelfID: 'pages', title: title || handle, visible: true, icon: iconSrc }, children: [], params: { pageID } }
  })
}

function getNamespaceLabel (value) {
  if (typeof value === 'string') value = filteredNamespaces.value.find(({ namespaceID: nid }) => nid === value) || {}
  return value.name
}
</script>
