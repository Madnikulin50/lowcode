<template>
  <div class="row g-0 mb-2">
    <div class="col-3 align-self-center">
      <div
        v-if="allowOmitStrategy"
        class="form-check"
      >
        <input
          :id="`use-${field}`"
          v-model="use"
          class="form-check-input"
          type="checkbox"
          :disabled="disabled"
        >
        <label
          class="form-check-label"
          :for="`use-${field}`"
        >{{ label }}</label>
      </div>
      <div
        v-else
        class="fw-bold"
      >
        {{ label }}
      </div>
    </div>

    <div class="col-3">
      <c-input-select
        v-show="strategy !== 'omit'"
        v-model="strategy"
        :options="strategies"
        :disabled="!use"
        label="text"
        :reduce="strategy => strategy.value"
        :clearable="false"
        size="sm"
      />
    </div>

    <div
      v-if="strategy === ''"
      class="col-6"
    >
      <input
        :value="storeIdent"
        class="form-control form-control-sm"
        :placeholder="t('ident.placeholder')"
        readonly
      >
    </div>

    <div
      v-else-if="showIdentInput"
      class="col-6"
    >
      <input
        v-model="draft.ident"
        class="form-control form-control-sm"
        :placeholder="t('ident.placeholder')"
        :disabled="disableIdentInput"
      >
    </div>
  </div>
</template>

<script setup lang="js">
defineOptions({ i18nOptions: { namespaces: 'module', keyPrefix: 'edit.config.dal.encoding-strategy' } })
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { defaultConfigDraft, types } from './encoding-strategy'

const prefixed$ = 'edit.config.dal.encoding-strategy.'
const { t: $t } = useI18n()
const t = (key) => $t(prefixed$ + key)

const props = defineProps({
  config: {
    type: Object,
    required: true,
  },
  field: {
    type: String,
    required: true,
  },
  label: {
    type: String,
    required: true,
  },
  isMulti: {
    type: Boolean,
    default: false,
  },
  storeIdent: {
    type: String,
    required: true,
  },
  defaultStrategy: {
    type: String,
    default: types.Plain,
  },
  allowOmitStrategy: {
    type: Boolean,
    default: true,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['change'])

const draft = ref(defaultConfigDraft(props.config, props.storeIdent))
const undoOmit = ref(props.defaultStrategy)

const strategies = computed(() => {
  return [
    { value: types.Plain, text: t('strategies.plain.label'), disabled: props.isMulti },
    { value: types.Alias, text: t('strategies.alias.label'), disabled: props.isMulti },
    { value: types.JSON, text: t('strategies.json.label') },
  ].filter(({ disabled }) => !disabled)
})

const showIdentInput = computed(() => {
  return [types.JSON, types.Alias, types.Plain].includes(strategy.value)
})

const disableIdentInput = computed(() => {
  return [types.Plain].includes(strategy.value)
})

const strategy = computed({
  get () {
    for (const t of Object.values(types)) {
      if (props.config[t] === undefined) {
        continue
      }
      return t
    }
    return props.defaultStrategy
  },
  set (strategy) {
    emit('change', { strategy, config: draft.value })
  },
})

const use = computed({
  get () {
    return strategy.value !== types.Omit
  },
  set (use) {
    if (strategy.value !== types.Omit) {
      undoOmit.value = strategy.value
    }
    strategy.value = use ? undoOmit.value : types.Omit
  },
})

watch(draft, (config) => {
  emit('change', { strategy: strategy.value, config })
}, { deep: true })
</script>
