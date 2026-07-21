import { defineStore } from 'pinia'
import { system } from 'corteza-lib/js/dist'
import { getSystemAPI } from './api'

export const useUserStore = defineStore('user', {
  state: () => ({
    pending: false,
    set: [],
  }),

  getters: {
    findByID: (state) => {
      return (ID) => state.set.find(({ userID }) => ID === userID)
    },

    findByUsername: (state) => (username) => {
      return state.set.filter(user => user.username === username)[0] || undefined
    },
  },

  actions: {
    setPending (val) { this.pending = val },

    updateSet (set) {
      set = (Array.isArray(set) ? set : [set]).filter(u => !!u).map(i => new system.User(i))

      if (this.set.length === 0) {
        this.set = set
        return
      }

      set.forEach(newItem => {
        const oldIndex = this.set.findIndex(({ userID }) => userID === newItem.userID)
        if (oldIndex > -1) {
          this.set.splice(oldIndex, 1, newItem)
        } else {
          this.set.push(newItem)
        }
      })
    },

    async load (filter) {
      this.setPending(true)
      const SystemAPI = getSystemAPI()
      return SystemAPI.userList(filter).then(({ set }) => {
        this.updateSet(set)
      }).finally(() => {
        this.setPending(false)
      })
    },

    push (user) {
      this.updateSet(user)
    },

    async fetchUsers (userID) {
      this.setPending(true)

      if (userID.length === 0) {
        return null
      }

      const SystemAPI = getSystemAPI()
      return SystemAPI.userList({ userID }).then(({ set }) => {
        this.updateSet(set)
      }).finally(() => {
        this.setPending(false)
      })
    },

    async resolveUsers (list) {
      if (list.length === 0) {
        return
      }

      const existing = new Set(this.set.map(({ userID }) => userID))
      list = [...new Set(list.filter(userID => userID && !existing.has(userID)))]

      if (list.length === 0) {
        return
      }

      this.setPending(true)
      const SystemAPI = getSystemAPI()
      return SystemAPI.userList({ userID: list, suspended: 1, deleted: 1 }).then(({ set }) => {
        this.updateSet(set)
      }).finally(() => {
        this.setPending(false)
      })
    },
  },
})
