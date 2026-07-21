<template>
  <div class="card shadow-sm" data-test-id="card-connection-settings">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div class="row">
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('form.name.label') }}</label>
          <small class="form-text text-muted">{{ $t('form.name.description') }}</small>
          <input
            v-model="connection.meta.name"
            type="text"
            class="form-control"
            :class="{ 'is-invalid': nameState === false }"
            required
            :placeholder="$t('form.name.placeholder')"
          >
        </div>
      </div>

      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('form.handle.label') }}</label>
          <input
            v-model="connection.handle"
            type="text"
            class="form-control"
            :class="{ 'is-invalid': handleState === false }"
            :disabled="isPrimary || disabled"
            :placeholder="$t('form.handle.placeholder')"
          >
          <div v-if="handleState === false" class="invalid-feedback">{{ $t('form.handle.invalid-characters') }}</div>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('form.location-name.label') }}</label>
          <small class="form-text text-muted">{{ $t('form.location-name.description') }}</small>
          <input
            v-model="connection.meta.location.properties.name"
            type="text"
            class="form-control"
            :placeholder="$t('form.location-name.placeholder')"
          >
        </div>
      </div>

      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('form.location-geometry.label') }}</label>
          <small class="form-text text-muted">{{ $t('form.location-geometry.description') }}</small>
          <c-location
            v-if="!disabled"
            v-model="locationCoordinates"
            :label="locationCoordinatesLabel"
            editable
          />
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('form.ownership.label') }}</label>
          <small class="form-text text-muted">{{ $t('form.ownership.description') }}</small>
          <input
            v-model="connection.meta.ownership"
            type="text"
            class="form-control"
            :placeholder="$t('form.ownership.placeholder')"
          >
        </div>
      </div>

      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('form.sensitivity-level.label') }}</label>
          <small class="form-text text-muted">{{ $t('form.sensitivity-level.description') }}</small>
          <c-sensitivity-level-picker
            v-model="connection.config.privacy.sensitivityLevelID"
            :options="sensitivityLevels"
            :placeholder="$t('form.sensitivity-level.placeholder')"
          />
        </div>
      </div>
    </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-input-confirm
        v-if="!fresh && !isPrimary && !disabled"
        :text="connection.deletedAt ? $t('label.undelete') : $t('label.delete')"
        variant="danger"
        size="md"
        @confirmed="$emit('delete')"
      />

      <c-button-submit
        :disabled="disabled || saveDisabled"
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit')"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { NoID } from 'corteza-lib/js/dist'
import { components, handle } from 'corteza-lib/vue/dist'
import CLocation from 'corteza-webapp-admin/src/components/CLocation'

const { CSensitivityLevelPicker } = components

const props = defineProps({
  connection: { type: Object, required: true },
  sensitivityLevels: { type: Array, required: true },
  disabled: { type: Boolean, default: false },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
  canCreate: { type: Boolean, required: true },
})

defineEmits(['delete', 'submit'])

const isPrimary = computed(() => {
  return props.connection.type === 'corteza::system:primary-dal-connection'
})

const fresh = computed(() => {
  return !props.connection.connectionID || props.connection.connectionID === NoID
})

const editable = computed(() => {
  return fresh.value ? props.canCreate : true
})

const nameState = computed(() => {
  return props.connection.meta.name ? null : false
})

const handleState = computed(() => {
  return handle.handleState(props.connection.handle)
})

const saveDisabled = computed(() => {
  return !editable.value || [nameState.value, handleState.value].includes(false)
})

const locationCoordinates = computed({
  get() {
    const { coordinates = [] } = (props.connection.meta.location || {}).geometry || {}
    return coordinates || []
  },
  set(value) {
    if (!props.connection.meta.location) {
      props.connection.meta.location = { type: 'Feature', geometry: { type: 'Point', coordinates: [] }, properties: { name: '' } }
    }
    if (!props.connection.meta.location.geometry) {
      props.connection.meta.location.geometry = { type: 'Point', coordinates: [] }
    }
    props.connection.meta.location.geometry.coordinates = value
  },
})

const locationCoordinatesLabel = computed(() => {
  return (locationCoordinates.value || []).map(c => c?.toFixed?.(7) || c).join(', ')
})
</script>