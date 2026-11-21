<template>
  <div class="p-4">
    <h1 class="h3 mb-3">
      {{ $t('title') }}
    </h1>

    <div
      v-if="loading"
      class="text-muted"
    >
      <span class="spinner-border spinner-border-sm" />
    </div>
    <div
      v-else-if="error"
      class="alert alert-danger"
    >
      {{ error }}
    </div>
    <div
      v-else-if="!documents.length"
      class="text-muted"
    >
      {{ $t('empty') }}
    </div>
    <div
      v-else
      class="list-group"
    >
      <router-link
        v-for="doc in documents"
        :key="doc.documentID"
        class="list-group-item list-group-item-action d-flex align-items-center gap-2"
        :to="viewLink(doc)"
      >
        <font-awesome-icon
          :icon="docIcon(doc)"
          class="text-secondary"
        />
        <span>{{ doc.title }}</span>
      </router-link>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'document' } })
import { computed, inject, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

const props = defineProps({
  namespace: { type: Object, required: true },
})

const route = useRoute()
const $ComposeAPI = inject('$ComposeAPI', null) || window.__composeAPI
const documents = ref([])
const loading = ref(false)
const error = ref('')

const namespaceID = computed(() => props.namespace?.namespaceID)
const slug = computed(() => route.params.slug || props.namespace?.slug || namespaceID.value)

watch(namespaceID, load, { immediate: true })

function docIcon (doc) {
  return doc.kind === 'pdf' ? ['far', 'file-pdf'] : ['fas', 'file-lines']
}

function viewLink (doc) {
  return { name: 'namespace.document', params: { slug: slug.value, documentID: doc.documentID } }
}

async function load () {
  if (!namespaceID.value) return
  loading.value = true
  error.value = ''
  try {
    const { set } = await $ComposeAPI.documentList({ namespaceID: namespaceID.value })
    documents.value = (set || []).filter(doc => doc.visible)
  } catch (e) {
    error.value = e?.message || String(e)
    documents.value = []
  } finally {
    loading.value = false
  }
}
</script>
