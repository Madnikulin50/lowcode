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
    <template v-if="field.isMulti">
      <div v-if="field.options.selectType === 'list'">
        <div v-for="option in selectOptions" :key="option.value" class="form-check d-block mb-1">
          <input
            :id="'opt-' + option.value"
            type="checkbox"
            class="form-check-input"
            :value="option.value"
            :checked="value && value.includes(option.value)"
            @change="toggleCheckboxValue(option.value)"
          />
          <label class="form-check-label" :for="'opt-' + option.value">
            <span :class="{ 'badge rounded-pill': field.options.displayType === 'badge' }" :style="getOptionStyle(option.value)">{{ option.text }}</span>
          </label>
        </div>
        <FieldErrors :errors="errors" />
      </div>

      <multi v-else v-model:value="value" :errors="errors" :single-input="field.options.selectType !== 'each'">
        <template #single>
          <c-input-select
            v-if="field.options.selectType === 'default'"
            ref="singleSelect"
            :options="selectOptions"
            :placeholder="t('kind.select.placeholder')"
            :reduce="o => o.value"
            :selectable="isSelectable"
            label="text"
            :badge="field.options.displayType === 'badge'"
            @input="selectChange"
          />
          <c-input-select
            v-if="field.options.selectType === 'multiple'"
            v-model="value"
            :options="selectOptions"
            :placeholder="t('kind.select.placeholder')"
            :reduce="o => o.value"
            :selectable="isSelectable"
            label="text"
            multiple
            :badge="field.options.displayType === 'badge'"
          />
        </template>
        <template #default="ctx">
          <c-input-select
            v-if="field.options.selectType === 'each'"
            :value="value[ctx.index]"
            :options="selectOptions"
            :reduce="o => o.value"
            :placeholder="t('kind.select.placeholder')"
            :selectable="isSelectable"
            label="text"
            :badge="field.options.displayType === 'badge'"
            @input="setMultiValue($event, ctx.index)"
          />
          <span v-else :class="{ 'badge rounded-pill': field.options.displayType === 'badge' }" :style="getOptionStyle(value[ctx.index])">{{ findLabel(value[ctx.index]) }}</span>
        </template>
      </multi>
    </template>

    <template v-else>
      <div v-if="field.options.selectType === 'list'" class="btn-group" data-bs-toggle="buttons">
        <label
          v-for="opt in selectOptions"
          :key="opt.value"
          class="btn btn-outline-primary btn-sm"
          :class="{ active: value === opt.value }"
        >
          <input
            type="radio"
            class="btn-check"
            :value="opt.value"
            :checked="value === opt.value"
            autocomplete="off"
            @change="value = opt.value"
          />
          {{ opt.text }}
        </label>
      </div>
      <c-input-select
        v-else
        v-model="value"
        :placeholder="t('kind.select.optionNotSelected')"
        :options="selectOptions"
        :reduce="o => o.value"
        :selectable="isSelectable"
        label="text"
        :badge="field.options.displayType === 'badge'"
      />
      <FieldErrors :errors="errors" />
    </template>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditorBase } from './base'
import FieldErrors from '../errors'
import multi from './multi'
import { badgeGradient } from 'corteza-webapp-compose/src/lib/color.js'

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
const { value, formGroupStyleClasses, labelColClass, contentColClass, label, hint, description, setMultiValue, getColor } = useEditorBase(props, emit)

const singleSelect = ref(null)

const selectOptions = computed(() => {
  return (props.field.options.options || []).filter(({ value: v = '', text = '' }) => v && text)
})

function selectChange (val) {
  const arr = Array.isArray(value.value) ? [...value.value] : []
  arr.push(val)
  value.value = arr
  if (singleSelect.value) {
    singleSelect.value.localValue = undefined
  }
}

function findLabel (v) {
  return (selectOptions.value.find(({ value: val }) => val === v) || {}).text || v
}

function isSelectable ({ value: val } = {}) {
  if (props.field.options.selectType === 'list') return true
  if (props.field.isMulti) {
    return !props.field.options.isUniqueMultiValue || !((value.value) || []).includes(val)
  }
  return value.value !== val
}

function getOptionStyle (v) {
  const style = {}
  if (props.field.options.displayType === 'badge') {
    const opt = selectOptions.value.find(({ value: val }) => val === v) || { style: {} }
    style.fontSize = '0.9rem'
    const fg = getColor(opt.style.textColor) || 'var(--dark)'
    const bg = getColor(opt.style.backgroundColor) || 'var(--extra-light)'
    style.color = fg
    const gradient = props.field.options.badgeGradient ? badgeGradient(bg) : undefined
    if (gradient) {
      style.background = gradient
    } else {
      style.backgroundColor = bg
    }
  }
  return style
}

function toggleCheckboxValue (val) {
  const arr = Array.isArray(value.value) ? [...value.value] : []
  const idx = arr.indexOf(val)
  if (idx >= 0) arr.splice(idx, 1)
  else arr.push(val)
  value.value = arr
}
</script>
