<template>
  <div class="card shadow-sm">
    <div class="card-header border-bottom">
      <h4 class="m-0">
        {{ t('ui.settings.editor.title') }}
      </h4>
    </div>

    <div class="card-body p-0">
      <div
        v-if="!sassInstalled"
        class="bg-warning rounded p-2 mb-3"
      >
        {{ t('ui.settings.editor.corteza-studio.sassNotInstalled') }}
        <a
          v-if="installSassDocs"
          :href="installSassDocs"
          target="_blank"
          class="text-dark"
        >
          {{ t('ui.settings.editor.corteza-studio.installSassDocs') }}
        </a>
      </div>

      <ul class="nav nav-tabs card-header-tabs" data-test-id="theme-tabs" role="tablist">
        <li
          v-for="theme in themes"
          :key="theme.id"
          class="nav-item"
          role="presentation"
        >
          <button
            class="nav-link"
            :class="{ active: theme.id === activeTab }"
            type="button"
            role="tab"
            @click="activeTab = theme.id"
          >
            {{ theme.title }}
          </button>
        </li>
      </ul>

      <div class="tab-content p-3">
        <div
          v-for="theme in themes"
          :key="theme.id"
          class="tab-pane"
          :class="{ show: theme.id === activeTab, active: theme.id === activeTab }"
          role="tabpanel"
        >
          <div
            v-if="theme.id !== 'general'"
            class="row"
          >
            <div
              v-for="key in themeVariables"
              :key="key"
              class="col-12 col-lg-6"
            >
              <div class="mb-3">
                <label class="form-label text-primary">{{ t(`theme.variables.${key}.label`) }}</label>
                <div class="form-text mb-2">{{ t(`theme.variables.${key}.description`) }}</div>
                <c-input-color-picker
                  ref="picker"
                  v-model="theme.variables[key]"
                  :default-value="theme.defaultVariables[key]"
                  :data-test-id="`input-${key}-color`"
                  :translations="{
                    modalTitle: t('ui.settings.editor.corteza-studio.colorPicker'),
                    defaultBtnLabel: t('ui.settings.editor.corteza-studio.label.default'),
                    light: t('ui.settings.editor.corteza-studio.tabs.light'),
                    dark: t('ui.settings.editor.corteza-studio.tabs.dark'),
                    cancelBtnLabel: t('admin.general.label.cancel'),
                    saveBtnLabel: t('admin.general.label.saveAndClose')
                  }"
                  :theme-settings="settings['ui.studio.themes']"
                />
              </div>
            </div>
          </div>

          <div
            v-else
            class="row"
          >
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="d-flex align-items-center form-label text-primary">
                  {{ t('ui.settings.editor.corteza-studio.mainLogo.title') }}

                  <c-input-confirm
                    v-if="uploadedFile('ui.main-logo')"
                    show-icon
                    class="ms-auto"
                    @confirmed="resetAttachment('ui.main-logo')"
                  />
                </label>

                <c-uploader-with-preview
                  :value="uploadedFile('ui.main-logo')"
                  endpoint="/settings/ui.main-logo"
                  :disabled="!canManage"
                  @upload="onUpload"
                />
              </div>
            </div>

            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="d-flex align-items-center form-label text-primary h-lg-100">
                  {{ t('ui.settings.editor.corteza-studio.iconLogo.title') }}

                  <c-input-confirm
                    v-if="uploadedFile('ui.icon-logo')"
                    show-icon
                    class="ms-auto"
                    @confirmed="resetAttachment('ui.icon-logo')"
                  />
                </label>

                <c-uploader-with-preview
                  :value="uploadedFile('ui.icon-logo')"
                  endpoint="/settings/ui.icon-logo"
                  :disabled="!canManage"
                  @upload="onUpload"
                />
              </div>
            </div>
          </div>

          <div class="row">
            <div class="col">
              <div class="mb-0">
                <label class="form-label text-primary">{{ t('ui.settings.editor.corteza-studio.custom-css') }}</label>
                <c-ace-editor
                  v-model="theme.customCSS"
                  auto-complete
                  lang="css"
                  min-height="400px"
                  show-line-numbers
                  :show-popout="true"
                  :auto-complete-suggestions="customCssAutocompleteVal"
                  @open="openCustomCSSModal(theme.id)"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <Teleport to="body">
        <div
          v-if="customCSSModal.show"
          class="modal fade show d-block"
          tabindex="-1"
          style="background: rgba(0,0,0,0.5);"
        >
          <div class="modal-dialog modal-xl modal-dialog-scrollable">
            <div class="modal-content">
              <div class="modal-header">
                <h5 class="modal-title">{{ t('ui.settings.editor.corteza-studio.custom-css') }}</h5>
                <button type="button" class="btn-close" @click="resetCustomCSSModal()"></button>
              </div>
              <div class="modal-body p-0">
                <c-ace-editor
                  v-model="customCSSModal.value"
                  auto-complete
                  lang="css"
                  min-height="80vh"
                  show-line-numbers
                  :border="false"
                  :show-popout="false"
                  :auto-complete-suggestions="customCssAutocompleteVal"
                />
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-outline-secondary" @click="resetCustomCSSModal()">{{ t('admin.general.label.cancel') }}</button>
                <button type="button" class="btn btn-primary" @click="saveCustomCSSModal()">{{ t('admin.general.label.saveAndClose') }}</button>
              </div>
            </div>
          </div>
        </div>
      </Teleport>
    </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        v-if="canManage"
        :processing="processing"
        :success="success"
        :text="t('admin.general.label.submit')"
        class="ms-auto"
        @submit="onSubmit"
      />
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'ui.settings', keyPrefix: 'editor.corteza-studio' } })
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import CUploaderWithPreview from 'corteza-webapp-admin/src/components/CUploaderWithPreview'
import { components } from 'corteza-lib/vue/dist'
import { CUSTOM_CSS_AUTO_COMPLETE_VALUES } from 'corteza-webapp-admin/src/lib/cssAutoComplete'

const { CInputColorPicker, CAceEditor } = components

const { t: _t } = useI18n()

function t(key, ...args) {
  if (key.startsWith('label.') || key.startsWith('admin.')) return _t(key, ...args)
  return _t('editor.corteza-studio.' + key, ...args)
}

const props = defineProps({
  settings: {
    type: Object,
    required: true,
  },
  processing: {
    type: Boolean,
    default: false,
  },
  success: {
    type: Boolean,
    default: false,
  },
  canManage: {
    type: Boolean,
    required: true,
  },
})

const emit = defineEmits(['submit'])

const themeTabs = ['general', 'light', 'dark']
const themeVariables = [
  'black', 'white', 'primary', 'secondary', 'success',
  'warning', 'danger', 'light', 'extra-light', 'body-bg',
  'sidebar-bg', 'topbar-bg',
]
const lightModeVariables = {
  black: '#0B344E',
  white: '#FFFFFF',
  primary: '#4e73df',
  secondary: '#858796',
  success: '#43AA8B',
  warning: '#E27646',
  danger: '#4e73df',
  light: '#F3F5F7',
  'extra-light': '#E4E9EF',
  'body-bg': '#F3F5F7',
  'sidebar-bg': '#FFFFFF',
  'topbar-bg': '#F3F5F7',
}
// Almost-black dark palette. Keep in sync with
// client3/lib/vue/src/scss/dark.scss and server/pkg/provision/stylesheet.go.
// Saving this screen writes ui.studio.themes, which compiles into custom.css
// and overrides the client fallback — stale navy values here are why the
// blue dark theme kept coming back.
const darkModeVariables = {
  black: '#EDEDED',
  white: '#121212',
  primary: '#6E8FF0',
  secondary: '#9A9A9A',
  success: '#43AA8B',
  warning: '#E27646',
  danger: '#F2555A',
  light: '#1A1A1A',
  'extra-light': '#242424',
  'body-bg': '#0A0A0A',
  'sidebar-bg': '#121212',
  'topbar-bg': '#121212',
}

const activeTab = ref('general')
const themes = ref([])
const customCSSModal = reactive({
  show: false,
  id: '',
  value: '',
})
const customCssAutocompleteVal = CUSTOM_CSS_AUTO_COMPLETE_VALUES

const sassInstalled = computed(() => {
  return props.settings['ui.studio.sass-installed']
})

const installSassDocs = computed(() => '')

watch(() => props.settings, (settings) => {
  const themesData = settings['ui.studio.themes'] || []
  const customCSS = settings['ui.studio.custom-css'] || []

  themes.value = themeTabs.map((id) => {
    const { title, values = '' } = themesData.find(t => t.id === id) || {}
    const defaultCustomCSS = customCSS.find(t => t.id === id) || {}

    let variables = JSON.parse(values || '{}')
    let defaultVariables

    if (['light', 'dark'].includes(id)) {
      if (!values) {
        variables = id === 'light' ? lightModeVariables : darkModeVariables
      }

      defaultVariables = id === 'light' ? lightModeVariables : darkModeVariables
    }

    return {
      id,
      title: title || t(`tabs.${id}`),
      variables,
      defaultVariables,
      customCSS: defaultCustomCSS.values || '',
    }
  })
}, { immediate: true })

function onSubmit () {
  emit('submit', {
    'ui.studio.themes': themes.value.map(theme => {
      return {
        id: theme.id,
        title: theme.title,
        values: JSON.stringify(theme.variables),
      }
    }),
    'ui.studio.custom-css': themes.value.map(theme => {
      return {
        id: theme.id,
        title: theme.title,
        values: theme.customCSS,
      }
    }),
  })
}

function openCustomCSSModal (id) {
  const { customCSS } = themes.value.find(t => t.id === id) || {}

  customCSSModal.id = id
  customCSSModal.value = customCSS
  customCSSModal.show = true
}

function saveCustomCSSModal () {
  themes.value.find(t => t.id === customCSSModal.id).customCSS = customCSSModal.value
}

function resetCustomCSSModal () {
  customCSSModal.id = ''
  customCSSModal.value = ''
  customCSSModal.show = false
}

function resetColor (key, theme) {
  theme.variables[key] = theme.id === 'light' ? lightModeVariables[key] : darkModeVariables[key]
}

function onUpload ({ name, value }) {
  props.settings[name] = value
}

function resetAttachment (name) {
  window.__systemAPI.settingsUpdate({ values: [{ name, value: undefined }], upload: {} })
    .then(() => {
      props.settings[name] = undefined
    })
}

function uploadedFile (name) {
  const localAttachment = /^attachment:(\d+)/

  switch (true) {
    case props.settings[name] && localAttachment.test(props.settings[name]):
      const [, attachmentID] = localAttachment.exec(props.settings[name])

      return window.__systemAPI.baseURL +
        window.__systemAPI.attachmentOriginalEndpoint({
          attachmentID,
          kind: 'settings',
          name,
        })
  }

  return undefined
}
</script>
