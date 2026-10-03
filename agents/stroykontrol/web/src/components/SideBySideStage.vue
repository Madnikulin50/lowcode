<script setup>
import { computed } from 'vue'

const props = defineProps({ pdUrl: String, rdUrl: String, pdLabel: String, rdLabel: String, zoomPct: Number })

const imgStyle = computed(() => props.zoomPct === 100 ? {} : { maxWidth: 'none', width: props.zoomPct + '%' })
</script>

<template>
  <div class="stage side-by-side">
    <div class="col">
      <div class="label">{{ pdLabel }}</div>
      <img :src="pdUrl" alt="ПД" :style="imgStyle" />
    </div>
    <div class="col">
      <div class="label">{{ rdLabel }}</div>
      <img :src="rdUrl" alt="РД" :style="imgStyle" />
    </div>
  </div>
</template>
