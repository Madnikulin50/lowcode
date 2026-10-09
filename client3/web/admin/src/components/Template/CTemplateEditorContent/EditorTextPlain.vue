<template>
  <c-ace-editor
    v-model="template.template"
    data-test-id="template-plain-text-editor"
    name="editor/plain_text"
    lang="plain_text"
    min-height="400px"
    show-line-numbers
    resizable
  />
</template>

<script setup>
import { components } from 'corteza-lib/vue/dist'
const { CAceEditor } = components

defineProps({
  template: { type: Object, required: true, default: () => ({}) },
})
</script>