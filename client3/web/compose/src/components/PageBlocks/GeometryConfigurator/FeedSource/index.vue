<template>
  <div class="row">
    <div
      v-for="(feed, i) in options.feeds"
      :key="i"
      class="col-12 p-0"
    >
      <div class="card list-background mx-3 mb-3">
        <div class="card-body">
          <h5
            v-if="feed.resource"
            class="d-flex align-items-center mb-3"
          >
            {{ $t('geometry.source.label') }} {{ i + 1 }}
            <c-input-confirm
              show-icon
              class="ms-auto mt-1"
              @confirmed="onRemoveFeed(i)"
            />
          </h5>
          <component
            :is="configurator(feed)"
            v-if="feed.resource && configurator(feed)"
            :feed="feed"
            :modules="modules"
            :page="page"
            :record="record"
          />
        </div>
      </div>
    </div>

    <div class="col-12">
      <button
        class="btn btn-primary test-feed-add"
        @click.prevent="handleAddButton"
      >
        {{ $t('geometry.addSource') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { usePageBlockBase } from '../../usePageBlockBase'
import { useStore } from '../../../../store'
import * as configs from './configs'
import { compose } from 'corteza-lib/js/dist'

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

const emit = defineEmits(['errors'])
const store = useStore()

const { options } = usePageBlockBase(props, emit)

const modules = computed(() => store.module.set)

if (options.value.feeds.length === 0) {
  props.block.options.feeds = []
}

function onRemoveFeed (i) {
  props.block.options.feeds.splice(i, 1)
}

function handleAddButton () {
  props.block.options.feeds.push(compose.PageBlockGeometry.makeFeed())
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
