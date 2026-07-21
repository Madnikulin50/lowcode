<template>
  <div class="tab-pane">
    <div
      v-if="enableFromRecordURL"
      class="mb-3"
    >
      <label class="form-label text-primary">{{ $t('iframe.srcFieldLabel') }}</label>
      <small class="form-text">{{ $t('iframe.srcFieldDesc') }}</small>
      <select
        v-model="options.srcField"
        class="form-select"
        :disabled="!fields.length"
      >
        <option value="">{{ $t('iframe.pickURLField') }}</option>
        <option
          v-for="f in fieldOptions"
          :key="f.value"
          :value="f.value"
        >
          {{ f.text }}
        </option>
      </select>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ $t('iframe.srcLabel') }}</label>
      <c-input-expression
        v-model="options.src"
        auto-complete
        lang="text"
        :suggestion-params="recordAutoCompleteParams"
        type="url"
      />
      <i18next
        path="interpolationFootnote"
        tag="small"
        class="text-muted"
      >
        <code>${record.values.fieldName}</code>
        <code>${recordID}</code>
        <code>${ownerID}</code>
        <span><code>${userID}</code>, <code>${user.name}</code></span>
      </i18next>
    </div>

    <div class="mb-3">
      <div class="form-check">
        <input
          v-model="options.displayAsImage"
          type="checkbox"
          class="form-check-input"
        />
        <label class="form-check-label">{{ $t('iframe.displayAsImageLabel') }}</label>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { usePageBlockBase } from './usePageBlockBase'
import { components } from 'corteza-lib/vue/dist'
import { NoID } from 'corteza-lib/js/dist'

const { CInputExpression } = components

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, required: false, default: '' },
  editable: { type: Boolean, required: false, default: false },
  resizing: { type: Boolean, required: false, default: false },
  magnified: { type: Boolean, required: false, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, required: false, default: false },
  errors: { type: Object, required: false, default: () => ({}) },
})

const emit = defineEmits(['errors'])

const { options } = usePageBlockBase(props, emit)

const fields = computed(() => {
  if (!props.module) return []
  return props.module.fields
    .filter(({ kind }) => kind === 'Url')
    .map(({ name, label }) => ({ value: name, text: label }))
    .sort((a, b) => a.text.localeCompare(b.text))
})

const fieldOptions = computed(() => [
  { value: '', text: 'Pick URL field' },
  ...fields.value,
])

const enableFromRecordURL = computed(() => props.page.moduleID !== '0')
</script>
