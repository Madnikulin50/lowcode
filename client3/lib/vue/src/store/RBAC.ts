import { defineStore } from 'pinia'

const resourcePrefix = 'corteza::'

interface Rule {
  resource: string
  operation: string
  allow: boolean
}

interface Fetcher {
  permissionsEffective: () => Promise<Rule[]>
}

export const useRBACStore = defineStore('rbac', {
  state: () => ({
    loaded: false,
    rules: [] as Rule[],
  }),

  getters: {
    can: (state) => {
      return (res: string, op: string): boolean => {
        return (state.rules.find(
          ({ resource, operation }) => resource === resourcePrefix + res && op === operation
        ) || { allow: false }).allow
      }
    },

    isLoaded: (state) => state.loaded,
  },

  actions: {
    load(...apis: Fetcher[]) {
      this.loaded = false
      Promise.all(apis.map(api =>
        api.permissionsEffective().catch(() => [])
      )).then((rr: Rule[][]) => {
        this.rules = ([] as Rule[])
          .concat(...rr)
          .filter(({ resource }) => resource.startsWith(resourcePrefix))
        this.loaded = true
      })
    },
  },
})
