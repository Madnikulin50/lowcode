<template>
  <div class="mb-3" :data-test-id="getFieldCypressId(label)" :class="formGroupStyleClasses">
    <div v-if="!valueOnly" class="d-flex align-items-center text-primary p-0">
      <span :title="label" class="d-inline-block mw-100">{{ label }}</span>
      <c-hint :tooltip="hint" />
      <slot name="tools" />
    </div>
    <div class="small text-muted" :class="{ 'mb-1': description }">{{ description }}</div>

    <multi v-if="field.isMulti" v-slot="ctx" v-model:value="value" :errors="errors">
      <div class="input-group input-group-sm">
        <span v-if="field.options.prefix" class="input-group-text">{{ field.options.prefix }}</span>
        <input
          :value="value[ctx.index]"
          type="number"
          class="form-control form-control-sm"
          autocomplete="off"
          :step="field.options.step"
          @input="setMultiValue($event.target.value, ctx.index)"
        />
        <span v-if="field.options.suffix" class="input-group-text">{{ field.options.suffix }}</span>
      </div>
    </multi>

    <template v-else>
      <div class="input-group input-group-sm">
        <span v-if="field.options.prefix" class="input-group-text">{{ field.options.prefix }}</span>
        <input v-model="value" type="number" class="form-control form-control-sm" autocomplete="off" :step="field.options.step" />
        <span v-if="field.options.suffix" class="input-group-text">{{ field.options.suffix }}</span>
      </div>
      <errors :errors="errors" />
    </template>
  </div>
</template>

<script setup>
import { useEditorBase } from './base'
import errors from '../errors'
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
const { value, formGroupStyleClasses, label, hint, description, getFieldCypressId, setMultiValue } = useEditorBase(props, emit)
</script>
