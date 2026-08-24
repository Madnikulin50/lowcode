<template>
  <div
    v-if="showModal"
    class="modal d-block magnification-modal"
    tabindex="-1"
    role="dialog"
    @click.self="onHidden"
  >
    <div class="modal-dialog modal-xl" :class="dialogClass" role="document">
      <div class="modal-content" :class="contentClass">
        <PageBlock
          v-if="showModal"
          :block="block"
          :blocks="page.blocks"
          :module="module"
          :record="record"
          :page="page"
          :namespace="namespace"
          magnified
        />
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n({ useScope: 'global' })
import { compose, NoID } from 'corteza-lib/js/dist'
import { fetchID } from 'corteza-webapp-compose/src/lib/block'
import { usePageStore } from 'corteza-webapp-compose/src/store/page'
import { useModuleStore } from 'corteza-webapp-compose/src/store/module'
import { composables } from 'corteza-lib/vue/dist'
import PageBlock from 'corteza-webapp-compose/src/components/PageBlocks/index'

const { toastErrorHandler } = composables.useToast()

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
})

const $ComposeAPI = window.__composeAPI
const route = useRoute()
const router = useRouter()
const pageStore = usePageStore()
const moduleStore = useModuleStore()

const showModal = ref(false)
const block = ref(undefined)
const record = ref(undefined)
const page = ref(undefined)
const module = ref(undefined)
const customBlock = ref(undefined)

const dialogClass = computed(() => {
  return block.value && block.value.options?.magnifyOption === 'fullscreen' ? 'h-100 mw-100 m-0 mh-100' : 'modal-max-width'
})

const contentClass = computed(() => {
  return `${block.value && block.value.options?.magnifyOption === 'fullscreen' ? 'mh-100 rounded-0' : 'modal-max-height'} position-initial`
})

const magnifiedBlockID = computed(() => route.query.magnifiedBlockID)

watch(magnifiedBlockID, (newVal) => {
  if (!newVal) {
    setDefaultValues()
  } else if (newVal !== fetchID(block.value)) {
    loadModal(newVal)
  }
}, { immediate: true })

function magnifyPageBlock(detail) {
  const { blockID, block: b } = detail || {}

  if (!blockID && !b) {
    onHidden()
    return
  }

  customBlock.value = b
  const magnifiedID = blockID || (b || {}).blockID
  loadModal(magnifiedID)

  router.push({
    query: {
      ...route.query,
      magnifiedBlockID: magnifiedID,
    },
  })
}

function loadModal(blockID) {
  const { recordID: paramsRecordID, pageID } = route.params
  const { recordID: queryRecordID, recordPageID } = route.query

  page.value = pageStore.getByID(recordPageID || pageID)

  if (!page.value) return

  const { namespaceID, moduleID } = page.value
  const recordID = paramsRecordID || queryRecordID
  block.value = customBlock.value || page.value.blocks.find(b => fetchID(b) === blockID)
  const mod = moduleID !== NoID ? moduleStore.getByID(moduleID) : undefined
  module.value = mod

  showModal.value = !!(block.value || {}).blockID

  if (recordID) {
    $ComposeAPI
      .recordRead({ namespaceID, moduleID, recordID })
      .then(raw => {
        record.value = new compose.Record(mod, raw)
      })
      .catch(toastErrorHandler($t('notification.record.loadFailed')))
  } else if (mod) {
    record.value = new compose.Record(mod, {})
  }
}

function onHidden() {
  if (route.query.magnifiedBlockID !== undefined) {
    router.push({
      query: {
        ...route.query,
        magnifiedBlockID: undefined,
      },
    })
  }
}

function setDefaultValues() {
  showModal.value = false
  block.value = undefined
  record.value = undefined
  page.value = undefined
  module.value = undefined
  customBlock.value = undefined
}

let magnifyHandler

function destroyEvents() {
  if (magnifyHandler) {
    window.removeEventListener('magnify-page-block', magnifyHandler)
    magnifyHandler = undefined
  }
}

onMounted(() => {
  magnifyHandler = (event) => magnifyPageBlock(event.detail)
  window.addEventListener('magnify-page-block', magnifyHandler)
})

onBeforeUnmount(() => {
  destroyEvents()
  setDefaultValues()
})
</script>

<style lang="scss">
.magnification-modal {
  .modal-dialog {
    display: flex;
    flex-direction: column;

    .modal-content {
      flex: 1 1 auto;
    }
  }
}

.position-initial {
  position: initial;
}

.modal-max-width {
  max-width: 97vw;
}

.modal-max-height {
  max-height: 80vh;
  overflow-y: auto;
}
</style>
