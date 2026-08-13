<template>
  <div :class="classes">
    <span
      v-for="(c, index) of localValue"
      :key="index"
      :class="{ 'd-block': field.options.multiDelimiter === '\n' }"
    >
      <span v-if="disableClick">
        {{ c.value[0] }}, {{ c.value[1] }}
        {{ index !== localValue.length - 1 ? field.options.multiDelimiter : '' }}
      </span>
      <a
        v-else
        class="text-nowrap text-primary pointer"
        @click.stop="openMap(index)"
      >
        {{ c.value[0] }}, {{ c.value[1] }}
        {{ index !== localValue.length - 1 ? field.options.multiDelimiter : '' }}
      </a>
    </span>

    <div v-if="map.show" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);" @click.self="closeMap">
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ field.label || field.name }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" @click="closeMap"></button>
          </div>
          <div class="modal-body p-0">
            <CMap
              :map="map"
              :markers="localValue"
              :hide-current-location-button="field.options.hideCurrentLocationButton"
              :hide-geo-search="field.options.hideGeoSearch"
              :labels="{ tooltip: { 'goToCurrentLocation': t('tooltip.goToCurrentLocation') } }"
              style="height: 75vh; width: 100%;"
              @on-geosearch-error="onGeoSearchError"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field', keyPrefix: 'kind.geometry' } })
import { ref, computed, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useViewerBase } from './useViewerBase'
import { compose } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
const { CMap } = components

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  field: { type: compose.ModuleField, required: true },
  record: { type: compose.Record, required: true },
  valueOnly: { type: Boolean, required: false },
  extraOptions: { type: Object, default: () => ({}) },
  includeStyles: { type: Boolean, default: false },
  disableClick: { type: Boolean, default: false },
})

const { t } = useI18n({ useScope: 'global', messages: {} })
const { value, classes } = useViewerBase(props)

const localValueIndex = ref(undefined)
const map = ref({
  show: false,
  zoom: 14,
  center: [30, 30],
  rotation: 0,
  attribution: '&copy; <a target="_blank" rel="noopener noreferrer" href="http://osm.org/copyright">OpenStreetMap</a>',
})

const localValue = computed(() => {
  if (props.field.isMulti) {
    return (value.value || []).map((v, i) => {
      return {
        value: JSON.parse(v || '{"coordinates":[]}').coordinates || [],
        opacity: localValueIndex.value === undefined || i === localValueIndex.value ? 1.0 : 0.6,
      }
    }).filter(c => c && c.value && c.value.length)
  }
  return [{ value: JSON.parse(value.value || '{"coordinates":[]}').coordinates || [] }].filter(c => c && c.value && c.value.length)
})

onBeforeUnmount(() => {
  setDefaultValues()
})

function openMap (index) {
  localValueIndex.value = index
  const { value: coords } = localValue.value[index] || {}
  map.value.center = coords || props.field.options.center
  map.value.zoom = index >= 0 ? 13 : props.field.options.zoom
  map.value.show = true
}

function closeMap () {
  map.value.show = false
}

function onGeoSearchError () {}

function setDefaultValues () {
  map.value = {}
  localValueIndex.value = undefined
}
</script>
