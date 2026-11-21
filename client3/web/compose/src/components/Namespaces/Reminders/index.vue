<template>
  <div class="d-flex flex-column h-100">
    <list
      v-if="!edit"
      :reminders="reminders"
      class="flex-fill"
      @edit="onEdit"
      @dismiss="onDismiss"
      @delete="onDelete"
    />

    <edit
      v-else
      :edit="edit"
      :disable-save="disableSave"
      :processing-save="processingSave"
      class="flex-fill"
      @dismiss="onDismiss"
      @back="onCancel()"
      @save="onSave"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue'
import List from './List'
import Edit from './Edit'
import { system, NoID } from 'corteza-lib/js/dist'

const props = defineProps({
  namespaceID: { type: String, default: '' },
})

const gp = getCurrentInstance()?.appContext?.app?.config?.globalProperties || {}
const $SystemAPI = gp.$SystemAPI || window.__systemAPI
const $auth = gp.$auth || window.__auth
const $Reminder = gp.$Reminder || window.__Reminder

const reminders = ref([])
const edit = ref(null)
const disableSave = ref(false)
const processingSave = ref(false)

onMounted(() => {
  fetchReminders()
  window.addEventListener('reminders.pull', fetchReminders)
  window.addEventListener('reminder.updated', fetchReminders)
  window.addEventListener('reminder.create', onEdit)
})

onBeforeUnmount(() => {
  destroyEvents()
  setDefaultValues()
})

function onEdit ({ reminderID, resource, assignedTo, payload = {} } = {}) {
  if (reminderID) {
    edit.value = reminders.value.find(r => r.reminderID === reminderID)
  } else {
    edit.value = {
      resource: resource || `namespace:${props.namespaceID}`,
      assignedTo: assignedTo || $auth.user.userID,
      payload,
    }
  }
  window.dispatchEvent(new CustomEvent('reminders.show'))
}

function onSave (r) {
  processingSave.value = true
  const endpoint = r.reminderID && r.reminderID !== NoID ? 'reminderUpdate' : 'reminderCreate'
  $SystemAPI[endpoint](r).then(() => {
    return fetchReminders()
  }).then(() => {
    onCancel()
  }).finally(() => {
    processingSave.value = false
  })
}

function onCancel () {
  edit.value = undefined
}

function onDismiss ({ reminderID }, value) {
  const endpoint = value ? 'reminderDismiss' : 'reminderUndismiss'
  $SystemAPI[endpoint]({ reminderID }).then(() => {
    fetchReminders()
  })
}

function onDelete ({ reminderID }) {
  $SystemAPI.reminderDelete({ reminderID }).then(() => {
    fetchReminders()
  })
}

async function fetchReminders () {
  return $SystemAPI.reminderList({
    assignedTo: $auth.user.userID,
    limit: 0,
  }).then(({ set: r = [] }) => {
    reminders.value = r.map(rr => new system.Reminder(rr))
    $Reminder.enqueue(reminders.value)
  })
}

function setDefaultValues () {
  reminders.value = []
  edit.value = null
  disableSave.value = false
  processingSave.value = false
}

function destroyEvents () {
  window.removeEventListener('reminders.pull', fetchReminders)
  window.removeEventListener('reminder.updated', fetchReminders)
  window.removeEventListener('reminder.create', onEdit)
}
</script>
