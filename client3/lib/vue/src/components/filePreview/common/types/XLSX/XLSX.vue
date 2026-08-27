<template>
  <div
    :class="['office-preview', 'xlsx-preview-host', inline ? 'inline' : '', attrs.onClick ? 'clickable' : '']"
    :style="previewStyle"
    @click.stop="onPreviewClick"
  >
    <PreviewStatus
      :inline="!!inline"
      :loading="loading"
      :error="loadError"
      :loading-label="labels.loading || 'Loading'"
      @open-preview="onPreviewClick"
    >
      <div class="xlsx-body w-100">
        <ul
          v-if="!inline && sheetNames.length > 1"
          class="nav nav-tabs px-2 pt-2"
        >
          <li
            v-for="name in sheetNames"
            :key="name"
            class="nav-item"
          >
            <button
              type="button"
              class="nav-link"
              :class="{ active: name === activeSheet }"
              @click.stop="activeSheet = name"
            >
              {{ name }}
            </button>
          </li>
        </ul>

        <div class="table-responsive">
          <table class="table table-sm table-bordered mb-0 xlsx-table">
            <tbody>
              <tr
                v-for="(row, r) in visibleRows"
                :key="r"
              >
                <td
                  v-for="(cell, c) in row"
                  :key="c"
                >
                  {{ formatCell(cell) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <p
          v-if="truncated"
          class="text-muted small px-2 py-1 mb-0"
        >
          {{ truncatedLabel }}
        </p>
      </div>
    </PreviewStatus>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, useAttrs } from 'vue'
import PreviewStatus from '../PreviewStatus.vue'
import { OFFICE_MAX_BYTES, assertPreviewSize, fetchBinary } from '../../binary.js'

defineOptions({ inheritAttrs: false })

const attrs = useAttrs()
const props = defineProps<{
  inline?: boolean
}>()

const emit = defineEmits<{
  (e: 'openPreview'): void
  (e: 'error', err: Error): void
}>()

const loading = ref(true)
const loadError = ref<Error | null>(null)
const sheets = ref<Record<string, any[][]>>({})
const sheetNames = ref<string[]>([])
const activeSheet = ref('')

const src = computed(() => (attrs as any).src)
const labels = computed(() => (attrs as any).labels || {})
const previewStyle = computed(() => (attrs as any).previewStyle || {})
const inline = computed(() => (attrs as any).inline ?? props.inline)

const maxRows = computed(() => inline.value ? 8 : 1000)
const maxCols = computed(() => inline.value ? 8 : 64)

const rawRows = computed(() => sheets.value[activeSheet.value] || [])
const totalRows = computed(() => rawRows.value.length)
const visibleRows = computed(() => {
  return rawRows.value
    .slice(0, maxRows.value)
    .map(row => (row || []).slice(0, maxCols.value))
})
const truncated = computed(() => totalRows.value > maxRows.value || rawRows.value.some(r => (r || []).length > maxCols.value))
const truncatedLabel = computed(() => {
  const tpl = labels.value.truncated || 'Showing #0 of #1 rows. Download the file to see all data.'
  return tpl.replace('#0', String(visibleRows.value.length)).replace('#1', String(totalRows.value))
})

function formatCell (cell: any) {
  if (cell == null) {
    return ''
  }
  if (cell instanceof Date) {
    return cell.toISOString().slice(0, 10)
  }
  return String(cell)
}

function onPreviewClick () {
  if (loadError.value) {
    init()
    return
  }
  if (inline.value) {
    emit('openPreview')
  }
}

async function init () {
  loading.value = true
  loadError.value = null
  try {
    assertPreviewSize((attrs as any).meta, OFFICE_MAX_BYTES, labels.value)
    const buffer = await fetchBinary(src.value)
    const XLSX = await import('xlsx')
    const wb = XLSX.read(buffer, { type: 'array', cellDates: true })
    const next: Record<string, any[][]> = {}
    for (const name of wb.SheetNames) {
      next[name] = XLSX.utils.sheet_to_json(wb.Sheets[name], { header: 1, raw: false, defval: '' }) as any[][]
    }
    sheets.value = next
    sheetNames.value = wb.SheetNames
    activeSheet.value = wb.SheetNames[0] || ''
  } catch (err: any) {
    loadError.value = err instanceof Error ? err : new Error(String(err))
    emit('error', loadError.value)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!src.value) {
    loadError.value = new Error('src.missing')
    loading.value = false
    return
  }
  init()
})
</script>

<style lang="scss" scoped>
.xlsx-preview-host {
  width: 100%;
  background: var(--white);

  &.inline {
    cursor: zoom-in;
    overflow: hidden;
    max-height: 240px;
    pointer-events: auto;
  }
}

.xlsx-table {
  font-size: 0.8rem;
  td {
    white-space: nowrap;
    max-width: 16rem;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}
</style>
