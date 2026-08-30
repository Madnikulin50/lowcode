<template>
  <c-translator-button
    v-if="canManageResourceTranslations && resourceTranslationsEnabled"
    :tooltip="$t('tooltip')"
    v-bind="{ ...$attrs, ...$props }"
    :resource="resource"
    :titles="titles"
    :fetcher="fetcher"
    :updater="updater"
  />
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'resource-translator', keyPrefix: 'resources.namespace' } })
import { ref, computed, inject } from 'vue'
import { compose } from 'corteza-lib/js/dist'
import { useI18n } from 'vue-i18n'
import { useStore } from '../../store'
import CTranslatorButton from 'corteza-webapp-compose/src/components/Translator/CTranslatorButton'

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  disabled: { type: Boolean, default: () => false },
  highlightKey: { type: String, default: '' },
})

const { t } = useI18n()
const store = useStore()
const $ComposeAPI = inject('$ComposeAPI')

const can = computed(() => store.rbac.can)
const currentLanguage = ref('')

const canManageResourceTranslations = computed(() => can.value('compose/', 'resource-translations.manage'))
const resourceTranslationsEnabled = computed(() => true)

const resource = computed(() => `compose:namespace/${props.namespace.namespaceID}`)

const titles = computed(() => {
  const { namespaceID, slug: handle } = props.namespace
  const ts = {}
  ts[resource.value] = t('title', { handle: handle || namespaceID })
  return ts
})

const fetcher = computed(() => {
  const { namespaceID } = props.namespace
  return () => $ComposeAPI.namespaceListTranslations({ namespaceID })
})

const updater = computed(() => {
  const { namespaceID } = props.namespace
  return (translations) => {
    return $ComposeAPI
      .namespaceUpdateTranslations({ namespaceID, translations })
      .then(() => fetcher.value())
      .then((translations) => {
        const find = (key) => translations.find(t => t.key === key && t.lang === currentLanguage.value && t.resource === resource.value)
        let tr
        tr = find('name')
        if (tr !== undefined) props.namespace.name = tr.message
        tr = find('meta.subtitle')
        if (tr !== undefined) props.namespace.meta.subtitle = tr.message
        tr = find('meta.description')
        if (tr !== undefined) props.namespace.meta.description = tr.message
        tr = find('meta.prompt')
        if (tr !== undefined) props.namespace.meta.prompt = tr.message
        tr = find('meta.help')
        if (tr !== undefined) props.namespace.meta.help = tr.message
      })
  }
})
</script>
