<template>
  <div class="card shadow-sm" data-test-id="card-user-info">
    <div class="card-header border-bottom">
      <h4 class="ae-section-title">{{ $t('title') }}</h4>
    </div>
    <div class="card-body">
      <form
        class="user-info-form"
        @submit.prevent="$emit('submit', user)"
      >
      <div class="row g-3">
        <div class="col-12 col-lg-6">
          <label class="form-label">{{ $t('email') }}</label>
          <input
            v-model="user.email"
            data-test-id="input-email"
            required
            :class="['form-control', { 'is-invalid': emailState === false }]"
            type="email"
          >
        </div>

        <div class="col-12 col-lg-6">
          <label class="form-label">{{ $t('name') }}</label>
          <input
            v-model="user.name"
            data-test-id="input-name"
            required
            class="form-control"
          >
        </div>

        <div class="col-12 col-lg-6">
          <label class="form-label">{{ $t('handle') }}</label>
          <input
            v-model="user.handle"
            data-test-id="input-handle"
            :placeholder="$t('system.users.editor.info.placeholder-handle')"
            :class="['form-control', { 'is-invalid': handleState === false }]"
          >
          <div v-if="handleState === false" class="invalid-feedback">
            {{ $t('system.users.editor.info.invalid-handle-characters') }}
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <label class="form-label">{{ $t('system.users.editor.info.userGroup.label') }}</label>
          <c-input-user-group
            v-model="user.userGroupID"
            :placeholder="$t('system.users.editor.info.userGroup.placeholder')"
          />
        </div>
      </div>

      <c-system-fields
        :id="user.userID"
        :resource="user"
      />

      <input
        type="submit"
        class="d-none"
        :disabled="saveDisabled"
      >
    </form>
    </div>
    <Teleport v-if="toolbarReady" to="#admin-toolbar">
      <div class="ae-toolbar">
        <c-input-confirm
          v-if="!fresh && user.canDeleteUser"
          :data-test-id="deletedButtonStatusCypressId"
          :text="getDeleteStatus"
          variant="danger"
          size="md"
          @confirmed="$emit('delete')"
        />

        <c-input-confirm
          v-if="!fresh"
          :data-test-id="suspendButtonStatusCypressId"
          :text="getSuspendStatus"
          variant="outline-secondary"
          size="md"
          @confirmed="$emit('status')"
        />

        <c-input-confirm
          v-if="!fresh"
          data-test-id="button-sessions-revoke"
          :text="$t('system.users.editor.info.revokeAllSession')"
          :disabled="user.userID === currentUserID"
          variant="outline-secondary"
          size="md"
          @confirmed="$emit('sessionsRevoke')"
        />

        <button
          v-if="!fresh && !user.emailConfirmed"
          class="btn btn-outline-secondary"
          @click="$emit('patch', '/emailConfirmed', true)"
        >
          {{ $t('system.users.editor.info.confirmEmail') }}
        </button>

        <c-corredor-manual-buttons
          ui-page="user/editor"
          ui-slot="infoFooter"
          resource-type="system:user"
          default-variant="outline-secondary"
          @click="dispatchCortezaSystemUserEvent($event, { user })"
        />

        <c-button-submit
          :disabled="saveDisabled"
          :processing="processing"
          :success="success"
          :text="$t('label.submit')"
          class="ms-auto"
          @submit="$emit('submit', user)"
        />
      </div>
    </Teleport>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.users', keyPrefix: 'editor.info' } })
import { computed, inject, onMounted, ref } from 'vue'

import { NoID } from 'corteza-lib/js/dist'
import { handle, components, useNsI18n } from 'corteza-lib/vue/dist'
const { CInputUserGroup } = components

const t = useNsI18n()
const $auth = inject('auth', {})
const toolbarReady = ref(false)
onMounted(() => { toolbarReady.value = true })

const props = defineProps({
  user: { type: Object, required: true },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  canCreate: { type: Boolean, required: true },
})

defineEmits(['submit', 'delete', 'status', 'patch', 'sessionsRevoke'])

const currentUserID = computed(() => $auth?.user?.userID)

const getDeleteStatus = computed(() => props.user.deletedAt ? t('undelete') : t('delete'))
const getSuspendStatus = computed(() => props.user.suspendedAt ? t('unsuspend') : t('suspend'))
const fresh = computed(() => !props.user.userID || props.user.userID === NoID)
const editable = computed(() => fresh.value ? props.canCreate : props.user.canUpdateUser)
const emailState = computed(() => {
  const { email } = props.user
  return email ? null : false
})
const handleState = computed(() => handle.handleState(props.user.handle))
const saveDisabled = computed(() => !editable.value || [emailState.value, handleState.value].includes(false))
const deletedButtonStatusCypressId = computed(() => `button-${getDeleteStatus.value.toLowerCase()}`)
const suspendButtonStatusCypressId = computed(() => `button-${getSuspendStatus.value.toLowerCase()}`)

function dispatchCortezaSystemUserEvent($event, { user }) {}
</script>
