<template>
  <div class="container-fluid d-flex flex-column h-100 pt-2 pb-3">
    <c-content-header :title="$t('ui.title.federation')" />
    <c-permission-list :roles="sortedRoles" :permissions="permissions" :role-permissions="rolePermissions" :can-grant="canGrant" :loaded="isLoaded" :processing="permission.processing" :success="permission.success" component="federation" class="flex-fill" @submit="onSubmit" @add="addRole" @hide="hideRole" />
  </div>
</template>
<script setup>
defineOptions({ i18nOptions: { namespaces: 'permissions' } })
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { usePermissionHelpers } from '../../../mixins/permissionHelpers'
import CPermissionList from '../../../components/Permissions/CPermissionList.vue'
const { t } = useI18n()
const ph = usePermissionHelpers(window.__FederationAPI)
const canGrant = computed(() => can('federation/', 'grant'))
function can(resource, operation) { return true }
const sortedRoles = computed(() => ph.roles.sort((a, b) => a.mode?.localeCompare(b.mode)))
const { isLoaded, permissions, rolePermissions, permission } = ph
onMounted(() => { if (window.__FederationAPI) ph.fetchPermissions(window.__FederationAPI) })
function onSubmit(roleRules) { ph.onSubmit(roleRules, window.__FederationAPI) }
function addRole(role) { ph.addRole(role) }
function hideRole(role) { ph.hideRole(role) }
</script>
