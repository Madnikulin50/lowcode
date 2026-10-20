<script setup>
import { names, nm } from '../constants.js'
import { usePageErrors, pageProblems } from '../composables/usePageImage.js'
import PageMissing from './PageMissing.vue'
import { computed } from 'vue'
import { useFitStage } from '../composables/useFit.js'
import FitControls from './FitControls.vue'

const props = defineProps({ pdUrl: String, rdUrl: String, zoomPct: Number, viewer: Object })

const imgStyle = computed(() => props.zoomPct === 100 ? {} : { maxWidth: 'none', width: props.zoomPct + '%' })

const { bottomRef, topRef, onAlign } = useFitStage(props)
const { st, fail, retry } = usePageErrors(props)
const problem = computed(() => pageProblems(props, st))
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
      <span><span class="swatch pd"></span>{{ nm('только в ПД') }}</span>
      <span><span class="swatch rd"></span>{{ nm('только в РД') }}</span>
      <span><span class="swatch both"></span>совпадает</span>
    </div>

    <FitControls :viewer="viewer" @align="onAlign" />

    <div v-if="problem" class="stage">
      <PageMissing v-for="p in problem" :key="p.label" :label="p.label" :reason="p.reason" retry @retry="retry" />
    </div>
    <div v-else :key="st.tick" class="stage diff-stage" :class="{ colorized: viewer.diffColorized }">
      <img ref="bottomRef" class="bottom" :src="pdUrl" :alt="names.ref" :style="imgStyle" @error="fail('pd', pdUrl)" />
      <img ref="topRef" class="top" :src="rdUrl" :alt="names.target" :style="imgStyle" @error="fail('rd', rdUrl)" />
    </div>
  </div>
</template>
