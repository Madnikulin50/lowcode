import { defineStore } from 'pinia'

export const useEtlStore = defineStore('etl', {
  state: () => ({
    rows: [],
    loading: false,
  }),
  getters: {},
  actions: {
    fetch() {},
  },
})
