<template>
  <div>
    <div v-for="agg in aggregationOptions" :key="agg.resource" class="my-2">
      <div v-if="agg.items.length" class="d-flex justify-content-between align-items-center" style="min-height: 25px;">
        <h6 class="text-primary d-flex mb-0">
          {{ agg.name }}
          <span v-if="groups[agg.name].length" class="badge border border-secondary ms-1 align-self-center">
            {{ groups[agg.name].length }}
          </span>
        </h6>
        <button v-if="groups[agg.name].length" class="btn btn-link text-muted p-0 m-0" size="sm" @click="clearGroup(agg.name)">
          {{ t('reset') }}
        </button>
      </div>
      <div class="mt-1 ms-2">
        <div v-for="(resource, i) in agg.items" :key="i" class="form-check mb-1">
          <input
            class="form-check-input"
            type="checkbox"
            :value="resource.name"
            :id="`filter-${agg.name}-${i}`"
            :checked="groups[agg.name].includes(resource.name)"
            :disabled="storeProcessing"
            @change="toggleGroupItem(agg.name, resource.name)"
          />
          <label class="form-check-label" :for="`filter-${agg.name}-${i}`">
            <div class="d-flex align-items-center">
              <span class="d-inline-block text-truncate">
                {{ getResourceDisplayName(agg.resource, resource) }}
              </span>
            </div>
          </label>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDiscoveryStore } from '../store'

const { t } = useI18n({
  useScope: 'local',
  messages: {},
})

const store = useDiscoveryStore()

const groups = ref({
  Module: [],
  Namespace: [],
})

const storeProcessing = computed(() => store.processing)
const storeResourceTypes = computed(() => store.resourceTypes)
const storeAggregations = computed(() => store.aggregations)
const storeModules = computed(() => store.modules)
const storeNamespaces = computed(() => store.namespaces)

const options = computed(() => [
  { text: t('types.namespace'), value: 'compose:namespace', icon: 'code-square' },
  { text: t('types.module'), value: 'compose:module', icon: 'list-ul' },
  { text: t('types.record'), value: 'compose:record', icon: 'file-earmark-text' },
])

const aggregationOptions = computed(() => {
  let namespaceOptions = storeAggregations.value.find(({ resource }) => resource === 'compose:namespace') || {}
  let moduleOptions = storeAggregations.value.find(({ resource }) => resource === 'compose:module') || {}

  namespaceOptions = {
    resource: 'compose:namespace',
    name: 'Namespace',
    hits: namespaceOptions.hits || 0,
    items: namespaceOptions.resource_name || [],
  }

  const missingModuleOptions = groups.value.Module.filter(name => !(moduleOptions.resource_name || []).some(o => o.name === name))
    .map(name => ({ name }))

  moduleOptions = {
    resource: 'compose:module',
    name: 'Module',
    hits: moduleOptions.hits || 0,
    items: [
      ...missingModuleOptions,
      ...(moduleOptions.resource_name || []),
    ],
  }

  return [namespaceOptions, moduleOptions]
})

watch(storeNamespaces, (namespace) => {
  groups.value.Namespace = namespace
}, { immediate: true })

watch(storeModules, (module) => {
  groups.value.Module = module
}, { immediate: true })

function getResourceDisplayName(type, { name, handle, slug }) {
  if (type === 'compose:namespace') {
    return name || slug || 'Unnamed namespace'
  } else if (type === 'compose:module') {
    return handle || name || 'Unnamed module'
  }
}

function toggleGroupItem(groupName, itemName) {
  const idx = groups.value[groupName].indexOf(itemName)
  if (idx >= 0) {
    groups.value[groupName].splice(idx, 1)
  } else {
    groups.value[groupName].push(itemName)
  }
  updateGroup(groupName)
}

function updateGroup(name) {
  store[`update${name}s`](groups.value[name])
}

function clearGroup(name) {
  groups.value[name] = []
  updateGroup(name)
}
</script>
