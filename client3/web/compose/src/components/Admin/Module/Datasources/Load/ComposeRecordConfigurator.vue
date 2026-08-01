<template>
  <div>
    <div class="row g-0">
      <div class="col">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('datasources.namespace') }}</label>
          <select
            v-model="namespace"
            class="form-select form-control"
            @change="selectNamespace"
          >
            <option :value="undefined">{{ $t('label.none') }}</option>
            <option
              v-for="ns in namespaces"
              :key="ns.namespaceID"
              :value="ns.namespaceID"
            >{{ ns.name }}</option>
          </select>
        </div>
      </div>
      <div class="col">
        <div
          v-if="namespace"
          class="mb-3"
        >
          <label class="form-label text-primary">{{ $t('datasources.module') }}</label>
          <select
            v-model="module"
            class="form-select form-control"
          >
            <option :value="undefined">{{ $t('label.none') }}</option>
            <option
              v-for="mod in modules"
              :key="mod.moduleID"
              :value="mod.moduleID"
            >{{ mod.name }}</option>
          </select>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="js">
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps({
  definition: {
    type: Object,
    required: true,
    default: () => ({}),
  },
})

const emit = defineEmits(['update:definition'])

const processing = ref(false)
const namespaces = ref([])
const modules = ref([])
const $ComposeAPI = window.__composeAPI

const namespace = computed({
  get () {
    return props.definition.namespaceID
  },
  set (namespaceID) {
    emit('update.definition', { ...props.definition, namespaceID })
  },
})

const module = computed({
  get () {
    return props.definition.moduleID
  },
  set (moduleID) {
    emit('update.definition', { ...props.definition, moduleID })
  },
})

const filter = computed({
  get () {
    return props.definition.filter
  },
  set (filter) {
    emit('update.definition', { ...props.definition, filter })
  },
})

const sort = computed({
  get () {
    return props.definition.sort
  },
  set (sort) {
    emit('update.definition', { ...props.definition, sort })
  },
})

watch(namespace, (ns) => {
  if (ns) {
    processing.value = true
    fetchModules(ns)
      .finally(() => {
        processing.value = false
      })
  }
}, { immediate: true })

onMounted(() => {
  processing.value = true
  fetchNamespaces()
    .then(() => {
      if (namespace.value) {
        return fetchModules(namespace.value)
      }
    }).finally(() => {
      processing.value = false
    })
})

function selectNamespace () {
  module.value = undefined
}

function fetchNamespaces () {
  return $ComposeAPI.namespaceList({ sort: 'name' }).then(({ set = [] }) => {
    namespaces.value = set
  }).catch((e) => {
    toastErrorHandler(t('notification.namespace.fetch-failed'))(e)
  })
}

function fetchModules (namespaceID) {
  return $ComposeAPI.moduleList({ namespaceID, sort: 'name' }).then(({ set = [] }) => {
    modules.value = set
  }).catch((e) => {
    toastErrorHandler(t('notification.module.fetch-failed'))(e)
  })
}

function toastErrorHandler (msg) {
  return (err) => {
    console.error(msg, err)
  }
}
</script>
