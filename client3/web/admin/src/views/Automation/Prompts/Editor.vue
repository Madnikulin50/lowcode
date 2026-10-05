<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="title">
      <button
        v-if="handle"
        class="btn btn-primary"
        @click="$router.push({ name: 'automation.prompts.new' })"
      >
        {{ $t('automation.prompts.editor.new') }}
      </button>
    </c-content-header>

    <c-prompt-editor-info
      :form="form"
      :is-new="!handle"
      :history="history"
      :viewing="viewing"
      :processing="info.processing"
      :success="info.success"
      :error="info.error"
      @submit="onSubmit"
      @delete="onDelete"
      @view="view"
      @activate="activate"
    />

    <prompt-eval
      v-if="handle && history.length"
      :handle="handle"
      :versions="history"
      :cases-text="form.casesText"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'automation.prompts' } })
import { computed, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import CPromptEditorInfo from '../../../components/Prompts/CPromptEditorInfo.vue'
import PromptEval from '../../../components/Prompts/PromptEval.vue'

const props = defineProps({ handle: { type: String, required: false, default: undefined } })

const router = useRouter()
const { t } = useI18n()
const api = () => window.__AutomationAPI

const history = ref([])
const viewing = ref(0)
const info = reactive({ processing: false, success: false, error: '' })
const form = reactive({ handle: '', description: '', text: '', note: '', casesText: '', activate: true })

const title = computed(() => props.handle ? t('automation.prompts.editor.title.edit') : t('automation.prompts.editor.title.create'))

function message (e) {
  return (e && e.response && e.response.data && e.response.data.error && e.response.data.error.message) || String((e && e.message) || e)
}

watch(() => props.handle, async () => {
  info.error = ''
  history.value = []
  viewing.value = 0

  if (!props.handle) {
    // the first version is live whether asked or not
    Object.assign(form, { handle: '', description: '', text: '', note: '', casesText: '', activate: true })
    return
  }

  form.handle = props.handle
  try {
    await loadHistory()
    const active = history.value.find(v => v.active) || history.value[history.value.length - 1]
    await view(active.version)
  } catch (e) {
    info.error = message(e)
  }
}, { immediate: true })

async function loadHistory () {
  history.value = (await api().promptHistory({ handle: props.handle })).set || []
}

async function view (version) {
  info.error = ''
  try {
    const { prompt, cases } = await api().promptVersion({ handle: props.handle, version })
    viewing.value = prompt.version
    Object.assign(form, {
      handle: props.handle,
      description: prompt.description || '',
      text: prompt.text || '',
      note: '',
      casesText: cases && cases.length ? JSON.stringify(cases, null, 2) : '',
      activate: false, // a new version is tested before it goes live
    })
  } catch (e) {
    info.error = message(e)
  }
}

async function onSubmit () {
  info.error = ''
  info.processing = true
  const target = props.handle || form.handle.trim()

  try {
    // an empty box clears the test cases: what is shown is what is saved
    const saved = await api().promptSave({
      handle: target,
      text: form.text,
      description: form.description,
      note: form.note,
      activate: form.activate,
      cases: form.casesText.trim() ? JSON.parse(form.casesText) : [],
    })

    info.success = true
    setTimeout(() => { info.success = false }, 2000)

    if (!props.handle) {
      router.push({ name: 'automation.prompts.edit', params: { handle: target } })
    } else {
      await loadHistory()
      await view(saved.version)
    }
  } catch (e) {
    info.error = message(e)
  } finally {
    info.processing = false
  }
}

async function activate (version) {
  info.error = ''
  try {
    await api().promptActivate({ handle: props.handle, version })
    await loadHistory()
  } catch (e) {
    info.error = message(e)
  }
}

async function onDelete () {
  info.error = ''
  try {
    await api().promptDelete({ handle: props.handle })
    router.push({ name: 'automation.prompts' })
  } catch (e) {
    info.error = message(e)
  }
}
</script>
