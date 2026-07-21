<template>
  <div class="popup-img-preview">
    <div
      v-if="isSvg"
      class="svg-preview"
    >
      <img
        :src="src"
        class="svg-lightbox-img"
      >
    </div>
    <photo-swipe
      v-else
      :is-open="true"
      :items="items"
      :options="options"
      @close="() => $emit('close')"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeUnmount, useAttrs } from 'vue'
import { PhotoSwipe } from 'v-photoswipe'

defineOptions({ inheritAttrs: false })

const attrs = useAttrs()

defineEmits<{
  (e: 'close'): void
}>()

const options = ref({
  index: 0,
  bgOpacity: 0,
  closeOnScroll: false,
  escKey: false,
  history: false,
  arrowKeys: false,
  modal: false,
  closeEl: false,
  captionEl: false,
  fullscreenEl: false,
  zoomEl: false,
  shareEl: false,
  counterEl: false,
  arrowEl: false,
  preloaderEl: false,
  clickToCloseNonZoomable: false,
})

const src = computed(() => (attrs as any).src)
const mime = computed(() => (attrs as any).mime)
const name = computed(() => (attrs as any).name || '')
const meta = computed(() => (attrs as any).meta || {})

const isSvg = computed(() =>
  mime.value === 'image/svg+xml' ||
  (name.value && name.value.toLowerCase().endsWith('.svg'))
)

const items = computed(() => {
  const { original, preview } = meta.value
  const image = (original || preview || {}).image
  if (!image) {
    emit('close')
    return []
  }

  return [{
    src: src.value,
    w: image.width,
    h: image.height,
  }]
})

const emit = defineEmits<{
  (e: 'close'): void
}>()

function setDefaultValues() {
  options.value = {} as any
}

onBeforeUnmount(() => {
  setDefaultValues()
})
</script>

<style lang="scss">
.popup-img-preview {
  .pswp {
    pointer-events: none;
    .pswp__img {
      pointer-events: all;
    }
  }
  .pswp__top-bar {
    display: none!important;
  }

  .svg-preview {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;

    .svg-lightbox-img {
      max-width: 90%;
      max-height: 90%;
      object-fit: contain;
    }
  }
}
</style>
