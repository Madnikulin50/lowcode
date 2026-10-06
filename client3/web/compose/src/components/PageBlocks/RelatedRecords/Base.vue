<template>
  <Wrap v-bind="$props" :scrollable-body="true" @refreshBlock="refreshCounts">
    <div
      v-if="!options.relations.length"
      class="p-3 text-secondary"
    >
      {{ $t('relatedRecords.notConfigured') }}
    </div>

    <div
      v-else-if="!hasRecord"
      class="p-3 text-secondary"
    >
      {{ $t('relatedRecords.saveFirst') }}
    </div>

    <div
      v-else
      class="accordion accordion-flush related-records"
    >
      <div
        v-for="s in visibleSections"
        :key="s.key"
        class="accordion-item"
      >
        <h2 class="accordion-header">
          <button
            class="accordion-button py-2"
            :class="{ collapsed: !isOpen(s.key) }"
            type="button"
            :aria-expanded="isOpen(s.key)"
            @click="toggle(s.key)"
          >
            <span class="fw-semibold">{{ s.title }}</span>
            <span
              v-if="options.showCounts && counts[s.key] !== undefined"
              class="badge rounded-pill text-bg-secondary ms-2"
            >{{ countLabel(counts[s.key]) }}</span>
            <span
              v-else-if="options.showCounts && loadingCounts"
              class="spinner-border spinner-border-sm text-secondary ms-2"
            />
          </button>
        </h2>

        <div
          v-if="visited[s.key]"
          v-show="isOpen(s.key)"
          class="accordion-body p-0"
        >
          <PageBlock
            v-bind="{ ...$props, page, blocks: [], block: s.block, blockIndex: -1 }"
            :module="s.module"
            :record="record"
            :magnified="magnified"
            :editable="false"
            :errors="{}"
          />
        </div>
      </div>

      <div
        v-if="!visibleSections.length && !loadingCounts"
        class="p-3 text-secondary"
      >
        {{ $t('relatedRecords.nothing') }}
      </div>
    </div>
  </Wrap>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose, NoID } from 'corteza-lib/js/dist'
import { debounce } from 'lodash'
import { usePageBlockBase } from './../usePageBlockBase'
import { useStore } from '../../../store'
import { usableRelations, relationKey, relationListBlock, childrenQuery, interpretCount, countLabel, COUNT_PAGE_SIZE } from 'corteza-webapp-compose/src/lib/related-records'
import Wrap from './../Wrap/index.js'
import PageBlock from './../index.js'

const { t: $t } = useI18n({ useScope: 'global' })

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

const $ComposeAPI = inject('$ComposeAPI')
const store = useStore()
const { options } = usePageBlockBase(props, emit)

const counts = reactive({})
const denied = reactive({}) // sections the user may not search: not shown
const open = reactive({})
const visited = reactive({}) // a section's list is only made when it is first opened
const loadingCounts = ref(false)

const hasRecord = computed(() => !!props.record && props.record.recordID && props.record.recordID !== NoID)

// the relations whose module and field still exist, as sections
const sections = computed(() => {
  return usableRelations(options.value.relations, id => store.module.getByID(id)).map(({ relation, module, field }) => {
    const block = compose.PageBlockMaker(relationListBlock(relation))
    // the section's own header is the accordion button: no card around the list
    block.style.wrap.kind = 'Plain'
    block.style.border.enabled = false
    return {
      key: relationKey(relation),
      relation,
      module,
      field,
      block,
      title: relation.title || module.name || module.handle,
    }
  })
})

const visibleSections = computed(() => sections.value.filter(s => {
  if (denied[s.key]) return false
  return !(options.value.hideEmptySections && counts[s.key] && counts[s.key].count === 0 && !counts[s.key].more)
}))

function isOpen (key) {
  return !!open[key]
}

function toggle (key) {
  open[key] = !open[key]
  if (open[key]) visited[key] = true
}

// sections that should be open when the block appears
watch(sections, (list) => {
  for (const s of list) {
    if (s.relation.expanded && open[s.key] === undefined) {
      open[s.key] = true
      visited[s.key] = true
    }
  }
}, { immediate: true })

async function countOf (section) {
  const { response } = $ComposeAPI.recordListCancellable({
    namespaceID: props.namespace.namespaceID,
    moduleID: section.module.moduleID,
    query: childrenQuery(section.field, props.record.recordID),
    limit: COUNT_PAGE_SIZE,
    incTotal: true,
  })

  try {
    // a query on a Record field gets total -1 from the server: see interpretCount
    const { set = [], filter = {} } = await response()
    counts[section.key] = interpretCount(set, filter)
    denied[section.key] = false
  } catch (e) {
    // no right to search this module (or it cannot be read): hide the section,
    // the others still work
    delete counts[section.key]
    denied[section.key] = true
  }
}

async function refreshCounts () {
  if (!hasRecord.value || props.loadingRecord) return
  if (!options.value.showCounts && !options.value.hideEmptySections) return

  loadingCounts.value = true
  try {
    await Promise.all(sections.value.map(countOf))
  } finally {
    loadingCounts.value = false
  }
}

const refreshSoon = debounce(refreshCounts, 300)

// a child list changed (added, edited, deleted a record) in one of the modules
function onRecordsUpdated ({ detail } = {}) {
  const moduleID = detail && detail.moduleID
  if (!moduleID || sections.value.some(s => s.module.moduleID === moduleID)) refreshSoon()
}

watch(() => [props.record && props.record.recordID, props.loadingRecord], refreshCounts, { immediate: true })
watch(() => options.value.relations, refreshSoon, { deep: true })

onMounted(() => {
  window.addEventListener('module-records-updated', onRecordsUpdated)
  window.addEventListener('refetch-records', refreshSoon)
})

onBeforeUnmount(() => {
  window.removeEventListener('module-records-updated', onRecordsUpdated)
  window.removeEventListener('refetch-records', refreshSoon)
  refreshSoon.cancel()
})
</script>
