<template><router-view v-if="loaded && i18nLoaded" /></template>
<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuth, useSettings, useToast, applyColorMode } from 'corteza-lib/vue/dist'
import { websocket } from 'corteza-lib/vue/dist'
const { t } = useI18n()
const { auth } = useAuth()
const { $Settings } = useSettings()
const toast = useToast()
const loaded = ref(false)
const i18nLoaded = ref(true)
onMounted(async () => {
  try { websocket.init(auth) } catch (e) {}
  try {
    const { user } = await auth.handle()
    await $Settings.init({ api: window.__systemAPI })
    const icon = $Settings.attachment('ui.iconLogo') || '/icon.svg'
    const favicon = document.getElementById('favicon')
    if (favicon) favicon.href = icon
    document.body.setAttribute('dir', user.meta?.preferredLanguage?.startsWith('ar') || user.meta?.preferredLanguage?.startsWith('he') || user.meta?.preferredLanguage?.startsWith('fa') ? 'rtl' : 'ltr')
    if (user.meta.theme) {
      applyColorMode(user.meta.theme)
    }
    loaded.value = true
  } catch (err) {
    if (err instanceof Error && err.message === 'Unauthenticated') { auth.startAuthenticationFlow(); return }
    throw err
  }
})
</script>