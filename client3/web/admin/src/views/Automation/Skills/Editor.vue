<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="title">
      <button
        v-if="handle"
        class="btn btn-primary"
        @click="$router.push({ name: 'automation.skills.new' })"
      >
        {{ $t('automation.skills.editor.new') }}
      </button>
    </c-content-header>

    <c-skill-editor-info
      :form="form"
      :is-new="!handle"
      :history="history"
      :viewing="viewing"
      :kit-options="kitOptions"
      :processing="info.processing"
      :success="info.success"
      :error="info.error"
      @submit="onSubmit"
      @delete="onDelete"
      @view="view"
      @activate="activate"
      @export="exportSkill"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'automation.skills' } })
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import CSkillEditorInfo from '../../../components/Skills/CSkillEditorInfo.vue'

const props = defineProps({ handle: { type: String, required: false, default: undefined } })

const router = useRouter()
const { t } = useI18n()
const api = () => window.__AutomationAPI

const history = ref([])
const viewing = ref(0)
const kitOptions = ref([])
const info = reactive({ processing: false, success: false, error: '' })
const form = reactive(emptyForm())

const title = computed(() => props.handle ? t('automation.skills.editor.title.edit') : t('automation.skills.editor.title.create'))

function emptyForm () {
  // the first version is live whether asked or not
  return { handle: '', description: '', text: '', note: '', activate: true, requires: [], resources: [] }
}

function message (e) {
  return (e && e.response && e.response.data && e.response.data.error && e.response.data.error.message) || String((e && e.message) || e)
}

// the toolkits that exist right now, to choose from; the editor works without them
onMounted(async () => {
  try {
    const raw = await window.__ComposeAPI.api().request({ method: 'get', url: '/chat/agents' })
    const res = (raw && raw.data && (raw.data.response || raw.data)) || {}
    const names = new Set()
    for (const k of [].concat(res.toolkits || [], res.kits || [])) {
      const n = typeof k === 'string' ? k : (k && (k.name || k.handle))
      if (n) names.add(n)
    }
    kitOptions.value = [...names].sort()
  } catch (e) {
    kitOptions.value = []
  }
})

watch(() => props.handle, async () => {
  info.error = ''
  history.value = []
  viewing.value = 0

  if (!props.handle) {
    Object.assign(form, emptyForm())
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
  history.value = (await api().skillHistory({ handle: props.handle })).set || []
}

async function view (version) {
  info.error = ''
  try {
    const { skill } = await api().skillVersion({ handle: props.handle, version })
    viewing.value = skill.version
    Object.assign(form, {
      handle: props.handle,
      description: skill.description || '',
      text: skill.text || '',
      note: '',
      activate: false, // a new version is tried before it goes live
      requires: [...(skill.requires || [])],
      resources: (skill.resources || []).map(f => ({ path: f.path, content: f.content || '' })),
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
    // what is shown is what is saved: an empty list clears
    const saved = await api().skillSave({
      handle: target,
      text: form.text,
      description: form.description,
      note: form.note,
      activate: form.activate,
      requires: form.requires,
      resources: form.resources.filter(f => f.path.trim()).map(f => ({ path: f.path.trim(), content: f.content })),
    })

    info.success = true
    setTimeout(() => { info.success = false }, 2000)

    if (!props.handle) {
      router.push({ name: 'automation.skills.edit', params: { handle: target } })
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
    await api().skillActivate({ handle: props.handle, version })
    await loadHistory()
  } catch (e) {
    info.error = message(e)
  }
}

async function onDelete () {
  info.error = ''
  try {
    await api().skillDelete({ handle: props.handle })
    router.push({ name: 'automation.skills' })
  } catch (e) {
    info.error = message(e)
  }
}

// the viewed version as SKILL.md (a zip when it has files)
async function exportSkill () {
  info.error = ''
  try {
    const blob = await api().skillExport({ handle: props.handle, version: viewing.value })
    const name = `${props.handle}.${blob.type === 'application/zip' ? 'zip' : 'md'}`
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = name
    document.body.appendChild(a)
    a.click()
    a.remove()
    URL.revokeObjectURL(url)
  } catch (e) {
    info.error = message(e)
  }
}
</script>
