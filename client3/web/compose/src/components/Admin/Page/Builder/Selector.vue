<template>
  <div class="container-fluid">
    <div class="row">
      <div class="col-12">
        <button
          v-for="(type) in types"
          :key="type.label"
          :disabled="isOptionDisabled(type)"
          class="btn btn-outline-secondary me-2 mb-2 text-body"
          @click="$emit('select', type.block)"
          @mouseover="current = type.image"
          @mouseleave="current = undefined"
        >
          {{ type.label }}
        </button>
      </div>

      <div
        class="col-12 d-flex align-items-center justify-content-center border rounded bg-light"
        style="height: 30vh;"
      >
        <img
          v-if="current"
          :src="current"
          class="img-fluid mx-auto mh-100"
        >
        <span v-else class="text-muted">Наведите на тип блока для предпросмотра</span>
      </div>

      <hr
        v-if="existingBlocks.length"
        class="w-100"
      >

      <div
        v-if="existingBlocks.length"
        class="col-12"
      >
        <div class="input-group d-flex w-100">
          <c-input-select
            v-model="selectedExistingBlock"
            :get-option-label="getBlockLabel"
            :get-option-key="b => b.blockID"
            :options="existingBlocks"
            :placeholder="$t('selector.selectableBlocks.placeholder')"
          />

          <button
            class="btn btn-extra-light d-flex align-items-center"
            :title="$t('selector.tooltip.clone.noRef')"
            :disabled="!selectedExistingBlock"
            @click="$emit('select', selectedExistingBlock.clone())"
          >
            <font-awesome-icon
              :icon="['far', 'clone']"
            />
          </button>

          <button
            class="btn btn-extra-light d-flex align-items-center"
            :title="$t('selector.tooltip.clone.ref')"
            :disabled="!selectedExistingBlock"
            @click="$emit('select', selectedExistingBlock)"
          >
            <font-awesome-icon
              :icon="['far', 'copy']"
            />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="js">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import * as images from '../../../../assets/PageBlocks'
import { compose } from 'corteza-lib/js/dist'

const { t } = useI18n()

defineOptions({
  i18nOptions: {
    namespaces: 'block',
  },
})

const props = defineProps({
  recordPage: {
    type: Boolean,
    default: false,
  },
  disabledKinds: {
    type: Array,
    default: () => [],
  },
  existingBlocks: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['select'])

const current = ref(undefined)
const selectedExistingBlock = ref(undefined)

const types = ref([
  { label: t('automation.label'), block: new compose.PageBlockAutomation(), image: images.Automation },
  { label: t('calendar.label'), block: new compose.PageBlockCalendar(), image: images.Calendar },
  { label: t('chart.label'), block: new compose.PageBlockChart(), image: images.Chart },
  { label: t('content.label'), block: new compose.PageBlockContent(), image: images.Content },
  { label: t('comment.label'), block: new compose.PageBlockComment(), image: images.Comment },
  { label: t('file.label'), block: new compose.PageBlockFile(), image: images.File },
  { label: t('iframe.label'), block: new compose.PageBlockIFrame(), image: images.IFrame },
  { label: t('metric.label'), block: new compose.PageBlockMetric(), image: images.Metric },
  { label: t('record.label'), block: new compose.PageBlockRecord(), image: images.Record, recordPageOnly: true },
  { label: t('recordList.label'), block: new compose.PageBlockRecordList(), image: images.RecordList },
  { label: t('recordOrganizer.label'), block: new compose.PageBlockRecordOrganizer(), image: images.RecordOrganizer },
  { label: t('recordRevisions.label'), block: new compose.PageBlockRecordRevisions(), image: images.RecordRevisions, recordPageOnly: true },
  { label: t('report.label'), block: new compose.PageBlockReport(), image: images.Report },
  { label: t('progress.label'), block: new compose.PageBlockProgress(), image: images.Progress },
  { label: t('geometry.label'), block: new compose.PageBlockGeometry(), image: images.Geometry },
  { label: t('tabs.label'), block: new compose.PageBlockTab(), image: images.Tabs },
  { label: t('navigation.label'), block: new compose.PageBlockNavigation(), image: images.Navigation },
  { label: t('ruleChain.label'), block: new compose.PageBlockRuleChain(), image: images.RuleChain },
])

function isOptionDisabled (type) {
  return (!props.recordPage && type.recordPageOnly) || props.disabledKinds.includes(type.block.kind)
}

function getBlockLabel ({ title, kind }) {
  return title || kind
}
</script>
