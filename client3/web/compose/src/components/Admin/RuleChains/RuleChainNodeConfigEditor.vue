<template>
  <div>
    <p
      v-if="description"
      class="small text-muted mb-2"
    >
      {{ description }}
    </p>

    <div
      v-if="!nodeType"
      class="small text-muted"
    >
      {{ $t('rulechain.edit.nodes.config.noType') }}
    </div>

    <template v-else>
      <div class="btn-group btn-group-sm mb-2" role="group">
        <button
          type="button"
          class="btn"
          :class="tab === 'form' ? 'btn-primary' : 'btn-outline-secondary'"
          :disabled="!!jsonError"
          @click="tab = 'form'"
        >
          {{ $t('rulechain.edit.nodes.config.form') }}
        </button>
        <button
          type="button"
          class="btn"
          :class="tab === 'json' ? 'btn-primary' : 'btn-outline-secondary'"
          @click="openJsonTab"
        >
          {{ $t('rulechain.edit.nodes.config.json') }}
        </button>
      </div>

      <div v-show="tab === 'form'">
        <RuleChainConfigField
          v-for="field in fields"
          :key="field.key"
          :field="field"
          :model-value="config[field.key]"
          :scope="displayScope"
          @update:model-value="setField(field, $event)"
        />
        <div
          v-if="!fields.length"
          class="small text-muted"
        >
          {{ $t('rulechain.edit.nodes.config.noFields') }}
        </div>
      </div>

      <div
        v-if="testable"
        v-show="tab === 'form'"
        class="border rounded p-2 mt-3 bg-light"
        data-test="node-test"
      >
        <label class="form-label small fw-bold text-muted mb-1">
          {{ $t('rulechain.nodeTest.title', 'Try this node') }}
        </label>
        <textarea
          v-model="testInput"
          rows="3"
          class="form-control form-control-sm font-monospace mb-2"
          spellcheck="false"
          :placeholder="$t('rulechain.nodeTest.inputPlaceholder', 'Sample input as JSON - its fields fill the variables used in the prompt')"
        />
        <button
          v-if="testRunning"
          type="button"
          class="btn btn-sm btn-outline-danger me-1"
          @click="stopNodeTest"
        >
          {{ $t('rulechain.nodeTest.stop', 'Stop') }}
        </button>
        <button
          type="button"
          class="btn btn-sm btn-outline-primary"
          :disabled="testRunning"
          @click="runNodeTest"
        >
          <span
            v-if="testRunning"
            class="spinner-border spinner-border-sm me-1"
          />
          {{ $t('rulechain.nodeTest.run', 'Run with the real model') }}
        </button>
        <span class="small text-muted ms-2">
          {{ $t('rulechain.nodeTest.hint', 'Nothing is saved; data changes by the agent are switched off.') }}
        </span>

        <div
          v-if="testRunning || testLive"
          class="border rounded bg-white p-2 mt-2 small"
        >
          <div class="text-muted mb-1">
            {{ testProgress }}
          </div>
          <pre
            v-if="testLive"
            class="mb-0"
            style="max-height: 20vh; overflow: auto; white-space: pre-wrap;"
          >{{ testLive }}</pre>
        </div>

        <div
          v-if="testError"
          class="alert alert-danger py-1 small mt-2 mb-0"
        >
          {{ testError }}
        </div>

        <template v-if="testResult">
          <div
            v-if="testResult.error"
            class="alert alert-danger py-1 small mt-2 mb-0"
          >
            {{ testResult.error }}
          </div>
          <RuleChainAiTrace :trace="testResult.trace" />
          <pre
            v-if="testResult.output"
            class="bg-white border rounded p-2 mt-2 mb-0 small"
            style="max-height: 20vh; overflow: auto; white-space: pre-wrap;"
          >{{ JSON.stringify(testResult.output, null, 2) }}</pre>
        </template>
      </div>

      <div v-show="tab === 'json'">
        <textarea
          v-model="jsonText"
          class="form-control form-control-sm font-monospace"
          :class="{ 'is-invalid': !!jsonError }"
          :rows="jsonRows"
          spellcheck="false"
          @input="onJsonEdit"
        />
        <div
          v-if="jsonError"
          class="invalid-feedback d-block"
        >
          {{ $t('rulechain.edit.nodes.config.invalidJson') }}: {{ jsonError }}
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { reactive, ref, watch, computed, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import RuleChainConfigField from './RuleChainConfigField.vue'
import RuleChainAiTrace from './RuleChainAiTrace.vue'
import {
  parseConfigText,
  stringifyConfig,
  applyTypeDefaults,
  serializeConfig,
  scopeWithDefaults,
  fieldsFromNodeType,
  resolveFields,
} from './rulechainConfig'

const { t } = useI18n()

const props = defineProps({
  modelValue: { type: String, default: '{}' },
  nodeType: { type: String, default: '' },
  fields: { type: Array, default: () => [] },
  nodeSchema: { type: Object, default: null },
  description: { type: String, default: '' },
  jsonRows: { type: Number, default: 10 },
})

const emit = defineEmits(['update:modelValue'])

// the AI nodes can be tried on their own (server: POST /admin/rulechain/node-test)
const testable = computed(() => props.nodeType === 'ai' || props.nodeType === 'ai.operation')
const testInput = ref('{}')
const testRunning = ref(false)
const testResult = ref(null)
const testError = ref('')
const testLive = ref('')
const testStatus = ref('')
const testAttempt = ref(0)
let testController = null

const testProgress = computed(() => {
  if (testStatus.value === 'warming') return t('rulechain.nodeTest.warming', 'The model is loading...')
  if (testStatus.value === 'using-tools') return t('rulechain.nodeTest.usingTools', 'The agent is using tools...')
  if (testAttempt.value > 1) return t('rulechain.nodeTest.attemptN', { n: testAttempt.value })
  return t('rulechain.nodeTest.working', 'Waiting for the model...')
})

function stopNodeTest () {
  if (testController) testController.abort()
}

async function runNodeTest () {
  testError.value = ''
  testResult.value = null

  let input
  try {
    input = JSON.parse(testInput.value || '{}')
  } catch (e) {
    testError.value = t('rulechain.test.input.invalid', 'Input must be valid JSON')
    return
  }

  const parsed = parseConfigText(jsonText.value)
  if (!parsed.ok) {
    testError.value = `${t('rulechain.edit.nodes.config.invalidJson')}: ${parsed.error}`
    return
  }

  testLive.value = ''
  testStatus.value = ''
  testAttempt.value = 0
  testController = new AbortController()
  testRunning.value = true
  try {
    const last = await window.__composeAPI.ruleChainNodeTestStream({
      type: props.nodeType,
      config: { ...config, ...serializeConfig(config, fields.value) },
      input,
    }, (ev) => {
      if (ev.status) testStatus.value = ev.status
      if (ev.attempt) {
        // a new try starts from a blank page: the last answer was rejected
        testAttempt.value = ev.attempt
        if (ev.attempt > 1) testLive.value = ''
      }
      if (ev.token) testLive.value += ev.token
    }, testController.signal)
    testResult.value = last.result
  } catch (e) {
    // stopping is the user's own doing, not a failure
    if (e && e.name !== 'AbortError') {
      testError.value = (e.response && e.response.data && e.response.data.error && e.response.data.error.message) || String(e.message || e)
    }
  } finally {
    testRunning.value = false
    testController = null
  }
}

const tab = ref('form')
const config = reactive({})
const jsonText = ref('{}')
const jsonError = ref('')
let lastEmitted
let hydrating = true

const fields = computed(() => {
  const fromProp = Array.isArray(props.fields) ? props.fields : []
  const fromType = fieldsFromNodeType(props.nodeSchema)
  const snapshot = { ...config }
  return resolveFields(fromProp.length ? fromProp : fromType, snapshot, props.nodeType)
})
const displayScope = computed(() => scopeWithDefaults(config, fields.value))

function replaceConfig (obj) {
  for (const key of Object.keys(config)) delete config[key]
  Object.assign(config, obj && typeof obj === 'object' ? obj : {})
}

function loadFromText (text, { forceJsonTab = false } = {}) {
  hydrating = true
  const parsed = parseConfigText(text)
  jsonText.value = (text && String(text).trim()) ? text : '{}'
  if (!parsed.ok) {
    jsonError.value = parsed.error
    if (forceJsonTab || tab.value === 'form') tab.value = 'json'
    nextTick(() => {
      requestAnimationFrame(() => { hydrating = false })
    })
    return
  }
  jsonError.value = ''
  replaceConfig(parsed.value)
  nextTick(() => {
    requestAnimationFrame(() => { hydrating = false })
  })
}

function emitConfig (text) {
  lastEmitted = text
  emit('update:modelValue', text)
}

function emitFromForm () {
  if (hydrating) return
  const text = stringifyConfig(serializeConfig(config, fields.value))
  jsonText.value = text
  jsonError.value = ''
  emitConfig(text)
}

function setField (field, value) {
  if (hydrating) return
  if (value === undefined) {
    delete config[field.key]
  } else {
    config[field.key] = value
  }
  emitFromForm()
}

function onJsonEdit () {
  const text = jsonText.value
  const parsed = parseConfigText(text)
  if (!parsed.ok) {
    jsonError.value = parsed.error
    emitConfig(text)
    return
  }
  jsonError.value = ''
  replaceConfig(parsed.value)
  emitConfig(text)
}

function openJsonTab () {
  if (!jsonError.value) {
    jsonText.value = stringifyConfig({
      ...config,
      ...serializeConfig(config, fields.value),
    })
  }
  tab.value = 'json'
}

watch(() => props.modelValue, (v) => {
  if (v === lastEmitted) return
  loadFromText(v)
}, { immediate: true })

watch(() => props.nodeType, (next, prev) => {
  if (hydrating) return
  if (!prev || prev === next) return
  if (!fields.value.length) return
  const nextConfig = applyTypeDefaults(config, fields.value)
  replaceConfig(nextConfig)
  emitFromForm()
  tab.value = 'form'
}, { flush: 'post' })
</script>
