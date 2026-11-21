<template>
  <c-translator-button
    v-if="canManageResourceTranslations && resourceTranslationsEnabled"
    v-bind="{ ...$attrs, ...$props }"
    :size="size"
    :tooltip="t('tooltip')"
    :resource="resource"
    :fetcher="fetcher"
    :updater="updater"
    class="ms-auto me-1 py-1 px-3"
  />
</template>

<script setup lang="js">
defineOptions({ i18nOptions: { namespaces: 'resource-translator', keyPrefix: 'resources.module.field' } })
import { computed } from 'vue'
import { useStore } from '../../../store'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import CTranslatorButton from 'corteza-webapp-compose/src/components/Translator/CTranslatorButton'

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

const keyPrefix = 'meta.bool'
const keySuffix = '.label'

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
        return set
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
        const find = (key) => {
          return translations.find(t => t.key === key && t.lang === currentLanguage && t.resource === resource.value)
        }
        let tr
        tr = find('meta.bool.true.label')
        if (tr !== undefined) {
          props.field.options.trueLabel = tr.message
        }
        tr = find('meta.bool.false.label')
        if (tr !== undefined) {
          props.field.options.falseLabel = tr.message
        }
      })
  }
})

const resourceTranslationsEnabled = computed(() => true)
const currentLanguage = computed(() => '')
</script>
