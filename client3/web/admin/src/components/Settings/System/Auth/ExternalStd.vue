<template>
  <div>
    <div class="mb-3">
      <div class="form-check">
        <input
          id="std-enabled"
          v-model="value.enabled"
          class="form-check-input-v3"
          type="checkbox"
        >
        <label
          class="form-check-label"
          for="std-enabled"
        >{{ t('enabled') }}</label>
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('clientKey') }}</label>
      <div class="input-group">
        <input
          v-model.trim="value.key"
          class="form-control"
          :required="value.enabled"
        >
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('clientSecret') }}</label>
      <div class="input-group">
        <input
          v-model.trim="value.secret"
          class="form-control"
          :required="value.enabled"
        >
      </div>
    </div>

    <security
      v-model="value.security"
    />
  </div>
</template>

<script setup>
import { useNsI18n } from 'corteza-lib/vue/dist'
defineOptions({ i18nOptions: { namespaces: 'system.settings', keyPrefix: 'editor.external.standard' } })

import Security from './ExternalSecurity'

const t = useNsI18n()

const props = defineProps({
  value: {
    type: Object,
    required: true,
    default: () => ({}),
  },
})
</script>
