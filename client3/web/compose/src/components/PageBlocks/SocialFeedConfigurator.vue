<template>
  <div>
    <h5>{{ $t('socialFeed.label') }}</h5>
    <div v-if="page.moduleID && page.moduleID !== '0'" class="mb-3">
      <label class="form-label text-primary">{{ $t('socialFeed.twitterProfileField') }}</label>
      <select v-model="options.profileSourceField" class="form-select form-control">
        <option v-for="opt in selectOptions" :key="opt" :value="opt">{{ opt }}</option>
      </select>
    </div>
    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('socialFeed.twitterProfileLabel') }}</label>
      <input v-model="options.profileUrl" class="form-control" />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  block: { type: Object, required: true },
  module: { type: Object, required: false },
  page: { type: Object, required: true },
})

const options = computed(() => props.block.options)

const selectOptions = computed(() => {
  if (!props.module) return []
  return props.module.fields.slice().sort((a, b) => a.name.localeCompare(b.name)).map(({ name }) => name)
})
</script>
