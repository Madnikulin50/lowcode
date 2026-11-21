<template>
  <div class="inline h-100">
    <div class="h-100">
      <component
        :is="previewType"
        v-bind="$attrs"
        :max-pages="1"
        :initial-scale="1.5"
        inline
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, useAttrs, type Component } from 'vue'
import { InlineIMG as IMG, PDF, DOCX, XLSX, PPTX, CAD, BIM, Hint, NoPreview } from './common/types'
import { getComponent } from './common/index.js'

const attrs = useAttrs()

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
