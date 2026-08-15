<template><router-view v-if="loaded && i18nLoaded" /></template>
<script setup>
import { ref, onMounted } from 'vue'
import { useAuth, useSettings, useToast } from 'corteza-lib/vue/dist'
import { websocket } from 'corteza-lib/vue/dist'

const { auth } = useAuth()
const { $Settings } = useSettings()
const toast = useToast()
const loaded = ref(false)
const i18nLoaded = ref(true)

onMounted(async () => {
  try { websocket.init() } catch (e) {}
  try {
    const { user } = await auth.handle()
    await $Settings.init({ api: window.__systemAPI })
    const icon = $Settings.attachment('ui.iconLogo') || '/icon.svg'
    const favicon = document.getElementById('favicon')
    if (favicon) { favicon.href = icon }
    if (user.meta.preferredLanguage) {
      window.__composeAPI?.setHeader?.('Accept-Language', user.meta.preferredLanguage)
      window.__composeAPI?.setHeader?.('Content-Language', user.meta.preferredLanguage)
    }
    if (user.meta.theme) { document.getElementsByTagName('html')[0].setAttribute('data-color-mode', user.meta.theme) }
    loaded.value = true
  } catch (err) {
    if (err instanceof Error && err.message === 'Unauthenticated') { auth.startAuthenticationFlow(); return }
    throw err
  }
})
</script>
