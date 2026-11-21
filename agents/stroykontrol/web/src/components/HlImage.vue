<script setup>
// A page image with an optional highlighted region (page fractions).
import { computed, ref, watch } from 'vue'
import { probeImage } from '../composables/usePageImage.js'
import PageMissing from './PageMissing.vue'

const props = defineProps({ src: String, alt: String, zoomPct: { type: Number, default: 100 }, box: Array })

const reason = ref('')
const attempt = ref(0)
watch(() => props.src, () => { reason.value = ''; attempt.value = 0 })
const shownSrc = computed(() => attempt.value ? props.src + (props.src.includes('?') ? '&' : '?') + 'retry=' + attempt.value : props.src)
async function onError () { reason.value = await probeImage(props.src) }

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
  <PageMissing v-if="reason" :label="alt" :reason="reason" retry @retry="reason = ''; attempt++" />
  <div v-else :style="wrapStyle">
    <img :src="shownSrc" :alt="alt" class="hl-img" @error="onError" />
    <div v-if="boxStyle" class="hl-box" :style="boxStyle"></div>
  </div>
</template>
