<template>
  <c-ace-editor
    v-model="template.template"
    data-test-id="template-html-editor"
    name="editor/html"
    lang="html"
    min-height="400px"
    show-line-numbers
    highlight-active-line
    show-print-margin
    :border="false"
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