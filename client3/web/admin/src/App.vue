<template>
  <router-view v-if="loaded && i18nLoaded && isRbacLoaded" />
  <c-toaster :toasts="toasts" />
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useStore } from './store'
import { useI18n } from 'vue-i18n'
import { system } from 'corteza-lib/js/dist'
import { websocket } from 'corteza-lib/vue/dist'
import { useAuth, useSettings, useToast, toasts, components } from 'corteza-lib/vue/dist'

const { CToaster } = components
const router = useRouter()
const store = useStore()
const { t } = useI18n()
const { auth } = useAuth()
const { $Settings } = useSettings()
const toast = useToast()

const loaded = ref(false)
const i18nLoaded = ref(true)

const isRbacLoaded = computed(() => store.rbac.isLoaded)

const notProduction = import.meta.env.DEV

function textDirectionality(lang) {
  if (!lang) return 'ltr'
  const rtlLangs = ['ar', 'he', 'fa', 'ur']
  const base = lang.split('-')[0]
  return rtlLangs.includes(base) ? 'rtl' : 'ltr'
}

function handleWebsocketMessage(event) {
  const msg = JSON.parse(event.data)
  switch (msg['@type']) {
    case 'workflowSessionPrompt':
      store.wfPrompts.new(msg['@value'])
      break
    case 'workflowSessionResumed':
      store.wfPrompts.clear(msg['@value'])
      break
    case 'notification':
      store.notifications.addNotification(msg['@value'])
      break
    case 'notification.read':
      store.notifications.updateReadNotification(msg['@value'])
      break
    case 'notification.unread':
      store.notifications.updateUnreadNotification(msg['@value'])
      break
    case 'notification.read.all':
      store.notifications.updateAllReadNotifications(msg['@value'])
      break
    case 'notification.unread.all':
      store.notifications.updateAllUnreadNotifications(msg['@value'])
      break
    case 'notification.delete':
      store.notifications.removeNotification(msg['@value'])
      break
    case 'error':
      toast.danger('Websocket message with error', msg['@value'])
  }
}

onMounted(async () => {
  try { websocket.init() } catch (e) {}
  window.addEventListener('websocket-message', handleWebsocketMessage)

  try {
    const { user } = await auth.handle()

    await $Settings.init({ api: window.__systemAPI })
    const icon = $Settings.attachment('ui.iconLogo') || '/icon.svg'
    const favicon = document.getElementById('favicon')
    if (favicon) { favicon.href = icon }


    if (user.meta.preferredLanguage) {
      // i18n change language
    }

    if (user.meta.theme) {
      const html = document.getElementsByTagName('html')[0]
      html.setAttribute('data-color-mode', user.meta.theme)
      html.setAttribute('data-bs-theme', user.meta.theme)
    }

    const bundleLoaderOpt = {
      bundle: 'admin',
      verbose: notProduction,
      ctx: new (await import('corteza-lib/vue/dist')).corredor.WebappCtx({
        $invoker: user,
        authToken: auth.accessToken,
      }),
    }

    store.wfPrompts.update()

    const enabledApis = [window.__systemAPI, window.__composeAPI, window.__automationAPI]
    if ($Settings.get('federation.enabled', false)) {
      enabledApis.push(window.__federationAPI)
    }

    store.rbac.load(enabledApis)
    store.notifications.fetchNotifications()

    // Bundle loading
    const loadBundle = bundleLoaderOpt.ctx.loadBundle?.(bundleLoaderOpt)
    if (loadBundle) {
      await loadBundle.then(() => window.__systemAPI.automationList({ excludeInvalid: true }))
    } else {
      await window.__systemAPI.automationList({ excludeInvalid: true })
    }

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
