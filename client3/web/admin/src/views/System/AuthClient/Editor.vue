<template>
  <div v-if="authclient" class="container pt-2 pb-3">
    <c-content-header :title="title">
      <button v-if="authClientID && canCreate" class="btn btn-primary" @click="$router.push({ name: 'system.authClient.new' })">{{ $t('new') }}</button>
      <c-permissions-button v-if="authClientID && canGrant" :title="authclient.meta.name || authclient.handle || authClientID" :target="authclient.meta.name || authclient.handle || authClientID" :resource="`corteza::system:auth-client/${authClientID}`"><font-awesome-icon :icon="['fas', 'lock']" /> {{ $t('permissions') }}</c-permissions-button>
    </c-content-header>
    <c-authclient-editor-info :key="authClientID" :resource="authclient" :processing="info.processing" :success="info.success" :can-delete="authclient && authclient.authClientID && !authclient.isDefault && authclient.canDeleteAuthClient" :can-create="canCreate" @submit="onSubmit($event)" @delete="onDelete($event)" @undelete="onUndelete($event)" />
  </div>
</template>
<script setup>
import { ref, computed, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { isEqual } from 'lodash'
import { system } from 'corteza-lib/js/dist'
import CAuthclientEditorInfo from '../../../components/Authclient/CAuthclientEditorInfo.vue'
const props = defineProps({ authClientID: { type: String, required: false, default: undefined } })
const router = useRouter()
const { t } = useI18n()
const authclient = ref(undefined)
const initialAuthclientState = ref(undefined)
const info = reactive({ processing: false, success: false })
const canCreate = computed(() => can('system/', 'auth-client.create'))
const canGrant = computed(() => can('system/', 'grant'))
const title = computed(() => props.authClientID ? t('title.edit') : t('title.create'))
function can(resource, operation) { return true }
function incLoader() {}
function decLoader() {}
watch(() => props.authClientID, () => {
  if (props.authClientID) { fetchAuthclient() } else { authclient.value = new system.AuthClient(); initialAuthclientState.value = authclient.value.clone() }
}, { immediate: true })
function fetchAuthclient() {
  incLoader()
  window.__systemAPI.authClientRead({ clientID: props.authClientID }).then(ac => { authclient.value = new system.AuthClient(ac); initialAuthclientState.value = authclient.value.clone() }).finally(() => decLoader())
}
function onSubmit(ac) {
  info.processing = true
  if (props.authClientID) {
    window.__systemAPI.authClientUpdate({ clientID: props.authClientID, ...ac }).then(a => { authclient.value = new system.AuthClient(a); initialAuthclientState.value = authclient.value.clone(); info.success = true; setTimeout(() => { info.success = false }, 2000) }).finally(() => { info.processing = false })
  } else {
    window.__systemAPI.authClientCreate({ ...ac }).then((a) => { authclient.value = new system.AuthClient(a); initialAuthclientState.value = authclient.value.clone(); const { authClientID } = a; info.success = true; setTimeout(() => { info.success = false }, 2000); router.push({ name: 'system.authClient.edit', params: { authClientID } }) }).finally(() => { info.processing = false })
  }
}
function onDelete() { incLoader(); window.__systemAPI.authClientDelete({ clientID: props.authClientID }).then(() => { fetchAuthclient(); authclient.value.deletedAt = new Date(); router.push({ name: 'system.authClient' }) }).finally(() => decLoader()) }
function onUndelete() { incLoader(); window.__systemAPI.authClientUndelete({ clientID: props.authClientID }).then(() => fetchAuthclient()).finally(() => decLoader()) }
</script>
