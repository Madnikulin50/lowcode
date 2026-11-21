import { defineStore } from 'pinia'

export const useUiStore = defineStore('ui', {
  state: () => ({ loader: 0 }),
  getters: {
    isLoading: (state) => state.loader > 0,
  },
  actions: {
    incLoader() { this.loader++ },
    decLoader() { if (this.loader > 0) this.loader-- },
    hideLoader() { this.loader = 0 },
  },
})
