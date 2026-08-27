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
import RuleChainConfigField from './RuleChainConfigField.vue'
import {
  parseConfigText,
  stringifyConfig,
  applyTypeDefaults,
  serializeConfig,
  scopeWithDefaults,
  fieldsFromNodeType,
  resolveFields,
} from './rulechainConfig'

const props = defineProps({
  modelValue: { type: String, default: '{}' },
  nodeType: { type: String, default: '' },
  fields: { type: Array, default: () => [] },
  nodeSchema: { type: Object, default: null },
  description: { type: String, default: '' },
  jsonRows: { type: Number, default: 10 },
})

const emit = defineEmits(['update:modelValue'])

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
