<template>
  <div>
    <div class="form-check mb-2">
      <input
        id="multiLine"
        v-model="f.options.multiLine"
        type="checkbox"
        class="form-check-input"
      />
      <label class="form-check-label" for="multiLine">{{ t('kind.string.multiLine') }}</label>
    </div>
    <div class="form-check mb-2">
      <input
        id="useRichTextEditor"
        v-model="f.options.useRichTextEditor"
        type="checkbox"
        class="form-check-input"
      />
      <label class="form-check-label" for="useRichTextEditor">{{ t('kind.string.richText') }}</label>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { useI18n } from 'vue-i18n'
import { useConfiguratorBase } from './base'
import { NoID } from 'corteza-lib/js/dist'

const props = defineProps({
  namespace: { type: Object, required: true },
  module: { type: Object, required: true },
  field: { type: Object, required: true },
  hasRecords: { type: Boolean, default: false },
})

const emit = defineEmits(['update:field'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { f, isNew, hasData } = useConfiguratorBase(props, emit)
</script>
