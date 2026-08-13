<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="t('list.title')" />
    <c-ai-models-editor
      :settings="settings"
      :can-manage="canManage"
      :processing="save.processing"
      :success="save.success"
      @submit="onSubmit"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.ai' } })

import { ref, reactive, computed, onMounted, inject } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { cloneDeep } from 'lodash'
import CAiModelsEditor from '../../../components/Settings/System/CAIModelsEditor.vue'

const router = useRouter()
const { t } = useI18n()
const $Settings = inject('$Settings', {})
const settings = ref({})
const save = reactive({ processing: false, success: false })
const canManage = computed(() => true)

onMounted(() => { fetchSettings() })

function fetchSettings () {
  window.__SystemAPI.settingsList({ prefix: 'ai.' }).then(s => {
    const map = {}
    s.forEach(({ name, value }) => { map[name] = value })
    // Defaults when never configured
    if (map['ai.enabled'] === undefined) map['ai.enabled'] = true
    if (!Array.isArray(map['ai.catalog'])) map['ai.catalog'] = []
    if (map['ai.roles.compose-chat'] === undefined) map['ai.roles.compose-chat'] = ''
    if (map['ai.roles.mcp-agent'] === undefined) map['ai.roles.mcp-agent'] = ''
    if (map['ai.roles.automation-chat'] === undefined) map['ai.roles.automation-chat'] = ''
    if (map['ai.roles.rulesgo-ai'] === undefined) map['ai.roles.rulesgo-ai'] = ''
    if (map['ai.ollama-url'] === undefined) map['ai.ollama-url'] = ''
    settings.value = map
  }).catch(() => {
    router.push({ name: 'dashboard' })
  })
}

function onSubmit (values) {
  save.processing = true
  const payload = Object.entries(values).map(([name, value]) => ({ name, value }))
  window.__SystemAPI.settingsUpdate({ values: payload }).then(() => {
    save.success = true
    setTimeout(() => { save.success = false }, 2000)
    settings.value = cloneDeep(values)
    if ($Settings && $Settings.fetch) $Settings.fetch()
  }).finally(() => {
    save.processing = false
  })
}
</script>
