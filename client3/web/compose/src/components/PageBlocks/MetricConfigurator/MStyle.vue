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

        <div class="col-12">
          <div class="mb-3">
            <label class="d-flex align-items-center text-primary form-label">
              <div class="d-flex align-items-center">
                {{ $t('metric.editStyle.thresholds.label') }}
                <button
                  class="btn btn-link text-decoration-none ms-1"
                  @click="addThreshold()"
                >
                  {{ $t('label.add-with-plus') }}
                </button>
              </div>
              <small class="text-muted d-block">
                {{ $t('metric.editStyle.thresholds.description') }}
              </small>
            </label>

            <div class="row">
              <div
                v-for="(t, i) in options.colorThresholds"
                :key="i"
                class="row align-items-center mt-2"
              >
                <div class="col">
                  <input
                    v-model="t.value"
                    placeholder="Threshold"
                    type="number"
                    class="form-control"
                  />
                </div>
                <div class="col d-flex align-items-center justify-content-center">
                  <select
                    v-model="t.variant"
                    class="form-select form-control"
                  >
                    <option
                      v-for="v in variants"
                      :key="v.value"
                      :value="v.value"
                    >
                      {{ v.text }}
                    </option>
                  </select>
                  <font-awesome-icon
                    :icon="['fas', 'times']"
                    class="pointer text-danger ms-3"
                    @click="removeThreshold(i)"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="mb-3">
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

        <div class="mb-3">
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

        <div class="mb-3">
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

        <div class="mb-3">
          <div class="form-check">
            <input
              v-model="options.notFitVertical"
              type="checkbox"
              class="form-check-input"
            />
            <label class="form-check-label">{{ $t('metric.editStyle.notFitVertical') }}</label>
          </div>
        </div>

        <div class="mb-3">
          <div class="form-check">
            <input
              v-model="options.notFitHorizontal"
              type="checkbox"
              class="form-check-input"
            />
            <label class="form-check-label">{{ $t('metric.editStyle.notFitHorizontal') }}</label>
          </div>
        </div>
      </fieldset>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, inject } from 'vue'
import { components } from 'corteza-lib/vue/dist'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
const { CInputColorPicker } = components

const props = defineProps({
  options: { type: Object, required: true, default: () => ({}) },
})

const $Settings = inject('$Settings')

const variants = ref([
  { text: t('variants.primary'), value: 'primary' },
  { text: t('variants.secondary'), value: 'secondary' },
  { text: t('variants.success'), value: 'success' },
  { text: t('variants.warning'), value: 'warning' },
  { text: t('variants.danger'), value: 'danger' },
  { text: t('variants.info'), value: 'info' },
  { text: t('variants.light'), value: 'light' },
  { text: t('variants.dark'), value: 'dark' },
])

const themeSettings = computed(() => $Settings.get('ui.studio.themes', []))

function addThreshold () {
  props.options.colorThresholds.push({ value: 0, variant: 'success' })
}

function removeThreshold (index) {
  if (index > -1) props.options.colorThresholds.splice(index, 1)
}
</script>
