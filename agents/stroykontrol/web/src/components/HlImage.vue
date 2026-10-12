<script setup>
// A page image with an optional highlighted region (page fractions).
import { computed } from 'vue'

const props = defineProps({ src: String, alt: String, zoomPct: { type: Number, default: 100 }, box: Array })

const wrapStyle = computed(() => props.zoomPct === 100
  ? { position: 'relative', display: 'inline-block', maxWidth: '100%' }
  : { position: 'relative', display: 'inline-block', width: props.zoomPct + '%' })

const boxStyle = computed(() => {
  if (!props.box) return null
  const [x0, y0, x1, y1] = props.box
  const pad = 0.004
  return {
    left: Math.max(0, x0 - pad) * 100 + '%',
    top: Math.max(0, y0 - pad) * 100 + '%',
    width: (Math.min(1, x1 + pad) - Math.max(0, x0 - pad)) * 100 + '%',
    height: (Math.min(1, y1 + pad) - Math.max(0, y0 - pad)) * 100 + '%',
  }
})
</script>

<template>
  <div :style="wrapStyle">
    <img :src="src" :alt="alt" class="hl-img" />
    <div v-if="boxStyle" class="hl-box" :style="boxStyle"></div>
  </div>
</template>
