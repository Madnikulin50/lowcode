<template>
  <tr>
    <td />
    <td>
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('navigation.fieldLabel') }}</label>
        <input v-model="options.item.label" class="form-control" type="text" />
      </div>
    </td>
    <td style="min-width: 200px;">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('navigation.composePage') }}</label>
        <c-input-select
          v-model="options.item.pageID"
          :placeholder="$t('navigation.none')"
          :options="pageList"
          :get-option-key="getOptionKey"
          label="title"
          :reduce="f => f.pageID"
          option-value="pageID"
          @input="updateLabelValue"
        />
      </div>
    </td>
    <td style="min-width: 200px;">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('navigation.pageLayout') }}</label>
        <c-input-select
          v-model="options.item.pageLayoutID"
          :placeholder="$t('navigation.defaultLayout')"
          :options="pageLayoutList"
          :get-option-key="getLayoutOptionKey"
          :get-option-label="getLayoutOptionLabel"
          :reduce="f => f.pageLayoutID"
          option-value="pageLayoutID"
          :loading="loadingPageLayouts"
          :disabled="!options.item.pageID"
        />
      </div>
    </td>
    <td style="min-width: 200px;">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('navigation.openIn') }}</label>
        <select v-model="options.item.target" class="form-select form-control">
          <option value="sameTab">{{ $t('navigation.sameTab') }}</option>
          <option value="newTab">{{ $t('navigation.newTab') }}</option>
        </select>
      </div>
    </td>
    <td v-if="selectedPageChildren(options.item.pageID).length > 0" class="align-middle text-center">
      <div class="mb-3">
        <label class="form-label text-nowrap text-primary">{{ $t('navigation.displaySubPages') }}</label>
        <div class="form-check form-switch">
          <input
            v-model="options.item.displaySubPages"
            class="form-check-input"
            type="checkbox"
            role="switch"
          />
        </div>
      </div>
    </td>
    <td />
  </tr>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { composables } from 'corteza-lib/vue/dist'
import { NoID, compose } from 'corteza-lib/js/dist'

const { toastErrorHandler } = composables.useToast()
const { t: $t } = useI18n({ useScope: 'global' })
const $ComposeAPI = window.__composeAPI

const props = defineProps({
  item: { type: Object, required: true },
  namespace: { type: compose.Namespace, required: true },
})

const loadingPageLayouts = ref(false)
const pageList = ref([])
const pageLayoutList = ref([])

const options = computed({
  get: () => props.item.options,
  set: (opts) => { props.item.options = opts },
})

watch(() => options.value.item.pageID, (pageID) => {
  if (pageID && pageID !== NoID) {
    loadPageLayouts(pageID)
  } else {
    pageLayoutList.value = []
  }
}, { immediate: true })

onMounted(() => {
  loadPages()
})

onBeforeUnmount(() => {
  setDefaultValues()
})

function selectedPageChildren(pageID) {
  return pageList.value.filter(value => value.selfID === pageID && value.moduleID === NoID) || []
}

function loadPages() {
  const { namespaceID } = props.namespace
  $ComposeAPI
    .pageList({ namespaceID, sort: 'title' })
    .then(({ set: pages }) => {
      pageList.value = pages.map(p => new compose.Page(p))
    })
    .catch(toastErrorHandler($t('notification.page.listFailed')))
}

function loadPageLayouts(pageID) {
  const { namespaceID } = props.namespace
  loadingPageLayouts.value = true
  $ComposeAPI
    .pageLayoutList({ namespaceID, pageID, sort: 'weight ASC' })
    .then(({ set: layouts }) => {
      pageLayoutList.value = layouts.map(pl => new compose.PageLayout(pl))
    })
    .catch(toastErrorHandler($t('notification.page-layout.listFailed')))
    .finally(() => {
      loadingPageLayouts.value = false
    })
}

function updateLabelValue(pageID) {
  const composePage = pageList.value.find(t => t.pageID === pageID)
  if (!options.value.item.label) {
    options.value.item.label = composePage ? composePage.title : ''
  }
  options.value.item.pageLayoutID = ''
  options.value.item.moduleID = composePage && composePage.moduleID !== NoID ? composePage.moduleID : ''
}

function getOptionKey({ pageID }) {
  return pageID
}

function getLayoutOptionKey({ pageLayoutID }) {
  return pageLayoutID
}

function getLayoutOptionLabel({ handle, meta = {}, pageLayoutID }) {
  return meta.title || handle || pageLayoutID
}

function setDefaultValues() {
  pageList.value = []
  pageLayoutList.value = []
}
</script>

<style lang="scss" scoped>
th, td {
  padding-left: 15px;
  padding-right: 15px;
}
</style>
