<template>
  <div v-if="userGroup" class="container pt-2 pb-3">
    <c-content-header :title="title">
      <button v-if="userGroupID && canCreate" class="btn btn-primary" @click="$router.push({ name: 'system.userGroup.new' })">{{ $t('new') }}</button>
      <c-permissions-button v-if="userGroupID && canGrant" :title="userGroup.name || userGroup.handle || userGroup.userGroupID" :target="userGroup.name || userGroup.handle || userGroup.userGroupID" :resource="`corteza::system:user-group/${userGroupID}`"><font-awesome-icon :icon="['fas', 'lock']" /> {{ $t('permissions') }}</c-permissions-button>
    </c-content-header>
    <c-user-group-editor-info :user-group="userGroup" :processing="info.processing" :success="info.success" :can-create="canCreate" @submit="onInfoSubmit" @delete="onDelete" />
    <c-user-group-editor-members v-if="canManageMembers" v-model="groupMembers.active" :processing="groupMembers.processing" :success="groupMembers.success" class="mt-3" @submit="onMembersSubmit" />
    <c-user-group-editor-roles v-if="canManageRoles" v-model="groupRoles.active" :processing="groupRoles.processing" :success="groupRoles.success" class="mt-3" @submit="onUserGroupRolesSubmit" />
  </div>
</template>
<script setup>
import { ref, computed, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { system } from 'corteza-lib/js/dist'
import { isEqual } from 'lodash'
import CUserGroupEditorInfo from '../../../components/UserGroup/CUserGroupEditorInfo.vue'
import CUserGroupEditorMembers from '../../../components/UserGroup/CUserGroupEditorMembers.vue'
import CUserGroupEditorRoles from '../../../components/UserGroup/CUserGroupEditorRoles.vue'
const props = defineProps({ userGroupID: { type: String, required: false, default: undefined } })
const router = useRouter()
const { t } = useI18n()
const userGroup = ref(undefined)
const initialUserGroupState = ref(undefined)
const info = reactive({ processing: false, success: false })
const groupMembers = reactive({ active: undefined, initial: undefined, processing: false, success: false })
const groupRoles = reactive({ active: undefined, initial: undefined, processing: false, success: false })
const canManageMembers = computed(() => userGroup.value && userGroup.value.canManageMembersOnUserGroup && userGroup.value.userGroupID && groupMembers.active)
const canManageRoles = computed(() => userGroup.value && userGroup.value.userGroupID && groupRoles.active)
const canCreate = computed(() => can('system/', 'user-group.create'))
const canGrant = computed(() => can('system/', 'grant'))
const title = computed(() => props.userGroupID ? t('title.edit') : t('title.create'))
function can(resource, operation) { return true }
function incLoader() {}
function decLoader() {}
watch(() => props.userGroupID, async () => {
  if (props.userGroupID) { fetchUserGroup(); fetchUserGroupRoles(); fetchUserGroupMembers() }
  else { userGroup.value = new system.UserGroup(); initialUserGroupState.value = userGroup.value.clone() }
}, { immediate: true })
function fetchUserGroup() { incLoader(); window.__systemAPI.userGroupRead({ userGroupID: props.userGroupID }).then(r => { userGroup.value = new system.UserGroup(r); initialUserGroupState.value = userGroup.value.clone() }).finally(() => decLoader()) }
async function fetchUserGroupRoles() { incLoader(); return window.__systemAPI.roleList({ userGroupID: props.userGroupID }).then(({ set = [] }) => { groupRoles.active = [...set.map(({ roleID }) => roleID)]; groupRoles.initial = [...set.map(({ roleID }) => roleID)] }).finally(() => decLoader()) }
function fetchUserGroupMembers() { incLoader(); return window.__systemAPI.userGroupMemberList({ userGroupID: props.userGroupID }).then((set = []) => { groupMembers.active = [...set]; groupMembers.initial = [...set] }).finally(() => decLoader()) }
function onUserGroupRolesSubmit() {
  groupRoles.processing = true; const userGroupID = props.userGroupID
  Promise.all([...groupRoles.initial.filter(roleID => !groupRoles.active.includes(roleID)).map(roleID => window.__systemAPI.roleMemberRemoveGroup({ roleID, userGroupID })), ...groupRoles.active.filter(roleID => !groupRoles.initial.includes(roleID)).map(roleID => window.__systemAPI.roleMemberAddGroup({ roleID, userGroupID }))])
    .then(async () => { groupRoles.success = true; setTimeout(() => { groupRoles.success = false }, 2000); await fetchUserGroupRoles() }).finally(() => { groupRoles.processing = false })
}
function onDelete() {
  incLoader()
  if (userGroup.value.deletedAt) { window.__systemAPI.userGroupUndelete({ userGroupID: props.userGroupID }).then(() => fetchUserGroup()).finally(() => decLoader()) }
  else { window.__systemAPI.userGroupDelete({ userGroupID: props.userGroupID }).then(() => { fetchUserGroup(); userGroup.value.deletedAt = new Date(); router.push({ name: 'system.user-group' }) }).finally(() => decLoader()) }
}
function onInfoSubmit(ug) {
  info.processing = true
  if (props.userGroupID) { window.__systemAPI.userGroupUpdate(ug).then(() => { fetchUserGroup(); info.success = true; setTimeout(() => { info.success = false }, 2000) }).finally(() => { info.processing = false }) }
  else { window.__systemAPI.userGroupCreate(ug).then(({ userGroupID }) => { info.success = true; setTimeout(() => { info.success = false }, 2000); router.push({ name: 'system.userGroup.edit', params: { userGroupID } }) }).finally(() => { info.processing = false }) }
}
function onMembersSubmit() {
  groupMembers.processing = true; const userGroupID = props.userGroupID
  Promise.all([...groupMembers.active.filter(userID => !groupMembers.initial.includes(userID)).map(userID => window.__systemAPI.userGroupMemberAdd({ userGroupID, userID }))])
    .then(() => { groupMembers.success = true; setTimeout(() => { groupMembers.success = false }, 2000); fetchUserGroupMembers() }).finally(() => { groupMembers.processing = false })
}
</script>
