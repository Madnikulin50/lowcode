<template>
  <div v-if="module">
    <div class="mb-3">
      <div class="form-check">
        <input
          id="records-read-only"
          v-model="readOnly"
          class="form-check-input"
          type="checkbox"
        >
        <label
          class="form-check-label"
          for="records-read-only"
        >{{ t('read-only.label') }}</label>
      </div>
      <div class="form-text">
        {{ t('read-only.description') }}
      </div>
    </div>

    <div class="mb-3">
      <div class="form-check">
        <input
          id="records-no-create"
          v-model="noCreate"
          class="form-check-input"
          type="checkbox"
        >
        <label
          class="form-check-label"
          for="records-no-create"
        >{{ t('no-create.label') }}</label>
      </div>
      <div class="form-text">
        {{ t('no-create.description') }}
      </div>
    </div>

    <div class="mb-3">
      <div class="form-check">
        <input
          id="records-no-delete"
          v-model="noDelete"
          class="form-check-input"
          type="checkbox"
        >
        <label
          class="form-check-label"
          for="records-no-delete"
        >{{ t('no-delete.label') }}</label>
      </div>
      <div class="form-text">
        {{ t('no-delete.description') }}
      </div>
    </div>
  </div>
</template>

<script setup lang="js">
defineOptions({ i18nOptions: { namespaces: 'module', keyPrefix: 'edit.config.records' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const prefixed$ = 'edit.config.records.'
const { t: $t } = useI18n()
const t = (key) => $t(prefixed$ + key)

const props = defineProps({
  module: {
    type: Object,
    required: true,
  },
})

const flag = (name) => computed({
  get: () => !!(props.module.config && props.module.config.records && props.module.config.records[name]),
  set: (value) => {
    if (!props.module.config) props.module.config = {}
    if (!props.module.config.records) props.module.config.records = {}
    props.module.config.records[name] = value
  },
})

const readOnly = flag('readOnly')
const noCreate = flag('noCreate')
const noDelete = flag('noDelete')
</script>
