<template>
  <c-translator-button
    v-if="canManageResourceTranslations && resourceTranslationsEnabled"
    v-bind="{ ...$attrs, ...$props }"
    :tooltip="t('tooltip')"
    :resource="resource"
    :titles="titles"
    :fetcher="fetcher"
    :updater="updater"
  />
</template>

<script setup lang="js">
defineOptions({ i18nOptions: { namespaces: 'resource-translator', keyPrefix: 'resources.module' } })
import { computed } from 'vue'
import { useStore } from '../../../store'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import CTranslatorButton from 'corteza-webapp-compose/src/components/Translator/CTranslatorButton'
import moduleResTr from 'corteza-webapp-compose/src/lib/resource-translations/module'

const prefixed$ = 'resources.module.'
const { t: $t } = useI18n()
const t = (key) => $t(prefixed$ + key)

const props = defineProps({
  module: {
    type: compose.Module,
    required: true,
  },
  highlightKey: {
    type: String,
    default: '',
  },
  disabled: {
    type: Boolean,
    default: () => false,
  },
})

const emit = defineEmits(['update:module'])

const store = useStore()
const $ComposeAPI = window.__composeAPI

const can = computed(() => store.getters['rbac/can'])

const canManageResourceTranslations = computed(() => can.value('compose/', 'resource-translations.manage'))

const resource = computed(() => {
  const { moduleID, namespaceID } = props.module
  return `compose:module/${namespaceID}/${moduleID}`
})

const titles = computed(() => {
  const { moduleID, handle, namespaceID, fields } = props.module
  const titles = {}
  titles[resource.value] = t('title', { handle: handle || moduleID })
  fields.forEach(({ fieldID, name }) => {
    titles[`compose:module-field/${namespaceID}/${moduleID}/${fieldID}`] = t('field.title', { name })
  })
  return titles
})

const fetcher = computed(() => {
  const { moduleID, namespaceID } = props.module
  return () => {
    return $ComposeAPI.moduleListTranslations({ namespaceID, moduleID })
  }
})

const updater = computed(() => {
  const { moduleID, namespaceID } = props.module
  return translations => {
    return $ComposeAPI
      .moduleUpdateTranslations({ namespaceID, moduleID, translations })
      .then(() => fetcher.value())
      .then((translations) => {
        moduleResTr(props.module, translations, currentLanguage.value, resource.value)
        return props.module
      })
      .then(module => {
        emit('update.module', module)
      })
  }
})

const resourceTranslationsEnabled = computed(() => true)
const currentLanguage = computed(() => '')
</script>
