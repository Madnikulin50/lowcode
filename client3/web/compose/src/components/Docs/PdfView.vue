<template>
  <div class="doc-pdf h-100">
    <div
      v-if="loading"
      class="d-flex align-items-center gap-2 text-muted p-3"
    >
      <span class="spinner-border spinner-border-sm" />
    </div>
    <div
      v-else-if="error"
      class="alert alert-warning m-3"
    >
      {{ error }}
    </div>
    <iframe
      v-else-if="src"
      class="doc-pdf-frame w-100 border-0"
      :src="src"
      :title="title"
    />
  </div>
</template>

<script setup>
import { inject, onBeforeUnmount, ref, watch } from 'vue'

const props = defineProps({
  namespaceID: { type: String, required: true },
  documentID: { type: String, required: true },
  title: { type: String, default: '' },
})

const $ComposeAPI = inject('$ComposeAPI', null) || window.__composeAPI
const src = ref('')
const loading = ref(false)
const error = ref('')

watch(() => [props.namespaceID, props.documentID], load, { immediate: true })

onBeforeUnmount(revoke)

async function load () {
  revoke()
  error.value = ''
  if (!props.namespaceID || !props.documentID) return
  loading.value = true
  try {
    const blob = await $ComposeAPI.documentFile({
      namespaceID: props.namespaceID,
      documentID: props.documentID,
    })
    if (blob?.type && blob.type.includes('json')) {
      error.value = 'PDF'
      return
    }
    src.value = URL.createObjectURL(blob)
  } catch (e) {
    error.value = e?.message || String(e)
  } finally {
    loading.value = false
  }
}

function revoke () {
  if (src.value) URL.revokeObjectURL(src.value)
  src.value = ''
}
</script>

<style scoped>
.doc-pdf-frame {
  min-height: 70vh;
  height: 100%;
}
</style>
