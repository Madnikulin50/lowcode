<template>
  <div
    v-if="user"
    class="container pt-2 pb-3"
  >
    <c-content-header :title="title">
      <button
        v-if="userID && canCreate"
        data-test-id="button-new-user"
        class="btn btn-primary"
        @click="$router.push({ name: 'system.user.new' })"
      >
        {{ $t('system.users.editor.new') }}
      </button>

      <c-permissions-button
        v-if="userID && canGrant"
        :title="user.name || user.handle || user.email || userID"
        :target="user.name || user.handle || user.email || userID"
        :resource="`corteza::system:user/${userID}`"
      >
        <font-awesome-icon :icon="['fas', 'lock']" />
        {{ $t('system.users.editor.permissions') }}
      </c-permissions-button>

      <c-corredor-manual-buttons
        ui-page="user/editor"
        ui-slot="toolbar"
        resource-type="system:user"
        default-variant="link"
        class="me-1"
        @click="dispatchCortezaSystemUserEvent($event, { user })"
      />
    </c-content-header>

    <c-user-editor-info
      :user="user"
      :processing="info.processing"
      :success="info.success"
      :can-create="canCreate"
      @submit="onInfoSubmit"
      @delete="onDelete"
      @status="onStatusChange"
      @patch="onPatch"
      @sessions-revoke="onSessionsRevoke"
    />

    <c-user-editor-avatar
      v-if="user && userID && $Settings.get('auth.internal.profile-avatar.Enabled', false)"
      :user="user"
      :processing-avatar="avatar.processing"
      :success="avatar.success"
      class="mt-3"
      @submit="onAvatarSubmit"
      @on-upload="onAvatarUpload"
      @reset-attachment="onResetAvatar"
    />

    <c-user-editor-roles
      v-if="user && userID && membership.active"
      v-model="membership.active"
      class="mt-3"
      :processing="roles.processing"
      :success="roles.success"
      @submit="onMembershipSubmit"
    />

    <c-user-editor-mfa
      v-if="user && userID"
      class="mt-3"
      :mfa="user.meta.securityPolicy.mfa"
      :processing="mfa.processing"
      :success="mfa.success"
      @patch="onPatch"
    />

    <c-user-editor-password
      v-if="user && userID"
      class="mt-3"
      :processing="password.processing"
      :success="password.success"
      :user-i-d="userID"
      @submit="onPasswordSubmit"
    />

    <c-user-editor-external-auth-providers
      v-if="user && userID"
      class="mt-3"
      :value="externalAuthProviders"
      @delete="onExternalAuthProviderDelete"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.users', keyPrefix: 'editor' } })
import { ref, reactive, computed, watch, onMounted, inject } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NoID, system } from 'corteza-lib/js/dist'
import { isEqual } from 'lodash'
import { useUiStore } from '../../../store/ui'
import CUserEditorExternalAuthProviders from '../../../components/User/CUserEditorExternalAuthProviders.vue'
import CUserEditorInfo from '../../../components/User/CUserEditorInfo.vue'
import CUserEditorMfa from '../../../components/User/CUserEditorMFA.vue'
import CUserEditorPassword from '../../../components/User/CUserEditorPassword.vue'
import CUserEditorRoles from '../../../components/User/CUserEditorRoles.vue'
import CUserEditorAvatar from '../../../components/User/CUserEditorAvatar.vue'

const props = defineProps({
  userID: { type: String, required: false, default: undefined },
})

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const ui = useUiStore()
const $auth = inject('auth', {})
const $Settings = inject('$Settings', {})

const user = ref(undefined)
const initialUserState = ref(undefined)
const membership = reactive({ active: undefined, initial: undefined })
const externalAuthProviders = ref([])
const info = reactive({ processing: false, success: false })
const avatar = reactive({ processing: false, success: false })
const password = reactive({ processing: false, success: false })
const mfa = reactive({ processing: false, success: false })
const roles = reactive({ processing: false, success: false })

const canCreate = computed(() => can('system/', 'user.create'))
const canGrant = computed(() => can('system/', 'grant'))
const title = computed(() => props.userID ? t('system.users.editor.title.edit') : t('system.users.editor.title.create'))

function can(resource, operation) { return true }

watch(() => props.userID, (userID) => {
  if (userID) {
    fetchUser()
    fetchMembership()
    fetchExternalAuthProviders()
  } else {
    user.value = new system.User()
    initialUserState.value = user.value.clone()
  }
}, { immediate: true })

function incLoader() { ui.incLoader() }
function decLoader() { ui.decLoader() }

function fetchUser() {
  incLoader()
  return window.__systemAPI.userRead({ userID: props.userID })
    .then(u => {
      user.value = new system.User(u)
      initialUserState.value = user.value.clone()
    })
    .finally(() => decLoader())
}

function fetchMembership() {
  incLoader()
  return window.__systemAPI.userMembershipList({ userID: props.userID })
    .then((set = []) => {
      membership.active = [...set]
      membership.initial = [...set]
    })
    .finally(() => decLoader())
}

function fetchExternalAuthProviders() {
  incLoader()
  return window.__systemAPI.userListCredentials({ userID: props.userID })
    .then((providers = []) => {
      externalAuthProviders.value = providers.map(({ credentialsID = '', label = '', kind = '' }) => ({ credentialsID, label, type: kind }))
    })
    .finally(() => decLoader())
}

function onInfoSubmit(u) {
  info.processing = true
  const payload = { ...u }
  if (payload.userID !== NoID) {
    window.__systemAPI.userUpdate(payload).then(u => {
      user.value = new system.User(u)
      initialUserState.value = user.value.clone()
      info.success = true
      setTimeout(() => { info.success = false }, 2000)
    }).finally(() => { info.processing = false })
  } else {
    window.__systemAPI.userCreate(payload).then(({ userID }) => {
      info.success = true
      setTimeout(() => { info.success = false }, 2000)
      router.push({ name: 'system.user.edit', params: { userID } })
    }).finally(() => { info.processing = false })
  }
}

function onAvatarSubmit(u) {
  avatar.processing = true
  const payload = { userID: u.userID, avatarColor: u.meta.avatarColor, avatarBgColor: u.meta.avatarBgColor }
  window.__systemAPI.userProfileAvatarInitial(payload).then(() => fetchUser()).then(() => {
    avatar.success = true
    setTimeout(() => { avatar.success = false }, 2000)
  }).finally(() => { avatar.processing = false })
}

function onDelete() {
  incLoader()
  if (user.value.deletedAt) {
    window.__systemAPI.userUndelete({ userID: props.userID }).then(() => fetchUser()).finally(() => decLoader())
  } else {
    window.__systemAPI.userDelete({ userID: props.userID }).then(() => {
      fetchUser()
      user.value.deletedAt = new Date()
      router.push({ name: 'system.user' })
    }).finally(() => decLoader())
  }
}

function onExternalAuthProviderDelete(credentialsID = '') {
  incLoader()
  window.__systemAPI.userDeleteCredentials({ userID: props.userID, credentialsID })
    .then(() => fetchExternalAuthProviders())
    .finally(() => decLoader())
}

function onPasswordSubmit(pwd = '') {
  password.processing = true
  window.__systemAPI.userSetPassword({ userID: props.userID, password: pwd })
    .then(() => {
      password.success = true
      setTimeout(() => { password.success = false }, 2000)
      fetchExternalAuthProviders()
    })
    .finally(() => { password.processing = false })
}

function onPatch(path, value) {
  const cfg = { method: 'patch', url: window.__systemAPI.userPartialUpdateEndpoint({ userID: props.userID }), data: [{ path, value, op: 'replace' }] }
  return window.__systemAPI.api().request(cfg).then(() => fetchUser())
}

function onMembershipSubmit() {
  roles.processing = true
  const userID = props.userID
  Promise.all([
    ...membership.initial.filter(roleID => !membership.active.includes(roleID)).map(roleID => window.__systemAPI.userMembershipRemove({ roleID, userID })),
    ...membership.active.filter(roleID => !membership.initial.includes(roleID)).map(roleID => window.__systemAPI.userMembershipAdd({ roleID, userID })),
  ]).then(() => {
    roles.success = true
    setTimeout(() => { roles.success = false }, 2000)
    fetchMembership()
  }).finally(() => { roles.processing = false })
}

function onStatusChange() {
  incLoader()
  if (user.value.suspendedAt) {
    window.__systemAPI.userUnsuspend({ userID: props.userID }).then(() => fetchUser()).finally(() => decLoader())
  } else {
    window.__systemAPI.userSuspend({ userID: props.userID }).then(() => fetchUser()).finally(() => decLoader())
  }
}

function onSessionsRevoke() {
  incLoader()
  window.__systemAPI.userSessionsRemove({ userID: props.userID }).then(() => fetchUser()).finally(() => decLoader())
}

function onAvatarUpload() {
  fetchUser()
}

function onResetAvatar() {
  avatar.processing = true
  window.__systemAPI.userDeleteAvatar({ userID: props.userID }).then(() => fetchUser()).finally(() => { avatar.processing = false })
}

function dispatchCortezaSystemUserEvent($event, { user }) {}
</script>
