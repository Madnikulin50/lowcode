<script setup>
import HlImage from './HlImage.vue'

defineProps({ pdUrl: String, rdUrl: String, pdLabel: String, rdLabel: String, zoomPct: Number, hlPd: Array, hlRd: Array })
</script>

<template>
  <div class="stage side-by-side">
    <div class="col">
      <div class="label">{{ pdLabel }}</div>
      <HlImage v-if="pdUrl" :src="pdUrl" alt="ПД" :zoom-pct="zoomPct" :box="hlPd" />
      <div v-else class="empty">Нет страницы</div>
    </div>
    <div class="col">
      <div class="label">{{ rdLabel }}</div>
      <HlImage v-if="rdUrl" :src="rdUrl" alt="РД" :zoom-pct="zoomPct" :box="hlRd" />
      <div v-else class="empty">Нет страницы</div>
    </div>
  </div>
</template>
