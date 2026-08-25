<template>
  <div class="tab-pane">
    <div class="row g-0">
      <div class="col-12">
        <div
          v-for="(v, i) in variables"
          :key="i"
          class="mb-2"
        >
          <button
            class="btn btn-outline-secondary me-1"
            :class="{ active: i === selectedIndex }"
            @click="selectedIndex = i"
          >
            {{ $t('label.edit') }}
          </button>
          <button
            class="btn btn-outline-danger me-2"
            @click="removeVariable(i)"
          >
            {{ $t('label.remove') }}
          </button>
          <span class="btn">
            {{ v.label || v.name || $t('variables.defaultVariableLabel') }}
            <small class="text-secondary">({{ v.kind }})</small>
          </span>
        </div>

        <button
          class="btn btn-link px-1"
          @click="addVariable"
        >
          + {{ $t('label.add') }}
        </button>
      </div>
    </div>

    <hr />

    <div
      v-if="selected"
      class="card mb-3"
    >
      <div class="card-body">
        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('variables.edit.nameLabel') }}</label>
              <input
                v-model.trim="selected.name"
                class="form-control"
                :placeholder="$t('variables.edit.namePlaceholder')"
              />
              <small class="text-muted">{{ $t('variables.edit.nameFootnote') }} <code>${variables.{{ selected.name || '...' }}}</code></small>
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('variables.edit.labelLabel') }}</label>
              <input
                v-model="selected.label"
                class="form-control"
                :placeholder="$t('variables.edit.labelPlaceholder')"
              />
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('variables.edit.kindLabel') }}</label>
              <c-input-select
                v-model="selected.kind"
                :options="availableKinds"
                :clearable="false"
              />
            </div>
          </div>

          <div class="col-12 col-lg-6">
            <div class="mb-3 pt-4">
              <div class="form-check">
                <input
                  id="variableIsRequired"
                  v-model="selected.isRequired"
                  type="checkbox"
                  class="form-check-input"
                />
                <label class="form-check-label" for="variableIsRequired">{{ $t('label.required') }}</label>
              </div>
            </div>
          </div>
        </div>

        <template v-if="mockField && kindConfigurator">
          <hr />
          <component
            :is="kindConfigurator"
            :namespace="namespace"
            :module="mockModule"
            :field="mockField"
          />
        </template>

        <hr />

        <div class="mb-1">
          <label class="form-label text-primary mb-0">{{ $t('variables.edit.defaultValueLabel') }}</label>
          <FieldEditor
            v-if="defaultValueMock"
            value-only
            v-bind="defaultValueMock"
            @change="onDefaultValueChange"
          />
        </div>
      </div>
    </div>

    <hr />

    <h5 class="mb-3">{{ $t('variables.appearance.label') }}</h5>
    <div class="row">
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('variables.appearance.showLabel') }}</label>
          <c-input-checkbox v-model="options.showVariableLabels" switch :labels="checkboxLabel" />
        </div>
      </div>

      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('variables.appearance.horizontalFieldLayoutEnabled') }}</label>
          <c-input-checkbox v-model="options.horizontalFieldLayoutEnabled" switch :labels="checkboxLabel" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import lodash from 'lodash'
import { compose, validator } from 'corteza-lib/js/dist'
import { usePageBlockBase } from '../usePageBlockBase'
import FieldEditor from '../../ModuleFields/Editor'
import * as Configurators from '../../ModuleFields/Configurator/loader'

const { t: $t } = useI18n({ useScope: 'global' })

const { merge } = lodash

// File/Geometry don't stringify sensibly into a ${variables.x} QL expression
const excludedKinds = ['File', 'Geometry']

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, required: false, default: '' },
  editable: { type: Boolean, required: false, default: false },
  resizing: { type: Boolean, required: false, default: false },
  magnified: { type: Boolean, required: false, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, required: false, default: false },
  errors: { type: Object, required: false, default: () => ({}) },
})

const emit = defineEmits(['errors'])
const { options } = usePageBlockBase(props, emit)

if (!Array.isArray(options.value.variables)) {
  options.value.variables = []
}
if (options.value.showVariableLabels === undefined) {
  options.value.showVariableLabels = false
}
if (options.value.horizontalFieldLayoutEnabled === undefined) {
  options.value.horizontalFieldLayoutEnabled = false
}

const checkboxLabel = computed(() => ({ on: $t('label.yes'), off: $t('label.no') }))

const availableKinds = [...compose.ModuleFieldRegistry.keys()].filter(k => !excludedKinds.includes(k))

const variables = computed(() => options.value.variables)
const selectedIndex = ref(variables.value.length ? 0 : -1)
const selected = computed(() => variables.value[selectedIndex.value])

const mockField = ref(null)
const mockModule = ref(null)

function rebuildMockField () {
  if (!selected.value) {
    mockField.value = null
    mockModule.value = null
    return
  }

  const f = compose.ModuleFieldMaker({
    name: selected.value.name || 'variable',
    label: selected.value.label,
    kind: selected.value.kind,
    isRequired: selected.value.isRequired,
    isMulti: selected.value.isMulti,
  })
  // Merge stored options onto the kind's own defaults, then share the same
  // object reference back onto the stored variable so edits made by the
  // kind-specific Configurator (which mutates field.options.* in place)
  // persist directly, without extra syncing.
  f.options = merge({}, f.options, selected.value.options)
  selected.value.options = f.options

  mockField.value = f
  mockModule.value = new compose.Module({ fields: [f] }, props.namespace)
}

watch(() => selectedIndex.value, rebuildMockField, { immediate: true })
watch(() => selected.value?.kind, (kind, oldKind) => {
  if (oldKind !== undefined && kind !== oldKind) {
    // Option shapes differ per kind — start clean rather than carrying over
    // incompatible fields from the previous kind.
    selected.value.options = {}
  }
  rebuildMockField()
})

const kindConfigurator = computed(() => {
  if (!selected.value) return null
  const keys = Object.keys(Configurators)
  const i = keys.map(k => k.toLowerCase()).indexOf(selected.value.kind.toLowerCase())
  return i >= 0 ? Configurators[keys[i]] : null
})

const defaultValueMock = computed(() => {
  if (!mockField.value || !selected.value) return null
  const f = compose.ModuleFieldMaker(mockField.value)
  f.apply({ name: 'defaultValue' })
  const mod = new compose.Module({ fields: [f] }, props.namespace)
  const record = new compose.Record(mod, { defaultValue: selected.value.defaultValue })
  return { namespace: props.namespace, module: mod, field: f, record, errors: new validator.Validated() }
})

function onDefaultValueChange (value) {
  if (selected.value) selected.value.defaultValue = value
}

function addVariable () {
  options.value.variables = [
    ...variables.value,
    { name: '', label: '', kind: 'String', isRequired: false, isMulti: false, options: {}, defaultValue: '' },
  ]
  selectedIndex.value = variables.value.length - 1
}

function removeVariable (i) {
  const vars = [...variables.value]
  vars.splice(i, 1)
  options.value.variables = vars
  if (selectedIndex.value >= vars.length) selectedIndex.value = vars.length - 1
}
</script>
