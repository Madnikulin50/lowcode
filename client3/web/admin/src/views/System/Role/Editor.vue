<template>
  <div v-if="role" class="container pt-2 pb-3">
    <c-content-header :title="title">
      <button v-if="roleID && canCreate" class="btn btn-primary" @click="$router.push({ name: 'system.role.new' })">{{ $t('system.roles.editor.new') }}</button>
      <c-permissions-button v-if="roleID && canGrant" :title="role.name || role.handle || role.roleID" :target="role.name || role.handle || role.roleID" :resource="`corteza::system:role/${roleID}`"><font-awesome-icon :icon="['fas', 'lock']" /> {{ $t('system.roles.editor.permissions') }}</c-permissions-button>
      <c-permission-clone v-if="roleID && canGrant" :role-id="roleID" />
      <c-corredor-manual-buttons ui-page="role/editor" ui-slot="toolbar" resource-type="system:role" default-variant="link" class="me-1" @click="dispatchCortezaSystemRoleEvent($event, { role })" />
    </c-content-header>
    <c-role-editor-info :role="role" :processing="info.processing" :success="info.success" :is-context="isContext" :can-create="canCreate" @submit="onInfoSubmit" @delete="onDelete" @status="onStatusChange" @update:is-context="isContext = $event" />
    <c-role-editor-members v-if="!isContext && canManageMembers" v-model="members.active" :processing="members.processing" :success="members.success" class="mt-3" @submit="onMembersSubmit" />
  </div>
</template>
<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.roles', keyPrefix: 'editor' } })
import { ref, computed, reactive, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { system } from 'corteza-lib/js/dist'
import { isEqual } from 'lodash'
import CPermissionClone from '../../../components/Permissions/CPermissionClone.vue'
import CRoleEditorInfo from '../../../components/Role/CRoleEditorInfo.vue'
import CRoleEditorMembers from '../../../components/Role/CRoleEditorMembers.vue'
const props = defineProps({ roleID: { type: String, required: false, default: undefined } })
const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const role = ref(undefined)
const initialRoleState = ref(undefined)
const isContext = ref(false)
const info = reactive({ processing: false, success: false })
const members = reactive({ active: undefined, initial: undefined, processing: false, success: false })
const canManageMembers = computed(() => role.value && role.value.canManageMembersOnRole && role.value.roleID && members.active && !role.value.isClosed && !role.value.isContext)
const canCreate = computed(() => can('system/', 'role.create'))
const canGrant = computed(() => can('system/', 'grant'))
const title = computed(() => props.roleID ? t('system.roles.editor.title.edit') : t('system.roles.editor.title.create'))
function can(resource, operation) { return true }
function incLoader() {}
function decLoader() {}
watch(() => props.roleID, () => {
  if (props.roleID) { fetchRole() } else { role.value = new system.Role(); initialRoleState.value = role.value.clone(); isContext.value = false }
}, { immediate: true })
watch(() => role.value?.isContext, (v) => { if (v) isContext.value = true }, { immediate: true })
function fetchRole() {
  incLoader()
  if (props.roleID === '1') { router.push({ name: 'system.role.list' }) }
  window.__systemAPI.roleRead({ roleID: props.roleID }).then(r => {
    role.value = new system.Role(r); initialRoleState.value = role.value.clone(); isContext.value = !!role.value.isContext
    if (role.value.canManageMembersOnRole && !role.value.isContext && !role.value.isClosed) {
      return window.__systemAPI.roleMemberList(r).then((mm = []) => { members.active = [...mm]; members.initial = [...mm] })
    }
  }).finally(() => decLoader())
}
function onDelete() {
  incLoader()
  if (role.value.deletedAt) { window.__systemAPI.roleUndelete({ roleID: props.roleID }).then(() => fetchRole()).finally(() => decLoader()) }
  else { window.__systemAPI.roleDelete({ roleID: props.roleID }).then(() => { fetchRole(); role.value.deletedAt = new Date(); router.push({ name: 'system.role' }) }).finally(() => decLoader()) }
}
function onInfoSubmit(r) {
  info.processing = true
  if (props.roleID) { window.__systemAPI.roleUpdate(r).then(() => { fetchRole(); info.success = true; setTimeout(() => { info.success = false }, 2000) }).finally(() => { info.processing = false }) }
  else { window.__systemAPI.roleCreate(r).then(({ roleID }) => { info.success = true; setTimeout(() => { info.success = false }, 2000); router.push({ name: 'system.role.edit', params: { roleID } }) }).finally(() => { info.processing = false }) }
}
function onStatusChange() {
  incLoader()
  if (role.value.archivedAt) { window.__systemAPI.roleUnarchive({ roleID: props.roleID }).then(() => fetchRole()).finally(() => decLoader()) }
  else { window.__systemAPI.roleArchive({ roleID: props.roleID }).then(() => fetchRole()).finally(() => decLoader()) }
}
function onMembersSubmit() {
  members.processing = true
  const { roleID } = role.value
  if (roleID) {
    Promise.all([...members.active.filter(userID => !members.initial.includes(userID)).map(userID => window.__systemAPI.roleMemberAdd({ roleID, userID })), ...members.initial.filter(userID => !members.active.includes(userID)).map(userID => window.__systemAPI.roleMemberRemove({ roleID, userID }))])
      .then(() => { fetchRole(); members.success = true; setTimeout(() => { members.success = false }, 2000) }).finally(() => { members.processing = false })
  }
}
function dispatchCortezaSystemRoleEvent($event, { role }) {}
</script>
