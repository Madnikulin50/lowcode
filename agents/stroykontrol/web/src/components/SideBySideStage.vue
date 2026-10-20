<script setup>
import { names, nm } from '../constants.js'
import HlImage from './HlImage.vue'
import PageMissing from './PageMissing.vue'

defineProps({ pdUrl: String, rdUrl: String, pdLabel: String, rdLabel: String, zoomPct: Number, hlPd: Array, hlRd: Array })
</script>

<template>
  <div class="stage side-by-side">
    <div class="col">
      <div class="label">{{ pdLabel }}</div>
      <HlImage v-if="pdUrl" :src="pdUrl" :alt="names.ref" :zoom-pct="zoomPct" :box="hlPd" />
      <PageMissing v-else :label="names.ref" reason="В этом файле нет такой страницы (номер больше, чем страниц в документе) или файл не выбран." />
    </div>
    <div class="col">
      <div class="label">{{ rdLabel }}</div>
      <HlImage v-if="rdUrl" :src="rdUrl" :alt="names.target" :zoom-pct="zoomPct" :box="hlRd" />
      <PageMissing v-else :label="names.target" reason="В этом файле нет такой страницы (номер больше, чем страниц в документе) или файл не выбран." />
    </div>
  </div>
</template>
