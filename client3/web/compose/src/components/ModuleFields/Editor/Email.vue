<template>
  <div class="mb-3" :class="formGroupStyleClasses">
    <div v-if="!valueOnly" :class="labelColClass">
      <div class="d-flex align-items-center text-primary p-0">
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
        type="email"
        class="form-control form-control-sm"
        @input="setMultiValue($event.target.value, ctx.index)"
      />
    </multi>

    <template v-else>
      <input v-model="value" type="email" class="form-control form-control-sm" />
      <FieldErrors :errors="errors" />
    </template>
    </div>
  </div>
</template>

<script setup>
import { useEditorBase } from './base'
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
const { value, formGroupStyleClasses, labelColClass, contentColClass, label, hint, description, setMultiValue } = useEditorBase(props, emit)
</script>
