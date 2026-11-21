<template>
  <div class="container-fluid d-flex flex-column py-3">
    <Teleport to="#topbar-title">
      {{ isCreate ? $t(kind === 'pdf' ? 'newPdf' : 'newMarkdown') : form.title || $t('edit') }}
    </Teleport>

    <h1 class="h3 mb-3">
      {{ isCreate ? $t(kind === 'pdf' ? 'newPdf' : 'newMarkdown') : $t('edit') }}
    </h1>

    <div
      v-if="error"
      class="alert alert-danger"
    >
      {{ error }}
    </div>

    <form @submit.prevent="save">
      <div class="mb-3">
        <label class="form-label">{{ $t('field.title') }}</label>
        <input
          v-model="form.title"
          class="form-control"
          required
        >
      </div>

      <div class="row">
        <div class="col-md-8 mb-3">
          <label class="form-label">{{ $t('field.handle') }}</label>
          <input
            v-model="form.handle"
            class="form-control"
          >
        </div>
        <div class="col-md-4 mb-3 d-flex align-items-end">
          <div class="form-check mb-2">
            <input
              id="doc-visible"
              v-model="form.visible"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="doc-visible"
            >
              {{ $t('field.visible') }}
            </label>
          </div>
        </div>
      </div>

      <div
        v-if="kind === 'markdown'"
        class="row mb-3"
      >
        <div class="col-lg-6 mb-3 mb-lg-0">
          <label class="form-label">{{ $t('field.body') }}</label>
          <textarea
            v-model="form.body"
            class="form-control font-monospace"
            rows="18"
          />
        </div>
        <div class="col-lg-6">
          <label class="form-label">{{ $t('preview') }}</label>
          <div class="border rounded p-3 bg-white overflow-auto doc-preview">
            <markdown-view :body="form.body" />
          </div>
        </div>
      </div>

      <div
        v-else
        class="mb-3"
      >
        <label class="form-label">{{ $t('field.file') }}</label>
        <input
          class="form-control"
          type="file"
          accept="application/pdf,.pdf"
          @change="onFile"
        >
        <div
          v-if="form.fileName"
          class="form-text"
        >
          {{ form.fileName }}
        </div>
      </div>

      <div class="d-flex gap-2">
        <button
          class="btn btn-primary"
          type="submit"
          :disabled="saving"
        >
          {{ $t('save') }}
        </button>
        <router-link
          class="btn btn-outline-secondary"
          :to="cancelLink"
        >
          {{ $t('cancel') }}
        </router-link>
        <button
          v-if="!isCreate"
          type="button"
          class="btn btn-outline-danger ms-auto"
          @click="remove"
        >
          {{ $t('delete') }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'document' } })
import { computed, inject, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import MarkdownView from '../../../components/Docs/MarkdownView.vue'

const props = defineProps({
  namespace: { type: Object, required: true },
})

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const $ComposeAPI = inject('$ComposeAPI', null) || window.__composeAPI
const saving = ref(false)
const error = ref('')
const file = ref(null)
const form = reactive({
  title: '',
  handle: '',
  body: '',
  visible: true,
  fileName: '',
})

const isCreate = computed(() => route.name === 'admin.documents.create')
const kind = computed(() => (isCreate.value ? (route.query.kind === 'pdf' ? 'pdf' : 'markdown') : (loadedKind.value || 'markdown')))
const loadedKind = ref('')
const namespaceID = computed(() => props.namespace?.namespaceID)
const slug = computed(() => route.params.slug || props.namespace?.slug || namespaceID.value)
const cancelLink = computed(() => ({ name: 'admin.documents', params: { slug: slug.value } }))

watch(() => [namespaceID.value, route.params.documentID, route.name], load, { immediate: true })

function onFile (event) {
  file.value = event.target.files?.[0] || null
}

async function load () {
  error.value = ''
  file.value = null
  if (isCreate.value || !namespaceID.value || !route.params.documentID) {
    form.title = ''
    form.handle = ''
    form.body = ''
    form.visible = true
    form.fileName = ''
    loadedKind.value = route.query.kind === 'pdf' ? 'pdf' : 'markdown'
    return
  }
  try {
    const res = await $ComposeAPI.documentRead({
      namespaceID: namespaceID.value,
      documentID: route.params.documentID,
    })
    const doc = res.document || {}
    loadedKind.value = doc.kind || 'markdown'
    form.title = doc.title || ''
    form.handle = doc.handle || ''
    form.body = doc.body || ''
    form.visible = doc.visible !== false
    form.fileName = doc.fileName || ''
  } catch (e) {
    error.value = e?.message || String(e)
  }
}

async function save () {
  saving.value = true
  error.value = ''
  const payload = {
    namespaceID: namespaceID.value,
    title: form.title,
    handle: form.handle,
    kind: kind.value,
    body: kind.value === 'markdown' ? form.body : '',
    visible: form.visible,
  }
  try {
    let documentID = route.params.documentID
    if (isCreate.value) {
      const res = await $ComposeAPI.documentCreate(payload)
      documentID = res.document?.documentID
    } else {
      await $ComposeAPI.documentUpdate({ ...payload, documentID })
    }
    if (kind.value === 'pdf' && file.value && documentID) {
      await $ComposeAPI.documentUpload({
        namespaceID: namespaceID.value,
        documentID,
        file: file.value,
      })
    }
    window.dispatchEvent(new CustomEvent('compose-documents-changed'))
    if (isCreate.value) {
      router.push({ name: 'admin.documents.edit', params: { slug: slug.value, documentID } })
    }
  } catch (e) {
    error.value = e?.message || String(e)
  } finally {
    saving.value = false
  }
}

async function remove () {
  if (!window.confirm(t('document.deleteConfirm'))) return
  try {
    await $ComposeAPI.documentDelete({
      namespaceID: namespaceID.value,
      documentID: route.params.documentID,
    })
    window.dispatchEvent(new CustomEvent('compose-documents-changed'))
    router.push(cancelLink.value)
  } catch (e) {
    error.value = e?.message || String(e)
  }
}
</script>

<style scoped>
.doc-preview {
  min-height: 24rem;
  max-height: 32rem;
}
</style>
