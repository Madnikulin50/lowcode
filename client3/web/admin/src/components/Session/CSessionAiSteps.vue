<template>
  <div
    v-if="steps.length"
    class="card shadow-sm mt-3"
  >
    <div class="card-header border-bottom">
      <h4 class="ae-section-title">
        {{ $t('automation.sessions.editor.aiSteps.title', 'AI steps') }}
      </h4>
    </div>

    <div class="card-body">
      <div
        v-for="(step, ix) in steps"
        :key="ix"
        class="border rounded p-2 mb-3"
      >
        <div class="d-flex flex-wrap align-items-center gap-1 mb-1">
          <strong>{{ $t('automation.sessions.editor.aiSteps.step', 'Step') }} {{ step.stepID }}</strong>
          <span
            v-if="step.trace.agent"
            class="badge bg-light text-dark border"
          >{{ step.trace.agent }}</span>
          <span
            v-if="step.trace.model"
            class="badge bg-light text-dark border"
          >{{ step.trace.model }}</span>
          <span class="badge bg-light text-dark border">{{ formatMs(step.trace.durationMs || 0) }}</span>
          <span
            v-if="step.trace.promptTokens || step.trace.completionTokens"
            class="badge bg-light text-dark border"
          >{{ step.trace.promptTokens || 0 }} / {{ step.trace.completionTokens || 0 }} {{ $t('automation.sessions.editor.aiSteps.tokens', 'tokens') }}</span>
          <span
            v-if="step.trace.attempts > 1"
            class="badge bg-warning text-dark"
          >{{ step.trace.attempts }} {{ $t('automation.sessions.editor.aiSteps.attempts', 'attempts') }}</span>
        </div>

        <div
          v-if="step.trace.error || step.error"
          class="text-danger small"
        >
          {{ step.trace.error || step.error }}
        </div>

        <div
          v-if="(step.trace.tools || []).length"
          class="small text-muted"
        >
          {{ $t('automation.sessions.editor.aiSteps.tools', 'Tools') }}: {{ step.trace.tools.join(', ') }}
        </div>

        <details
          v-if="(step.trace.rejected || []).length"
          class="small mt-1"
        >
          <summary>{{ $t('automation.sessions.editor.aiSteps.rejected', 'Rejected answers') }} ({{ step.trace.rejected.length }})</summary>
          <ul class="mb-0 ps-3">
            <li
              v-for="(r, rix) in step.trace.rejected"
              :key="rix"
            >
              {{ r }}
            </li>
          </ul>
        </details>

        <details
          v-if="step.trace.prompt"
          class="small mt-1"
        >
          <summary>{{ $t('automation.sessions.editor.aiSteps.prompt', 'Prompt') }}</summary>
          <pre class="bg-light border rounded p-2 mb-0 ai-pre">{{ step.trace.prompt }}</pre>
        </details>

        <details
          v-if="step.trace.response"
          class="small mt-1"
        >
          <summary>{{ $t('automation.sessions.editor.aiSteps.response', 'Response') }}</summary>
          <pre class="bg-light border rounded p-2 mb-0 ai-pre">{{ step.trace.response }}</pre>
        </details>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { aiSteps } from './typed'

const props = defineProps({
  // the session's stacktrace (recorded when tracing is on for the workflow)
  stacktrace: { type: Array, default: () => [] },
})

const steps = computed(() => aiSteps(props.stacktrace))

function formatMs (ms) {
  return ms >= 1000 ? `${(ms / 1000).toFixed(1)} s` : `${ms} ms`
}
</script>

<style scoped>
.ai-pre {
  max-height: 25vh;
  overflow: auto;
  white-space: pre-wrap;
}
</style>
