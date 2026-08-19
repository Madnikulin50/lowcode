<template>
  <div
    class="h-100 p-2"
  >
    <Teleport to="#topbar-title">
      {{ title }}
    </Teleport>

    <Teleport to="#topbar-tools">
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
        @ai-search="handleAiSearch"
      />

      <div
        v-if="modulePage"
        class="btn-group"
      >
        <router-link
          :disabled="!modulePage"
          :to="modulePage"
          class="btn btn-primary d-flex align-items-center"
          style="margin-right:2px;"
        >
          {{ $t('edit.edit') }}
          <font-awesome-icon
            :icon="['far', 'edit']"
            size="sm"
            class="ms-2"
          />
        </router-link>
      </div>
    </Teleport>

    <record-list-base
      v-if="block && page && module"
      :block="block"
      :page="page"
      :module="module"
      :namespace="namespace"
      :block-index="0"
      class="p-2"
      @save-fields="handleFieldsSave"
    />

    <ai-chat-modal
      :namespace="namespace?.namespaceID"
      :page="page?.pageID"
      :module="module?.moduleID"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'module' } })
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useStore } from '../../../../store'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { compose, NoID } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import RecordListBase from 'corteza-webapp-compose/src/components/PageBlocks/RecordListBase'
import AiChatModal from 'corteza-webapp-compose/src/components/Public/Page/AiChat/Modal.vue'

const { CInputSearch } = components
const { t } = useI18n()
const store = useStore()
const route = useRoute()

const props = defineProps({
  namespace: {
    type: Object,
    required: false,
    default: undefined,
  },
  moduleID: {
    type: String,
    required: false,
    default: '',
  },
})

const block = ref(undefined)
const aiPrompt = ref('')
const getModuleByID = computed(() => store.getters['module/getByID'])
const recordPaginationUsable = computed(() => store.getters['ui/recordPaginationUsable'])

const enableAI = computed(() => true)

const title = computed(() => {
  const { name, handle } = module.value || {}
  return t('allRecords.list.title', { name: name || handle, interpolation: { escapeValue: false } })
})

const module = computed(() => {
  if (props.moduleID) {
    return getModuleByID.value(props.moduleID)
  }
  return undefined
})

const modulePage = computed(() => {
  if (module.value) {
    return { name: 'admin.modules.edit', params: { moduleID: module.value.moduleID }, query: null }
  }
  return undefined
})

const page = computed(() => {
  if (!module.value) {
    return undefined
  }
  const { moduleID } = module.value
  return new compose.Page({ pageID: moduleID })
})

function initBlock () {
  const mdl = module.value
  if (!mdl) return

  const { meta = { ui: {} }, moduleID, config = { type: 'basic' } } = mdl
  let fields = ((meta.ui || {}).admin || {}).fields || []
  fields = fields.length
    ? fields
    : (config.type === 'basic'
      ? [...mdl.fields.slice(0, 10), ...mdl.systemFields()]
      : [...mdl.fields.slice(0, 10)])

  const b = new compose.PageBlockRecordList({
    blockIndex: 0,
    options: {
      moduleID,
      fields,
      hideRecordReminderButton: true,
      hideRecordViewButton: true,
      hideRecordCloneButton: false,
      hideRecordPermissionsButton: false,
      selectable: true,
      allowExport: true,
      perPage: 1000,
      fullPageNavigation: true,
      showTotalCount: true,
      showDeletedRecordsOption: config.type === 'basic',
      presort: 'createdAt DESC',
      enableRecordPageNavigation: true,
      hideConfigureFieldsButton: false,
      inlineRecordEditEnabled: true,
      customFilterPresets: true,
    },
  })

  b.options = {
    ...b.options,
    allRecords: true,
    rowViewUrl: 'admin.modules.record.view',
    rowEditUrl: 'admin.modules.record.edit',
    rowCreateUrl: 'admin.modules.record.create',
  }

  block.value = b

  if (recordPaginationUsable.value) {
    store.dispatch('ui/setRecordPaginationUsable', false)
  } else {
    store.dispatch('ui/clearRecordPagination')
  }
}

watch(module, (mdl) => {
  if (mdl && !block.value) {
    initBlock()
  }
}, { immediate: true })

watch(() => props.moduleID, () => {
  if (module.value) {
    if (!block.value) {
      initBlock()
      return
    }
    const { meta = { ui: {} }, moduleID } = module.value || {}
    let fields = ((meta.ui || {}).admin || {}).fields || []
    fields = fields.length ? fields : [...module.value.fields.slice(0, 10), ...module.value.systemFields()]
    block.value.options.moduleID = moduleID
    block.value.options.fields = fields
  }
})

onBeforeUnmount(() => {
  block.value = undefined
})

function handleFieldsSave (fields = []) {
  fields = fields.map((f) => f.fieldID && f.fieldID !== NoID ? f.fieldID : f.name).filter(f => !!f)
  if (!module.value.meta.ui) {
    module.value.meta.ui = { admin: { fields } }
  } else {
    module.value.meta.ui.admin = { ...(module.value.meta.ui.admin || {}), fields }
  }
  store.dispatch('module/update', module.value)
}

function handleAiSearch (query) {
  window.dispatchEvent(new CustomEvent('show-chat-modal', { detail: {
    namespace: props.namespace?.namespaceID,
    module: module.value?.moduleID,
    page: module.value?.moduleID,
    prompt: query,
  } }))
}
</script>
