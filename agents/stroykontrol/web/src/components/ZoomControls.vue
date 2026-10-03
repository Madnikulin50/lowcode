<script setup>
import { ZOOM_LEVELS } from '../constants.js'

defineProps({ viewer: Object })
</script>

<template>
  <div class="zoom">
    <button :disabled="viewer.zoomIdx <= 0" @click="viewer.zoomIdx--">−</button>
    <span class="zlevel">{{ ZOOM_LEVELS[viewer.zoomIdx] }}%</span>
    <button :disabled="viewer.zoomIdx >= ZOOM_LEVELS.length - 1" @click="viewer.zoomIdx++">+</button>
    <button @click="viewer.zoomIdx = ZOOM_LEVELS.indexOf(100)">Сброс</button>
  </div>
</template>
