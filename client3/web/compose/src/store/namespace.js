import { defineStore } from 'pinia'
import { compose } from 'corteza-lib/js/dist'
import * as request from '../lib/request'
import { getComposeAPI } from './api'

export const useNamespaceStore = defineStore('namespace', {
  state: () => ({
    loading: false,
    pending: false,
    set: [],
  }),

  getters: {
    getByID: (state) => {
      return (ID) => state.set.find(({ namespaceID }) => ID === namespaceID)
    },

    getByUrlPart: (state) => {
      return (urlPart) => state.set.find(({ slug, namespaceID }) => (urlPart === slug) || (urlPart === namespaceID))
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
        const oldIndex = this.set.findIndex(({ namespaceID }) => namespaceID === newItem.namespaceID)
        if (oldIndex > -1) {
          this.set.splice(oldIndex, 1, newItem)
        } else {
          this.set.push(newItem)
        }
      })
    },
    removeFromSet (removedSet) {
      (removedSet || []).forEach(removedItem => {
        const i = this.set.findIndex(({ namespaceID }) => namespaceID === removedItem.namespaceID)
        if (i > -1) {
          this.set.splice(i, 1)
        }
      })
    },
    clearSet () {
      this.pending = false
      this.set.splice(0)
    },

    async load ({ force = false } = {}) {
      if (!force && this.set.length > 1) {
        return new Promise((resolve) => resolve(this.set))
      }
      this.setLoading(true)
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.namespaceList({}, { timeout: 30000 }).then(({ set }) => {
        if (set && set.length > 0) {
          this.updateSet(set.map(n => new compose.Namespace(n)))
        }
        return this.set
      }).finally(() => {
        this.setLoading(false)
        this.setPending(false)
      })
    },

    async findByID ({ namespaceID, force = false } = {}) {
      if (!force) {
        const oldItem = this.getByID(namespaceID)
        if (oldItem) {
          return new Promise((resolve) => resolve(oldItem))
        }
      }
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.namespaceRead({ namespaceID }).then(raw => {
        const namespace = new compose.Namespace(raw)
        this.updateSet([namespace])
        return namespace
      }).finally(() => {
        this.setPending(false)
      })
    },

    async create (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.namespaceCreate(item, request.config(item)).then(raw => {
        const namespace = new compose.Namespace(raw)
        this.updateSet([namespace])
        return namespace
      }).finally(() => {
        this.setPending(false)
      })
    },

    async clone (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.namespaceClone(item).then(raw => {
        const namespace = new compose.Namespace(raw)
        this.updateSet([namespace])
        return namespace
      }).finally(() => {
        this.setPending(false)
      })
    },

    async update (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.namespaceUpdate(item, request.config(item)).then(raw => {
        const namespace = new compose.Namespace(raw)
        this.updateSet([namespace])
        return namespace
      }).finally(() => {
        this.setPending(false)
      })
    },

    async delete (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.namespaceDelete(item).then(() => {
        this.removeFromSet([item])
        return true
      }).finally(() => {
        this.setPending(false)
      })
    },
  },
})
