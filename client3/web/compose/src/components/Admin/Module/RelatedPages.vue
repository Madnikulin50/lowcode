<template>
  <div v-if="namespace.canManageNamespace">
    <div
      v-if="recordPage"
      class="dropdown flex-fill related-pages-dropdown"
    >
      <button
        id="related-pages-dropdown"
        class="btn btn-primary dropdown-toggle"
        :style="size === 'md' ? '' : ''"
        type="button"
        data-bs-toggle="dropdown"
        :disabled="!module.name"
      >
        {{ $t('related-pages') }}
      </button>
      <ul
        class="dropdown-menu"
        aria-labelledby="related-pages-dropdown"
      >
        <li v-if="recordPage">
          <router-link
            data-test-id="dropdown-link-record-page-edit"
            class="dropdown-item"
            :disabled="!namespace.canManageNamespace"
            :to="{ name: 'admin.pages.builder', params: { pageID: recordPage.pageID } }"
          >
            {{ $t('recordPage.edit') }}
          </router-link>
        </li>

        <li v-if="recordListPage">
          <router-link
            class="dropdown-item"
            :to="{ name: 'admin.pages.builder', params: { pageID: recordListPage.pageID } }"
          >
            {{ $t('recordListPage.edit') }}
          </router-link>
        </li>

        <li v-else>
          <button
            data-test-id="dropdown-link-record-list-page-create"
            class="dropdown-item"
            :disabled="processing"
            @click.prevent.stop="handleRecordListPageCreation"
          >
            {{ $t('recordListPage.create') }}
          </button>
        </li>
      </ul>
    </div>

    <button
      v-else
      data-test-id="button-record-page-create"
      class="btn btn-primary"
      :disabled="processing || !module.name"
      @click.stop.prevent="handleRecordPageCreation"
    >
      {{ $t('recordPage.create') }}
    </button>
  </div>
</template>

<script setup lang="js">
import { ref, computed } from 'vue'
import { useStore } from '../../../store'
import { useI18n } from 'vue-i18n'
import { compose, NoID } from 'corteza-lib/js/dist'

const { t } = useI18n()

defineOptions({
  i18nOptions: {
    namespaces: 'module',
  },
})

const props = defineProps({
  namespace: {
    type: Object,
    required: true,
  },
  module: {
    type: Object,
    required: true,
  },
  size: {
    type: String,
    default: 'md',
  },
  boundary: {
    type: String,
    default: 'viewport',
  },
})

const store = useStore()

const processing = ref(false)

const pages = computed(() => store.getters['page/set'])

const recordPage = computed(() => {
  return pages.value.find(p => p.moduleID === props.module.moduleID)
})

const recordListPage = computed(() => {
  return pages.value.find(p => {
    return p.blocks.find(b => b.kind === 'RecordList' && b.options.moduleID === props.module.moduleID)
  })
})

function createPage (...args) {
  return store.dispatch('page/create', ...args)
}

function updatePage (...args) {
  return store.dispatch('page/update', ...args)
}

function createPageLayout (...args) {
  return store.dispatch('pageLayout/create', ...args)
}

function handleRecordPageCreation () {
  processing.value = true

  const { name, moduleID } = props.module
  const { namespaceID } = props.namespace

  const blocks = [new compose.PageBlockRecord({ xywh: [0, 0, 48, 82] })]
  const selfID = (recordListPage.value || {}).pageID || NoID

  const page = new compose.Page({
    namespaceID,
    moduleID,
    selfID,
    title: t('forModule.recordPage', { name, interpolation: { escapeValue: false } }),
    blocks,
  })

  createPage(page).then(({ pageID, title, blocks }) => {
    const pageLayout = new compose.PageLayout({ namespaceID, pageID, handle: 'primary', blocks, meta: { title } })
    return createPageLayout(pageLayout)
  }).catch(toastErrorHandler(t('notification.module.recordPage.createFailed')))
    .finally(() => {
      processing.value = false
    })
}

function handleRecordListPageCreation () {
  processing.value = true

  const { namespaceID } = props.namespace
  const { name, moduleID } = props.module

  const blocks = [new compose.PageBlockRecordList({
    xywh: [0, 0, 48, 82],
    options: {
      moduleID,
      fields: [],
    },
  })]

  const page = new compose.Page({
    title: name,
    namespaceID,
    blocks,
    visible: true,
  })

  createPage(page)
    .then(({ pageID, title, blocks }) => {
      const pageLayout = new compose.PageLayout({ namespaceID, pageID, handle: 'primary', blocks, meta: { title } })
      return Promise.all([
        updatePage({ ...recordPage.value, selfID: pageID }),
        createPageLayout(pageLayout),
      ])
    })
    .catch(toastErrorHandler(t('notification.module.recordPage.createFailed')))
    .finally(() => {
      processing.value = false
    })
}

function toastErrorHandler (msg) {
  return (err) => {
    console.error(msg, err)
  }
}
</script>
