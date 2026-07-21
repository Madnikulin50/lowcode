import { defineStore } from 'pinia'
export const useDiscoveryStore = defineStore('discovery', {
  state: () => ({ processing: false, resourceTypes: ['compose:record'], aggregations: [], modules: [], namespaces: [] }),
  getters: {
    isProcessing: (state) => state.processing,
    resourceTypesList: (state) => state.resourceTypes,
    aggregationsList: (state) => state.aggregations,
    modulesList: (state) => state.modules,
    namespacesList: (state) => state.namespaces,
  },
  actions: {
    async fetchData({ query, modules, namespaces, size } = {}) {
      this.processing = true
      try {
        const response = await window.__discoveryAPI?.query({ query, modules: modules || this.modules, namespaces: namespaces || this.namespaces, size, resourceTypes: this.resourceTypes })
        if (response) this.aggregations = response.aggregations
        return response
      } finally { this.processing = false }
    },
    updateResourceTypes(resourceTypes) { this.resourceTypes = resourceTypes },
    updateModules(modules) { this.modules = modules },
    updateNamespaces(namespaces) { this.namespaces = namespaces },
  },
})