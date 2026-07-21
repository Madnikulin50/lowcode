<template>
  <c-input-select
    :value="value"
    :options="columns"
    :get-option-label="getColumnLabel"
    :get-option-key="getOptionKey"
    :placeholder="$t('label.none')"
    :reduce="r => r.name"
    append-to-body
    v-bind="$attrs"
  />
</template>

<script setup>
const props = defineProps({
  columns: {
    type: Array,
    required: true,
  },
  value: {
    type: String,
    default: '',
  },
})

function getColumnLabel ({ name, label }) {
  return `${label} (${name})`
}

function getOptionKey ({ name }) {
  return name
}
</script>

<style lang="scss">
.column-selector {
  min-width: 15vw !important;
  max-width: 150px !important;
}
</style>
