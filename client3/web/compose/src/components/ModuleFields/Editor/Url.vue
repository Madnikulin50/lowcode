<template>
  <div class="mb-3" :class="formGroupStyleClasses">
    <div v-if="!valueOnly" :class="labelColClass">
      <div class="d-flex align-items-center text-primary px-0">
        <span :title="label" class="d-inline-block mw-100">{{ label }}</span>
        <c-hint :tooltip="hint" />
        <slot name="tools" />
      </div>
      <div class="small text-muted" :class="{ 'mb-1': description }">{{ description }}</div>
    </div>
    <div :class="contentColClass">
    <multi v-if="field.isMulti" v-slot="ctx" v-model:value="value" :errors="errors">
      <input
        :value="value[ctx.index]"
        type="url"
        class="form-control form-control-sm"
        :placeholder="t('kind.url.example')"
        @input="setMultiValue(fixUrl($event.target.value), ctx.index)"
      />
    </multi>

    <template v-else>
      <input
        v-model="value"
        type="url"
        class="form-control form-control-sm"
        :placeholder="t('kind.url.example')"
      />
      <FieldErrors :errors="errors" />
    </template>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { useI18n } from 'vue-i18n'
import { useEditorBase } from './base'
import { trimUrlFragment, trimUrlQuery, trimUrlPath, onlySecureUrl } from '../url'
import FieldErrors from '../errors'
import multi from './multi'

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
const { value, formGroupStyleClasses, labelColClass, contentColClass, label, hint, description, setMultiValue } = useEditorBase(props, emit)

function fixUrl (value) {
  if (props.field.options.trimFragment) value = trimUrlFragment(value)
  if (props.field.options.trimQuery) value = trimUrlQuery(value)
  if (props.field.options.trimPath) value = trimUrlPath(value)
  if (props.field.options.onlySecure) value = onlySecureUrl(value)
  return value
}
</script>
