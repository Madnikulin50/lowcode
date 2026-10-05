<template>
  <div class="d-flex flex-column w-100 h-100 flex-grow-1" style="min-width: 0">
    <Teleport
      defer
      to="#topbar-title"
    >
      {{ document?.title || $t('title') }}
    </Teleport>

    <Teleport
      defer
      to="#topbar-tools"
    >
      <router-link
        class="btn btn-outline-secondary btn-sm"
        :to="{ name: 'namespace.documents', params: { slug } }"
      >
        {{ $t('toc') }}
      </router-link>
      <router-link
        v-if="prev"
        class="btn btn-outline-secondary btn-sm"
        :to="viewLink(prev)"
      >
        {{ $t('previous') }}
      </router-link>
      <router-link
        v-if="next"
        class="btn btn-outline-secondary btn-sm"
        :to="viewLink(next)"
      >
        {{ $t('next') }}
      </router-link>
    </Teleport>

    <div
      v-if="loading"
      class="p-4 text-muted"
    >
      <span class="spinner-border spinner-border-sm" />
    </div>
    <div
      v-else-if="error"
      class="alert alert-danger m-4"
    >
      {{ error }}
    </div>
    <div
      v-else-if="document?.kind === 'markdown'"
      class="doc-view p-4 overflow-auto w-100 flex-grow-1"
    >
      <markdown-view class="w-100" :body="document.body" />
    </div>
    <pdf-view
      v-else-if="document?.kind === 'pdf' && document.attachmentID && document.attachmentID !== '0'"
      class="flex-grow-1"
      :namespaceID="namespaceID"
      :documentID="document.documentID"
      :title="document.title"
    />
    <div
      v-else
      class="p-4 text-muted"
    >
      {{ $t('noFile') }}
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'document' } })
import { computed, inject, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import MarkdownView from '../../../components/Docs/MarkdownView.vue'
import PdfView from '../../../components/Docs/PdfView.vue'

const props = defineProps({
  namespace: { type: Object, required: true },
})

const route = useRoute()
const $ComposeAPI = inject('$ComposeAPI', null) || window.__composeAPI
const document = ref(null)
const documents = ref([])
const loading = ref(false)
const error = ref('')

const namespaceID = computed(() => props.namespace?.namespaceID)
const slug = computed(() => route.params.slug || props.namespace?.slug || namespaceID.value)
const index = computed(() => documents.value.findIndex(doc => String(doc.documentID) === String(route.params.documentID)))
const prev = computed(() => (index.value > 0 ? documents.value[index.value - 1] : null))
const next = computed(() => (index.value >= 0 && index.value < documents.value.length - 1 ? documents.value[index.value + 1] : null))

watch(() => [namespaceID.value, route.params.documentID], load, { immediate: true })

function viewLink (doc) {
  return { name: 'namespace.document', params: { slug: slug.value, documentID: doc.documentID } }
}

async function load () {
  if (!namespaceID.value || !route.params.documentID) return
  loading.value = true
  error.value = ''
  try {
    const [res, list] = await Promise.all([
      $ComposeAPI.documentRead({
        namespaceID: namespaceID.value,
        documentID: route.params.documentID,
      }),
      $ComposeAPI.documentList({ namespaceID: namespaceID.value }),
    ])
    document.value = res.document
    documents.value = (list.set || []).filter(doc => doc.visible)
  } catch (e) {
    document.value = null
    error.value = e?.message || String(e)
  } finally {
    loading.value = false
  }
}
</script>
