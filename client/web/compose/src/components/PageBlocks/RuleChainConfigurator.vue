<template>
  <div>
    <div class="mb-3">
      <label class="form-label">{{ $t('ruleChain.config.chainID') }}</label>
      <select
        :value="options.chainID"
        class="form-select"
        @change="$emit('update:options', { ...options, chainID: $event.target.value })"
      >
        <option value="">{{ $t('ruleChain.config.selectChain') }}</option>
        <option
          v-for="chain in availableChains"
          :key="chain.id"
          :value="chain.id"
        >
          {{ chain.name }} ({{ chain.nodeCount }} nodes)
        </option>
      </select>
      <small class="form-text text-muted">{{ $t('ruleChain.config.chainIDHelp') }}</small>
    </div>

    <div class="mb-3">
      <label class="form-label">{{ $t('ruleChain.config.label') }}</label>
      <input
        :value="options.label"
        class="form-control"
        :placeholder="$t('ruleChain.config.labelPlaceholder')"
        @input="$emit('update:options', { ...options, label: $event.target.value })"
      />
    </div>

    <div class="row mb-3">
      <div class="col-6">
        <label class="form-label">{{ $t('ruleChain.config.variant') }}</label>
        <select
          :value="options.variant || 'primary'"
          class="form-select"
          @change="$emit('update:options', { ...options, variant: $event.target.value })"
        >
          <option value="primary">Primary</option>
          <option value="secondary">Secondary</option>
          <option value="success">Success</option>
          <option value="danger">Danger</option>
          <option value="warning">Warning</option>
          <option value="info">Info</option>
          <option value="outline-primary">Outline Primary</option>
          <option value="outline-secondary">Outline Secondary</option>
        </select>
      </div>
      <div class="col-6">
        <label class="form-label">{{ $t('ruleChain.config.size') }}</label>
        <select
          :value="options.size || ''"
          class="form-select"
          @change="$emit('update:options', { ...options, size: $event.target.value })"
        >
          <option value="">{{ $t('ruleChain.config.sizeDefault') }}</option>
          <option value="sm">Small</option>
          <option value="lg">Large</option>
        </select>
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label">{{ $t('ruleChain.config.icon') }}</label>
      <input
        :value="options.icon || 'play'"
        class="form-control"
        placeholder="play"
        @input="$emit('update:options', { ...options, icon: $event.target.value })"
      />
      <small class="form-text text-muted">{{ $t('ruleChain.config.iconHelp') }}</small>
    </div>

    <div class="mb-3">
      <label class="form-label">{{ $t('ruleChain.config.context') }}</label>
      <textarea
        :value="contextString"
        class="form-control"
        rows="3"
        :placeholder="contextPlaceholder"
        @input="onContextInput"
      />
      <small class="form-text text-muted">
        {{ $t('ruleChain.config.contextHelp') }}
      </small>
    </div>
  </div>
</template>

<script>
import base from './base'

export default {
  extends: base,

  data () {
    return {
      availableChains: [],
      chainsLoaded: false,
    }
  },

  computed: {
    contextPlaceholder () {
      return '{\n  "extraField": "value"\n}'
    },

    contextString () {
      try {
        return JSON.stringify(this.options.context || {}, null, 2)
      } catch {
        return '{}'
      }
    },
  },

  mounted () {
    this.fetchChains()
  },

  methods: {
    async fetchChains () {
      try {
        const response = await this.$SystemAPI.api().request({
          method: 'get',
          url: this.$ComposeAPI.baseURL + '/rulechain/',
        })
        this.availableChains = response?.data?.response?.chains || response?.data?.chains || []
      } catch {
        this.availableChains = []
      }
      this.chainsLoaded = true
    },

    onContextInput (e) {
      try {
        const ctx = JSON.parse(e.target.value)
        this.$emit('update:options', { ...this.options, context: ctx })
      } catch {
        // ignore invalid JSON while typing
      }
    },
  },
}
</script>
