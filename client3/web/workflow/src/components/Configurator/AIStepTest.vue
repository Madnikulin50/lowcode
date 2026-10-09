<template>
  <div
    :id="modalId"
    class="modal fade"
    tabindex="-1"
  >
    <div class="modal-dialog modal-lg modal-dialog-scrollable">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">
            {{ t('steps.function.configurator.test.title') }}: {{ functionRef }}
          </h5>
          <button
            type="button"
            class="btn-close"
            data-bs-dismiss="modal"
          />
        </div>

        <div class="modal-body">
          <p class="text-muted small">
            {{ t('steps.function.configurator.test.hint') }}
          </p>

          <label class="form-label text-primary">{{ t('steps.function.configurator.test.arguments') }}</label>
          <textarea
            v-model="argsText"
            rows="8"
            class="form-control font-monospace mb-1"
            :class="{ 'is-invalid': !!argsError }"
            spellcheck="false"
          />
          <div
            v-if="argsError"
            class="invalid-feedback d-block"
          >
            {{ argsError }}
          </div>
          <div
            v-else-if="expressionArgs.length"
            class="form-text"
          >
            {{ t('steps.function.configurator.test.expressions') }}: <var>{{ expressionArgs.join(', ') }}</var>
          </div>

          <div
            v-if="error"
            class="alert alert-danger mt-3 mb-0"
          >
            {{ error }}
          </div>

          <div
            v-if="running || live"
            class="border rounded p-2 mt-3"
          >
            <div class="small text-muted mb-1">
              <span
                v-if="running"
                class="spinner-border spinner-border-sm me-1"
              />
              {{ progressLine }}
            </div>
            <pre
              v-if="live"
              class="mb-0 trace-pre"
            >{{ live }}</pre>
          </div>

          <template v-if="result">
            <div
              v-if="trace"
              class="border rounded p-2 mt-3 small"
            >
              <div class="d-flex flex-wrap align-items-center gap-1">
                <span class="fw-bold text-muted text-uppercase me-1">{{ t('steps.function.configurator.test.trace') }}</span>
                <span
                  v-if="trace.agent"
                  class="badge bg-light text-dark border"
                >{{ trace.agent }}</span>
                <span
                  v-if="trace.model"
                  class="badge bg-light text-dark border"
                >{{ trace.model }}</span>
                <span class="badge bg-light text-dark border">{{ formatMs(trace.durationMs || 0) }}</span>
                <span
                  v-if="trace.promptTokens || trace.completionTokens"
                  class="badge bg-light text-dark border"
                >{{ trace.promptTokens || 0 }} / {{ trace.completionTokens || 0 }} {{ t('steps.function.configurator.test.tokens') }}</span>
                <span
                  v-if="trace.attempts > 1"
                  class="badge bg-warning text-dark"
                >{{ trace.attempts }} {{ t('steps.function.configurator.test.attempts') }}</span>
              </div>
              <details
                v-if="(trace.rejected || []).length"
                class="mt-1"
              >
                <summary>{{ t('steps.function.configurator.test.rejected') }} ({{ trace.rejected.length }})</summary>
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
                <summary>{{ t('steps.function.configurator.test.prompt') }}</summary>
                <pre class="bg-light border rounded p-2 mb-0 trace-pre">{{ trace.prompt }}</pre>
              </details>
            </div>

            <label class="form-label text-primary mt-3">{{ t('steps.function.configurator.test.results') }}</label>
            <pre class="bg-light border rounded p-2 mb-0 trace-pre">{{ resultText }}</pre>
          </template>
        </div>

        <div class="modal-footer">
          <button
            v-if="running"
            type="button"
            class="btn btn-outline-danger"
            @click="stop"
          >
            {{ t('steps.function.configurator.test.stop') }}
          </button>
          <button
            type="button"
            class="btn btn-primary"
            :disabled="running"
            @click="run"
          >
            <span
              v-if="running"
              class="spinner-border spinner-border-sm me-1"
            />
            {{ t('steps.function.configurator.test.run') }}
          </button>
          <button
            type="button"
            class="btn btn-outline-secondary"
            data-bs-dismiss="modal"
          >
            {{ t('cancel') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { seedTestArgs } from '../../lib/ai-step-test'

const { t } = useI18n()
const $AutomationAPI = inject('automationAPI', {})

const props = defineProps({
  functionRef: { type: String, required: true },
  // the step's arguments as the configurator holds them
  args: { type: Array, default: () => [] },
})

const modalId = 'ai-step-test-modal'
const argsText = ref('{}')
const argsError = ref('')
const error = ref('')
const result = ref(null)
const running = ref(false)
const expressionArgs = ref([])

// what the model has produced so far, and what it is doing
const live = ref('')
const status = ref('')
const attempt = ref(0)
let controller = null

const progressLine = computed(() => {
  if (status.value === 'warming') return t('steps.function.configurator.test.warming')
  if (status.value === 'using-tools') return t('steps.function.configurator.test.usingTools')
  if (attempt.value > 1) return t('steps.function.configurator.test.attemptN', { n: attempt.value })
  return running.value ? t('steps.function.configurator.test.working') : ''
})

const trace = computed(() => (result.value && result.value.trace) || null)
const resultText = computed(() => {
  if (!result.value) return ''
  const { trace: _trace, ...rest } = result.value
  return JSON.stringify(rest, null, 2)
})

function formatMs (ms) {
  return ms >= 1000 ? `${(ms / 1000).toFixed(1)} s` : `${ms} ms`
}

// called by the configurator right before it shows the modal
function prepare () {
  const { values, expressions } = seedTestArgs(props.args)
  argsText.value = JSON.stringify(values, null, 2)
  expressionArgs.value = expressions
  argsError.value = ''
  error.value = ''
  result.value = null
  live.value = ''
  status.value = ''
  attempt.value = 0
}

function stop () {
  if (controller) controller.abort()
}

async function run () {
  argsError.value = ''
  error.value = ''
  result.value = null

  let args
  try {
    args = JSON.parse(argsText.value || '{}')
  } catch (e) {
    argsError.value = e.message
    return
  }

  live.value = ''
  status.value = ''
  attempt.value = 0
  controller = new AbortController()
  running.value = true
  try {
    const last = await $AutomationAPI.aiStepTestStream({ ref: props.functionRef, args }, (ev) => {
      if (ev.status) status.value = ev.status
      if (ev.attempt) {
        // a new try starts from a blank page: the last answer was rejected
        attempt.value = ev.attempt
        if (ev.attempt > 1) live.value = ''
      }
      if (ev.token) live.value += ev.token
    }, controller.signal)
    result.value = last.result
  } catch (e) {
    // stopping is the user's own doing, not a failure
    if (e && e.name !== 'AbortError') {
      error.value = (e.response && e.response.data && e.response.data.error && e.response.data.error.message) || String(e.message || e)
    }
  } finally {
    running.value = false
    controller = null
  }
}

defineExpose({ prepare, modalId })
</script>

<style scoped>
.trace-pre {
  max-height: 25vh;
  overflow: auto;
  white-space: pre-wrap;
}
</style>
