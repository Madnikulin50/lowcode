<template>
  <div
    class="doc-markdown"
    v-html="html"
  />
</template>

<script setup>
import { computed } from 'vue'
import { markdownToSafeHtml } from '../../lib/markdown'

const props = defineProps({
  body: { type: String, default: '' },
})

const html = computed(() => markdownToSafeHtml(props.body))
</script>

<style scoped>
.doc-markdown {
  width: 100%;
  max-width: none;
}

.doc-markdown :deep(h1),
.doc-markdown :deep(h2),
.doc-markdown :deep(h3) {
  margin-top: 1.25rem;
  margin-bottom: 0.5rem;
}

.doc-markdown :deep(p),
.doc-markdown :deep(ul),
.doc-markdown :deep(ol) {
  margin-bottom: 0.75rem;
}

.doc-markdown :deep(pre) {
  padding: 0.75rem 1rem;
  background: var(--bs-light, #f8f9fc);
  border-radius: 0.375rem;
  overflow: auto;
}

.doc-markdown :deep(img) {
  max-width: 100%;
}

.doc-markdown :deep(table) {
  width: 100%;
  margin-bottom: 1rem;
}

.doc-markdown :deep(a) {
  word-break: break-word;
}
</style>
