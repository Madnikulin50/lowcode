import { defineStore } from 'pinia'
import { getSystemAPI } from './api'

export const useLanguagesStore = defineStore('languages', {
  state: () => ({
    pending: false,
    set: [],
  }),

  getters: {
    default: (state) => state.set.length > 0 ? state.set[0] : undefined,
  },

  actions: {
    setPending (val) { this.pending = val },

    async load () {
      this.setPending(true)
      const SystemAPI = getSystemAPI()
      return SystemAPI.localeList().then(({ set }) => {
        this.set = set
      }).finally(() => {
        this.setPending(false)
      })
    },
  },
})
