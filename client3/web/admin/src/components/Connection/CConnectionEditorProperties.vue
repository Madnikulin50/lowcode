<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">{{ $t('title') }}</h4>
    </div>

    <div v-for="prop in list" :key="prop" class="row">
      <div class="col-12">
        <div class="form-check mb-1">
          <input
            :id="'prop-' + prop"
            v-model="properties[prop].enabled"
            type="checkbox"
            class="form-check-input-v3"
          >
          <label :for="'prop-' + prop" class="form-check-label">
            {{ $t('form.' + kebabCase(prop) + '.checkbox.label') }}
          </label>
        </div>
        <div class="ms-4 mb-3">
          <label class="form-label text-primary">{{ $t('form.' + kebabCase(prop) + '.notes.label') }}</label>
          <small class="form-text text-muted">{{ $t('form.' + kebabCase(prop) + '.notes.description') }}</small>
          <textarea v-model="properties[prop].notes" class="form-control" />
        </div>
      </div>
    </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        :disabled="disabled"
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
import { kebabCase } from 'lodash'

defineProps({
  properties: { type: Object, required: true },
  disabled: { type: Boolean, default: false },
  processing: { type: Boolean, value: false },
  success: { type: Boolean, value: false },
})

defineEmits(['submit'])

const list = [
  'dataAtRestEncryption',
  'dataAtRestProtection',
  'dataAtTransitEncryption',
  'dataRestoration',
]
</script>
