<template>
  <div class="card result shadow-sm h-100" :class="{ 'shadow': hovered }" @mouseover="onHover" @mouseleave="onLeave">
    <a v-if="hit.value?.url" :href="hit.value.url" target="_blank" rel="noopener noreferrer" class="stretched-link" />
    <component :is="component" v-bind="$props" />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import * as Results from './loader'

const props = defineProps({
  index: {
    type: Number,
    required: true,
  },
  hit: {
    type: Object,
    required: true,
  },
  showMap: {
    type: Boolean,
    required: true,
  },
})

const emit = defineEmits(['hover'])

const hovered = ref(false)

const component = computed(() => {
  const { type } = props.hit
  const resourceType = type.split(':')[1]
  const keys = Object.keys(Results)
  const i = keys.map(c => c.toLocaleLowerCase()).findIndex(c => c === resourceType)
  return Results[keys[i]]
})

function onHover() {
  hovered.value = true
  emit('hover', props.hit.value?.recordID)
}

function onLeave() {
  hovered.value = false
  emit('hover', undefined)
}
</script>

<style lang="scss" scoped>
.result {
  transition: all 0.3s ease-in;
}
</style>
