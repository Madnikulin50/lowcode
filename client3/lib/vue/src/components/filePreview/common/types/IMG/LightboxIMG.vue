<template>
  <div class="popup-img-preview d-flex align-items-center justify-content-center h-100">
    <img
      :src="src"
      :alt="name"
      class="lightbox-img"
    >
  </div>
</template>

<script setup lang="ts">
import { computed, useAttrs } from 'vue'

defineOptions({ inheritAttrs: false })

const attrs = useAttrs()

const src = computed(() => (attrs as any).src)
const name = computed(() => (attrs as any).name || '')
</script>

<style lang="scss" scoped>
.popup-img-preview {
  width: 100%;
  height: 100%;
  background: var(--body-bg);
}

.lightbox-img {
  max-width: 90%;
  max-height: 90%;
  object-fit: contain;
}
</style>
