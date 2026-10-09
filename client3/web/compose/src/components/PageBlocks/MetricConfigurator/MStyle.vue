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
                {{ balloon ? $t('metric.editStyle.balloonThresholds.label') : $t('metric.editStyle.thresholds.label') }}
                <button
                  type="button"
                  class="btn btn-link text-decoration-none ms-1"
                  @click="addThreshold()"
                >
                  {{ $t('label.add-with-plus') }}
                </button>
              </div>
            </label>
            <small class="text-muted d-block mb-2">
              {{ balloon ? $t('metric.editStyle.balloonThresholds.description') : $t('metric.editStyle.thresholds.description') }}
            </small>

            <div
              v-for="(t, i) in options.colorThresholds"
              :key="i"
              class="row align-items-center mt-2 g-2"
            >
              <div :class="balloon ? 'col-3' : 'col-4'">
                <input
                  v-model.number="t.value"
                  :placeholder="$t('metric.editStyle.thresholds.valuePlaceholder')"
                  type="number"
                  class="form-control"
                />
              </div>
              <div :class="balloon ? 'col-4' : 'col-5'" class="d-flex align-items-center gap-2">
                <c-input-color-picker
                  v-if="balloon"
                  v-model="t.variant"
                  :translations="{
                    modalTitle: $t('metric.editStyle.colorPicker'),
                    light: $t('themes.labels.light'),
                    dark: $t('themes.labels.dark'),
                    cancelBtnLabel: $t('label.cancel'),
                    saveBtnLabel: $t('label.saveAndClose')
                  }"
                  :theme-settings="themeSettings"
                  class="flex-grow-1"
                />
                <select
                  v-else
                  v-model="t.variant"
                  class="form-select form-control flex-grow-1"
                >
                  <option
                    v-for="v in variants"
                    :key="v.value"
                    :value="v.value"
                  >
                    {{ v.text }}
                  </option>
                </select>
              </div>
              <div class="col-3">
                <select
                  v-model="t.icon"
                  class="form-select form-control"
                  :title="$t('metric.editStyle.thresholds.icon')"
                >
                  <option
                    v-for="ic in iconOptions"
                    :key="ic.value"
                    :value="ic.value"
                  >
                    {{ ic.text }}
                  </option>
                </select>
              </div>
              <div class="col-auto d-flex align-items-center">
                <font-awesome-icon
                  :icon="['fas', 'times']"
                  class="pointer text-danger"
                  @click="removeThreshold(i)"
                />
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
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, inject } from 'vue'
import { components } from 'corteza-lib/vue/dist'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
const { CInputColorPicker } = components

const props = defineProps({
  options: { type: Object, required: true, default: () => ({}) },
  /** When true — thresholds drive balloon fill color via color picker */
  balloon: { type: Boolean, default: false },
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

const iconOptions = computed(() => [
  { value: '', text: t('metric.editStyle.thresholds.icons.none') },
  { value: 'arrow-up', text: t('metric.editStyle.thresholds.icons.arrowUp') },
  { value: 'arrow-down', text: t('metric.editStyle.thresholds.icons.arrowDown') },
  { value: 'arrow-right', text: t('metric.editStyle.thresholds.icons.arrowRight') },
  { value: 'alert', text: t('metric.editStyle.thresholds.icons.alert') },
  { value: 'alert-circle', text: t('metric.editStyle.thresholds.icons.alertCircle') },
])

const themeSettings = computed(() => $Settings.get('ui.studio.themes', []))

function addThreshold () {
  if (!Array.isArray(props.options.colorThresholds)) {
    props.options.colorThresholds = []
  }
  props.options.colorThresholds.push({
    value: 0,
    variant: props.balloon ? '#f64e60' : 'success',
    icon: '',
  })
}

function removeThreshold (index) {
  if (index > -1) props.options.colorThresholds.splice(index, 1)
}
</script>
