<template>
  <Teleport v-if="title" to="#topbar-title">
    <h3 class="me-auto">{{ title }}</h3>
  </Teleport>

  <div class="d-flex justify-content-end flex-wrap gap-1 mb-2 flex-fill-child">
    <slot />
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, required: false, default: '' },
})
</script>