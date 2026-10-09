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
          <table class="xlsx-table">
            <colgroup>
              <col
                v-for="col in grid.cols"
                :key="col.c"
                :style="{ width: `${col.widthPx}px` }"
              >
            </colgroup>
            <tbody>
              <tr
                v-for="row in grid.rows"
                :key="row.r"
                :style="{ height: `${row.heightPx}px` }"
              >
                <template
                  v-for="cell in row.cells"
                  :key="cell.c"
                >
                  <td
                    v-if="!cell.covered"
                    :rowspan="cell.rowspan > 1 ? cell.rowspan : undefined"
                    :colspan="cell.colspan > 1 ? cell.colspan : undefined"
                    :style="cell.style"
                  >
                    {{ cell.text }}
                  </td>
                </template>
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
import { ref, shallowRef, computed, onMounted, useAttrs } from 'vue'
import PreviewStatus from '../PreviewStatus.vue'
import { OFFICE_MAX_BYTES, assertPreviewSize, fetchBinary } from '../../binary.js'
import { readLayout, ptToPx, colWidthToPx, DEFAULT_ROW_HEIGHT_PT, DEFAULT_COL_WIDTH_PX } from './layout.js'

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
const sheets = shallowRef<Record<string, any>>({})
const layouts = shallowRef<Record<string, any>>({})
let utils: any = null
const sheetNames = ref<string[]>([])
const activeSheet = ref('')

const src = computed(() => (attrs as any).src)
const labels = computed(() => (attrs as any).labels || {})
const previewStyle = computed(() => (attrs as any).previewStyle || {})
const inline = computed(() => (attrs as any).inline ?? props.inline)

const maxRows = computed(() => inline.value ? 8 : 1000)
const maxCols = computed(() => inline.value ? 8 : 64)

// Visible (non-hidden) row/column indexes of the active sheet, in sheet coordinates.
const sheetRange = computed(() => {
  const ws = sheets.value[activeSheet.value]
  if (!ws || !ws['!ref'] || !utils) {
    return { rows: [] as number[], cols: [] as number[] }
  }
  const { s, e } = utils.decode_range(ws['!ref'])
  const rows: number[] = []
  const cols: number[] = []
  for (let r = s.r; r <= e.r; r++) {
    if (!ws['!rows']?.[r]?.hidden) rows.push(r)
  }
  for (let c = s.c; c <= e.c; c++) {
    if (!ws['!cols']?.[c]?.hidden) cols.push(c)
  }
  return { rows, cols }
})

const totalRows = computed(() => sheetRange.value.rows.length)

const grid = computed(() => {
  const ws = sheets.value[activeSheet.value]
  const layout = layouts.value[activeSheet.value] || {}
  const borders: Map<string, any> = layout.borders || new Map()
  const rowIdx = sheetRange.value.rows.slice(0, maxRows.value)
  const colIdx = sheetRange.value.cols.slice(0, maxCols.value)
  if (!ws) {
    return { rows: [], cols: [] }
  }

  const rowPos = new Map(rowIdx.map((r, i) => [r, i]))
  const colPos = new Map(colIdx.map((c, i) => [c, i]))

  // Merges → span at the top-left visible cell, every other cell is covered.
  const spans = new Map<string, { rowspan: number, colspan: number }>()
  const covered = new Set<string>()
  for (const m of ws['!merges'] || []) {
    const mr = rowIdx.filter(r => r >= m.s.r && r <= m.e.r)
    const mc = colIdx.filter(c => c >= m.s.c && c <= m.e.c)
    if (!mr.length || !mc.length) continue
    spans.set(`${mr[0]}:${mc[0]}`, { rowspan: mr.length, colspan: mc.length })
    for (const r of mr) {
      for (const c of mc) {
        if (r !== mr[0] || c !== mc[0]) covered.add(`${r}:${c}`)
      }
    }
  }

  const cols = colIdx.map(c => ({
    c,
    widthPx: colWidthToPx(ws['!cols']?.[c]) ?? layout.defaultColWidthPx ?? DEFAULT_COL_WIDTH_PX,
  }))

  const rows = rowIdx.map(r => {
    const info = ws['!rows']?.[r]
    const heightPx = info?.hpt != null
      ? ptToPx(info.hpt)
      : info?.hpx ?? ptToPx(layout.defaultRowHeightPt ?? DEFAULT_ROW_HEIGHT_PT)

    return {
      r,
      heightPx,
      cells: colIdx.map(c => {
        const key = `${r}:${c}`
        const span = spans.get(key)
        let style = borders.get(key)
        if (span) {
          // Merged range draws its outer edges from the bottom/right corner cells too.
          const lastR = rowIdx[(rowPos.get(r) as number) + span.rowspan - 1]
          const lastC = colIdx[(colPos.get(c) as number) + span.colspan - 1]
          const br = borders.get(`${lastR}:${lastC}`) || {}
          const tr = borders.get(`${r}:${lastC}`) || {}
          const bl = borders.get(`${lastR}:${c}`) || {}
          style = {
            ...style,
            ...(tr.borderRight && { borderRight: tr.borderRight }),
            ...(bl.borderBottom && { borderBottom: bl.borderBottom }),
            ...(br.borderRight && { borderRight: br.borderRight }),
            ...(br.borderBottom && { borderBottom: br.borderBottom }),
          }
        }
        return {
          c,
          text: formatCell(ws[utils.encode_cell({ r, c })]),
          style,
          rowspan: span?.rowspan || 1,
          colspan: span?.colspan || 1,
          covered: covered.has(key),
        }
      }),
    }
  })

  return { rows, cols }
})

const truncated = computed(() => totalRows.value > maxRows.value || sheetRange.value.cols.length > maxCols.value)
const truncatedLabel = computed(() => {
  const tpl = labels.value.truncated || 'Showing #0 of #1 rows. Download the file to see all data.'
  return tpl.replace('#0', String(grid.value.rows.length)).replace('#1', String(totalRows.value))
})

function formatCell (cell: any) {
  if (cell == null) {
    return ''
  }
  if (cell.w != null) {
    return cell.w
  }
  if (cell.v instanceof Date) {
    return cell.v.toISOString().slice(0, 10)
  }
  return cell.v == null ? '' : String(cell.v)
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
    const wb = XLSX.read(buffer, { type: 'array', cellDates: true, cellStyles: true })
    utils = XLSX.utils
    layouts.value = await readLayout(buffer)
    sheets.value = wb.Sheets
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
  border-collapse: collapse;
  table-layout: fixed;
  width: max-content;

  td {
    padding: 0 3px;
    line-height: 1.2;
    vertical-align: bottom;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    // Excel-like gridlines as a shadow, so they never win border-collapse
    // conflicts against real (dashed/dotted) borders coming as inline styles.
    border: none;
    box-shadow: inset -1px -1px 0 #e1e1e1;
  }
}
</style>
