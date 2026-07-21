<template>
  <div class="card">
    <div class="card-body">
      <slot name="title" />

      <fieldset>
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('metric.editStyle.color') }}</label>
          <c-input-color-picker
            v-model="options.color"
            :translations="{
              modalTitle: $t('metric.editStyle.colorPicker'),
              light: $t('themes.labels.light'),
              dark: $t('themes.labels.dark'),
              cancelBtnLabel: $t('label.cancel'),
              saveBtnLabel: $t('label.saveAndClose')
            }"
            :theme-settings="themeSettings"
            class="mb-1"
          />
        </div>

        <div
          v-if="labelColor"
          class="mb-3"
        >
          <label class="form-label text-primary">{{ $t('metric.editStyle.labelColor') }}</label>
          <c-input-color-picker
            v-model="options.labelColor"
            :translations="{
              modalTitle: $t('metric.editStyle.colorPicker'),
              light: $t('themes.labels.light'),
              dark: $t('themes.labels.dark'),
              cancelBtnLabel: $t('label.cancel'),
              saveBtnLabel: $t('label.saveAndClose')
            }"
            :theme-settings="themeSettings"
            class="mb-1"
          />
        </div>

        <div
          v-if="backgroundColor"
          class="mb-3"
        >
          <label class="form-label text-primary">{{ $t('metric.editStyle.backgroundColor') }}</label>
          <c-input-color-picker
            v-model="options.backgroundColor"
            :translations="{
              modalTitle: $t('geometry.recordFeed.colorPicker'),
              light: $t('themes.labels.light'),
              dark: $t('themes.labels.dark'),
              cancelBtnLabel: $t('label.cancel'),
              saveBtnLabel: $t('label.saveAndClose')
            }"
            :theme-settings="themeSettings"
            class="mb-1"
          />
        </div>

        <div
          v-if="fontSize"
          class="mb-3"
        >
          <label class="form-label text-primary">{{ $t('metric.editStyle.fontSize') }}</label>
          <input
            v-model="options.fontSize"
            type="number"
            class="form-control mb-1"
            placeholder="16"
            min="0.1"
            step="0.1"
          />
        </div>
      </fieldset>
    </div>
  </div>
</template>

<script setup>
import { computed, inject } from 'vue'
import { components } from 'corteza-lib/vue/dist'
const { CInputColorPicker } = components

const props = defineProps({
  valueColor: { type: Boolean, required: false, default: true },
  valueColorGrades: { type: Boolean, required: false, default: false },
  labelColor: { type: Boolean, required: false, default: () => true },
  backgrundColor: { type: Boolean, required: false, default: () => true },
  fontSize: { type: Boolean, required: false, default: () => true },
  options: { type: Object, required: true, default: () => ({}) },
})

const $Settings = inject('$Settings')

const themeSettings = computed(() => $Settings.get('ui.studio.themes', []))
</script>
