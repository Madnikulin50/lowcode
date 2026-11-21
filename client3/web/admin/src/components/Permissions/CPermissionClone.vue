<template>
  <div class="d-inline-block">
    <button
      type="button"
      class="btn btn-outline-secondary me-2"
      data-test-id="button-clone"
      @click="showModal = true"
    >
      {{ $t('ui.clone.label') }}
    </button>

    <div
      class="modal fade"
      :class="{ show: showModal }"
      :style="{ display: showModal ? 'block' : 'none' }"
      tabindex="-1"
      @click.self="onModalHide"
    >
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ $t('ui.clone.title') }}</h5>
            <button type="button" class="btn-close" @click="onModalHide" />
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <small class="form-text text-muted">{{ $t('ui.clone.description') }}</small>
              <c-input-role
                v-model="selectedRoles"
                data-test-id="select-role-list"
                :selectable="r => !selectedRoles.some(rr => rr.roleID === r.roleID)"
                :placeholder="$t('ui.clone.pick-role')"
                multiple
              />
            </div>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-outline-secondary"
              @click="onModalHide"
            >
              {{ $t('label.cancel') }}
            </button>
            <button
              type="button"
              class="btn btn-primary"
              :disabled="!selectedRoles.length || processingSubmit"
              @click="clonePermissions"
            >
              {{ $t('ui.clone.clone') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="showModal"
      class="modal-backdrop fade show"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'permissions' } })
import { ref } from 'vue'
import { components } from 'corteza-lib/vue/dist'

const { CInputRole } = components

defineProps({
  roleId: { type: String, required: false, default: undefined },
})

const showModal = ref(false)
const selectedRoles = ref([])
const processingSubmit = ref(false)

function onModalHide() {
  showModal.value = false
}

function clonePermissions() {
  processingSubmit.value = true

  const cloneToRoleID = selectedRoles.value.map(({ roleID }) => roleID)

  window.__SystemAPI.roleCloneRules({ roleID: props.roleId, cloneToRoleID })
    .then(() => {
      selectedRoles.value = []
      window.__toastSuccess(t('notification.permissions.clone.success'))
    })
    .catch(() => {
      window.__toastError(t('notification.permissions.clone.error'))
    })
    .finally(() => {
      processingSubmit.value = false
    })
}
</script>