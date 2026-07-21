<template>
  <div class="row">
    <div
      v-for="(feed, i) in options.feeds"
      :key="i"
      class="col-12 p-0"
    >
      <div class="card list-background mx-3 mb-3">
        <div class="card-body">
          <h5 class="d-flex align-items-center mb-3">
            {{ $t('calendar.source.label') }} {{ i + 1 }}
            <c-input-confirm
              show-icon
              class="ms-auto mt-1"
              @confirmed="onRemoveFeed(i)"
            />
          </h5>

          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('calendar.eventSource') }}</label>
            <c-input-select
              v-model="feed.resource"
              :options="feedSources"
              :clearable="false"
              label="text"
              :reduce="o => o.value"
            />
          </div>

          <component
            :is="configurator(feed)"
            v-if="feed.resource && configurator(feed)"
            :feed="feed"
            :modules="modules"
            :page="page"
            :record="record"
            :module="module"
          />

          <div class="row">
            <div class="col-12 col-lg-6">
              <div class="mb-3">
                <label class="form-label text-primary">{{ $t('calendar.colorLabel') }}</label>
                <c-input-color-picker
                  v-model="feed.options.color"
                  :translations="{
                    modalTitle: $t('calendar.colorPicker'),
                    light: $t('themes.labels.light'),
                    dark: $t('themes.labels.dark'),
                    cancelBtnLabel: $t('label.cancel'),
                    saveBtnLabel: $t('label.saveAndClose')
                  }"
                  :theme-settings="themeSettings"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="col-12">
      <button
        class="btn btn-primary test-feed-add"
        @click.prevent="handleAddButton"
      >
        {{ $t('calendar.addEventsSource') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed, inject } from 'vue'
import { usePageBlockBase } from '../../usePageBlockBase'
import { useStore } from '../../../../store'
import * as configs from './configs'
import { compose } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
const { CInputColorPicker } = components

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: compose.Page, required: true },
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

const emit = defineEmits(['errors'])
const store = useStore()
const $Settings = inject('$Settings')

const { options } = usePageBlockBase(props, emit)

const modules = computed(() => store.module.set)

const feedSources = computed(() =>
  Object.entries(compose.PageBlockCalendar.feedResources).map(([key, value]) => ({
    value,
    text: key,
  }))
)

const themeSettings = computed(() => $Settings.get('ui.studio.themes', []))

if (options.value.feeds.length === 0) {
  props.block.options.feeds = []
}

function onRemoveFeed (i) {
  props.block.options.feeds.splice(i, 1)
}

function handleAddButton () {
  props.block.options.feeds.push(compose.PageBlockCalendar.makeFeed())
}

function configurator (feed) {
  if (!feed.resource) return
  const r = feed.resource.split(':').pop()
  return configs[r]
}
</script>

<style lang="scss" scoped>
.list-background {
  background-color: var(--body-bg);
}
</style>
