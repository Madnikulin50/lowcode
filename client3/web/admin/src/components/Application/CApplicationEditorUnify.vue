<template>
  <div class="card shadow-sm" data-test-id="card-application-selector">
    <div class="card-header border-bottom">
      <h4 data-test-id="card-title" class="m-0">
        {{ $t('title') }}
      </h4>
    </div>

    <div class="card-body">
  <form @submit.prevent="$emit('submit', { unify, unifyAssets })">
        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.applications.editor.unify.name.label') }}</label>
              <small class="form-text text-muted">{{ $t('system.applications.editor.unify.name.description') }}</small>
              <input
                v-model="unify.name"
                type="text"
                class="form-control"
                data-test-id="input-name"
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary d-flex align-items-center gap-1">
                {{ $t('system.applications.editor.unify.logo.label') }}
                <button
                  v-if="showLogoPreview"
                  type="button"
                  class="btn btn-link d-flex align-items-center border-0 p-0 ms-2"
                  data-test-id="button-logo-show"
                  data-bs-toggle="modal"
                  data-bs-target="#logo-preview-modal"
                >
                  <font-awesome-icon :icon="['fas', 'eye']" />
                </button>
                <button
                  v-if="showLogoPreview"
                  type="button"
                  class="btn btn-outline-secondary btn-sm py-0 ms-2"
                  data-test-id="button-logo-reset"
                  @click="resetLogo()"
                >
                  {{ $t('system.applications.editor.unify.logo.reset') }}
                </button>
              </label>
              <small class="form-text text-muted">{{ $t('system.applications.editor.unify.logo.description') }}</small>
              <input
                type="file"
                class="form-control"
                accept="image/*"
                data-test-id="file-logo-upload"
                @change="onLogoChange"
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.applications.editor.unify.url.label') }}</label>
              <small class="form-text text-muted">{{ $t('system.applications.editor.unify.url.description') }}</small>
              <input
                v-model="unify.url"
                type="text"
                class="form-control"
                data-test-id="input-url"
              >
            </div>
          </div>
  
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.applications.editor.unify.listed') }}</label>
              <c-input-checkbox
                v-model="unify.listed"
                data-test-id="checkbox-listed"
                switch
                :labels="checkboxLabel"
              />
            </div>
          </div>
  
          <div v-if="canPin" class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.applications.editor.unify.pinned') }}</label>
              <c-input-checkbox
                v-model="unify.pinned"
                data-test-id="checkbox-pinned"
                switch
                :labels="checkboxLabel"
              />
            </div>
          </div>
  
          <div class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('system.applications.editor.unify.config.label') }}</label>
              <small class="form-text text-muted">{{ $t('system.applications.editor.unify.config.description') }}</small>
              <textarea
                v-model="unify.config"
                class="form-control"
                :class="{ 'is-invalid': configState === false }"
                data-test-id="textarea-config"
                rows="10"
              />
            </div>
          </div>
        </div>
      </form>
  </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        :disabled="disabled"
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit', { unify, unifyAssets })"
      />
    </div>
  </div>

  <div
    id="logo-preview-modal"
    class="modal fade"
    tabindex="-1"
  >
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-body p-1">
          <img
            data-test-id="img-logo-preview"
            :src="unify.logo"
            class="img-fluid"
          >
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'

const { t } = useI18n()

const props = defineProps({
  unify: {
    type: Object,
    required: true,
  },
  application: {
    type: Object,
    required: true,
  },
  canPin: {
    type: Boolean,
    required: true,
  },
  processing: {
    type: Boolean,
    value: false,
  },
  success: {
    type: Boolean,
    value: false,
  },
})

defineEmits(['submit', 'change-detected'])

const unifyAssets = ref({
  icon: undefined,
  logo: undefined,
})

const checkboxLabel = {
  on: t('label.general.yes'),
  off: t('label.general.no'),
}

const disabled = computed(() => {
  return validConfig.value === false
})

const validConfig = computed(() => {
  if (!props.unify) {
    return null
  }

  try {
    if ((props.unify.config || '').trim() !== '') {
      JSON.parse(props.unify.config)
    }
    return null
  } catch (e) {
    return false
  }
})

const configState = computed(() => {
  if (((props.unify || {}).config || '').trim() === '') {
    return null
  } else {
    return validConfig.value
  }
})

const showLogoPreview = computed(() => {
  return props.unify.logoID !== NoID
})

onMounted(() => {
  props.unify.name = props.unify.name ? props.unify.name : props.application.name
})

function resetLogo() {
  props.unify.logo = undefined
  props.unify.logoID = NoID
}

function onLogoChange(event) {
  const file = event.target.files[0]
  if (file) {
    unifyAssets.value.logo = file
  }
}
</script>