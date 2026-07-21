<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="$t('compose.settings.list.title')" />
    <c-compose-editor-basic :basic="settings" :processing="basic.processing" :success="basic.success" :can-manage="canManage" @submit="onSubmit($event, 'basic')" />
    <c-compose-editor-ui :settings="settings" :processing="ui.processing" :success="ui.success" :can-manage="canManage" class="mt-3" @submit="onSubmit($event, 'ui')" />
  </div>
</template>
<script setup>
import { ref, reactive, computed, onMounted, inject } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { isEqual, cloneDeep } from 'lodash'
import CComposeEditorBasic from '../../../components/Settings/Compose/CComposeEditorBasic.vue'
import CComposeEditorUI from '../../../components/Settings/Compose/CComposeEditorUI.vue'
const router = useRouter()
const { t } = useI18n()
const prefix = 'compose.'
const settings = ref({})
const initialSettingsState = ref({})
const basic = reactive({ processing: false, success: false })
const ui = reactive({ processing: false, success: false })
const canManage = computed(() => can('system/', 'settings.manage'))
function can(resource, operation) { return true }
function incLoader() {} function decLoader() {}
onMounted(() => { fetchSettings() })
function fetchSettings() { incLoader(); window.__SystemAPI.settingsList({ prefix }).then(s => { settings.value = {}; initialSettingsState.value = {}; s.forEach(({ name, value }) => { settings.value[name] = value; initialSettingsState.value[name] = cloneDeep(value) }) }).catch(() => {}).finally(() => decLoader()) }
function onSubmit(s, type) { const obj = type === 'basic' ? basic : ui; obj.processing = true; const values = Object.entries(s).map(([name, value]) => ({ name, value })); window.__SystemAPI.settingsUpdate({ values }).then(() => { obj.success = true; setTimeout(() => { obj.success = false }, 2000); initialSettingsState.value = cloneDeep(settings.value) }).catch(() => {}).finally(() => { obj.processing = false }) }
</script>
