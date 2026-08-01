import { type App } from 'vue'
import { createI18n, type I18n, type I18nOptions } from 'vue-i18n'
import i18next, { InitOptions } from 'i18next'
import http from 'i18next-http-backend'
import detector from 'i18next-browser-languagedetector'
import multiload from 'i18next-multiload-backend-adapter'
import moment from 'moment'
import Pseudo from 'i18next-pseudo'

interface Options {
  app: string;
  resources: object;
  lng: string;
  fallbackLng: string | false;
  ns: string | Array<string>;
  fallbackNS: string | false;
  defaultNS: string;
  baseURL: string;
  pseudo: boolean;
}

function flattenMessages(nestedMessages: Record<string, unknown>, prefix = ''): Record<string, string> {
  return Object.keys(nestedMessages).reduce((messages: Record<string, string>, key: string) => {
    const value = nestedMessages[key]
    const prefixedKey = prefix ? `${prefix}.${key}` : key

    if (typeof value === 'string') {
      messages[prefixedKey] = value.replace(/\{\{(\w+)\}\}/g, '{$1}')
    } else if (typeof value === 'object' && value !== null) {
      Object.assign(messages, flattenMessages(value as Record<string, unknown>, prefixedKey))
    }

    return messages
  }, {})
}

function mergeMessages(
  merged: Record<string, string>,
  raw: Record<string, unknown>,
  ns: string,
): void {
  const prefixed = flattenMessages(raw, ns)
  const plain = flattenMessages(raw)

  // Add namespace-prefixed keys (e.g. general.label.loading)
  Object.assign(merged, prefixed)

  // Also add unprefixed aliases for backward compat (e.g. system.items.apigw)
  for (const [k, v] of Object.entries(plain)) {
    if (!(k in merged)) {
      merged[k] = v
    }
  }

  // Also add aliases without the top-level section prefix
  // e.g. general.customCSSClass.label -> customCSSClass.label
  for (const [k, v] of Object.entries(plain)) {
    const dotIndex = k.indexOf('.')
    if (dotIndex > 0) {
      const shortKey = k.substring(dotIndex + 1)
      if (shortKey && !(shortKey in merged)) {
        merged[shortKey] = v
      }
    }
  }
}

export default function (app: App, appName: string | Partial<Options>, ...namespaces: Array<string>): I18n {
  const devMode = process.env.NODE_ENV !== 'production'
  const defNS = 'translation'

  let opt: Partial<Options> = {}
  if (typeof appName === 'string') {
    opt = { app: appName }
  } else {
    opt = appName
  }

  const { lng = 'ru', fallbackLng = 'en', fallbackNS = false } = opt

  let ns: Array<string> = []
  if (!Array.isArray(opt.ns)) {
    ns = [opt.ns || defNS]
  } else {
    ns = opt.ns
  }
  ns.push(...namespaces)

  const defaultNS = opt.defaultNS || ns[0]

  if (!opt.baseURL) {
    const { CortezaAPI = undefined } = window as unknown as Record<string, string | undefined>
    if (!CortezaAPI) {
      throw new Error('config.js missing or window.CortezaAPI not set')
    }
    opt.baseURL = `${CortezaAPI}/system`
  }

  const pseudo = devMode && (
    !!opt.pseudo ||
    !!(window as unknown as Record<string, unknown>).i18nPseudoModeEnabled ||
    window.location.search.indexOf('i18nPseudoModeEnabled') > -1
  )

  let postProcess: Array<string> = []
  if (pseudo) {
    postProcess = ['pseudo']
  }

  const options: InitOptions = {
    debug: devMode,
    lng,
    fallbackLng,
    ns,
    fallbackNS,
    defaultNS,
    postProcess,
    load: 'languageOnly',
    initImmediate: false,
    detection: {
      order: ['querystring', 'localStorage', 'cookie', 'navigator'],
      caches: devMode ? [] : ['localStorage', 'cookie'],
    },
    backend: {
      backend: http,
      backendOption: {
        loadPath: `${opt.baseURL}/locale/{{lng}}/corteza-webapp-${opt.app}`,
      },
    },
  }

  i18next
    .use(detector)
    .use(multiload)
    .use(new Pseudo({ enabled: pseudo }))
    .init(options)

  moment.locale(lng)

  function loadAllNamespaces(locale: string, target: Record<string, string>): void {
    const all = i18next.services.resourceStore.data[locale]
    if (!all) return
    Object.keys(all).forEach(n => {
      const loaded = i18next.getResourceBundle(locale, n) || {}
      const flat = flattenMessages(loaded as Record<string, unknown>)
      if (!messages[n]) messages[n] = flat
      mergeMessages(target, loaded as Record<string, unknown>, n)
    })
  }

  const messages: Record<string, Record<string, string>> = {}
  const merged: Record<string, string> = {}
  loadAllNamespaces(lng || 'en', merged)

  const i18n = createI18n({
    legacy: false,
    locale: lng || 'en',
    fallbackLocale: fallbackLng,
    messages: {
      [lng || 'en']: merged,
      ...(fallbackLng && fallbackLng !== lng ? { [fallbackLng as string]: merged } : {}),
    },
    missingWarn: false,
    fallbackWarn: false,
  })

  i18next.on('loaded', () => {
    const locale = i18next.language || lng || 'en'
    const mergedMessages: Record<string, string> = {}
    loadAllNamespaces(locale, mergedMessages)
    if (Object.keys(mergedMessages).length > 0) {
      i18n.global.setLocaleMessage(locale, mergedMessages)
    }
  })

  app.use(i18n)
  return i18n
}
