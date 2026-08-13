<template>
  <div>
    <div class="mb-3">
      <label class="text-primary form-label">{{ $t('filter.namespace.label') }}</label>
      <c-input-select
        class="namespace-selector"
        :options="namespace.options"
        :model-value="namespace.values"
        :get-option-label="getNamespaceOptionLabel"
        :get-option-key="n => `corteza::compose:namespace/${n.namespaceID}`"
        :reduce="n => `corteza::compose:namespace/${n.namespaceID}`"
        :placeholder="$t('filter.namespace.placeholder')"
        :loading="namespace.processing"
        :filterable="false"
        multiple
        @search="searchNamespaces"
        @update:model-value="updateNamespaces"
      />
    </div>

    <div
      v-for="ns in namespace.values"
      :key="ns"
      class="mb-3"
    >
      <label class="text-primary form-label">{{ getModuleLabel(ns) }}</label>
      <c-input-select
        class="module-selector"
        :options="getModulesForNamespace(ns.split('/')[1])"
        :model-value="getModuleValuesForNamespace(ns.split('/')[1])"
        :get-option-key="m => `corteza::compose:module/${m.namespaceID}/${m.moduleID}`"
        :get-option-label="getModuleOptionLabel"
        :reduce="m => `corteza::compose:module/${m.namespaceID}/${m.moduleID}`"
        :placeholder="$t('filter.module.placeholder')"
        :loading="module.processing"
        :filterable="false"
        multiple
        @search="query => searchModulesForNamespace(query, ns.split('/')[1])"
        @update:model-value="modules => updateModulesForNamespace(modules, ns.split('/')[1])"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'general' } })
import { ref, reactive, computed, onMounted, inject, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { debounce } from 'lodash'
import { components } from 'corteza-lib/vue/dist'

const { CInputSelect } = components

const { t } = useI18n()

const props = defineProps({
  namespaceLabels: { type: Array, default: () => [] },
  moduleLabels: { type: Array, default: () => [] },
})

const emit = defineEmits(['change'])

const $ComposeAPI = inject('$ComposeAPI', {})

const namespace = reactive({
  processing: false,
  values: [],
  options: [],
  filter: { query: null, limit: 20, sort: 'name DESC' },
})

const module = reactive({
  processing: false,
  values: [],
  options: [],
  filter: { query: null, limit: 20, sort: 'name DESC' },
})

const modulesByNamespace = computed(() => {
  const grouped = {}
  module.options.forEach(mod => {
    if (!grouped[mod.namespaceID]) {
      grouped[mod.namespaceID] = []
    }
    grouped[mod.namespaceID].push(mod)
  })
  return grouped
})

onMounted(() => {
  fetchNamespaces().then(() => {
    initializeFromProps()
  })
})

watch(() => props.namespaceLabels, (newVal) => {
  if (newVal && newVal.length > 0 && namespace.options.length > 0) {
    initializeFromProps()
  }
})

function initializeFromProps() {
  namespace.values = [...(props.namespaceLabels || [])]

  if (namespace.values.length > 0) {
    fetchModules().then(() => {
      module.values = [...(props.moduleLabels || [])]
    })
  }
}

function fetchNamespaces() {
  namespace.processing = true

  return $ComposeAPI.namespaceList(namespace.filter).then(({ set = [] } = {}) => {
    const namespacePromises = []

    if (props.namespaceLabels && props.namespaceLabels.length > 0 && !namespace.filter.query) {
      const namespaceIDs = props.namespaceLabels.map(label => label.split('/')[1]).filter(Boolean)

      namespaceIDs.forEach(namespaceID => {
        if (!set.some(n => n.namespaceID === namespaceID)) {
          namespacePromises.push(
            $ComposeAPI.namespaceRead({ namespaceID })
              .then(n => [n])
              .catch(() => []),
          )
        }
      })
    }

    return Promise.all(namespacePromises).then(results => {
      namespace.options = [...set, ...results.flat()].sort((a, b) =>
        (a.name || '').localeCompare(b.name || ''),
      )
    }).catch(() => {
      namespace.options = []
    })
  }).finally(() => {
    namespace.processing = false
  })
}

function fetchModules() {
  if (!namespace.values || namespace.values.length === 0) {
    module.options = []
    return Promise.resolve()
  }

  module.processing = true

  const namespaceIDs = namespace.values.map(label => label.split('/')[1])

  const promises = namespaceIDs.map(nsID =>
    $ComposeAPI.moduleList({
      namespaceID: nsID,
      ...module.filter,
    }).then(({ set }) => set),
  )

  return Promise.all(promises).then(results => {
    module.options = results.flat()
  }).catch(() => {
    module.options = []
  }).finally(() => {
    module.processing = false
  })
}

const searchNamespaces = debounce(function (query) {
  if (query !== namespace.filter.query) {
    namespace.filter.query = query
  }
  fetchNamespaces()
}, 300)

const searchModulesForNamespace = debounce(function (query, namespaceID) {
  if (query !== module.filter.query) {
    module.filter.query = query
  }
  fetchModules()
}, 300)

function updateNamespaces(namespaceLabels) {
  namespace.values = namespaceLabels || []

  if (namespace.values.length > 0) {
    const selectedNsIDs = new Set(namespace.values.map(label => label.split('/')[1]))
    module.values = module.values.filter(moduleLabel => {
      const nsID = moduleLabel.split('/')[1]
      return selectedNsIDs.has(nsID)
    })

    fetchModules()
  } else {
    module.options = []
    module.values = []
  }

  emitChange()
}

function updateModulesForNamespace(moduleLabels, namespaceID) {
  module.values = module.values.filter(label => {
    const nsID = label.split('/')[1]
    return nsID !== namespaceID
  })

  if (moduleLabels && moduleLabels.length > 0) {
    module.values = [...module.values, ...moduleLabels]
  }

  emitChange()
}

function emitChange() {
  emit('change', {
    namespaceLabels: [...namespace.values],
    moduleLabels: [...module.values],
  })
}

function getNamespaceOptionLabel({ name, handle } = {}) {
  return name || handle || 'Unnamed Namespace'
}

function getModuleOptionLabel(mod) {
  return mod.name || mod.handle || 'Unnamed Module'
}

function getModuleLabel(namespaceLabel) {
  const namespaceID = namespaceLabel.split('/')[1]
  const ns = namespace.options.find(n => n.namespaceID === namespaceID)
  const nsLabel = ns ? getNamespaceOptionLabel(ns) : namespaceID
  return t('filter.module.template', { namespace: nsLabel })
}

function getModulesForNamespace(namespaceID) {
  return modulesByNamespace.value[namespaceID] || []
}

function getModuleValuesForNamespace(namespaceID) {
  return module.values.filter(label => {
    const nsID = label.split('/')[1]
    return nsID === namespaceID
  })
}

function reset() {
  namespace.values = []
  module.values = []
  module.options = []
  emitChange()
}

defineExpose({ reset })
</script>

<style scoped>
.namespace-selector .vs__selected {
  background-color: var(--primary) !important;
  color: white !important;
}

.module-selector .vs__selected {
  background-color: var(--extra-light) !important;
  color: var(--dark) !important;
}
</style>
