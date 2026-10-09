<template>
  <c-rich-text-input
    v-model="label"
    :labels="{
      urlPlaceholder: t('steps.content.configurator.urlPlaceholder'),
      ok: t('steps.content.configurator.ok'),
    }"
    class="m-3"
    @input="emit('update-value', $event)"
  />
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
const { CRichTextInput } = components

const { t } = useI18n()

const props = defineProps({
  item: { type: Object, default: () => ({}) },
  edges: { type: Object, default: () => ({}) },
  outEdges: { type: Number, default: 0 },
  isSubworkflow: { type: Boolean, default: false },
})

const emit = defineEmits(['update-value', 'update-default-value'])

const label = computed({
  get () { return props.item.node.value },
  set (label) { props.item.node.value = label },
})
</script>