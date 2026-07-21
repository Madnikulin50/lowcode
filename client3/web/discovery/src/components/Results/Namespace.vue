<template>
  <div>
    <div class="card-header border-bottom">
      <div class="d-flex flex-wrap align-items-center justify-content-between gap-2">
        <h5 class="text-primary text-capitalize text-truncate mb-0">
          {{ hit.value?.name || hit.value?.slug }}
        </h5>
        <span class="text-nowrap">
          <span class="rounded-circle bg-light text-dark d-inline-flex align-items-center justify-content-center" style="width: 24px; height: 24px;">
            <font-awesome-icon :icon="['fas', 'code-square']" style="font-size: 12px;" />
          </span>
          {{ t('types.namespace') }}
        </span>
      </div>
      <div class="d-flex justify-content-between small">
        <slot name="header" />
      </div>
    </div>
    <div class="card-body">
      <div class="d-flex flex-wrap" style="gap: 2rem;">
        <div v-for="(item, name, i) in limitData()" :key="i" class="mb-0" style="min-width: 20rem; max-width: 100%;">
          <label class="text-capitalize text-primary form-label">{{ name }}</label>
          <div>{{ item }}</div>
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

const blacklistedFields = computed(() => [
  ...defaultBlacklistedFields,
  'meta',
  'namespace',
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
