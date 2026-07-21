<template>
  <div>
    <router-link :to="link">
      {{ link.label }}
    </router-link>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  link: { label: string; to: string }
}>()
</script>
