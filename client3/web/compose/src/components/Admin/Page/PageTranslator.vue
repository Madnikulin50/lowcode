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
defineOptions({ i18nOptions: { namespaces: 'resource-translator', keyPrefix: 'resources.page' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose, NoID } from 'corteza-lib/js/dist'
import { useStore } from '../../../store'
import CTranslatorButton from 'corteza-webapp-compose/src/components/Translator/CTranslatorButton'

const prefixed$ = 'resources.page.'
const { t: $t } = useI18n()
const t = (key) => $t(prefixed$ + key)

const props = defineProps({
  page: {
    type: compose.Page,
    required: true,
  },
  pageLayouts: {
    type: Array,
    default: () => [],
  },
  pageLayout: {
    type: compose.PageLayout,
    required: false,
    default: undefined,
  },
  block: {
    type: compose.PageBlock,
    required: false,
    default: undefined,
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

const emit = defineEmits(['update:page', 'update:block', 'update:pageLayout', 'update:pageLayouts'])

const store = useStore()
const $ComposeAPI = window.__composeAPI

const can = computed(() => store.rbac.can)

const canManageResourceTranslations = computed(() => can.value('compose/', 'resource-translations.manage'))

const resource = computed(() => {
  const { pageID, namespaceID } = props.page
  return `compose:page/${namespaceID}/${pageID}`
})

const titles = computed(() => {
  const titles = {}
  if (props.block) {
    const { title, blockID } = props.block
    titles[resource.value] = t('block.title', { title, blockID })
  } else {
    const { pageID, handle } = props.page
    titles[resource.value] = t('title', { handle: handle || pageID })
  }

  if (props.pageLayout) {
    const { namespaceID, pageID, pageLayoutID, handle, meta } = props.pageLayout
    titles[`compose:page-layout/${namespaceID}/${pageID}/${pageLayoutID}`] = t('layout.title', { handle: handle || meta.title || pageLayoutID })
  } else {
    props.pageLayouts.forEach(({ namespaceID, pageID, pageLayoutID, handle, meta }) => {
      titles[`compose:page-layout/${namespaceID}/${pageID}/${pageLayoutID}`] = t('layout.title', { handle: handle || meta.title || pageLayoutID })
    })
  }

  return titles
})

const fetcher = computed(() => {
  const { pageID, namespaceID } = props.page
  return () => {
    return $ComposeAPI.pageListTranslations({ namespaceID, pageID })
      .then(set => {
        if (props.block) {
          set = set.filter(({ key }) => key.startsWith(`pageBlock.${props.block.blockID}.`))
        }
        if (props.pageLayout) {
          set = set.filter(({ resource }) => resource.endsWith(`${pageID}`) || resource.endsWith(`/${props.pageLayout.pageLayoutID}`))
        }
        return set
      })
  }
})

const updater = computed(() => {
  const { pageID, namespaceID } = props.page
  return translations => {
    return $ComposeAPI
      .pageUpdateTranslations({ namespaceID, pageID, translations })
      .then(() => fetcher.value())
      .then((translations) => {
        const find = (key) => {
          return translations.find(t => t.key === key && t.lang === currentLanguage.value && t.resource === resource.value)
        }
        const layoutResource = ({ pageLayoutID, pageID, namespaceID }) => `compose:page-layout/${namespaceID}/${pageID}/${pageLayoutID}`

        let tr = find('title')
        if (tr !== undefined) {
          props.page.title = tr.message
        }

        tr = find('description')
        if (tr !== undefined) {
          props.page.description = tr.message
        }

        if (props.page.moduleID && props.page.moduleID !== NoID) {
          tr = find('recordToolbar.new.label')
          if (tr) props.page.config.buttons.new.label = tr.message

          tr = find('recordToolbar.edit.label')
          if (tr) props.page.config.buttons.edit.label = tr.message

          tr = find('recordToolbar.submit.label')
          if (tr) props.page.config.buttons.submit.label = tr.message

          tr = find('recordToolbar.delete.label')
          if (tr) props.page.config.buttons.delete.label = tr.message

          tr = find('recordToolbar.clone.label')
          if (tr) props.page.config.buttons.clone.label = tr.message

          tr = find('recordToolbar.back.label')
          if (tr) props.page.config.buttons.back.label = tr.message
        }

        const updateBlockTranslations = block => {
          if (block.blockID === NoID) return block
          block.title = (find(`pageBlock.${block.blockID}.title`) || {}).message
          block.description = (find(`pageBlock.${block.blockID}.description`) || {}).message

          switch (true) {
            case block instanceof compose.PageBlockAutomation:
              block.options.buttons.forEach((btn, index) => {
                tr = find(`pageBlock.${block.blockID}.button.${btn.buttonID || index}.label`)
                if (tr) btn.label = tr.message
              })
              break
            case block instanceof compose.PageBlockRecordList:
              block.options.selectionButtons.forEach((btn, index) => {
                tr = find(`pageBlock.${block.blockID}.button.${btn.buttonID || index}.label`)
                if (tr) btn.label = tr.message
              })
              break
            case block instanceof compose.PageBlockContent:
              tr = find(`pageBlock.${block.blockID}.content.body`)
              if (tr) block.options.body = tr.message
              break
          }
          return block
        }

        const updatePageLayoutTranslations = pageLayout => {
          if (pageLayout.pageLayoutID === NoID) return pageLayout
          const find = (key) => {
            return translations.find(t => t.key === key && t.lang === currentLanguage.value && t.resource === layoutResource(pageLayout))
          }
          let tr = find('title')
          if (tr !== undefined) pageLayout.meta.title = tr.message

          tr = find('description')
          if (tr !== undefined) pageLayout.meta.description = tr.message

          if (pageLayout.moduleID && pageLayout.moduleID !== NoID) {
            tr = find('config.buttons.new.label')
            if (tr) pageLayout.config.buttons.new.label = tr.message

            tr = find('config.buttons.edit.label')
            if (tr) pageLayout.config.buttons.edit.label = tr.message

            tr = find('config.buttons.submit.label')
            if (tr) pageLayout.config.buttons.submit.label = tr.message

            tr = find('config.buttons.delete.label')
            if (tr) pageLayout.config.buttons.delete.label = tr.message

            tr = find('config.buttons.clone.label')
            if (tr) pageLayout.config.buttons.clone.label = tr.message

            tr = find('config.buttons.back.label')
            if (tr) pageLayout.config.buttons.back.label = tr.message
          }
          return pageLayout
        }

        if (props.block) {
          props.block = updateBlockTranslations(props.block)
        } else {
          props.page.blocks = props.page.blocks.map(block => updateBlockTranslations(block))
        }

        if (props.pageLayout) {
          props.pageLayout = updatePageLayoutTranslations(props.pageLayout)
        } else {
          props.pageLayouts = props.pageLayouts.map(pageLayout => updatePageLayoutTranslations(pageLayout))
        }

        return props.page
      })
      .then(page => {
        emit('update.page', page)
        if (props.block) {
          emit('update.block', props.block)
        }
        if (props.pageLayout) {
          emit('update.pageLayout', props.pageLayout)
        } else {
          emit('update.pageLayouts', props.pageLayouts)
        }
      })
  }
})

const resourceTranslationsEnabled = computed(() => true)
const currentLanguage = computed(() => '')
</script>
