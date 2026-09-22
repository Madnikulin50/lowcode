<template>
  <div class="row g-3 ae-system-fields">
    <div
      v-if="id && id !== '0'"
      class="col-12 col-lg-6"
    >
      <label class="form-label">{{ $t('id') }}</label>
      <p class="ae-field-value mb-0">{{ id }}</p>
    </div>

    <div
      v-for="(f, i) in visibleFields"
      :key="i"
      class="col-12 col-lg-6"
    >
      <label
        class="form-label"
        :data-test-id="`input-${generateTestID(f)}`"
      >{{ $t(f) || $t(label) }}</label>
      <p class="ae-field-value mb-0">{{ getFieldValue(f) }}</p>
    </div>
    <slot name="custom-field" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { getSystemFields, kebabize } from 'corteza-webapp-admin/src/lib/sysFields'
import { fmt } from 'corteza-lib/js/dist'

const { t } = useI18n()

const props = defineProps({
  resource: { type: Object, required: true },
  label: { type: String, default: '' },
  id: { type: String, default: '' },
})

const systemFields = computed(() => getSystemFields(props.resource))
const visibleFields = computed(() => systemFields.value.filter(f => getFieldValue(f) !== '0'))

function generateTestID(field) {
  return kebabize(field)
}

function getFieldValue(field) {
  const isTimeValue = field.substring(field.length - 2) === 'At'
  const value = isTimeValue ? fmt.fullDateTime(props.resource[field]) : props.resource[field]
  return value
}
</script>
