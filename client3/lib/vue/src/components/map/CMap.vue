<template>
  <div class="position-relative">
    <div
      v-if="!hideGeoSearch"
      class="geosearch-container"
      @mouseover="disableMap"
      @mouseleave="enableMap"
    >
      <CInputSearch
        v-model="geoSearch.query"
        :placeholder="labels.geosearchInputPlaceholder"
        autocomplete="off"
        :debounce="300"
        @input="onGeoSearch"
      />

      <div class="geosearch-results">
        <div
          v-for="(result, idx) in geoSearch.results"
          :key="idx"
          class="geosearch-result"
          @click="placeGeoSearchMarker(result)"
        >
          {{ result.label }}
        </div>
      </div>
    </div>

    <l-map
      ref="map"
      :zoom="mapOptions.zoom"
      :center="mapOptions.center"
      :min-zoom="mapOptions.minZoom"
      :max-zoom="mapOptions.maxZoom"
      :bounds="mapOptions.bounds"
      :max-bounds="mapOptions.maxBounds"
      class="w-100 h-100"
      @click="onMapClick"
      @locationfound="onLocationFound"
      @update:zoom="onZoom"
      @update:center="onCenter"
      @update:bounds="onBoundsUpdate"
    >
      <l-tile-layer
        :url="tileLayerURL"
        :attribution="tileAttribution"
        :max-zoom="tileMaxZoom"
        :min-zoom="tileMinZoom"
      />

      <l-polygon
        v-for="(geometry, i) in polygons"
        :key="`polygon-${i}`"
        :lat-lngs="geometry.map((value: any) => value.geometry)"
        :color="(geometry.find((g: any) => g) || {}).color"
      />

      <l-marker
        v-for="(marker, i) in markerValues"
        :key="`marker-${i}`"
        :lat-lng="marker.value"
        :icon="getIcon(marker.color)"
        :opacity="marker.opacity || 1.0"
        @click="onMarkerClick(i, marker)"
      >
        <l-tooltip
          v-if="$slots['marker-tooltip'] || marker.title"
          :options="{
            offset: [-1, 5],
            direction: 'bottom',
          }"
        >
          <slot
            name="marker-tooltip"
            :marker="marker"
          >
            {{ marker.title }}
          </slot>
        </l-tooltip>
      </l-marker>

      <l-marker
        v-if="geoSearch.marker"
        :lat-lng="geoSearch.marker.latlng"
        :icon="getIcon(getCSSVariable('--secondary'))"
      >
        <l-tooltip
          :options="{
            offset: [-1, 5],
            direction: 'bottom',
          }"
        >
          {{ geoSearch.marker.title }}
        </l-tooltip>
      </l-marker>

      <l-control class="leaflet-bar">
        <a
          v-if="!hideCurrentLocationButton"
          title="Go to current location"
          role="button"
          class="d-flex justify-content-center align-items-center"
          @click="goToCurrentLocation"
        >
          <font-awesome-icon
            :icon="['fas', 'location-crosshairs']"
          />
        </a>
      </l-control>
    </l-map>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, getCurrentInstance, nextTick } from 'vue'
import { divIcon, latLng, latLngBounds } from 'leaflet'
import {
  OpenStreetMapProvider,
  OpenCageProvider,
  EsriProvider,
  GeoapifyProvider,
  GeocodeEarthProvider,
  GoogleProvider,
  LocationIQProvider,
} from 'leaflet-geosearch'
import { isNumber } from 'lodash'
import { LControl, LMap, LMarker, LPolygon, LTileLayer, LTooltip } from '@vue-leaflet/vue-leaflet'
import CInputSearch from '../input/CInputSearch.vue'

import 'leaflet/dist/leaflet.css'

const { proxy }: any = getCurrentInstance()!

const props = withDefaults(defineProps<{
  hideCurrentLocationButton?: boolean
  labels?: Record<string, any>
  map?: Record<string, any>
  markers?: any[]
  polygons?: any[]
  hideGeoSearch?: boolean
  disabled?: boolean
}>(), {
  hideCurrentLocationButton: false,
  labels: () => ({}),
  map: () => ({}),
  markers: () => [],
  polygons: () => [],
  hideGeoSearch: false,
  disabled: false,
})

const emit = defineEmits<{
  'on-geosearch-error': []
  'location-found': [value: any]
  'on-marker-click': [value: any]
  'on-map-click': [value: any]
  'on-zoom': [value: any]
  'on-center': [value: any]
  'on-bounds-update': [value: any]
}>()

const map = ref<typeof LMap>()

const geoSearch = ref<{
  query: string
  results: any[]
  marker: any | null
}>({
  query: '',
  results: [],
  marker: null,
})

const DEFAULT_TILE_URL = 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'
const DEFAULT_ATTRIBUTION = '&copy; <a target="_blank" href="http://osm.org/copyright">OpenStreetMap</a>'

const markerValues = computed(() => props.markers.map((m: any) => ({
  ...m,
  value: getLatLng(m.value),
})).filter((c: any) => c.value) || [])

const settingsTileURL = computed(() => String(proxy.$Settings?.get?.('ui.map.tileURL', '') || '').trim())
const settingsAttribution = computed(() => String(proxy.$Settings?.get?.('ui.map.attribution', '') || '').trim())
const settingsMaxZoom = computed(() => Number(proxy.$Settings?.get?.('ui.map.maxZoom', 0) || 0))
const settingsMinZoom = computed(() => Number(proxy.$Settings?.get?.('ui.map.minZoom', 0) || 0))

const tileLayerURL = computed(() => {
  if (props.map?.tileURL) return props.map.tileURL
  if (settingsTileURL.value) return settingsTileURL.value
  return DEFAULT_TILE_URL
})

const tileAttribution = computed(() => {
  if (props.map?.attribution) return props.map.attribution
  if (settingsAttribution.value) return settingsAttribution.value
  return DEFAULT_ATTRIBUTION
})

const tileMaxZoom = computed(() => {
  if (isNumber(props.map?.maxZoom)) return props.map.maxZoom
  if (settingsMaxZoom.value > 0) return settingsMaxZoom.value
  return 18
})

const tileMinZoom = computed(() => {
  if (isNumber(props.map?.minZoom)) return props.map.minZoom
  if (settingsMinZoom.value > 0) return settingsMinZoom.value
  return 0
})

const mapOptions = computed(() => {
  const mapOpts = { ...props.map }
  const defaultOptions = {
    zoom: 3,
    center: [30, 30],
    rotation: 0,
    attribution: tileAttribution.value,
    maxZoom: tileMaxZoom.value,
    minZoom: tileMinZoom.value,
  }
  mapOpts.bounds = mapOpts.bounds ? latLngBounds(mapOpts.bounds) : null
  if (!isNumber(mapOpts.maxZoom)) mapOpts.maxZoom = defaultOptions.maxZoom
  if (!isNumber(mapOpts.minZoom)) mapOpts.minZoom = defaultOptions.minZoom
  if (!mapOpts.attribution) mapOpts.attribution = defaultOptions.attribution
  const merged = { ...defaultOptions, ...mapOpts }
  const c = merged.center
  if (!Array.isArray(c) || c.length < 2 || c[0] == null || c[1] == null || !Number.isFinite(Number(c[0])) || !Number.isFinite(Number(c[1]))) {
    merged.center = defaultOptions.center
  }
  return merged
})

const geoSearchApiKey = computed(() => proxy.$Settings.get('ui.location.geoSearchApiKey', ''))
const geoSearchProviderName = computed(() => proxy.$Settings.get('ui.location.geoSearchProvider', ''))

const geoSearchProvider = computed(() => {
  const providerName = (geoSearchProviderName.value || 'openstreetmap').toLowerCase()
  const apiKey = geoSearchApiKey.value
  const providers: Record<string, () => any> = {
    openstreetmap: () => new OpenStreetMapProvider(),
    opencage: () => new OpenCageProvider({ params: { key: apiKey } }),
    esri: () => new EsriProvider(),
    geoapify: () => new GeoapifyProvider({ params: { apiKey } }),
    geocodeearth: () => new GeocodeEarthProvider({ params: { api_key: apiKey } }),
    google: () => new GoogleProvider({ apiKey }),
    locationiq: () => new LocationIQProvider({ params: { key: apiKey } }),
  }
  if (providers[providerName]) return providers[providerName]()
  console.warn(`Unknown geosearch provider: ${providerName}, falling back to OpenStreetMap`)
  return new OpenStreetMapProvider()
})

onMounted(() => {
  if (map.value && map.value.leafletObject) {
    onBoundsUpdate(map.value.leafletObject.getBounds())
  }
})

function onGeoSearch(query: string) {
  if (!query || !geoSearchProvider.value) {
    geoSearch.value.results = []
    geoSearch.value.marker = null
    return
  }
  geoSearchProvider.value.search({ query }).then((results: any[]) => {
    geoSearch.value.results = results.map((result: any) => {
      let lat: number, lng: number
      if (result.y !== undefined && result.x !== undefined) {
        lat = result.y
        lng = result.x
      } else if (result.raw) {
        lat = result.raw.lat || result.raw.latitude
        lng = result.raw.lon || result.raw.lng || result.raw.longitude
      }
      return { ...result, latlng: { lat, lng } }
    })
  }).catch(() => {
    emit('on-geosearch-error')
  })
}

function placeGeoSearchMarker(result: any) {
  const mapObj = map.value?.leafletObject
  const zoom = mapObj ? (mapObj._zoom >= 15 ? mapObj._zoom : 15) : 15
  mapObj?.flyTo([result.latlng.lat, result.latlng.lng], zoom, { animate: false })
  geoSearch.value.marker = { title: result.label, latlng: result.latlng }
  geoSearch.value.results = []
  onMapClick(result)
}

function getLatLng(coordinates: any[] = [undefined, undefined]) {
  const [lat, lng] = coordinates
  if (isNumber(lat) && isNumber(lng)) {
    return latLng(lat, lng)
  }
}

function onLocationFound({ latitude, longitude }: any) {
  const mapObj = map.value?.leafletObject
  const zoom = mapObj ? (mapObj._zoom >= 15 ? mapObj._zoom : 15) : 15
  mapObj?.flyTo([latitude, longitude], zoom)
  emit('location-found', { latlng: { lat: latitude, lng: longitude } })
}

function disableMap() {
  if (props.disabled) {
    const mapObj = map.value?.leafletObject
    mapObj?._handlers.forEach((handler: any) => handler.disable())
  }
}

function enableMap() {
  if (props.disabled) {
    const mapObj = map.value?.leafletObject
    mapObj?._handlers.forEach((handler: any) => handler.enable())
  }
}

function onMarkerClick(index: number, marker: any) {
  emit('on-marker-click', { index, marker })
}

function goToCurrentLocation() {
  map.value?.leafletObject?.locate()
}

function onMapClick(e: any) {
  emit('on-map-click', e)
}

function getCSSVariable(variable: string) {
  return getComputedStyle(document.documentElement).getPropertyValue(variable)
}

function getIcon(markerColor = getCSSVariable('--primary')) {
  const markerIconHtml = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 34.892337" height="60" width="40" style="margin-top: -40px;margin-left: -15px;height: 35px;">
      <g transform="translate(-814.59595,-274.38623)">
        <g transform="matrix(1.1855854,0,0,1.1855854,-151.17715,-57.3976)">
          <path d="m 817.11249,282.97118 c -1.25816,1.34277 -2.04623,3.29881 -2.01563,5.13867 0.0639,3.84476 1.79693,5.3002 4.56836,10.59179 0.99832,2.32851 2.04027,4.79237 3.03125,8.87305 0.13772,0.60193 0.27203,1.16104 0.33416,1.20948 0.0621,0.0485 0.19644,-0.51262 0.33416,-1.11455 0.99098,-4.08068 2.03293,-6.54258 3.03125,-8.87109 2.77143,-5.29159 4.50444,-6.74704 4.56836,-10.5918 0.0306,-1.83986 -0.75942,-3.79785 -2.01758,-5.14062 -1.43724,-1.53389 -3.60504,-2.66908 -5.91619,-2.71655 -2.31115,-0.0475 -4.4809,1.08773 -5.91814,2.62162 z" style="fill:${markerColor};stroke:${markerColor};"/>
          <circle r="3.0355" cy="288.25278" cx="823.03064" id="path3049" style="display:inline;fill:#FFFFFF;"/>
        </g>
      </g>
    </svg>`
  return divIcon({ className: 'marker-pin', html: markerIconHtml })
}

function onZoom(e: any) {
  emit('on-zoom', e)
}

function onCenter(e: any) {
  emit('on-center', e)
}

function onBoundsUpdate(value: any) {
  nextTick(() => {
    setTimeout(() => {
      map.value?.leafletObject?.invalidateSize()
    }, 100)
  })
  value = value || map.value?.leafletObject?.getBounds()
  emit('on-bounds-update', value)
}
</script>

<style lang="scss">
.leaflet-touch .leaflet-bar {
  border: 1px solid transparent;
  border-radius: 0.3rem;
}

.leaflet-bar a {
  background-color: var(--white) !important;
  color: var(--primary) !important;
  text-decoration: none !important;

  &:hover {
    background-color: var(--white) !important;
    transition: background-color 0.15s ease;
  }
}

.geosearch-result {
  &:hover {
    background-color: var(--light) !important;
    color: var(--black);
  }

  &:active {
    color: var(--white) !important;
    background-color: var(--primary) !important;
  }
}
</style>

<style scoped>
.geosearch-container {
  position: absolute;
  display: block;
  height: auto;
  width: 50%;
  max-width: 50%;
  cursor: auto;
  z-index: 1030;
  left: 50%;
  transform: translateX(-50%);
  top: 10px;
}

.geosearch-results {
  margin: 1px;
  border-radius: 2px;
  background-color: var(--white);
  max-height: 50%;
  overflow: auto;
}

.geosearch-result {
  border-radius: 2px;
  line-height: 32px;
  padding: 0 8px;
  font-size: 12px;
  white-space: nowrap;
}

.geosearch-result:hover {
  background-color: var(--gray-200);
  cursor: pointer;
}
</style>
