<template>
  <div v-if="queue" class="container pt-2 pb-3">
    <c-content-header :title="title" class="mb-2">
      <button v-if="queueID && canCreate" class="btn btn-primary" @click="$router.push({ name: 'system.queue.new' })">{{ $t('system.queues.editor.new') }}</button>
    </c-content-header>
    <c-queue-editor-info :queue="queue" :processing="info.processing" :success="info.success" :can-create="canCreate" :consumers="consumers" @delete="onDelete" @submit="onSubmit" />
  </div>
</template>
<script setup>
import { ref, computed, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { isEqual, cloneDeep } from 'lodash'
import CQueueEditorInfo from '../../../components/Queues/CQueueEditorInfo.vue'
const props = defineProps({ queueID: { type: String, required: false, default: undefined } })
const router = useRouter()
const { t } = useI18n()
const queue = ref(undefined)
const initialQueueState = ref(undefined)
const consumers = ref([])
const info = reactive({ processing: false, success: false })
const canCreate = computed(() => can('system/', 'queue.create'))
const title = computed(() => queue.value?.queueID ? t('system.queues.editor.title.edit') : t('system.queues.editor.title.new'))
function can(resource, operation) { return true }
function incLoader() {}
function decLoader() {}
watch(() => props.queueID, () => {
  fetchQueueConsumers()
  if (props.queueID) { fetchQueue() } else { queue.value = { consumer: 'corteza', meta: { poll_delay: '', dispatch_events: false }, queue: '' }; initialQueueState.value = cloneDeep(queue.value) }
}, { immediate: true })
function fetchQueue() { incLoader(); window.__systemAPI.queuesRead({ queueID: props.queueID }).then(q => { queue.value = q; initialQueueState.value = cloneDeep(q) }).finally(() => decLoader()) }
function fetchQueueConsumers() { consumers.value = [{ value: 'store', text: 'Store' }, { value: 'eventbus', text: 'Eventbus' }, { value: 'corteza', text: 'Lowcooode' }, { value: 'redis', text: 'Redis' }] }
function onSubmit(q) {
  incLoader()
  if (props.queueID) { window.__systemAPI.queuesUpdate(q).then(q => { queue.value = q; initialQueueState.value = cloneDeep(q); info.success = true; setTimeout(() => { info.success = false }, 2000) }).finally(() => decLoader()) }
  else { window.__systemAPI.queuesCreate(q).then(({ queueID }) => { info.success = true; setTimeout(() => { info.success = false }, 2000); router.push({ name: 'system.queue.edit', params: { queueID } }) }).finally(() => decLoader()) }
}
function onDelete() {
  incLoader(); const { deletedAt = '' } = queue.value; const method = deletedAt ? 'queuesUndelete' : 'queuesDelete'
  window.__systemAPI[method]({ queueID: props.queueID }).then(() => { fetchQueue(); if (!deletedAt) { queue.value.deletedAt = new Date(); router.push({ name: 'system.queue' }) } }).finally(() => decLoader())
}
</script>
