import { defineStore } from 'pinia'
import { getComposeAPI } from './api'

const pendingBatches = new Map()
const inflightIDs = new Set()
let flushTimer = null

function flushResolves (store) {
  const batches = new Map(pendingBatches)
  pendingBatches.clear()

  for (const [, { ids, resolvers, namespaceID, moduleID }] of batches) {
    const recordIDs = [...ids]

    if (recordIDs.length === 0) {
      resolvers.forEach(r => r())
      continue
    }

    recordIDs.forEach(id => inflightIDs.add(id))
    store.setPending(true)

    const query = recordIDs.map(id => `recordID = ${id}`).join(' OR ')
    const ComposeAPI = getComposeAPI()

    ComposeAPI.recordList({ namespaceID, moduleID, query, deleted: 1 })
      .then(({ set }) => {
        store.updateSet(set)
      })
      .finally(() => {
        recordIDs.forEach(id => inflightIDs.delete(id))
        store.setPending(false)
        resolvers.forEach(r => r())
      })
  }
}

export const useRecordStore = defineStore('record', {
  state: () => ({
    pending: false,
    set: [],
  }),

  getters: {
    findByID: (state) => {
      return (ID) => state.set.find(({ recordID }) => ID === recordID)
    },

    findByIDs: (state) => {
      return (IDs) => {
        const idSet = new Set(IDs.flat())
        return state.set.filter(({ recordID }) => idSet.has(recordID))
      }
    },
  },

  actions: {
    setPending (val) { this.pending = val },

    updateSet (set) {
      set = (Array.isArray(set) ? set : [set]).filter(r => !!r)

      if (this.set.length === 0) {
        this.set = set.map(r => JSON.parse(JSON.stringify(r)))
        return
      }

      const indexByID = new Map(this.set.map(({ recordID }, i) => [recordID, i]))

      set.forEach(newItem => {
        newItem = JSON.parse(JSON.stringify(newItem))

        const oldIndex = indexByID.get(newItem.recordID)
        if (oldIndex !== undefined) {
          this.set.splice(oldIndex, 1, newItem)
        } else {
          indexByID.set(newItem.recordID, this.set.length)
          this.set.push(newItem)
        }
      })
    },

    clearSet () {
      this.pending = false
      this.set.splice(0)

      inflightIDs.clear()
      clearTimeout(flushTimer)
      pendingBatches.clear()
    },

    resolveRecords ({ namespaceID, moduleID, recordIDs }) {
      if (recordIDs.length === 0) {
        return Promise.resolve()
      }

      const knownIDs = new Set(this.set.map(({ recordID }) => recordID))
      recordIDs = recordIDs.filter(id => !knownIDs.has(id) && !inflightIDs.has(id))

      if (recordIDs.length === 0) {
        return Promise.resolve()
      }

      const key = `${namespaceID}/${moduleID}`

      if (!pendingBatches.has(key)) {
        pendingBatches.set(key, { ids: new Set(), resolvers: [], namespaceID, moduleID })
      }

      const batch = pendingBatches.get(key)
      recordIDs.forEach(id => batch.ids.add(id))

      const promise = new Promise(resolve => {
        batch.resolvers.push(resolve)
      })

      clearTimeout(flushTimer)
      flushTimer = setTimeout(() => flushResolves(this), 50)

      return promise
    },

    updateRecords (records) {
      this.updateSet(records)
    },

    push (record) {
      this.updateSet(record)
    },
  },
})
