<template>
  <div>
    <div class="mb-3">
      <label class="form-label text-primary">{{ t('kind.string.displayType.label') }}</label>
      <div class="btn-group" data-bs-toggle="buttons">
        <label
          v-for="opt in displayOptions"
          :key="opt.value"
          class="btn btn-outline-primary btn-sm"
          :class="{ active: displayType === opt.value }"
        >
          <input
            type="radio"
            class="btn-check"
            :value="opt.value"
            :checked="displayType === opt.value"
            autocomplete="off"
            @change="onDisplayType(opt.value)"
          />
          {{ opt.text }}
        </label>
      </div>
      <div class="form-text">{{ t('kind.string.displayType.description') }}</div>
    </div>

    <template v-if="displayType === 'json'">
      <div class="mb-3">
        <label class="form-label text-primary">{{ t('kind.string.json.layout.label') }}</label>
        <div class="btn-group" data-bs-toggle="buttons">
          <label
            v-for="opt in layoutOptions"
            :key="opt.value"
            class="btn btn-outline-primary btn-sm"
            :class="{ active: (f.options.jsonLayout || 'chips') === opt.value }"
          >
            <input
              type="radio"
              class="btn-check"
              :value="opt.value"
              :checked="(f.options.jsonLayout || 'chips') === opt.value"
              autocomplete="off"
              @change="f.options.jsonLayout = opt.value"
            />
            {{ opt.text }}
          </label>
        </div>
      </div>

      <div
        v-if="(f.options.jsonLayout || 'chips') !== 'pretty'"
        class="mb-3"
      >
        <label class="form-label text-primary">{{ t('kind.string.json.template.label') }}</label>
        <input
          v-model="f.options.jsonTemplate"
          type="text"
          class="form-control form-control-sm"
          :placeholder="t('kind.string.json.template.placeholder')"
        >
        <div class="form-text">{{ t('kind.string.json.template.description') }}</div>
      </div>

      <div
        v-if="(f.options.jsonLayout || 'chips') === 'table'"
        class="mb-3"
      >
        <label class="form-label text-primary">{{ t('kind.string.json.fields.label') }}</label>
        <input
          v-model="f.options.jsonFields"
          type="text"
          class="form-control form-control-sm"
          :placeholder="t('kind.string.json.fields.placeholder')"
        >
        <div class="form-text">{{ t('kind.string.json.fields.description') }}</div>
      </div>

      <div
        v-if="(f.options.jsonLayout || 'chips') === 'chips'"
        class="mb-3"
      >
        <label class="form-label text-primary">{{ t('kind.string.json.variantField.label') }}</label>
        <input
          v-model="f.options.jsonVariantField"
          type="text"
          class="form-control form-control-sm"
          :placeholder="t('kind.string.json.variantField.placeholder')"
        >
        <div class="form-text">{{ t('kind.string.json.variantField.description') }}</div>
      </div>

      <div
        v-if="(f.options.jsonLayout || 'chips') === 'chips'"
        class="mb-3"
      >
        <label class="form-label text-primary">{{ t('kind.string.json.variants.label') }}</label>
        <table
          v-if="variants.length"
          class="table table-sm align-middle"
        >
          <thead>
            <tr>
              <th>{{ t('kind.string.json.variants.match') }}</th>
              <th>{{ t('kind.string.json.variants.color') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(rule, index) in variants"
              :key="index"
            >
              <td>
                <input
                  v-model="rule.value"
                  type="text"
                  class="form-control form-control-sm"
                >
              </td>
              <td style="width: 9rem;">
                <select
                  v-model="rule.variant"
                  class="form-select form-select-sm"
                >
                  <option
                    v-for="name in variantNames"
                    :key="name"
                    :value="name"
                  >{{ name }}</option>
                </select>
              </td>
              <td class="text-end" style="width: 2rem;">
                <button
                  type="button"
                  class="btn btn-link btn-sm text-danger p-0"
                  @click="removeVariant(index)"
                >
                  <font-awesome-icon :icon="['fas', 'times']" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <button
          type="button"
          class="btn btn-outline-primary btn-sm"
          @click="addVariant"
        >
          {{ t('kind.string.json.variants.add') }}
        </button>
      </div>
    </template>

    <template v-else>
      <div class="form-check mb-2">
        <input
          id="multiLine"
          v-model="f.options.multiLine"
          type="checkbox"
          class="form-check-input"
        />
        <label class="form-check-label" for="multiLine">{{ t('kind.string.multiLine') }}</label>
      </div>
      <div class="form-check mb-2">
        <input
          id="useRichTextEditor"
          v-model="f.options.useRichTextEditor"
          type="checkbox"
          class="form-check-input"
        />
        <label class="form-check-label" for="useRichTextEditor">{{ t('kind.string.richText') }}</label>
      </div>
    </template>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConfiguratorBase } from './base'
import { JSON_VARIANTS } from 'corteza-webapp-compose/src/lib/json-field'

const props = defineProps({
  namespace: { type: Object, required: true },
  module: { type: Object, required: true },
  field: { type: Object, required: true },
  hasRecords: { type: Boolean, default: false },
})

const emit = defineEmits(['update:field'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { f } = useConfiguratorBase(props, emit)

const variantNames = JSON_VARIANTS

const displayType = computed(() => {
  const v = f.value.options.displayType
  if (v === 'ports') return 'json'
  return v || 'text'
})

const displayOptions = computed(() => [
  { text: t('kind.string.displayType.text'), value: 'text' },
  { text: t('kind.string.displayType.json'), value: 'json' },
])

const layoutOptions = computed(() => [
  { text: t('kind.string.json.layout.chips'), value: 'chips' },
  { text: t('kind.string.json.layout.table'), value: 'table' },
  { text: t('kind.string.json.layout.kv'), value: 'kv' },
  { text: t('kind.string.json.layout.pretty'), value: 'pretty' },
])

const variants = computed(() => {
  if (!Array.isArray(f.value.options.jsonVariants)) f.value.options.jsonVariants = []
  return f.value.options.jsonVariants
})

function onDisplayType (value) {
  f.value.options.displayType = value
  if (value === 'json') {
    if (!f.value.options.jsonLayout) f.value.options.jsonLayout = 'chips'
    if (!Array.isArray(f.value.options.jsonVariants)) f.value.options.jsonVariants = []
  }
}

function addVariant () {
  variants.value.push({ value: '', variant: 'secondary' })
}

function removeVariant (index) {
  variants.value.splice(index, 1)
}
</script>
