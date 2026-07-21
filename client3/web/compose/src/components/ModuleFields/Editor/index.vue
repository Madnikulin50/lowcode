<template>
  <component
    :is="component"
    v-if="component"
    v-bind="{ ...$attrs, field }"
  />
  <code v-else>{{ t('field.unknownFieldKind', { kind: field.kind }) }}</code>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import * as Editors from './loader'

const props = defineProps({
  field: { type: Object, required: true },
})

const { t } = useI18n({ useScope: 'global', messages: {} })

const component = computed(() => {
  const kind = props.field.kind.toLocaleLowerCase()
  const keys = Object.keys(Editors)
  const i = keys.map(c => c.toLocaleLowerCase()).findIndex(c => c === kind)
  return i >= 0 ? Editors[keys[i]] : null
})
</script>
