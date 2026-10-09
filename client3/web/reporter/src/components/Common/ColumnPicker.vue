<template>
  <div class="d-flex">
    <c-item-picker
      :model-value="selected"
      :options="options"
      :labels="{
        searchPlaceholder: t('list.searchPlaceholder'),
        availableItems: t('available-columns'),
        selectAllItems: t('select-all'),
        selectedItems: t('selected-columns'),
        unselectAllItems: t('unselect-all'),
        noItemsFound: t('no-columns-found'),
      }"
      style="max-height: 40vh;"
      @update:model-value="onUpdate"
    >
      <template #default="{ field }">
        <b class="cursor-default text-dark">
          <template v-if="field.label">{{ field.label }} ({{ field.name }})</template>
          <template v-else>{{ field.name }}</template>
        </b>
      </template>
    </c-item-picker>
  </div>
</template>
<script setup>
defineOptions({ i18nOptions: { namespaces: 'builder' } })
import { computed } from 'vue'

import { components, useNsI18n } from 'corteza-lib/vue/dist'
const { CItemPicker } = components

const props = defineProps({
  allItems: { type: Array, required: true },
  selectedItems: { type: Array, required: true },
  enableSorting: { type: Boolean, default: false },
})
const emit = defineEmits(['update:selected-items'])

const t = useNsI18n()

const selected = computed({
  get: () => props.selectedItems.map(({ name }) => name),
  set: (val) => emit('update.selected-items', val),
})

const options = computed(() => {
  const fields = [...props.allItems]
  if (props.enableSorting) fields.sort((a, b) => a.label.localeCompare(b.label))
  return fields.map(field => ({
    value: field.name,
    text: [field.name, field.label, field.kind, field.system ? 'system' : ''].join(' '),
    field,
  }))
})

function onUpdate(val) {
  selected.value = val
}
</script>