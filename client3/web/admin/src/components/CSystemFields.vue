<template>
  <div class="row py-4">
    <div v-if="id && id !== '0'" class="col-12">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('id') }}</label>
        <p class="form-control-plaintext">{{ id }}</p>
      </div>
    </div>

    <div v-for="(f, i) in systemFields" :key="i" class="col-12">
      <div v-if="getFieldValue(f) !== '0'" class="mb-3">
        <label class="form-label text-primary" :data-test-id="`input-${generateTestID(f)}`">{{ $t(f) || $t(label) }}</label>
        <p class="form-control-plaintext">{{ getFieldValue(f) }}</p>
      </div>
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

function generateTestID(field) {
  return kebabize(field)
}

function getFieldValue(field) {
  const isTimeValue = field.substring(field.length - 2) === 'At'
  const value = isTimeValue ? fmt.fullDateTime(props.resource[field]) : props.resource[field]
  return value
}
</script>