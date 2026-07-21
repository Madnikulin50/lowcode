<template>
  <div v-if="workflow" class="container pt-2 pb-3">
    <c-content-header :title="title">
      <button v-if="workflowID && canCreate" class="btn btn-primary" @click="$router.push({ name: 'automation.workflow.new' })">{{ $t('new') }}</button>
      <c-permissions-button v-if="workflowID && canGrant" :title="workflow.meta.name || workflow.handle || workflowID" :target="workflow.meta.name || workflow.handle || workflowID" :resource="`corteza::automation:workflow/${workflowID}`"><font-awesome-icon :icon="['fas', 'lock']" /> {{ $t('permissions') }}</c-permissions-button>
    </c-content-header>
    <c-workflow-editor-info :workflow="workflow" :processing="info.processing" :success="info.success" :can-create="canCreate" @submit="onInfoSubmit" @delete="onDelete" />
    <c-workflow-editor-triggers v-if="workflowID" :triggers="triggers" :processing="info.processing" :success="info.success" />
  </div>
</template>
<script setup>
import { ref, computed, reactive, watch, inject } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { isEqual, cloneDeep } from 'lodash'
import CWorkflowEditorInfo from '../../../components/Workflow/CWorkflowEditorInfo.vue'
import CWorkflowEditorTriggers from '../../../components/Workflow/CWorkflowEditorTriggers.vue'
const props = defineProps({ workflowID: { type: String, required: false, default: undefined } })
const router = useRouter()
const { t } = useI18n()
const $auth = inject('auth', {})
const workflow = ref(undefined)
const initialWorkflowState = ref(undefined)
const triggers = ref([])
const info = reactive({ processing: false, success: false })
const canCreate = computed(() => can('automation/', 'workflow.create'))
const canGrant = computed(() => can('automation/', 'grant'))
function can(resource, operation) { return true }
const userID = computed(() => $auth.user?.userID)
const title = computed(() => props.workflowID ? t('title.edit') : t('title.create'))
function incLoader() {} function decLoader() {}
watch(() => props.workflowID, () => { if (props.workflowID) { fetchWorkflow(); fetchTriggers() } else { workflow.value = { ownedBy: userID.value, runAs: userID.value, enabled: true, meta: { name: '' } }; initialWorkflowState.value = cloneDeep(workflow.value) } }, { immediate: true })
function fetchWorkflow() { incLoader(); window.__AutomationAPI.workflowRead({ workflowID: props.workflowID }).then(prepare).finally(() => decLoader()) }
function fetchTriggers() { incLoader(); window.__AutomationAPI.triggerList({ workflowID: props.workflowID, disabled: 1 }).then(({ set = [] }) => { triggers.value = set }).finally(() => decLoader()) }
function onInfoSubmit(w) { info.processing = true; if (props.workflowID) { window.__AutomationAPI.workflowUpdate(w).then(() => { fetchWorkflow(); info.success = true; setTimeout(() => { info.success = false }, 2000) }).finally(() => { info.processing = false }) } else { window.__AutomationAPI.workflowCreate(w).then(({ workflowID }) => { info.success = true; setTimeout(() => { info.success = false }, 2000); router.push({ name: 'automation.workflow.edit', params: { workflowID } }) }).finally(() => { info.processing = false }) } }
function onDelete() { incLoader(); if (workflow.value.deletedAt) { window.__AutomationAPI.workflowUndelete({ workflowID: props.workflowID }).then(() => fetchWorkflow()).finally(() => decLoader()) } else { window.__AutomationAPI.workflowDelete({ workflowID: props.workflowID }).then(() => { fetchWorkflow(); workflow.value.deletedAt = new Date(); router.push({ name: 'automation.workflow' }) }).finally(() => decLoader()) } }
function prepare(w = {}) { workflow.value = w; initialWorkflowState.value = cloneDeep(w) }
</script>
