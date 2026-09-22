<template>
  <Teleport v-if="ready && title" to="#topbar-title">
    <div class="c-page-title">
      <div class="c-page-title__name">{{ title }}</div>
      <div v-if="subtitle" class="c-page-title__sub">{{ subtitle }}</div>
    </div>
  </Teleport>
  <Teleport v-if="ready && hasActions" to="#topbar-tools">
    <div class="c-page-actions d-flex align-items-center flex-wrap gap-1">
      <slot />
    </div>
  </Teleport>
</template>

<script setup>
import { computed, onMounted, ref, useSlots } from 'vue'

defineProps({
  title: { type: String, default: '' },
  subtitle: { type: String, default: '' },
})

const slots = useSlots()
const ready = ref(false)
const hasActions = computed(() => !!slots.default)

onMounted(() => { ready.value = true })
</script>
