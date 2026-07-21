<template>
  <div>
    <div class="form-check mb-2">
      <input
        id="isRequired"
        v-model="field.isRequired"
        type="checkbox"
        class="form-check-input"
        :disabled="!field.cap.required || showValueExpr"
      />
      <label class="form-check-label" for="isRequired">{{ t('label.required') }}</label>
    </div>
    <div class="form-check mb-2">
      <input
        id="isMulti"
        v-model="field.isMulti"
        type="checkbox"
        class="form-check-input"
        :disabled="!field.cap.multi"
      />
      <label class="form-check-label" for="isMulti">{{ t('label.multi') }}</label>
    </div>
    <div class="form-check mb-2">
      <input
        id="showValueExpr"
        v-model="showValueExpr"
        type="checkbox"
        class="form-check-input"
        :disabled="field.isRequired || defaultValueEnabled"
      />
      <label class="form-check-label" for="showValueExpr">{{ t('valueExpr.label') }}</label>
    </div>
    <div v-if="showDefaultValue" class="form-check mb-2">
      <input
        id="defaultValue"
        type="checkbox"
        class="form-check-input"
        :checked="defaultValueEnabled"
        :disabled="!!showValueExpr"
        @change="toggleDefaultValue()"
      />
      <label class="form-check-label" for="defaultValue">{{ t('defaultValue') }}</label>
    </div>

    <hr />

    <div v-if="showValueExpr" class="mb-3 mt-2">
      <label class="form-label text-primary">{{ t('valueExpr.label') }}</label>
      <div class="input-group input-group-sm">
        <span class="input-group-text">ƒ</span>
        <input
          v-model="field.expressions.value"
          type="text"
          class="form-control form-control-sm"
          :placeholder="t('valueExpr.placeholder')"
        />
        <a
          :href="documentationURL"
          target="_blank"
          class="btn btn-outline-secondary d-flex justify-content-center align-items-center text-primary"
        >?</a>
      </div>
      <div class="form-text">{{ t('valueExpr.description') }}</div>
    </div>

    <div v-else-if="showDefaultField" class="mb-3">
      <label class="form-label text-primary mb-0">{{ t('defaultFieldValue') }}</label>
      <FieldEditor
        value-only
        v-bind="mock"
        class="mb-1"
      />
    </div>

    <hr v-if="showValueExpr || showDefaultField" />

    <div class="mb-3 mt-2">
      <label class="form-label text-primary">{{ t(`options.description.label.${noDescriptionEdit ? 'default' : 'view'}`) }}</label>
      <div class="input-group input-group-sm">
        <input
          v-model="field.options.description.view"
          type="text"
          class="form-control form-control-sm"
          :placeholder="t(`options.description.placeholder.${noDescriptionEdit ? 'default' : 'view'}`)"
        />
        <FieldTranslator
          v-if="field"
          :field="field"
          :module="module"
          :disabled="isNew"
          highlight-key="meta.description.view"
        />
      </div>
    </div>

    <div v-if="!noDescriptionEdit" class="mb-3 mt-2">
      <label class="form-label text-primary">{{ t('options.description.label.edit') }}</label>
      <div class="input-group input-group-sm">
        <input
          v-model="field.options.description.edit"
          type="text"
          class="form-control form-control-sm"
          :placeholder="t('options.description.placeholder.edit')"
        />
        <FieldTranslator
          v-if="field"
          :field="field"
          :module="module"
          :disabled="isNew"
          highlight-key="meta.description.edit"
        />
      </div>
    </div>

    <div class="form-check mb-2">
      <input
        id="noDescriptionEdit"
        type="checkbox"
        class="form-check-input"
        tabindex="-1"
        :checked="noDescriptionEdit"
        @change="field.options.description.edit = $event.target.checked ? undefined : ''"
      />
      <label class="form-check-label" for="noDescriptionEdit">{{ t('options.description.same') }}</label>
    </div>

    <hr />

    <div class="mb-3 mt-2">
      <label class="form-label text-primary">{{ t(`options.hint.label.${noHintEdit ? 'default' : 'view'}`) }}</label>
      <div class="input-group input-group-sm">
        <input
          v-model="field.options.hint.view"
          type="text"
          class="form-control form-control-sm"
          :placeholder="t(`options.hint.placeholder.${noHintEdit ? 'default' : 'view'}`)"
        />
        <FieldTranslator
          v-if="field"
          :field="field"
          :module="module"
          :disabled="isNew"
          highlight-key="meta.hint.view"
        />
      </div>
    </div>

    <div v-if="!noHintEdit" class="mb-3 mt-2">
      <label class="form-label text-primary">{{ t('options.hint.label.edit') }}</label>
      <div class="input-group input-group-sm">
        <input
          v-model="field.options.hint.edit"
          type="text"
          class="form-control form-control-sm"
          :placeholder="t('options.hint.placeholder.edit')"
        />
        <FieldTranslator
          v-if="field"
          :field="field"
          :module="module"
          :disabled="isNew"
          highlight-key="meta.hint.edit"
        />
      </div>
    </div>

    <div class="form-check mb-2">
      <input
        id="noHintEdit"
        type="checkbox"
        class="form-check-input"
        tabindex="-1"
        :checked="noHintEdit"
        @change="field.options.hint.edit = $event.target.checked ? undefined : ''"
      />
      <label class="form-check-label" for="noHintEdit">{{ t('options.hint.same') }}</label>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose, validator, NoID } from 'corteza-lib/js/dist'
import FieldEditor from '../Editor'
import FieldTranslator from 'corteza-webapp-compose/src/components/Admin/Module/FieldTranslator'

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  module: { type: compose.Module, required: true },
  field: { type: compose.ModuleField, required: true },
})

const { t } = useI18n({ useScope: 'global', messages: {} })

const showValueExpr = ref(false)

const mock = ref({
  show: true,
  namespace: undefined,
  module: undefined,
  field: undefined,
  record: undefined,
  errors: new validator.Validated(),
})

const noDescriptionEdit = computed(() => props.field.options.description.edit === undefined)
const noHintEdit = computed(() => props.field.options.hint.edit === undefined)
const showDefaultValue = computed(() => !['File'].includes(props.field.kind))
const defaultValueEnabled = computed(() => !!props.field.defaultValue && props.field.defaultValue.length > 0)
const showDefaultField = computed(() => defaultValueEnabled.value && mock.value.show && mock.value.field)
const isNew = computed(() => props.field.fieldID === NoID)

const documentationURL = computed(() => {
  const [year, month] = (typeof VERSION !== 'undefined' ? VERSION : '2024.0').split('.')
  return `https://docs.cortezaproject.org/corteza-docs/${year}.${month}/integrator-guide/expr/index.html`
})

watch(defaultValueEnabled, (val) => {
  if (val) showValueExpr.value = false
})

watch(() => mock.value.record && mock.value.record.values, {
  handler (vals) {
    if (!vals) return
    const { defValField: dv } = vals
    let arr = dv
    if (!Array.isArray(arr)) arr = [arr]
    else if (!arr.length) arr = [undefined]
    props.field.defaultValue = arr.map(v => {
      if (v !== undefined && v !== null && v.toString) v = v.toString()
      const def = { name: props.field.name }
      if (v) def.value = v
      return def
    })
  },
  deep: true,
})

watch(() => props.field.options, {
  handler (options) {
    if (mock.value.field) mock.value.field.options = options
  },
  deep: true,
})

watch(() => props.field.isMulti, () => {
  if (props.field.defaultValue && props.field.defaultValue.length) {
    initMocks(props.field.defaultValue)
  }
})

let { defaultValue, expressions } = props.field
if (!defaultValue) defaultValue = []
if (defaultValue.length) initMocks(defaultValue)
if (!props.field.options.hint.edit) props.field.options.hint.edit = undefined
if (!props.field.options.description.edit) props.field.options.description.edit = undefined
showValueExpr.value = expressions && expressions.value && expressions.value.length > 0
if (!props.field.expressions.value) props.field.expressions.value = ''

onBeforeUnmount(() => {
  if (showValueExpr.value) {
    props.field.required = false
    props.field.defaultValue = []
  } else {
    props.field.expressions.value = undefined
  }
})

function initMocks (defaultValue = []) {
  if (props.field.isMulti) {
    defaultValue = defaultValue.map(v => (v || {}).value).filter(v => v)
  } else {
    defaultValue = (defaultValue[0] || {}).value
  }
  mock.value.namespace = props.namespace
  mock.value.field = compose.ModuleFieldMaker(props.field)
  mock.value.field.apply({ label: mock.value.field.label || 'Default value' })
  mock.value.field.apply({ name: 'defValField' })
  mock.value.module = new compose.Module({ fields: [mock.value.field] }, props.namespace)
  mock.value.record = new compose.Record(mock.value.module, { defValField: defaultValue })
}

function toggleDefaultValue () {
  if (defaultValueEnabled.value) props.field.defaultValue = []
  else initMocks()
}
</script>
