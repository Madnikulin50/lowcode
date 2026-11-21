<template>
  <div class="mb-3">
    <label
      v-if="label"
      class="form-label text-primary"
    >
      {{ label }}
    </label>
    <div
      v-if="description"
      class="form-text mb-1"
    >
      {{ description }}
    </div>
    <div class="input-group">
      <c-rich-text-input
        v-if="rich"
        :model-value="modelValue"
        :placeholder="placeholder"
        body-class="form-control"
        min-body-height="8rem"
        output-format="markdown"
        :to-markdown="toMarkdown"
        :to-html="toHtml"
        @update:model-value="$emit('update:modelValue', $event)"
      />
      <textarea
        v-else
        :value="modelValue"
        class="form-control"
        :placeholder="placeholder"
        :rows="rows"
        @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
      />
      <slot name="append" />
    </div>
  </div>
</template>

<script setup lang="ts">
import CRichTextInput from '../input/CRichTextInput/index.vue'

defineProps({
  modelValue: { type: String, default: '' },
  label: { type: String, default: '' },
  description: { type: String, default: '' },
  placeholder: { type: String, default: '' },
  rows: { type: Number, default: 6 },
  rich: { type: Boolean, default: false },
  toMarkdown: { type: Function, default: undefined },
  toHtml: { type: Function, default: undefined },
})

defineEmits(['update:modelValue'])
</script>
