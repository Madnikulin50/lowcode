<template>
  <router-view v-if="loaded && i18nLoaded" />
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue'
import { useRouter } from 'vue-router'
import { useStore } from './store'
import { corredor, websocket, i18n, applyColorMode } from 'corteza-lib/vue/dist'
import { compose } from 'corteza-lib/js/dist'

const router = useRouter()
const store = useStore()

const loaded = ref(false)
const i18nLoaded = ref(false)

const notProduction = import.meta.env.DEV
const verboseEventbus = window.location.search.includes('verboseEventbus')

const app = getCurrentInstance()?.appContext?.app
const $auth = app?.config?.globalProperties?.$auth || window.__auth
const $ComposeAPI = app?.config?.globalProperties?.$ComposeAPI || window.__composeAPI
const $SystemAPI = app?.config?.globalProperties?.$SystemAPI || window.__systemAPI
const $AutomationAPI = app?.config?.globalProperties?.$AutomationAPI || window.__automationAPI
const $Settings = app?.config?.globalProperties?.$Settings || window.__settings

let websocketCleanup = null

onMounted(async () => {
  $auth.handle().then(async ({ user }) => {
    await $Settings.init({ api: $SystemAPI }).then(() => {
      const icon = $Settings.attachment('ui.iconLogo') || '/icon.svg'
      const favicon = document.getElementById('favicon')
      if (favicon) favicon.href = icon
    })


    if (user.meta.preferredLanguage) {
      $ComposeAPI.setHeader('Accept-Language', user.meta.preferredLanguage)
        .setHeader('Content-Language', user.meta.preferredLanguage)
    }

    if (user.meta.theme) {
      applyColorMode(user.meta.theme)
    }

    const bundleLoaderOpt = {
      bundle: 'compose',
      verbose: notProduction || verboseEventbus,
      ctx: new corredor.ComposeCtx(
        { $invoker: $auth.user, authToken: $auth.accessToken },
        {
          $SystemAPI,
          $ComposeAPI,
          $store: Object.assign(store, {
            getters: new Proxy({}, {
              get(_, path) {
                const [ns, prop] = String(path).split('/')
                return store[ns]?.[prop]
              },
            }),
          }),
          $emit: (name, params) => window.dispatchEvent(new CustomEvent(name, { detail: params })),
          $router: router,
        },
      ),
    }

    store.wfPrompts.update()
    store.rbac.load($ComposeAPI, $SystemAPI, $AutomationAPI)
    store.notifications.fetchNotifications()
    store.drafts.init({
      composeAPI: $ComposeAPI,
      systemAPI: $SystemAPI,
      resourceType: 'compose:record',
    })

    try { websocket.init($auth) } catch (e) {}
    websocketCleanup = handleWebSocket()

    const loadBundle = bundleLoaderOpt.ctx.loadBundle?.(bundleLoaderOpt)
    if (loadBundle) {
      loadBundle.then(() => $ComposeAPI.automationList({ excludeInvalid: true }))
        .then(/* makeAutomationScriptsRegistrator pending */)
    }

    console.log('token after handle:', $auth.accessToken)
    i18nLoaded.value = true
    loaded.value = true

    const url = new URL(window.location.href)
    if (url.searchParams.get('code')) {
      url.searchParams.delete('code')
      window.location.replace(url.toString())
    }
  }).catch((err) => {
    if (err instanceof Error && err.message === 'Unauthenticated') {
      $auth.startAuthenticationFlow()
      return
    }
    throw err
  })
})

function handleWebSocket() {
  const handler = ({ data }) => {
    const msg = JSON.parse(data)
    switch (msg['@type']) {
      case 'workflowSessionPrompt':
        store.wfPrompts.new(msg['@value'])
        break
      case 'workflowSessionResumed':
        store.wfPrompts.clear(msg['@value'])
        break
      case 'reminder':
        window.__reminder?.enqueueRaw(msg['@value'])
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
        console.warn('Websocket message with error', msg['@value'])
    }
  }
  window.addEventListener('websocket-message', handler)
  return () => window.removeEventListener('websocket-message', handler)
}

onBeforeUnmount(() => {
  websocketCleanup?.()
})

function textDirectionality(lang) {
  const rtlLanguages = ['ar', 'he', 'fa', 'ur']
  if (!lang) return 'ltr'
  const base = lang.split('-')[0]
  return rtlLanguages.includes(base) ? 'rtl' : 'ltr'
}
</script>
