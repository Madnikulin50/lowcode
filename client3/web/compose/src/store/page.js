import { defineStore } from 'pinia'
import { compose } from 'corteza-lib/js/dist'
import * as request from '../lib/request'
import { getComposeAPI } from './api'
import { restoreBlockHelp } from '../help/appDocs'

function pageFromRaw (raw) {
  const page = new compose.Page(raw)
  restoreBlockHelp(page, raw)
  return page
}

export const usePageStore = defineStore('page', {
  state: () => ({
    loading: false,
    pending: false,
    set: [],
  }),

  getters: {
    getByID: (state) => {
      return (ID) => state.set.find(({ pageID }) => ID === pageID)
    },

    getByHandle: (state) => {
      return (handle) => state.set.find((p) => handle === p.handle)
    },

    homePage: (state) => state.set.find(p => p.visible && p.firstLevel && !p.isRecordPage),
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
        const oldIndex = this.set.findIndex(({ pageID }) => pageID === newItem.pageID)
        if (oldIndex > -1) {
          this.set.splice(oldIndex, 1, newItem)
        } else {
          this.set.push(newItem)
        }
      })
    },
    removeFromSet (removedSet) {
      (removedSet || []).forEach(removedItem => {
        const i = this.set.findIndex(({ pageID }) => pageID === removedItem.pageID)
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
      return ComposeAPI.pageList({ namespaceID, sort: 'weight ASC' }).then(({ set }) => {
        if (set && set.length > 0) {
          this.updateSet(set.map(p => pageFromRaw(p)))
        }
        return this.set
      }).finally(() => {
        this.setLoading(false)
        this.setPending(false)
      })
    },

    async findByID ({ namespaceID, pageID, force = false } = {}) {
      if (!force) {
        const oldItem = this.getByID(pageID)
        if (oldItem) {
          return new Promise((resolve) => resolve(oldItem))
        }
      }
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.pageRead({ namespaceID, pageID }).then(raw => {
        const page = pageFromRaw(raw)
        this.updateSet([page])
        return page
      }).finally(() => {
        this.setPending(false)
      })
    },

    async create (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.pageCreate(item, request.config(item)).then(raw => {
        const page = pageFromRaw(raw)
        this.updateSet([page])
        return page
      }).finally(() => {
        this.setPending(false)
      })
    },

    async update (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.pageUpdate(item, request.config(item)).then(raw => {
        const page = pageFromRaw(raw)
        this.updateSet([page])
        return page
      }).finally(() => {
        this.setPending(false)
      })
    },

    async delete (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.pageDelete(item).then(() => {
        this.removeFromSet([item])
        const { namespaceID } = item || {}
        if (namespaceID) {
          this.load({ namespaceID: item.namespaceID, clear: true })
        }
        return true
      }).finally(() => {
        this.setPending(false)
      })
    },
  },
})
