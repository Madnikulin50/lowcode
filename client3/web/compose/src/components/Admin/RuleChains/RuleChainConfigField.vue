<template>
  <div
    v-if="visible"
    :class="compact ? 'mb-0' : 'mb-2'"
  >
    <label
      v-if="field.widget !== 'bool'"
      class="form-label small fw-bold text-muted mb-1"
      :title="field.help || undefined"
    >
      {{ field.label }}
      <span
        v-if="field.required"
        class="text-danger"
      >*</span>
    </label>

    <select
      v-if="field.widget === 'enum'"
      class="form-select form-select-sm"
      :class="{ 'is-invalid': showError }"
      :value="stringValue"
      @change="emitString($event.target.value)"
    >
      <option value="">
        —
      </option>
      <option
        v-if="stringValue && !(field.options || []).includes(stringValue)"
        :value="stringValue"
      >
        {{ stringValue }}
      </option>
      <option
        v-for="opt in field.options || []"
        :key="opt"
        :value="opt"
      >
        {{ opt }}
      </option>
    </select>

    <input
      v-else-if="field.widget === 'number'"
      type="number"
      class="form-control form-control-sm"
      :class="{ 'is-invalid': showError }"
      :placeholder="field.placeholder"
      :value="numberDisplay"
      @input="emitNumber($event.target.value)"
      @blur="touched = true"
    >

    <div
      v-else-if="field.widget === 'bool'"
      class="form-check form-switch"
    >
      <input
        class="form-check-input"
        type="checkbox"
        :checked="!!modelValue"
        @change="emit('update:modelValue', $event.target.checked)"
      >
      <label class="form-check-label small">
        {{ field.label }}
      </label>
    </div>

    <textarea
      v-else-if="field.widget === 'textarea'"
      class="form-control form-control-sm font-monospace"
      :class="{ 'is-invalid': showError }"
      :rows="field.rows || 8"
      :placeholder="field.placeholder || (field.template ? '{{variable}}' : '')"
      spellcheck="false"
      :value="stringValue"
      @input="emitString($event.target.value)"
      @blur="touched = true"
    />

    <RuleChainCodeEditor
      v-else-if="field.widget === 'code'"
      :model-value="stringValue"
      :lang="aceLang"
      @update:model-value="emitString($event)"
    />

    <textarea
      v-else-if="field.widget === 'json'"
      class="form-control form-control-sm font-monospace"
      :class="{ 'is-invalid': !!jsonError || showError }"
      :rows="field.rows || 5"
      spellcheck="false"
      :value="jsonText"
      @input="onJsonInput($event.target.value)"
      @blur="touched = true"
    />

    <template v-else-if="field.widget === 'keymap'">
      <div
        v-for="(row, ix) in keymapRows"
        :key="ix"
        class="d-flex flex-wrap align-items-end gap-1 border rounded p-2 mb-2"
      >
        <div class="flex-grow-1" style="min-width: 7rem">
          <label class="form-label small fw-bold text-muted mb-1">
            {{ $t('rulechain.edit.nodes.config.key') }}
          </label>
          <input
            class="form-control form-control-sm"
            :placeholder="$t('rulechain.edit.nodes.config.key')"
            :value="row.key"
            @input="updateKeymapRow(ix, 'key', $event.target.value)"
          >
        </div>
        <div class="flex-grow-1" style="min-width: 7rem">
          <label class="form-label small fw-bold text-muted mb-1">
            {{ $t('rulechain.edit.nodes.config.value') }}
          </label>
          <input
            class="form-control form-control-sm"
            :placeholder="$t('rulechain.edit.nodes.config.value')"
            :value="row.value"
            @input="updateKeymapRow(ix, 'value', $event.target.value)"
          >
        </div>
        <button
          type="button"
          class="btn btn-sm btn-outline-danger mb-1"
          :title="$t('rulechain.edit.nodes.config.remove')"
          @click="removeKeymapRow(ix)"
        >
          ×
        </button>
      </div>
      <button
        type="button"
        class="btn btn-sm btn-outline-secondary"
        @click="addKeymapRow"
      >
        + {{ $t('rulechain.edit.nodes.config.addField') }}
      </button>
    </template>

    <template v-else-if="field.widget === 'stringlist'">
      <div
        v-for="(item, ix) in stringList"
        :key="ix"
        class="d-flex gap-1 mb-1"
      >
        <input
          class="form-control form-control-sm"
          :value="item"
          @input="updateStringList(ix, $event.target.value)"
        >
        <button
          type="button"
          class="btn btn-sm btn-outline-danger"
          @click="removeStringList(ix)"
        >
          ×
        </button>
      </div>
      <button
        type="button"
        class="btn btn-sm btn-outline-secondary"
        @click="addStringList"
      >
        + {{ $t('rulechain.edit.nodes.config.addEntry') }}
      </button>
    </template>

    <template v-else-if="field.widget === 'objectlist'">
      <div
        v-for="(item, ix) in objectList"
        :key="ix"
        class="rcf-object-item border rounded px-2 pt-2 pb-1 mb-2"
      >
        <div
          v-for="sub in itemFields"
          :key="sub.key"
          class="rcf-object-col"
          :class="sub.widget === 'number' || sub.widget === 'bool' ? 'rcf-object-col-sm' : ''"
        >
          <RuleChainConfigField
            compact
            :field="sub"
            :model-value="item[sub.key]"
            :scope="item"
            @update:model-value="updateObjectItem(ix, sub, $event)"
          />
        </div>
        <button
          type="button"
          class="btn btn-sm btn-outline-danger rcf-object-remove"
          :title="$t('rulechain.edit.nodes.config.remove')"
          @click="removeObjectItem(ix)"
        >
          ×
        </button>
      </div>
      <button
        type="button"
        class="btn btn-sm btn-outline-secondary"
        @click="addObjectItem"
      >
        + {{ $t('rulechain.edit.nodes.config.addEntry') }}
      </button>
    </template>

    <input
      v-else
      type="text"
      class="form-control form-control-sm"
      :class="{ 'is-invalid': showError }"
      :placeholder="field.placeholder || (field.template ? '{{variable}}' : '')"
      :value="stringValue"
      @input="emitString($event.target.value)"
      @blur="touched = true"
    >

    <div
      v-if="field.help && !compact && !showError && !jsonError"
      class="form-text small"
    >
      {{ field.help }}
    </div>
    <div
      v-if="jsonError"
      class="invalid-feedback d-block"
    >
      {{ jsonError }}
    </div>
    <div
      v-else-if="showError"
      class="invalid-feedback d-block"
    >
      {{ $t('rulechain.edit.nodes.config.required') }}
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { isFieldVisible, isEmptyValue, inferFieldsFromConfig, OBJECT_LIST_PRESETS } from './rulechainConfig'
import RuleChainCodeEditor from './RuleChainCodeEditor.vue'

defineOptions({ name: 'RuleChainConfigField' })

const props = defineProps({
  field: { type: Object, required: true },
  modelValue: { default: undefined },
  scope: { type: Object, default: () => ({}) },
  compact: { type: Boolean, default: false },
})

const emit = defineEmits(['update:modelValue'])

const touched = ref(false)
const jsonError = ref('')
const jsonText = ref('')
const keymapRows = ref([])
const suppressKeymapWatch = ref(false)

const visible = computed(() => isFieldVisible(props.field, props.scope))

const stringValue = computed(() => {
  if (props.modelValue === undefined || props.modelValue === null) return ''
  return String(props.modelValue)
})

const numberDisplay = computed(() => {
  if (props.modelValue === undefined || props.modelValue === null || props.modelValue === '') return ''
  return props.modelValue
})

const showError = computed(() => {
  return !!(props.field.required && touched.value && isEmptyValue(props.modelValue))
})

const aceLang = computed(() => {
  const lang = props.field.lang || 'javascript'
  if (lang === 'go') return 'golang'
  return lang
})

const stringList = computed(() => Array.isArray(props.modelValue) ? props.modelValue : [])
const objectList = computed(() => Array.isArray(props.modelValue) ? props.modelValue : [])
const itemFields = computed(() => {
  const listed = props.field.itemFields
  if (Array.isArray(listed) && listed.length) return listed
  const preset = OBJECT_LIST_PRESETS[props.field.key]
  if (preset) return preset
  const first = objectList.value.find((item) => item && typeof item === 'object' && !Array.isArray(item))
  return first ? inferFieldsFromConfig(first) : []
})

watch(() => props.modelValue, (v) => {
  if (props.field.widget === 'json') {
    if (jsonError.value) return
    jsonText.value = v === undefined ? '' : stringifyJson(v)
  }
  if (props.field.widget === 'keymap' && !suppressKeymapWatch.value) {
    keymapRows.value = objectToRows(v)
  }
}, { immediate: true, deep: true })

function emitString (v) {
  if (v === stringValue.value) return
  emit('update:modelValue', v)
}

function emitNumber (raw) {
  if (raw === '' || raw === null) {
    emit('update:modelValue', undefined)
    return
  }
  const n = Number(raw)
  emit('update:modelValue', Number.isNaN(n) ? undefined : n)
}

function stringifyJson (v) {
  if (v === undefined || v === '') return ''
  try {
    return JSON.stringify(v, null, 2)
  } catch (e) {
    return ''
  }
}

function onJsonInput (text) {
  jsonText.value = text
  const str = (text || '').trim()
  if (!str) {
    jsonError.value = ''
    emit('update:modelValue', undefined)
    return
  }
  try {
    emit('update:modelValue', JSON.parse(str))
    jsonError.value = ''
  } catch (e) {
    jsonError.value = e.message
  }
}

function objectToRows (obj) {
  if (!obj || typeof obj !== 'object' || Array.isArray(obj)) return [{ key: '', value: '' }]
  const rows = Object.entries(obj).map(([key, value]) => ({
    key,
    value: value === undefined || value === null ? '' : String(value),
  }))
  return rows.length ? rows : [{ key: '', value: '' }]
}

function rowsToObject (rows) {
  const out = {}
  for (const row of rows) {
    const key = (row.key || '').trim()
    if (!key) continue
    out[key] = row.value
  }
  return out
}

function emitKeymap () {
  suppressKeymapWatch.value = true
  emit('update:modelValue', rowsToObject(keymapRows.value))
  requestAnimationFrame(() => { suppressKeymapWatch.value = false })
}

function updateKeymapRow (ix, prop, value) {
  const next = keymapRows.value.slice()
  next[ix] = { ...next[ix], [prop]: value }
  keymapRows.value = next
  emitKeymap()
}

function addKeymapRow () {
  keymapRows.value = [...keymapRows.value, { key: '', value: '' }]
}

function removeKeymapRow (ix) {
  const next = keymapRows.value.filter((_, i) => i !== ix)
  keymapRows.value = next.length ? next : [{ key: '', value: '' }]
  emitKeymap()
}

function updateStringList (ix, value) {
  const next = stringList.value.slice()
  next[ix] = value
  emit('update:modelValue', next)
}

function addStringList () {
  emit('update:modelValue', [...stringList.value, ''])
}

function removeStringList (ix) {
  emit('update:modelValue', stringList.value.filter((_, i) => i !== ix))
}

function addObjectItem () {
  const item = {}
  for (const sub of itemFields.value) {
    if (sub.default !== undefined) item[sub.key] = sub.default
    else if (sub.widget === 'number') item[sub.key] = undefined
    else if (sub.widget === 'bool') item[sub.key] = false
    else item[sub.key] = ''
  }
  emit('update:modelValue', [...objectList.value, item])
}

function removeObjectItem (ix) {
  emit('update:modelValue', objectList.value.filter((_, i) => i !== ix))
}

function updateObjectItem (ix, sub, value) {
  const next = objectList.value.map((item, i) => {
    if (i !== ix) return item
    const copy = { ...item }
    if (value === undefined || (isEmptyValue(value) && !sub.required)) {
      delete copy[sub.key]
    } else {
      copy[sub.key] = value
    }
    return copy
  })
  emit('update:modelValue', next)
}
</script>

<style scoped>
.rcf-object-item {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 0.5rem;
}
.rcf-object-col {
  flex: 1 1 8rem;
  min-width: 8rem;
}
.rcf-object-col-sm {
  flex: 0 0 5.75rem;
  min-width: 5.75rem;
  max-width: 7.5rem;
}
.rcf-object-remove {
  flex: 0 0 auto;
  margin-bottom: 0.125rem;
}
</style>
