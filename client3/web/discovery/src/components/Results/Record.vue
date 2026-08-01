<template>
  <div>
    <div class="card-header border-bottom">
      <div class="d-flex flex-wrap align-items-center justify-content-between gap-2">
        <h5 class="text-primary text-capitalize text-truncate mb-0">
          <span v-if="hit.value?.namespace?.name || hit.value?.namespace?.handle">
            {{ hit.value.namespace.name || hit.value.namespace.handle }}
          </span>
          <span v-if="hit.value?.namespace?.name || hit.value?.namespace?.handle" class="mx-2">
            <font-awesome-icon :icon="['fas', 'arrow-right']" />
          </span>
          <span>{{ hit.value?.module?.name || hit.value?.module?.handle }}</span>
        </h5>
        <span class="text-nowrap">
          <span class="rounded-circle bg-light text-dark d-inline-flex align-items-center justify-content-center" style="width: 24px; height: 24px;">
            <font-awesome-icon :icon="['fas', 'file-earmark-text']" style="font-size: 12px;" />
          </span>
          {{ t('types.record') }}
        </span>
      </div>
    </div>
    <div class="card-body d-flex flex-column flex-wrap gap-2">
      <div v-if="Object.keys(hit.value?.labels || {}).includes('federation')">
        <span class="badge text-white border border-secondary h6 mb-0">
          {{ t('federated') }}
        </span>
      </div>
      <div v-if="recordValues.length" class="d-flex flex-wrap gap-2 flex-grow-1">
        <div v-for="(item, i) in recordValues" :key="i" class="mb-0" style="min-width: 20rem; max-width: 100%; white-space: pre-line;">
          <label class="text-capitalize text-primary form-label">{{ item.label || item.name }}</label>
          <div>{{ item.value }}</div>
        </div>
      </div>
      <div v-if="systemValues.length" class="d-flex flex-wrap gap-2 flex-grow-1">
        <div v-for="item in systemValues" :key="item.name" class="mb-0" style="min-width: 20rem; max-width: 100%;">
          <label class="text-capitalize text-primary form-label">{{ item.label }}</label>
          <div>{{ item.value }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { filters } from 'corteza-lib/vue/dist'

const { t } = useI18n({
  useScope: 'local',
  messages: {},
})

const props = defineProps({
  index: {
    type: Number,
    required: true,
  },
  hit: {
    type: Object,
    required: true,
  },
  showMap: {
    type: Boolean,
    required: true,
  },
})

const defaultBlacklistedFields = ['deleted', 'created', 'updated', 'security', 'catch_all']

const createdBy = computed(() => {
  const { by } = props.hit.value?.created || {}
  return by
})

const createdAt = computed(() => {
  const { at } = props.hit.value?.created || {}
  return at ? filters.locFullDateTime(at) : at
})

const updatedAt = computed(() => {
  const { at } = props.hit.value?.updated || {}
  return at ? filters.locFullDateTime(at) : at
})

const blacklistedFields = computed(() => defaultBlacklistedFields)

function limitData() {
  const out = {}
  if (props.hit.value) {
    for (const key in props.hit.value) {
      const value = props.hit.value[key]
      if (!!value && blacklistedFields.value.indexOf(key) < 0) {
        out[key] = value
      }
    }
  }
  if (createdBy.value) out.createdBy = createdBy.value
  if (createdAt.value) out.createdAt = createdAt.value
  if (updatedAt.value) out.updatedAt = updatedAt.value
  return out
}

const recordID = computed(() => props.hit.value?.recordID)

const recordValues = computed(() => {
  const { values = [] } = props.hit.value || {}
  return values.map(({ name, label, value = [] }) => {
    if (value) {
      value = value.map(v => {
        return (v !== null ? v : '').toString().includes('{"coordinates":[')
          ? ((JSON.parse(v || '{}') || {}).coordinates || []).join(', ')
          : v
      }).join('\n')
    }
    return { name, label, value }
  })
})

const systemValues = computed(() => [
  { name: 'recordID', label: t('recordID'), value: recordID.value },
  { name: 'createdBy', label: t('createdBy'), value: createdBy.value },
  { name: 'createdAt', label: t('createdAt'), value: createdAt.value },
  { name: 'updatedAt', label: t('updatedAt'), value: updatedAt.value },
].filter(v => v.value))
</script>
