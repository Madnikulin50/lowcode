<template>
  <tr v-if="value">
    <td class="handle align-middle pe-2">
      <font-awesome-icon
        :icon="['fas', 'bars']"
        class="text-secondary grab"
      />
    </td>

    <td style="min-width: 250px;">
      <input
        v-model="value.name"
        class="form-control form-control-sm"
        :title="nameState === false ? $t('module.edit.tooltip.name') : ''"
        required
        :readonly="hasData"
        :class="{ 'is-invalid': nameState === false }"
        type="text"
      >
    </td>

    <td style="min-width: 250px;">
      <div class="input-group input-group-sm">
        <input
          v-model="value.label"
          class="form-control"
          type="text"
        >

        <field-translator
          :field="value"
          :module="module"
          :disabled="isNew"
          button-variant="extra-light"
          highlight-key="label"
        />
      </div>
    </td>

    <td style="min-width: 250px;">
      <div class="input-group input-group-sm field-type">
        <c-input-select
          v-model="value.kind"
          :title="hasData ? $t('field.not-configurable') : ''"
          :options="fieldKinds"
          :reduce="kind => kind.kind"
          :disabled="hasData"
          :clearable="false"
          @input="$emit('updateKind')"
        />

        <button
          data-test-id="button-configure-field"
          class="btn btn-extra-light"
          :disabled="isEditDisabled"
          @click.prevent="$emit('edit')"
        >
          <font-awesome-icon
            :icon="['fas', 'wrench']"
          />
        </button>
      </div>
    </td>

    <td />
    <td />

    <td class="align-middle text-center">
      <div class="form-check form-switch">
        <input
          :id="`isRequired-${value.fieldID}`"
          v-model="value.isRequired"
          class="form-check-input"
          style="min-width: auto;"
          type="checkbox"
          :disabled="!value.cap.required"
        >
      </div>
    </td>

    <td class="align-middle text-center">
      <div class="form-check form-switch ms-2" >
        <input
          :id="`isMulti-${value.fieldID}`"
          v-model="value.isMulti"
          class="form-check-input"
          style="min-width: auto;"
          type="checkbox"
          :disabled="!value.cap.multi"
        >
      </div>
    </td>

    <td
      class="text-end align-middle"
      style="min-width: 7rem;"
    >
      <c-permissions-button
        v-if="canGrant && !isNew"
        button-variant="outline-extra-light"
        size="sm"
        :title="value.label || value.name || value.fieldID"
        :target="value.label || value.name || value.fieldID"
        :tooltip="$t('permissions.resources.compose.module-field.tooltip')"
        :resource="`corteza::compose:module-field/${module.namespaceID}/${module.moduleID}/${value.fieldID}`"
        class="text-dark border-0 me-2"
      />

      <c-input-confirm
        show-icon
        @confirmed="$emit('delete')"
      />
    </td>
  </tr>
</template>

<script setup lang="js">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import FieldTranslator from 'corteza-webapp-compose/src/components/Admin/Module/FieldTranslator'
import { compose, NoID } from 'corteza-lib/js/dist'

const { t } = useI18n()

defineOptions({
  i18nOptions: {
    namespaces: 'general',
  },
})

const props = defineProps({
  value: {
    type: Object,
    default: null,
  },
  modelValue: {
    type: Object,
    required: true,
  },
  module: {
    type: compose.Module,
    required: true,
  },
  canGrant: {
    type: Boolean,
    required: false,
  },
  hasRecords: {
    type: Boolean,
    required: true,
  },
  isDuplicate: {
    type: Boolean,
    required: false,
  },
})

const emit = defineEmits(['updateKind', 'edit', 'delete'])

const value = computed(() => props.modelValue ?? props.value)

const nameState = computed(() => {
  if (hasData.value) {
    return null
  }
  if (props.isDuplicate) {
    return false
  }
  return props.modelValue.isValid ? null : false
})

const hasData = computed(() => !isNew.value && props.hasRecords)

const isNew = computed(() => props.module?.moduleID === NoID || props.modelValue?.fieldID === NoID)

const fieldKinds = computed(() => {
  return [...compose.ModuleFieldRegistry.keys()]
    .map(kind => {
      return { kind, label: t('fieldKinds.' + kind + '.label') }
    }).sort((a, b) => a.label.localeCompare(b.label))
})

const isEditDisabled = computed(() => !props.modelValue.cap.configurable || nameState.value !== null)
</script>

<style lang="scss" scoped>
td {
  input,
  .input-group {
    min-width: 150px;
  }

  .handle {
    width: 30px;
  }
}
</style>
