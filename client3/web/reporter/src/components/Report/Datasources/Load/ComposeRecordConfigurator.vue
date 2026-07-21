<template>
  <div>
    <div class="row">
      <div class="col">
        <div class="mb-3">
          <label class="text-primary form-label">{{ t('datasources.namespace') }}</label>
          <select v-model="namespace" class="form-select" @change="selectNamespace">
            <option :value="undefined">{{ t('label.none') }}</option>
            <option v-for="ns in namespaces" :key="ns.namespaceID" :value="ns.namespaceID">{{ ns.name }}</option>
          </select>
        </div>
      </div>
      <div class="col">
        <div v-if="definition.namespaceID" class="mb-3">
          <label class="text-primary form-label">{{ t('datasources.module') }}</label>
          <select v-model="module" class="form-select">
            <option :value="undefined">{{ t('label.none') }}</option>
            <option v-for="mod in modules" :key="mod.moduleID" :value="mod.moduleID">{{ mod.name }}</option>
          </select>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'

const props = defineProps({
  definition: { type: Object, required: true },
})
const emit = defineEmits(['update:definition'])

const { t } = useI18n()
const toast = useToast()
const toastErrorHandler = toast.toastErrorHandler

const processing = ref(false)
const namespaces = ref([])
const modules = ref([])

const namespace = computed({
  get: () => props.definition.namespaceID,
  set: (namespaceID) => { emit('update.definition', { ...props.definition, namespaceID }) },
})

const module = computed({
  get: () => props.definition.moduleID,
  set: (moduleID) => { emit('update.definition', { ...props.definition, moduleID }) },
})

watch(namespace, (ns) => {
  if (ns) {
    processing.value = true
    fetchModules(ns).finally(() => { processing.value = false })
  }
})

onMounted(() => {
  processing.value = true
  fetchNamespaces().then(() => {
    if (namespace.value) return fetchModules(namespace.value)
  }).finally(() => { processing.value = false })
})

function selectNamespace() { module.value = undefined }

function fetchNamespaces() {
  return window.__composeAPI.namespaceList({ sort: 'name' })
    .then(({ set = [] }) => { namespaces.value = set })
    .catch((e) => { toastErrorHandler(t('notification.namespace.fetch-failed'))(e) })
}

function fetchModules(namespaceID) {
  return window.__composeAPI.moduleList({ namespaceID, sort: 'name' })
    .then(({ set = [] }) => { modules.value = set })
    .catch((e) => { toastErrorHandler(t('notification.module.fetch-failed'))(e) })
}
</script>