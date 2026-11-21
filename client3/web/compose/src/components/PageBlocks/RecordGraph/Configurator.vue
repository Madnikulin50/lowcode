<template>
  <div>
    <h5>{{ $t('recordGraph.label') }}</h5>

    <div
      v-if="!pageModuleID"
      class="alert alert-warning"
    >
      {{ $t('recordGraph.recordPageOnly') }}
    </div>

    <template v-else>
      <div class="row g-3 mb-3">
        <div class="col-12 col-lg-4">
          <label class="form-label text-primary">{{ $t('recordGraph.depth.label') }}</label>
          <select
            v-model.number="options.depth"
            class="form-select"
          >
            <option :value="0">
              {{ $t('recordGraph.depth.0') }}
            </option>
            <option :value="1">
              {{ $t('recordGraph.depth.1') }}
            </option>
            <option :value="2">
              {{ $t('recordGraph.depth.2') }}
            </option>
          </select>
          <small class="text-muted">{{ $t('recordGraph.depth.footnote') }}</small>
        </div>

        <div class="col-12 col-lg-4">
          <label class="form-label text-primary">{{ $t('recordGraph.maxNodes.label') }}</label>
          <input
            v-model.number="options.maxNodes"
            type="number"
            :min="limits.maxNodes.min"
            :max="limits.maxNodes.max"
            class="form-control"
          >
          <small class="text-muted">{{ $t('recordGraph.maxNodes.footnote') }}</small>
        </div>

        <div class="col-12 col-lg-4">
          <label class="form-label text-primary">{{ $t('recordGraph.perModuleLimit.label') }}</label>
          <input
            v-model.number="options.perModuleLimit"
            type="number"
            :min="limits.perModuleLimit.min"
            :max="limits.perModuleLimit.max"
            class="form-control"
          >
          <small class="text-muted">{{ $t('recordGraph.perModuleLimit.footnote') }}</small>
        </div>
      </div>

      <div class="row g-3 mb-3">
        <div class="col-12 col-lg-4">
          <label class="form-label text-primary">{{ $t('recordGraph.layout.label') }}</label>
          <select
            v-model="options.layout"
            class="form-select"
          >
            <option
              v-for="l in layouts"
              :key="l"
              :value="l"
            >
              {{ $t(`recordGraph.layout.${l}`) }}
            </option>
          </select>
          <small class="text-muted">{{ $t('recordGraph.layout.footnote') }}</small>
        </div>

        <div class="col-12 col-lg-4">
          <label class="form-label text-primary">{{ $t('recordGraph.clusterFrom.label') }}</label>
          <input
            v-model.number="options.clusterFrom"
            type="number"
            :min="limits.clusterFrom.min"
            :max="limits.clusterFrom.max"
            class="form-control"
          >
          <small class="text-muted">{{ $t('recordGraph.clusterFrom.footnote') }}</small>
        </div>

        <div class="col-12 col-lg-4 d-flex align-items-end">
          <div class="form-check form-switch mb-2">
            <input
              id="rg-toolbar"
              v-model="options.showToolbar"
              class="form-check-input"
              type="checkbox"
            >
            <label
              class="form-check-label"
              for="rg-toolbar"
            >{{ $t('recordGraph.showToolbar') }}</label>
          </div>
        </div>
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('recordGraph.displayOption.label') }}</label>
        <select
          v-model="options.displayOption"
          class="form-select"
        >
          <option
            v-for="o in displayOptions"
            :key="o"
            :value="o"
          >
            {{ $t(`recordGraph.displayOption.${o}`) }}
          </option>
        </select>
        <div class="form-text">
          {{ $t('recordGraph.displayOption.footnote') }}
        </div>
      </div>

      <div class="mb-3">
        <div class="form-check form-switch">
          <input
            id="rg-parents"
            v-model="options.showParents"
            class="form-check-input"
            type="checkbox"
          >
          <label
            class="form-check-label"
            for="rg-parents"
          >{{ $t('recordGraph.showParents') }}</label>
        </div>
        <div class="form-check form-switch">
          <input
            id="rg-children"
            v-model="options.showChildren"
            class="form-check-input"
            type="checkbox"
          >
          <label
            class="form-check-label"
            for="rg-children"
          >{{ $t('recordGraph.showChildren') }}</label>
        </div>
        <div class="form-check form-switch">
          <input
            id="rg-legend"
            v-model="options.showLegend"
            class="form-check-input"
            type="checkbox"
          >
          <label
            class="form-check-label"
            for="rg-legend"
          >{{ $t('recordGraph.showLegend') }}</label>
        </div>
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('recordGraph.labels.label') }}</label>
        <div class="form-text mb-2">
          {{ $t('recordGraph.labels.footnote') }}
          {{ $t('recordGraph.labels.example') }}: <code v-pre>{{position_number}} · {{work_name}}</code>
        </div>

        <div
          v-for="m in modulesInGraph"
          :key="m.moduleID"
          class="border rounded p-2 mb-2"
        >
          <div class="fw-semibold mb-1">
            {{ m.name || m.handle }}
          </div>
          <input
            :value="templateOf(m.moduleID)"
            class="form-control font-monospace"
            :placeholder="placeholder"
            spellcheck="false"
            @input="setTemplate(m.moduleID, $event.target.value)"
          >
          <div class="d-flex flex-wrap gap-1 mt-1">
            <button
              v-for="f in fieldsOf(m)"
              :key="f.name"
              type="button"
              class="btn btn-sm btn-outline-secondary py-0"
              :title="f.kind"
              @click="addField(m.moduleID, f.name)"
            >
              {{ f.label || f.name }}
            </button>
          </div>
          <div
            v-if="selectFields(m).length"
            class="mt-2"
          >
            <label class="form-label small text-primary mb-0">{{ $t('recordGraph.statusField.label') }}</label>
            <select
              :value="statusOf(m.moduleID)"
              class="form-select form-select-sm"
              @change="setStatus(m.moduleID, $event.target.value)"
            >
              <option value="">
                {{ $t('recordGraph.statusField.none') }}
              </option>
              <option
                v-for="f in selectFields(m)"
                :key="f.name"
                :value="f.name"
              >
                {{ f.label || f.name }}
              </option>
            </select>
          </div>
        </div>
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('recordGraph.exclude.label') }}</label>
        <div
          v-if="candidates.length"
          class="d-flex flex-wrap gap-3"
        >
          <div
            v-for="m in candidates"
            :key="m.moduleID"
            class="form-check"
          >
            <input
              :id="`rg-ex-${m.moduleID}`"
              class="form-check-input"
              type="checkbox"
              :checked="options.excludeModules.includes(m.moduleID)"
              @change="toggle(m.moduleID)"
            >
            <label
              class="form-check-label"
              :for="`rg-ex-${m.moduleID}`"
            >{{ m.name || m.handle }}</label>
          </div>
        </div>
        <span
          v-else
          class="text-muted"
        >{{ $t('recordGraph.exclude.none') }}</span>
        <div class="form-text">
          {{ $t('recordGraph.exclude.footnote') }}
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import { useStore } from '../../../store'
import { reachableModules } from 'corteza-webapp-compose/src/lib/record-graph'

const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  block: { type: Object, required: true },
  module: { type: Object, required: false, default: undefined },
  page: { type: Object, required: false, default: undefined },
})

const store = useStore()
const options = computed(() => props.block.options)
const limits = compose.PageBlockRecordGraphLimits
const displayOptions = ['sameTab', 'newTab', 'modal', 'doNothing']
const layouts = ['force', 'radial', 'layers']

// the module of the record page the block is on
const pageModuleID = computed(() => props.module?.moduleID || props.page?.moduleID || '')

// the modules a graph around this module can reach (two steps), to name and to leave out
const modulesInGraph = computed(() => reachableModules(store.module.set, pageModuleID.value))
const candidates = computed(() => modulesInGraph.value.filter(m => m.moduleID !== pageModuleID.value))

// ---- how the records of a module are named

// not in a template: the double braces would be taken for a Vue interpolation
const placeholder = ['{{', '}}'].join('field')

const entryOf = moduleID => options.value.labels.find(l => l.moduleID === moduleID)
const templateOf = moduleID => (entryOf(moduleID) || {}).template || ''
const statusOf = moduleID => (entryOf(moduleID) || {}).statusField || ''

// one entry per module holds its template and its status field; it goes when both are empty
function setEntry (moduleID, patch) {
  const labels = options.value.labels
  const i = labels.findIndex(l => l.moduleID === moduleID)
  const next = { moduleID, template: '', statusField: '', ...(i >= 0 ? labels[i] : {}), ...patch }

  if (!next.template.trim() && !next.statusField) {
    if (i >= 0) labels.splice(i, 1)
  } else if (i >= 0) {
    labels.splice(i, 1, next)
  } else {
    labels.push(next)
  }
}

const setTemplate = (moduleID, template) => setEntry(moduleID, { template })
const setStatus = (moduleID, statusField) => setEntry(moduleID, { statusField })

// the Select fields: their colours can ring the nodes
const selectFields = m => m.fields.filter(f => f.kind === 'Select' && !f.isMulti)

// a click on a field adds it to the template, after a separator when there is text already
function addField (moduleID, name) {
  const current = templateOf(moduleID)
  setTemplate(moduleID, `${current}${current.trim() ? ' · ' : ''}{{${name}}}`)
}

// the fields that can go into a name: readable ones, not references or files
const fieldsOf = m => [...m.fields.filter(f => !['Record', 'File', 'User', 'Geometry'].includes(f.kind)), { name: 'createdAt', label: 'createdAt', kind: 'DateTime' }]

function toggle (moduleID) {
  const list = options.value.excludeModules
  const i = list.indexOf(moduleID)
  if (i >= 0) list.splice(i, 1)
  else list.push(moduleID)
}
</script>
