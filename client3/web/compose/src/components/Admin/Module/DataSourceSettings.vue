<template>
  <div v-if="module">
    <div class="d-flex justify-content-end mb-2">
      <button
        class="btn btn-outline-primary btn-sm"
        type="button"
        :disabled="previewing"
        @click="runPreview"
      >
        <span
          v-if="previewing"
          class="spinner-border spinner-border-sm me-1"
        />
        {{ t('preview.button') }}
      </button>
    </div>

    <configurator
      v-if="module"
      :items="items"
      :current-index="dsState.currentIndex"
      draggable
      @select="setCurrentDatasource"
      @add="openDatasourceSelector()"
      @delete="deleteCurrentDataSource"
    >
      <template #label="{ item: { step } }">
        <span class="d-inline-block text-truncate">
          {{ datasourceLabel(step, dsState.currentIndex) }}
        </span>
      </template>
      <template #configurator>
        <component
          :is="getDatasourceComponent(items[dsState.currentIndex])"
          v-if="currentDatasourceStep"
          :index="dsState.currentIndex"
          :datasources="items"
          v-model:step="currentDatasourceStep"
          :creating="items[dsState.currentIndex].meta.creating"
        />
      </template>
    </configurator>

    <div
      id="datasource-selector-modal"
      ref="selectorModal"
      class="modal fade"
      tabindex="-1"
    >
      <div class="modal-dialog modal-lg modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t('add.datasource') }}</h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
            />
          </div>
          <div class="modal-body px-0 py-3">
            <selector
              :items="dsState.types"
              display-mode="text"
              @select="addDatasource"
            />
          </div>
        </div>
      </div>
    </div>

    <div
      ref="previewModalEl"
      class="modal fade"
      tabindex="-1"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-xl modal-dialog-scrollable">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ t('preview.title') }}</h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              aria-label="Close"
            />
          </div>
          <div class="modal-body">
            <p
              v-if="previewError"
              class="text-danger small mb-0"
            >
              {{ previewError }}
            </p>
            <div
              v-else-if="previewColumns.length"
              class="preview-table-wrap"
            >
              <table class="table record-list-table table-hover mb-0">
                <thead>
                  <tr class="text-muted small">
                    <th
                      v-for="c in previewColumns"
                      :key="c.name"
                    >
                      <div class="d-flex align-items-center text-nowrap">
                        {{ c.label || c.name }}
                      </div>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(row, ri) in previewRows"
                    :key="ri"
                  >
                    <td
                      v-for="(cell, ci) in row"
                      :key="ci"
                      class="record-value"
                    >
                      {{ cell }}
                    </td>
                  </tr>
                  <tr v-if="!previewRows.length">
                    <td
                      :colspan="previewColumns.length"
                      class="text-muted small"
                    >
                      {{ t('preview.empty') }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p
              v-else
              class="text-muted small mb-0"
            >
              {{ t('preview.empty') }}
            </p>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-secondary"
              data-bs-dismiss="modal"
            >
              {{ t('preview.close') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="js">
defineOptions({ i18nOptions: { namespaces: 'module', keyPrefix: 'edit.config.datasource' } })
import { ref, reactive, computed, onMounted, onBeforeUnmount, getCurrentInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import { Modal } from 'bootstrap'
import { compose, NoID, reporter } from 'corteza-lib/js/dist'
import datasources from 'corteza-webapp-compose/src/components/Admin/Module/Datasources/loader'
import Configurator from 'corteza-webapp-compose/src/components/Common/Configurator'
import Selector from 'corteza-webapp-compose/src/components/Common/Selector'
import * as displayElementThumbnails from 'corteza-webapp-compose/src/assets/DisplayElements'
import { cloneDeep } from 'lodash'

const prefixed$ = 'edit.config.datasource.'
const { t: $t } = useI18n()
const t = (key) => $t(prefixed$ + key)
const tG = (key) => $t(key)

const { $ComposeAPI } = getCurrentInstance().appContext.config.globalProperties

const previewModalEl = ref(undefined)
const previewModal = ref(undefined)
const previewing = ref(false)
const previewColumns = ref([])
const previewRows = ref([])
const previewError = ref('')

const props = defineProps({
  module: {
    type: compose.Module,
    required: true,
  },
})

const processing = ref(false)
const connections = ref([])

const displayElements = reactive({
  showSelector: false,
  currentIndex: undefined,
  types: [
    { label: t('display-elements.types.text'), kind: 'Text', value: displayElementThumbnails.Text },
    { label: t('display-elements.types.metric'), kind: 'Metric', value: displayElementThumbnails.Metric },
    { label: t('display-elements.types.table'), kind: 'Table', value: displayElementThumbnails.Table },
    { label: t('display-elements.types.chart'), kind: 'Chart', value: displayElementThumbnails.Chart },
  ],
})

const dsState = reactive({
  showSelector: false,
  showConfigurator: false,
  processing: false,
  currentIndex: undefined,
  tempItems: [],
  types: [
    { label: t('types.load.label'), kind: 'Load', value: t('types.load.data-from-resource') },
    { label: t('types.link.label'), kind: 'Link', value: t('types.link.load-datasources') },
    { label: t('types.join.label'), kind: 'Join', value: t('types.join.load-datasources') },
    { label: t('types.aggregate.label'), kind: 'Aggregate', value: t('types.aggregate.load-datasource') },
  ],
})

const editor = ref(undefined)

const items = computed({
  get () {
    const dataSource = props.module.config.dataSource ?? {}
    return dataSource.items ?? []
  },
  set (value) {
    props.module.config.dataSource = props.module.config.dataSource ?? {}
    props.module.config.dataSource = { items: value }
  },
})

const currentDatasourceStep = computed({
  get () {
    return dsState.currentIndex !== undefined ? items.value[dsState.currentIndex].step : undefined
  },
  set (step) {
    if (dsState.currentIndex !== undefined) {
      items.value[dsState.currentIndex].step = step
    }
  },
})

onMounted(() => {
  dsState.showConfigurator = true
  const dataSource = props.module.config.dataSource ?? {}
  items.value = cloneDeep(dataSource.items ?? []).map(ds => {
    ds.meta.creating = false
    return ds
  })
  dsState.currentIndex = items.value.length ? 0 : undefined

  previewModal.value = new Modal(previewModalEl.value)
  previewModalEl.value.addEventListener('hidden.bs.modal', () => {
    previewColumns.value = []
    previewRows.value = []
    previewError.value = ''
  })
})

onBeforeUnmount(() => {
  setDefaultValues()
  previewModal.value?.dispose()
})

function setDefaultValues () {
  processing.value = false
  connections.value = []
  dsState.currentIndex = undefined
  dsState.showSelector = false
}

function getDatasourceComponent ({ step }) {
  let datasourceComponent
  if (step) {
    for (const s in step) {
      datasourceComponent = datasources(s)
      if (datasourceComponent) break
    }
  }
  return datasourceComponent
}

function datasourceLabel (datasource, currentIndex) {
  for (const v of Object.values(datasource)) {
    if (v && v.name) return v.name
  }
  return `${tG('datasources.source')} ${currentIndex}`
}

function openDatasourceSelector () {
  dsState.showSelector = true
}

function setCurrentDatasource (index) {
  dsState.currentIndex = index
}

function deleteCurrentDataSource () {
  items.value.splice(dsState.currentIndex, 1)
  dsState.currentIndex = items.value.length ? 0 : undefined
}

function addDatasource (kind = '') {
  if (kind) {
    let step
    switch (kind) {
      case 'Aggregate':
        step = reporter.StepFactory({
          aggregate: { name: 'Aggregate', keys: [], columns: [], filter: {}, sort: '' },
        })
        break
      case 'Link':
        step = reporter.StepFactory({
          link: { name: 'Link', foreignColumn: '', foreignSource: '', localColumn: '', localSource: '' },
        })
        break
      case 'Join':
        step = reporter.StepFactory({
          join: { name: 'Join', foreignColumn: '', foreignSource: '', localColumn: '', localSource: '' },
        })
        break
      default:
        step = reporter.StepFactory({
          load: { name: 'Load', source: 'composeRecords', definition: {}, filter: {}, sort: '' },
        })
    }
    const newItems = [...items.value]
    newItems.push({ step, meta: {} })
    items.value = newItems
  }

  dsState.currentIndex = items.value.length - 1
  dsState.showSelector = false
  dsState.showConfigurator = true
}

function runPreview () {
  previewing.value = true
  previewError.value = ''
  previewColumns.value = []
  previewRows.value = []
  $ComposeAPI.datasourcePreview({
    namespaceID: props.module.namespaceID || NoID,
    moduleID: props.module.moduleID || NoID,
    datasource: props.module.config.dataSource ?? {},
    limit: 20,
  }).then((frame) => {
    previewColumns.value = (frame?.columns || [])
    previewRows.value = (frame?.rows || [])
    previewModal.value?.show()
  }).catch((e) => {
    previewError.value = e.message || String(e)
    previewModal.value?.show()
  }).finally(() => {
    previewing.value = false
  })
}
</script>

<style lang="scss">
.preview-table-wrap {
  border-radius: 0.5rem;
  border: 1px solid var(--bs-border-color, #dee2e6);
}

.record-list-table {
  border-collapse: separate;
  border-spacing: 0;

  thead {
    th {
      background: #f8f9fa;
      border-bottom: 2px solid var(--bs-border-color, #dee2e6);
      font-size: 0.8125rem;
      font-weight: 600;
      color: var(--bs-secondary-color, #6c757d);
      text-transform: uppercase;
      letter-spacing: 0.025em;
      padding: 0.625rem 0.75rem;
      white-space: nowrap;

      position: sticky;
      top: 0;
      z-index: 2;
    }
  }

  tbody {
    td {
      padding: 0.625rem 0.75rem;
      vertical-align: middle;
      border-bottom: 1px solid var(--bs-border-color-translucent, rgba(0,0,0,0.05));
      font-size: 0.875rem;
    }

    tr:last-child td {
      border-bottom: none;
    }

    tr:hover {
      background-color: rgba(var(--bs-primary-rgb, 13 110 253), 0.03);
    }
  }
}
</style>
