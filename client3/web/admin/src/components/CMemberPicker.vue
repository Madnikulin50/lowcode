<template>
  <div
    data-test-id="member-picker"
    class="d-flex flex-column"
  >
    <c-input-user
      data-test-id="input-member-picker"
      :selectable="r => !props.value.includes(r.userID)"
      :placeholder="$t('admin.picker.member.placeholder')"
      clear-on-select
      @input-object="addUser"
    />

    <div
      v-if="preloading"
      class="spinner-border mx-auto my-4"
    />

    <table
      v-else-if="getSelectedUsers.length"
      class="table table-sm table-hover w-100 p-0 mb-0 mt-1"
    >
      <tbody>
        <tr
          v-for="user in getSelectedUsers"
          :key="user.userID"
          data-test-id="selected-row-list"
        >
          <td class="align-middle">
            {{ getUserLabel(user) }}
          </td>
          <td
            v-if="!props.noRemove"
            class="text-end"
          >
            <c-input-confirm
              data-test-id="button-remove-user"
              show-icon
              @confirmed="removeUser(user.userID)"
            />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
const { CInputUser } = components

const { t } = useI18n()

const props = defineProps({
  value: {
    type: Array,
    default: () => ([]),
  },
  noRemove: {
    type: Boolean,
    required: false,
    default: false,
  },
})

const emit = defineEmits(['input'])

const fetching = ref(false)
const preloading = ref(false)
const filter = ref('')
const selectedUsers = ref([])

const getSelectedUsers = computed(() => {
  return selectedUsers.value.filter(({ userID }) => props.value.includes(userID))
})

onMounted(() => {
  preloadSelected()
})

function addUser(user) {
  if (!props.value.includes(user.userID)) {
    selectedUsers.value.push(user)
    emit('input', [...props.value, user.userID])
  }
}

function removeUser(userID) {
  selectedUsers.value = selectedUsers.value.filter(({ userID: rID }) => rID !== userID)
  emit('input', props.value.filter(v => v !== userID))
}

function preloadSelected() {
  if (!props.value.length) {
    return
  }

  preloading.value = true

  return window.__systemAPI.userList({ userID: props.value, suspended: 1, deleted: 1 })
    .then(({ set }) => {
      selectedUsers.value = set || []
    })
    .finally(() => {
      preloading.value = false
    })
    .catch(window.__toastError(t('notification.user.fetch.error')))
}

function getUserLabel({ name, handle, userID, email }) {
  return name || handle || email || userID
}
</script>
