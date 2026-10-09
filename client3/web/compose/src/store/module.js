import { defineStore } from 'pinia'
import { compose } from 'corteza-lib/js/dist'
import * as request from '../lib/request'
import { getComposeAPI } from './api'

export const useModuleStore = defineStore('module', {
  state: () => ({
    loading: false,
    pending: false,
    set: [],
  }),

  getters: {
    getByID: (state) => {
      return (ID) => state.set.find(({ moduleID }) => ID === moduleID)
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
        const oldIndex = this.set.findIndex(({ moduleID }) => moduleID === newItem.moduleID)
        if (oldIndex > -1) {
          this.set.splice(oldIndex, 1, newItem)
        } else {
          this.set.push(newItem)
        }
      })
    },
    removeFromSet (removedSet) {
      (removedSet || []).forEach(removedItem => {
        const i = this.set.findIndex(({ moduleID }) => moduleID === removedItem.moduleID)
        if (i > -1) {
          this.set.splice(i, 1)
        }
      })
    },
    clearSet () {
      this.pending = false
      this.set.splice(0)
    },

    async load ({ namespace, clear = false, force = false } = {}) {
      if (clear) {
        this.clearSet()
      }
      if (!force && this.set.length > 1) {
        return new Promise((resolve) => resolve(this.set))
      }
      this.setLoading(true)
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.moduleList({ namespaceID: namespace.namespaceID, sort: 'name ASC' }).then(({ set }) => {
        if (set && set.length > 0) {
          this.updateSet(set.map(m => new compose.Module(m, namespace)))
        }
        return this.set
      }).finally(() => {
        this.setLoading(false)
        this.setPending(false)
      })
    },

    async findByID ({ namespace, moduleID, force = false } = {}) {
      if (!force) {
        const oldItem = this.getByID(moduleID)
        if (oldItem) {
          return new Promise((resolve) => resolve(oldItem))
        }
      }
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.moduleRead({ namespaceID: namespace.namespaceID, moduleID }).then(raw => {
        const module = new compose.Module(raw, namespace)
        this.updateSet([module])
        return module
      }).finally(() => {
        this.setPending(false)
      })
    },

    async create (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.moduleCreate(item, request.config(item)).then(raw => {
        const module = new compose.Module(raw, raw.namespace)
        this.updateSet([module])
        return module
      }).finally(() => {
        this.setPending(false)
      })
    },

    async update (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.moduleUpdate(item, request.config(item)).then(raw => {
        const module = new compose.Module(raw, raw.namespace)
        this.updateSet([module])
        return module
      }).finally(() => {
        this.setPending(false)
      })
    },

    async delete (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.moduleDelete(item).then(() => {
        this.removeFromSet([item])
        return true
      }).finally(() => {
        this.setPending(false)
      })
    },
  },
})
