<template>
  <Wrap
    v-bind="$props"
    @refreshBlock="refresh"
  >
    <div
      v-if="isProcessing"
      class="d-flex align-items-center justify-content-center h-100"
    >
      <span class="spinner-border spinner-border-sm" />
    </div>

    <div
      v-else
      class="w-100 h-100"
    >
      <c-map
        v-if="map"
        :map="{ ...map, maxBounds: map.bounds }"
        :labels="{ tooltip: { 'goToCurrentLocation': $t('geometry.tooltip.goToCurrentLocation') } }"
        :markers="localValue"
        :disabled="editable"
        :hide-geo-search="options.hideGeoSearch"
        :polygons="geometries"
        class="w-100 h-100"
        @on-marker-click="onMarkerCLick"
        @on-geosearch-error="onGeoSearchError"
      />
    </div>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch, onMounted, onBeforeUnmount, inject } from 'vue'
import { compose, NoID } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import axios from 'axios'
import { evaluatePrefilter, isFieldInFilter } from 'corteza-webapp-compose/src/lib/record-filter'
import { isNumber } from 'lodash'
import { useStore } from '../../store'
import { usePageBlockBase } from './usePageBlockBase'
import Wrap from './Wrap/index.js'
import { useRouter } from 'vue-router'

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
const store = useStore()
const router = useRouter()
const $auth = inject('$auth')
const $ComposeAPI = inject('$ComposeAPI')

const { options, isProcessing, processing, inModal, refreshBlock, setBaseDefaultValues, getColor } = usePageBlockBase(props, emit)

const map = ref(undefined)
const show = ref(false)
const geometries = ref([])
const colors = ref([])
const markers = ref([])
const cancelTokenSource = axios.CancelToken.source()
const getModuleByID = computed(() => store.module.getByID)
const pages = computed(() => store.page.set)

const localValue = computed(() => {
  const values = []
  geometries.value.forEach((geo) => {
    geo.filter(({ displayMarker }) => displayMarker).forEach(value => {
      value.markers.forEach(subValue => {
        if (subValue) values.push({ title: value.title, value: subValue || {}, color: getColor(value.color), recordID: value.recordID, moduleID: value.moduleID })
      })
    })
  })
  return values
})

watch(() => props.record?.recordID, () => { loadEvents() }, { immediate: true })
watch(() => options.value, () => { loadEvents() }, { deep: true })

onMounted(() => {
  const bounds = options.value.bounds
  refreshBlock(refresh)
  createEvents()
})

onBeforeUnmount(() => {
  setDefaultValues()
  abortRequests()
  destroyEvents()
})

function createEvents () {
  window.addEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.addEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.addEventListener('refetch-records', refresh)
}

function refetchOnPrefilterValueChange ({ fieldName }) {
  const { feeds } = options.value
  if (feeds.some(({ options: o }) => isFieldInFilter(fieldName, o.prefilter))) refresh()
}

function refreshOnRelatedRecordsUpdate ({ moduleID } = {}) {
  const hasMatchingModule = options.value.feeds.some(({ options: o = {} }) => o.moduleID === moduleID)
  if (hasMatchingModule) refresh()
}

function loadEvents () {
  geometries.value = []
  processing.value = true
  colors.value = options.value.feeds.map(feed => feed.options.color)
  const { bounds, center, zoomStarting, zoomMin, zoomMax } = options.value
  map.value = { bounds, center, zoom: zoomStarting, zoomMin, zoomMax }
  Promise.all(options.value.feeds.filter(f => f.isValid()).map((feed, idx) => {
    return store.module.findByID({ namespace: props.namespace, moduleID: feed.options.moduleID }).then(module => {
      const f = compose.PageBlockGeometry.makeFeed(feed)
      if (f.options.prefilter) {
        f.options.prefilter = evaluatePrefilter(f.options.prefilter, {
          record: props.record, user: $auth.user || {}, recordID: (props.record || {}).recordID || NoID,
          ownerID: (props.record || {}).ownedBy || NoID, userID: ($auth.user || {}).userID || NoID,
        })
      }
      return compose.PageBlockGeometry.RecordFeed($ComposeAPI, module, props.namespace, f, { cancelToken: cancelTokenSource.token })
        .then(records => {
          const mapModuleField = module.fields.find(field => field.name === f.geometryField)
          if (!mapModuleField) return
          geometries.value[idx] = records.map(record => {
            let geometry = record.values[f.geometryField]
            let mkrs = []
            if (mapModuleField.isMulti) {
              geometry = geometry.map(value => parseGeometryField(value))
              mkrs = geometry
            } else {
              geometry = parseGeometryField(geometry)
              mkrs = [geometry]
            }
            return { title: record.values[f.titleField], geometry: f.displayPolygon ? geometry : [], markers: mkrs, color: f.options.color, displayMarker: f.displayMarker, recordID: record.recordID, moduleID: record.moduleID }
          }).filter(g => g && g.markers.length)
        })
    })
  })).catch(error => {
    if (axios.isCancel(error)) return
    console.error('Geometry records load failed:', error)
  }).finally(() => { setTimeout(() => { processing.value = false }, 300) })
}

function parseGeometryField (value) {
  value = JSON.parse(value || '{"coordinates":[]}').coordinates || []
  return value.every(isNumber) ? value : []
}

function onMarkerCLick ({ marker: { recordID, moduleID } }) {
  const page = pages.value.find(p => p.moduleID === moduleID)
  if (!page) return
  const route = { name: 'page.record', params: { recordID, pageID: page.pageID } }
  if (options.value.displayOption === 'modal' || inModal.value) {
    window.dispatchEvent(new CustomEvent('show-record-modal', { detail: { recordID, recordPageID: page.pageID } }))
  } else if (options.value.displayOption === 'newTab') {
    window.open(router.resolve(route).href)
  } else {
    router.push(route)
  }
}

function refresh () { loadEvents() }
function onGeoSearchError () { console.error('Geo search error') }

function abortRequests () { cancelTokenSource.cancel(`abort-request-${props.block.blockID}`) }
function destroyEvents () {
  window.removeEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.removeEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.removeEventListener('refetch-records', refresh)
}
function setDefaultValues () {}
</script>

<style>
.leaflet-touch .leaflet-control-attribution, .leaflet-touch .leaflet-control-layers, .leaflet-touch .leaflet-bar {
  box-shadow: none;
  height: 0px;
}
</style>
