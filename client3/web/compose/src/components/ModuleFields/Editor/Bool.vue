<template>
  <div class="mb-3" :class="formGroupStyleClasses">
    <template v-if="field.options.switch">
      <div v-if="!valueOnly" class="d-flex align-items-center text-primary p-0">
        <span :title="label" class="d-inline-block mw-100 pt-0" :class="{ 'py-1': !horizontal }">{{ label }}</span>
        <c-hint :tooltip="hint" />
        <slot name="tools" />
      </div>
      <div class="small text-muted" :class="{ 'mb-1': description }">{{ description }}</div>
    </template>

    <CInputCheckbox
      v-model="value"
      :switch="field.options.switch"
      :labels="field.options.switch ? checkboxLabel : {}"
    >
      <div v-if="!field.options.switch" class="d-flex align-items-center text-primary">
        {{ label }}
        <c-hint :tooltip="hint" />
      </div>
    </CInputCheckbox>

    <div v-if="!valueOnly && !field.options.switch" class="small text-muted">{{ description }}</div>
    <errors :errors="errors" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import errors from '../errors'
import { useEditorBase } from './base'

const props = defineProps({
  namespace: { type: Object, required: true },
  field: { type: Object, required: true },
  record: { type: Object, required: true },
  errors: { type: Object, required: true },
  valueOnly: { type: Boolean, default: false },
  horizontal: { type: Boolean, default: false },
  extraOptions: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['change', 'update:preventPopoverClose'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { formGroupStyleClasses, label, hint, description } = useEditorBase(props, emit)

const value = computed({
  get () { return props.record.values[props.field.name] === '1' },
  set (val) {
    props.record.values[props.field.name] = val ? '1' : '0'
    emit('change', val)
  },
})

const checkboxLabel = computed(() => ({
  on: props.field.options.trueLabel || t('label.yes'),
  off: props.field.options.falseLabel || t('label.no'),
}))
</script>
