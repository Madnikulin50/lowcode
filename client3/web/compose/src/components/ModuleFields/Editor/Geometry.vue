 <template>
   <div class="mb-3" :class="formGroupStyleClasses">
     <div v-if="!valueOnly" class="d-flex align-items-center text-primary p-0">
       <span :title="label" class="d-inline-block mw-100">{{ label }}</span>
       <c-hint :tooltip="hint" />
       <slot name="tools" />
     </div>
     <div class="small text-muted" :class="{ 'mb-1': description }">{{ description }}</div>

     <button v-if="field.isMulti" class="btn btn-outline-secondary btn-sm w-100 mb-3" :title="t('tooltip.openMap')" @click="openMap()">
       <font-awesome-icon :icon="['fas', 'map-marked-alt']" class="text-primary" />
     </button>

     <multi v-if="field.isMulti" v-slot="ctx" v-model:value="localValue" :errors="errors" :default-value="{ coordinates: [] }">
       <div class="input-group input-group-sm">
         <input v-model="localValue[ctx.index].coordinates[0]" type="number" step="0.000001" class="form-control form-control-sm" :placeholder="t('latitude')" />
         <input v-model="localValue[ctx.index].coordinates[1]" type="number" step="0.000001" class="form-control form-control-sm" :placeholder="t('longitude')" />
         <button class="btn btn-outline-secondary d-flex align-items-center" :title="t('tooltip.openMap')" @click="openMap(ctx.index)">
           <font-awesome-icon :icon="['fas', 'map-marked-alt']" class="text-primary" />
         </button>
         <button v-if="!field.options.hideCurrentLocationButton" class="btn btn-outline-secondary d-flex align-items-center" :title="t('tooltip.useCurrentLocation')" @click="useCurrentLocation(ctx.index)">
           <font-awesome-icon :icon="['fas', 'location-arrow']" class="text-primary" />
         </button>
       </div>
     </multi>

     <template v-else>
       <div class="input-group input-group-sm">
         <input v-model="localValue.coordinates[0]" type="number" step="0.000001" class="form-control form-control-sm" :placeholder="t('latitude')" />
         <input v-model="localValue.coordinates[1]" type="number" step="0.000001" class="form-control form-control-sm" :placeholder="t('longitude')" />
         <button class="btn btn-outline-secondary d-flex align-items-center" :title="t('tooltip.openMap')" @click="openMap()">
           <font-awesome-icon :icon="['fas', 'map-marked-alt']" class="text-primary" />
         </button>
         <button v-if="!field.options.hideCurrentLocationButton" class="btn btn-outline-secondary d-flex align-items-center" :title="t('tooltip.useCurrentLocation')" @click="useCurrentLocation()">
           <font-awesome-icon :icon="['fas', 'location-arrow']" class="text-primary" />
         </button>
       </div>
       <errors :errors="errors" />
     </template>

     <div v-if="map.show" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);" @click.self="closeMap">
       <div class="modal-dialog modal-xl">
         <div class="modal-content">
           <div class="modal-header">
             <h5 class="modal-title">{{ field.label || field.name }}</h5>
             <button type="button" class="btn-close" @click="closeMap"></button>
           </div>
           <div class="modal-body p-0">
             <CMap
               :map="map"
               :hide-geo-search="field.options.hideGeoSearch"
               :hide-current-location-button="field.options.hideCurrentLocationButton"
               :markers="markers"
               :labels="{ tooltip: { 'goToCurrentLocation': t('tooltip.goToCurrentLocation') } }"
               style="height: 75vh; width: 100%; cursor: pointer;"
               @on-map-click="placeMarker"
               @on-marker-click="removeMarker"
               @location-found="placeMarker($event, localValueIndex, true)"
               @on-geosearch-error="onGeoSearchError"
             />
           </div>
           <div class="modal-footer d-flex align-items-center">
             <span>{{ t('clickToPlaceMarker') }}</span>
             <button class="btn btn-outline-secondary btn-sm ms-auto" @click="closeMap">{{ t('label.cancel') }}</button>
             <button class="btn btn-primary btn-sm" @click="saveMapValue">{{ t('label.save') }}</button>
           </div>
         </div>
       </div>
     </div>
   </div>
 </template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field', keyPrefix: 'kind.geometry' } })
import { ref, computed, watch, onBeforeUnmount, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditorBase } from './base'
import { NoID } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import { isNumber } from 'lodash'
import errors from '../errors'
import multi from './multi'
const { CMap } = components

const props = defineProps({
  namespace: { type: Object, required: true },
  field: { type: Object, required: true },
  record: { type: Object, required: true },
  errors: { type: Object, required: true },
  valueOnly: { type: Boolean, default: false },
  horizontal: { type: Boolean, default: false },
  extraOptions: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['change', 'update:preventPopoverClose'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { value, formGroupStyleClasses, label, hint, description } = useEditorBase(props, emit)

const $root = inject('$root')

const localValue = ref(undefined)
const localValueIndex = ref(undefined)
const map = ref({ show: false, value: undefined })

const markers = computed(() => {
  if (!map.value.value) return []
  let m = [{ value: map.value.value.coordinates, opacity: 1.0 }]
  if (props.field.isMulti) {
    m = map.value.value.map(({ coordinates }, i) => ({
      value: coordinates && coordinates.length ? coordinates : undefined,
      opacity: localValueIndex.value === undefined || i === localValueIndex.value ? 1.0 : 0.6,
    }))
  }
  return m
})

watch(localValue, (val) => {
    const newValue = props.field.isMulti ? val.filter(v => (v || {}).coordinates).map(v => JSON.stringify(v)) : JSON.stringify(val)
    if (JSON.stringify(newValue) === JSON.stringify(value.value)) return
    value.value = newValue
  }, { deep: true })

watch(() => props.field.isMulti, () => {
    if (props.field.isMulti) {
      localValue.value = (value.value || []).map(v => JSON.parse(v || '{"coordinates":[]}'))
    } else {
      localValue.value = JSON.parse(value.value || '{"coordinates":[]}')
    }
  }, { immediate: true })

watch(() => props.field.options.prefillWithCurrentLocation, (v) => {
    if (v && (!props.record || props.record.recordID === NoID)) {
      useCurrentLocation()
    }
  }, { immediate: true })

onBeforeUnmount(() => {
  setDefaultValues()
})

function openMap (index) {
  map.value.value = props.field.isMulti ? [...localValue.value] : localValue.value
  localValueIndex.value = index

  const firstCoordinates = (index >= 0 ? localValue.value[index] : localValue.value) || {}
  const areCoordinatesValid = firstCoordinates.coordinates && firstCoordinates.coordinates.length === 2 && firstCoordinates.coordinates.every(isNumber)

  firstCoordinates.coordinates = areCoordinatesValid ? [...firstCoordinates.coordinates] : []

  map.value.center = areCoordinatesValid ? firstCoordinates.coordinates : props.field.options.center
  map.value.zoom = areCoordinatesValid ? 13 : props.field.options.zoom
  map.value.show = true
}

function closeMap () {
  map.value.show = false
}

function placeMarker (e, index = localValueIndex.value, useMap = true) {
  const { lat = 0, lng = 0 } = e.latlng || {}
  const coords = { coordinates: [Math.round(lat * 1e7) / 1e7, Math.round(lng * 1e7) / 1e7] }

  if (props.field.isMulti) {
    if (index >= 0) {
      useMap ? map.value.value.splice(index, 1, coords) : localValue.value.splice(index, 1, coords)
    } else {
      useMap ? map.value.value.push(coords) : localValue.value.push(coords)
    }
  } else {
    useMap ? map.value.value = coords : localValue.value = coords
  }
}

function removeMarker ({ index }) {
  if (props.field.isMulti) {
    map.value.value.splice(index, 1)
  } else {
    map.value.value = { coordinates: [] }
  }
}

function saveMapValue () {
  localValue.value = map.value.value
  closeMap()
}

function useCurrentLocation (index) {
  try {
    if (!navigator.geolocation) return
    navigator.geolocation.getCurrentPosition(
      ({ coords }) => {
        const latlng = { lat: coords.latitude, lng: coords.longitude }
        placeMarker({ latlng }, index, false)
      },
      () => {},
    )
  } catch (_) {}
}

function onGeoSearchError () {}

function setDefaultValues () {
  localValue.value = undefined
  localValueIndex.value = undefined
  map.value = {}
}
</script>
