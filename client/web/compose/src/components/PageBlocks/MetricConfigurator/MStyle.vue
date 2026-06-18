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

      <b-col
        cols="12"
      >
        <b-form-group>
          <template #label>
            <div
              class="d-flex align-items-center text-primary"
            >
              {{ $t('metric.editStyle.thresholds.label') }}
              <b-button
                variant="link"
                class="text-decoration-none ml-1"
                @click="addThreshold()"
              >
                {{ $t('general:label.add-with-plus') }}
              </b-button>
            </div>

            <small
              class="text-muted"
            >
              {{ $t('metric.editStyle.thresholds.description') }}
            </small>
          </template>

          <b-row
            v-for="(t, i) in options.colorThresholds"
            :key="i"
            align-v="center"
            :class="{ 'mt-2': i }"
          >
            <b-col>
              <b-form-input
                v-model="t.value"
                :placeholder="'Threshold'"
                type="number"
                number
              />
            </b-col>

            <b-col
              class="d-flex align-items-center justify-content-center"
            >
              <b-form-select
                v-model="t.variant"
                :options="variants"
              />

              <font-awesome-icon
                :icon="['fas', 'times']"
                class="pointer text-danger ml-3"
                @click="removeThreshold(i)"
              />
            </b-col>
          </b-row>
        </b-form-group>
      </b-col>

      <b-form-group
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

      <b-form-group>
        <b-form-checkbox
          v-model="options.notFitVertical"
        >
          {{ $t('metric.editStyle.notFitVertical') }}
        </b-form-checkbox>
      </b-form-group>
      <b-form-group>
        <b-form-checkbox
          v-model="options.notFitHorizontal"
        >
          {{ $t('metric.editStyle.notFitHorizontal') }}
        </b-form-checkbox>
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
    options: {
      type: Object,
      required: true,
      default: () => ({}),
    },
  },
  data () {
    return {
      variants: [
        { text: this.$t('general:variants.primary'), value: 'primary' },
        { text: this.$t('general:variants.secondary'), value: 'secondary' },
        { text: this.$t('general:variants.success'), value: 'success' },
        { text: this.$t('general:variants.warning'), value: 'warning' },
        { text: this.$t('general:variants.danger'), value: 'danger' },
        { text: this.$t('general:variants.info'), value: 'info' },
        { text: this.$t('general:variants.light'), value: 'light' },
        { text: this.$t('general:variants.dark'), value: 'dark' },
      ],
    }
  },

  computed: {
    themeSettings () {
      return this.$Settings.get('ui.studio.themes', [])
    },
  },

  methods: {
    addThreshold () {
      this.options.colorThresholds.push({ value: 0, variant: 'success' })
    },

    removeThreshold (index) {
      if (index > -1) {
        this.options.colorThresholds.splice(index, 1)
      }
    },
  },
}
</script>
