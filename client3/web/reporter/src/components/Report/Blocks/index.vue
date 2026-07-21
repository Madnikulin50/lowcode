<template>
  <Wrap v-if="block" :block="block" :index="index" :wrap="block.wrap">
    <template #header>
      <div v-if="block.title || block.description" class="d-flex flex-column p-3 border-bottom gap-1">
        <h4 v-if="block.title" class="text-primary text-truncate mb-0">{{ block.title }}</h4>
        <p v-if="block.description" class="text-dark text-wrap mb-0">{{ block.description }}</p>
      </div>
    </template>
    <template #default>
      <split
        v-if="showDisplayElements"
        ref="split"
        :direction="block.layout"
        :gutter-size="12"
        class="h-100"
        @onDragEnd="setDisplayElementSizes"
      >
        <split-area
          v-for="(element, displayElementIndex) in block.elements"
          :key="displayElementIndex"
          :size="element.meta.size"
          :min-size="0"
          :class="{
            'overflow-hidden h-100': element.kind !== 'Text',
            'w-100': block.elements.length === 1,
          }"
          class="position-relative"
        >
          <div v-if="processing" class="d-flex align-items-center justify-content-center h-100">
            <div class="spinner-border" />
          </div>
          <display-element
            v-else
            :display-element="element"
            :labels="{
              previous: t('table.view.previous'),
              next: t('table.view.next'),
            }"
            @update="updateDataframes({ displayElementIndex, definition: $event })"
          />
        </split-area>
      </split>
    </template>
  </Wrap>
</template>
<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useToast } from 'corteza-lib/vue/dist'
import { reporter } from 'corteza-lib/js/dist'
import { cloneDeep } from 'lodash'
import Wrap from './Wrap/index.js'
import { Split, SplitArea } from 'vue-split-panel'
import DisplayElement from './DisplayElements/Viewers/index.js'

const props = defineProps({
  index: { type: Number, default: -1 },
  block: { type: Object, required: true },
  scenario: { type: Object, default: () => ({}) },
  reportID: { type: String, required: false, default: '0' },
})
const emit = defineEmits(['item-updated'])

const { t } = useI18n()
const toast = useToast()
const toastErrorHandler = toast.toastErrorHandler

const processing = ref(false)
const dataframes = ref({})
const showDisplayElements = ref(false)

watch(() => props.block.blockID, () => { renderBlock() }, { immediate: true })
watch(() => props.block.elements.length, (length, oldLength) => {
  const addedOrRemoved = length !== oldLength && oldLength
  if (addedOrRemoved) {
    const defaultSize = Math.floor(100 / length)
    props.block.elements = props.block.elements.map(e => { e.meta.size = defaultSize; return e })
  }
  renderBlock()
})

function renderBlock() {
  const { elements = [] } = props.block || {}
  if (elements.length) {
    runReport()
    showDisplayElements.value = false
    nextTick(() => { showDisplayElements.value = true })
  }
}

function setDisplayElementSizes(sizes = []) {
  sizes.forEach((size, index) => { props.block.elements[index].meta.size = size })
  emit('item-updated', props.index)
}

function getScenarioDefinition(element) {
  const scenarioDefinition = {}
  if (props.scenario.filters) {
    Object.keys(props.scenario.filters).map(k => {
      const v = props.scenario.filters[k]
      scenarioDefinition[k] = { ref: k, filter: cloneDeep(v) }
    })
  }
  return scenarioDefinition
}

function runReport() {
  processing.value = true
  dataframes.value = {}
  const frames = []
  props.block.elements.forEach((element, key) => {
    element = reporter.DisplayElementMaker(element)
    if (element && element.kind !== 'Text') {
      if (element.elementID === '0') element.elementID = `${key}`
      const scDefs = getScenarioDefinition(element)
      const { dataframes: dfs = [] } = element.reportDefinitions(scDefs)
      dfs.forEach(frame => { frame.filter = cloneDeep(frame.filter) })
      frames.push(...dfs.filter(({ source }) => source))
    }
  })
  if (frames.length) {
    window.__systemAPI.reportRun({ frames, reportID: props.reportID })
      .then(({ frames: resultFrames = [] }) => {
        props.block.elements = props.block.elements.map((element, key) => {
          if (element.elementID === '0') element.elementID = `${key}`
          const dataframes = resultFrames.filter(({ name }) => name === element.elementID)
          return { ...element, dataframes }
        })
      }).catch((e) => { toastErrorHandler(t('notification.report.runFailed'))(e) })
      .finally(() => { processing.value = false })
  } else {
    processing.value = false
  }
}

function updateDataframes({ displayElementIndex, definition }) {
  const element = reporter.DisplayElementMaker(props.block.elements[displayElementIndex])
  const frames = []
  if (element && element.kind !== 'Text') {
    const scenarioDefinition = getScenarioDefinition(element)
    Object.entries(definition).forEach(([key, value]) => { definition[key] = { ...value, ...scenarioDefinition[key] } })
    const { dataframes: dfs = [] } = element.reportDefinitions(definition)
    frames.push(...dfs.filter(({ source }) => source))
    if (frames.length) {
      window.__systemAPI.reportRun({ frames, reportID: props.reportID })
        .then(({ frames: resultFrames = [] }) => {
          const found = props.block.elements.find(({ elementID }) => elementID === element.elementID)
          if (found) found.dataframes = resultFrames
        }).catch((e) => { toastErrorHandler(t('notification.report.runFailed'))(e) })
    }
  }
}
</script>
<style lang="scss">
.split .gutter { background-color: transparent; }
</style>
