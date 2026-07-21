<template>
  <c-map
    :map="{
      zoom,
    }"
    hide-geo-search
    hide-current-location-button
    :markers="validMarkerValues"
    style="min-height: 400px; height: 100% !important;"
  >
    <template #marker-tooltip="{ marker }">
      <h5 class="text-primary">{{ t('map.server-details') }}</h5>

      <div class="mb-2">
        <label class="text-primary form-label">{{ t('map.name') }}</label>
        <div>{{ marker.meta.name }}</div>
      </div>

      <div class="mb-2">
        <label class="text-primary form-label">{{ t('map.location') }}</label>
        <div>{{ getLocationName(marker) }}</div>
      </div>
    </template>
  </c-map>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'

const { CMap } = components
const { t } = useI18n()

const props = defineProps({
  connections: { type: Array, required: true },
})

const zoom = 2

const validMarkerValues = computed(() => {
  return props.connections
    .filter(({ meta = {} }) => {
      const { location = {} } = meta
      const { geometry = {} } = location
      const { coordinates = [] } = geometry
      return coordinates && !!coordinates.length
    })
    .map((connection) => {
      return {
        id: connection.id,
        value: getLocationCoordinates(connection),
        ...connection,
      }
    })
})

function getLocationCoordinates ({ meta = {} }) {
  const { location = {} } = meta
  const { geometry = {} } = location
  return geometry.coordinates
}

function getLocationName (connection) {
  return connection.meta.location.properties.name || t('map.unnamed-location')
}
</script>

<style lang="scss">
.vl-style-text {
  color: var(--white);
}
</style>