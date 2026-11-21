<template>
  <div>
    <h5 class="d-flex align-items-center mb-3">
      {{ t('duplicationDetection') }}
      <c-hint
        :tooltip="t('tooltip.performance')"
        icon-class="text-warning"
      />
    </h5>

    <div
      v-for="(rule, index) in rules"
      :key="index"
    >
      <label class="d-flex align-items-center text-primary">
        {{ t('uniqueValueConstraint', { index: index + 1 }) }}
        <c-input-confirm
          show-icon
          size="md"
          button-class="px-2"
          class="ms-2"
          @confirmed="rules.splice(index, 1)"
        />
      </label>

      <div class="d-flex align-items-center justify-content-between flex-wrap w-100">
        <div class="mb-3">
          <c-input-select
            v-model="rule.currentField"
            :placeholder="t('searchFields')"
            :get-option-label="getOptionLabel"
            :get-option-key="getOptionKey"
            :options="filterFieldOptions(rule)"
            :reduce="o => o.name"
            style="min-width: 300px;"
            @input="updateRuleConstraint(rule)"
          />
        </div>

        <div class="mb-3">
          <label class="form-label text-primary ms-auto">{{ t('preventRecordsSave') }}</label>
          <c-input-checkbox
            v-model="rule.strict"
            switch
            :labels="checkboxLabel"
          />
        </div>
      </div>

      <c-form-table-wrapper
        v-if="rule.constraints && rule.constraints.length > 0"
        hide-add-button
      >
        <table class="table table-sm table-borderless mb-0">
          <thead>
            <tr class="text-primary">
              <th scope="col">
                {{ t("field") }}
              </th>
              <th scope="col">
                {{ t("type") }}
              </th>
              <th
                scope="col"
                style="width: 250px;"
              >
                {{ t("valueModifiers") }}
              </th>
              <th
                scope="col"
                style="width: 250px;"
              >
                {{ t("multiValues") }}
              </th>
              <th
                scope="col"
                style="width: 5rem;"
              />
            </tr>
          </thead>

          <tbody v-if="rule.constraints">
            <tr
              v-for="(constraint, consIndex) in rule.constraints"
              :key="`constraint-${consIndex}`"
            >
              <td>{{ getOptionLabel(getField(constraint.attribute)) }}</td>

              <td>{{ getField(constraint.attribute).kind }}</td>

              <td>
                <select
                  v-model="constraint.modifier"
                  class="form-select form-control form-select-sm"
                >
                  <option
                    v-for="opt in modifierOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >{{ opt.text }}</option>
                </select>
              </td>

              <td>
                <select
                  v-model="constraint.multiValue"
                  class="form-select form-control form-select-sm"
                  :disabled="!getField(constraint.attribute).isMulti"
                >
                  <option
                    v-for="opt in multiValueOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >{{ opt.text }}</option>
                </select>
              </td>

              <td class="text-end align-middle">
                <c-input-confirm
                  show-icon
                  @confirmed="rule.constraints.splice(consIndex, 1)"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </c-form-table-wrapper>

      <hr>
    </div>

    <div class="d-flex">
      <button
        class="btn btn-primary d-flex align-items-center"
        @click="addNewConstraint"
      >
        <font-awesome-icon
          :icon="['fas', 'plus']"
          class="me-2"
        />
        {{ t("addNewConstraint") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="js">
defineOptions({ i18nOptions: { namespaces: 'module', keyPrefix: 'edit.config.uniqueValues' } })
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'

const prefixed$ = 'edit.config.uniqueValues.'
const { t: $t } = useI18n()
const t = (key) => $t(prefixed$ + key)

const props = defineProps({
  module: {
    type: compose.Module,
    required: true,
  },
})

const checkboxLabel = ref({
  on: t('label.yes'),
  off: t('label.no'),
})

const rules = computed({
  get () {
    return props.module.config.recordDeDup.rules
  },
  set (value) {
    props.module.config.recordDeDup.rules = value
  },
})

const modifierOptions = computed(() => {
  const ruleModifiers = rules.value.reduce((acc, { constraints }) => {
    if (!constraints) return acc
    constraints.forEach(({ modifier }) => {
      if (!acc.includes(modifier)) acc.push(modifier)
    })
    return acc
  }, [])

  return [
    { value: 'ignore-case', text: t('ignoreCase') },
    { value: 'fuzzy-match', text: t('fuzzyMatch'), legacy: true },
    { value: 'sounds-like', text: t('soundsLike'), legacy: true },
    { value: 'case-sensitive', text: t('caseSensitive') },
  ].filter(({ value, legacy }) => !legacy || ruleModifiers.includes(value))
})

const multiValueOptions = computed(() => {
  return [
    { value: 'one-of', text: t('oneOf') },
    { value: 'equal', text: t('equal') },
  ]
})

function addNewConstraint () {
  rules.value.push({
    name: '',
    strict: true,
    constraints: [],
  })
}

function updateRuleConstraint (rule) {
  rule.currentField = props.module.fields.find(({ name }) => name === rule.currentField)

  if (!rule.constraints) {
    rule.constraints = []
  }

  rule.constraints.push({
    attribute: rule.currentField.name,
    modifier: 'case-sensitive',
    multiValue: 'equal',
    type: rule.currentField.kind,
    isMulti: rule.currentField.isMulti,
  })

  rule.currentField = undefined
}

function filterFieldOptions (rule) {
  const selectedFields = rule.constraints ? rule.constraints.map(({ attribute }) => attribute) : []
  return props.module.fields.filter(({ name }) => !selectedFields.includes(name))
}

function getField (attribute) {
  const field = props.module.fields.find(({ name }) => name === attribute)
  return field || {}
}

function getOptionLabel ({ kind, label, name }) {
  return label || name || kind
}

function getOptionKey ({ fieldID }) {
  return fieldID
}
</script>

<style lang="scss" scoped>
.list-background {
  background-color: var(--body-bg);
}
</style>
