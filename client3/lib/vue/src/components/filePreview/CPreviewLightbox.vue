<template>
  <c-lightbox>
    <template
      v-for="(_, slot) of slots"
      #[slot]="scope"
    >
      <slot
        :name="slot"
        v-bind="scope"
      />
    </template>

    <component
      :is="previewType"
      v-if="previewType"
      v-bind="$attrs"
    />
  </c-lightbox>
</template>

<script setup lang="ts">
import { computed, useAttrs, useSlots, type Component } from 'vue'
import { LightboxIMG as IMG, PDF, DOCX, XLSX, PPTX, CAD, BIM, Hint, NoPreview } from './common/types'
import { CLightbox } from '../lightbox/index.ts'
import { getComponent } from './common/index.js'

const attrs = useAttrs()
const slots = useSlots()

const previewComponents: Record<string, Component> = {
  IMG, PDF, DOCX, XLSX, PPTX, CAD, BIM, Hint, NoPreview,
}

const previewType = computed(() => {
  const name = getComponent({
    type: attrs.mime as string,
    src: attrs.src as string,
    name: attrs.name as string,
    meta: attrs.meta as Record<string, any>,
  })
  return name ? previewComponents[name] : undefined
})
</script>
