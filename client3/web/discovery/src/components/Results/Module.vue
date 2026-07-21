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
          <span>{{ hit.value?.name || hit.value?.handle }}</span>
        </h5>
        <span class="text-nowrap">
          <span v-if="Object.keys(hit.value?.labels || {}).includes('federation')" class="badge border border-secondary me-1 h5 p-2 mb-0">
            {{ t('federated') }}
          </span>
          <span class="rounded-circle bg-light text-dark d-inline-flex align-items-center justify-content-center" style="width: 24px; height: 24px;">
            <font-awesome-icon :icon="['fas', 'list-ul']" style="font-size: 12px;" />
          </span>
          {{ t('types.module') }}
        </span>
      </div>
      <div class="d-flex justify-content-between small">
        <slot name="header" />
      </div>
    </div>
    <div class="card-body d-flex flex-wrap" style="gap: 2rem;">
      <div v-for="(value, name, i) in limitData()" :key="i" class="mb-0" style="min-width: 20rem; max-width: 100%;">
        <label class="text-capitalize text-primary form-label">{{ name }}</label>
        <div>{{ value }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

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

import { filters } from 'corteza-lib/vue/dist'

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

const blacklistedFields = computed(() => [
  ...defaultBlacklistedFields,
  'meta',
  'fields',
  'namespace',
  'labels',
  'module',
])

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
</script>
