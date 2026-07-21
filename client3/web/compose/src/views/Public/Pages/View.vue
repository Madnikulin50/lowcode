<template>
  <div
    v-if="!!page"
    class="d-flex w-100 overflow-hidden"
  >
    <Teleport
      v-if="!isRecordPage"
      to="#topbar-title"
    >
      {{ pageTitle }}
    </Teleport>

    <Teleport
      v-if="!isRecordPage"
      to="#topbar-tools"
    >
        <c-input-search
          v-if="enableAI"
          v-model.trim="aiPrompt"
          class="me-2"
          :ai="true"
          :aria-label="$t('AI')"
          :placeholder="$t('aiChat.startPrompt')"
          :autocomplete="'off'"
          submittable
          @search="handleAiSearch"
        />
        <div
          v-if="page && page.canUpdatePage"
          class="btn-group btn-group-sm"
        >
          <router-link
            data-test-id="button-page-builder"
            :to="pageBuilder"
            class="btn btn-primary d-flex align-items-center"
          >
            {{ $t('label.pageBuilder') }}
            <font-awesome-icon
              :icon="['fas', 'tools']"
              class="ms-2"
            />
          </router-link>

          <router-link
            data-test-id="button-page-edit"
            :to="pageEditor"
            class="btn btn-primary d-flex align-items-center"
            :title="$t('tooltip.edit.page')"
            style="margin-left:2px;"
            data-bs-toggle="tooltip"
            data-bs-boundary="body"
          >
            <font-awesome-icon
              :icon="['far', 'edit']"
            />
          </router-link>

          <page-translator
            v-if="trPage"
            data-test-id="button-page-translations"
            :page="trPage"
            :page-layout="layout"
            button-variant="primary"
            style="margin-left:2px;"
          />
      </div>

    </Teleport>

    <div class="flex-grow-1 overflow-auto d-flex p-2 w-100">
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
import bus from '../../../lib/bus'
import Grid from 'corteza-webapp-compose/src/components/Public/Page/Grid'
import RecordModal from 'corteza-webapp-compose/src/components/Public/Record/Modal'
import MagnificationModal from 'corteza-webapp-compose/src/components/Public/Page/Block/Modal'
import AiChatModal from 'corteza-webapp-compose/src/components/Public/Page/AiChat/Modal'
import PageTranslator from 'corteza-webapp-compose/src/components/Admin/Page/PageTranslator'

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

const layouts = ref([])
const layout = ref(undefined)
const blocks = ref(undefined)
const tempRecord = ref(undefined)
const pageTitle = ref('')
const aiPrompt = ref('')

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
  const [pageID = '', pageLayoutID = ''] = value || []
  const [oldPageID = ''] = oldValue || []

  if (!pageID || pageID === NoID) return

  if (pageID !== oldPageID) {
    layouts.value = pageLayoutStore.getByPageID(props.page.pageID)
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

async function determineLayout({ pageLayoutID, redirectOnFail = true } = {}) {
  if (isRecordPage.value) {
    resetErrors()
  }

  let expressions = {}

  if (layouts.value.some(({ config = {} }) => config.visibility?.expression)) {
    expressions = await evaluateLayoutExpressions()
  }

  const matchedLayout = layouts.value.find(l => {
    if (pageLayoutID && l.pageLayoutID !== pageLayoutID) return false
    const { expression, roles = [] } = l.config.visibility || {}
    if (expression && !expressions[l.pageLayoutID]) return false
    if (!roles.length) return true
    return $auth.user.roles.some(roleID => roles.includes(roleID))
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

async function prepareBlocks() {
  blocks.value = undefined
  const tempBlocks = []
  const layoutBlocks = layout.value?.blocks || []
  const tabbedIDs = new Set()

  layoutBlocks.forEach(({ blockID, xywh }) => {
    const block = props.page.blocks.find(b => b.blockID === blockID)
    if (block) {
      block.xywh = xywh
      tempBlocks.push(block)
      if (block.kind === 'Tabs') {
        const { tabs = [] } = block.options
        tabs.forEach(t => {
          if (t.blockID && !layoutBlocks.some(b => b.blockID === t.blockID)) {
            tabbedIDs.add(t.blockID)
          }
        })
      }
    }
  })

  props.page.blocks.forEach(block => {
    if (tabbedIDs.has(block.blockID)) {
      tempBlocks.push(block)
    }
  })

  return evaluateBlocks(tempBlocks)
}

async function evaluateBlocks(pageBlocks = props.page.blocks) {
  let layoutBlocksExpressions = {}

  if (pageBlocks.some(({ meta = {} }) => (meta.visibility || {}).expression)) {
    layoutBlocksExpressions = await evaluateBlocksExpressions(pageBlocks)
  }

  pageBlocks.forEach(block => {
    const { meta = {} } = block
    const blockID = fetchID(block)
    const visibility = meta.visibility || {}
    const { roles = [] } = visibility
    const validExpression = !visibility.expression || layoutBlocksExpressions[blockID]
    const validRole = !roles.length || $auth.user.roles.some(roleID => roles.includes(roleID))
    const showBlock = block && validExpression && validRole
    if (!block.meta) block.meta = {}
    block.meta.invisible = !showBlock
  })

  pageBlocks.forEach(block => {
    if (block.kind === 'Tabs' && !block.meta.invisible) {
      const { tabs = [] } = block.options
      const hasVisibleTab = tabs.some(t => {
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
  const { moduleID, namespaceID, pageID } = props.page
  window.dispatchEvent(new CustomEvent('show-chat-modal', { detail: {
    namespace: namespaceID,
    module: moduleID,
    page: pageID,
    prompt: query,
  } }))
}
</script>
