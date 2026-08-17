<template>
  <div class="block-selector">
    <div class="d-flex flex-wrap gap-2 align-items-center mb-3">
      <c-input-search
        v-model="query"
        class="flex-grow-1"
        :placeholder="$t('selector.searchPlaceholder')"
        autocomplete="off"
      />
    </div>

    <div class="d-flex flex-wrap gap-1 mb-3" role="tablist">
      <button
        v-for="cat in categories"
        :key="cat.id"
        type="button"
        class="btn btn-sm"
        :class="category === cat.id ? 'btn-primary' : 'btn-outline-secondary'"
        @click="category = cat.id"
      >
        {{ cat.label }}
      </button>
    </div>

    <div
      v-if="filteredTypes.length"
      class="row g-3"
    >
      <div
        v-for="type in filteredTypes"
        :key="type.block.kind"
        class="col-12 col-sm-6 col-xl-4"
      >
        <button
          type="button"
          class="block-type-card w-100 text-start border rounded-3 bg-white overflow-hidden"
          :class="{ disabled: isOptionDisabled(type) }"
          :disabled="isOptionDisabled(type)"
          :title="disabledHint(type)"
          @click="$emit('select', type.block)"
        >
          <div class="block-type-card__preview border-bottom">
            <block-kind-preview :kind="type.block.kind" />
          </div>
          <div class="p-3">
            <div class="d-flex align-items-center gap-2 mb-1">
              <font-awesome-icon
                :icon="type.icon"
                class="text-primary flex-shrink-0"
              />
              <span class="fw-semibold text-body text-truncate">{{ type.label }}</span>
              <span
                v-if="type.recordPageOnly"
                class="badge bg-light text-secondary border ms-auto flex-shrink-0"
              >{{ $t('selector.recordPageOnly') }}</span>
            </div>
            <p class="small text-muted mb-0">
              {{ type.description }}
            </p>
          </div>
        </button>
      </div>
    </div>

    <div
      v-else
      class="text-center text-muted py-5"
    >
      {{ $t('selector.empty') }}
    </div>

    <div
      v-if="existingBlocks.length"
      class="reuse-panel mt-4 pt-3 border-top"
    >
      <div class="fw-semibold mb-2">
        {{ $t('selector.reuse.title') }}
      </div>
      <div class="input-group">
        <c-input-select
          v-model="selectedExistingBlock"
          :get-option-label="getBlockLabel"
          :get-option-key="b => b.blockID"
          :options="existingBlocks"
          :placeholder="$t('selector.selectableBlocks.placeholder')"
        />
        <button
          type="button"
          class="btn btn-outline-secondary d-flex align-items-center gap-1"
          :title="$t('selector.tooltip.clone.noRef')"
          :disabled="!selectedExistingBlock"
          @click="$emit('select', selectedExistingBlock.clone())"
        >
          <font-awesome-icon :icon="['far', 'clone']" />
          <span class="d-none d-md-inline">{{ $t('selector.reuse.clone') }}</span>
        </button>
        <button
          type="button"
          class="btn btn-outline-secondary d-flex align-items-center gap-1"
          :title="$t('selector.tooltip.clone.ref')"
          :disabled="!selectedExistingBlock"
          @click="$emit('select', selectedExistingBlock)"
        >
          <font-awesome-icon :icon="['far', 'copy']" />
          <span class="d-none d-md-inline">{{ $t('selector.reuse.reference') }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({
  i18nOptions: {
    namespaces: 'block',
  },
})

import { ref, computed } from 'vue'
import { components, useNsI18n } from 'corteza-lib/vue/dist'
import { compose } from 'corteza-lib/js/dist'
import BlockKindPreview from './BlockKindPreview.vue'

const { CInputSearch } = components

const t = useNsI18n()

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

defineEmits(['select'])

const query = ref('')
const category = ref('all')
const selectedExistingBlock = ref(undefined)

const categories = computed(() => [
  { id: 'all', label: t('selector.categories.all') },
  { id: 'data', label: t('selector.categories.data') },
  { id: 'visualize', label: t('selector.categories.visualize') },
  { id: 'content', label: t('selector.categories.content') },
  { id: 'layout', label: t('selector.categories.layout') },
  { id: 'automation', label: t('selector.categories.automation') },
])

const catalog = [
  { kind: 'RecordList', category: 'data', icon: ['fas', 'table'], recordPageOnly: false, block: new compose.PageBlockRecordList() },
  { kind: 'Record', category: 'data', icon: ['far', 'file-alt'], recordPageOnly: true, block: new compose.PageBlockRecord() },
  { kind: 'RecordOrganizer', category: 'data', icon: ['fas', 'columns'], recordPageOnly: false, block: new compose.PageBlockRecordOrganizer() },
  { kind: 'RecordRevisions', category: 'data', icon: ['far', 'clock'], recordPageOnly: true, block: new compose.PageBlockRecordRevisions() },
  { kind: 'Comment', category: 'data', icon: ['fas', 'comments'], recordPageOnly: false, block: new compose.PageBlockComment() },
  { kind: 'Chart', category: 'visualize', icon: ['fas', 'chart-pie'], recordPageOnly: false, block: new compose.PageBlockChart() },
  { kind: 'Metric', category: 'visualize', icon: ['fas', 'chart-bar'], recordPageOnly: false, block: new compose.PageBlockMetric() },
  { kind: 'Progress', category: 'visualize', icon: ['fas', 'tasks'], recordPageOnly: false, block: new compose.PageBlockProgress() },
  { kind: 'Calendar', category: 'visualize', icon: ['fas', 'calendar-alt'], recordPageOnly: false, block: new compose.PageBlockCalendar() },
  { kind: 'Geometry', category: 'visualize', icon: ['fas', 'map-marked-alt'], recordPageOnly: false, block: new compose.PageBlockGeometry() },
  { kind: 'Report', category: 'visualize', icon: ['fas', 'book'], recordPageOnly: false, block: new compose.PageBlockReport() },
  { kind: 'Content', category: 'content', icon: ['fas', 'paragraph'], recordPageOnly: false, block: new compose.PageBlockContent() },
  { kind: 'File', category: 'content', icon: ['fas', 'paperclip'], recordPageOnly: false, block: new compose.PageBlockFile() },
  { kind: 'IFrame', category: 'content', icon: ['fas', 'external-link-alt'], recordPageOnly: false, block: new compose.PageBlockIFrame() },
  { kind: 'Tabs', category: 'layout', icon: ['fas', 'folder'], recordPageOnly: false, block: new compose.PageBlockTab() },
  { kind: 'Navigation', category: 'layout', icon: ['fas', 'bars'], recordPageOnly: false, block: new compose.PageBlockNavigation() },
  { kind: 'Automation', category: 'automation', icon: ['fas', 'play'], recordPageOnly: false, block: new compose.PageBlockAutomation() },
  { kind: 'RuleChain', category: 'automation', icon: ['fas', 'project-diagram'], recordPageOnly: false, block: new compose.PageBlockRuleChain() },
  { kind: 'AiChat', category: 'automation', icon: ['fas', 'brain'], recordPageOnly: false, block: new compose.PageBlockAiChat() },
]

const types = computed(() => catalog.map(def => ({
  ...def,
  label: t(labelKey(def.kind)),
  description: t(`selector.types.${camelKind(def.kind)}.description`),
})))

const filteredTypes = computed(() => {
  const q = query.value.trim().toLowerCase()
  return types.value.filter(type => {
    if (category.value !== 'all' && type.category !== category.value) return false
    if (!q) return true
    return type.label.toLowerCase().includes(q) ||
      type.description.toLowerCase().includes(q) ||
      type.kind.toLowerCase().includes(q)
  })
})

function camelKind (kind) {
  return kind.charAt(0).toLowerCase() + kind.slice(1)
}

function labelKey (kind) {
  const keys = {
    RecordList: 'recordList.label',
    RecordOrganizer: 'recordOrganizer.label',
    RecordRevisions: 'recordRevisions.label',
    RuleChain: 'ruleChain.label',
    AiChat: 'aiChat.label',
    IFrame: 'iframe.label',
  }
  return keys[kind] || `${camelKind(kind)}.label`
}

function isOptionDisabled (type) {
  return (!props.recordPage && type.recordPageOnly) || props.disabledKinds.includes(type.block.kind)
}

function disabledHint (type) {
  if (!props.recordPage && type.recordPageOnly) return t('selector.recordPageOnlyHint')
  if (props.disabledKinds.includes(type.block.kind)) return t('selector.disabledHere')
  return ''
}

function getBlockLabel ({ title, kind }) {
  return title || kind
}
</script>

<style lang="scss" scoped>
.block-type-card {
  padding: 0;
  color: inherit;
  cursor: pointer;
  box-shadow: 0 0.125rem 0.35rem rgba(58, 59, 69, 0.08);
  transition: box-shadow 0.15s ease, transform 0.15s ease, border-color 0.15s ease;

  &:hover:not(:disabled) {
    border-color: var(--bs-primary) !important;
    box-shadow: 0 0.35rem 0.85rem rgba(78, 115, 223, 0.18);
    transform: translateY(-1px);
  }

  &:focus-visible {
    outline: 2px solid var(--bs-primary);
    outline-offset: 2px;
  }

  &.disabled,
  &:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
}

.block-type-card__preview {
  background: #f4f6fb;
  pointer-events: none;
}
</style>
