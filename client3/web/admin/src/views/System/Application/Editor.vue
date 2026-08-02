<template>
  <div v-if="application" class="container pt-2 pb-3">
    <c-content-header :title="title">
      <button v-if="applicationID && canCreate" class="btn btn-primary" @click="$router.push({ name: 'system.application.new' })">{{ $t('system.applications.editor.new') }}</button>
      <c-permissions-button v-if="applicationID && canGrant" :title="application.name || applicationID" :target="application.name || applicationID" :resource="`corteza::system:application/${applicationID}`"><font-awesome-icon :icon="['fas', 'lock']" /> {{ $t('system.applications.editor.permissions') }}</c-permissions-button>
    </c-content-header>
    <c-application-editor-info :application="application" :processing="info.processing" :success="info.success" :can-create="canCreate" @submit="onInfoSubmit" @delete="onDelete" />
    <c-application-editor-unify v-if="applicationID && application.unify && application.applicationID" class="mt-3" :unify="application.unify" :application="application" :can-pin="canPin" :processing="unify.processing" :success="unify.success" @change-detected="unifyAssetStateChange = true" @submit="onUnifySubmit" />
  </div>
</template>
<script setup>
import { ref, computed, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { isEqual } from 'lodash'
import { system } from 'corteza-lib/js/dist'
import CApplicationEditorInfo from '../../../components/Application/CApplicationEditorInfo.vue'
import CApplicationEditorUnify from '../../../components/Application/CApplicationEditorUnify.vue'
const props = defineProps({ applicationID: { type: String, required: false, default: undefined } })
const router = useRouter()
const { t } = useI18n()
const application = ref(undefined)
const initialApplicationState = ref(undefined)
const info = reactive({ processing: false, success: false })
const unify = reactive({ processing: false, success: false })
const unifyAssetStateChange = ref(false)
const canCreate = computed(() => can('system/', 'application.create'))
const canGrant = computed(() => can('system/', 'grant'))
const canPin = computed(() => can('system/', 'pin'))
const title = computed(() => props.applicationID ? t('system.applications.editor.title.edit') : t('system.applications.editor.title.create'))
function can(resource, operation) { return true }
function incLoader() {}
function decLoader() {}
watch(() => props.applicationID, () => {
  if (props.applicationID) { fetchApplication() } else { application.value = new system.Application(); initialApplicationState.value = application.value.clone() }
}, { immediate: true })
function fetchApplication() {
  incLoader()
  window.__systemAPI.applicationRead({ applicationID: props.applicationID, incFlags: 1 }).then((a = {}) => {
    if (!a.unify) a.unify = { listed: true, pinned: false, name: application.value.name, config: '', icon: '', logo: '', url: '' }
    a.unify.pinned = (a.flags || []).includes('pinned')
    a.unify.name = a.unify.name ? a.unify.name : a.name
    application.value = new system.Application(a)
    initialApplicationState.value = application.value.clone()
  }).finally(() => decLoader())
}
function onInfoSubmit(a) {
  info.processing = true
  if (props.applicationID) {
    a = { ...a, unify: initialApplicationState.value.unify }
    window.__systemAPI.applicationUpdate(a).then(a => {
      initialApplicationState.value = new system.Application({ ...a, unify: initialApplicationState.value.unify })
      application.value = new system.Application({ ...a, unify: application.value.unify })
      info.success = true; setTimeout(() => { info.success = false }, 2000)
    }).finally(() => { info.processing = false })
  } else {
    window.__systemAPI.applicationCreate(a).then(({ applicationID }) => { info.success = true; setTimeout(() => { info.success = false }, 2000); router.push({ name: 'system.application.edit', params: { applicationID } }) }).finally(() => { info.processing = false })
  }
}
async function onUnifySubmit({ unify: u, unifyAssets }) {
  unify.processing = true
  if (unifyAssets.logo || unifyAssets.icon) {
    try { const assets = await uploadAssets(unifyAssets); u = { ...u, ...assets } } catch (e) { unify.processing = false; return }
  }
  if (props.applicationID) {
    const flagPayload = { applicationID: props.applicationID, flag: 'pinned', ownedBy: '0' }
    if (u.pinned) { await window.__systemAPI.applicationFlagCreate(flagPayload).catch(() => {}) } else { await window.__systemAPI.applicationFlagDelete(flagPayload).catch(() => {}) }
    return window.__systemAPI.applicationUpdate({ ...initialApplicationState.value, unify: u }).then(() => {
      application.value = new system.Application({ ...application.value, unify: u })
      initialApplicationState.value = new system.Application({ ...initialApplicationState.value, unify: u })
      unifyAssetStateChange.value = false; unify.success = true; setTimeout(() => { unify.success = false }, 2000)
    }).finally(() => { unify.processing = false })
  }
}
async function uploadAssets(assets) {
  const rr = {}
  const rq = async (file) => { const formData = new FormData(); formData.append('upload', file); const rsp = await window.__systemAPI.api().request({ method: 'post', url: window.__systemAPI.applicationUploadEndpoint(), data: formData, headers: { 'Content-Type': 'multipart/form-data' } }); if (rsp.data.error) throw new Error(rsp.data.error.message); return rsp.data.response }
  const baseURL = window.__systemAPI.baseURL
  if (assets.logo) { const rsp = await rq(assets.logo); rr.logo = baseURL + rsp.url; rr.logoID = rsp.attachmentID; assets.logo = undefined }
  if (assets.icon) { const rsp = await rq(assets.icon); rr.icon = baseURL + rsp.url; rr.iconID = rsp.attachmentID; assets.icon = undefined }
  return rr
}
function onDelete() {
  incLoader()
  if (application.value.deletedAt) { window.__systemAPI.applicationUndelete({ applicationID: props.applicationID }).then(() => fetchApplication()).finally(() => decLoader()) }
  else { window.__systemAPI.applicationDelete({ applicationID: props.applicationID }).then(() => { fetchApplication(); application.value.deletedAt = new Date(); router.push({ name: 'system.application' }) }).finally(() => decLoader()) }
}
</script>
