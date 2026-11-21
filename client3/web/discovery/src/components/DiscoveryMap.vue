<template>
  <c-map
    style="height: calc(100vh - 64px);"
    :map="{ zoom, center }"
    :markers="makerValues"
    @on-marker-click="onMarkerClick"
    @on-map-click="clearClickedMarker"
  />
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { components } from 'corteza-lib/vue/dist'

const { CMap } = components

const props = defineProps({
  markers: {
    type: Array,
    required: true,
  },
  hoverIndex: {
    type: Number,
    required: false,
    default: undefined,
  },
})

const emit = defineEmits(['marker-clicked'])

const zoom = ref(3)
const center = ref([30, 30])
const rotation = ref(0)
const attribution = ref('&copy; <a target="_blank" rel="noopener noreferrer" href="http://osm.org/copyright">OpenStreetMap</a>')
const clickedMarker = ref(undefined)

const makerValues = computed(() => {
  return props.markers.map((marker, i) => ({
    id: i,
    value: marker.coordinates,
    opacity: [props.hoverIndex, clickedMarker.value].includes(i) ? 1.0 : 0.6,
    color: 'var(--primary)',
  }))
})

watch(() => props.markers, (markers = []) => {
  if (markers.length) {
    const { coordinates = [30, 30] } = markers[0]
    center.value = coordinates
  }
}, { immediate: true })

watch(() => props.hoverIndex, (hoverIndex) => {
  if (hoverIndex) {
    const marker = props.markers.find(({ id }) => id === hoverIndex)
    if (marker?.coordinates) {
      center.value = marker.coordinates
    }
  }
  clickedMarker.value = undefined
})

function onMarkerClick({ index }) {
  if (index === clickedMarker.value) {
    clickedMarker.value = undefined
  } else {
    clickedMarker.value = index
  }
  emit('marker-clicked', clickedMarker.value)
}

function clearClickedMarker() {
  clickedMarker.value = undefined
  emit('marker-clicked', clickedMarker.value)
}
</script>

<style lang="scss">
.vl-style-text {
  color: var(--white);
}
</style>
