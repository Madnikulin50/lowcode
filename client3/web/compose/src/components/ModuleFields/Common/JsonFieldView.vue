<template>
  <span
    v-if="parsed.empty"
    class="text-muted small"
  >{{ emptyLabel }}</span>
  <span
    v-else-if="!parsed.ok"
    class="text-danger small"
    :title="parsed.raw"
  >{{ invalidLabel }}</span>
  <pre
    v-else-if="layout === 'pretty'"
    class="json-pretty mb-0"
  >{{ pretty }}</pre>
  <dl
    v-else-if="layout === 'kv' && isPlainObject"
    class="json-kv mb-0"
  >
    <div
      v-for="(val, key) in parsed.value"
      :key="key"
      class="json-kv-row"
    >
      <dt>{{ key }}</dt>
      <dd>{{ formatCell(val) }}</dd>
    </div>
  </dl>
  <div
    v-else-if="layout === 'table'"
    class="table-responsive"
  >
    <table class="table table-sm table-borderless align-middle mb-0 json-table">
      <thead>
        <tr>
          <th
            v-for="col in table.columns"
            :key="col"
          >{{ col }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(row, i) in table.rows"
          :key="i"
        >
          <td
            v-for="col in table.columns"
            :key="col"
          >{{ row.cells[col] }}</td>
        </tr>
      </tbody>
    </table>
  </div>
  <div
    v-else
    class="json-chips"
  >
    <span
      v-for="(item, i) in items"
      :key="i"
      class="json-chip badge rounded-pill"
      :class="`text-bg-${itemVariant(item, opts)}`"
      :title="itemTooltip(item)"
    >{{ formatItem(item, opts) }}</span>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import {
  parseJSONValue,
  jsonOptions,
  asItems,
  formatItem,
  itemTooltip,
  itemVariant,
  tableRows,
  prettyJSON,
} from 'corteza-webapp-compose/src/lib/json-field'

const props = defineProps({
  value: { type: [String, Array, Object, Number, Boolean], default: '' },
  field: { type: Object, default: () => ({}) },
  emptyLabel: { type: String, default: '' },
  invalidLabel: { type: String, default: '' },
})

const opts = computed(() => jsonOptions(props.field))
const parsed = computed(() => parseJSONValue(props.value))
const layout = computed(() => opts.value.layout)
const items = computed(() => asItems(parsed.value.value).filter(item => item != null && item !== ''))
const isPlainObject = computed(() => parsed.value.value && typeof parsed.value.value === 'object' && !Array.isArray(parsed.value.value))
const table = computed(() => tableRows(parsed.value.value, opts.value))
const pretty = computed(() => prettyJSON(parsed.value.value))

function formatCell (val) {
  if (val == null) return ''
  if (typeof val === 'object') return prettyJSON(val)
  return String(val)
}
</script>

<style lang="scss" scoped>
.json-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.28rem;
  align-items: center;
}

.json-chip {
  font-family: var(--font-medium), ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.01em;
  padding: 0.28em 0.65em;
  line-height: 1.35;
  white-space: nowrap;
}

.json-pretty {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.78rem;
  background: var(--extra-light, #f8f9fc);
  border-radius: 0.4rem;
  padding: 0.6rem 0.75rem;
  max-height: 16rem;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
}

.json-table {
  font-size: 0.8rem;

  th {
    color: var(--primary);
    font-weight: 600;
    text-transform: none;
  }

  td {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  }
}

.json-kv {
  display: grid;
  gap: 0.2rem;
  margin: 0;
}

.json-kv-row {
  display: grid;
  grid-template-columns: minmax(5rem, 30%) 1fr;
  gap: 0.5rem;
  font-size: 0.82rem;

  dt {
    color: var(--secondary);
    font-weight: 600;
    margin: 0;
  }

  dd {
    margin: 0;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    word-break: break-word;
  }
}
</style>
