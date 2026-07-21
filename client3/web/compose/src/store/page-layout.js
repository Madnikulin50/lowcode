import { defineStore } from 'pinia'
import { compose } from 'corteza-lib/js/dist'
import * as request from '../lib/request'
import { getComposeAPI } from './api'

export const usePageLayoutStore = defineStore('pageLayout', {
  state: () => ({
    loading: false,
    pending: false,
    set: [],
  }),

  getters: {
    getByID: (state) => {
      return (ID) => state.set.find(({ pageLayoutID }) => ID === pageLayoutID)
    },

    getByHandle: (state) => {
      return (handle) => state.set.find((pl) => handle === pl.handle)
    },

    getByPageID: (state) => {
      return (ID) => state.set.filter(({ pageID }) => ID === pageID).sort((a, b) => a.weight - b.weight)
    },
  },

  actions: {
    setLoading (val) { this.loading = val },
    setPending (val) { this.pending = val },
    updateSet (set) {
      set = set.map(i => Object.freeze(i))
      if (this.set.length === 0) {
        this.set = set
        return
      }
      set.forEach(newItem => {
        const oldIndex = this.set.findIndex(({ pageLayoutID }) => pageLayoutID === newItem.pageLayoutID)
        if (oldIndex > -1) {
          this.set.splice(oldIndex, 1, newItem)
        } else {
          this.set.push(newItem)
        }
      })
    },
    removeFromSet (removedSet) {
      (removedSet || []).forEach(removedItem => {
        const i = this.set.findIndex(({ pageLayoutID }) => pageLayoutID === removedItem.pageLayoutID)
        if (i > -1) {
          this.set.splice(i, 1)
        }
      })
    },
    clearSet () {
      this.pending = false
      this.set.splice(0)
    },

    async load ({ namespaceID, clear = false, force = false } = {}) {
      if (clear) {
        this.clearSet()
      }
      if (!force && this.set.length > 1) {
        return new Promise((resolve) => resolve(this.set))
      }
      this.setLoading(true)
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.pageLayoutListNamespace({ namespaceID, sort: 'weight ASC' }).then(({ set }) => {
        if (set && set.length > 0) {
          this.updateSet(set.map(pl => new compose.PageLayout(pl)))
        }
        return this.set
      }).finally(() => {
        this.setLoading(false)
        this.setPending(false)
      })
    },

    async findByID ({ namespaceID, pageID, pageLayoutID, force = false } = {}) {
      if (!force) {
        const oldItem = this.getByID(pageLayoutID)
        return new Promise((resolve) => resolve(oldItem))
      }
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.pageLayoutRead({ namespaceID, pageID, pageLayoutID }).then(pl => {
        const pageLayout = new compose.PageLayout(pl)
        this.updateSet([pageLayout])
        return pageLayout
      }).finally(() => {
        this.setPending(false)
      })
    },

    async findByPageID ({ namespaceID, pageID, force = false } = {}) {
      if (!force) {
        const oldItems = this.getByPageID(pageID)
        return new Promise((resolve) => resolve(oldItems))
      }
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.pageLayoutList({ namespaceID, pageID, sort: 'weight ASC' }).then(({ set }) => {
        this.updateSet(set.map(pl => new compose.PageLayout(pl)))
        return set
      }).finally(() => {
        this.setPending(false)
      })
    },

    async create (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.pageLayoutCreate(item, request.config(item)).then(pl => {
        const pageLayout = new compose.PageLayout(pl)
        this.updateSet([pageLayout])
        return pageLayout
      }).finally(() => {
        this.setPending(false)
      })
    },

    async update (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.pageLayoutUpdate(item, request.config(item)).then(pl => {
        const pageLayout = new compose.PageLayout(pl)
        this.updateSet([pageLayout])
        return pageLayout
      }).finally(() => {
        this.setPending(false)
      })
    },

    async delete (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.pageLayoutDelete(item).then(() => {
        this.removeFromSet([item])
        return true
      }).finally(() => {
        this.setPending(false)
      })
    },
  },
})
