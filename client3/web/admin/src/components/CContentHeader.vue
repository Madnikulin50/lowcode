<template>
  <div class="d-flex justify-content-end flex-wrap gap-1 mb-2 flex-fill-child">
    <h3 class="me-auto">{{ title }}</h3>
    <slot />
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, required: false, default: '' },
})
</script>