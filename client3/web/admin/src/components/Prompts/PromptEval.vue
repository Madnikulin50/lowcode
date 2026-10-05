<template>
  <div class="card shadow-sm mt-3">
    <div class="card-header d-flex align-items-center justify-content-between">
      <h5 class="mb-0">{{ $t('automation.prompts.eval.title') }}</h5>
      <div class="d-flex align-items-center gap-2">
        <button
          v-if="running"
          class="btn btn-sm btn-outline-danger"
          @click="stop"
        >
          {{ $t('automation.prompts.eval.stop') }}
        </button>
        <button
          class="btn btn-sm btn-primary"
          :disabled="running || !chosen.length"
          @click="run"
        >
          <span
            v-if="running"
            class="spinner-border spinner-border-sm me-1"
          />
          {{ $t('automation.prompts.eval.run') }}
        </button>
      </div>
    </div>

    <div class="card-body">
      <p class="small text-muted">
        {{ $t('automation.prompts.eval.hint') }}
      </p>

      <div class="mb-3">
        <span class="me-2 small fw-bold">{{ $t('automation.prompts.eval.versions') }}</span>
        <div
          v-for="v in versions"
          :key="v.version"
          class="form-check form-check-inline"
        >
          <input
            :id="'eval-v-' + v.version"
            v-model="chosen"
            class="form-check-input"
            type="checkbox"
            :value="v.version"
            :disabled="running"
          >
          <label
            class="form-check-label"
            :for="'eval-v-' + v.version"
          >
            v{{ v.version }}<span
              v-if="v.active"
              class="badge bg-success ms-1"
            >{{ $t('automation.prompts.active') }}</span>
          </label>
        </div>
      </div>

      <div
        v-if="error"
        class="alert alert-danger py-2"
      >
        {{ error }}
      </div>

      <div
        v-for="v in shown"
        :key="v.version"
        class="mb-3"
      >
        <div class="d-flex align-items-center gap-2 mb-1">
          <strong>v{{ v.version }}</strong>
          <span
            v-if="v.report"
            class="badge"
            :class="v.report.passRate === 1 ? 'bg-success' : v.report.passRate >= 0.5 ? 'bg-warning text-dark' : 'bg-danger'"
          >
            {{ v.report.passed }} / {{ v.report.total }} ({{ Math.round(v.report.passRate * 100) }}%)
          </span>
          <span
            v-if="v.report"
            class="small text-muted"
          >
            {{ formatMs(v.report.durationMs) }}<template v-if="v.report.promptTokens || v.report.completionTokens">
              · {{ v.report.promptTokens || 0 }} / {{ v.report.completionTokens || 0 }} {{ $t('automation.prompts.eval.tokens') }}
            </template>
          </span>
          <span
            v-else-if="running"
            class="small text-muted"
          >{{ $t('automation.prompts.eval.running') }}</span>
        </div>

        <table
          v-if="v.cases.length"
          class="table table-sm align-middle mb-0"
        >
          <tbody>
            <tr
              v-for="(c, ix) in v.cases"
              :key="ix"
            >
              <td style="width: 1.5rem">
                <font-awesome-icon
                  :icon="['fas', c.pass ? 'check-circle' : 'exclamation-circle']"
                  :class="c.pass ? 'text-success' : 'text-danger'"
                />
              </td>
              <td style="width: 25%">
                {{ c.case }}
              </td>
              <td class="small">
                <div
                  v-for="(f, fix) in c.failures || []"
                  :key="fix"
                  class="text-danger"
                >
                  {{ f }}
                </div>
                <details v-if="c.response">
                  <summary class="text-muted">
                    {{ $t('automation.prompts.eval.response') }}
                  </summary>
                  <pre class="mb-0 eval-pre">{{ c.response }}</pre>
                </details>
              </td>
              <td
                class="small text-muted text-end"
                style="width: 5rem"
              >
                {{ formatMs(c.durationMs) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue'

const props = defineProps({
  handle: { type: String, required: true },
  // [{version, active}] - the prompt's versions, oldest first
  versions: { type: Array, default: () => [] },
  // unsaved test cases from the editor (JSON text); empty: use the stored ones
  casesText: { type: String, default: '' },
})

const chosen = ref([])
const running = ref(false)
const error = ref('')
const live = ref({}) // version -> {cases: [], report}
let controller = null

// by default compare what is live against the newest version
watch(() => [props.handle, props.versions], () => {
  const active = props.versions.find(v => v.active)
  const latest = props.versions[props.versions.length - 1]
  chosen.value = [...new Set([active && active.version, latest && latest.version].filter(Boolean))]
  live.value = {}
  error.value = ''
}, { immediate: true })

const shown = computed(() => Object.keys(live.value)
  .map(Number)
  .sort((a, b) => a - b)
  .map(version => ({ version, ...live.value[version] })))

function formatMs (ms) {
  return ms >= 1000 ? `${(ms / 1000).toFixed(1)} s` : `${ms || 0} ms`
}

function stop () {
  if (controller) controller.abort()
}

onBeforeUnmount(stop)

async function run () {
  error.value = ''

  let cases
  if (props.casesText.trim()) {
    try {
      cases = JSON.parse(props.casesText)
    } catch (e) {
      error.value = e.message
      return
    }
  }

  const versions = [...chosen.value].sort((a, b) => a - b)
  live.value = Object.fromEntries(versions.map(v => [v, { cases: [], report: null }]))

  controller = new AbortController()
  running.value = true
  try {
    const last = await window.__AutomationAPI.promptEvalStream(
      { handle: props.handle, versions, cases },
      (ev) => {
        if (ev.case && live.value[ev.version]) live.value[ev.version].cases.push(ev.case)
      },
      controller.signal,
    )

    for (const [key, report] of Object.entries(last.reports || {})) {
      const version = Number(key.replace(/^v/, ''))
      if (live.value[version]) live.value[version].report = report
    }
  } catch (e) {
    // stopping is the user's own doing, not a failure
    if (e && e.name !== 'AbortError') error.value = e.message || String(e)
  } finally {
    running.value = false
    controller = null
  }
}
</script>

<style scoped>
.eval-pre {
  max-height: 12rem;
  overflow: auto;
  white-space: pre-wrap;
}
</style>
