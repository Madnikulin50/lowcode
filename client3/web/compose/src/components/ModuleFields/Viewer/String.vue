<template>
  <json-field-view
    v-if="showJSON"
    :value="value"
    :field="field"
    :empty-label="$t('kind.string.json.empty')"
    :invalid-label="$t('kind.string.json.invalid')"
  />
  <div
    v-else
    class="rt-content"
  >
    <p
      v-if="formatted"
      :style="{ 'white-space': field.options.useRichTextEditor ? 'pre-line' : undefined }"
      :class="[{ multiline: field.isMulti || field.options.multiLine }, ...classes]"
      v-html="formatted"
    />
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { computed } from 'vue'
import { useViewerBase } from './useViewerBase'
import { compose } from 'corteza-lib/js/dist'
import { isJSONDisplay } from 'corteza-webapp-compose/src/lib/json-field'
import JsonFieldView from '../Common/JsonFieldView.vue'

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  field: { type: compose.ModuleField, required: true },
  record: { type: Object, required: true },
  valueOnly: { type: Boolean, required: false },
  extraOptions: { type: Object, default: () => ({}) },
  includeStyles: { type: Boolean, default: false },
  disableClick: { type: Boolean, default: false },
})

const { value, formatted, classes } = useViewerBase(props)

const showJSON = computed(() => isJSONDisplay(props.field))
</script>
