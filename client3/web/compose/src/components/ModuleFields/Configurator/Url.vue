<template>
  <div>
    <div class="form-check mb-2">
      <input id="trimPath" v-model="f.options.trimPath" type="checkbox" class="form-check-input" />
      <label class="form-check-label" for="trimPath">{{ t('kind.url.label') }}</label>
    </div>
    <div class="form-check mb-2">
      <input id="trimFragment" v-model="f.options.trimFragment" type="checkbox" class="form-check-input" />
      <label class="form-check-label" for="trimFragment">{{ t('kind.url.trimHash') }}</label>
    </div>
    <div class="form-check mb-2">
      <input id="trimQuery" v-model="f.options.trimQuery" type="checkbox" class="form-check-input" />
      <label class="form-check-label" for="trimQuery">{{ t('kind.url.trimQuestionMark') }}</label>
    </div>
    <div class="form-check mb-2">
      <input id="onlySecure" v-model="f.options.onlySecure" type="checkbox" class="form-check-input" />
      <label class="form-check-label" for="onlySecure">{{ t('kind.url.sshOnly') }}</label>
    </div>
    <div class="form-check mb-2">
      <input id="outputPlain" v-model="f.options.outputPlain" type="checkbox" class="form-check-input" />
      <label class="form-check-label" for="outputPlain">{{ t('kind.url.preventToLink') }}</label>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { useI18n } from 'vue-i18n'
import { useConfiguratorBase } from './base'

const props = defineProps({
  namespace: { type: Object, required: true },
  module: { type: Object, required: true },
  field: { type: Object, required: true },
  hasRecords: { type: Boolean, default: false },
})

const emit = defineEmits(['update:field'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { f } = useConfiguratorBase(props, emit)
</script>
