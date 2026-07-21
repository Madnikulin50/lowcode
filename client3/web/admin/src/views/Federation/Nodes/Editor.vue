<template>
  <div class="container pt-2 pb-3">
    <c-content-header :title="title">
      <button v-if="nodeID" class="btn btn-link" @click="generate.modal = true">{{ $t('generateUri') }}</button>
    </c-content-header>
    <c-federation-editor-info :node="node" :processing="info.processing" :success="info.success" :can-create="canCreate" @submit="onInfoSubmit" @delete="onDelete" />
    <div v-if="generate.modal" class="modal fade show d-block" tabindex="-1" style="background:rgba(0,0,0,0.5)">
      <div class="modal-dialog modal-lg modal-dialog-centered">
        <div class="modal-content px-5">
          <div class="modal-body text-center px-5">
            <font-awesome-icon size="7x" :icon="['fas', 'share-alt']" class="text-light mb-2" />
            <h2>{{ $t('generate.description') }}</h2>
            <div class="input-group input-group-lg mt-5">
              <input v-model="generate.email" type="email" class="form-control" placeholder="email@example.com" />
              <button class="btn btn-primary" :disabled="!generate.url || !generate.email" @click="sendEmail()">{{ generate.processing ? $t('loading') : generate.success ? '✓' : $t('generate.sendEmail') }}</button>
            </div>
            <div class="mt-3">
              <p>{{ $t('generate.subject') }} <strong>{{ $t('generate.invitation') }}</strong></p>
              <p class="mt-4">{{ $t('generate.hello') }}</p>
              <p>{{ $t('generate.body', { userLabel }) }}</p>
              <p class="text-center text-break"><i>{{ generate.url || $t('generate.notGenerated') }}</i></p>
              <p>{{ $t('generate.kindRegards') }}</p>
            </div>
            <hr class="my-3" />
            <span class="text-break">{{ generate.url || $t('generate.notGenerated') }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, computed, reactive, watch, inject } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { cloneDeep, isEqual } from 'lodash'
import CFederationEditorInfo from '../../../components/Federation/CFederationEditorInfo.vue'
const props = defineProps({ nodeID: { type: String, required: false, default: undefined } })
const router = useRouter()
const { t } = useI18n()
const $auth = inject('auth', {})
const node = ref({})
const initialNodeState = ref({})
const info = reactive({ processing: false, success: false })
const generate = reactive({ modal: false, processing: false, success: false, email: '', url: '' })
const canCreate = computed(() => can('federation/', 'node.create'))
function can(resource, operation) { return true }
const userLabel = computed(() => $auth.user?.name || $auth.user?.email)
const title = computed(() => props.nodeID ? t('title.edit') : t('title.create'))
function incLoader() {} function decLoader() {}
watch(() => props.nodeID, () => { if (props.nodeID) { fetchNode(); fetchGeneratedUrl() } else { node.value = { name: '', baseURL: '', contact: '' }; initialNodeState.value = { name: '', baseURL: '', contact: '' } } }, { immediate: true })
function fetchNode() { incLoader(); window.__FederationAPI.nodeRead({ nodeID: props.nodeID }).then(n => { node.value = n; initialNodeState.value = cloneDeep(n) }).finally(() => decLoader()) }
function fetchGeneratedUrl() { incLoader(); window.__FederationAPI.nodeGenerateUri({ nodeID: props.nodeID }).then(url => { generate.url = url }).finally(() => decLoader()) }
function onInfoSubmit(n) { info.processing = true; const payload = { ...n }; if (payload.nodeID) { window.__FederationAPI.nodeUpdate(payload).then(n => { node.value = n; initialNodeState.value = cloneDeep(n); info.success = true; setTimeout(() => { info.success = false }, 2000) }).finally(() => { info.processing = false }) } else { window.__FederationAPI.nodeCreate(payload).then(({ nodeID }) => { info.success = true; setTimeout(() => { info.success = false }, 2000); router.push({ name: 'federation.nodes.edit', params: { nodeID } }) }).finally(() => { info.processing = false }) } }
function onDelete() { incLoader(); if (node.value.deletedAt) { window.__FederationAPI.nodeUndelete({ nodeID: props.nodeID }).then(() => fetchNode()).finally(() => decLoader()) } else { window.__FederationAPI.nodeDelete({ nodeID: props.nodeID }).then(() => { fetchNode(); node.value.deletedAt = new Date(); router.push({ name: 'federation.nodes' }) }).finally(() => decLoader()) } }
async function sendEmail() { generate.processing = true; const html = `<p class="mt-4">Hello,</p><p>${userLabel.value} is sending you an invitation for Corteza Federated Network.</p><p>To start sharing data between organizations, go to the admin panel of your Corteza application, click on &ldquo;Federation&rdquo; and select &ldquo;Pair Federation Network&rdquo; on top right corner.<br />Copy the link below and await confirmation from another administrator.</p><blockquote><p class="text-center text-break"><em>${generate.url}</em></p></blockquote><p>Kind regards, Corteza team.</p>`; await window.__ComposeAPI.notificationEmailSend({ to: [generate.email], subject: t('generate.invitation'), content: { html } }).then(() => { generate.email = ''; generate.success = true; setTimeout(() => { generate.success = false }, 2000) }).finally(() => { generate.processing = false }) }
</script>
