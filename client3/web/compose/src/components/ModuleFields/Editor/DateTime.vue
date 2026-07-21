<template>
  <div class="mb-3 date-time-field-editor" :class="formGroupStyleClasses">
     <div v-if="!valueOnly" class="d-flex align-items-center text-primary p-0">
        <span :title="label" class="d-inline-block mw-100">{{ label }}</span>
        <c-hint :tooltip="hint" />
        <slot name="tools" />
      </div>
      <div class="small text-muted" :class="{ 'mb-1': description }">{{ description }}</div>

    <multi v-if="field.isMulti" v-slot="ctx" v-model:value="value" :errors="errors">
      <CInputDateTime
        :value="value[ctx.index]"
        :no-date="field.options.onlyTime"
        :no-time="field.options.onlyDate"
        :only-future="field.options.onlyFutureValues"
        :only-past="field.options.onlyPastValues"
        :labels="dtLabels"
        @input="setMultiValue($event, ctx.index)"
      />
    </multi>

    <template v-else>
      <CInputDateTime
        v-model="value"
        :no-date="field.options.onlyTime"
        :no-time="field.options.onlyDate"
        :only-future="field.options.onlyFutureValues"
        :only-past="field.options.onlyPastValues"
        :labels="dtLabels"
      />
      <errors :errors="errors" />
    </template>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditorBase } from './base'
import { components } from 'corteza-lib/vue/dist'
import errors from '../errors'
import multi from './multi'
const { CInputDateTime } = components

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
const { value, formGroupStyleClasses, label, hint, description, setMultiValue } = useEditorBase(props, emit)

const dtLabels = computed(() => ({
  clear: t('label.clear'),
  none: t('label.none'),
  now: t('label.now'),
  today: t('label.today'),
}))
</script>

<style>
.date-time-field-editor .form-row div {
  position: initial;
}
</style>
