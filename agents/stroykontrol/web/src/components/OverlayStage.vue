<script setup>
import { ref, computed } from 'vue'
import { useFitStage } from '../composables/useFit.js'
import FitControls from './FitControls.vue'

const props = defineProps({ pdUrl: String, rdUrl: String, zoomPct: Number, viewer: Object })

const opacity = ref(50)
const imgStyle = computed(() => props.zoomPct === 100 ? {} : { maxWidth: 'none', width: props.zoomPct + '%' })

const { bottomRef, topRef, onAlign } = useFitStage(props)
</script>

<template>
  <div>
    <div class="opacity-row">
      <span>Прозрачность РД поверх ПД:</span>
      <input type="range" min="0" max="100" v-model.number="opacity" />
    </div>
    <FitControls :viewer="viewer" @align="onAlign" />
    <div class="stage overlay-stage">
      <img ref="bottomRef" class="bottom" :src="pdUrl" alt="ПД" :style="imgStyle" />
      <img ref="topRef" class="top" :src="rdUrl" alt="РД" :style="{ ...imgStyle, opacity: opacity / 100 }" />
    </div>
  </div>
</template>
