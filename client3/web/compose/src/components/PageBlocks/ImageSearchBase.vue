<template>
  <Wrap v-bind="$props">
    <div class="image-search-block h-100 p-3">
      <div v-if="!fields.length && !extraQuery" class="text-secondary small">
        Блок не настроен: укажите поля записи для запроса в настройках блока.
      </div>

      <div v-else-if="!query" class="text-secondary small">
        Заполните поля {{ fields.join(', ') }}, чтобы искать изображения.
      </div>

      <div v-else-if="loading" class="d-flex align-items-center justify-content-center h-100">
        <span class="spinner-border spinner-border-sm" />
      </div>

      <div v-else-if="error" class="text-danger small">
        {{ error }}
      </div>

      <div v-else-if="!results.length" class="text-secondary small">
        По запросу «{{ query }}» ничего не найдено.
      </div>

      <!-- one result: fill the block as fully as the image's own aspect
           ratio allows, no crop, no scrollbar -->
      <a
        v-else-if="results.length === 1"
        :href="results[0].image"
        target="_blank"
        rel="noopener noreferrer"
        class="image-search-single"
        :title="results[0].title"
      >
        <img :src="results[0].image" :alt="results[0].title">
      </a>

      <!-- several results: an even grid, cropped to square tiles so nothing
           forces the block taller than its own allotted height -->
      <div v-else class="image-search-grid">
        <a
          v-for="(r, i) in results"
          :key="i"
          :href="r.image"
          target="_blank"
          rel="noopener noreferrer"
          class="image-search-item"
          :title="r.title"
        >
          <img :src="r.thumbnail || r.image" :alt="r.title" loading="lazy">
        </a>
      </div>
    </div>
  </Wrap>
</template>

<script setup>
import { ref, computed, watch, onMounted, inject } from 'vue'
import { useRoute } from 'vue-router'
import { NoID } from 'corteza-lib/js/dist'
import Wrap from './Wrap/index.js'

const props = defineProps({
  blockIndex: { type: Number, default: -1 },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  blocks: { type: Array, default: () => [] },
  block: { type: Object, required: true },
  module: { type: Object, default: undefined },
  record: { type: Object, default: undefined },
  mode: { type: String, default: '' },
  editable: { type: Boolean, default: false },
  resizing: { type: Boolean, default: false },
  magnified: { type: Boolean, default: false },
  unsavedBlocks: { type: Set, default: () => new Set() },
  loadingRecord: { type: Boolean, default: false },
  errors: { type: Object, default: () => ({}) },
})

const $ComposeAPI = inject('$ComposeAPI', window.__composeAPI)
const route = useRoute()

const fields = computed(() => props.block.options?.fields || [])
const extraQuery = computed(() => props.block.options?.extraQuery || '')
const limit = computed(() => props.block.options?.limit || 6)

function asRecordID (id) {
  if (id == null || id === '' || id === NoID || id === '0' || id === 0) return ''
  return String(id)
}
// Same reasoning as RiskBase.vue: this component instance is reused across
// /record/:id1 -> /record/:id2 (same route, including the prev/next "< >"
// switcher, which pushes {params: {...route.params, recordID}}), so
// props.record can lag behind the URL — route.params goes first, and we
// fetch the record ourselves by that live id rather than trusting whatever
// props.record happened to hold when this last re-rendered.
const recordID = computed(() =>
  asRecordID(route.params?.recordID) || asRecordID(route.query?.recordID) || asRecordID(props.record?.recordID),
)

const record = ref(null)
const loading = ref(false)
const error = ref('')
const results = ref([])

// types.Record.Values comes back either as an array of {name, value} pairs
// or already flattened into an object, depending on the caller — handle
// both, same as RuleChainBase.vue's recordPayload() does.
function fieldValue (name) {
  const src = record.value?.values
  if (!src) return ''
  if (Array.isArray(src)) {
    const hit = src.find((row) => row?.name === name)
    return hit ? hit.value : ''
  }
  return src[name] ?? ''
}

const query = computed(() => {
  const parts = fields.value.map(fieldValue).filter(Boolean)
  if (extraQuery.value) parts.push(extraQuery.value)
  return parts.join(' ').trim()
})

async function loadRecord () {
  record.value = null
  if (!recordID.value || !props.module?.moduleID || !props.namespace?.namespaceID) return
  try {
    record.value = await $ComposeAPI.recordRead({
      namespaceID: props.namespace.namespaceID,
      moduleID: props.module.moduleID,
      recordID: recordID.value,
    })
  } catch {
    record.value = null
  }
}

async function search () {
  results.value = []
  error.value = ''
  if (!query.value || !$ComposeAPI) return
  loading.value = true
  try {
    const { results: set } = await $ComposeAPI.imageSearch({ q: query.value, limit: limit.value })
    results.value = set || []
  } catch (e) {
    error.value = e?.response?.data?.error?.message || e?.message || 'Поиск не удался'
  } finally {
    loading.value = false
  }
}

async function refresh () {
  await loadRecord()
  await search()
}

onMounted(refresh)
watch(recordID, refresh)
// Belt-and-suspenders — see RiskBase.vue's identical watcher for why.
watch(() => route.fullPath, refresh)
watch(() => [fields.value.join(','), extraQuery.value, limit.value], () => { if (record.value) search() })
</script>

<style scoped>
.image-search-single {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}
.image-search-single img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: 0.375rem;
}
.image-search-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(6rem, 1fr));
  gap: 0.5rem;
}
.image-search-item {
  display: block;
  aspect-ratio: 1 / 1;
  overflow: hidden;
  border-radius: 0.375rem;
  background: var(--bs-tertiary-bg, #f1f3f5);
}
.image-search-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
</style>
