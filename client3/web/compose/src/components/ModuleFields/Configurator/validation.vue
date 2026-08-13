<template>
  <div v-if="loaded">
    <div class="mb-3">
      <label class="d-flex align-items-center text-primary">
        {{ t('sanitizers.label') }}
        <a :href="`${documentationURL}#value-sanitizers`" target="_blank" class="btn btn-link btn-sm p-0 ms-auto">{{ t('label.examples') }}</a>
      </label>
      <div class="form-text">{{ t('sanitizers.description') }}</div>
      <c-form-table-wrapper
        :labels="{ addButton: t('label.add') }"
        @add-item="field.expressions.sanitizers.push('')"
      >
        <FieldExpressions
          v-model="field.expressions.sanitizers"
          :placeholder="t('sanitizers.expression.placeholder')"
          @remove="onRemove('sanitizers', $event)"
        />
      </c-form-table-wrapper>
    </div>

    <hr />

    <div class="mb-3 mt-3">
      <label class="d-flex align-items-center text-primary">
        {{ t('validators.label') }}
        <a :href="`${documentationURL}#value-validators`" target="_blank" class="btn btn-link btn-sm p-0 ms-auto">{{ t('label.examples') }}</a>
      </label>
      <div class="form-text">{{ t('validators.description') }}</div>
      <c-form-table-wrapper
        :labels="{ addButton: t('label.add') }"
        @add-item="field.expressions.validators.push({ test: '', error: '' })"
      >
        <FieldExpressions
          v-model="field.expressions.validators"
          v-slot="{ value }"
          :no-prompt="isValidatorEmpty"
          @remove="onRemove('validators', $event)"
        >
          <input
            v-model="value.test"
            type="text"
            class="form-control form-control-sm"
            :placeholder="t('validators.expression.placeholder')"
          />
          <button
            class="btn btn-warning btn-sm"
            :title="t('validators.error.tooltip')"
          >!</button>
          <input
            v-model="value.error"
            type="text"
            class="form-control form-control-sm"
            :placeholder="t('validators.error.placeholder')"
          />
          <FieldTranslator
            v-if="field"
            :field="field"
            :module="module"
            :highlight-key="`expression.validator.${value.validatorID}.error`"
            :disabled="isNew(value)"
          />
        </FieldExpressions>
      </c-form-table-wrapper>

      <div class="form-check mt-3">
        <input
          id="disableDefaultValidators"
          v-model="field.expressions.disableDefaultValidators"
          type="checkbox"
          class="form-check-input"
          :true-value="true"
          :false-value="false"
          :disabled="!field.expressions.validators || field.expressions.validators.length === 0"
        />
        <label class="form-check-label" for="disableDefaultValidators">{{ t('validators.disableBuiltIn') }}</label>
      </div>
    </div>

    <hr />

    <div class="mb-3 mt-3">
      <label class="d-flex align-items-center text-primary">
        {{ t('constraints.description') }}
        <c-hint :tooltip="t('constraints.tooltip.performance')" icon-class="text-warning" />
      </label>
      <c-input-checkbox
        v-model="fieldConstraint.exists"
        switch
        :labels="{ on: t('label.yes'), off: t('label.no') }"
        @change="toggleFieldConstraint"
      />
      <div v-if="fieldConstraint.exists" class="row mt-4">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('constraints.valueModifiers') }}</label>
            <select v-model="constraint.modifier" class="form-select form-control form-select-sm">
              <option v-for="opt in modifierOptions" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
            </select>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('constraints.multiValues') }}</label>
            <select v-model="constraint.multiValue" class="form-select form-control form-select-sm" :disabled="!field.isMulti">
              <option v-for="opt in multiValueOptions" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
            </select>
          </div>
        </div>
        <div v-if="fieldConstraint.total" class="col-12">
          <i>{{ t('constraints.totalFieldConstraintCount', { total: fieldConstraint.total }) }}</i>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, getCurrentInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose, NoID } from 'corteza-lib/js/dist'
import FieldExpressions from 'corteza-webapp-compose/src/components/Common/Module/FieldExpressions'
import FieldTranslator from 'corteza-webapp-compose/src/components/Admin/Module/FieldTranslator'

const props = defineProps({
  field: { type: compose.ModuleField, required: true },
  module: { type: compose.Module, required: true },
})

const { t } = useI18n({ useScope: 'global', messages: {} })

const loaded = ref(false)
const fieldConstraint = ref({
  ruleIndex: null,
  total: 0,
  exists: false,
  index: null,
})
const rule = ref({})

const { proxy } = getCurrentInstance()

const documentationURL = computed(() => {
  const [year, month] = (typeof VERSION !== 'undefined' ? VERSION : '2024.0').split('.')
  return `https://docs.cortezaproject.org/corteza-docs/${year}.${month}/integrator-guide/compose-configuration/index.html`
})

const modifierOptions = computed(() => [
  { value: 'ignore-case', text: t('constraints.ignoreCase') },
  { value: 'fuzzy-match', text: t('constraints.fuzzyMatch') },
  { value: 'sounds-like', text: t('constraints.soundsLike') },
  { value: 'case-sensitive', text: t('constraints.caseSensitive') },
])

const multiValueOptions = computed(() => [
  { value: 'one-of', text: t('constraints.oneOf') },
  { value: 'equal', text: t('constraints.equal') },
])

const constraint = computed({
  get () {
    if (props.module.config.recordDeDup.rules[fieldConstraint.value.ruleIndex]) {
      return props.module.config.recordDeDup.rules[fieldConstraint.value.ruleIndex].constraints[fieldConstraint.value.index]
    }
    return {}
  },
  set (value) {
    if (props.module.config.recordDeDup.rules[fieldConstraint.value.ruleIndex]) {
      props.module.config.recordDeDup.rules[fieldConstraint.value.ruleIndex].constraints[fieldConstraint.value.index] = value
    }
  },
})

onMounted(() => {
  checkForFieldConstraint()

  if (!props.field.expressions.sanitizers) {
    props.field.expressions.sanitizers = []
  }
  if (!props.field.expressions.validators) {
    props.field.expressions.validators = []
  }
  if (!props.field.expressions.disableDefaultValidators) {
    props.field.expressions.disableDefaultValidators = false
  }
  if (!props.field.expressions.formatters) {
    props.field.expressions.formatters = []
  }
  if (!props.field.expressions.disableDefaultFormatters) {
    props.field.expressions.disableDefaultFormatters = false
  }

  loaded.value = true
})

function isNew (value) {
  return !(value.validatorID && value.validatorID !== NoID)
}

function isValidatorEmpty ({ error = '', test = '' } = {}) {
  return error.length === 0 && test.length === 0
}

function checkForFieldConstraint () {
  props.module.config.recordDeDup.rules.forEach((rule, x) => {
    const { constraints } = rule
    constraints.forEach((constraint, i) => {
      if (constraint.attribute === props.field.name) {
        if (constraints.length === 1) {
          fieldConstraint.value.exists = true
          fieldConstraint.value.index = i
          fieldConstraint.value.ruleIndex = x
        }
        fieldConstraint.value.total += 1
      }
    })
  })
}

function toggleFieldConstraint (value) {
  if (!value) {
    props.module.config.recordDeDup.rules.splice(fieldConstraint.value.ruleIndex, 1)
    fieldConstraint.value.ruleIndex = null
    fieldConstraint.value.index = null
  } else if (fieldConstraint.value.ruleIndex == null) {
    props.module.config.recordDeDup.rules.push({
      name: '',
      strict: true,
      constraints: [{
        attribute: props.field.name,
        modifier: 'case-sensitive',
        multiValue: 'equal',
        type: props.field.kind,
      }],
    })
    fieldConstraint.value.ruleIndex = props.module.config.recordDeDup.rules.length - 1
    fieldConstraint.value.index = props.module.config.recordDeDup.rules[fieldConstraint.value.ruleIndex].constraints.length - 1
  }
}

function onRemove (type, index) {
  props.field.expressions[type].splice(index, 1)
}
</script>
