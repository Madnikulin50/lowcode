<template>
  <div
    v-if="!!page"
    class="d-flex w-100 h-100 overflow-hidden"
  >
    <Teleport
      v-if="!isRecordPage"
      defer
      to="#topbar-title"
    >
      <span class="d-inline-flex align-items-center gap-1">
        {{ pageTitle }}
        <c-help-trigger
          v-if="pageHelp.hasAny"
          v-bind="pageHelp.triggerProps"
          :label="helpOpenLabel"
        />
      </span>
    </Teleport>

    <Teleport
      v-if="!isRecordPage && enableAI"
      defer
      to="#topbar-search"
    >
      <c-input-search
        v-model.trim="aiPrompt"
        ai
        :aria-label="$t('AI')"
        :placeholder="$t('aiChat.startPrompt')"
        :autocomplete="'off'"
        submittable
        @search="handleAiSearch"
        @ai-search="handleAiSearch"
      />
    </Teleport>

    <Teleport
      defer
      to="#topbar-menu-items"
    >
      <div style="display: contents;">
        <a
          v-for="format in ['pdf', 'xlsx', 'pptx']"
          :key="format"
          href="#"
          class="dropdown-item"
          :class="{ disabled: exporting }"
          :data-test-id="`button-page-export-${format}`"
          @click.prevent="exportPage(format)"
        >
          Экспорт в {{ format.toUpperCase() }}
        </a>

        <template v-if="!isRecordPage && page.canUpdatePage">
          <a
            href="#"
            class="dropdown-item"
            data-test-id="button-page-export-schedule"
            @click.prevent="showExportSchedule = true"
          >
            Рассылка по почте
          </a>

          <router-link
            data-test-id="button-page-builder"
            :to="pageBuilder"
            class="dropdown-item"
          >
            {{ $t('label.pageBuilder') }}
          </router-link>

          <router-link
            data-test-id="button-page-edit"
            :to="pageEditor"
            class="dropdown-item"
          >
            {{ $t('navigation.editPage') }}
          </router-link>

          <page-translator
            data-test-id="button-page-translations"
            :page="trPage"
            :page-layout="layout"
            button-variant="link"
            button-class="dropdown-item rounded-0 text-dark text-decoration-none justify-content-start"
          >
            {{ $t('navigation.translation') }}
          </page-translator>
        </template>
      </div>
    </Teleport>

    <div class="flex-grow-1 overflow-auto d-flex flex-column p-2 w-100 h-100">
      <page-help-bar
        v-if="pageHelp.hasApp"
        class="flex-shrink-0"
        :page="page"
        :namespace="namespace"
      />
      <div class="flex-grow-1 overflow-auto d-flex w-100 h-100">
        <router-view
          v-if="isRecordPage"
          :namespace="namespace"
          :module="module"
          :page="page"
        />

        <div
          v-else-if="!layout"
          class="d-flex align-items-center justify-content-center w-100"
        >
          <span class="spinner-border" />
        </div>

        <grid
          v-else-if="blocks"
          :namespace="namespace"
          :module="module"
          :page="page"
          :blocks="blocks"
        />
      </div>
    </div>

    <export-schedule-modal
      v-if="!isRecordPage && page.canUpdatePage"
      v-model:show="showExportSchedule"
      :namespace-id="namespace.namespaceID"
      :page-id="page.pageID"
    />

    <record-modal
      :namespace="namespace"
    />

    <magnification-modal
      :namespace="namespace"
    />
    <ai-chat-modal
      :namespace="namespace.namespaceID"
      :page="page.pageID"
      :module="page.moduleID"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'page' } })
import { ref, computed, watch, onBeforeUnmount, getCurrentInstance } from 'vue'
import { useRouter, useRoute, onBeforeRouteLeave, onBeforeRouteUpdate } from 'vue-router'
import { compose, NoID } from 'corteza-lib/js/dist'
import { components, composables } from 'corteza-lib/vue/dist'
import { usePageStore } from '../../../store/page'
import { useModuleStore } from '../../../store/module'
import { usePageLayoutStore } from '../../../store/page-layout'
import { useUiStore } from '../../../store/ui'
import { useRecordStore } from '../../../store/record'
import { fetchID } from 'corteza-webapp-compose/src/lib/block'
import { normalizeXYWH } from 'corteza-webapp-compose/src/lib/block-layout'
import bus from '../../../lib/bus'
import Grid from 'corteza-webapp-compose/src/components/Public/Page/Grid'
import RecordModal from 'corteza-webapp-compose/src/components/Public/Record/Modal'
import MagnificationModal from 'corteza-webapp-compose/src/components/Public/Page/Block/Modal'
import AiChatModal from 'corteza-webapp-compose/src/components/Public/Page/AiChat/Modal'
import PageTranslator from 'corteza-webapp-compose/src/components/Admin/Page/PageTranslator'
import ExportScheduleModal from 'corteza-webapp-compose/src/components/Public/Page/ExportScheduleModal.vue'
import { useHelp } from '../../../composables/useHelp'
import { pageHelpDocs, hydratePageDocs } from '../../../help/appDocs'
import PageHelpBar from '../../../components/Common/PageHelpBar.vue'

const { CInputSearch } = components
const { toastErrorHandler, toastWarning } = composables.useToast()

const { proxy } = getCurrentInstance()

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  page: { type: compose.Page, required: true },
})

const router = useRouter()
const route = useRoute()
const pageStore = usePageStore()
const moduleStore = useModuleStore()
const pageLayoutStore = usePageLayoutStore()
const uiStore = useUiStore()
const recordStore = useRecordStore()

const $auth = window.__auth
const $SystemAPI = window.__systemAPI

const exporting = ref(false)
const showExportSchedule = ref(false)

// Server renders the page in headless browser with the print rules (see shims/print.css)
async function exportPage (format) {
  if (exporting.value) return
  exporting.value = true

  try {
    const query = new URLSearchParams({ format, url: window.location.href })
    const url = `${window.CortezaAPI}/compose/namespace/${props.namespace.namespaceID}/page/${props.page.pageID}/export?${query}`
    const rsp = await fetch(url, { headers: { Authorization: `Bearer ${$auth.accessTokenFn()}` } })
    if (!rsp.ok) {
      throw new Error(await rsp.text())
    }

    const link = document.createElement('a')
    link.href = URL.createObjectURL(await rsp.blob())
    link.download = `${pageTitle.value || 'page'}.${format}`
    link.click()
    URL.revokeObjectURL(link.href)
  } catch (e) {
    toastErrorHandler('Не удалось выполнить экспорт страницы')(e)
  } finally {
    exporting.value = false
  }
}

const layouts = ref([])
const layout = ref(undefined)
const blocks = ref(undefined)
const tempRecord = ref(undefined)
const pageTitle = ref('')
const aiPrompt = ref('')

const pageHelp = useHelp('compose.page.view', computed(() => pageHelpDocs(props.namespace, props.page)))
watch(
  () => [props.namespace, props.page],
  () => { hydratePageDocs(props.namespace, props.page) },
  { immediate: true },
)
const helpOpenLabel = computed(() => {
  const value = proxy.$t('help.bar.open')
  return !value || value === 'help.bar.open' ? 'Справка' : value
})

const recordID = computed(() => route.params.recordID || '')

const isRecordPage = computed(() => recordID.value || route.name === 'page.record.create')

const module = computed(() => {
  if (props.page.moduleID && props.page.moduleID !== NoID) {
    return moduleStore.getByID(props.page.moduleID)
  }
  return undefined
})

const trPage = computed({
  get: () => props.page.clone(),
  set: (v) => pageStore.updateSet([v]),
})

const pageEditor = computed(() => ({
  name: 'admin.pages.edit',
  params: { pageID: props.page.pageID },
}))

const pageBuilder = computed(() => {
  const { pageLayoutID } = layout.value || {}
  return { name: 'admin.pages.builder', params: { pageID: props.page.pageID }, query: { layoutID: pageLayoutID } }
})

const moduleEditor = computed(() => {
  if (!module.value) return undefined
  return { name: 'admin.modules.edit', params: { moduleID: module.value.moduleID } }
})

const enableAI = computed(() => true)

const uniqueID = computed(() => [props.page.pageID, route.query.layoutID])

onBeforeRouteLeave((to, from, next) => {
  uiStore.setPreviousPages([])
  next()
})

onBeforeRouteUpdate((to, from, next) => {
  const { recordID: toRecordID } = to.params
  const { recordID: fromRecordID } = from.params
  const fromToRecordPage = fromRecordID && toRecordID !== fromRecordID
  const fromNormalToRecordPage = from.name === 'page' && to.name !== 'page'
  if (fromNormalToRecordPage || fromToRecordPage) {
    uiStore.pushPreviousPages(from)
  }
  next()
})

watch(uniqueID, (value, oldValue) => {
  const [pageID = '', pageLayoutID = ''] = (Array.isArray(value) ? value : []).map(v => v || '')
  const [oldPageID = ''] = (Array.isArray(oldValue) ? oldValue : []).map(v => v || '')

  if (!pageID || pageID === NoID) return

    if (pageID !== oldPageID) {
    layouts.value = pageLayoutStore.getByPageID(props.page.pageID) || []
    layout.value = undefined
    pageTitle.value = props.page.title
  }

  if (!isRecordPage.value) {
    determineLayout({ pageLayoutID }).then(b => {
      if (b) blocks.value = b
    }).finally(() => {
      processing.value = false
    })
  }

  if (uiStore.recordPaginationUsable) {
    uiStore.setRecordPaginationUsable(false)
  } else if (pageID !== oldPageID) {
    uiStore.clearRecordPagination()
    recordStore.clearSet()
  }
}, { immediate: true })

watch(() => props.page.handle, (handle, oldHandle) => {
  if (handle !== oldHandle) {
    uiStore.setPageHandle(handle)
  }
}, { immediate: true })

watch(() => layout.value?.handle, (handle, oldHandle) => {
  if (handle !== oldHandle) {
    uiStore.setLayoutHandle(handle)
  }
}, { immediate: true })

onBeforeUnmount(() => {
  uiStore.setPageHandle('')
  uiStore.setLayoutHandle('')
  setDefaultValues()
  recordStore.clearSet()
})

const processing = ref(false)

function userMatchesVisibilityRoles (roles = []) {
  const have = new Set(($auth.user?.roles || []).map(String))
  return (roles || []).some(roleID => have.has(String(roleID)))
}

async function determineLayout({ pageLayoutID, redirectOnFail = true } = {}) {
  if (isRecordPage.value) {
    resetErrors()
  }

  let expressions = {}

  if ((layouts.value || []).some(({ config = {} }) => config.visibility?.expression)) {
    expressions = await evaluateLayoutExpressions()
  }

  const matchedLayout = (layouts.value || []).find(l => {
    if (pageLayoutID && l.pageLayoutID !== pageLayoutID) return false
    const { expression, roles = [] } = l.config?.visibility || {}
    if (expression && !expressions[l.pageLayoutID]) return false
    if (!roles.length) return true
    return userMatchesVisibilityRoles(roles)
  })

  if (!matchedLayout) {
    toastWarning(proxy.$t('notification.page.page-layout.notFound.view'))
    if (redirectOnFail) {
      router.go(-1)
    }
    return
  }

  layout.value = matchedLayout

  if (isRecordPage.value) {
  } else {
    const { handle, meta = {} } = matchedLayout
    pageTitle.value = (meta.title || props.page.title) || handle || proxy.$t('navigation.noPageTitle')
    document.title = pageTitle.value
  }

  return prepareBlocks()
}

async function evaluateLayoutExpressions() {
  const expressions = {}
  const variables = expressionVariables()

  layouts.value.forEach(l => {
    const { config = {} } = l
    if (!config.visibility?.expression) return
    variables.layout = l
    expressions[l.pageLayoutID] = config.visibility.expression
  })

  return $SystemAPI.expressionEvaluate({ variables, expressions }).catch(e => {
    toastErrorHandler(proxy.$t('notification.evaluate.failed'))(e)
    Object.keys(expressions).forEach(key => (expressions[key] = false))
    return expressions
  })
}

function expressionVariables() {
  const record = tempRecord.value
  return {
    user: $auth.user,
    record: record ? record.serialize() : {},
    screen: {
      width: window.innerWidth,
      height: window.innerHeight,
      userAgent: navigator.userAgent,
      breakpoint: getBreakpoint(),
    },
    oldLayout: layout.value,
    layout: undefined,
    ...(isRecordPage.value && {
      isView: false,
      isCreate: false,
      isEdit: false,
    }),
  }
}

function getBreakpoint() {
  const width = window.innerWidth
  if (width < 576) return 'xs'
  if (width >= 576 && width < 768) return 'sm'
  if (width >= 768 && width < 992) return 'md'
  if (width >= 992 && width < 1200) return 'lg'
  if (width >= 1200 && width < 1400) return 'xl'
  return 'xxl'
}

// A block shown only inside a Tabs block is not part of the layout: it is handed to the grid
// hidden (so it is not drawn on its own) and rendered by the tab. A copy, so that the page's
// own block never gets the flag saved.
function hiddenCopy(block) {
  return Object.assign(Object.create(Object.getPrototypeOf(block)), block, { meta: { ...block.meta, hidden: true } })
}

async function prepareBlocks() {
  blocks.value = undefined
  const tempBlocks = []
  const layoutBlocks = layout.value?.blocks || []
  const tabbedIDs = new Set()
  const pageBlocks = props.page?.blocks || []

  layoutBlocks.forEach(({ blockID, xywh }) => {
    const block = pageBlocks.find(b => b.blockID === blockID)
    if (block) {
      block.xywh = normalizeXYWH(xywh)
      tempBlocks.push(block)
      if (block.kind === 'Tabs') {
        const { tabs = [] } = block.options || {}
        ;(tabs || []).forEach(t => {
          if (t.blockID && !layoutBlocks.some(b => b.blockID === t.blockID)) {
            tabbedIDs.add(t.blockID)
          }
        })
      }
    }
  })

  pageBlocks.forEach(block => {
    if (tabbedIDs.has(block.blockID)) {
      tempBlocks.push(hiddenCopy(block))
    }
  })

  return evaluateBlocks(tempBlocks)
}

async function evaluateBlocks(pageBlocks = props.page?.blocks) {
  pageBlocks = pageBlocks || []
  let layoutBlocksExpressions = {}

  if (pageBlocks.some(({ meta = {} }) => (meta.visibility || {}).expression)) {
    layoutBlocksExpressions = await evaluateBlocksExpressions(pageBlocks)
  }

  pageBlocks.forEach(block => {
    const { meta = {} } = block || {}
    const blockID = fetchID(block)
    const visibility = meta.visibility || {}
    const { roles = [] } = visibility
    const validExpression = !visibility.expression || layoutBlocksExpressions[blockID]
    const validRole = !roles.length || userMatchesVisibilityRoles(roles)
    const showBlock = block && validExpression && validRole
    if (!block.meta) block.meta = {}
    block.meta.invisible = !showBlock
  })

  pageBlocks.forEach(block => {
    if (block.kind === 'Tabs' && !block.meta.invisible) {
      const { tabs = [] } = block.options || {}
      const hasVisibleTab = (tabs || []).some(t => {
        const b = pageBlocks.find(b2 => fetchID(b2) === t.blockID)
        return b ? !b.meta.invisible : !!t.title
      })
      if (!hasVisibleTab) {
        block.meta.invisible = true
      }
    }
  })

  return pageBlocks
}

async function evaluateBlocksExpressions(pageBlocks) {
  const expressions = {}
  const variables = expressionVariables()

  pageBlocks.forEach(block => {
    const { visibility } = block.meta
    if (!(visibility || {}).expression) return
    expressions[fetchID(block)] = visibility.expression
  })

  return $SystemAPI.expressionEvaluate({ variables, expressions }).catch(e => {
    toastErrorHandler(proxy.$t('notification.evaluate.failed'))(e)
    Object.keys(expressions).forEach(key => (expressions[key] = false))
    return expressions
  })
}

function resetErrors() {}

function setDefaultValues() {
  layouts.value = []
  layout.value = undefined
  blocks.value = undefined
  pageTitle.value = ''
}

function handleAiSearch(query) {
  const { moduleID, namespaceID, pageID } = props.page || {}
  window.dispatchEvent(new CustomEvent('show-chat-modal', { detail: {
    namespace: namespaceID,
    module: moduleID,
    page: pageID,
    prompt: query,
  } }))
}
</script>
