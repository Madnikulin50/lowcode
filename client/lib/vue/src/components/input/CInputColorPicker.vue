<template>
  <div>
    <div class="d-flex align-items-center">
      <b-button
        :style="`color: ${viewColor}; fill: ${viewColor};`"
        class="p-0 rounded-circle bg-white border-white shadow-none"
        @click="toggleMenu"
      >
        <svg
          viewBox="0 0 32 32"
          :style="{ width: width, height: height }"
          class="border border-light rounded-circle"
        >
          <pattern
            id="checkerboard"
            width="12"
            height="12"
            patternUnits="userSpaceOnUse"
            fill="FFF"
          >
            <rect
              fill="#7080707f"
              x="0"
              width="6"
              height="6"
              y="0"
            />
            <rect
              fill="#7080707f"
              x="6"
              width="6"
              height="6"
              y="6"
            />
          </pattern>

          <circle
            cx="16"
            cy="16"
            r="16"
            fill="url(#checkerboard)"
          />

          <circle
            cx="16"
            cy="16"
            r="16"
          />
        </svg>
      </b-button>
      <span
        v-if="showText"
        class="ml-2"
      >
        {{ value }}
      </span>
    </div>

    <b-modal
      :visible="showModal"
      :title="translations.modalTitle"
      centered
      size="md"
      :hide-header="!Boolean(translations.modalTitle)"
      body-class="p-0"
      no-fade
      @hide="closeMenu"
    >
      <chrome
        :value="currentColor"
        class="w-100 shadow-none"
        @input="updateColor"
      />

      <div
        v-if="themes.length > 0"
        class="d-flex flex-column border-top p-3 gap-2"
      >
        <b-form-group
          v-for="theme in themes"
          :key="theme.id"
          :label="translations[theme.id]"
          label-class="text-primary"
          class="mb-0"
        >
          <div
            class="d-flex flex-wrap border"
          >
            <b-button
              v-for="variable in themeVariables"
              :key="variable.key"
              v-b-tooltip.noninteractive.hover="{ title: variable.label, boundary: 'body' }"
              class="swatch flex-grow-1 rounded-0"
              :style="{ backgroundColor: theme.values[variable.key], borderColor: theme.values[variable.key] }"
              @click="setColor(theme.values[variable.key], variable.key)"
            />
          </div>
        </b-form-group>
      </div>

      <template #modal-footer>
        <b-button
          v-if="defaultValue"
          variant="outline-secondary"
          @click="setColor()"
        >
          {{ translations.defaultBtnLabel || 'Default' }}
        </b-button>
        <slot name="footer" />

        <b-button
          variant="outline-light"
          class="ml-auto text-primary border-0"
          @click="closeMenu"
        >
          {{ translations.cancelBtnLabel || 'Cancel' }}
        </b-button>

        <b-button
          variant="primary"
          @click="saveColor"
        >
          {{ translations.saveBtnLabel || 'Save' }}
        </b-button>
      </template>
    </b-modal>
  </div>
</template>

<script>
import {Chrome} from 'vue-color'
import {debounce} from 'lodash'

export default {
  name: 'CInputColorPicker',

  components: {
    Chrome,
  },

  props: {
    value: {
      type: String,
      default: 'black',
    },

    defaultValue: {
      type: String,
      default: '',
    },

    translations: {
      type: Object,
      default: () => ({}),
    },

    width: {
      type: String,
      default: '32px',
    },

    height: {
      type: String,
      default: '32px',
    },

    showText: {
      type: Boolean,
      default: true,
    },

    themeSettings: {
      type: Array,
      default: () => [],
    },

    themeVariables: {
      type: Array,
      default: () => [
        {
          key: 'white',
          label: 'White',
        },
        {
          key: 'black',
          label: 'Black',
        },
        {
          key: 'primary',
          label: 'Primary',
        },
        {
          key: 'secondary',
          label: 'Secondary',
        },
        {
          key: 'success',
          label: 'Success',
        },
        {
          key: 'warning',
          label: 'Warning',
        },
        {
          key: 'danger',
          label: 'Danger',
        },
        {
          key: 'light',
          label: 'Light',
        },
        {
          key: 'extra-light',
          label: 'Extra light',
        },
      ],
    },
  },

  data () {
    return {
      showModal: false,
      currentColor: '',
      colorValue: ''
    }
  },

  computed: {
    viewColor () {
      return this.colorFromValue(this.value)
    },
    themes () {
      return this.themeSettings
        .filter((theme) => theme.id !== 'general') // remove general theme
        .map((theme) => {
          return {
            id: theme.id,
            values: JSON.parse(theme.values),
          }
        })
    },
  },



  watch: {
    value: {
      immediate: true,
      handler: function (value) {
        this.currentValue = value

        this.currentColor = this.colorFromValue(value)

        if (!value && this.defaultValue) {
          this.$emit('input', this.defaultValue)
        }
      },
    },
  },

  methods: {
    colorFromValue (value) {
      if (value[0] === "#") {
        return value
      }
      return this.themes[0].values[value] || value
    },

    updateColor: debounce(function ({ hex8 = '' }) {
      if (this.currentColor !== hex8) {
        this.currentColor = hex8
        this.currentValue = hex8
      }
    }, 300),

    setColor (color = this.defaultValue, key) {
      if (key === undefined) {
        key = color
      }
      this.currentColor = color
      this.currentValue = key
    },

    saveColor () {
       this.$emit('input', this.currentValue)
      this.closeMenu()
    },

    toggleMenu () {
      if (this.showModal) {
        this.closeMenu()
      } else {
        this.openMenu()
      }
    },

    openMenu () {
      this.showModal = true
    },

    closeMenu () {
      this.showModal = false
    },
  },
}
</script>

<style lang="scss" scoped>
.swatch {
  height: 58px;
  min-width: 50px;
}
</style>

<style lang="scss">
.vc-chrome {
  font-family: var(--font-medium) !important;

  .vc-chrome-body {
    background: var(--white) !important;

    .vc-input__input {
      color: var(--black) !important;
      background-color: var(--white) !important;
    }

    .vc-input__label {
      color: var(--black) !important;
    }

    .vc-chrome-toggle-btn {
      path {
        fill: var(--black) !important;
      }

      .vc-chrome-toggle-icon-highlight {
        background: var(--light) !important;
      }
    }
  }
}
</style>
