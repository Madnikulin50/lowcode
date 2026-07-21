<template>
  <div>
    <div class="mb-3">
      <label class="text-primary form-label">{{ t('label-column') }}</label>
      <ColumnSelector v-model="options.valueColumn" :columns="valueColumns" style="min-width:100% !important" />
    </div>
    <div class="row">
      <div class="col">
        <div class="mb-3">
          <label class="text-primary form-label">{{ t('format') }}</label>
          <input v-model="options.format" class="form-control" placeholder="0.00" />
        </div>
      </div>
      <div class="col">
        <div class="mb-3">
          <label class="text-primary form-label">{{ t('prefix') }}</label>
          <input v-model="options.prefix" class="form-control" placeholder="$" />
        </div>
      </div>
      <div class="col">
        <div class="mb-3">
          <label class="text-primary form-label">{{ t('suffix') }}</label>
          <input v-model="options.suffix" class="form-control" placeholder="USD/mo" />
        </div>
      </div>
    </div>
    <div class="row">
      <div class="col">
        <div class="mb-3">
          <label class="text-primary form-label">{{ t('color.text') }}</label>
          <c-input-color-picker
            v-model="options.color"
            :translations="{
              modalTitle: t('color.picker'),
              light: t('themes.labels.light'),
              dark: t('themes.labels.dark'),
              cancelBtnLabel: t('label.cancel'),
              saveBtnLabel: t('label.saveAndClose'),
            }"
            :theme-settings="themeSettings"
          />
        </div>
      </div>
      <div class="col">
        <div class="mb-3">
          <label class="text-primary form-label">{{ t('color.background') }}</label>
          <c-input-color-picker
            v-model="options.backgroundColor"
            :translations="{
              modalTitle: t('color.picker'),
              light: t('themes.labels.light'),
              dark: t('themes.labels.dark'),
              cancelBtnLabel: t('label.cancel'),
              saveBtnLabel: t('label.saveAndClose'),
            }"
            :theme-settings="themeSettings"
          />
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSettings } from 'corteza-lib/vue/dist'
import ColumnSelector from '../../../../Common/ColumnSelector.vue'
import { components } from 'corteza-lib/vue/dist'
const { CInputColorPicker } = components

const props = defineProps({
  displayElementOptions: { type: Object, default: () => ({}) },
  columns: { type: Array, required: true },
  datasource: { type: Object, default: undefined },
})
const emit = defineEmits(['update:displayElementOptions'])

const { t } = useI18n()
const { settings } = useSettings()

const options = computed({
  get: () => props.displayElementOptions || {},
  set: (val) => emit('update.displayElementOptions', val),
})

const valueColumns = computed(() => {
  const cols = props.columns.length ? props.columns[0] : []
  return [...cols.filter(({ kind }) => ['Number'].includes(kind))].sort((a, b) => a.label.localeCompare(b.label))
})

const themeSettings = computed(() => settings?.get('ui.studio.themes', []))
</script>