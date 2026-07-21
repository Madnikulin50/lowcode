<template>
  <div>
    <div class="mb-3">
      <label class="form-label text-primary">{{ t('kind.file.view.maxSizeLabel') }}</label>
      <div class="input-group input-group-sm">
        <input v-model="f.options.maxSize" type="number" class="form-control form-control-sm" />
      </div>
    </div>

    <div class="mb-3 mt-2">
      <label class="form-label text-primary">{{ t('kind.file.view.mimetypesLabel') }}</label>
      <div class="form-text">{{ t('kind.file.view.mimetypesFootnote') }}</div>
      <input v-model="f.options.mimetypes" type="text" class="form-control form-control-sm" />
    </div>

    <div class="mb-3 mt-2">
      <div class="form-text">{{ t('kind.file.view.webcam.enable.footnote') }}</div>
      <div class="form-check">
        <input id="enableWebcam" v-model="f.options.enableWebcam" type="checkbox" class="form-check-input" />
        <label class="form-check-label" for="enableWebcam">{{ t('kind.file.view.webcam.enable.label') }}</label>
      </div>
    </div>

    <div class="mb-3">
      <label class="form-label text-primary">{{ t('kind.file.view.modeLabel') }}</label>
      <div class="form-text">{{ t('kind.file.view.modeFootnote') }}</div>
      <div class="btn-group btn-group-sm" data-bs-toggle="buttons">
        <label
          v-for="opt in modes"
          :key="opt.value"
          class="btn btn-outline-secondary btn-sm"
          :class="{ active: f.options.mode === opt.value }"
        >
          <input
            type="radio"
            class="btn-check"
            :value="opt.value"
            :checked="f.options.mode === opt.value"
            autocomplete="off"
            @change="f.options.mode = opt.value"
          />
          {{ opt.text }}
        </label>
      </div>
    </div>

    <div class="mb-3">
      <div class="form-check" v-if="enablePreviewStyling">
        <input id="hideFileName" v-model="f.options.hideFileName" type="checkbox" class="form-check-input" />
        <label class="form-check-label" for="hideFileName">{{ t('kind.file.view.showName') }}</label>
      </div>
      <div class="form-check">
        <input id="clickToView" v-model="f.options.clickToView" type="checkbox" class="form-check-input" />
        <label class="form-check-label" for="clickToView">{{ t('kind.file.view.clickToView') }}</label>
      </div>
      <div class="form-check">
        <input id="enableDownload" v-model="f.options.enableDownload" type="checkbox" class="form-check-input" />
        <label class="form-check-label" for="enableDownload">{{ t('kind.file.view.enableDownload') }}</label>
      </div>
    </div>

    <template v-if="enablePreviewStyling">
      <hr />
      <h5 class="mb-2">{{ t('kind.file.view.previewStyle') }}</h5>
      <small>{{ t('kind.file.view.description') }}</small>

      <div class="row mb-2 mt-2">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.file.view.height') }}</label>
            <input v-model="f.options.height" type="text" class="form-control form-control-sm" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.file.view.width') }}</label>
            <input v-model="f.options.width" type="text" class="form-control form-control-sm" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.file.view.maxHeight') }}</label>
            <input v-model="f.options.maxHeight" type="text" class="form-control form-control-sm" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.file.view.maxWidth') }}</label>
            <input v-model="f.options.maxWidth" type="text" class="form-control form-control-sm" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.file.view.borderRadius') }}</label>
            <input v-model="f.options.borderRadius" type="text" class="form-control form-control-sm" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.file.view.margin') }}</label>
            <input v-model="f.options.margin" type="text" class="form-control form-control-sm" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ t('kind.file.view.background') }}</label>
            <CInputColorPicker
              v-model="f.options.backgroundColor"
              :translations="{
                modalTitle: t('kind.file.view.colorPicker'),
                light: t('themes.labels.light'),
                dark: t('themes.labels.dark'),
                cancelBtnLabel: t('label.cancel'),
                saveBtnLabel: t('label.saveAndClose')
              }"
              :theme-settings="themeSettings"
            />
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConfiguratorBase } from './base'
import { components } from 'corteza-lib/vue/dist'
const { CInputColorPicker } = components

const props = defineProps({
  namespace: { type: Object, required: true },
  module: { type: Object, required: true },
  field: { type: Object, required: true },
  hasRecords: { type: Boolean, default: false },
})

const emit = defineEmits(['update:field'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { f, isNew, hasData } = useConfiguratorBase(props, emit)

const $settings = inject('$Settings')

const modes = computed(() => [
  { value: 'list', text: t('kind.file.view.list') },
  { value: 'gallery', text: t('kind.file.view.gallery') },
])

const enablePreviewStyling = computed(() => {
  return f.value.options.mode === 'gallery'
})

const themeSettings = computed(() => {
  return ($settings.value || {}).get ? $settings.value.get('ui.studio.themes', []) : []
})
</script>
