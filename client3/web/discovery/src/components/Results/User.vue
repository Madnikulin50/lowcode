<template>
  <div>
    <div class="card-header border-bottom">
      <div class="d-flex align-items-center justify-content-between">
        <h5 class="text-primary text-capitalize text-truncate me-2 mb-0">
          {{ t('types.user') }}
        </h5>
        <span class="rounded-circle bg-light text-dark d-inline-flex align-items-center justify-content-center" style="width: 24px; height: 24px;" title="User">
          <font-awesome-icon :icon="['fas', 'user']" style="font-size: 12px;" />
        </span>
      </div>
      <div class="d-flex justify-content-between small">
        <slot name="header" />
      </div>
    </div>
    <div class="card-body pb-0">
      <div v-for="(value, name, i) in limitData()" :key="i" class="mb-0" style="min-width: 200px; max-width: 100%;">
        <label class="text-capitalize text-primary form-label">{{ name }}</label>
        <div>{{ value }}</div>
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

const blacklistedFields = computed(() => [
  ...defaultBlacklistedFields,
  'meta',
  'fields',
  'namespace',
  'labels',
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
