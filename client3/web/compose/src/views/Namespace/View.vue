<template>
  <div
    id="namespace-view"
    class="d-flex w-100 h-100"
  >
    <router-view
      v-if="loaded && namespace"
      :namespace="namespace"
    />

    <div
      v-else
      class="loader flex-column align-items-center justify-content-center w-100 h-50"
    >
      <h1>
        {{ namespace ? (namespace.name || namespace.slug || namespace.namespaceID) : '...' }}
      </h1>

      <div class="d-flex align-items-center justify-content-center mt-4">
        <span class="spinner-border" />
        <h4 class="mb-0 ms-2">
          {{ $t('label.loading') }}
        </h4>
      </div>
    </div>

    <attachment-modal />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'general' } })
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useNamespaceStore } from '../../store/namespace'
import { useModuleStore } from '../../store/module'
import { useChartStore } from '../../store/chart'
import { usePageStore } from '../../store/page'
import { usePageLayoutStore } from '../../store/page-layout'
import { composables } from 'corteza-lib/vue/dist'
import AttachmentModal from 'corteza-webapp-compose/src/components/Public/Page/Attachment/Modal'

const { t } = useI18n()
const router = useRouter()
const nsStore = useNamespaceStore()
const moduleStore = useModuleStore()
const chartStore = useChartStore()
const pageStore = usePageStore()
const pageLayoutStore = usePageLayoutStore()

const { toastDanger } = composables.useToast()

const props = defineProps({
  slug: {
    required: true,
    type: String,
  },
})

const loaded = ref(false)
const error = ref('')
const namespace = ref(null)

const parts = computed(() => ({
  namespace: nsStore.pending,
  module: moduleStore.pending,
  page: pageStore.pending,
  chart: chartStore.pending,
  pageLayout: pageLayoutStore.pending,
}))

watch(() => props.slug, (slug) => {
  loaded.value = false

  let ns = nsStore.getByUrlPart(slug)

  if (!ns) {
    nsStore.load({ force: true }).then(() => {
      ns = nsStore.getByUrlPart(slug)
      namespace.value = ns
      prepareNamespace()
    }).catch((err) => {
      errHandler(err)
      loaded.value = true
      router.push({ name: 'root' })
    })
  } else {
    namespace.value = ns
    prepareNamespace()
  }
}, { immediate: true })

error.value = ''

onBeforeUnmount(() => {
  setDefaultValues()
})

function prepareNamespace() {
  if (!namespace.value) {
    router.push({ name: 'root' })
    return
  }

  if (!namespace.value.enabled) {
    toastDanger(t('notification.namespace.disabled'))
    router.push({ name: 'root' })
    return
  }

  const p = { namespace: namespace.value, namespaceID: namespace.value.namespaceID, clear: true }

  moduleStore.clearSet()
  chartStore.clearSet()
  pageStore.clearSet()

  window.dispatchEvent(new CustomEvent('check-namespace-sidebar', { detail: !namespace.value.meta.hideSidebar }))

  // allSettled: a failed module/chart/page load must not leave the spinner forever.
  // The previous `.catch(errHandler)` re-rejected, so Promise.all never reached `.then`.
  Promise.allSettled([
    moduleStore.load(p),
    chartStore.load(p),
    pageStore.load(p),
    pageLayoutStore.load(p),
  ]).then((results) => {
    const seen = new Set()
    for (const r of results) {
      if (r.status !== 'rejected') continue
      const key = r.reason?.message || String(r.reason)
      if (seen.has(key)) continue
      seen.add(key)
      errHandler(r.reason)
    }
  }).finally(() => {
    setTimeout(() => {
      loaded.value = true
    }, 500)
  })
}

function errHandler(err) {
  if (!err) return
  switch ((err.response || {}).status) {
    case 403:
      error.value = t('notification.general.composeAccessNotAllowed')
      toastDanger(error.value)
      return
  }
  const msg = err.message || (typeof err === 'string' ? err : '')
  if (msg) toastDanger(msg)
}

function setDefaultValues() {
  loaded.value = false
  error.value = ''
  namespace.value = null
}
</script>

<style lang="scss" scoped>
.error {
  font-size: 24px;
  background-color: var(--white);
  width: 100vw;
  height: 20vh;
  padding: 60px;
  top: 40vh;
}

.loader {
  height: calc(100vh - 2 * #{var(--topbar-height)});
  display: flex;
  align-items: center;
  justify-content: space-around;

  .pending {
    width: 30px;
  }

  .logo {
    height: 30px;
    background-position: center;
    background-repeat: no-repeat;
    background-size: 130px;
  }
}
</style>
