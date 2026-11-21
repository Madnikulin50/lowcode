<script setup>
// Only place we need raw markup (inline SVG icons, or a plain status glyph
// like '⏳') — everything else stays text-safe by going through {{ }}.
defineProps({ svg: String, cls: { type: String, default: '' } })
</script>

<template>
  <span :class="cls" v-html="svg"></span>
</template>
