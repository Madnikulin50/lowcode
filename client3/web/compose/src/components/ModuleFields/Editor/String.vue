<template>
  <div class="mb-3" :data-test-id="getFieldCypressId(label)" :class="formGroupStyleClasses">

      <div v-if="!valueOnly" class="d-flex align-items-center text-primary p-0">
        <span :title="label" class="d-flex">{{ label }}</span>
        <c-hint :tooltip="hint" />
        <slot name="tools" />
      </div>
      <div class="small text-muted" :class="{ 'mb-1': description }">{{ description }}</div>


    <multi v-if="field.isMulti" v-slot="ctx" v-model:value="value" :errors="errors">
      <CRichTextInput
        v-if="field.options.useRichTextEditor"
        :value="value[ctx.index]"
        :labels="richTextLabels"
        @input="setMultiValue($event, ctx.index)"
      />
      <textarea
        v-else-if="field.options.multiLine"
        :value="value[ctx.index]"
        class="form-control"
        @input="setMultiValue($event.target.value, ctx.index)"
      ></textarea>
      <input
        v-else
        :value="value[ctx.index]"
        type="text"
        class="form-control form-control-sm"
        @input="setMultiValue($event.target.value, ctx.index)"
      />
    </multi>

    <template v-else>
      <CRichTextInput
        v-if="field.options.useRichTextEditor"
        v-model="value"
        :labels="richTextLabels"
      />
      <textarea
        v-else-if="field.options.multiLine"
        v-model="value"
        class="form-control"
      ></textarea>
      <input
        v-else
        v-model="value"
        type="text"
        class="form-control form-control-sm"
      />
      <errors :errors="errors" />
    </template>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditorBase } from './base'
import { components } from 'corteza-lib/vue/dist'
import errors from '../errors'
import multi from './multi'
const { CRichTextInput } = components

const props = defineProps({
  namespace: { type: Object, required: true },
  field: { type: Object, required: true },
  record: { type: Object, required: true },
  errors: { type: Object, required: true },
  valueOnly: { type: Boolean, default: false },
  horizontal: { type: Boolean, default: false },
  extraOptions: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['change', 'update:preventPopoverClose'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { value, formGroupStyleClasses, label, hint, description, getFieldCypressId, setMultiValue } = useEditorBase(props, emit)

const richTextLabels = computed(() => ({
  urlPlaceholder: t('content.urlPlaceholder'),
  ok: t('content.ok'),
  emojiPicker: {
    search: t('content.emojiPicker.search'),
    searchResults: t('content.emojiPicker.searchResults'),
    frequentlyUsed: t('content.emojiPicker.frequentlyUsed'),
    noResults: t('content.emojiPicker.noResults'),
    quickReactions: t('content.emojiPicker.quickReactions'),
  },
}))
</script>
