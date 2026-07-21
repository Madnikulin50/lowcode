<template>
  <div v-if="module">
    <div class="mb-3">
      <div class="form-check">
        <input
          id="record-revisions-enabled"
          v-model="module.config.recordRevisions.enabled"
          class="form-check-input"
          type="checkbox"
        >
        <label
          class="form-check-label"
          for="record-revisions-enabled"
        >{{ t('enabled') }}</label>
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('ident.label') }}</label>
      <div class="form-text mb-2">{{ identDescription }}</div>
      <input
        v-model="module.config.recordRevisions.ident"
        class="form-control form-control-sm"
        :disabled="!module.config.recordRevisions.enabled"
        :placeholder="t('ident.placeholder')"
      >
    </div>
  </div>
</template>

<script setup lang="js">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const prefixed$ = 'edit.config.record-revisions.'
const { t: $t } = useI18n()
const t = (key) => $t(prefixed$ + key)

defineProps({
  module: {
    type: Object,
    required: true,
  },
})

const identDescription = computed(() => t('ident.description', { interpolation: { prefix: '{{{', suffix: '}}}' } }))
</script>
