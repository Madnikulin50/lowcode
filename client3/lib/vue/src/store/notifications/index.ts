import { defineStore } from 'pinia'
import { apiClients, system } from '@cortezaproject/corteza-js'

export const useNotificationsStore = defineStore('notifications', {
  state: () => ({
    notifications: [] as Array<system.Notification>,
    visible: false,
    pageCursor: null as string | null,
    muted: localStorage.getItem('notificationsMuted') === 'true' || false,
  }),

  getters: {
    hasUnread: (state) => state.notifications.some(notification => !notification.readAt),
    hasRead: (state) => state.notifications.some(notification => !!notification.readAt),
    unreadCount: (state) => state.notifications.filter(notification => !notification.readAt).length,
    hasMorePages: (state) => !!state.pageCursor,
  },

  actions: {
    toggleVisibility () {
      this.visible = !this.visible
    },

    async fetchNotifications ({ unreadOnly = true } = {}) {
      const api = (window as any).__systemAPI as apiClients.System | undefined
      if (!api) return
      return api.notificationList({
        limit: 25,
        sort: this.pageCursor ? '' : 'createdAt DESC, readAt DESC',
        read: unreadOnly ? 0 : 1,
        pageCursor: this.pageCursor,
      }).then((response: any) => {
        const set = (response.set || []) as Array<system.Notification>
        if (this.pageCursor) {
          this.notifications = [...this.notifications, ...set.map((n: system.Notification) => new system.Notification(n))]
        } else {
          this.notifications = set.map((n: system.Notification) => new system.Notification(n))
        }
        const filter = (response.filter || {}) as { nextPage?: string }
        this.pageCursor = filter.nextPage || null
      })
    },

    async markAsRead (notificationID: string) {
      const api = (window as any).__systemAPI as apiClients.System | undefined
      if (!api) return
      return api.notificationMarkAsRead({ notificationID })
        .then(() => {
          const n = this.notifications.find(x => String(x.notificationID) === String(notificationID))
          if (n) n.readAt = new Date()
        })
    },

    async markAsUnread (notificationID: string) {
      const api = (window as any).__systemAPI as any
      if (!api) return
      return api.notificationMarkAsUnread({ notificationID })
        .then(() => {
          const n = this.notifications.find(x => String(x.notificationID) === String(notificationID))
          if (n) n.readAt = undefined
        })
    },

    async markAllAsRead () {
      if (!this.notifications.length) return Promise.resolve()
      const api = (window as any).__systemAPI as apiClients.System | undefined
      if (!api) return
      return api.notificationMarkAllAsRead()
        .then(() => {
          const now = new Date()
          this.notifications.forEach(n => { if (!n.readAt) n.readAt = now })
        })
    },

    async markAllAsUnread () {
      if (!this.notifications.length) return Promise.resolve()
      const api = (window as any).__systemAPI as any
      if (!api) return
      return api.notificationMarkAllAsUnread()
        .then(() => {
          this.notifications.forEach(n => { n.readAt = undefined })
        })
    },

    addNotification (notification: system.Notification) {
      this.notifications.unshift(new system.Notification(notification))
    },

    async deleteNotification (notificationID: string) {
      const api = (window as any).__systemAPI as apiClients.System | undefined
      if (!api) return
      return api.notificationDelete({ notificationID })
        .then(() => {
          this.notifications = this.notifications.filter(n => String(n.notificationID) !== String(notificationID))
        })
    },

    setPageCursor (pageCursor: string | null) {
      this.pageCursor = pageCursor
    },

    toggleMuted () {
      this.muted = !this.muted
      localStorage.setItem('notificationsMuted', String(this.muted))
    },

    updateReadNotification (notification: system.Notification) {
      const existing = this.notifications.find(n =>
        String(n.notificationID) === String(notification.notificationID),
      )
      if (!existing) return
      const hasReadAt = Object.prototype.hasOwnProperty.call(notification, 'readAt')
      if (hasReadAt) {
        existing.readAt = notification.readAt ? new Date(notification.readAt as any) : undefined
      } else {
        existing.readAt = new Date()
      }
    },

    updateUnreadNotification (notification: system.Notification) {
      const existing = this.notifications.find(n =>
        String(n.notificationID) === String(notification.notificationID),
      )
      if (existing) existing.readAt = undefined
    },

    updateAllReadNotifications (notifications: Array<system.Notification>) {
      const now = new Date()
      if (Array.isArray(notifications) && notifications.length > 0) {
        const ids = notifications.map(n => String(n.notificationID))
        this.notifications.forEach(n => {
          if (ids.includes(String(n.notificationID))) n.readAt = now
        })
      } else {
        this.notifications.forEach(n => { if (!n.readAt) n.readAt = now })
      }
    },

    updateAllUnreadNotifications (notifications: Array<system.Notification>) {
      if (Array.isArray(notifications) && notifications.length > 0) {
        const ids = notifications.map(n => String(n.notificationID))
        this.notifications.forEach(n => {
          if (ids.includes(String(n.notificationID))) n.readAt = undefined
        })
      } else {
        this.notifications.forEach(n => { n.readAt = undefined })
      }
    },

    removeNotification (notification: system.Notification) {
      this.notifications = this.notifications.filter(n => n.notificationID !== notification.notificationID)
    },
  },
})
