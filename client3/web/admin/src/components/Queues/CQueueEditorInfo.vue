<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div class="card-body">
  <form @submit.prevent="$emit('submit', queue)">
        <div class="row g-3 p-3">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('name') }}</label>
              <input
                v-model="queue.queue"
                type="text"
                class="form-control"
                :class="{ 'is-invalid': handleState === false }"
                data-test-id="input-name"
              >
              <div v-if="handleState === false" class="invalid-feedback" data-test-id="feedback-invalid-name">
                {{ $t('system.queues.editor.info.invalid-handle-characters') }}
              </div>
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('consumer') }}</label>
              <select
                v-model="queue.consumer"
                class="form-select"
                data-test-id="input-consumer"
              >
                <option v-for="opt in consumers" :key="opt" :value="opt">{{ opt }}</option>
              </select>
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.queues.editor.info.poll_delay') }}</label>
              <small class="form-text text-muted">{{ metaPollDelayDescription }}</small>
              <input
                v-model="(queue.meta || {}).poll_delay"
                type="text"
                class="form-control"
                :class="{ 'is-invalid': durationState === false }"
                data-test-id="input-polling"
              >
            </div>
          </div>
  
          <div v-if="isMetaDispatchEvents" class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.queues.editor.info.dispatch_events') }}</label>
              <small class="form-text text-muted">{{ $t('system.queues.editor.info.dispatch_events_desc') }}</small>
              <div class="form-check">
                <input
                  id="checkbox-dispatch-events"
                  v-model="queue.meta.dispatch_events"
                  type="checkbox"
                  class="form-check-input-v3"
                >
                <label class="form-check-label" for="checkbox-dispatch-events">{{ $t("dispatch_events") }}</label>
              </div>
            </div>
          </div>
        </div>
  
        <c-system-fields :resource="queue" />
      </form>
  </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-input-confirm
        v-if="queue && queue.queueID && queue.canDeleteQueue"
        :data-test-id="deleteButtonStatusCypressId"
        :text="getDeleteStatus"
        variant="danger"
        size="md"
        @confirmed="$emit('delete')"
      />

      <c-button-submit
        :disabled="saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit', queue)"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'system.queues', keyPrefix: 'editor.info' } })
import { computed } from 'vue'

import { NoID } from 'corteza-lib/js/dist'
import { handle, useNsI18n } from 'corteza-lib/vue/dist'

const t = useNsI18n()
const props = defineProps({
  consumers: { type: Array, required: true },
  queue: { type: Object, required: true },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  canCreate: { type: Boolean, required: true },
})

defineEmits(['submit', 'delete'])

const fresh = computed(() => !props.queue.queueID || props.queue.queueID === NoID)
const editable = computed(() => fresh.value ? props.canCreate : true)

const saveDisabled = computed(() => !editable.value || [durationState.value, handleState.value].includes(false))

const durationState = computed(() => {
  const pd = (props.queue.meta || {}).poll_delay || ''
  const m = pd.match(/^((\d+h)?(\d+m)?(\d+s)?)|(\s)$/g)
  if (m && m.length && m[0] === pd) return null
  return false
})

const handleState = computed(() => {
  const { queue = '' } = props.queue
  return queue ? handle.handleState(queue) : false
})

const isMetaPollDelay = computed(() => {
  if (props.queue.queueID) return ((props.queue.meta || {}).poll_delay || '') === ''
  return true
})

const isMetaDispatchEvents = computed(() => {
  return ((props.queue || {}).meta || {}).dispatch_events === null
})

const getDeleteStatus = computed(() => props.queue.deletedAt ? t('undelete') : t('delete'))
const deleteButtonStatusCypressId = computed(() => `button-${getDeleteStatus.value.toLowerCase()}`)

const metaPollDelayDescription = computed(() => {
  return ((props.queue || {}).meta || {}).poll_delay
    ? t('poll_delay_set')
    : t('poll_delay_empty')
})
</script>
