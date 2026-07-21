<template>
  <c-input-select
    :model-value="value"
    :options="columns"
    :get-option-label="getColumnLabel"
    :get-option-key="getOptionKey"
    :placeholder="t('label.none')"
    :reduce="r => r.name"
    append-to-body
    @update:model-value="$emit('input', $event)"
  />
</template>
<script setup>
import { useI18n } from 'vue-i18n'

const props = defineProps({
  columns: { type: Array, required: true },
  value: { type: String, default: '' },
})
defineEmits(['input'])

const { t } = useI18n()

function getColumnLabel({ name, label }) { return `${label} (${name})` }
function getOptionKey({ name }) { return name }
</script>
<style lang="scss">
.column-selector { min-width: 15vw !important; max-width: 150px !important; }
</style>