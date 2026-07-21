<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="$t('system.settings.list.title')" />
    <template v-if="Object.keys(getAuth).length">
      <c-system-editor-auth :settings="getAuth" :processing="auth.processing" :success="auth.success" :can-manage="canManage" @submit="onAuthSubmit" />
      <c-system-editor-external v-model="settings" class="mt-3" :processing="external.processing" :success="external.success" :can-manage="canManage" @submit="onExternalSubmit" />
      <c-system-editor-auth-bg-screen :settings="getAuthBackground" :can-manage="canManage" :processing="authBackground.processing" :success="authBackground.success" class="mt-3" @on-upload="onBackgroundImageUpload" @reset-attachment="onResetBackgroundImage" @submit="onAuthBackgroundSubmit" />
    </template>
  </div>
</template>
<script setup>
import { ref, reactive, computed, onMounted, inject } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import CSystemEditorAuth from '../../../components/Settings/System/CSystemEditorAuth.vue'
import CSystemEditorExternal from '../../../components/Settings/System/CSystemEditorExternal.vue'
import CSystemEditorAuthBgScreen from '../../../components/Settings/System/CSystemEditorAuthBgScreen.vue'
const router = useRouter()
const { t } = useI18n()
const $Settings = inject('$Settings', {})
const settings = ref([])
const auth = reactive({ processing: false, success: false })
const external = reactive({ processing: false, success: false })
const authBackground = reactive({ processing: false, success: false })
const canManage = computed(() => can('system/', 'settings.manage'))
function can(resource, operation) { return true }
function incLoader() {} function decLoader() {}
const getAuth = computed(() => filterSettings('auth'))
const getAuthBackground = computed(() => filterSettings('auth.ui'))
onMounted(() => { fetchSettings() })
function fetchSettings() { incLoader(); window.__systemAPI.settingsList().then(s => { settings.value = s }).catch(e => { router.push({ name: 'dashboard' }) }).finally(() => decLoader()) }
function filterSettings(prefix) { if (settings.value.length > 0) return settings.value.reduce((map, obj) => { const { name, value } = obj; if (name.startsWith(prefix)) map[name] = value; return map }, {}); return {} }
function onAuthSubmit(a) { auth.processing = true; const values = Object.entries(a).map(([name, value]) => ({ name, value })); window.__systemAPI.settingsUpdate({ values }).then(() => { auth.success = true; setTimeout(() => { auth.success = false }, 2000); $Settings.fetch() }).finally(() => { auth.processing = false }) }
function onExternalSubmit(e) { external.processing = true; window.__systemAPI.settingsUpdate({ values: e }).then(() => { external.success = true; setTimeout(() => { external.success = false }, 2000); $Settings.fetch() }).finally(() => { external.processing = false; fetchSettings() }) }
function onBackgroundImageUpload() { fetchSettings() }
function onResetBackgroundImage(name) { window.__systemAPI.settingsUpdate({ values: [{ name, value: undefined }], upload: {} }).then(() => fetchSettings()) }
function onAuthBackgroundSubmit(a) { authBackground.processing = true; const values = [{ name: 'auth.ui.styles', value: a }]; window.__systemAPI.settingsUpdate({ values }).then(() => { authBackground.success = true; setTimeout(() => { authBackground.success = false }, 2000); $Settings.fetch() }).finally(() => { authBackground.processing = false }) }
</script>
