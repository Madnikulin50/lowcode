<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">
        {{ t('geosearch.title') }}
      </h4>
    </div>

    <form
      @submit.prevent="emit('submit', settings)"
    >
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
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: _t } = useI18n()

function t(key, ...args) {
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

const locationSettings = ref({})

const selectedProvider = computed(() => {
  return providerOptions.find(p => p.value === locationSettings.value.geoSearchProvider)
})

const requiresApiKey = computed(() => {
  return selectedProvider.value?.requiresApiKey || false
})

const isSubmitDisabled = computed(() => {
  return requiresApiKey.value && !locationSettings.value.geoSearchApiKey
})

watch(() => props.settings, (settings) => {
  locationSettings.value = settings['ui.location'] || {}

  if (!locationSettings.value.geoSearchProvider) {
    locationSettings.value.geoSearchProvider = 'openstreetmap'
  }

  if (!locationSettings.value.geoSearchApiKey) {
    locationSettings.value.geoSearchApiKey = ''
  }
}, { immediate: true })

watch(() => locationSettings.value.geoSearchProvider, () => {
  locationSettings.value.geoSearchApiKey = ''
})

function onSubmit () {
  emit('submit', { 'ui.location': locationSettings.value })
}
</script>
