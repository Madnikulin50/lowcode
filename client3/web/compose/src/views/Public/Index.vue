<template>
  <div
    v-if="namespace"
    class="d-flex flex-column h-100 w-100"
  >
    <div
      v-if="showSteps"
      class="d-flex flex-column m-5 vh-75"
    >
      <h1 class="display-3">
        {{ $t('label.welcome') }}
      </h1>

      <p class="lead">
        {{ $t('message.noPages') }}
        <span v-if="namespace.canManageNamespace">
          {{ $t('message.startBuilding') }}
        </span>
        <span v-else>
          {{ $t('message.notifyAdministrator') }}
        </span>
      </p>

      <div
        v-if="namespace.canManageNamespace"
        class="container-fluid align-items-center border-top steps"
      >
        <div class="row align-items-center text-center justify-content-between">
          <div class="col">
            <circle-step
              step-number="1"
              :done="hasModules"
            >
              <router-link
                v-if="!hasModules"
                :to="{ name: 'admin.modules.create' }"
                class="btn btn-outline-primary"
                :class="{ disabled: !namespace.canCreateModule }"
                data-test-id="button-module-create"
              >
                {{ $t('step.module.create') }}
              </router-link>
              <router-link
                v-else
                :to="{ name: 'admin.modules' }"
                class="btn btn-primar"
                :class="{ disabled: !namespace.canManageNamespace }"
                data-test-id="button-module-view"
              >
                {{ $t('step.module.view') }}
              </router-link>
            </circle-step>
          </div>
          <div class="col">
            <hr>
          </div>
          <div class="col">
            <circle-step
              :done="hasCharts"
              :disabled="!hasModules"
              optional
            >
              <router-link
                v-if="!hasCharts"
                :to="{ name: 'admin.charts.create', params: { category: 'generic' } }"
                class="btn btn-outline-primary btn-lg"
                :class="{ disabled: !hasModules || !namespace.canCreateChart }"
              >
                {{ $t('step.chart.create') }}
              </router-link>
              <router-link
                v-else
                :to="{ name: 'admin.charts' }"
                class="btn btn-primary btn-lg"
                :class="{ disabled: !namespace.canManageNamespace }"
              >
                {{ $t('step.chart.view') }}
              </router-link>
            </circle-step>
          </div>
          <div class="col">
            <hr>
          </div>
          <div class="col">
            <circle-step
              step-number="2"
              :done="hasPages"
            >
              <button
                v-if="!hasPages"
                data-test-id="button-page-build"
                :disabled="!namespace.canCreatePage"
                class="btn btn-outline-primary btn-lg"
                @click="createNewPage"
              >
                {{ $t('step.page.create') }}
              </button>
              <router-link
                v-else
                :to="{ name: 'admin.pages' }"
                class="btn btn-primary btn-lg"
                :class="{ disabled: !namespace.canManageNamespace }"
                data-test-id="button-page-view"
              >
                {{ $t('step.page.view') }}
              </router-link>
            </circle-step>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="d-flex flex-column h-100 w-100">
      <router-view
        class="flex-grow-1 overflow-auto"
        :namespace="namespace"
        :page="page"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'onboarding' } })
import { ref, computed, watch, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue'
import { useRouter } from 'vue-router'
import { compose } from 'corteza-lib/js/dist'
import { composables } from 'corteza-lib/vue/dist'
import { usePageStore } from '../../store/page'
import { usePageLayoutStore } from '../../store/page-layout'
import { useModuleStore } from '../../store/module'
import { useChartStore } from '../../store/chart'
import CircleStep from 'corteza-webapp-compose/src/components/Common/CircleStep'

const { toastErrorHandler } = composables.useToast()
const { proxy } = getCurrentInstance()

const demoPageHandle = 'demo_page'

const props = defineProps({
  pageID: {
    type: String,
    required: false,
    default: '',
  },
  namespace: {
    type: compose.Namespace,
    required: true,
  },
})

const router = useRouter()
const pageStore = usePageStore()
const pageLayoutStore = usePageLayoutStore()
const moduleStore = useModuleStore()
const chartStore = useChartStore()

const navVisible = ref(false)
const documentWidth = ref(0)
const loaded = ref(false)

const page = computed(() => {
  return pageStore.getByID(props.pageID) || new compose.Page()
})

const showSteps = computed(() => !props.pageID && loaded.value)

const hasModules = computed(() => !!moduleStore.set.length)

const hasCharts = computed(() => !!chartStore.set.length)

const hasPages = computed(() => {
  return pageStore.set.filter(p => p.visible || p.handle === demoPageHandle).length > 0
})

watch(() => props.pageID, (pageID) => {
  if (!pageID) {
    const homePage = pageStore.homePage
    if (homePage?.pageID) {
      router.replace({ name: 'page', params: { pageID: homePage.pageID } })
    } else {
      document.title = props.namespace?.name || props.namespace?.slug || proxy.$t('label.app-name.namespace.view')
      loaded.value = true
    }
  }
}, { immediate: true })

onMounted(() => {
  documentWidth.value = document.body.offsetWidth
  const onResize = () => {
    documentWidth.value = document.body.offsetWidth
  }
  window.addEventListener('resize', onResize)
})

onBeforeUnmount(() => {
  window.onresize = null
  setDefaultValues()
})

function createNewPage() {
  const namespaceID = props.namespace?.namespaceID
  const newPage = new compose.Page({
    namespaceID,
    title: 'Demo Page',
    handle: demoPageHandle,
    visible: true,
    blocks: [],
  })

  pageStore.create(newPage).then(({ pageID, title }) => {
    const pageLayout = new compose.PageLayout({ namespaceID, pageID, handle: 'primary', meta: { title } })
    return pageLayoutStore.create(pageLayout).then(() => {
      router.push({ name: 'admin.pages.builder', params: { pageID } })
    })
  }).catch(toastErrorHandler(proxy.$t('notification.page.saveFailed')))
}

function setDefaultValues() {
  navVisible.value = false
  documentWidth.value = 0
  loaded.value = false
}
</script>

<style lang="scss" scoped>
.steps {
  padding: 0;
  padding-top: 20vh;
}
</style>
