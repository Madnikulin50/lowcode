<template>
  <Teleport v-if="title && ready" to="#topbar-title">
    <h3 class="me-auto">{{ title }}</h3>
  </Teleport>
  <div class="d-flex justify-content-end flex-wrap gap-1 mb-2 flex-fill-child">
    <slot />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

defineProps({ title: { type: String, default: '' } })

const ready = ref(false)
onMounted(() => { ready.value = true })
</script>
