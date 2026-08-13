<template>
  <div>
    <div class="card shadow-sm">
      <div class="card-header border-bottom">
        <h4 class="m-0">
          {{ t('geosearch.title') }}
        </h4>
      </div>

      <form @submit.prevent="onSubmit">
        <div class="card-body">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('geosearch.provider.label') }}</label>
            <div class="form-text mb-2">{{ t('geosearch.provider.description') }}</div>
            <c-input-select
              v-model="locationSettings.geoSearchProvider"
              :options="providerOptions"
              :reduce="o => o.value"
              label="text"
              :clearable="false"
            />
          </div>

          <div
            v-if="requiresApiKey"
            class="mb-3"
          >
            <label class="form-label text-primary">{{ t('geosearch.apiKey.label') }}</label>
            <input
              v-model="locationSettings.geoSearchApiKey"
              class="form-control"
              type="text"
              :placeholder="t('geosearch.apiKey.placeholder')"
            >
          </div>
        </div>
      </form>
    </div>

    <div class="card shadow-sm mt-3">
      <div class="card-header border-bottom">
        <h4 class="m-0">
          {{ t('map.title') }}
        </h4>
      </div>

      <div class="card-body">
        <p class="text-muted small">{{ t('map.help') }}</p>

        <div class="mb-3">
          <label class="form-label text-primary">{{ t('map.tileSource.label') }}</label>
          <div class="form-text mb-2">{{ t('map.tileSource.description') }}</div>
          <c-input-select
            v-model="mapSettings.tileSource"
            :options="tileSourceOptions"
            :reduce="o => o.value"
            label="text"
            :clearable="false"
          />
        </div>

        <div class="mb-3">
          <label class="form-label text-primary">{{ t('map.tileURL.label') }}</label>
          <div class="form-text mb-2">{{ t('map.tileURL.description') }}</div>
          <input
            v-model="mapSettings.tileURL"
            class="form-control font-monospace"
            type="text"
            :placeholder="tileURLPlaceholder"
            :disabled="!canManage"
          >
        </div>

        <div class="row g-3 mb-3">
          <div class="col-md-6">
            <label class="form-label text-primary">{{ t('map.minZoom.label') }}</label>
            <input
              v-model.number="mapSettings.minZoom"
              class="form-control"
              type="number"
              min="0"
              max="22"
              :disabled="!canManage"
            >
          </div>
          <div class="col-md-6">
            <label class="form-label text-primary">{{ t('map.maxZoom.label') }}</label>
            <input
              v-model.number="mapSettings.maxZoom"
              class="form-control"
              type="number"
              min="0"
              max="22"
              :disabled="!canManage"
            >
          </div>
        </div>

        <div class="mb-0">
          <label class="form-label text-primary">{{ t('map.attribution.label') }}</label>
          <div class="form-text mb-2">{{ t('map.attribution.description') }}</div>
          <input
            v-model="mapSettings.attribution"
            class="form-control"
            type="text"
            :placeholder="t('map.attribution.placeholder')"
            :disabled="!canManage"
          >
        </div>
      </div>

      <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
        <c-button-submit
          v-if="canManage"
          :processing="processing"
          :success="success"
          :disabled="isSubmitDisabled"
          :text="t('admin.general.label.submit')"
          class="ms-auto"
          @submit="onSubmit"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'ui.settings', keyPrefix: 'editor.location' } })
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: _t } = useI18n()

function t (key, ...args) {
  if (key.startsWith('label.') || key.startsWith('admin.')) return _t(key, ...args)
  return _t('editor.location.' + key, ...args)
}

const props = defineProps({
  settings: {
    type: Object,
    required: true,
  },
  processing: {
    type: Boolean,
    default: false,
  },
  success: {
    type: Boolean,
    default: false,
  },
  canManage: {
    type: Boolean,
    required: true,
  },
})

const emit = defineEmits(['submit'])

const DEFAULT_OSM_URL = 'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'

const providerOptions = [
  { value: 'openstreetmap', text: 'OpenStreetMap', requiresApiKey: false },
  { value: 'opencage', text: 'OpenCage', requiresApiKey: true },
  { value: 'esri', text: 'Esri', requiresApiKey: false },
  { value: 'geoapify', text: 'Geoapify', requiresApiKey: true },
  { value: 'geocodeearth', text: 'Geocode Earth', requiresApiKey: true },
  { value: 'google', text: 'Google Maps', requiresApiKey: true },
  { value: 'locationiq', text: 'LocationIQ', requiresApiKey: true },
  { value: 'mapbox', text: 'Mapbox', requiresApiKey: true },
  { value: 'pelias', text: 'Pelias', requiresApiKey: true },
]

const tileSourceOptions = [
  { value: 'online', text: 'Online (OpenStreetMap / custom URL)' },
  { value: 'local', text: 'Local / offline tile server' },
]

const locationSettings = ref({})
const mapSettings = ref({})

const selectedProvider = computed(() => {
  return providerOptions.find(p => p.value === locationSettings.value.geoSearchProvider)
})

const requiresApiKey = computed(() => {
  return selectedProvider.value?.requiresApiKey || false
})

const tileURLPlaceholder = computed(() => {
  if (mapSettings.value.tileSource === 'local') {
    return 'http://localhost:8081/{z}/{x}/{y}.png'
  }
  return DEFAULT_OSM_URL
})

const isSubmitDisabled = computed(() => {
  if (requiresApiKey.value && !locationSettings.value.geoSearchApiKey) return true
  if (mapSettings.value.tileSource === 'local' && !String(mapSettings.value.tileURL || '').trim()) return true
  return false
})

watch(() => props.settings, (settings) => {
  locationSettings.value = { ...(settings['ui.location'] || {}) }

  if (!locationSettings.value.geoSearchProvider) {
    locationSettings.value.geoSearchProvider = 'openstreetmap'
  }
  if (!locationSettings.value.geoSearchApiKey) {
    locationSettings.value.geoSearchApiKey = ''
  }

  mapSettings.value = { ...(settings['ui.map'] || {}) }
  if (!mapSettings.value.tileSource) {
    mapSettings.value.tileSource = 'online'
  }
  if (mapSettings.value.tileURL === undefined || mapSettings.value.tileURL === null) {
    mapSettings.value.tileURL = ''
  }
  if (!mapSettings.value.maxZoom) {
    mapSettings.value.maxZoom = 18
  }
  if (mapSettings.value.minZoom === undefined || mapSettings.value.minZoom === null || mapSettings.value.minZoom === '') {
    mapSettings.value.minZoom = 0
  }
  if (mapSettings.value.attribution === undefined || mapSettings.value.attribution === null) {
    mapSettings.value.attribution = ''
  }
}, { immediate: true })

watch(() => locationSettings.value.geoSearchProvider, () => {
  locationSettings.value.geoSearchApiKey = ''
})

function onSubmit () {
  emit('submit', {
    'ui.location': locationSettings.value,
    'ui.map': {
      tileSource: mapSettings.value.tileSource || 'online',
      tileURL: String(mapSettings.value.tileURL || '').trim(),
      minZoom: Number(mapSettings.value.minZoom) || 0,
      maxZoom: Number(mapSettings.value.maxZoom) || 18,
      attribution: String(mapSettings.value.attribution || '').trim(),
    },
  })
}
</script>
