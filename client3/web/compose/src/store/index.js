import { useNamespaceStore } from './namespace'
import { useModuleStore } from './module'
import { usePageStore } from './page'
import { usePageLayoutStore } from './page-layout'
import { useChartStore } from './chart'
import { useRecordStore } from './record'
import { useUserStore } from './user'
import { useEtlStore } from './etl'
import { useLanguagesStore } from './languages'
import { useUiStore } from './ui'
import { useRBACStore } from 'corteza-lib/vue/dist'
import { useWfPromptsStore } from 'corteza-lib/vue/dist'
import { useNotificationsStore } from 'corteza-lib/vue/dist'
import { useDraftsStore } from 'corteza-lib/vue/dist'

function makeGettersProxy(stores) {
  return new Proxy({}, {
    get(_, key) {
      const [ns, getterName] = String(key).split('/')
      if (ns && getterName && stores[ns]) {
        return stores[ns][getterName]
      }
      return undefined
    },
  })
}

function makeDispatch(stores) {
  return (path, ...args) => {
    const [ns, action] = String(path).split('/')
    if (ns && action && stores[ns] && stores[ns][action]) {
      return stores[ns][action](...args)
    }
    throw new Error(`Unknown store action: ${path}`)
  }
}

export function useStore() {
  const stores = {
    namespace: useNamespaceStore(),
    module: useModuleStore(),
    page: usePageStore(),
    pageLayout: usePageLayoutStore(),
    chart: useChartStore(),
    record: useRecordStore(),
    user: useUserStore(),
    languages: useLanguagesStore(),
    etl: useEtlStore(),
    ui: useUiStore(),
    rbac: useRBACStore(),
    wfPrompts: useWfPromptsStore(),
    notifications: useNotificationsStore(),
    drafts: useDraftsStore(),
  }

  return {
    ...stores,
    getters: makeGettersProxy(stores),
    dispatch: makeDispatch(stores),
  }
}
