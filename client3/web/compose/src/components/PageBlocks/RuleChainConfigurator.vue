<template>
  <div>
    <div class="mb-3">
      <label class="form-label">Rule Chain ID</label>
      <select
        :value="block.options.chainID"
        class="form-select"
        @change="$emit('update:block', { ...block, options: { ...block.options, chainID: $event.target.value } })"
      >
        <option value="">Select a chain</option>
        <option v-for="chain in availableChains" :key="chain.id" :value="chain.id">
          {{ chain.name }} ({{ chain.nodeCount }} nodes)
        </option>
      </select>
      <small class="form-text text-muted">Select which rule chain to run when the button is clicked</small>
    </div>

    <div class="mb-3">
      <label class="form-label">Button Label</label>
      <input
        :value="block.options.label"
        class="form-control"
        placeholder="Run Rule Chain"
        @input="$emit('update:block', { ...block, options: { ...block.options, label: $event.target.value } })"
      />
    </div>

    <div class="row mb-3">
      <div class="col-6">
        <label class="form-label">Variant</label>
        <select
          :value="block.options.variant || 'primary'"
          class="form-select"
          @change="$emit('update:block', { ...block, options: { ...block.options, variant: $event.target.value } })"
        >
          <option value="primary">Primary</option>
          <option value="secondary">Secondary</option>
          <option value="success">Success</option>
          <option value="danger">Danger</option>
          <option value="warning">Warning</option>
          <option value="info">Info</option>
          <option value="outline-primary">Outline Primary</option>
        </select>
      </div>
      <div class="col-6">
        <label class="form-label">Size</label>
        <select
          :value="block.options.size || ''"
          class="form-select"
          @change="$emit('update:block', { ...block, options: { ...block.options, size: $event.target.value } })"
        >
          <option value="">Default</option>
          <option value="sm">Small</option>
          <option value="lg">Large</option>
        </select>
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label">Icon (FontAwesome)</label>
      <input
        :value="block.options.icon || 'play'"
        class="form-control"
        placeholder="play"
        @input="$emit('update:block', { ...block, options: { ...block.options, icon: $event.target.value } })"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, inject } from 'vue'

const props = defineProps({
  block: { type: Object, required: true },
})

defineEmits(['update:block'])

const $ComposeAPI = inject('$ComposeAPI')
const availableChains = ref([])

onMounted(async () => {
  try {
    const { data } = await $ComposeAPI.api().request({
      method: 'get',
      url: $ComposeAPI.baseURL + '/rulechain/',
    })
    availableChains.value = data?.response?.chains || data?.chains || []
  } catch {
    availableChains.value = []
  }
})
</script>
