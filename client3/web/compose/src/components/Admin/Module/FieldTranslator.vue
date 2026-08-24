<template>
  <c-translator-button
    v-if="canManageResourceTranslations && resourceTranslationsEnabled"
    v-bind="$props"
    :tooltip="t('tooltip')"
    :size="size"
    :resource="resource"
    :titles="titles"
    :fetcher="fetcher"
    :updater="updater"
    class="ms-auto py-1 px-3"
  />
</template>

<script setup lang="js">
defineOptions({ i18nOptions: { namespaces: 'resource-translator', keyPrefix: 'resources.module.field' } })
import { computed } from 'vue'
import { useStore } from '../../../store'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import CTranslatorButton from 'corteza-webapp-compose/src/components/Translator/CTranslatorButton'
import moduleFieldResTr from 'corteza-webapp-compose/src/lib/resource-translations/module-field'

const prefixed$ = 'resources.module.field.'
const { t: $t } = useI18n()
const t = (key, params) => $t(prefixed$ + key, params)

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

const can = computed(() => store.getters['rbac/can'])

const canManageResourceTranslations = computed(() => can.value('compose/', 'resource-translations.manage'))

const resource = computed(() => {
  const { fieldID } = props.field
  const { moduleID, namespaceID } = props.module
  return `compose:module-field/${namespaceID}/${moduleID}/${fieldID}`
})

const titles = computed(() => {
  const { fieldID, name } = props.field
  const titles = {}
  titles[resource.value] = t('title', { name: name || fieldID })
  return titles
})

const fetcher = computed(() => {
  const { moduleID, namespaceID } = props.module
  return () => {
    return $ComposeAPI
      .moduleListTranslations({ namespaceID, moduleID })
      .then(set => {
        return set
          .filter(({ resource: r }) => resource.value === r)
          .filter(({ key }) => !key.startsWith('meta.options'))
          .filter(({ key }) => !key.startsWith('meta.bool'))
      })
  }
})

const updater = computed(() => {
  const { moduleID, namespaceID } = props.module
  return translations => {
    return $ComposeAPI
      .moduleUpdateTranslations({ namespaceID, moduleID, translations })
      .then(() => fetcher.value())
      .then((translations) => {
        moduleFieldResTr(props.field, translations, currentLanguage.value, resource.value)
      })
  }
})

const resourceTranslationsEnabled = computed(() => true)
const currentLanguage = computed(() => '')
</script>
