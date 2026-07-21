<template>
  <c-translator-button
    v-if="canManageResourceTranslations && resourceTranslationsEnabled"
    v-bind="{ ...$attrs, ...$props }"
    :tooltip="t('tooltip')"
    :size="size"
    :resource="resource"
    :fetcher="fetcher"
    :updater="updater"
    :key-prettyfier="keyPrettifyer"
    class="ms-auto me-1 py-1 px-3"
  />
</template>

<script setup lang="js">
import { computed } from 'vue'
import { useStore } from '../../../store'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import CTranslatorButton from 'corteza-webapp-compose/src/components/Translator/CTranslatorButton'
import moduleFieldSelectResTr from 'corteza-webapp-compose/src/lib/resource-translations/module-field-select'

const prefixed$ = 'resources.module.field.'
const { t: $t } = useI18n()
const t = (key) => $t(prefixed$ + key)

const props = defineProps({
  field: {
    type: compose.ModuleField,
    required: true,
  },
  module: {
    type: compose.Module,
    required: true,
  },
  size: {
    type: String,
    default: 'lg',
  },
  disabled: {
    type: Boolean,
    default: () => false,
  },
  highlightKey: {
    type: String,
    default: '',
  },
})

const store = useStore()
const $ComposeAPI = window.__composeAPI

const keyPrefix = 'meta.options.'
const keySuffix = '.text'

function optionValueFromKey (key) {
  return key.substring(keyPrefix.length, key.length - keySuffix.length)
}

const can = computed(() => store.getters['rbac/can'])

const canManageResourceTranslations = computed(() => can.value('compose/', 'resource-translations.manage'))

const resource = computed(() => {
  const { fieldID } = props.field
  const { moduleID, namespaceID } = props.module
  return `compose:module-field/${namespaceID}/${moduleID}/${fieldID}`
})

const fetcher = computed(() => {
  const { moduleID, namespaceID } = props.module
  return () => {
    return $ComposeAPI
      .moduleListTranslations({ namespaceID, moduleID })
      .then(set => {
        set = set
          .filter(({ resource: r }) => resource.value === r)
          .filter(({ key }) => key.startsWith(keyPrefix) && key.endsWith(keySuffix))

        set
          .filter(({ lang }) => currentLanguage.value === lang)
          .forEach(rt => {
            const op = props.field.options.options
              .find(op => typeof op === 'object' && rt.key === `${keyPrefix}${op.value}${keySuffix}`)
            if (op) {
              rt.message = op.text
            }
          })

        return set
      })
  }
})

const keyPrettifyer = computed(() => optionValueFromKey)

const updater = computed(() => {
  const { moduleID, namespaceID } = props.module
  return translations => {
    return $ComposeAPI
      .moduleUpdateTranslations({ namespaceID, moduleID, translations })
      .then(() => {
        moduleFieldSelectResTr(props.field, translations, currentLanguage.value, resource.value)
        fetcher.value()
      })
  }
})

const resourceTranslationsEnabled = computed(() => true)
const currentLanguage = computed(() => '')
</script>
