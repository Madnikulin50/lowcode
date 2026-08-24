import { defineStore } from 'pinia'
import { compose } from 'corteza-lib/js/dist'
import { getComposeAPI } from './api'

function instantiateChart (c) {
  try {
    return new compose.Chart(c)
  } catch (err) {
    console.error('[chart store] failed to instantiate chart', c?.handle || c?.chartID, err)
    return null
  }
}

export const useChartStore = defineStore('chart', {
  state: () => ({
    loading: false,
    pending: false,
    set: [],
  }),

  getters: {
    getByID: (state) => {
      return (ID) => state.set.find(({ chartID }) => ID === chartID)
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
        const oldIndex = this.set.findIndex(({ chartID }) => chartID === newItem.chartID)
        if (oldIndex > -1) {
          this.set.splice(oldIndex, 1, newItem)
        } else {
          this.set.push(newItem)
        }
      })
    },
    removeFromSet (removedSet) {
      (removedSet || []).forEach(removedItem => {
        const i = this.set.findIndex(({ chartID }) => chartID === removedItem.chartID)
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
      return ComposeAPI.chartList({ namespaceID, sort: 'name ASC' }, { timeout: 30000 }).then(({ set }) => {
        if (set && set.length > 0) {
          this.updateSet(set.map(instantiateChart).filter(Boolean))
        }
        return this.set
      }).finally(() => {
        this.setLoading(false)
        this.setPending(false)
      })
    },

    async findByID ({ namespaceID, chartID, force = false } = {}) {
      if (!force) {
        const oldItem = this.getByID(chartID)
        if (oldItem) {
          return new Promise((resolve) => resolve(oldItem))
        }
      }
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.chartRead({ namespaceID, chartID }).then(raw => {
        const chart = new compose.Chart(raw)
        this.updateSet([chart])
        return chart
      }).finally(() => {
        this.setPending(false)
      })
    },

    async create (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.chartCreate(item).then(raw => {
        const chart = new compose.Chart(raw)
        this.updateSet([chart])
        return chart
      }).finally(() => {
        this.setPending(false)
      })
    },

    async update (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.chartUpdate(item).then(raw => {
        const chart = new compose.Chart(raw)
        this.updateSet([chart])
        return chart
      }).finally(() => {
        this.setPending(false)
      })
    },

    async delete (item) {
      this.setPending(true)
      const ComposeAPI = getComposeAPI()
      return ComposeAPI.chartDelete(item).then(() => {
        this.removeFromSet([item])
        return true
      }).finally(() => {
        this.setPending(false)
      })
    },
  },
})
