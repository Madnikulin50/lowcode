<template>
  <Wrap v-bind="$props">
    <div class="p-3">
      <div v-if="result" class="mb-3 p-2 border rounded" :class="result.success ? 'border-success bg-success-subtle' : 'border-danger bg-danger-subtle'">
        <div v-if="result.success" class="text-success">
          <font-awesome-icon :icon="['fas', 'check-circle']" class="me-1" />
          Chain executed successfully
        </div>
        <div v-else class="text-danger">
          <font-awesome-icon :icon="['fas', 'exclamation-circle']" class="me-1" />
          {{ result.error || 'Execution failed' }}
        </div>
        <pre v-if="result.output" class="mt-2 mb-0 small">{{ result.output }}</pre>
      </div>

      <button
        class="btn"
        :class="btnClass"
        :disabled="running"
        @click="runChain"
      >
        <span v-if="running" class="spinner-border spinner-border-sm me-1" role="status" />
        <font-awesome-icon :icon="icon" class="me-1" />
        {{ label }}
      </button>
    </div>
  </Wrap>
</template>

<script setup>
import { ref, computed, inject } from 'vue'
import { NoID } from 'corteza-lib/js/dist'

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, default: undefined },
  record: { type: Object, default: undefined },
})

const $ComposeAPI = inject('$ComposeAPI')

const running = ref(false)
const result = ref(null)

const chainID = computed(() => props.block.options?.chainID || '')
const label = computed(() => props.block.options?.label || 'Run Rule Chain')
const icon = computed(() => ['fas', props.block.options?.icon || 'play'])
const btnClass = computed(() => `btn-${props.block.options?.variant || 'primary'} ${(props.block.options?.size) ? 'btn-' + props.block.options.size : ''}`)

async function runChain() {
  if (!chainID.value) return
  running.value = true
  result.value = null
  try {
    const { data } = await $ComposeAPI.api().request({
      method: 'post',
      url: $ComposeAPI.baseURL + '/pageblock/trigger',
      data: {
        chainID: chainID.value,
        pageID: props.page?.pageID,
        moduleID: props.module?.moduleID,
        namespaceID: props.namespace?.namespaceID,
        recordID: props.record?.recordID,
        record: props.record,
        context: props.block.options?.context || {},
      },
    })
    result.value = data?.response || data || { success: true }
  } catch (err) {
    result.value = { success: false, error: err.message || 'Request failed' }
  } finally {
    running.value = false
  }
}
</script>
