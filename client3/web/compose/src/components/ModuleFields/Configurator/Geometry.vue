<template>
  <div class="d-flex flex-column">
    <div class="mb-3">
      <div class="form-check">
        <input id="prefillWithCurrentLocation" v-model="f.options.prefillWithCurrentLocation" type="checkbox" class="form-check-input" />
        <label class="form-check-label" for="prefillWithCurrentLocation">{{ t('prefillWithCurrentLocation') }}</label>
      </div>
      <div class="form-check">
        <input id="hideCurrentLocationButton" v-model="f.options.hideCurrentLocationButton" type="checkbox" class="form-check-input" />
        <label class="form-check-label" for="hideCurrentLocationButton">{{ t('hideCurrentLocationButton') }}</label>
      </div>
      <div class="form-check">
        <input id="hideGeoSearch" v-model="f.options.hideGeoSearch" type="checkbox" class="form-check-input" />
        <label class="form-check-label" for="hideGeoSearch">{{ t('hideGeoSearch') }}</label>
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary mb-0">{{ t('initialZoomAndPosition') }}</label>
      <CMap
        :map="{ zoom, center }"
        :labels="{ tooltip: { 'goToCurrentLocation': t('tooltip.goToCurrentLocation') } }"
        style="height: 50vh;"
        @on-zoom="f.options.zoom = $event"
        @on-center="f.options.center = $event"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConfiguratorBase } from './base'
import { components } from 'corteza-lib/vue/dist'
const { CMap } = components

const props = defineProps({
  namespace: { type: Object, required: true },
  module: { type: Object, required: true },
  field: { type: Object, required: true },
  hasRecords: { type: Boolean, default: false },
})

const emit = defineEmits(['update:field'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { f, isNew, hasData } = useConfiguratorBase(props, emit)

const center = computed(() => f.value.options.center || [30, 30])
const zoom = computed(() => f.value.options.zoom || 3)
</script>
