<template>
  <div class="card h-100 rounded-0">
    <div class="card-body p-0 overflow-auto">
      <div
        class="table-responsive"
        :class="{ 'table-dark': options?.dark }"
      >
        <table
          class="table table-sm mb-0 mb-0"
          :class="{
            'table-bordered': options?.bordered,
            'table-borderless': options?.borderless,
            'table-striped': options?.striped,
            'table-hover': options?.hover,
          }"
        >
          <colgroup
            v-for="(cg, i) in tabelify.colgroups"
            :key="i"
            :class="{ local: cg.isLocal, foreign: !cg.isLocal }"
            :span="cg.size"
          />

          <thead
            :class="options?.headVariant === 'dark' ? 'table-dark' : 'table-light'"
          >
            <tr v-if="dataframes.length > 1">
              <th
                v-for="(c, i) in tabelify.header"
                :key="i"
                v-bind="c.column ? c.column.attrs : {}"
                class="border-0"
              >
                <p
                  v-if="c.sourceName"
                  class="m-0"
                >
                  {{ c.sourceName }}
                </p>
              </th>
            </tr>
            <tr>
              <th
                v-for="(c, i) in tabelify.header"
                :key="i"
                class="border-0"
              >
                <div class="d-flex align-items-center">
                  <div
                    v-if="c.column ? c.column.label : ''"
                    class="d-flex text-nowrap"
                  >
                    {{ c.column.label }}
                  </div>

                  <button
                    v-if="!c.meta.tmp_noSort"
                    class="btn btn-outline-extra-light d-flex align-items-center text-secondary d-print-none border-0 px-1 ms-1"
                    @click="handleSort(c.meta.sortKey)"
                  >
                    <font-awesome-layers class="d-print-none">
                      <font-awesome-icon
                        :icon="['fas', 'angle-up']"
                        class="mb-1"
                        :class="{ 'text-primary': sort.field === c.meta.sortKey && !sort.descending }"
                      />
                      <font-awesome-icon
                        :icon="['fas', 'angle-down']"
                        class="mt-1"
                        :class="{ 'text-primary': sort.field === c.meta.sortKey && sort.descending }"
                      />
                    </font-awesome-layers>
                  </button>
                </div>
              </th>
            </tr>
          </thead>

          <tbody>
            <tr
              v-for="(r, i) in tabelify.rows"
              :key="i"
              :class="{
                separator: !!(r[0] || {}).separator,
              }"
            >
              <td
                v-for="(c, j) in r"
                :key="j"
                v-bind="c.attrs || {}"
              >
                {{ c.value }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="card-footer d-flex rounded-0 bg-light p-2">
      <button
        data-bs-toggle="tooltip"
        title="Export to CSV"
        class="btn btn-outline-extra-light text-dark border-0 me-auto py-1 px-2"
        @click="exportCSV"
      >
        <font-awesome-icon :icon="['fas', 'download']" />
      </button>

      <div class="btn-group ms-auto gap-1">
        <button
          v-if="hasPrevPage"
          class="btn btn-outline-extra-light d-flex align-items-center justify-content-center text-dark border-0 p-1"
          @click="goToPage()"
        >
          <font-awesome-icon :icon="['fas', 'angle-double-left']" />
        </button>

        <button
          v-if="hasPrevPage"
          class="btn btn-outline-extra-light d-flex align-items-center justify-content-center text-dark border-0 p-1"
          @click="goToPage('prevPage')"
        >
          <font-awesome-icon :icon="['fas', 'angle-left']" class="me-1" />
          {{ labels?.previous || 'Previous' }}
        </button>

        <button
          v-if="hasNextPage"
          class="btn btn-outline-extra-light d-flex align-items-center justify-content-center text-dark border-0 p-1"
          @click="goToPage('nextPage')"
        >
          {{ labels?.next || 'Next' }}
          <font-awesome-icon :icon="['fas', 'angle-right']" class="ms-1" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = defineProps<{
  displayElement: any
  labels?: Record<string, any>
}>()

const emit = defineEmits<{
  'update': [value: any]
}>()

const sort = ref<{ field: string; descending: boolean }>({
  field: '',
  descending: false,
})

const cursors = ref<any[]>([])
const cursor = ref<any>(undefined)

const dataframes = computed(() => props.displayElement?.dataframes || [])
const options = computed(() => props.displayElement?.options || undefined)

const localDataframe = computed(() => dataframes.value[0])

const indexFrames = computed(() => {
  const ix: Record<string, any> = {}
  for (const df of dataframes.value || []) {
    if (!df.relSource) continue
    if (!ix[df.relSource]) ix[df.relSource] = {}
    if (!ix[df.relSource][df.refValue]) ix[df.relSource][df.refValue] = []
    ix[df.relSource][df.refValue].push(df)
  }
  return ix
})

const tabelify = computed(() => {
  if (!dataframes.value.length) return { rows: [], header: [], colgroups: [] }
  return tabelifyFrame(localDataframe.value)
})

const hasPrevPage = computed(() => !!cursors.value.length)

const nextPage = computed(() => {
  if (localDataframe.value) {
    return localDataframe.value.paging?.nextPage
  }
  return undefined
})

const hasNextPage = computed(() => !!nextPage.value)

watch(localDataframe, (dataframe, oldDataframe) => {
  if (dataframe && !oldDataframe) {
    const firstField = dataframe.sort.includes(',') ? dataframe.sort.split(',')[0] : dataframe.sort
    if (firstField.includes('DESC')) {
      sort.value.descending = true
      sort.value.field = firstField.split(' ')[0]
    } else {
      sort.value.field = firstField
    }
  }
}, { immediate: true })

function keyColumns(frame: any) {
  const foreignFrames = getForeignFrames(frame)
  const keys: Record<string, number> = {}
  if (foreignFrames === undefined) return keys
  for (const ff of Object.values(foreignFrames) as any[][]) {
    for (const f of ff) {
      keys[f.relColumn] = frame.columns.findIndex(({ name }: any) => name === f.relColumn)
    }
  }
  return keys
}

function getForeignFrames(frame: any) {
  return indexFrames.value[frame.ref]
}

function tabelifyFrame(frame: any): { rows: any[]; header: any[]; colgroups: any[] } {
  const outRows: any[] = []
  const isLocal = frame.ref === localDataframe.value.ref

  const selectedCols = new Set<number>()
  for (const c of (options.value?.columns?.[frame.ref]) || []) {
    selectedCols.add(frame.columns.findIndex(({ name }: any) => name === c.name))
  }

  const hSeanFrames: Record<string, boolean> = {}
  const outHeader = [...selectedCols].map((index: number) => {
    const column = frame.columns[index]
    const columnName = column ? column.name : ''
    return {
      column,
      meta: {
        ref: frame.ref,
        sortKey: isLocal ? columnName : `${frame.ref}.${columnName}`,
        tmp_noSort: !isLocal,
      },
    }
  })

  const outColgroups = [{ size: outHeader.length, isLocal }]

  const relFrames = getForeignFrames(frame)
  const usedKeys = keyColumns(frame)

  for (const r of frame.rows || []) {
    let maxSize = 1

    const row = tabelifyRow(r, [...selectedCols])

    const auxRows: any[] = []
    for (const colIndex of Object.values(usedKeys) as number[]) {
      const relFrame = relFrames[r[colIndex]]
      for (const rf of relFrame || []) {
        maxSize = Math.max(maxSize, rf.rows.length)
      }
      for (const rf of relFrame || []) {
        const aux = tabelifyFrame(rf)
        if (!hSeanFrames[rf.ref]) {
          const x = [...aux.header]
          x[0].sourceName = rf.ref
          outHeader.push(...x)
          hSeanFrames[rf.ref] = true
          outColgroups.push(...aux.colgroups)
        }
        if (aux.rows.length < maxSize) {
          for (const col of aux.rows[aux.rows.length - 1]) {
            col.attrs = { rowspan: (maxSize - aux.rows.length) + 1 }
          }
        }
        auxRows.push(aux.rows)
      }
    }

    if (auxRows.length > 0) {
      for (const c of row) {
        c.attrs = { rowspan: maxSize }
      }
      if (row.length) {
        row[0].separator = true
      }
      const merged = mergeRows(auxRows).pop()
      row.push(...merged[0])
      outRows.push(row)
      outRows.push(...merged.slice(1))
    } else {
      outRows.push(row)
    }
  }

  return { rows: outRows, header: outHeader, colgroups: outColgroups }
}

function mergeRows(auxRows: any[], _maxSize?: number): any[] {
  if (auxRows.length <= 1) return auxRows
  const a = auxRows[auxRows.length - 2]
  const b = auxRows[auxRows.length - 1]
  const tmpRows: any[] = []
  for (let i = 0; i < a.length; i++) {
    const row = a[i]
    if (i >= b.length) {
      tmpRows.push(row)
      continue
    }
    row.push(...b[i])
    tmpRows.push(row)
  }
  if (b.length > a.length) {
    tmpRows.push(...b.slice(a.length))
  }
  auxRows.splice(auxRows.length - 2, 2, tmpRows)
  return mergeRows(auxRows)
}

function tabelifyRow(row: any[], selectedCols: number[] = []) {
  return selectedCols.map(index => ({ value: row[index] }))
}

function handleSort(fieldName: string) {
  let relatedDatasource: string | undefined
  if (fieldName) {
    const { field, descending } = sort.value
    if (cursor.value) {
      cursor.value = undefined
      cursors.value = []
    }
    if (fieldName === field) {
      sort.value.descending = !descending
    } else {
      sort.value.field = fieldName
      sort.value.descending = false
    }
  }
  updateDefinition(relatedDatasource)
}

function goToPage(dir?: string) {
  switch (dir) {
    case 'nextPage':
      if (!cursors.value.length) cursors.value.push(undefined)
      cursors.value.push(nextPage.value)
      cursor.value = nextPage.value
      break
    case 'prevPage':
      cursors.value.pop()
      cursor.value = cursors.value.pop()
      break
    default:
      cursor.value = undefined
      cursors.value = []
  }
  updateDefinition()
}

function updateDefinition(updatedDatasource?: string) {
  if (localDataframe.value) {
    let ref = localDataframe.value.ref
    if (updatedDatasource) {
      ref = (dataframes.value.find(({ ref }: any) => ref === updatedDatasource) || {}).ref
    }
    if (ref) {
      const def: any = { ref }
      const { field, descending } = sort.value
      if (field) {
        def.sort = descending ? `${field} DESC` : field
      }
      if (cursor.value) {
        let { limit } = localDataframe.value.paging || {}
        if (!limit) {
          limit = options.value?.datasources?.[0]?.paging?.limit || 20
        }
        def.paging = { cursor: cursor.value, limit }
      }
      const definition: Record<string, any> = {}
      definition[ref] = def
      emit('update', definition)
    }
  }
}

function exportCSV() {
  const header = tabelify.value.header.map(({ column }: any) => {
    const value = column ? column.label : ''
    return value.includes(',') ? `"${value}"` : value
  })
  const rows = tabelify.value.rows.map((row: any[]) => row.map(({ value }: any) => value.includes(',') ? `"${value}"` : value))
  const csv = [header, ...rows].map((row: string[]) => row.join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${props.displayElement?.name || 'report'}-${new Date().toISOString()}.csv`
  a.click()
  URL.revokeObjectURL(url)
}
</script>

<style scoped lang="scss">
.foreign {
  border: 1px solid rgba(var(--black), 0.8);
  border-top: 0;
  border-bottom: 0;
}

.separator {
  border-top: 2px solid rgba(var(--black), 0.5);
}

.card-rounded {
  border-radius: 1rem 1rem 0 0;
}
</style>
