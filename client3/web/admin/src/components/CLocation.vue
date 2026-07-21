<template>
  <div>
    <button
      v-if="label || props.editable"
      :class="label ? 'btn btn-link p-0 border-0' : 'btn btn-primary'"
      @click="openMap"
    >
      <span v-if="label">
        {{ label }}
      </span>

      <span v-else>
        <font-awesome-icon
          :icon="['fas', 'map-marked-alt']"
        />
        {{ $t('openMap') }}
      </span>
    </button>

    <div
      v-if="map.show"
      class="modal fade show d-block"
      tabindex="-1"
      @click.self="map.show = false"
    >
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-body p-0">
            <c-map
              :map="map"
              :markers="[{ value: props.value }]"
              style="height: 75vh; width: 100%;"
              @on-marker-click="removeMarker"
              @on-map-click="placeMarker"
            />
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="map.show"
      class="modal-backdrop fade show"
      @click="map.show = false"
    />
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { components } from 'corteza-lib/vue/dist'
const { CMap } = components

const { t } = useI18n()

const props = defineProps({
  value: {
    type: Array,
    required: true,
  },
  editable: {
    type: Boolean,
    default: false,
  },
  label: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['input'])

const map = reactive({
  show: false,
  zoom: 3,
  center: [30, 30],
  rotation: 0,
  attribution: '&copy; <a target="_blank" rel="noopener noreferrer" href="http://osm.org/copyright">OpenStreetMap</a>',
})

function openMap() {
  map.show = true
}

function placeMarker({ latlng = {} }) {
  const { lat = 0, lng = 0 } = latlng
  emit('input', [lat, lng])
}

function removeMarker() {
  emit('input', [])
}
</script>

<style lang="scss">
</style>
