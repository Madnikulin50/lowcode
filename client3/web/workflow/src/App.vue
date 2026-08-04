<template>
  <router-view v-if="loaded && i18nLoaded" />
  <c-toaster :toasts="toasts" />
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuth, useSettings, useToast, toasts, components } from 'corteza-lib/vue/dist'
import { websocket } from 'corteza-lib/vue/dist'

const { CToaster } = components
const router = useRouter()
const { t } = useI18n()
const { auth } = useAuth()
const { $Settings } = useSettings()
const toast = useToast()

const loaded = ref(false)
const i18nLoaded = ref(true)

function handleWebsocketMessage(event) {
  const msg = JSON.parse(event.data)
  switch (msg['@type']) {
    case 'notification':
      break
    case 'notification.read':
      break
    case 'notification.unread':
      break
    case 'notification.read.all':
      break
    case 'notification.unread.all':
      break
    case 'notification.delete':
      break
    case 'error':
      toast.toastDanger('Websocket message with error', msg['@value'])
  }
}

onMounted(async () => {
  try { websocket.init(auth) } catch (e) {}
  window.addEventListener('websocket-message', handleWebsocketMessage)

  try {
    const { user } = await auth.handle()

    await $Settings.init({ api: window.__systemAPI })
    const icon = $Settings.attachment('ui.iconLogo') || '/icon.svg'
    const favicon = document.getElementById('favicon')
    if (favicon) { favicon.href = icon }

    if (user.meta.preferredLanguage) {
    }

    if (user.meta.theme) {
      document.getElementsByTagName('html')[0].setAttribute('data-color-mode', user.meta.theme)
    }

    i18nLoaded.value = true
    loaded.value = true

    const url = new URL(window.location.href)
    if (url.searchParams.get('code')) {
      url.searchParams.delete('code')
      window.location.replace(url.toString())
    }
  } catch (err) {
    if (err instanceof Error && err.message === 'Unauthenticated') {
      auth.startAuthenticationFlow()
      return
    }
    throw err
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('websocket-message', handleWebsocketMessage)
})
</script>
