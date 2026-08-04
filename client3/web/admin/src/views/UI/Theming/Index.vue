<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="$t('ui.settings.editor.corteza-studio.title')" />
    <CUIBrandingEditor v-if="settings" :settings="settings" :processing="branding.processing" :success="branding.success" :can-manage="canManage" @submit="onSubmit($event, 'branding')" />
  </div>
</template>
<script setup>
import { ref, reactive, computed, onMounted, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import CUIBrandingEditor from '../../../components/Settings/UI/CUIBrandingEditor.vue'
const { t } = useI18n()
const $Settings = inject('$Settings', {})
const settings = ref(undefined)
const branding = reactive({ processing: false, success: false })
const canManage = computed(() => can('system/', 'settings.manage'))
function can(resource, operation) { return true }
function incLoader() {} function decLoader() {}
function fetchSettings() { incLoader(); $Settings.fetch(); window.__SystemAPI.settingsList({ prefix: 'ui' }).then(s => { settings.value = {}; s.forEach(({ name, value }) => { settings.value[name] = value }) }).finally(() => decLoader()) }
function onSubmit(s, type) { branding.processing = true; const values = Object.entries(s).map(([name, value]) => ({ name, value })); window.__SystemAPI.settingsUpdate({ values }).then(() => fetchSettings().then(() => { if (type === 'branding' && settings.value['ui.studio.sass-installed']) { setTimeout(() => { const stylesheet = document.querySelector('link#corteza-custom-css'); if (stylesheet) stylesheet.href = 'custom.css?v=' + new Date().getTime() }, 1000) } })).then(() => { branding.success = true; setTimeout(() => { branding.success = false }, 2000) }).finally(() => { branding.processing = false }) }
onMounted(() => { fetchSettings() })
</script>
