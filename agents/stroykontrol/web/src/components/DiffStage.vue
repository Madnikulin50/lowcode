<script setup>
import { computed } from 'vue'
import { useFitStage } from '../composables/useFit.js'
import FitControls from './FitControls.vue'

const props = defineProps({ pdUrl: String, rdUrl: String, zoomPct: Number, viewer: Object })

const imgStyle = computed(() => props.zoomPct === 100 ? {} : { maxWidth: 'none', width: props.zoomPct + '%' })

const { bottomRef, topRef, onAlign } = useFitStage(props)
</script>

<template>
  <div>
    <div class="diff-toggle-row">
      <span>Режим:</span>
      <div class="switch">
        <button :class="{ active: !viewer.diffColorized }" @click="viewer.diffColorized = false">Серый</button>
        <button :class="{ active: viewer.diffColorized }" @click="viewer.diffColorized = true">Красный/зелёный</button>
      </div>
    </div>

    <div v-if="viewer.diffColorized" class="diff-legend">
      <span><span class="swatch pd"></span>только в ПД</span>
      <span><span class="swatch rd"></span>только в РД</span>
      <span><span class="swatch both"></span>совпадает</span>
    </div>

    <FitControls :viewer="viewer" @align="onAlign" />

    <div class="stage diff-stage" :class="{ colorized: viewer.diffColorized }">
      <img ref="bottomRef" class="bottom" :src="pdUrl" alt="ПД" :style="imgStyle" />
      <img ref="topRef" class="top" :src="rdUrl" alt="РД" :style="imgStyle" />
    </div>
  </div>
</template>
