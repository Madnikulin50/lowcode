<template>
  <div v-if="module">
    <h5>{{ $t('record-duplication.title') }}</h5>

    <div class="mb-3 my-4">
      <label class="form-label text-primary pb-0">{{ $t('record-duplication.strict-fields.label') }}</label>
      <small class="form-text d-block mb-2">{{ $t('record-duplication.strict-fields.description') }}</small>
      <field-picker
        :module="module"
        v-model:fields="strictFields"
        :field-subset="module.fields"
        disable-system-fields
        style="height: 35vh;"
        class="mt-3"
      />
    </div>

    <hr>

    <div class="mb-3 mt-4">
      <label class="form-label text-primary pb-0">{{ $t('record-duplication.non-strict-fields.label') }}</label>
      <small class="form-text d-block mb-2">{{ $t('record-duplication.non-strict-fields.description') }}</small>
      <field-picker
        :module="module"
        v-model:fields="nonStrictFields"
        :field-subset="module.fields"
        disable-system-fields
        style="height: 35vh;"
        class="mt-3"
      />
    </div>
  </div>
</template>

<script setup lang="js">
import { computed } from 'vue'
import { compose } from 'corteza-lib/js/dist'
import FieldPicker from 'corteza-webapp-compose/src/components/Common/FieldPicker'

defineOptions({
  i18nOptions: {
    namespaces: 'module',
    keyPrefix: 'edit.config.validation',
  },
})

const props = defineProps({
  module: {
    type: compose.Module,
    required: true,
  },
})

const strictFields = computed({
  get () {
    return getRuleFields(true).map(({ attributes }) => {
      return { name: attributes[0] }
    })
  },
  set (fields = []) {
    const fieldNames = fields.map(({ name }) => name)
    props.module.config.recordDeDup.rules = [
      ...getRuleFields(false).filter(({ attributes }) => !fieldNames.includes(attributes[0])),
      ...fieldNames.map(name => ({
        name: 'case-sensitive',
        strict: true,
        attributes: [name],
      })),
    ]
  },
})

const nonStrictFields = computed({
  get () {
    return getRuleFields(false).map(({ attributes }) => {
      return { name: attributes[0] }
    })
  },
  set (fields = []) {
    const fieldNames = fields.map(({ name }) => name)
    props.module.config.recordDeDup.rules = [
      ...getRuleFields(true).filter(({ attributes }) => !fieldNames.includes(attributes[0])),
      ...fieldNames.map(name => ({
        name: 'case-sensitive',
        strict: false,
        attributes: [name],
      })),
    ]
  },
})

function getRuleFields (strictValue) {
  return props.module.config.recordDeDup.rules.filter(({ name, strict, attributes = [] }) => {
    return strict === strictValue && name === 'case-sensitive' && attributes.length
  })
}
</script>
