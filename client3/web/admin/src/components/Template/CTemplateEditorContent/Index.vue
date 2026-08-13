<template>
  <div class="row">
    <div class="col-12 col-lg-3 mb-3">
      <editor-toolbox
        :template="template"
        :partials="partials"
      />
    </div>

    <div class="col-12 col-lg-9 mb-3">
      <div v-if="!template.partial" class="card shadow-sm">
        <div class="card-header border-bottom">
          <h4 class="m-0">{{ $t('system.templates.editor.content.preview.title') }}</h4>
        </div>
        <div class="card-body p-0">
          <c-ace-editor
            v-model="previewData"
            data-test-id="template-preview-output"
            name="preview-data"
            lang="json"
            min-height="300px"
            show-line-numbers
            highlight-active-line
            resizable
            :border="false"
          />
        </div>
        <div class="card-footer border-top d-flex justify-content-end flex-wrap flex-fill-child gap-1">
          <button
            v-if="canPreviewHTML"
            type="button"
            class="btn btn-outline-secondary"
            data-test-id="button-preview-html-template"
            @click="openPreview('html')"
          >
            {{ $t('system.templates.editor.content.preview.html') }}
          </button>
          <button
            v-if="canPreviewPDF"
            type="button"
            class="btn btn-outline-secondary"
            data-test-id="button-preview-pdf-template"
            @click="openPreview('pdf')"
          >
            {{ $t('system.templates.editor.content.preview.pdf') }}
          </button>
        </div>
      </div>
    </div>

    <div class="col-12">
      <div class="card shadow-sm">
        <div class="card-header border-bottom d-flex align-items-center">
          <h4 class="m-0">{{ $t('system.templates.editor.content.title') }}</h4>
          <span
            v-if="template.partial"
            class="badge bg-primary ms-2"
            data-test-id="badge-partial-template"
          >
            {{ $t('system.templates.editor.content.partial') }}
          </span>
        </div>
        <div class="card-body p-0">
          <component :is="editor" :template="template" />
        </div>
        <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
          <c-button-submit
            :disabled="!canCreate"
            :processing="processing"
            :success="success"
            :text="$t('admin.general.label.submit')"
            class="ms-auto"
            @submit="$emit('submit', template)"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.templates', keyPrefix: 'editor.content' } })
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useListHelpers } from 'corteza-webapp-admin/src/mixins/listHelpers'
import EditorToolbox from './EditorToolbox'
import EditorTextHtml from './EditorTextHtml'
import EditorTextPlain from './EditorTextPlain'
import EditorUnsupported from './EditorUnsupported'
import { components } from 'corteza-lib/vue/dist'

const { CAceEditor } = components
const { t } = useI18n()
const { incLoader, decLoader } = useListHelpers()

const props = defineProps({
  template: { type: Object, required: true, default: () => ({}) },
  partials: { type: Array, required: false, default: () => [] },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  canCreate: { type: Boolean, required: true },
})

defineEmits(['submit'])

const previewData = ref('{\n  "variables": {\n    "param1": "value1",\n    "param2": {\n      "nestedParam1": "value2"\n    }\n  },\n  "options": {\n    "documentSize": "A4",\n    "contentScale": "1",\n    "orientation": "portrait",\n    "margin": "0.3"\n  }\n}\n')
const previewBlob = ref('')
const availableDrivers = ref([])

const editor = computed(() => {
  switch (props.template.type) {
    case 'text/html': return EditorTextHtml
    case 'text/plain': return EditorTextPlain
    default: return EditorUnsupported
  }
})

const canPreviewHTML = computed(() => {
  return availableDrivers.value.find(({ outputTypes }) => outputTypes.includes('text/html'))
})

const canPreviewPDF = computed(() => {
  return availableDrivers.value.find(({ outputTypes }) => outputTypes.includes('application/pdf'))
})

onMounted(async () => {
  availableDrivers.value = await window.__SystemAPI.templateRenderDrivers()
    .then(rsp => rsp.set)
    .catch(() => [])
})

function openPreview(ext) {
  incLoader()

  const cfg = {
    method: 'post',
    responseType: 'blob',
    url: window.__SystemAPI.templateRenderEndpoint({
      templateID: props.template.templateID,
      filename: 'preview',
      ext,
    }),
    data: JSON.parse(previewData.value),
  }

  window.__SystemAPI.api().request(cfg)
    .then(r => {
      previewBlob.value = window.URL.createObjectURL(r.data)
      window.open(previewBlob.value, '_newtab')
    })
    .catch(() => {})
    .finally(() => {
      decLoader()
    })
}
</script>