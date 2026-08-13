<template>
  <div>
    <div class="mb-3">
      <label class="form-label text-primary">{{ t('kind.bool.checkedValueLabel') }}</label>
      <div class="input-group input-group-sm">
        <input
          v-model="f.options.trueLabel"
          type="text"
          class="form-control form-control-sm"
          :placeholder="t('kind.bool.checkedValuePlaceholder')"
        />
        <FieldBoolTranslator
          v-if="field"
          :field="field"
          :module="module"
          highlight-key="meta.bool.true.label"
          size="sm"
          :disabled="isNew"
        />
      </div>
    </div>
    <div class="mb-3">
      <label class="form-label text-primary">{{ t('kind.bool.uncheckedValueLabel') }}</label>
      <div class="input-group input-group-sm">
        <input
          v-model="f.options.falseLabel"
          type="text"
          class="form-control form-control-sm"
          :placeholder="t('kind.bool.uncheckedValuePlaceholder')"
        />
        <FieldBoolTranslator
          v-if="field"
          :field="field"
          :module="module"
          highlight-key="meta.bool.fals.label"
          size="sm"
          :disabled="isNew"
        />
      </div>
    </div>
    <div class="mb-3">
      <label class="form-label text-primary">{{ t('kind.bool.toggleTypeLabel') }}</label>
      <div class="form-check">
        <input
          id="switch"
          v-model="f.options.switch"
          type="checkbox"
          class="form-check-input"
        />
        <label class="form-check-label" for="switch">{{ t('kind.bool.toggleType') }}</label>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { useI18n } from 'vue-i18n'
import { useConfiguratorBase } from './base'
import { NoID } from 'corteza-lib/js/dist'
import FieldBoolTranslator from 'corteza-webapp-compose/src/components/Admin/Module/FieldBoolTranslator'

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
