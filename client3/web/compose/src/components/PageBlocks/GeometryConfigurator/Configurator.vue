<template>
  <div>
    <div>
      <c-map
        :map="mapOptions"
        :labels="{
          tooltip: { 'goToCurrentLocation': $t('geometry.tooltip.goToCurrentLocation') },
        }"
        :hide-geo-search="options.hideGeoSearch"
        class="w-100 cursor-pointer"
        style="height: 45vh;"
        @on-bounds-update="boundsUpdated"
        @on-center="updateCenter"
        @on-zoom="options.zoomStarting = $event"
      />
      <div class="form-text" id="password-help-block">
        {{ $t('geometry.mapHelpText') }}
      </div>
    </div>

    <hr />

    <div class="row">
      <div class="col-12 col-lg-4">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('geometry.zoom.zoomStartingLabel') }}</label>
          <input
            v-model="options.zoomStarting"
            type="number"
            class="form-control"
            readonly
          />
        </div>
      </div>

      <div class="col-12 col-lg-4">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('geometry.zoom.zoomMinLabel') }}</label>
          <small class="d-block">{{ options.zoomMin }}</small>
          <input
            v-model="options.zoomMin"
            type="range"
            class="form-range"
            min="1"
            max="18"
          />
        </div>
      </div>

      <div class="col-12 col-lg-4">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('geometry.zoom.zoomMaxLabel') }}</label>
          <small class="d-block">{{ options.zoomMax }}</small>
          <input
            v-model="options.zoomMax"
            type="range"
            class="form-range"
            min="1"
            max="18"
          />
        </div>
      </div>

      <div class="col-12 col-lg-4">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('geometry.onMarkerClick') }}</label>
          <select
            v-model="options.displayOption"
            class="form-select"
          >
            <option
              v-for="opt in displayOptions"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.text }}
            </option>
          </select>
        </div>
      </div>

      <div class="col-12 col-lg-4">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('geometry.bounds.lockBounds') }}</label>
          <c-input-checkbox
            v-model="options.lockBounds"
            switch
            :labels="checkboxLabel"
          />
        </div>
      </div>

      <div class="col-12 col-lg-4">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('geometry.hideGeoSearch') }}</label>
          <c-input-checkbox
            v-model="options.hideGeoSearch"
            switch
            :labels="checkboxLabel"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onBeforeUnmount } from 'vue'
import { components } from 'corteza-lib/vue/dist'
import { usePageBlockBase } from '../usePageBlockBase'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
const { CMap } = components

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, required: false, default: '' },
  editable: { type: Boolean, required: false, default: false },
  resizing: { type: Boolean, required: false, default: false },
  magnified: { type: Boolean, required: false, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, required: false, default: false },
  errors: { type: Object, required: false, default: () => ({}) },
})

const emit = defineEmits(['errors'])

const { options, setBaseDefaultValues } = usePageBlockBase(props, emit)

const map = ref({})
const localValue = ref({ coordinates: [] })
const center = ref([])
const bounds = ref(null)
const checkboxLabel = ref({ on: t('label.yes'), off: t('label.no') })

const displayOptions = computed(() => [
  { value: 'sameTab', text: t('geometry.openInSameTab') },
  { value: 'newTab', text: t('geometry.openInNewTab') },
  { value: 'modal', text: t('geometry.openInModal') },
])

const mapOptions = computed({
  get: () => ({
    zoom: options.value.zoomStarting,
    minZoom: options.value.zoomMin,
    maxZoom: options.value.zoomMax,
    center: options.value.center,
    bounds: bounds.value,
    maxBounds: options.value.bounds,
  }),
  set: (val) => {
    options.value.zoomStarting = val.zoom
    options.value.zoomMin = val.minZoom
    options.value.zoomMax = val.maxZoom
    options.value.center = val.center
    options.value.bounds = val.center
    bounds.value = val.maxBounds
  },
})

onBeforeUnmount(() => { setDefaultValues() })

function updateCenter (coordinates) {
  let { lat = 0, lng = 0 } = coordinates || {}
  lat = Math.round(lat * 1e7) / 1e7
  lng = Math.round(lng * 1e7) / 1e7
  options.value.center = [lat, lng]
}

function boundsUpdated (coordinates) {
  bounds.value = coordinates
  updateBounds(options.value.lockBounds)
}

function updateBounds (value) {
  if (value) {
    const b = bounds.value || {}
    const { _northEast, _southWest } = b
    options.value.bounds = [Object.values(_northEast), Object.values(_southWest)]
  } else {
    options.value.bounds = null
  }
}

function setDefaultValues () {
  map.value = {}
  localValue.value = { coordinates: [] }
  center.value = []
  bounds.value = null
}
</script>
