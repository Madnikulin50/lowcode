<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="$t('ui.settings.editor.location.title')" />
    <CUILocationSettings v-if="settings" :settings="settings" :processing="location.processing" :success="location.success" :can-manage="canManage" @submit="onSubmit($event, 'location')" />
  </div>
</template>
<script setup>
defineOptions({ i18nOptions: { namespaces: 'ui.settings', keyPrefix: 'editor.location' } })
import { ref, reactive, computed, onMounted, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import CUILocationSettings from '../../../components/Settings/UI/CUILocationSettings.vue'
const { t } = useI18n()
const $Settings = inject('$Settings', {})
const settings = ref(undefined)
const location = reactive({ processing: false, success: false })
const canManage = computed(() => can('system/', 'settings.manage'))
function can(resource, operation) { return true }
function incLoader() {} function decLoader() {}
function fetchSettings() { incLoader(); $Settings.fetch(); window.__SystemAPI.settingsList({ prefix: 'ui' }).then(s => { settings.value = {}; s.forEach(({ name, value }) => { settings.value[name] = value }) }).finally(() => decLoader()) }
function onSubmit(s, type) { location.processing = true; const values = Object.entries(s).map(([name, value]) => ({ name, value })); window.__SystemAPI.settingsUpdate({ values }).then(() => fetchSettings()).then(() => { location.success = true; setTimeout(() => { location.success = false }, 2000) }).finally(() => { location.processing = false }) }
onMounted(() => { fetchSettings() })
</script>
