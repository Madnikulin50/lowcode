<template>
  <div
    v-if="trace"
    class="border rounded bg-white p-2 mt-1 small"
    data-test="ai-trace"
  >
    <div class="d-flex flex-wrap align-items-center gap-1">
      <span class="fw-bold text-muted text-uppercase me-1">
        {{ $t('rulechain.trace.title', 'AI trace') }}
      </span>
      <span
        v-if="trace.agent"
        class="badge bg-light text-dark border"
        :title="$t('rulechain.trace.agent', 'Agent')"
      >{{ trace.agent }}</span>
      <span
        v-if="trace.model"
        class="badge bg-light text-dark border"
        :title="$t('rulechain.trace.model', 'Model')"
      >{{ trace.model }}</span>
      <span
        v-if="trace.durationMs !== undefined"
        class="badge bg-light text-dark border"
      >{{ formatMs(trace.durationMs) }}</span>
      <span
        v-if="tokens"
        class="badge bg-light text-dark border"
        :title="$t('rulechain.trace.tokensHint', 'Prompt / completion tokens')"
      >{{ trace.promptTokens || 0 }} / {{ trace.completionTokens || 0 }} {{ $t('rulechain.trace.tokens', 'tokens') }}</span>
      <span
        v-if="trace.attempts > 1"
        class="badge bg-warning text-dark"
      >{{ trace.attempts }} {{ $t('rulechain.trace.attempts', 'attempts') }}</span>
      <span
        v-if="trace.error"
        class="badge bg-danger"
      >{{ $t('rulechain.trace.failed', 'failed') }}</span>
    </div>

    <div
      v-if="trace.error"
      class="text-danger mt-1"
    >
      {{ trace.error }}
    </div>

    <div
      v-if="(trace.tools || []).length"
      class="mt-1 text-muted"
    >
      {{ $t('rulechain.trace.tools', 'Tools') }}: {{ trace.tools.join(', ') }}
    </div>

    <details
      v-if="(trace.rejected || []).length"
      class="mt-1"
    >
      <summary>{{ $t('rulechain.trace.rejected', 'Rejected answers') }} ({{ trace.rejected.length }})</summary>
      <ul class="mb-0 ps-3">
        <li
          v-for="(r, ix) in trace.rejected"
          :key="ix"
        >
          {{ r }}
        </li>
      </ul>
    </details>

    <details
      v-if="trace.prompt"
      class="mt-1"
    >
      <summary>{{ $t('rulechain.trace.prompt', 'Prompt') }}</summary>
      <pre class="bg-light border rounded p-2 mb-0 trace-pre">{{ trace.prompt }}</pre>
    </details>

    <details
      v-if="trace.response"
      class="mt-1"
      :open="!!trace.error"
    >
      <summary>{{ $t('rulechain.trace.response', 'Response') }}</summary>
      <pre class="bg-light border rounded p-2 mb-0 trace-pre">{{ trace.response }}</pre>
    </details>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  // CallTrace as returned by the server (server/pkg/aiagent/trace.go)
  trace: { type: Object, default: null },
})

const tokens = computed(() => !!(props.trace && (props.trace.promptTokens || props.trace.completionTokens)))

function formatMs (ms) {
  return ms >= 1000 ? `${(ms / 1000).toFixed(1)} s` : `${ms} ms`
}
</script>

<style scoped>
.trace-pre {
  max-height: 20vh;
  overflow: auto;
  white-space: pre-wrap;
}
</style>
