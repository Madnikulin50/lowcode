<template>
  <div>
    <div data-test-id="record-list-configurator">
      <h5 class="mb-3">{{ $t('recordList.record.generalLabel') }}</h5>
      <div class="row">
        <div class="col-12" :class="isInlineEditorAllowed ? 'col-lg-6' : 'col-lg-12'">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('module') }}</label>
            <c-input-select v-model="options.moduleID" :options="modules" label="name" :reduce="o => o.moduleID" :placeholder="$t('recordList.modulePlaceholder')" default-value="0" required />
          </div>
        </div>
        <div v-if="isInlineEditorAllowed" class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="d-flex align-items-center form-label text-primary">
              {{ $t('recordList.record.inlineEditorAllow') }}
              <c-hint :tooltip="$t('recordList.tooltip.performance.impact')" icon-class="text-warning" />
            </label>
            <c-input-checkbox v-model="options.editable" switch :labels="checkboxLabel" />
          </div>
        </div>
      </div>
    </div>

    <template v-if="recordListModule">
      <hr>

      <div class="">
        <div class="mb-3">
          <h5 class="d-flex align-items-center mb-1">
            {{ $t('module.general.fields') }}
            <c-hint :tooltip="$t('recordList.tooltip.performance.moduleFields')" icon-class="text-warning" />
          </h5>
          <small class="text-muted">{{ $t('recordList.moduleFieldsFootnote') }}</small>
        </div>
        <div class="row">
          <div class="col-12">
            <FieldPicker :module="recordListModule" v-model:fields="options.fields" class="mb-3" style="height: 50vh;" />
          </div>
          <div v-if="onRecordPage" class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.refField.label') }}</label>
              <c-input-select v-model="options.refField" :options="parentFields" :placeholder="$t('label.none')" :reduce="f => f.name" />
              <small class="text-muted d-block">{{ $t('recordList.refField.footnote') }}</small>
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.hideConfigureFieldsButton') }}</label>
              <c-input-checkbox v-model="options.hideConfigureFieldsButton" switch invert :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-md-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.textStyles') }}</label>
              <ColumnPicker size="sm" variant="outline-secondary" :module="recordListModule" :fields="options.textStyles.wrappedFields || []" :field-subset="options.fields.length ? options.fields : recordListModule.fields" @updateFields="onUpdateTextWrapOption">
                {{ $t('recordList.record.configureWrappedFields') }}
              </ColumnPicker>
            </div>
          </div>
        </div>
      </div>

      <hr>

      <div v-if="options.editable" class="">
        <h5 class="mb-3">{{ $t('recordList.record.inlineEditor') }}</h5>
        <div v-if="recordListModule && options.editable" class="mb-3">
          <label class="form-label text-primary">{{ $t('recordList.editFields') }}</label>
          <FieldPicker :module="recordListModule" v-model:fields="options.editFields" :field-subset="editableFieldSubset" style="height: 50vh;" />
        </div>
        <div class="row mt-3">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.positionField.label') }}</label>
              <c-input-select v-model="options.positionField" :options="positionFields" :placeholder="$t('recordList.positionField.placeholder')" :reduce="f => f.name" label="label" />
              <small class="text-muted d-block">{{ $t('recordList.positionField.footnote') }}</small>
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div v-if="options.positionField" class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.draggable') }}</label>
              <c-input-checkbox v-model="options.draggable" switch :labels="checkboxLabel" />
            </div>
          </div>
        </div>
      </div>

      <hr v-if="options.editable">

      <div class="">
        <h5 class="mb-3">{{ $t('recordList.record.prefilterLabel') }}</h5>
        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.prefilterHideSearch') }}</label>
              <c-input-checkbox v-model="options.hideSearch" switch invert :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.filterHide') }}</label>
              <c-input-checkbox v-model="options.hideFiltering" switch invert :labels="checkboxLabel" />
            </div>
          </div>
          <div v-if="!options.hideSearch" class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.searchableFields') }}</label>
              <ColumnPicker size="sm" variant="outline-secondary" :module="recordListModule" :fields="options.searchableFields" :field-subset="queryableFields" @updateFields="onUpdateSearchableFields">
                {{ $t('recordList.record.configureSearchableFields') }}
              </ColumnPicker>
              <small class="text-muted d-block">{{ $t('recordList.record.searchableFieldsFootnote') }}</small>
            </div>
          </div>
        </div>

        <Prefilter :record="record" :module="recordListModule" :namespace="namespace" :options="options" :page="page" />

        <hr>

        <div class="row">
          <div class="col">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.setCustomFilterPresets') }}</label>
              <c-input-checkbox v-model="options.customFilterPresets" switch :labels="checkboxLabel" />
            </div>
          </div>
        </div>

        <div class="row">
          <div class="col">
            <div class="mb-3">
              <label class="form-label text-primary mb-1">{{ $t('recordList.filter.presets') }}</label>
              <c-form-table-wrapper :loading="fetchingRoles" :labels="{ addButton: $t('label.add') }" @add-item="addFilterPreset">
                <table class="table table-sm table-borderless mb-2">
                  <draggable item-key="id" :list="options.filterPresets" group="sort" handle=".grab" tag="tbody">
                    <template #item="{ element, index }">
                      <tr :key="index">
                        <td class="grab text-center align-middle" style="width: 40px;">
                          <font-awesome-icon :icon="['fas', 'bars']" class="text-secondary" />
                        </td>
                        <td class="align-middle" style="min-width: 150px;">
                          <div class="input-group">
                            <input v-model="element.name" class="form-control" :placeholder="$t('recordList.filter.name.placeholder')" />
                            <RecordListFilter class="d-print-none" :target="`record-filter-${index}`" :namespace="namespace" :module="recordListModule" :record-list-filter="element.filter" variant="extra-light" inactive-icon-class="text-light" button-class="px-2 pt-2" :button-style="{ borderTopLeftRadius: 0, borderBottomLeftRadius: 0 }" @filter="(f) => onFilter(f, index)" />
                          </div>
                        </td>
                        <td class="text-center align-middle" style="min-width: 200px;">
                          <c-input-role :value="getResolvedRoles(element)" :placeholder="$t('recordList.filter.role.placeholder')" :visible="isRoleVisible" multiple @input="onRoleChange(element, $event)" />
                        </td>
                        <td class="text-end align-middle" style="min-width: 80px; width: 80px;">
                          <c-input-confirm show-icon @confirmed="options.filterPresets.splice(index, 1)" />
                        </td>
                      </tr>
                    </template>
                  </draggable>
                </table>
              </c-form-table-wrapper>
            </div>
          </div>
        </div>
      </div>
      <hr>

      <div v-if="!options.positionField" class="">
        <h5 class="mb-3">{{ $t('recordList.record.presortLabel') }}</h5>
        <div class="row">
          <div class="col">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.presortHideSort') }}</label>
              <c-input-checkbox v-model="options.hideSorting" switch invert :labels="checkboxLabel" />
            </div>
          </div>
        </div>
        <div class="row">
          <div class="col">
            <c-input-presort v-model="options.presort" :fields="recordListModuleFields" :labels="{ ascending: $t('label.ascending'), descending: $t('label.descending'), none: $t('label.none'), placeholder: $t('recordList.record.presortPlaceholder'), footnote: $t('recordList.record.presortFootnote'), toggleInput: $t('recordList.record.presortToggleInput'), addButton: $t('label.add'), title: $t('recordList.record.presortInputLabel') }" allow-text-input />
          </div>
        </div>
      </div>
      <hr v-if="!options.positionField">

      <div class="">
        <h5 class="mb-3">{{ $t('recordList.record.pagingLabel') }}</h5>
        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.hidePaging') }}</label>
              <c-input-checkbox v-model="options.hidePaging" switch invert :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="d-flex align-items-center form-label text-primary p-0">
                {{ $t('recordList.record.fullPageNavigation') }}
                <c-hint :tooltip="$t('recordList.tooltip.performance.impact')" icon-class="text-warning" />
              </label>
              <c-input-checkbox v-model="options.fullPageNavigation" switch :labels="checkboxLabel" data-test-id="hide-page-navigation" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="d-flex align-items-center form-label text-primary">
                {{ $t('recordList.record.perPage') }}
                <c-hint :tooltip="$t('recordList.tooltip.performance.perPage')" icon-class="text-warning" />
              </label>
              <input v-model.number="options.perPage" data-test-id="input-records-per-page" type="number" class="form-control mb-2" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.showRecordPerPageOption') }}</label>
              <c-input-checkbox v-model="options.showRecordPerPageOption" switch :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.showTotalCount') }}</label>
              <c-input-checkbox v-model="options.showTotalCount" data-test-id="show-total-record-count" switch :labels="checkboxLabel" />
            </div>
          </div>
        </div>
      </div>

      <hr>

      <div class="">
        <h5 class="mb-3">{{ $t('recordList.summaries.label') }}</h5>
        <div class="row">
          <div class="col-12">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.summaries.customSummaries.description') }}</label>
              <c-input-checkbox v-model="options.customSummaries" switch :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12">
            <c-form-table-wrapper :loading="fetchingRoles" :labels="{ addButton: $t('label.add') }" @add-item="addSummary">
              <table class="table table-sm table-borderless mb-2">
                <draggable item-key="id" :list="options.summaries" group="sort" handle=".grab" tag="tbody">
                  <template #item="{ element, index }">
                    <tr :key="index">
                      <td class="grab text-center align-middle" style="width: 40px;">
                        <font-awesome-icon :icon="['fas', 'bars']" class="text-secondary" />
                      </td>
                      <td style="min-width: 200px;">
                        <input v-model="element.label" class="form-control" :placeholder="$t('recordList.summaries.param.label.placeholder')" />
                      </td>
                      <td style="min-width: 200px;">
                        <c-input-select v-model="element.field" :options="recordListModuleFields" :reduce="field => field.name" :placeholder="$t('recordList.summaries.param.field.placeholder')" />
                      </td>
                      <td style="min-width: 200px;">
                        <c-input-select v-model="element.metric" :options="summaryMetrics" :reduce="m => m.value" :placeholder="$t('recordList.summaries.param.metric.placeholder')" />
                      </td>
                      <td style="min-width: 200px;">
                        <c-input-role :value="getResolvedRoles(element)" :placeholder="$t('recordList.summaries.param.role.placeholder')" :visible="isRoleVisible" multiple @input="onRoleChange(element, $event)" />
                      </td>
                      <td class="text-end align-middle">
                        <c-input-confirm show-icon @confirmed="options.summaries.splice(index, 1)" />
                      </td>
                    </tr>
                  </template>
                </draggable>
              </table>
            </c-form-table-wrapper>
          </div>
        </div>
      </div>

      <hr>

      <div class="">
        <h5 class="mb-3">{{ $t('recordList.record.recordsLabel') }}</h5>
        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.recordDisplayOptions') }}</label>
              <select v-model="options.recordDisplayOption" class="form-select form-control">
                <option v-for="opt in recordDisplayOptionsOnSelect" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
              </select>
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.recordSelectorDisplayOptions') }}</label>
              <select v-model="options.recordSelectorDisplayOption" class="form-select form-control">
                <option v-for="opt in recordDisplayOptions" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
              </select>
            </div>
          </div>
        </div>

        <div v-if="options.recordDisplayOption === 'sameTabSelectedModule'" class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.recordDisplayTargetModuleID') }}</label>
              <c-input-select v-model="options.recordDisplayTargetModuleID" :options="modules" label="name" :reduce="module => module.moduleID" :placeholder="$t('recordList.modulePlaceholder')" default-value="0" required />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.recordDisplayTargetModuleField') }}</label>
              <div class="input-group"><input v-model="options.recordDisplayTargetModuleField" class="form-control" /></div>
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.recordDisplayListField') }}</label>
              <div class="input-group"><input v-model="options.recordDisplayListField" class="form-control" /></div>
            </div>
          </div>
        </div>

        <div class="row">
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.hideAddButton') }}</label>
              <c-input-checkbox v-model="options.hideAddButton" switch invert :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.addRecordOptions') }}</label>
              <select v-model="options.addRecordDisplayOption" class="form-select form-control" :disabled="options.hideAddButton">
                <option v-for="opt in recordCreateOptions" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
              </select>
            </div>
          </div>
          <div class="col-12 col-md-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.editMode') }}</label>
              <c-input-checkbox v-model="options.openRecordInEditMode" switch :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-md-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.selectable') }}</label>
              <c-input-checkbox v-model="options.selectable" switch :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.hideImportButton') }}</label>
              <c-input-checkbox v-model="options.hideImportButton" switch invert :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.export.allow') }}</label>
              <c-input-checkbox v-model="options.allowExport" switch :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="d-flex form-label text-primary">
                {{ $t('recordList.inlineEdit.enabled') }}
                <ColumnPicker :module="recordListModule" :fields="options.inlineEditFields" :field-subset="options.fields.length ? options.fields : recordListModule.fields" :button-tooltip="{ title: $t('recordList.inlineEdit.fields.configure.tooltip'), boundary: 'body' }" variant="outline-extra-light" button-class="d-flex align-items-center text-secondary border-0 py-0 px-1 ms-1" size="sm" @updateFields="onUpdateInlineEditableFields">
                  <font-awesome-icon :icon="['fas', 'cog']" />
                </ColumnPicker>
              </label>
              <c-input-checkbox v-model="options.inlineRecordEditEnabled" switch :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.inlineEdit.allowAddField') }}</label>
              <c-input-checkbox v-model="options.inlineRecordEditAllowAddField" switch :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.enableBulkRecordEdit') }}</label>
              <c-input-checkbox v-model="options.bulkRecordEditEnabled" switch :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.inlineValueFilter.enabled') }}</label>
              <c-input-checkbox v-model="options.inlineValueFiltering" switch :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="d-flex align-items-center form-label text-primary mb-0">
                {{ $t('recordList.enableRecordPageNavigation') }}
                <c-hint :tooltip="$t('recordList.tooltip.performance.impact')" icon-class="text-warning" />
              </label>
              <c-input-checkbox v-model="options.enableRecordPageNavigation" switch :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.showDeletedRecordsOption') }}</label>
              <c-input-checkbox v-model="options.showDeletedRecordsOption" switch :labels="checkboxLabel" />
            </div>
          </div>
          <div class="col-12 col-lg-6">
            <div class="mb-3">
              <label class="form-label text-primary">{{ $t('recordList.record.buttons') }}</label>
              <div class="form-check">
                <input v-model="options.hideRecordViewButton" class="form-check-input" type="checkbox" id="hideRecordViewButton" />
                <label class="form-check-label" for="hideRecordViewButton">{{ $t('recordList.hideRecordViewButton') }}</label>
              </div>
              <div class="form-check">
                <input v-model="options.hideRecordEditButton" class="form-check-input" type="checkbox" id="hideRecordEditButton" />
                <label class="form-check-label" for="hideRecordEditButton">{{ $t('recordList.hideRecordEditButton') }}</label>
              </div>
              <div class="form-check">
                <input v-model="options.hideRecordCloneButton" class="form-check-input" type="checkbox" id="hideRecordCloneButton" />
                <label class="form-check-label" for="hideRecordCloneButton">{{ $t('recordList.hideRecordCloneButton') }}</label>
              </div>
              <div class="form-check">
                <input v-model="options.hideRecordReminderButton" class="form-check-input" type="checkbox" id="hideRecordReminderButton" />
                <label class="form-check-label" for="hideRecordReminderButton">{{ $t('recordList.hideRecordReminderButton') }}</label>
              </div>
              <div class="form-check">
                <input v-model="options.hideRecordPermissionsButton" class="form-check-input" type="checkbox" id="hideRecordPermissionsButton" />
                <label class="form-check-label" for="hideRecordPermissionsButton">{{ $t('recordList.hideRecordPermissionsButton') }}</label>
              </div>
              <div class="form-check">
                <input v-model="options.hideRecordDeleteButton" class="form-check-input" type="checkbox" id="hideRecordDeleteButton" />
                <label class="form-check-label" for="hideRecordDeleteButton">{{ $t('recordList.hideRecordDeleteButton') }}</label>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <AutomationTab v-bind="$props" :module="recordListModule" :buttons="options.selectionButtons" @update:buttons="options.selectionButtons = $event" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { useStore } from '../../store'
import { NoID } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import ColumnPicker from 'corteza-webapp-compose/src/components/Admin/Module/Records/ColumnPicker'
import FieldPicker from 'corteza-webapp-compose/src/components/Common/FieldPicker'
import RecordListFilter from 'corteza-webapp-compose/src/components/Common/RecordListFilter'
import Draggable from 'vuedraggable'
import Prefilter from './RecordList/Prefilter.vue'
import AutomationTab from './Shared/AutomationTab'

const { CInputPresort, CInputRole } = components
const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  block: { type: Object, required: true },
  module: { type: Object, required: false },
  namespace: { type: Object, required: true },
  page: { type: Object, required: true },
  record: { type: Object, required: false, default: undefined },
})

const store = useStore()
const $SystemAPI = inject('$SystemAPI')

const options = computed(() => props.block.options)
const checkboxLabel = computed(() => ({ on: $t('label.yes'), off: $t('label.no') }))
const fetchingRoles = ref(false)
const resolvedRoles = ref({})

const getModuleByID = computed(() => store.module.getByID)
const modules = computed(() => store.module.set)
const pages = computed(() => store.page.set)

const recordDisplayOptions = computed(() => [
  { value: 'sameTab', text: $t('recordList.record.openInSameTab') },
  { value: 'newTab', text: $t('recordList.record.openInNewTab') },
  { value: 'modal', text: $t('recordList.record.openInModal') },
  { value: 'doNothing', text: $t('recordList.record.doNothing') },
])

const recordDisplayOptionsOnSelect = computed(() => [
  { value: 'sameTab', text: $t('recordList.record.openInSameTab') },
  { value: 'sameTabSelectedModule', text: $t('recordList.record.openInSameTabSelectedModule') },
  { value: 'newTab', text: $t('recordList.record.openInNewTab') },
  { value: 'modal', text: $t('recordList.record.openInModal') },
  { value: 'doNothing', text: $t('recordList.record.doNothing') },
])

const recordCreateOptions = computed(() => [
  { value: 'sameTab', text: $t('recordList.record.createInSameTab') },
  { value: 'newTab', text: $t('recordList.record.createInNewTab') },
  { value: 'modal', text: $t('recordList.record.createInModal') },
])

const recordListModule = computed(() => {
  if (options.value.moduleID !== NoID) return getModuleByID.value(options.value.moduleID)
  return undefined
})

const recordListModuleFields = computed(() => {
  if (!recordListModule.value) return []
  return [
    ...recordListModule.value.fields,
    ...recordListModule.value.systemFields().map(sf => ({ label: $t(`system.${sf.name}`), name: sf.name === 'recordID' ? 'ID' : sf.name })),
  ].map(({ name, label }) => ({ name, label }))
})

const onRecordPage = computed(() => props.page?.moduleID !== NoID)

const recordListModuleRecordPage = computed(() => {
  if (options.value.moduleID !== NoID) return pages.value.find(p => p.moduleID === options.value.moduleID)
  return undefined
})

const parentFields = computed(() => {
  if (!recordListModule.value) return []
  return recordListModule.value.fields.filter(({ kind, options: opts }) => kind === 'Record' && props.record && opts.moduleID === props.record.moduleID)
})

const positionFields = computed(() => {
  if (!recordListModule.value) return []
  return recordListModule.value.fields.filter(({ kind, isMulti }) => kind === 'Number' && !isMulti)
})

const isInlineEditorAllowed = computed(() => !!recordListModule.value)

const summaryMetrics = computed(() => [
  { value: 'sum', label: $t('recordList.summaries.metrics.sum.label') },
  { value: 'min', label: $t('recordList.summaries.metrics.min.label') },
  { value: 'max', label: $t('recordList.summaries.metrics.max.label') },
  { value: 'avg', label: $t('recordList.summaries.metrics.avg.label') },
  { value: 'emptyCount', label: $t('recordList.summaries.metrics.emptyCount.label') },
  { value: 'notEmptyCount', label: $t('recordList.summaries.metrics.notEmptyCount.label') },
  { value: 'uniqueCount', label: $t('recordList.summaries.metrics.uniqueCount.label') },
])

const queryableFields = computed(() => {
  if (!recordListModule.value) return []
  return [...recordListModule.value.fields, ...recordListModule.value.systemFields()].filter(f => f.isQueryable)
})

const editableFieldSubset = computed(() => {
  if (!recordListModule.value) return []
  return options.value.fields.length ? options.value.fields : [...recordListModule.value.fields, recordListModule.value.systemFields().find(f => f.name === 'ownedBy')]
})

onMounted(() => { fetchRoles() })
onBeforeUnmount(() => { setDefaultValues() })

function fetchRoles() {
  const { filterPresets = [], summaries = [] } = options.value
  if (!filterPresets.length && !summaries.length) return
  fetchingRoles.value = true
  const rolesToResolve = []
  filterPresets.forEach(preset => { preset.roles.forEach(r => { if (!rolesToResolve.includes(r)) rolesToResolve.push(r) }) })
  summaries.forEach(summary => { summary.roles.forEach(r => { if (!rolesToResolve.includes(r)) rolesToResolve.push(r) }) })
  $SystemAPI.roleList({ roleID: rolesToResolve }).then(({ set }) => {
    set.forEach(role => { resolvedRoles.value[role.roleID] = role })
  }).finally(() => { fetchingRoles.value = false })
}

function onRoleChange(resource, value) {
  value.forEach(r => { if (!resolvedRoles.value[r.roleID]) resolvedRoles.value[r.roleID] = r })
  resource.roles = value.map(r => r.roleID)
}

function getResolvedRoles({ roles = [] }) { return roles.map(roleID => resolvedRoles.value[roleID]) }
function isRoleVisible({ meta }) { return !(meta?.context?.resourceTypes) }
function onFilter(filter = [], index) { options.value.filterPresets[index].filter = filter }
function addFilterPreset() { options.value.filterPresets.push({ name: '', filter: [], roles: [] }) }
function setDefaultValues() { checkboxLabel.value = {}; resolvedRoles.value = {} }
function onUpdateInlineEditableFields(fields = []) { options.value.inlineEditFields = fields.map(f => f.fieldID && f.fieldID !== NoID ? f.fieldID : f.name).filter(f => !!f) }
function onUpdateSearchableFields(fields = []) { options.value.searchableFields = fields.map(f => f.fieldID && f.fieldID !== NoID ? f.fieldID : f.name).filter(f => !!f) }
function onUpdateTextWrapOption(fields = []) { if (options.value.textStyles?.wrappedFields) options.value.textStyles.wrappedFields = fields.map(f => f.fieldID && f.fieldID !== NoID ? f.fieldID : f.name).filter(f => !!f) }
function addSummary() { options.value.summaries.push({ label: '', field: '', metric: '', roles: [] }) }
</script>
<style>
.w-fit { width: fit-content; }
</style>
<style lang="scss" scoped>
.list-background { background-color: var(--body-bg); }
</style>
