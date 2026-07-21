<template>
  <div class="container-fluid d-flex justify-content-center vh-100 logo">
    <img
      v-if="logo"
      :src="logo"
      class="w-25 my-auto"
    >
  </div>
</template>

<script setup lang="ts">
defineProps<{
  logo?: string
}>()
</script>

<style lang="scss" scoped>
.logo {
  img {
    animation: pulse 4.2s infinite;
  }

  @keyframes pulse {
    0% {
      opacity: 0;
    }
    65% {
      opacity: 1;
    }
    100% {
      opacity: 0;
    }
  }
}
</style>
