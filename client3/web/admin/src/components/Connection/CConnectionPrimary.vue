<template>
  <div class="card shadow-sm" data-test-id="card-primary-database">
    <div class="card-header border-bottom">
      <h4 class="d-flex align-items-center gap-1 mb-0">
        {{ $t('title') }}
        <router-link
          v-if="connection"
          class="btn btn-outline-light d-flex align-items-center text-primary border-0 p-1"
          data-test-id="button-edit"
          :to="{ name: 'system.connection.edit', params: { connectionID: (connection || {}).connectionID } }"
        >
          <font-awesome-icon :icon="['far', 'edit']" />
        </router-link>
      </h4>
    </div>

    <div v-if="connection">
      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('name') }}</label>
            <p class="form-control-plaintext">{{ (connection.meta || {}).name }}</p>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('handle') }}</label>
            <p class="form-control-plaintext">{{ connection.handle }}</p>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('location') }}</label>
            <c-location
              :value="locationCoordinates"
              :label="locationName || locationCoordinatesLabel"
            />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('ownership') }}</label>
            <p class="form-control-plaintext">{{ (connection.meta || {}).ownership }}</p>
          </div>
        </div>
      </div>

      <div class="row">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('sensitivity-level') }}</label>
            <p class="form-control-plaintext">{{ sensitivityLevelName }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import CLocation from 'corteza-webapp-admin/src/components/CLocation'
import { NoID } from 'corteza-lib/js/dist'

const loading = ref(false)
const connection = ref(undefined)
const sensitivityLevel = ref(undefined)

const locationCoordinates = computed(() => {
  const { coordinates = [] } = ((connection.value || {}).meta || {}).location?.geometry || {}
  return coordinates || []
})

const locationCoordinatesLabel = computed(() => {
  return (locationCoordinates.value || []).map(c => c?.toFixed?.(7) || c).join(', ')
})

const locationName = computed(() => {
  return ((connection.value || {}).meta || {}).location?.properties?.name
})

const sensitivityLevelName = computed(() => {
  const { sensitivityLevelID, handle, meta = {} } = sensitivityLevel.value || {}
  return meta.name || handle || sensitivityLevelID || 'N/A'
})

onMounted(() => {
  fetchPrimaryConnection()
})

function fetchPrimaryConnection() {
  loading.value = true

  return window.__SystemAPI.dalConnectionList({ type: 'corteza::system:primary-dal-connection' })
    .then(({ set = [] }) => {
      connection.value = set.find(({ type }) => type === 'corteza::system:primary-dal-connection')

      if (!connection.value) {
        return
      }

      const { sensitivityLevelID } = (connection.value.config || {}).privacy || {}

      if (sensitivityLevelID && sensitivityLevelID !== NoID) {
        return window.__SystemAPI.dalSensitivityLevelRead({ sensitivityLevelID })
          .then(sl => {
            sensitivityLevel.value = sl
          })
      }
    })
    .catch(() => {})
    .finally(() => {
      loading.value = false
    })
}
</script>