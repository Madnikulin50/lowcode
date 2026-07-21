<template>
  <Wrap
    :scrollable-body="false"
  >
    <div class="h-100 w-100 card overflow-hidden bg-transparent">
      <ul
        class="nav border-0 h-100 overflow-auto"
        :class="navClasses"
        :style="{ justifyContent: options.display.alignment }"
      >
        <li
          v-for="(navItem, index) in options.navigationItems"
          :key="`navItem-${index}`"
          class="nav-item d-flex align-items-center"
          :style="{ order: index }"
        >
          <template v-if="navItem.type === 'dropdown' || isComposeDropdownPage(navItem)">
            <button
              :id="`dropdown-popover-${index}-${block.blockID}`"
              class="btn btn-link p-0 w-100 h-100"
              :style="{ color: navItem.options.textColor, background: navItem.options.backgroundColor }"
              data-bs-toggle="popover"
            >
              {{ displayDropdownText(navItem) }}
              <span class="ms-1">
                <font-awesome-icon :icon="['fas', 'chevron-down']" size="sm" />
              </span>
            </button>
            <div class="popover-container">
              <div
                class="popover"
                data-bs-content
              >
                <template v-if="navItem.type === 'dropdown'">
                  <div
                    v-for="(dropdown, dIndex) in navItem.options.item.dropdown.items"
                    :key="`dropdown-${dIndex}`"
                  >
                    <a
                      class="dropdown-item"
                      :href="dropdown.url"
                      :target="selectTargetOption(dropdown.target)"
                      :style="{ order: dIndex * 2 }"
                    >
                      {{ dropdown.label }}
                    </a>
                    <hr v-if="dropdown.delimiter" class="my-1" :style="{ order: dIndex + 1 }">
                  </div>
                </template>
                <template v-else>
                  <router-link
                    :to="{ name: 'page', params: { pageID: navItem.options.item.pageID } }"
                    :target="selectTargetOption(navItem.options.item.target)"
                    class="dropdown-item"
                    style="white-space: normal;"
                  >
                    {{ navItem.options.item.label }}
                  </router-link>
                  <hr v-if="getSubPages(navItem.options.item.pageID).length > 0" class="my-1">
                  <div
                    v-for="(dropdown, dIndex) in getSubPages(navItem.options.item.pageID)"
                    :key="`dropdown-${dIndex}`"
                  >
                    <router-link
                      :to="{ name: 'page', params: { pageID: dropdown.pageID } }"
                      :target="selectTargetOption(navItem.options.item.target)"
                      :style="{ order: dIndex * 2 }"
                      class="dropdown-item"
                      style="white-space: normal;"
                    >
                      {{ dropdown.title }}
                    </router-link>
                  </div>
                </template>
              </div>
            </div>
          </template>
          <router-link
            v-else
            class="nav-link h-100 w-100 d-flex align-items-center justify-content-center"
            :class="navItemLinkClass(navItem)"
            :style="{ color: navItem.options.textColor, background: navItem.options.backgroundColor, justifyContent: options.display.alignment }"
            :to="generateToAttributeLink(navItem) || undefined"
            :href="generateHrefAttributeLink(navItem) || undefined"
            :target="selectTargetOption(navItem.options.item.target)"
            :disabled="!navItem.options.enabled"
          >
            {{ navItem.options.item.label }}
          </router-link>
        </li>
      </ul>
    </div>
  </Wrap>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NoID } from 'corteza-lib/js/dist'
import { usePageStore } from 'corteza-webapp-compose/src/store/page'
import { usePageBlockBase } from '../usePageBlockBase'
import Wrap from '../Wrap/index.js'

const { t: $t } = useI18n({ useScope: 'global' })
const route = useRoute()
const pageStore = usePageStore()

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  record: { type: Object, required: false, default: undefined },
  mode: { type: String, required: false, default: '' },
  editable: { type: Boolean, required: false, default: false },
  resizing: { type: Boolean, required: false, default: false },
  magnified: { type: Boolean, required: false, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, required: false, default: false },
  errors: { type: Object, required: false, default: () => ({}) },
})

const emit = defineEmits(['errors', 'refreshBlock'])

const pages = computed(() => pageStore.set || [])

const { options, inModal, themeSettings } = usePageBlockBase(props, emit)

const navClasses = computed(() => {
  const appearance = options.value.display?.appearance
  return {
    'nav-tabs': appearance === 'tabs',
    'nav-pills': appearance === 'pills',
    'nav-sm': appearance === 'small',
    'nav-fill': options.value.display?.justify === 'justify',
  }
})

function navItemLinkClass(navItem) {
  return [
    !navItem.options.enabled ? 'disabled' : '',
  ]
}

function isComposeDropdownPage(navItem) {
  return navItem.type === 'compose' && navItem.options.item.displaySubPages
}

function getSubPages(pageID) {
  return pages.value.filter(value => value.selfID === pageID && value.moduleID === NoID) || []
}

function selectTargetOption(target) {
  switch (target) {
    case 'sameTab': return '_self'
    case 'newTab': return '_blank'
  }
}

function displayDropdownText(navItem) {
  if (navItem.type === 'dropdown') return navItem.options.item.dropdown.label
  return navItem.options.item.label
}

function generateToAttributeLink(navItem) {
  if (['dropdown', 'text-section'].includes(navItem.type) || isComposeDropdownPage(navItem)) return

  if (navItem.type === 'compose') {
    const pageID = navItem.options.item.pageID
    const pageLayoutID = navItem.options.item.pageLayoutID
    const moduleID = navItem.options.item.moduleID

    if (inModal.value && pageID === route.query.recordPageID) {
      return { ...route, query: { ...route.query, modalLayoutID: pageLayoutID } }
    }

    const isSamePage = pageID === route.params.pageID

    if (moduleID) {
      return isSamePage
        ? { ...route, query: { layoutID: pageLayoutID } }
        : { name: 'page.record.create', params: { pageID }, query: { layoutID: pageLayoutID } }
    }

    return isSamePage
      ? { ...route, query: { layoutID: pageLayoutID } }
      : { name: 'page', params: { pageID }, query: { layoutID: pageLayoutID } }
  }
}

function generateHrefAttributeLink(navItem) {
  if (['dropdown', 'text-section'].includes(navItem.type) || isComposeDropdownPage(navItem)) return
  return navItem.type === 'url' ? navItem.options.item.url : ''
}
</script>

<style lang="scss" scoped>
.nav-link:hover {
  text-decoration: underline !important;
}
</style>
