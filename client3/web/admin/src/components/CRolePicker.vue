<template>
  <div
    data-test-id="role-picker"
    class="d-flex flex-column"
  >
    <c-input-role
      data-test-id="input-role-picker"
      :selectable="r => !props.value.includes(r.roleID)"
      :placeholder="$t('admin.picker.role.placeholder')"
      :visible="isRoleVisible"
      clear-on-select
      @input="addRole($event)"
    />

    <div
      v-if="preloading"
      class="spinner-border mx-auto my-4"
    />

    <table
      v-else-if="getSelectedRoles.length"
      class="table table-sm table-hover w-100 p-0 mb-0 mt-1"
    >
      <tbody>
        <tr
          v-for="role in getSelectedRoles"
          :key="role.roleID"
          data-test-id="selected-row-list"
        >
          <td class="align-middle">
            {{ getRoleLabel(role) }}
          </td>
          <td class="text-end">
            <c-input-confirm
              data-test-id="button-remove-role"
              show-icon
              @confirmed="removeRole(role.roleID)"
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
const { CInputRole } = components

const { t } = useI18n()

const props = defineProps({
  value: {
    type: Array,
    default: () => ([]),
  },
})

const emit = defineEmits(['input'])

const fetching = ref(false)
const preloading = ref(false)
const filter = ref('')
const selectedRoles = ref([])

const getSelectedRoles = computed(() => {
  return selectedRoles.value.filter(({ roleID }) => props.value.includes(roleID))
})

onMounted(() => {
  preloadSelected()
})

function addRole(role) {
  if (!props.value.includes(role.roleID)) {
    selectedRoles.value.push(role)
    emit('input', [...props.value, role.roleID])
  }
}

function removeRole(roleID) {
  selectedRoles.value = selectedRoles.value.filter(({ roleID: rID }) => rID !== roleID)
  emit('input', props.value.filter(v => v !== roleID))
}

function preloadSelected() {
  if (!props.value.length) {
    return
  }

  preloading.value = true

  return window.__systemAPI.roleList({ roleID: props.value })
    .then(({ set }) => {
      selectedRoles.value = set || []
    })
    .finally(() => {
      preloading.value = false
    })
    .catch(window.__toastError(t('notification.role.fetch.error')))
}

function getRoleLabel({ name, handle, roleID }) {
  return name || handle || roleID
}

function isRoleVisible({ isClosed, meta = {} }) {
  return !(isClosed || (meta.context && meta.context.resourceTypes))
}
</script>

<style lang="scss">
.results {
  z-index: 100;
  .filtered-role {
    cursor: pointer;
    &:hover {
      background-color: var(--light);
    }
  }
}
</style>
