<script setup>
defineProps({ viewer: Object, pageCount: Number })
</script>

<template>
  <div class="pager">
    <button :disabled="viewer.page <= 1" @click="viewer.page--">‹ Пред.</button>
    <span>Стр. {{ viewer.page }} / {{ pageCount }}</span>
    <button :disabled="viewer.page >= pageCount" @click="viewer.page++">След. ›</button>
  </div>
</template>
