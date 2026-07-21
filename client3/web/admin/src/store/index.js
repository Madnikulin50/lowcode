import { defineStore } from 'pinia'
import { store as cvStore } from 'corteza-lib/vue/dist'
import { useUiStore } from './ui'

const useRbacStore = defineStore('rbac', {
  state: () => ({
    isLoaded: false,
    rules: [],
  }),
  getters: {
    can: (state) => (resource, operation) => {
      // simplified: always true for now
      return true
    },
  },
  actions: {
    load(apis) {
      this.isLoaded = true
    },
  },
})

const useWfPromptsStore = defineStore('wfPrompts', {
  state: () => ({
    prompts: [],
  }),
  actions: {
    new(prompt) {
      this.prompts.push(prompt)
    },
    clear(value) {
      this.prompts = this.prompts.filter(p => p !== value)
    },
    update() {},
  },
})

const useNotificationsStore = defineStore('notifications', {
  state: () => ({
    notifications: [],
  }),
  actions: {
    addNotification(n) {
      this.notifications.push(n)
    },
    removeNotification(n) {
      this.notifications = this.notifications.filter(x => x !== n)
    },
    updateReadNotification(n) {},
    updateUnreadNotification(n) {},
    updateAllReadNotifications(n) {},
    updateAllUnreadNotifications(n) {},
    fetchNotifications() {},
  },
})

export function useStore() {
  return {
    ui: useUiStore(),
    rbac: useRbacStore(),
    wfPrompts: useWfPromptsStore(),
    notifications: useNotificationsStore(),
  }
}
