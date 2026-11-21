import moment from 'moment'
import { system } from 'corteza-lib/js/dist'

function intervalToMS (from, to) {
  if (!from || !to) {
    throw new Error('intervalToMS.invalidArgs')
  }
  return to - from
}

export class ReminderService {
  constructor (app, { api, fetchOffset = 1000 * 60 * 5, resource = null } = {}) {
    if (!api) {
      throw new Error('reminderService.invalidParams')
    }

    this.app = app
    this.api = api
    this.fetchOffset = fetchOffset
    this.resource = resource
    this.rootInstance = null

    this.set = []
    this.nextRemindAt = null
    this.tHandle = null
  }

  init ({ filter = {} }) {
    this.filter = {
      scheduledOnly: true,
      excludeDismissed: true,
      ...filter,
    }

    this.prefetch().then(rr => {
      this.enqueue(rr)
    })
  }

  setRootInstance (rootInstance) {
    this.rootInstance = rootInstance
  }

  async prefetch () {
    return this.api.reminderList({
      limit: 0,
      resource: this.resource,
      scheduledUntil: moment().add(this.fetchOffset, 'min').toISOString(),
      ...this.filter,
    }).then(({ set }) => {
      return (set || []).map(r => new system.Reminder(r))
    })
  }

  enqueueRaw (raw) {
    this.enqueue([new system.Reminder(raw)])
  }

  enqueue (set) {
    set.forEach(r => {
      const i = this.set.findIndex(({ reminderID }) => reminderID === r.reminderID)
      if (i > -1) {
        this.set.splice(i, 1, r)
      } else {
        this.set.push(r)
      }
    })

    const { changed, time } = this.findNextProcessTime(this.set, this.nextRemindAt)
    if (changed) {
      this.nextRemindAt = time
      this.scheduleReminderProcess(this.nextRemindAt)
    }
  }

  dequeue (IDs = []) {
    this.set = this.set.filter(({ reminderID }) => !IDs.includes(reminderID))

    const { changed, time } = this.findNextProcessTime(this.set, null)
    if (changed) {
      this.nextRemindAt = time
      this.scheduleReminderProcess(this.nextRemindAt)
    }
  }

  findNextProcessTime (set = [], time = null) {
    let changed = false
    set.forEach(r => {
      if (!r.dismissedAt && (!time || r.remindAt < time)) {
        time = r.remindAt
        changed = true
      }
    })

    return { changed, time }
  }

  scheduleReminderProcess (at, now = new Date()) {
    if (!at) {
      return
    }

    const t = intervalToMS(now, at)

    if (this.tHandle != null) {
      window.clearTimeout(this.tHandle)
    }
    this.tHandle = window.setTimeout(this.processQueue.bind(this), t)
  }

  processQueue (now = new Date()) {
    let nextRemindAt = null

    this.set.forEach(r => {
      if (!r.dismissedAt && now >= r.remindAt) {
        if (this.rootInstance) {
          this.rootInstance.$emit('reminder.show', r)
        } else {
          console.warn('ReminderService: No root instance available to emit reminder.show event')
        }
        r.processed = true
      } else if (now < r.remindAt && (!nextRemindAt || r.remindAt < nextRemindAt)) {
        nextRemindAt = r.remindAt
      }
    })

    this.nextRemindAt = nextRemindAt
    this.set = this.set.filter(({ processed }) => !processed)

    if (this.nextRemindAt === null) {
      this.tHandle = null
    } else {
      this.scheduleReminderProcess(this.nextRemindAt)
    }
  }
}

export default {
  install (app, opts) {
    const reminderService = new ReminderService(app, opts)
    app.config.globalProperties.$Reminder = reminderService

    app.mixin({
      created () {
        if (this.$root === this) {
          reminderService.setRootInstance(this)
        }
      },
    })
  },
}
