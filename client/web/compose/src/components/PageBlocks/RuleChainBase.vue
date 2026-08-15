<template>
  <wrap v-bind="$props" v-on="$listeners">
    <div class="p-3">
      <div v-if="result" class="mb-3 p-2 border rounded" :class="result.success ? 'border-success bg-success-subtle' : 'border-danger bg-danger-subtle'">
        <div v-if="result.success" class="text-success">
          <font-awesome-icon :icon="['fas', 'check-circle']" class="me-1" />
          {{ $t('ruleChain.success') }}
        </div>
        <div v-else class="text-danger">
          <font-awesome-icon :icon="['fas', 'exclamation-circle']" class="me-1" />
          {{ result.error || $t('ruleChain.error') }}
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
  </wrap>
</template>

<script>
import { NoID } from 'corteza-lib/js/dist'
import base from './base'

export default {
  extends: base,

  data () {
    return {
      running: false,
      result: null,
    }
  },

  computed: {
    chainID () {
      return this.options.chainID || ''
    },

    label () {
      return this.options.label || this.$t('ruleChain.run')
    },

    icon () {
      const iconOpt = this.options.icon || 'play'
      return ['fas', iconOpt]
    },

    btnClass () {
      const variant = this.options.variant || 'primary'
      const size = this.options.size || ''
      return `btn-${variant} ${size ? 'btn-' + size : ''}`
    },

    contextData () {
      return {
        pageID: this.page.pageID || NoID,
        moduleID: this.module?.moduleID || NoID,
        namespaceID: this.namespace?.namespaceID || NoID,
        recordID: this.record?.recordID || NoID,
        record: this.record,
        userID: this.$auth?.user?.userID || NoID,
        ...(this.options.context || {}),
      }
    },
  },

  methods: {
    async runChain () {
      if (!this.chainID) {
        return
      }

      this.running = true
      this.result = null

      try {
        const response = await this.$ComposeAPI.api().request({
          method: 'post',
          url: this.$ComposeAPI.baseURL + '/pageblock/trigger',
          data: {
            chainID: this.chainID,
            blockID: this.block.title || '',
            pageID: this.page.pageID,
            moduleID: this.module?.moduleID,
            namespaceID: this.namespace?.namespaceID,
            recordID: this.record?.recordID,
            record: this.record,
            context: this.options.context || {},
          },
        })

        this.result = response?.data?.response || response?.data || { success: true }
      } catch (err) {
        this.result = { success: false, error: err.message || 'Request failed' }
      } finally {
        this.running = false
      }
    },
  },
}
</script>
