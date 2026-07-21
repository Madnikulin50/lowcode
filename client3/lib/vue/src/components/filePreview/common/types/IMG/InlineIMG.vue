<template>
  <div class="d-flex inline h-100 mb-1">
    <img
      ref="image"
      :key="src"
      :src="src"
      :title="title"
      :alt="alt"
      :class="getClass"
      :style="previewStyle"
      :width="getWidth"
      :height="getHeight"
      @click.stop="$emit('openPreview', {})"
      @error.once="reloadBrokenImage"
      @load="loaded=true"
    >
  </div>
</template>

<script setup lang="ts">
import { ref, computed, useAttrs } from 'vue'

defineOptions({ inheritAttrs: false })

const attrs = useAttrs()

const props = defineProps<{
  inline?: boolean
  alt?: string | null
  title?: string | null
}>()

defineEmits<{
  (e: 'openPreview', payload: Record<string, never>): void
}>()

const image = ref<HTMLImageElement | null>(null)
const loaded = ref(false)

const src = computed(() => (attrs as any).src)
const mime = computed(() => (attrs as any).mime)
const name = computed(() => (attrs as any).name || '')
const meta = computed(() => (attrs as any).meta || {})
const previewStyle = computed(() => (attrs as any).previewStyle || {})
const previewClass = computed(() => (attrs as any).previewClass || [])

const isSvg = computed(() =>
  mime.value === 'image/svg+xml' ||
  (name.value && name.value.toLowerCase().endsWith('.svg'))
)

const getClass = computed(() => {
  const rtr = [...previewClass.value]
  if ((attrs as any).onClick) {
    rtr.push('clickable')
  }
  if (loaded.value) {
    rtr.push('loaded')
  }
  return rtr
})

const getWidth = computed(() => {
  if (isSvg.value) return undefined
  return meta.value.preview?.image?.width
})

const getHeight = computed(() => {
  if (isSvg.value) return undefined
  return meta.value.preview?.image?.height
})

function reloadBrokenImage(ev: Event) {
  const target = ev.target as HTMLImageElement
  if (target && target.src) {
    window.setTimeout(() => {
      if (!target || !target.src) return
      target.src = target.src
    }, 500)
  }
}
</script>

<style scoped lang="scss">
div {
  object-fit: contain;

  img {
    &.loaded {
      width: auto;
      height: auto;
      display: block;
    }
  }

  &.inline {
    img {
      cursor: zoom-in;
    }

    img.disable-zoom-cursor {
      cursor: default;
    }
  }
}
</style>
