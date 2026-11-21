<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="$t('ui.settings.editor.navigation.title')" />
    <CUITopbarSettings v-if="settings" :settings="settings" :processing="topbar.processing" :success="topbar.success" :can-manage="canManage" @submit="onSubmit($event, 'topbar')" />
  </div>
</template>
<script setup>
defineOptions({ i18nOptions: { namespaces: 'ui.settings', keyPrefix: 'editor.navigation' } })
import { ref, reactive, computed, onMounted, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import CUITopbarSettings from '../../../components/Settings/UI/CUITopbarSettings.vue'
const { t } = useI18n()
const $Settings = inject('$Settings', {})
const settings = ref(undefined)
const topbar = reactive({ processing: false, success: false })
const canManage = computed(() => can('system/', 'settings.manage'))
function can(resource, operation) { return true }
function incLoader() {} function decLoader() {}
function fetchSettings() { incLoader(); $Settings.fetch(); window.__SystemAPI.settingsList({ prefix: 'ui' }).then(s => { settings.value = {}; s.forEach(({ name, value }) => { settings.value[name] = value }) }).finally(() => decLoader()) }
function onSubmit(s, type) { topbar.processing = true; const values = Object.entries(s).map(([name, value]) => ({ name, value })); window.__SystemAPI.settingsUpdate({ values }).then(() => fetchSettings()).then(() => { topbar.success = true; setTimeout(() => { topbar.success = false }, 2000) }).finally(() => { topbar.processing = false }) }
onMounted(() => { fetchSettings() })
</script>
