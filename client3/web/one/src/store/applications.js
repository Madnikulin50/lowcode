import { defineStore } from 'pinia'

export const useApplicationsStore = defineStore('applications', {
  state: () => ({ set: [] }),
  getters: {
    all: (state) => state.set,
    unifyOnly: (state) => state.set.filter(({ unify: { listed } = { listed: false } }) => listed),
  },
  actions: {
    async load() {
      const api = window.__systemAPI
      if (api?.applicationList) {
        const { set } = await api.applicationList({ sort: 'weight', incFlags: 0 })
        this.set = set
      }
    },
    async reorder(applicationIDs) {
      await window.__systemAPI.applicationReorder({ applicationIDs })
      return this.load()
    },
    async pin({ applicationID, ownedBy }) {
      try {
        await window.__systemAPI.applicationFlagCreate({ applicationID, flag: 'pinned', ownedBy })
        return this.load()
      } catch {}
    },
    async unpin({ applicationID, ownedBy }) {
      try {
        await window.__systemAPI.applicationFlagDelete({ applicationID, flag: 'pinned', ownedBy })
        return this.load()
      } catch {}
    },
  },
})
