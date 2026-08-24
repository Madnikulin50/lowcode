<template>
  <div
    v-if="namespace"
    class="py-3 d-flex flex-column flex-grow-1"
    style="min-height: 0"
  >
    <Teleport to="#topbar-title">
      {{ $t('edit.title') }}
    </Teleport>

    <div
      v-if="loading"
      class="d-flex align-items-center justify-content-center h-100"
    >
      <span class="spinner-border" />
    </div>

    <div
      v-else
      class="d-flex flex-column flex-grow-1"
      style="min-height: 0"
      @submit.prevent="handleSave"
    >
      <div class="container-fluid flex-grow-1 d-flex flex-column" style="min-height: 0">
      <div class="row flex-grow-1" style="min-height: 0">
        <div class="col d-flex flex-column" style="min-height: 0">
          <div class="card shadow-sm d-flex flex-column flex-grow-1" style="min-height: 0">
            <div class="card-header d-flex py-3 align-items-center border-bottom gap-1">
              <export
                v-if="namespace.canExportCharts"
                :list="[chart]"
                type="chart"
              />

              <c-permissions-button
                v-if="namespace.canGrant"
                :title="chart.name || chart.handle || chart.chartID"
                :target="chart.name || chart.handle || chart.chartID"
                :resource="`corteza::compose:chart/${namespace.namespaceID}/${chart.chartID}`"
                :button-label="$t('label.permissions')"
                class="btn-lg"
              />
            </div>

            <div class="overflow-auto" style="flex: 1 1 0%; min-height: 0;">
            <div class="row pb-5">
              <div
                class="col-12 col-lg-7 border-end"
              >
                <div class="pt-3 px-3">
                  <h5>
                    {{ $t('generalSettings') }}
                  </h5>
                  <div
                    v-if="modules"
                    class="row"
                  >
                    <div class="col-12 col-lg-6">
                      <div class="mb-3">
                        <label class="form-label text-primary">
                          {{ $t('name') }}
                        </label>
                        <input
                          v-model="chart.name"
                          class="form-control"
                          :class="{ 'is-invalid': nameState === false }"
                          :placeholder="$t('placeholder.name')"
                        />
                      </div>
                    </div>

                    <div class="col-12 col-lg-6">
                      <div class="mb-3">
                        <label class="form-label text-primary">
                          {{ $t('handle') }}
                        </label>
                        <input
                          v-model="chart.handle"
                          class="form-control mb-1"
                          :class="{ 'is-invalid': handleState === false }"
                          :placeholder="$t('placeholder.handle')"
                        />
                        <div
                          v-if="handleState === false"
                          class="invalid-feedback d-block"
                        >
                          {{ $t('placeholder.invalid-handle-characters') }}
                        </div>
                      </div>
                    </div>

                    <div class="col-12 col-lg-6">
                      <div class="mb-3">
                        <label class="form-label text-primary">
                          {{ $t('colorScheme.label') }}
                        </label>
                        <div class="input-group d-flex w-100">
                          <c-input-select
                            v-model="chart.config.colorScheme"
                            :options="colorSchemes"
                            :reduce="cs => cs.id"
                            label="name"
                            :get-option-key="o => o.id"
                            :placeholder="$t('colorScheme.placeholder')"
                          >
                            <template #option="option">
                              <p class="mb-1">
                                {{ option.name }}
                              </p>
                              <div
                                v-for="(color, index) in option.colors"
                                :key="index"
                                :style="`background: ${color};`"
                                class="d-inline-block color-box me-1 mb-1"
                              />
                            </template>

                            <template
                              v-if="canManageColorSchemes"
                              #list-header
                            >
                              <li class="border-bottom text-center mb-1">
                                <button
                                  class="btn btn-link text-decoration-none"
                                  @click="createColorScheme"
                                >
                                  {{ $t('colorScheme.custom.add') }}
                                </button>
                              </li>
                            </template>
                          </c-input-select>

                          <button
                            v-if="showEditColorSchemeButton"
                            data-bs-toggle="tooltip"
                            :title="$t('colorScheme.custom.edit')"
                            class="btn btn-extra-light d-flex align-items-center"
                            @click="editColorScheme()"
                          >
                            <font-awesome-icon :icon="['far', 'edit']" />
                          </button>
                        </div>

                        <template
                          v-if="currentColorScheme"
                        >
                          <div
                            v-for="(color, index) in currentColorScheme.colors"
                            :key="`${currentColorScheme.value}-${index}`"
                            :style="`background: ${color};`"
                            class="d-inline-block color-box me-1"
                          />
                        </template>
                      </div>
                    </div>

                    <div class="col-12 col-lg-6 mt-2 mt-md-0">
                      <div class="mb-3">
                        <label class="form-label text-primary">
                          {{ $t('edit.animation.label') }}
                        </label>
                        <c-input-checkbox
                          v-model="chart.config.noAnimation"
                          :labels="checkboxLabel"
                          switch
                          invert
                        />
                      </div>
                    </div>
                  </div>
                </div>
                <hr v-if="modules">

                <component
                  :is="reportEditor"
                  v-if="chart && editReport"
                  v-model:report="editReport"
                  :chart="chart"
                  :modules="modules"
                  :supported-metrics="1"
                />

                <hr>

                <div class="px-3">
                  <h5 class="mb-3">
                    {{ $t('edit.toolbox.label') }}
                  </h5>

                  <div class="row">
                    <div class="col-12 col-lg-6">
                      <div class="mb-3">
                        <label class="form-label text-primary">
                          {{ $t('edit.toolbox.saveAsImage.label') }}
                        </label>
                        <c-input-checkbox
                          v-model="chart.config.toolbox.saveAsImage"
                          switch
                          :labels="checkboxLabel"
                        />
                      </div>

                      <div class="mb-3">
                        <label class="form-label text-primary">
                          {{ $t('edit.toolbox.showDataTable.label') }}
                        </label>
                        <c-input-checkbox
                          v-model="chart.config.toolbox.showDataTable"
                          switch
                          :labels="checkboxLabel"
                        />
                      </div>
                    </div>

                    <div
                      v-if="hasAxis"
                      class="col-12 col-lg-6"
                    >
                      <div class="mb-3">
                        <label class="form-label text-primary">
                          {{ $t('edit.toolbox.timeline.label') }}
                        </label>
                        <div
                          class="btn-group"
                          data-bs-toggle="buttons"
                        >
                          <label
                            v-for="opt in timelineOptions"
                            :key="opt.value"
                            class="btn btn-outline-secondary btn-sm"
                            :class="{ active: chart.config.toolbox.timeline === opt.value }"
                          >
                            <input
                              v-model="chart.config.toolbox.timeline"
                              type="radio"
                              class="btn-check"
                              autocomplete="off"
                              :value="opt.value"
                            />
                            {{ opt.text }}
                          </label>
                      </div>
                    </div>


                  </div>
                </div>
                  <div class="col-12 col-lg-6 mt-2 mt-md-0">
                      <div v-if="hasGradient" class="mb-3">
                        <label class="form-label text-primary">
                          {{ $t('edit.gradient.label', 'Gradient') }}
                        </label>
                        <c-input-select
                          v-model="chart.config.gradient"
                          :options="gradientOptions"
                          :reduce="opt => opt.value"
                          label="text"
                          :clearable="false"
                          :searchable="false"
                          :placeholder="$t('edit.gradient.placeholder', 'None')"
                        />
                      </div>
                    </div>
                </div>
              </div>

              <div
                class="col-12 col-lg-5"
              >
                <div
                  class="d-flex flex-column position-sticky"
                  style="top: 0;"
                >
                  <button
                    data-bs-toggle="tooltip"
                    :title="$t('edit.loadData')"
                    :disabled="processing || !reportsValid"
                    class="btn btn-outline-light btn-lg d-flex align-items-center text-primary ms-auto border-0 px-2 mt-2 me-2"
                    @click.prevent="update"
                  >
                    <font-awesome-icon :icon="['fa', 'sync']" />
                  </button>

                  <chart-component
                    ref="chartComponentRef"
                    :chart="chart"
                    :reporter="reporter"
                    style="min-height: 400px;"
                    @updated="onUpdated"
                  />
                </div>
            </div>
            </div>
          </div>
          </div>
        </div>
      </div>
    </div>

    <div
      class="modal fade"
      :class="{ show: colorSchemeModal.show }"
      :style="{ display: colorSchemeModal.show ? 'block' : 'none' }"
      tabindex="-1"
      style="background-color: rgba(0,0,0,0.5);"
    >
      <div class="modal-dialog modal-dialog-centered modal-md">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">
              {{ colorSchemeModalTitle }}
            </h5>
            <button
              type="button"
              class="btn-close"
              @click="closeColorSchemeModal"
            />
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label text-primary">
                {{ $t('colorScheme.custom.modal.name.label') }}
              </label>
              <input
                v-model="colorSchemeModal.colorscheme.name"
                class="form-control"
              />
            </div>

            <div class="mb-0">
              <label class="form-label text-primary">
                {{ $t('colorScheme.custom.modal.colors.label') }}
                <button
                  class="btn btn-outline-light btn-sm text-primary border-0"
                  @click="addColor"
                >
                  <font-awesome-icon :icon="['fa', 'plus']" />
                </button>
              </label>

              <c-input-color-picker
                v-for="(color, index) in colorSchemeModal.colorscheme.colors"
                :key="index"
                v-model="colorSchemeModal.colorscheme.colors[index]"
                :show-text="false"
                data-test-id="input-scheme-color"
                :translations="{
                  modalTitle: $t('colorScheme.pickAColor'),
                  light: $t('themes.labels.light'),
                  dark: $t('themes.labels.dark'),
                  cancelBtnLabel: $t('label.cancel'),
                  saveBtnLabel: $t('label.saveAndClose')
                }"
                :theme-settings="themeSettings"
                class="d-inline-flex me-1"
              >
                <template #footer>
                  <c-input-confirm
                    variant="danger"
                    size="md"
                    show-icon
                    @confirmed="removeColor(index)"
                  />
                </template>
              </c-input-color-picker>
            </div>
          </div>
          <div class="modal-footer">
            <c-input-confirm
              v-if="colorSchemeModal.edit"
              :disabled="colorSchemeModal.processing"
              variant="danger"
              size="md"
              show-icon
              @confirmed="deleteColorScheme()"
            />

            <button
              class="btn btn-outline-secondary ms-auto"
              :disabled="colorSchemeModal.processing"
              @click="closeColorSchemeModal"
            >
              {{ $t('label.cancel') }}
            </button>

            <button
              class="btn btn-primary"
              :disabled="!colorSchemeModal.colorscheme.name || !colorSchemeModal.colorscheme.colors.length || colorSchemeModal.processing"
              @click="saveColorScheme"
            >
              {{ $t('label.saveAndClose') }}
            </button>
          </div>
        </div>
      </div>
    </div>
    </div>

    <Teleport to="#admin-toolbar">
      <editor-toolbar
        :processing="processing"
        :processing-save="processingSave"
        :processing-clone="processingClone"
        :processing-save-and-close="processingSaveAndClose"
        :processing-delete="processingDelete"
        :hide-delete="hideDelete"
        :hide-save="hideSave"
        :hide-clone="!isEdit"
        :disable-save="disableSave"
        @delete="handleDelete()"
        @save="handleSave()"
        @clone="handleClone()"
        @saveAndClose="handleSave({ closeOnSuccess: true })"
        @back="router.push(previousPage || { name: 'admin.charts' })"
      />
    </Teleport>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'chart' } })
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'

const chartComponentRef = ref(null)
import { useStore } from '../../../store'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getCurrentInstance } from 'vue'
import { isEqual, debounce } from 'lodash'
import EditorToolbar from 'corteza-webapp-compose/src/components/Admin/EditorToolbar'
import { compose, NoID, shared } from 'corteza-lib/js/dist'
import Export from 'corteza-webapp-compose/src/components/Admin/Export'
import ChartComponent from 'corteza-webapp-compose/src/components/Chart'
import { handle, components, composables } from 'corteza-lib/vue/dist'
import draggable from 'vuedraggable'
import ReportItem from 'corteza-webapp-compose/src/components/Chart/ReportItem'
import Reports from 'corteza-webapp-compose/src/components/Chart/Report'
import { chartConstructor } from 'corteza-webapp-compose/src/lib/charts'
import { evaluatePrefilter } from 'corteza-webapp-compose/src/lib/record-filter'

const { CInputCheckbox, CInputColorPicker } = components
const { colorschemes } = shared

const { t } = useI18n()
const store = useStore()
const router = useRouter()
const route = useRoute()

const { $ComposeAPI, $SystemAPI, $Settings, $auth } = getCurrentInstance().appContext.config.globalProperties

const props = defineProps({
  namespace: { type: compose.Namespace, required: true },
  chartID: { type: String, required: false, default: NoID },
  category: { type: String, required: false, default: '' },
})

const chart = ref(undefined)
const initialChartState = ref(undefined)
const processing = ref(false)
const processingSave = ref(false)
const processingClone = ref(false)
const processingSaveAndClose = ref(false)
const processingDelete = ref(false)
const loading = ref(false)
const editReportIndex = ref(undefined)
const customColorSchemes = ref([])
const colorSchemeModal = ref({
  show: false,
  processing: false,
  edit: false,
  colorscheme: {},
})

const checkboxLabel = computed(() => ({
  on: t('label.yes'),
  off: t('label.no'),
}))

const gradientOptions = [
  { value: '', text: t('edit.gradient.options.none', 'None') },
  { value: 'lightToDark', text: t('edit.gradient.options.lightToDark', 'Light → Dark') },
  { value: 'darkToLight', text: t('edit.gradient.options.darkToLight', 'Dark → Light') },
]

const modules = computed(() => store.getters['module/set'])
const modByID = computed(() => store.getters['module/getByID'])
const previousPage = computed(() => store.getters['ui/previousPage'])
const can = computed(() => store.getters['rbac/can'])

const colorSchemes = computed(() => {
  const getCssVariable = (variableName) => {
    return getComputedStyle(document.documentElement).getPropertyValue(variableName).trim()
  }
  const capitalize = w => `${w[0].toUpperCase()}${w.slice(1)}`
  const splicer = sc => {
    const rr = (/(\D+)(\d+)$/gi).exec(sc)
    return { label: rr[1], count: rr[2] }
  }
  const colors = ['blue', 'indigo', 'purple', 'pink', 'red', 'orange', 'yellow', 'green', 'teal', 'cyan'].map(variable => {
    return getCssVariable(`--${variable}`)
  })
  const rr = []
  const scheme = { id: 'lowcode', name: 'Lowcode Platform', colors }
  rr.push(scheme)
  rr.push(...customColorSchemes.value)
  for (const g in colorschemes) {
    for (const sc in colorschemes[g]) {
      const gn = splicer(sc)
      rr.push({
        id: `${g}.${sc}`,
        name: `${capitalize(g)}: ${capitalize(gn.label)} (${t('colorLabel', gn)})`,
        colors: [...colorschemes[g][sc]],
      })
    }
  }
  return rr
})

const currentColorScheme = computed(() => colorSchemes.value.find(({ id }) => id === chart.value.config.colorScheme))

const canManageColorSchemes = computed(() => can.value('system/', 'settings.manage'))

const colorSchemeModalTitle = computed(() => t(`colorScheme.custom.${colorSchemeModal.value.edit ? 'edit' : 'add'}`))

const defaultReport = { moduleID: undefined, metrics: [{ field: 'count' }], dimensions: [{ field: 'createdAt', modifier: 'MONTH' }] }

const nameState = computed(() => chart.value && chart.value.name.length > 0 ? null : false)

const handleState = computed(() => chart.value ? handle.handleState(chart.value.handle) : null)

const supportsMultipleReports = computed(() => chart.value instanceof compose.FunnelChart)

const reportsValid = computed(() => {
  if (!reports.value) return false
  return !reports.value.find(({ moduleID }) => !moduleID)
})

const reportEditor = computed(() => {
  if (!chart.value) return undefined
  if (chart.value instanceof compose.FunnelChart) return Reports.FunnelChart
  else if (chart.value instanceof compose.GaugeChart) return Reports.GaugeChart
  else if (chart.value instanceof compose.RadarChart) return Reports.RadarChart
  return Reports.GenericChart
})

const reports = computed({
  get () { return chart.value.config.reports },
  set (r) { chart.value.config.reports = r },
})

const editReport = computed({
  get () {
    if (editReportIndex.value !== undefined) return reports.value[editReportIndex.value]
    return undefined
  },
  set (v) { reports.value.splice(editReportIndex.value, 1, v) },
})

const disableSave = computed(() => !chart.value || [nameState.value, handleState.value].includes(false))

const hideDelete = computed(() => !isEdit.value || !chart.value.canDeleteChart || !!chart.value.deletedAt)

const hideSave = computed(() => isEdit.value && !chart.value.canUpdateChart)

const isEdit = computed(() => chart.value && chart.value.chartID !== NoID)

const hasAxis = computed(() => reports.value.some(({ metrics = [] }) => metrics.some(m => ['bar', 'line', 'scatter', 'waterfall', 'boxplot', 'candlestick', 'heatmap', 'parallel'].includes(m.type))))

const hasPie = computed(() => {
  const r = chart.value?.config?.reports
  if (!r) return false
  return r.some(({ metrics = [] }) => metrics.some(m => ['pie', 'doughnut'].includes(m.type)))
})

const hasGradient = computed(() => hasAxis.value || hasPie.value)

const timelineOptions = computed(() => [
  { value: '', text: t('edit.toolbox.timeline.options.none') },
  { value: 'x', text: t('edit.toolbox.timeline.options.x') },
  { value: 'y', text: t('edit.toolbox.timeline.options.y') },
  { value: 'xy', text: t('edit.toolbox.timeline.options.xy') },
])

const showEditColorSchemeButton = computed(() => {
  const { config = {} } = chart.value || {}
  return config.colorScheme && config.colorScheme.includes('custom') && canManageColorSchemes.value
})

const themeSettings = computed(() => $Settings.get('ui.studio.themes', []))

const defaultReportComputed = computed(() => Object.assign({}, defaultReport))

const { toastErrorHandler, toastSuccess } = composables.useToast()

watch(() => props.chartID, () => { fetchChart() }, { immediate: true })

watch(() => chart.value ? chart.value.config : undefined, (value, oldValue) => {
  if (value && oldValue) onConfigUpdate()
}, { deep: true })

function moduleName (moduleID) {
  const m = modByID.value(moduleID)
  return m ? m.name : ''
}

function fetchChart (chartID = props.chartID) {
  const namespaceID = props.namespace?.namespaceID

  if (canManageColorSchemes.value) fetchCustomColorSchemes()

  if (chartID === NoID) {
    let c = new compose.Chart({ namespaceID })
    switch (props.category) {
      case 'gauge': c = new compose.GaugeChart(c); break
      case 'funnel': c = new compose.FunnelChart(c); break
      case 'radar': c = new compose.RadarChart(c); break
    }
    chart.value = c
    initialChartState.value = chart.value.clone()
    document.title = t('label.app-name.chart.create')
    onEditReport(0)
  } else {
    loading.value = true
    processing.value = true
    store.dispatch('chart/findByID', { namespaceID, chartID, force: true }).then((c) => {
      chart.value = chartConstructor(c)
      initialChartState.value = chart.value.clone()
      document.title = t('label.app-name.chart.edit', { label: chart.value.name, interpolation: { escapeValue: false } })
      onEditReport(0)
    }).catch(e => {
      toastErrorHandler(t('notification.chart.loadFailed'))(e)
      router.push({ name: 'admin.charts' })
    }).finally(() => {
      setTimeout(() => { loading.value = false; processing.value = false }, 300)
    })
  }
}

function reporter (r) {
  const nr = { ...r }
  if (nr.filter) {
    nr.filter = evaluatePrefilter(nr.filter, {
      record: { values: {} },
      user: $auth.user || {},
      recordID: NoID,
      ownerID: NoID,
      userID: ($auth.user || {}).userID || NoID,
    })
  }
  return $ComposeAPI.recordReport({ namespaceID: props.namespace?.namespaceID, ...nr })
}

function update () {
  if (chartComponentRef.value) chartComponentRef.value.updateChart()
}

const onConfigUpdate = debounce(function () { update() }, 300)

function onUpdated () { processing.value = false }

function handleSave ({ ch = chart.value, closeOnSuccess = false, isClone = false } = {}) {
  const toggleProcessing = (value = true) => {
    if (closeOnSuccess) processingSaveAndClose.value = value
    else if (isClone) processingClone.value = value
    else processingSave.value = value
  }

  processing.value = true
  toggleProcessing()

  const resourceTranslationLanguage = currentLanguage()
  const c = Object.assign({}, ch, resourceTranslationLanguage)

  if (ch.chartID === NoID) {
    store.dispatch('chart/create', c).then(newChart => {
      loading.value = true
      chart.value = chartConstructor(newChart)
      initialChartState.value = chart.value.clone()
      document.title = t('label.app-name.chart.edit', { label: chart.value.name, interpolation: { escapeValue: false } })
      toastSuccess(t('notification.chart.created'))
      toggleProcessing(false)
      if (closeOnSuccess) router.push({ name: 'admin.charts' })
      else router.push({ name: 'admin.charts.edit', params: { chartID: newChart.chartID } })
    }).catch(e => {
      toastErrorHandler(t('notification.chart.createFailed'))(e)
      processing.value = false
      toggleProcessing(false)
    })
  } else {
    store.dispatch('chart/update', c).then(updatedChart => {
      chart.value = chartConstructor(updatedChart)
      initialChartState.value = chart.value.clone()
      document.title = t('label.app-name.chart.edit', { label: chart.value.name, interpolation: { escapeValue: false } })
      toastSuccess(t('notification.chart.updated'))
      if (closeOnSuccess) router.push({ name: 'admin.charts' })
    }).catch(toastErrorHandler(t('notification.chart.updateFailed')))
      .finally(() => {
        processing.value = false
        setTimeout(() => { toggleProcessing(false) }, 300)
      })
  }
}

function currentLanguage () { return undefined }

function handleDelete () {
  processing.value = true
  processingDelete.value = true
  store.dispatch('chart/delete', chart.value).then(() => {
    chart.value.deletedAt = new Date()
    initialChartState.value = chart.value.clone()
    toastSuccess(t('notification.chart.deleted'))
    router.push({ name: 'admin.charts' })
  }).catch(toastErrorHandler(t('notification.chart.deleteFailed')))
    .finally(() => { processing.value = false; processingDelete.value = false })
}

function handleClone () {
  const ch = chart.value.clone()
  ch.chartID = NoID
  ch.name = `${chart.value.name} (copy)`
  ch.handle = ''
  handleSave({ ch, isClone: true })
}

function onEditReport (i) { editReportIndex.value = i }

function onRemoveReport (i) {
  reports.value.splice(i, 1)
  if (editReportIndex.value === i) editReportIndex.value = undefined
}

function onAddReport () { reports.value.push(chart.value.defReport()) }

async function fetchCustomColorSchemes () {
  return $SystemAPI.settingsList({ prefix: 'ui.charts.colorSchemes' })
    .then(settings => {
      const { value = [] } = settings[0] || {}
      customColorSchemes.value = value
    })
    .catch(toastErrorHandler(t('notification.chart.colorScheme.fetch.failed')))
}

async function saveColorScheme () {
  colorSchemeModal.value.processing = true
  const action = colorSchemeModal.value.edit ? 'update' : 'create'

  if (colorSchemeModal.value.edit) {
    const index = customColorSchemes.value.findIndex(({ id }) => id === colorSchemeModal.value.colorscheme.id)
    customColorSchemes.value.splice(index, 1, colorSchemeModal.value.colorscheme)
  } else {
    customColorSchemes.value.push(colorSchemeModal.value.colorscheme)
  }

  const values = [{ name: 'ui.charts.colorSchemes', value: customColorSchemes.value }]

  return $SystemAPI.settingsUpdate({ values })
    .then(() => {
      chart.value.config.colorScheme = colorSchemeModal.value.colorscheme.id
      closeColorSchemeModal()
      toastSuccess(t(`notification:chart.colorScheme.${action}.success`))
      return $Settings.fetch().then(() => update())
    })
    .catch(toastErrorHandler(t(`notification:chart.colorScheme.${action}.failed`)))
    .finally(() => { colorSchemeModal.value.processing = false })
}

async function deleteColorScheme () {
  colorSchemeModal.value.processing = true
  const value = customColorSchemes.value.filter(({ id }) => id !== colorSchemeModal.value.colorscheme.id)
  const values = [{ name: 'ui.charts.colorSchemes', value }]

  return $SystemAPI.settingsUpdate({ values })
    .then(() => {
      chart.value.config.colorScheme = undefined
      customColorSchemes.value = value
      closeColorSchemeModal()
      toastSuccess(t('notification.chart.colorScheme.delete.success'))
      return $Settings.fetch()
    })
    .catch(toastErrorHandler(t('notification.chart.colorScheme.delete.failed')))
    .finally(() => { colorSchemeModal.value.processing = false })
}

function createColorScheme () {
  colorSchemeModal.value.edit = false
  colorSchemeModal.value.colorscheme = { id: `custom-${Date.now()}`, name: '', colors: ['#6C757D', '#000000'] }
  colorSchemeModal.value.show = true
}

function editColorScheme () {
  colorSchemeModal.value.edit = true
  colorSchemeModal.value.colorscheme = {
    id: currentColorScheme.value.id,
    name: currentColorScheme.value.name,
    colors: [...currentColorScheme.value.colors],
  }
  colorSchemeModal.value.show = true
}

function closeColorSchemeModal () { colorSchemeModal.value.show = false }

function addColor () { colorSchemeModal.value.colorscheme.colors.push('#000000') }

function removeColor (index) { colorSchemeModal.value.colorscheme.colors.splice(index, 1) }

function checkUnsavedChanges (next) {
  const { chartID = NoID } = chart.value || {}
  if (chartID === NoID) return next(window.confirm(t('editor.unsavedChanges')))
  const chartState = chart.value ? chart.value.clone() : {}
  const initState = initialChartState.value ? initialChartState.value.clone() : {}
  return next(isEqual(chartState, initState) || window.confirm(t('editor.unsavedChanges')))
}
</script>

<style lang="scss">
.chart-preview {
  max-height: 50%;
}
.color-box {
  width: 18px;
  height: 8px;
}
</style>
