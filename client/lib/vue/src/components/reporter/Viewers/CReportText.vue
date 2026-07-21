<template>
  <div
    v-if="options"
    class="rt-content p-3"
    v-html="options.value"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  displayElement: any
  labels?: Record<string, any>
}>()

const options = computed(() => props.displayElement?.options || undefined)
</script>
