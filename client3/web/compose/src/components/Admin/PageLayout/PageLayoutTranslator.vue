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
defineOptions({ i18nOptions: { namespaces: 'resource-translator', keyPrefix: 'resources.page-layout' } })
import { computed } from 'vue'
import { useStore } from '../../../store'
import { useI18n } from 'vue-i18n'
import { compose, NoID } from 'corteza-lib/js/dist'
import CTranslatorButton from 'corteza-webapp-compose/src/components/Translator/CTranslatorButton'

const prefixed$ = 'resources.page-layout.'
const { t: $t } = useI18n()
const t = (key, params) => $t(prefixed$ + key, params)

const props = defineProps({
  pageLayout: {
    type: compose.PageLayout,
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

const emit = defineEmits(['update:pageLayout'])

const store = useStore()
const $ComposeAPI = window.__composeAPI

const can = computed(() => store.getters['rbac/can'])

const canManageResourceTranslations = computed(() => can.value('compose/', 'resource-translations.manage'))

const resource = computed(() => {
  const { pageID, namespaceID, pageLayoutID } = props.pageLayout
  return `compose:page-layout/${namespaceID}/${pageID}/${pageLayoutID}`
})

const titles = computed(() => {
  const titles = {}
  const { pageID, handle, meta } = props.pageLayout
  titles[resource.value] = t('title', { handle: handle || meta.title || pageID })
  return titles
})

const fetcher = computed(() => {
  const { pageLayoutID, pageID, namespaceID } = props.pageLayout
  return () => {
    return $ComposeAPI.pageLayoutListTranslations({ namespaceID, pageID, pageLayoutID })
  }
})

const updater = computed(() => {
  const { pageID, namespaceID, pageLayoutID } = props.pageLayout
  return translations => {
    return $ComposeAPI
      .pageLayoutUpdateTranslations({ namespaceID, pageID, pageLayoutID, translations })
      .then(() => fetcher.value())
      .then((translations) => {
        const find = (key) => {
          return translations.find(t => t.key === key && t.lang === currentLanguage.value && t.resource === resource.value)
        }

        let tr = find('title')
        if (tr !== undefined) {
          props.pageLayout.meta.title = tr.message
        }

        tr = find('description')
        if (tr !== undefined) {
          props.pageLayout.meta.description = tr.message
        }

        if (props.pageLayout.moduleID && props.pageLayout.moduleID !== NoID) {
          tr = find('config.buttons.new.label')
          if (tr) props.pageLayout.config.buttons.new.label = tr.message

          tr = find('config.buttons.edit.label')
          if (tr) props.pageLayout.config.buttons.edit.label = tr.message

          tr = find('config.buttons.submit.label')
          if (tr) props.pageLayout.config.buttons.submit.label = tr.message

          tr = find('config.buttons.delete.label')
          if (tr) props.pageLayout.config.buttons.delete.label = tr.message

          tr = find('config.buttons.clone.label')
          if (tr) props.pageLayout.config.buttons.clone.label = tr.message

          tr = find('config.buttons.back.label')
          if (tr) props.pageLayout.config.buttons.back.label = tr.message
        }

        return props.page
      })
      .then(page => {
        emit('update.pageLayout', page)
      })
  }
})

const resourceTranslationsEnabled = computed(() => true)
const currentLanguage = computed(() => '')
</script>
