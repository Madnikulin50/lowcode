<template>
  <div class="mb-3" :data-test-id="getFieldCypressId(label)" :class="formGroupStyleClasses">
    <div v-if="!valueOnly" :class="labelColClass">
      <div class="d-flex align-items-center text-primary p-0">
        <span :title="label" class="d-flex">{{ label }}</span>
        <c-hint :tooltip="hint" />
        <slot name="tools" />
      </div>
      <div class="small text-muted" :class="{ 'mb-1': description }">{{ description }}</div>
    </div>
    <div :class="contentColClass">
    <template v-if="showJSON && !field.isMulti">
      <div
        v-if="jsonError"
        class="alert alert-warning py-1 px-2 small mb-2"
      >{{ t('kind.string.json.invalid') }}</div>
      <div
        v-if="layout === 'pretty' || jsonError"
        class="mb-2"
      >
        <textarea
          v-model="prettyText"
          class="form-control font-monospace"
          rows="8"
          @change="commitPretty"
        />
      </div>
      <template v-else>
        <div
          v-if="isKeyedObject"
          class="table-responsive mb-2"
        >
          <table class="table table-sm align-middle mb-0">
            <thead>
              <tr>
                <th>{{ t('kind.string.json.key') }}</th>
                <th>{{ t('kind.string.json.value') }}</th>
                <th />
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, index) in kvRows"
                :key="index"
              >
                <td>
                  <input
                    v-model="row.key"
                    type="text"
                    class="form-control form-control-sm"
                    @change="commitObject"
                  >
                </td>
                <td>
                  <input
                    v-model="row.value"
                    type="text"
                    class="form-control form-control-sm"
                    @change="commitObject"
                  >
                </td>
                <td class="text-end" style="width: 2.5rem;">
                  <button
                    type="button"
                    class="btn btn-link btn-sm text-danger p-0"
                    :title="t('kind.string.json.remove')"
                    @click="removeKv(index)"
                  >
                    <font-awesome-icon :icon="['fas', 'times']" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div
          v-else
          class="table-responsive mb-2"
        >
          <table
            v-if="objectRows.length"
            class="table table-sm align-middle mb-0"
          >
            <thead>
              <tr>
                <th
                  v-for="col in columns"
                  :key="col"
                >{{ col }}</th>
                <th />
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, index) in objectRows"
                :key="index"
              >
                <td
                  v-for="col in columns"
                  :key="col"
                >
                  <input
                    v-model="row[col]"
                    type="text"
                    class="form-control form-control-sm"
                    @change="commitRows"
                  >
                </td>
                <td class="text-end" style="width: 2.5rem;">
                  <button
                    type="button"
                    class="btn btn-link btn-sm text-danger p-0"
                    :title="t('kind.string.json.remove')"
                    @click="removeRow(index)"
                  >
                    <font-awesome-icon :icon="['fas', 'times']" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          <p
            v-else
            class="text-muted small mb-0"
          >{{ t('kind.string.json.empty') }}</p>
        </div>
        <button
          type="button"
          class="btn btn-outline-primary btn-sm"
          @click="addItem"
        >
          <font-awesome-icon :icon="['fas', 'plus']" class="me-1" />
          {{ t('kind.string.json.add') }}
        </button>
      </template>
      <FieldErrors :errors="errors" />
    </template>

    <multi v-else-if="field.isMulti" v-slot="ctx" v-model:value="value" :errors="errors">
      <CRichTextInput
        v-if="field.options.useRichTextEditor"
        :value="value[ctx.index]"
        :labels="richTextLabels"
        @input="setMultiValue($event, ctx.index)"
      />
      <textarea
        v-else-if="field.options.multiLine"
        :value="value[ctx.index]"
        class="form-control"
        @input="setMultiValue($event.target.value, ctx.index)"
      ></textarea>
      <input
        v-else
        :value="value[ctx.index]"
        type="text"
        class="form-control form-control-sm"
        @input="setMultiValue($event.target.value, ctx.index)"
      />
    </multi>

    <template v-else>
      <CRichTextInput
        v-if="field.options.useRichTextEditor"
        v-model="value"
        :labels="richTextLabels"
      />
      <textarea
        v-else-if="field.options.multiLine"
        v-model="value"
        class="form-control"
      ></textarea>
      <input
        v-else
        v-model="value"
        type="text"
        class="form-control form-control-sm"
      />
      <FieldErrors :errors="errors" />
    </template>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditorBase } from './base'
import { components } from 'corteza-lib/vue/dist'
import FieldErrors from '../errors'
import multi from './multi'
import {
  isJSONDisplay,
  jsonOptions,
  parseJSONValue,
  stringifyJSON,
  prettyJSON,
  asItems,
  detectFields,
  blankRow,
} from 'corteza-webapp-compose/src/lib/json-field'
const { CRichTextInput } = components

const props = defineProps({
  namespace: { type: Object, required: true },
  field: { type: Object, required: true },
  record: { type: Object, required: true },
  errors: { type: Object, required: true },
  valueOnly: { type: Boolean, default: false },
  horizontal: { type: Boolean, default: false },
  extraOptions: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['change', 'update:preventPopoverClose'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { value, formGroupStyleClasses, labelColClass, contentColClass, label, hint, description, getFieldCypressId, setMultiValue } = useEditorBase(props, emit)

const showJSON = computed(() => isJSONDisplay(props.field))
const opts = computed(() => jsonOptions(props.field))
const layout = computed(() => opts.value.layout)
const parsed = computed(() => parseJSONValue(value.value))
const jsonError = computed(() => showJSON.value && !parsed.value.ok && !parsed.value.empty)
const isKeyedObject = computed(() => parsed.value.ok && parsed.value.value && typeof parsed.value.value === 'object' && !Array.isArray(parsed.value.value))

const objectRows = ref([])
const kvRows = ref([])
const prettyText = ref('')
const columns = ref([])

watch(() => value.value, () => {
  if (!showJSON.value) return
  const next = parseJSONValue(value.value)
  prettyText.value = next.ok ? prettyJSON(next.value) : String(value.value || '')
  if (!next.ok) return
  if (next.value && typeof next.value === 'object' && !Array.isArray(next.value)) {
    kvRows.value = Object.entries(next.value).map(([key, val]) => ({
      key,
      value: typeof val === 'object' ? stringifyJSON(val) : String(val ?? ''),
    }))
    return
  }
  const items = asItems(next.value).map(item => {
    if (item && typeof item === 'object' && !Array.isArray(item)) return { ...item }
    return { value: item == null ? '' : String(item) }
  })
  const cols = opts.value.fields.length ? opts.value.fields : (detectFields(items).length ? detectFields(items) : ['value'])
  columns.value = cols
  objectRows.value = items.map(item => {
    const row = {}
    for (const col of cols) row[col] = item[col] == null ? '' : String(item[col])
    return row
  })
}, { immediate: true })

function commitRows () {
  const rows = objectRows.value.map(row => {
    const out = {}
    let primitive = columns.value.length === 1 && columns.value[0] === 'value'
    for (const col of columns.value) {
      const raw = row[col]
      out[col] = coerceCell(raw)
    }
    return primitive ? out.value : out
  }).filter(row => {
    if (row == null || row === '') return false
    if (typeof row === 'object') return Object.values(row).some(v => v !== '' && v != null)
    return true
  })
  value.value = stringifyJSON(rows)
}

function commitObject () {
  const out = {}
  kvRows.value.forEach(({ key, value: val }) => {
    const k = String(key || '').trim()
    if (!k) return
    out[k] = coerceCell(val)
  })
  value.value = stringifyJSON(out)
}

function commitPretty () {
  const next = parseJSONValue(prettyText.value)
  if (!next.ok) {
    value.value = prettyText.value
    return
  }
  value.value = stringifyJSON(next.value)
}

function addItem () {
  if (isKeyedObject.value) {
    kvRows.value.push({ key: '', value: '' })
    return
  }
  if (!columns.value.length) columns.value = opts.value.fields.length ? opts.value.fields : ['value']
  objectRows.value.push(blankRow(columns.value))
}

function removeRow (index) {
  objectRows.value.splice(index, 1)
  commitRows()
}

function removeKv (index) {
  kvRows.value.splice(index, 1)
  commitObject()
}

function coerceCell (raw) {
  if (raw == null) return ''
  const s = String(raw).trim()
  if (s === '') return ''
  if (/^-?\d+(\.\d+)?$/.test(s)) return Number(s)
  if (s === 'true') return true
  if (s === 'false') return false
  if ((s.startsWith('{') && s.endsWith('}')) || (s.startsWith('[') && s.endsWith(']'))) {
    try { return JSON.parse(s) } catch { return s }
  }
  return s
}

const richTextLabels = computed(() => ({
  urlPlaceholder: t('content.urlPlaceholder'),
  ok: t('content.ok'),
  emojiPicker: {
    search: t('content.emojiPicker.search'),
    searchResults: t('content.emojiPicker.searchResults'),
    frequentlyUsed: t('content.emojiPicker.frequentlyUsed'),
    noResults: t('content.emojiPicker.noResults'),
    quickReactions: t('content.emojiPicker.quickReactions'),
  },
}))
</script>
