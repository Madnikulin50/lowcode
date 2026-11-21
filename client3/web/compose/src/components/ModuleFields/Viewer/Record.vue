<template>
  <span v-if="processing" class="spinner-border spinner-border-sm text-primary"></span>
  <div v-else @click.stop>
    <span
      v-for="(v, index) of formattedValue"
      :key="v.id || index"
      :class="{ 'd-block': field.options.multiDelimiter === '\n', 'mt-1': field.options.multiDelimiter === '\n' && index !== 0 }"
    >
      <template v-if="disableClick">
        <FieldViewer v-if="v.record" :field="labelField" :record="v.record" :namespace="namespace" disable-click value-only />
        <template v-else>{{ v.id }}</template>
        {{ index !== formattedValue.length - 1 ? field.options.multiDelimiter : '' }}
      </template>
      <a
        v-else-if="['modal', 'newTab'].includes(extraOptions.recordSelectorDisplayOption)"
        href="#"
        :class="{ 'text-decoration-none default-cursor': !v.to}"
        @click="onRecordSelectorClick($event, v.to)"
      >
        <FieldViewer v-if="v.record" :field="labelField" :record="v.record" :namespace="namespace" disable-click value-only />
        <template v-else>{{ v.id }}</template>
        {{ index !== formattedValue.length - 1 ? field.options.multiDelimiter : '' }}
      </a>
      <router-link
        v-else
        :to="v.to"
        :class="{ 'text-decoration-none default-cursor': !v.to}"
      >
        <FieldViewer v-if="v.record" :field="labelField" :record="v.record" :namespace="namespace" disable-click value-only />
        <template v-else>{{ v.id }}</template>
        {{ index !== formattedValue.length - 1 ? field.options.multiDelimiter : '' }}
      </router-link>
    </span>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, defineAsyncComponent } from 'vue'
import { useRouter } from 'vue-router'
import { compose, NoID } from 'corteza-lib/js/dist'
import { useViewerBase } from './useViewerBase'
import { useModuleStore } from 'corteza-webapp-compose/src/stores/module'
import { useRecordStore } from 'corteza-webapp-compose/src/stores/record'
import { usePageStore } from 'corteza-webapp-compose/src/stores/page'

const FieldViewer = defineAsyncComponent(() => import('./index.vue'))

const router = useRouter()
const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  field: { type: compose.ModuleField, required: true },
  record: { type: compose.Record, required: true },
  valueOnly: { type: Boolean, required: false },
  extraOptions: { type: Object, default: () => ({}) },
  includeStyles: { type: Boolean, default: false },
  disableClick: { type: Boolean, default: false },
})

const { value, inModal } = useViewerBase(props)

const moduleStore = useModuleStore()
const recordStore = useRecordStore()
const pageStore = usePageStore()

const processing = ref(false)
const recordValues = ref({})

const recordPage = computed(() => pageStore.set.find(p => p.moduleID === props.field.options.moduleID))

const recordModule = computed(() => {
  if (!props.field.options.moduleID) return undefined
  return moduleStore.getByID(props.field.options.moduleID)
})

const labelField = computed(() => {
  if (!props.field.options.labelField || !recordModule.value) return undefined
  let lf = recordModule.value.fields.find(({ name }) => name === props.field.options.labelField)
  if (!lf) return undefined
  lf = compose.ModuleFieldMaker(lf)
  if (lf.kind === 'Record' && props.field.options.recordLabelField) {
    lf.options.labelField = props.field.options.recordLabelField
  }
  return lf
})

const formattedValue = computed(() => {
  const arr = Array.isArray(value.value) ? value.value : [value.value]
  return arr.map(recordID => {
    let record = recordStore.findByID(recordID)
    if (record) record = new compose.Record(recordModule.value, record)
    return { id: recordID, to: linkToRecord(recordID), record }
  })
})

function linkToRecord (recordID) {
  if (!recordPage.value || !recordID) return ''
  return { name: 'page.record', params: { pageID: recordPage.value.pageID, recordID } }
}

onMounted(() => {
  formatRecordValues(value.value)
})

onBeforeUnmount(() => {
  setDefaultValues()
})

function formatRecordValues (recordIDs) {
  recordIDs = Array.isArray(recordIDs) ? recordIDs : [recordIDs].filter(v => v) || []
  const { namespaceID = NoID } = props.namespace
  const { moduleID = NoID, recordLabelField } = props.field.options

  if (!recordIDs.length || [moduleID, namespaceID].includes(NoID) || !labelField.value || !recordModule.value) return

  const records = recordStore.findByIDs(recordIDs).map(r => new compose.Record(recordModule.value, r))

  if (labelField.value.kind === 'Record' && recordLabelField) {
    processing.value = true
    setTimeout(() => { processing.value = false }, 300)
  } else if (labelField.value.kind === 'User') {
    processing.value = true
    setTimeout(() => { processing.value = false }, 300)
  }
}

function onRecordSelectorClick (e, route) {
  e.preventDefault()
  if (!route) return

  if (props.extraOptions.recordSelectorDisplayOption === 'modal' || inModal.value) {
    if (route.params?.recordID && route.params?.pageID) {
      window.dispatchEvent(new CustomEvent('show-record-modal', { detail: { recordID: route.params.recordID, recordPageID: route.params.pageID } }))
    }
  } else if (props.extraOptions.recordSelectorDisplayOption === 'newTab') {
    const resolved = router.resolve(route)
    if (resolved?.href) window.open(resolved.href, '_blank')
  }
}

function setDefaultValues () {
  processing.value = false
  recordValues.value = {}
}
</script>

<style lang="scss" scoped>
.default-cursor {
  cursor: default;
}
</style>
