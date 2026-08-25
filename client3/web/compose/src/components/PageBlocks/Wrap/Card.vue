<template>
  <div class="h-100">
    <div
      :id="blockID"
      class="card d-flex flex-column h-100 shadow-sm overflow-hidden position-static"
      :class="[blockClass, cardClass, customCSSClass]"
    >
      <div
        v-if="$slots.header"
        class="card-header border-bottom text-nowrap ps-3 pe-2"
        :class="[textClass, headerClass]"
      >
        <slot name="header" />
      </div>
      <div
        v-else-if="showHeader"
        class="card-header border-bottom text-nowrap ps-3 pe-2"
        :class="[textClass, headerClass]"
      >
        <div class="d-flex flex-column gap-1">
          <div
            v-if="blockTitle || showOptions"
            class="d-flex"
          >
            <h5
              v-if="blockTitle"
              :title="blockTitle"
              class="text-truncate mb-0"
            >
              {{ blockTitle }}

              <slot name="title-badge" />
            </h5>

            <div
              v-if="showOptions"
              class="ms-auto"
            >
              <button
                v-if="block.options.showRefresh"
                class="btn btn-outline-light d-print-none text-secondary px-2 py-1 border-0"
                title="Refresh"
                @click="$emit('refreshBlock')"
              >
                <font-awesome-icon :icon="['fa', 'sync']" />
              </button>

              <button
                v-if="showMagnifyButton"
                class="btn btn-outline-light d-print-none text-secondary px-2 py-1 border-0"
                :title="isBlockMagnified ? '' : 'Magnify'"
                @click="onMagnify"
              >
                <font-awesome-icon :icon="['fas', isBlockMagnified ? 'times' : 'search-plus']" />
              </button>
            </div>
          </div>

          <div
            v-if="blockDescription"
            :title="blockDescription"
            class="card-text text-dark text-wrap"
          >
            {{ blockDescription }}
          </div>
        </div>
      </div>

      <div v-if="$slots.toolbar">
        <slot name="toolbar" />
      </div>

      <div
        class="card-body p-0 flex-fill fixed-corner-container"
        :class="[bodyClass, { 'overflow-auto': scrollableBody }]"
      >
        <slot />
      </div>

      <div
        v-if="$slots.footer"
        class="card-footer p-0 bg-light"
      >
        <slot name="footer" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { compose, NoID } from 'corteza-lib/js/dist'
import { evaluatePrefilter } from 'corteza-webapp-compose/src/lib/record-filter'

const $auth = window.__auth
const route = useRoute()

const props = defineProps({
  block: { type: compose.PageBlock, required: true },
  record: { type: compose.Record, required: false, default: undefined },
  scrollableBody: { type: Boolean, required: false, default: true },
  cardClass: { type: String, required: false, default: '' },
  bodyClass: { type: String, required: false, default: '' },
  headerClass: { type: String, required: false, default: '' },
  magnified: { type: Boolean, required: false, default: false },
})

const emit = defineEmits(['refreshBlock'])

const blockID = computed(() => {
  const { blockID, meta } = props.block || {}
  return meta?.customID || blockID
})

const customCSSClass = computed(() => props.block?.meta?.customCSSClass)

const blockClass = computed(() => {
  return [
    'block',
    { border: props.block?.style?.border?.enabled },
    props.block?.kind,
  ]
})

const isBlockMagnified = computed(() => {
  const { magnifiedBlockID } = route.query
  return props.magnified && magnifiedBlockID === props.block?.blockID
})

const isAnotherBlockMagnified = computed(() => {
  const { magnifiedBlockID } = route.query
  return magnifiedBlockID && magnifiedBlockID !== props.block?.blockID
})

const showMagnifyButton = computed(() => {
  return (props.block?.options?.magnifyOption || isBlockMagnified.value) && !isAnotherBlockMagnified.value
})

const showHeader = computed(() => {
  return [props.block?.title, props.block?.description, props.block?.options?.showRefresh, showMagnifyButton.value].some(c => !!c)
})

const showOptions = computed(() => {
  return [props.block?.options?.magnifyOption, props.block?.options?.showRefresh, showMagnifyButton.value].some(c => !!c)
})

const textClass = computed(() => {
  return `text-${props.block?.style?.variants?.headerText || ''}`
})

const magnifyParams = computed(() => {
  if (!props.block) return
  const params = props.block.blockID === NoID ? { block: props.block } : { blockID: props.block.blockID }
  return isBlockMagnified.value ? undefined : params
})

const blockTitle = computed(() => {
  try {
    return evaluatePrefilter(props.block.title, {
      record: props.record,
      user: $auth.user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: ($auth.user || {}).userID || NoID,
    })
  } catch (e) {
    return e
  }
})

const blockDescription = computed(() => {
  try {
    return evaluatePrefilter(props.block.description, {
      record: props.record,
      user: $auth.user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: ($auth.user || {}).userID || NoID,
    })
  } catch (e) {
    return e
  }
})

function onMagnify() {
  window.dispatchEvent(new CustomEvent('magnify-page-block', {
    detail: isBlockMagnified.value ? undefined : magnifyParams.value,
  }))
}
</script>

<style scoped>
.fixed-corner-container {
  position: relative;
}
</style>
