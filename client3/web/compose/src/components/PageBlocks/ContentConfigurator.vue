<template>
  <div class="tab-pane">
    <div class="mb-3">
      <c-rich-text-input
        v-model="options.body"
        :labels="{
          urlPlaceholder: $t('content.urlPlaceholder'),
          ok: $t('content.ok'),
          emojiPicker: {
            search: $t('content.emojiPicker.search'),
            searchResults: $t('content.emojiPicker.searchResults'),
            frequentlyUsed: $t('content.emojiPicker.frequentlyUsed'),
            noResults: $t('content.emojiPicker.noResults'),
            quickReactions: $t('content.emojiPicker.quickReactions'),
          },
        }"
      />

      <i18next
        path="interpolationFootnote"
        tag="small"
        class="text-muted"
      >
        <code>${record.values.fieldName}</code>
        <code>${recordID}</code>
        <code>${ownerID}</code>
        <span><code>${userID}</code>, <code>${user.name}</code></span>
      </i18next>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { usePageBlockBase } from './usePageBlockBase'
import { components } from 'corteza-lib/vue/dist'
const { CRichTextInput } = components

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, required: false, default: '' },
  editable: { type: Boolean, required: false, default: false },
  resizing: { type: Boolean, required: false, default: false },
  magnified: { type: Boolean, required: false, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, required: false, default: false },
  errors: { type: Object, required: false, default: () => ({}) },
})

const emit = defineEmits(['errors'])

const { options } = usePageBlockBase(props, emit)
</script>
