<template>
  <div class="mb-3">
    <C-rich-text-input
      v-if="options"
      v-model="options.value"
      :labels="{ urlPlaceholder: t('label.urlPlaceholder'), ok: t('label.ok') }"
    />
  </div>
</template>
<script setup>
defineOptions({ i18nOptions: { namespaces: 'general' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
const { CRichTextInput } = components

const props = defineProps({
  displayElementOptions: { type: Object, default: () => ({}) },
  columns: { type: Array, required: true },
  datasource: { type: Object, default: undefined },
})
const emit = defineEmits(['update:displayElementOptions'])

const { t } = useI18n()

const options = computed({
  get: () => props.displayElementOptions,
  set: (val) => emit('update.displayElementOptions', val),
})
</script>