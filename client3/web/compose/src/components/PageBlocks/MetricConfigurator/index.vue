<template>
  <div>
    <ul class="nav nav-tabs" role="tablist">
      <li class="nav-item" role="presentation">
        <button
          class="nav-link"
          :class="{ active: activeTab === 0 }"
          @click="activeTab = 0"
        >
          {{ $t('metric.edit.tabTitle') }}
        </button>
      </li>
    </ul>

    <div class="tab-content">
      <div class="tab-pane active">
        <div class="row g-0">
          <div class="col-12">
            <div
              v-for="(m, i) in metrics"
              :key="i"
              class="mb-2"
            >
              <button
                class="btn btn-outline-secondary me-1"
                @click="editMetric(m)"
              >
                {{ $t('label.edit') }}
              </button>
              <button
                class="btn btn-outline-danger me-2"
                @click="removeMetric(i)"
              >
                {{ $t('label.remove') }}
              </button>
              <span class="btn">
                {{ m.label || $t('metric.defaultMetricLabel') }}
              </span>
            </div>

            <button
              class="btn btn-link px-1"
              @click="addMetric"
            >
              + {{ $t('label.add') }}
            </button>
          </div>
        </div>

        <hr />

        <div class="row mt-3">
          <div
            v-if="edit"
            class="col-12 col-lg-7"
          >
            <div class="card mb-5">
              <div class="card-body">
                <fieldset>
                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.labelLabel') }}</label>
                    <input
                      v-model="edit.label"
                      class="form-control mb-1"
                      :placeholder="$t('metric.edit.labelPlaceholder')"
                    />
                  </div>
                </fieldset>

                <template v-if="likeRecordList !== true">
                  <div class="mb-3">
                    <div class="form-check">
                      <input
                        v-model="edit.showLabel"
                        type="checkbox"
                        class="form-check-input"
                      />
                      <label class="form-check-label">{{ $t('metric.edit.showLabel') }}</label>
                    </div>
                  </div>

                  <div class="mb-3">
                    <div class="form-check">
                      <input
                        v-model="edit.showLabelTooltip"
                        type="checkbox"
                        class="form-check-input"
                      />
                      <label class="form-check-label">{{ $t('metric.edit.showLabelTooltip') }}</label>
                    </div>
                  </div>
                </template>

                <fieldset>
                  <h5>{{ $t('metric.edit.dimensionLabel') }}</h5>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.moduleLabel') }}</label>
                    <c-input-select
                      v-model="edit.moduleID"
                      :options="modules"
                      label="name"
                      class="mt-1"
                      :reduce="o => o.moduleID"
                      :placeholder="$t('metric.edit.modulePlaceholder')"
                    />
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.filterLabel') }}</label>
                    <c-input-expression
                      v-model="edit.filter"
                      auto-complete
                      placeholder="(A > B) OR (A < C)"
                      class="mb-1"
                      min-height="3.448rem"
                      :suggestion-params="recordAutoCompleteParams"
                    />
                    <i18next
                      path="metric.edit.filterFootnote"
                      tag="small"
                      class="d-block text-muted"
                    >
                      <code>${record.values.fieldName}</code>
                      <code>${recordID}</code>
                      <code>${ownerID}</code>
                      <span><code>${userID}</code>, <code>${user.name}</code></span>
                    </i18next>
                  </div>

                  <div v-if="selectedMetricModule" class="row">
                    <div class="col-12 col-lg-6">
                      <div class="mb-3">
                        <label class="form-label text-primary">{{ $t('metric.edit.dimensionFieldLabel') }}</label>
                        <c-input-select
                          v-model="edit.dimensionField"
                          :options="dimensionFields"
                          :clearable="true"
                          :disabled="!!edit.periodCompareEnabled"
                          :get-option-label="o => o.label || o.name"
                          :reduce="o => o.name"
                          :placeholder="$t('metric.edit.dimensionFieldPlaceholder')"
                        />
                        <small class="text-muted d-block">{{ $t('metric.edit.dimensionFieldFootnote') }}</small>
                      </div>
                    </div>
                    <div class="col-12 col-lg-6">
                      <div class="mb-3">
                        <label class="form-label text-primary">{{ $t('metric.edit.bucketLabel') }}</label>
                        <c-input-select
                          v-model="edit.bucketSize"
                          :options="dimensionModifiers"
                          :clearable="false"
                          :disabled="!isTemporalField(edit.dimensionField)"
                          :get-option-label="o => o.text"
                          :reduce="o => o.value"
                          :placeholder="$t('metric.edit.bucketPlaceholder')"
                        />
                      </div>
                    </div>
                  </div>

                  <template v-if="edit.dimensionField">
                    <div class="row">
                      <div class="col-12 col-lg-6">
                        <div class="mb-3">
                          <label class="form-label text-primary">{{ $t('metric.edit.topNLabel') }}</label>
                          <input
                            v-model.number="edit.topN"
                            type="number"
                            min="0"
                            class="form-control mb-1"
                            :placeholder="$t('metric.edit.topNPlaceholder')"
                          />
                        </div>
                      </div>
                      <div class="col-12 col-lg-6">
                        <div class="mb-3">
                          <label class="form-label text-primary">{{ $t('metric.edit.sortDirectionLabel') }}</label>
                          <c-input-select
                            v-model="edit.sortDirection"
                            :options="sortDirectionOptions"
                            :clearable="false"
                            :get-option-label="o => o.label"
                            :reduce="o => o.value"
                          />
                        </div>
                      </div>
                    </div>

                    <div class="mb-3">
                      <label class="form-label text-primary">{{ $t('metric.edit.compareValuesLabel') }}</label>
                      <div class="row g-2">
                        <div class="col-6">
                          <input
                            :value="(edit.compareValues || [])[0] || ''"
                            class="form-control"
                            :placeholder="$t('metric.edit.compareValue1Placeholder')"
                            @input="setCompareValue(0, $event.target.value)"
                          />
                        </div>
                        <div class="col-6">
                          <input
                            :value="(edit.compareValues || [])[1] || ''"
                            class="form-control"
                            :placeholder="$t('metric.edit.compareValue2Placeholder')"
                            @input="setCompareValue(1, $event.target.value)"
                          />
                        </div>
                      </div>
                      <small class="text-muted d-block">{{ $t('metric.edit.compareValuesFootnote') }}</small>
                    </div>
                  </template>
                </fieldset>

                <fieldset v-if="selectedMetricModule">
                  <h5>{{ $t('metric.edit.periodCompare.label') }}</h5>

                  <div class="mb-3">
                    <div class="form-check">
                      <input
                        v-model="edit.periodCompareEnabled"
                        type="checkbox"
                        class="form-check-input"
                        :disabled="!!edit.dimensionField"
                      />
                      <label class="form-check-label">{{ $t('metric.edit.periodCompare.enable') }}</label>
                    </div>
                  </div>

                  <template v-if="edit.periodCompareEnabled">
                    <div class="row">
                      <div class="col-12 col-lg-6">
                        <div class="mb-3">
                          <label class="form-label text-primary">{{ $t('metric.edit.periodCompare.dateFieldLabel') }}</label>
                          <c-input-select
                            v-model="edit.periodDateField"
                            :options="dateFields"
                            :clearable="false"
                            :get-option-label="o => o.label || o.name"
                            :reduce="o => o.name"
                            :placeholder="$t('metric.edit.periodCompare.dateFieldPlaceholder')"
                          />
                        </div>
                      </div>
                      <div class="col-12 col-lg-6">
                        <div class="mb-3">
                          <label class="form-label text-primary">{{ $t('metric.edit.periodCompare.granularityLabel') }}</label>
                          <c-input-select
                            v-model="edit.periodGranularity"
                            :options="periodGranularityOptions"
                            :clearable="false"
                            :get-option-label="o => o.label"
                            :reduce="o => o.value"
                          />
                        </div>
                      </div>
                      <div class="col-12 col-lg-6">
                        <div class="mb-3">
                          <label class="form-label text-primary">{{ $t('metric.edit.periodCompare.modeLabel') }}</label>
                          <c-input-select
                            v-model="edit.periodCompareMode"
                            :options="periodCompareModeOptions"
                            :clearable="false"
                            :get-option-label="o => o.label"
                            :reduce="o => o.value"
                          />
                        </div>
                      </div>
                      <div class="col-12 col-lg-6">
                        <div class="mb-3">
                          <label class="form-label text-primary">{{ $t('metric.edit.periodCompare.trendPositiveLabel') }}</label>
                          <c-input-select
                            v-model="edit.trendPositive"
                            :options="trendPositiveOptions"
                            :clearable="false"
                            :get-option-label="o => o.label"
                            :reduce="o => o.value"
                          />
                        </div>
                      </div>
                    </div>
                  </template>
                </fieldset>

                <fieldset v-if="selectedMetricModule">
                  <h5>{{ $t('metric.edit.metricLabel') }}</h5>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.metricFieldLabel') }}</label>
                    <c-input-select
                      v-model="edit.metricField"
                      :placeholder="$t('metric.edit.metricFieldSelect')"
                      :options="metricFields"
                      :get-option-key="getOptionMetricFieldKey"
                      :get-option-label="getOptionMetricFieldLabel"
                      :reduce="f => f.name"
                      @input="onMetricFieldChange"
                    />
                  </div>

                  <div
                    v-if="edit.metricField !== 'number_expression'"
                    class="mb-3"
                  >
                    <label class="form-label text-primary">{{ $t('metric.edit.metricAggregateLabel') }}</label>
                    <c-input-select
                      v-model="edit.operation"
                      :disabled="edit.metricField === 'count' || edit.metricField === 'number_expression'"
                      :placeholder="$t('metric.edit.metricSelectAggregate')"
                      :options="aggregationOperations"
                      :get-option-key="getOptionAggregationOperationKey"
                      :reduce="a => a.operation"
                    />
                  </div>

                  <div
                    v-if="edit.metricField === 'number_expression'"
                    class="mb-3"
                  >
                    <label class="form-label text-primary">{{ $t('metric.edit.expressionFieldLabel') }}</label>
                    <c-input-expression
                      v-model="edit.expression"
                      auto-complete
                      placeholder="$t('metric.edit.expressionPlaceholder')"
                      class="mb-1"
                      min-height="3.448rem"
                      :suggestion-params="recordAutoCompleteParams"
                    />
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.transformFunctionLabel') }}</label>
                    <c-input-expression
                      v-model="edit.transformFx"
                      auto-complete
                      placeholder="v"
                      class="mb-1"
                      min-height="3.448rem"
                      :suggestion-params="recordAutoCompleteParams"
                    />
                    <small>{{ $t('metric.edit.transformFunctionDescription') }}</small>
                    <i18next
                      path="metric.edit.transformFootnote"
                      tag="small"
                      class="d-block text-muted"
                    >
                      <code>${record.values.fieldName}</code>
                      <code>${recordID}</code>
                      <code>${ownerID}</code>
                      <span><code>${userID}</code>, <code>${user.name}</code></span>
                    </i18next>
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.numberFormat') }}</label>
                    <input
                      v-model="edit.numberFormat"
                      class="form-control mb-1"
                      placeholder="0.00"
                    />
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.prefixLabel') }}</label>
                    <input
                      v-model="edit.prefix"
                      class="form-control mb-1"
                      placeholder="$"
                    />
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.edit.suffixLabel') }}</label>
                    <input
                      v-model="edit.suffix"
                      class="form-control mb-1"
                      placeholder="USD/mo"
                    />
                  </div>

                  <div class="mb-3">
                    <label class="form-label text-primary">
                      {{ $t('metric.drillDown.label') }}
                    </label>
                    <c-input-checkbox
                      v-model="edit.drillDown.enabled"
                      switch
                      :labels="checkboxLabel"
                    />
                    <small class="form-text">{{ $t('metric.drillDown.description') }}</small>
                    <div class="input-group">
                      <c-input-select
                        v-model="edit.drillDown.blockID"
                        :options="drillDownOptions"
                        :disabled="!edit.drillDown.enabled"
                        :get-option-label="o => o.title || o.kind"
                        :reduce="option => option.blockID"
                        :clearable="true"
                        :placeholder="$t('metric.drillDown.openInModal')"
                        append-to-body
                        class="flex-grow-1"
                      />
                      <column-picker
                        :module="selectedMetricModule"
                        :disabled="!!edit.drillDown.blockID || !edit.drillDown.enabled"
                        :fields="selectedDrilldownFields"
                        variant="extra-light"
                        size="md"
                        @updateFields="onUpdateFields"
                      >
                        <font-awesome-icon :icon="['fas', 'wrench']" />
                      </column-picker>
                    </div>
                  </div>
                </fieldset>
              </div>
            </div>

            <div class="mb-3">
              <div class="form-check">
                <input
                  v-model="likeRecordList"
                  type="checkbox"
                  class="form-check-input"
                  id="metric-like-record"
                />
                <label class="form-check-label" for="metric-like-record">{{ $t('metric.likeRecordList') }}</label>
              </div>
            </div>

            <hr />
            <h5 class="mb-3">{{ $t('metric.appearance.label') }}</h5>
            <div class="row">
              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">{{ $t('metric.appearance.density') }}</label>
                  <select v-model="options.density" class="form-select form-control">
                    <option value="comfortable">{{ $t('metric.appearance.densityOptions.comfortable') }}</option>
                    <option value="compact">{{ $t('metric.appearance.densityOptions.compact') }}</option>
                  </select>
                </div>
              </div>
              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">{{ $t('metric.appearance.hideEmptyMetrics') }}</label>
                  <c-input-checkbox v-model="options.hideEmptyMetrics" switch :labels="checkboxLabel" />
                </div>
              </div>
              <div class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">{{ $t('metric.appearance.showEmptyPlaceholder') }}</label>
                  <small class="text-muted d-block mb-1">{{ $t('metric.appearance.showEmptyPlaceholderDescription') }}</small>
                  <c-input-checkbox
                    v-model="options.showEmptyPlaceholder"
                    switch
                    :labels="checkboxLabel"
                    :disabled="options.hideEmptyMetrics"
                  />
                </div>
              </div>
              <div v-if="edit" class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">{{ $t('metric.appearance.role') }}</label>
                  <small class="text-muted d-block mb-1">{{ $t('metric.appearance.roleDescription') }}</small>
                  <select
                    class="form-select form-control"
                    :value="edit.role === 'topK' ? 'balloon' : (edit.role || 'default')"
                    @change="setMetricRole($event.target.value)"
                  >
                    <option
                      v-for="opt in metricRoleOptions"
                      :key="opt.value"
                      :value="opt.value"
                    >
                      {{ opt.text }}
                    </option>
                  </select>
                </div>
              </div>
              <div v-if="edit && isBalloonRole(edit)" class="col-12 col-lg-6">
                <div class="mb-3">
                  <label class="form-label text-primary">{{ $t('metric.appearance.balloonFullWidth') }}</label>
                  <small class="text-muted d-block mb-1">{{ $t('metric.appearance.balloonFullWidthDescription') }}</small>
                  <c-input-checkbox v-model="edit.balloonFullWidth" switch :labels="checkboxLabel" />
                </div>
              </div>
            </div>

            <div v-if="likeRecordList" class="mb-3">
              <h5 class="mb-2">{{ $t('metric.appearance.sections') }}</h5>
              <small class="text-muted d-block mb-2">{{ $t('metric.appearance.sectionsDescription') }}</small>
              <c-form-table-wrapper
                :labels="{ addButton: $t('metric.appearance.addSection') }"
                class="my-2"
                @add-item="addSection"
              >
                <table v-if="options.sections?.length" class="table table-sm table-borderless align-middle">
                  <thead>
                    <tr>
                      <th class="text-primary">{{ $t('metric.appearance.sectionTitle') }}</th>
                      <th class="text-primary">{{ $t('metric.appearance.sectionMetrics') }}</th>
                      <th style="width: 4rem;" />
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(section, i) in options.sections" :key="i">
                      <td style="min-width: 12rem;">
                        <input
                          v-model="section.title"
                          type="text"
                          class="form-control form-control-sm"
                          :placeholder="$t('metric.appearance.sectionTitlePlaceholder')"
                        >
                      </td>
                      <td style="min-width: 16rem;">
                        <c-input-select
                          v-model="section.metrics"
                          :options="sectionMetricOptions"
                          multiple
                          :close-on-select="false"
                          :get-option-label="o => o.label"
                          :get-option-key="o => o.value"
                          :reduce="o => o.value"
                          :placeholder="$t('metric.appearance.sectionMetricsPlaceholder')"
                        />
                      </td>
                      <td class="text-end">
                        <c-input-confirm show-icon @confirmed="removeSection(i)" />
                      </td>
                    </tr>
                  </tbody>
                </table>
              </c-form-table-wrapper>
            </div>

            <template v-if="likeRecordList">
              <div class="row g-0">
                <div class="col-12 col-lg-6">
                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('metric.appearance.itemsPerRow.label') }}</label>
                    <small class="text-muted d-block mb-1">{{ $t('metric.appearance.itemsPerRow.description') }}</small>
                    <select
                      class="form-select form-control"
                      :value="itemsPerRowModel"
                      @change="onItemsPerRowChange"
                    >
                      <option value="1">{{ $t('metric.appearance.itemsPerRow.one') }}</option>
                      <option value="2">{{ $t('metric.appearance.itemsPerRow.two') }}</option>
                      <option value="4">{{ $t('metric.appearance.itemsPerRow.four') }}</option>
                      <option value="auto">{{ $t('metric.appearance.itemsPerRow.auto') }}</option>
                    </select>
                  </div>
                </div>

                <div class="col-12 col-lg-6">
                  <div class="mb-3">
                    <label class="form-label text-primary">{{ $t('record.horizontalFormLayout') }}</label>
                    <c-input-checkbox
                      v-model="options.horizontalFieldLayoutEnabled"
                      switch
                      :disabled="itemsPerRowModel === 'auto'"
                      :labels="checkboxLabel"
                    />
                  </div>
                </div>
              </div>
            </template>

            <m-style
              class="mt-2"
              :options="edit.valueStyle"
              :balloon="isBalloonRole(edit)"
            >
              <template #title>
                <h5>{{ $t('metric.editStyle.valueLabel') }}</h5>
              </template>
            </m-style>
          </div>

          <div class="col-12 col-lg-5">
            <div
              v-if="metrics.length"
              class="d-flex flex-column position-sticky pt-2"
              style="top: 0;"
            >
              <button
                title="Refresh data"
                class="btn btn-outline-light d-flex align-items-center text-primary ms-auto border-0 px-2 mt-2 me-2 btn-lg"
                @click.prevent="refreshMetric"
              >
                <font-awesome-icon :icon="['fas', 'sync']" />
              </button>

              <div
                class="mt-2"
                style="height: 400px;"
              >
                <metric-base
                  :key="'ipr-' + itemsPerRowModel"
                  v-bind="$props"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, watch, onBeforeUnmount, inject } from 'vue'
import { usePageBlockBase } from '../usePageBlockBase'
import { useStore } from '../../../store'
import MStyle from './MStyle'
import MetricBase from '../MetricBase'
import ColumnPicker from 'corteza-webapp-compose/src/components/Admin/Module/Records/ColumnPicker'
import { compose, NoID } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
const { CInputExpression } = components

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
const store = useStore()
const $auth = inject('$auth')

const { options, setBaseDefaultValues } = usePageBlockBase(props, emit)
const activeTab = ref(0)
const edit = ref(undefined)

// Normalize appearance defaults for older blocks
if (!options.value.density) options.value.density = 'comfortable'
{
  const per = options.value.itemsPerRow == null ? '' : String(options.value.itemsPerRow)
  if (['1', '2', '4', 'auto'].includes(per)) {
    options.value.itemsPerRow = per
  } else {
    const legacy = options.value.recordFieldLayoutOption
    options.value.itemsPerRow = legacy === 'wrap' ? '2' : legacy === 'noWrap' ? 'auto' : '1'
  }
}
if (options.value.hideEmptyMetrics === undefined) options.value.hideEmptyMetrics = false
if (options.value.showEmptyPlaceholder === undefined) options.value.showEmptyPlaceholder = true
if (!Array.isArray(options.value.sections)) options.value.sections = []

const checkboxLabel = ref({ on: t('label.yes'), off: t('label.no') })

const dimensionModifiers = computed(() =>
  compose.chartUtil.dimensionFunctions.map(df => ({ ...df, text: t(`chart.edit.dimension.function.${df.text}`) }))
)

// Kinds a dimension breakdown can sensibly group by — mirrors Chart's own
// dimension-field picker (ReportEdit.vue).
const dimensionFieldKinds = ['DateTime', 'Select', 'Number', 'Bool', 'String', 'Record', 'User']
const dimensionFields = computed(() => fields.value.filter(f =>
  dimensionFieldKinds.includes(f.kind) && !f.options?.useRichTextEditor && !f.options?.multiLine))

const dateFields = computed(() => fields.value.filter(f => f.kind === 'DateTime'))

const sortDirectionOptions = computed(() => [
  { value: 'value-desc', label: t('metric.edit.sortDirectionOptions.valueDesc') },
  { value: 'value-asc', label: t('metric.edit.sortDirectionOptions.valueAsc') },
  { value: 'label-asc', label: t('metric.edit.sortDirectionOptions.labelAsc') },
  { value: 'label-desc', label: t('metric.edit.sortDirectionOptions.labelDesc') },
])

const periodGranularityOptions = computed(() => [
  { value: 'day', label: t('metric.edit.periodCompare.granularityOptions.day') },
  { value: 'week', label: t('metric.edit.periodCompare.granularityOptions.week') },
  { value: 'month', label: t('metric.edit.periodCompare.granularityOptions.month') },
  { value: 'quarter', label: t('metric.edit.periodCompare.granularityOptions.quarter') },
  { value: 'year', label: t('metric.edit.periodCompare.granularityOptions.year') },
])

const periodCompareModeOptions = computed(() => [
  { value: 'previous-period', label: t('metric.edit.periodCompare.modeOptions.previousPeriod') },
  { value: 'year-over-year', label: t('metric.edit.periodCompare.modeOptions.yearOverYear') },
])

const trendPositiveOptions = computed(() => [
  { value: 'increase', label: t('metric.edit.periodCompare.trendPositiveOptions.increase') },
  { value: 'decrease', label: t('metric.edit.periodCompare.trendPositiveOptions.decrease') },
])

function setCompareValue (index, value) {
  if (!edit.value) return
  const next = [...(edit.value.compareValues || ['', ''])]
  next[index] = value
  edit.value.compareValues = next
}

const aggregationOperations = computed(() => [
  { label: t('metric.edit.operationSum'), operation: 'sum' },
  { label: t('metric.edit.operationMax'), operation: 'max' },
  { label: t('metric.edit.operationMin'), operation: 'min' },
  { label: t('metric.edit.operationAvg'), operation: 'avg' },
  { label: t('metric.edit.operationUniqueCount'), operation: 'uniqueCount' },
])

const modules = computed(() => store.module.set)
const getModuleByID = computed(() => store.module.getByID)

const fields = computed(() => {
  if (!edit.value || !edit.value.moduleID) return []
  return getModuleByID.value(edit.value.moduleID).fields
})

const selectedDrilldownFields = computed(() => {
  if (!edit.value || !edit.value.drillDown?.recordListOptions?.fields) return []
  return edit.value.drillDown.recordListOptions.fields
})

const metricFields = computed(() => {
  return [
    { name: 'count', label: 'Count' },
    { name: 'number_expression', label: 'Number Expression' },
    ...fields.value.filter(f => f.kind === 'Number').sort((a, b) => a.label.localeCompare(b.label)),
  ]
})

const likeRecordList = computed({
  get: () => options.value.likeRecordList !== false,
  set: (check) => { options.value.likeRecordList = check },
})

const itemsPerRowModel = computed(() => {
  const v = props.block?.options?.itemsPerRow ?? options.value?.itemsPerRow
  const s = v == null ? '' : String(v)
  return ['1', '2', '4', 'auto'].includes(s) ? s : '1'
})

function onItemsPerRowChange (e) {
  const v = e.target?.value || '1'
  if (props.block?.options) props.block.options.itemsPerRow = v
  options.value.itemsPerRow = v
  if (v === '1') options.value.recordFieldLayoutOption = 'default'
  else if (v === '2') options.value.recordFieldLayoutOption = 'wrap'
  else options.value.recordFieldLayoutOption = 'noWrap'
  if (v === 'auto') options.value.horizontalFieldLayoutEnabled = false
}

const metrics = computed({
  get: () => options.value.metrics,
  set: (m) => { options.value.metrics = m },
})

const metricRoleOptions = computed(() => [
  { value: 'default', text: t('metric.appearance.roleOptions.default') },
  { value: 'title', text: t('metric.appearance.roleOptions.title') },
  { value: 'badge', text: t('metric.appearance.roleOptions.badge') },
  { value: 'meta', text: t('metric.appearance.roleOptions.meta') },
  { value: 'hero', text: t('metric.appearance.roleOptions.hero') },
  { value: 'balloon', text: t('metric.appearance.roleOptions.balloon') },
])

const sectionMetricOptions = computed(() =>
  (metrics.value || []).map((m, i) => ({
    value: i,
    label: m.label || t('metric.defaultMetricLabel') + ` #${i + 1}`,
  }))
)

const drillDownOptions = computed(() =>
  props.page.blocks.filter(({ blockID, kind, options: o = {} }) =>
    kind === 'RecordList' && blockID !== NoID && o.moduleID === edit.value?.moduleID)
)

const selectedMetricModule = computed(() => {
  if (!edit.value?.moduleID) return undefined
  return getModuleByID.value(edit.value.moduleID)
})

const recordAutoCompleteParams = computed(() => processRecordAutoCompleteParams({ module: selectedMetricModule.value, operators: true }))

function processRecordAutoCompleteParams ({ module: mod, operators = false } = {}) {
  const { fields = [] } = mod || {}
  const moduleFields = fields.map(({ name }) => name)
  const userProperties = ($auth?.user?.properties?.()) || []

  return [
    ...(operators ? ['AND', 'OR'] : []),
    { interpolate: true, value: 'userID' },
    { interpolate: true, value: 'user', properties: userProperties },
    ...moduleFields,
  ]
}

watch(() => edit.value?.dimensionField, (df, prev) => {
  if (!edit.value) return
  if (!isTemporalField(df)) {
    edit.value.bucketSize = undefined
    edit.value.dateFormat = undefined
  } else {
    edit.value.dateFormat = edit.value.dateFormat || 'YYYY-MM-DD'
  }
  // Breakdown/pairwise and period-compare are mutually exclusive per metric.
  // Skip the initial bind (prev === undefined) so a saved period-compare
  // metric is not wiped when the configurator opens.
  if (df && prev !== undefined) edit.value.periodCompareEnabled = false
})

watch(() => edit.value?.periodCompareEnabled, (enabled) => {
  if (!edit.value || !enabled) return
  edit.value.dimensionField = ''
  edit.value.compareValues = undefined
})

if (!metrics.value.length) { addMetric() }
edit.value = metrics.value[0]

onBeforeUnmount(() => { setDefaultValues() })

function setDefaultValues () {
  edit.value = undefined
}

function addMetric () {
  const m = props.block.makeMetric()
  metrics.value.push(m)
  editMetric(m)
}

function editMetric (m) { edit.value = m }
function removeMetric (i) {
  metrics.value.splice(i, 1)
  // Prune section indices after removal
  if (Array.isArray(options.value.sections)) {
    options.value.sections.forEach(s => {
      s.metrics = (s.metrics || [])
        .filter(idx => idx !== i)
        .map(idx => (idx > i ? idx - 1 : idx))
    })
  }
  edit.value = undefined
}
function isTemporalField (name) { return !!fields.value.find(f => f.name === name && f.kind === 'DateTime') }
function getOptionMetricFieldKey ({ name }) { return name }
function getOptionMetricFieldLabel ({ name, label }) { return label || name }
function getOptionAggregationOperationKey ({ operation }) { return operation }

function onMetricFieldChange (field) {
  if (field === 'count' || field === 'number_expression') {
    edit.value.operation = undefined
  } else if (!edit.value.operation) {
    edit.value.operation = aggregationOperations.value[0].operation
  }
}

function onUpdateFields (fields) { edit.value.drillDown.recordListOptions.fields = fields }

function setMetricRole (role) {
  if (!edit.value) return
  if (role === 'title') {
    metrics.value.forEach(m => {
      if (m !== edit.value && m.role === 'title') m.role = 'default'
    })
  }
  const next = role === 'topK' ? 'balloon' : (role || 'default')
  edit.value.role = next
  if (next === 'balloon') {
    if (!edit.value.valueStyle) edit.value.valueStyle = {}
    if (!Array.isArray(edit.value.valueStyle.colorThresholds)) {
      edit.value.valueStyle.colorThresholds = []
    }
  }
}

function isBalloonRole (m) {
  return m?.role === 'balloon' || m?.role === 'topK'
}

function addSection () {
  if (!Array.isArray(options.value.sections)) options.value.sections = []
  options.value.sections.push({ title: '', metrics: [] })
}

function removeSection (i) {
  options.value.sections.splice(i, 1)
}

function refreshMetric () {
  window.dispatchEvent(new CustomEvent('metric.update'))
}
</script>
