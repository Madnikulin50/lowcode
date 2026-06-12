<template>
  <b-card no-body>
    <slot name="title" />

    <fieldset>
      <b-form-group
        :label="$t('metric.editStyle.color')"
        label-class="text-primary"
      >
        <c-input-color-picker
          v-model="options.color"
          :translations="{
            modalTitle: $t('metric.editStyle.colorPicker'),
            light: $t('general:themes.labels.light'),
            dark: $t('general:themes.labels.dark'),
            cancelBtnLabel: $t('general:label.cancel'),
            saveBtnLabel: $t('general:label.saveAndClose')
          }"
          :theme-settings="themeSettings"
          class="mb-1"
        />
      </b-form-group>

      <b-form-group
        v-if="labelColor"
        :label="$t('metric.editStyle.labelColor')"
        label-class="text-primary"
      >
        <c-input-color-picker
          v-model="options.labelColor"
          :translations="{
            modalTitle: $t('metric.editStyle.colorPicker'),
            light: $t('general:themes.labels.light'),
            dark: $t('general:themes.labels.dark'),
            cancelBtnLabel: $t('general:label.cancel'),
            saveBtnLabel: $t('general:label.saveAndClose')
          }"
          :theme-settings="themeSettings"
          class="mb-1"
        />
      </b-form-group>

      <b-form-group
        v-if="backgroundColor"
        :label="$t('metric.editStyle.backgroundColor')"
        label-class="text-primary"
      >
        <c-input-color-picker
          v-model="options.backgroundColor"
          :translations="{
            modalTitle: $t('geometry.recordFeed.colorPicker'),
            light: $t('general:themes.labels.light'),
            dark: $t('general:themes.labels.dark'),
            cancelBtnLabel: $t('general:label.cancel'),
            saveBtnLabel: $t('general:label.saveAndClose')
          }"
          :theme-settings="themeSettings"
          class="mb-1"
        />
      </b-form-group>

      <b-form-group
        v-if="fontSize"
        :label="$t('metric.editStyle.fontSize')"
        label-class="text-primary"
      >
        <b-form-input
          v-model="options.fontSize"
          type="number"
          placeholder="16"
          min="0.1"
          step="0.1"
          class="mb-1"
        />
      </b-form-group>
    </fieldset>
  </b-card>
</template>

<script>
import { components } from 'corteza-lib/vue/dist'
const { CInputColorPicker } = components

export default {
  i18nOptions: {
    namespaces: 'block',
  },

  components: {
    CInputColorPicker,
  },

  props: {
    valueColor: {
      type: Boolean,
      required: false,
      default: true,
    },
    valueColorGrades: {
      type: Boolean,
      required: false,
      default: false,
    },
    labelColor: {
      type: Boolean,
      required: false,
      default: () => true,
    },
    backgrundColor: {
      type: Boolean,
      required: false,
      default: () => true,
    },
    fontSize: {
      type: Boolean,
      required: false,
      default: () => true,
    },
    options: {
      type: Object,
      required: true,
      default: () => ({}),
    },
  },

  computed: {
    themeSettings () {
      return this.$Settings.get('ui.studio.themes', [])
    },
  },
}
</script>
