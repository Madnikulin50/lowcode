import { defineStore } from 'pinia'
export const useLabelsStore = defineStore('labels', {
  state: () => ({ namespaces: {}, modules: {} }),
  getters: {
    getNamespace: (state) => (namespaceID) => state.namespaces[namespaceID],
    getModule: (state) => (moduleID) => state.modules[moduleID],
  },
  actions: {
    setNamespace(namespaceID, name) { this.namespaces[namespaceID] = name },
    setModule(moduleID, name) { this.modules[moduleID] = name },
    async resolveNamespace({ namespaceID, api }) {
      if (this.namespaces[namespaceID] !== undefined) return this.namespaces[namespaceID]
      this.setNamespace(namespaceID, null)
      try { const ns = await api.namespaceRead({ namespaceID }); this.setNamespace(namespaceID, ns.name || ns.slug || namespaceID); return this.namespaces[namespaceID] }
      catch { this.setNamespace(namespaceID, namespaceID); return namespaceID }
    },
    async resolveModule({ moduleID, namespaceID, api }) {
      if (this.modules[moduleID] !== undefined) return this.modules[moduleID]
      this.setModule(moduleID, null)
      try { const mod = await api.moduleRead({ namespaceID, moduleID }); this.setModule(moduleID, mod.name || mod.handle || moduleID); return this.modules[moduleID] }
      catch { this.setModule(moduleID, moduleID); return moduleID }
    },
    async resolveMultipleNamespaces(namespaceIDs, api) { return Promise.all(namespaceIDs.map(id => this.resolveNamespace({ namespaceID: id, api }))) },
    async resolveMultipleModules(modules, api) { return Promise.all(modules.map(({ moduleID, namespaceID }) => this.resolveModule({ moduleID, namespaceID, api }))) },
  },
})
