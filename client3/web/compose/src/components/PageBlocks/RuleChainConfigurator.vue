<template>
  <div>
    <div class="mb-3">
      <label class="form-label">Rule Chain ID</label>
      <select
        v-model="options.chainID"
        class="form-select"
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
        v-model="options.label"
        class="form-control"
        placeholder="Run Rule Chain"
      />
    </div>

    <div class="row mb-3">
      <div class="col-6">
        <label class="form-label">Variant</label>
        <select
          v-model="options.variant"
          class="form-select"
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
          v-model="options.size"
          class="form-select"
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
        v-model="options.icon"
        class="form-control"
        placeholder="play"
      />
    </div>

    <div class="form-check form-switch mb-3">
      <input
        id="rulechain-reload-on-success"
        v-model="options.reloadOnSuccess"
        class="form-check-input"
        type="checkbox"
        role="switch"
      />
      <label class="form-check-label" for="rulechain-reload-on-success">{{ reloadOnSuccessLabel }}</label>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, onMounted, inject } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  block: { type: Object, required: true },
})

const { t, locale } = useI18n({ useScope: 'global' })

const options = computed(() => {
  if (!props.block.options || typeof props.block.options !== 'object') {
    props.block.options = {}
  }
  return props.block.options
})

const reloadOnSuccessLabel = computed(() => {
  const v = t('ruleChain.reloadOnSuccess')
  if (v && !String(v).includes('reloadOnSuccess')) return v
  const loc = String(locale.value || '').split('-')[0]
  return loc === 'en' ? 'Reload page on successful completion' : 'Обновить страницу при успешном завершении'
})

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
