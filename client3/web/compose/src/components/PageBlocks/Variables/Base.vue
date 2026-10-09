<template>
  <Wrap
    v-bind="$props"
    @refreshBlock="refresh"
  >
    <div
      v-if="!variables.length"
      class="d-flex h-100 align-items-center justify-content-center"
    >
      <p class="mb-0 my-3">
        {{ $t('noConfiguration') }}
      </p>
    </div>

    <div
      v-else
      class="d-flex flex-wrap gap-3 p-3"
    >
      <div
        v-for="field in mockModule.fields"
        :key="field.name"
        class="variable-field"
      >
        <FieldEditor
          :value-only="!options.showVariableLabels"
          :horizontal="options.horizontalFieldLayoutEnabled"
          :namespace="namespace"
          :field="field"
          :record="mockRecord"
          :errors="errors"
          @change="onFieldChange(field.name, $event)"
        />
      </div>
    </div>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { computed, onMounted } from 'vue'
import { compose, validator } from 'corteza-lib/js/dist'
import { usePageBlockBase } from '../usePageBlockBase'
import { useStore } from '../../../store'
import FieldEditor from '../../ModuleFields/Editor'
import Wrap from '../Wrap/index.js'

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
const store = useStore()
const { options } = usePageBlockBase(props, emit)

const errors = new validator.Validated()

const pageID = computed(() => props.page.pageID)
const variables = computed(() => options.value.variables || [])

// One mock Module shared by all variables in this block instance, and one
// mock Record holding all of their current values (seeded from the
// session-only pageVariables store, falling back to each variable's
// configured defaultValue).
const mockModule = computed(() => new compose.Module({
  fields: variables.value.map(v => ({
    name: v.name,
    kind: v.kind,
    label: v.label,
    isRequired: v.isRequired,
    isMulti: v.isMulti,
    options: v.options,
  })),
}, props.namespace))

const mockRecord = computed(() => {
  const values = {}
  for (const v of variables.value) {
    const stored = store.pageVariables.getValue(pageID.value, v.name)
    values[v.name] = stored !== undefined ? stored : v.defaultValue
  }
  return new compose.Record(mockModule.value, values)
})

function onFieldChange (name, value) {
  store.pageVariables.setValue(pageID.value, name, value)
  window.dispatchEvent(new CustomEvent('page-variable-change', {
    detail: { pageID: pageID.value, fieldName: name },
  }))
}

// Sibling blocks (RecordList/Chart/Metric) may have already run their first
// fetch before this block mounts and seeds its defaults into the store —
// their prefilter would have evaluated ${variables.x} as empty. Dispatching
// here lets them self-correct with a second fetch once real values exist.
function seedDefaults () {
  for (const v of variables.value) {
    const wasUnset = store.pageVariables.setDefaultIfUnset(pageID.value, v.name, v.defaultValue)
    if (wasUnset) {
      window.dispatchEvent(new CustomEvent('page-variable-change', {
        detail: { pageID: pageID.value, fieldName: v.name },
      }))
    }
  }
}

function refresh () {
  seedDefaults()
}

onMounted(() => {
  seedDefaults()
})
</script>

<style lang="scss" scoped>
.variable-field {
  min-width: 12rem;

  // With labels shown (and especially in horizontal mode, where the label
  // sits beside the input in its own column) a narrow chip clips the label
  // — give it more room, closer to how the Record block sizes its fields.
  &:has(.value-only) {
    min-width: 12rem;
  }

  &:not(:has(.value-only)) {
    min-width: 18rem;
    flex-grow: 1;
  }
}
</style>
