<template>
  <div class="card border">
    <div class="card-body">
      <h5 class="card-title">{{ sourceLabel }}</h5>
      <p v-if="workflow?.meta" class="mb-3">
        {{ workflow.meta.description || $t('noDescription') }}
        <var v-if="trigger">{{ $t('stepID', { stepID: trigger.stepID }) }}</var>
      </p>
      <p v-else-if="script" class="mb-3">{{ script.description || $t('noDescription') }}</p>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('buttonLabel') }}</label>
        <c-input-expression v-model="button.label" auto-complete :suggestion-params="recordAutoCompleteParams" :page="page" />
        <small class="text-muted d-block mt-1">
          <code>${record.values.fieldName}</code>
          <code>${recordID}</code>
          <code>${ownerID}</code>
          <code>${userID}</code>, <code>${user.name}</code>
        </small>
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('buttonVariant') }}</label>
        <select v-model="button.variant" class="form-select form-control w-100">
          <option v-for="({ variant, label }) in variants" :key="variant" :value="variant">{{ label }}</option>
        </select>
      </div>
    </div>
    <div class="card-footer text-end pt-0">
      <c-input-confirm show-icon @confirmed="$emit('delete', button)" />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block', keyPrefix: 'automation' } })
import { computed, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose, NoID } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'

const { CInputExpression } = components
const { t: $t } = useI18n({ useScope: 'global', messages: {}, keyPrefix: 'block:automation' })

const props = defineProps({
  button: { type: Object, required: true },
  script: { type: Object, required: false, default: undefined },
  trigger: { type: Object, required: false, default: undefined },
  page: { type: compose.Page, required: true },
  block: { type: compose.PageBlock, required: true },
  module: { type: compose.Module, required: false, default: undefined },
  record: { type: compose.Record, required: false, default: undefined },
})

defineEmits(['delete'])

const $auth = inject('$auth', undefined)

const recordAutoCompleteParams = computed(() => processRecordAutoCompleteParams({ module: props.module, operators: true }))

function processRecordAutoCompleteParams ({ module: mod, operators = false } = {}) {
  const { fields = [] } = mod || {}
  const moduleFields = fields.map(({ name }) => name)
  const userProperties = ($auth?.user?.properties?.()) || []

  const recordSuggestions = props.record
    ? [
        ...(['ownerID', 'recordID'].map(value => ({ interpolate: true, value }))),
        {
          interpolate: true,
          value: 'record',
          properties: [
            { value: 'values', properties: Object.keys(props.record.values) || [] },
            ...(props.record.properties || []),
          ],
        },
      ]
    : []

  return [
    ...recordSuggestions,
    ...(operators ? ['AND', 'OR'] : []),
    { interpolate: true, value: 'userID' },
    { interpolate: true, value: 'user', properties: userProperties },
    ...moduleFields,
  ]
}

const variants = computed(() => ['primary', 'secondary', 'light', 'dark', 'success', 'danger', 'warning'].map(variant => ({ variant, label: $t(`variants.${variant}`) })))

const sourceLabel = computed(() => {
  const w = workflow.value
  if (w) return w.meta?.name || $t('noLabel')
  if (props.button.script) return props.button.script
  return $t('dummyButtonLabel')
})

const workflow = computed(() => props.trigger?.workflow)
const isNew = computed(() => props.block.blockID === NoID)
</script>
