<template>
  <Wrap v-bind="{ ...$props, ...$attrs }" :scrollable-body="false" @refreshBlock="refresh(true, false)">
    <template v-if="recordListModule && isFederated" #title-badge>
      <span class="badge bg-primary d-inline-block mb-0 ms-2">{{ $t('recordList.federated') }}</span>
    </template>

    <template #toolbar>
      <div v-if="recordListModule && needHeaderBlock" ref="toolbar" class="d-flex flex-column gap-2 p-3 d-print-none">
        <div class="d-flex align-items-center justify-content-between gap-1">
          <div class="d-flex align-items-center flex-grow-1 flex-wrap flex-fill-child gap-1">
            <template v-if="recordListModule.canCreateRecord">
              <template v-if="inlineEditing">
                <button v-if="!options.hideAddButton" data-test-id="button-add-record" class="btn btn-primary" @click="addInlineRecord()">+ {{ $t('recordList.addRecord') }}</button>
              </template>
              <template v-else-if="!inlineEditing && (recordPageID || options.allRecords)">
                <button v-if="!options.hideAddButton" data-test-id="button-add-record" class="btn btn-primary" @click="handleAddRecord()">+ {{ $t('recordList.addRecord') }}</button>
                <ImporterModal v-if="!options.hideImportButton" :module="recordListModule" :namespace="namespace" @importSuccessful="onImportSuccessful" />
              </template>
            </template>
            <ExporterModal v-if="options.allowExport && !inlineEditing" :module="recordListModule" :filter="filter.query" :selection="selected" :selected-all-records="selectedAllRecords" :processing="processing" :preselected-fields="fields.map(({ moduleField }) => moduleField)" @export="onExport" />

            <div v-if="filterPresets.length" ref="filterPresets" class="dropdown">
              <button class="btn btn-outline-secondary dropdown-toggle" data-bs-toggle="dropdown" aria-expanded="false">{{ $t('recordList.filter.filters.label') }}</button>
              <ul class="dropdown-menu shadow-sm">
                <li v-for="(f, idx) in filterPresets" :key="idx" class="d-flex align-items-center justify-content-between">
                  <button class="dropdown-item" @click="updateFilter(f.filter, f.name)">{{ f.name }}</button>
                  <c-input-confirm v-if="!f.roles" show-icon class="me-1" @confirmed="removeRecordListFilterPreset(f.name)" />
                </li>
              </ul>
            </div>

            <ColumnPicker v-if="!options.hideConfigureFieldsButton" :module="recordListModule" :fields="fields.map(({ moduleField }) => moduleField)" @updateFields="onUpdateFields">{{ $t('module.allRecords.columns.title') }}</ColumnPicker>
          </div>
          <div v-if="!options.hideSearch" class="flex-fill">
            <c-input-search :value="query" :placeholder="$t('label.search', { default: 'Search' })" :ai="false" submittable @search="handleSearch" @ai-search="handleAiSearch" />
          </div>
        </div>

        <div v-if="options.showDeletedRecordsOption || groupRecordListFilter.length" class="d-flex align-items-start flex-wrap gap-1">
          <div v-if="groupedByConnector.length" class="d-flex align-items-center flex-wrap gap-2">
            <div v-for="(segment, segmentIdx) in groupedByConnector" :key="`segment-${segmentIdx}`" class="d-flex align-items-center gap-2">
              <div class="d-flex flex-wrap align-items-center gap-1 border rounded p-1">
                <template v-for="(filterGroup, groupIdx) in segment.groups" :key="`group-${filterGroup.originalIndex}`">
                  <div  class="d-flex flex-wrap align-items-center gap-1">
                    <div v-for="(f, filterIndex) in filterGroup.filter" :key="filterIndex" class="active-filter d-flex align-items-center rounded gap-1 ps-2 pe-1 py-1 bg-light">
                      <span class="field-label">{{ f.label || f.name }}</span>
                      <span>{{ $t(`recordList.filter.operatorLabels.${formatActiveFilterOperator(f.operator)}`) }}</span>
                      <template v-if="f.value">
                        <template v-if="isBetweenOperator(f.operator)">
                          <FieldViewer v-if="f.value.start" value-only :field="f.field" :record="f.record[0]" :module="recordListModule" :namespace="namespace" class="fw-bold text-primary" />
                          <span v-else class="text-primary fw-bold">{{ $t('recordList.filter.nil') }}</span>
                          <span class="text-lowercase">{{ $t('recordList.filter.conditions.and') }}</span>
                          <FieldViewer v-if="f.value.end" value-only :field="f.field" :record="f.record[1]" :module="recordListModule" :namespace="namespace" class="fw-bold text-primary" />
                          <span v-else class="text-primary fw-bold">{{ $t('recordList.filter.nil') }}</span>
                        </template>
                        <FieldViewer v-else value-only :field="f.field" :record="f.record" :module="recordListModule" :namespace="namespace" class="fw-bold text-primary" />
                      </template>
                      <span v-else class="text-primary fw-bold">{{ $t('recordList.filter.nil') }}</span>
                      <button class="btn btn-outline-secondary d-flex align-items-center p-1 active-filter-close-btn bg-transparent border-0" @click="removeFilter(filterGroup.originalIndex, filterIndex)">
                        <font-awesome-icon :icon="['fas', 'times']" />
                      </button>
                    </div>
                  </div>
                  <span v-if="groupIdx < segment.groups.length - 1" :key="`and-${filterGroup.originalIndex}`" class="text-secondary text-uppercase">{{ $t('recordList.filter.conditions.and') }}</span>
                </template>
              </div>
              <span v-if="segment.connector" class="text-secondary text-uppercase">{{ $t('recordList.filter.conditions.or') }}</span>
            </div>
            <button v-if="groupRecordListFilter.length" class="btn btn-outline-extra-light btn-sm text-primary border-0 text-nowrap" @click="resetFilter()">{{ $t('recordList.filter.reset') }}</button>
          </div>
          <div v-if="options.showDeletedRecordsOption" class="d-flex align-items-center ms-auto">
            <button class="btn btn-outline-extra-light btn-sm text-primary border-0 text-nowrap" @click="handleShowDeleted()">{{ showingDeletedRecords ? $t('recordList.showRecords.existing') : $t('recordList.showRecords.deleted') }}</button>
          </div>
        </div>

        <div v-if="(options.selectable && selected.length) || (inlineEditing && dirtyRecordsCount > 1)" class="d-flex align-items-center flex-wrap align-items-center">
          <div v-if="options.selectable && selected.length" class="me-1">{{ selectedRecordsDisplayText }}</div>
          <button v-if="!inlineEditing && options.selectable && selected.length" class="btn btn-outline-extra-light btn-sm text-primary border-0" @click="selectAllRecords()">{{ selectedAllRecords ? $t('recordList.unselectAllRecords') : $t('recordList.selectAllRecords') }}</button>
          <div class="d-flex align-items-center ms-auto gap-1">
            <AutomationButtons v-if="options.selectable && selected.length" class="d-inline m-0" :buttons="options.selectionButtons" :module="recordListModule" :extra-event-args="{ selected, filter }" v-bind="$props" @refresh="refresh()" />
            <div v-if="processingDirtyRecords === 'save'" class="spinner-border spinner-border-sm text-secondary" />
            <button v-else-if="inlineEditing && dirtyRecordsCount > 0" class="btn btn-outline-extra-light d-flex align-items-center justify-content-center border-0" style="width: 2rem; height: 2rem;" data-bs-toggle="tooltip" :title="$t('recordList.tooltip.saveChanges')" :disabled="!!processingDirtyRecords" @click="handleSaveDirtyRecords()">
              <font-awesome-icon :icon="['fas', 'check']" class="text-primary" />
            </button>
            <div v-if="processingDirtyRecords === 'deny'" class="spinner-border spinner-border-sm text-secondary" />
            <button v-else-if="inlineEditing && dirtyRecordsCount > 0" class="btn btn-outline-extra-light d-flex align-items-center justify-content-center border-0" style="width: 2rem; height: 2rem;" data-bs-toggle="tooltip" :title="$t('recordList.tooltip.discardChanges')" :disabled="!!processingDirtyRecords" @click="handleDenyDirtyRecords()">
              <font-awesome-icon :icon="['fas', 'times']" class="text-secondary" />
            </button>
            <div v-if="inlineEditing && dirtyRecordsCount > 0 && options.selectable && selected.length && ((options.bulkRecordEditEnabled && canUpdateSelectedRecords && !showingDeletedRecords) || (canDeleteSelectedRecords && !areAllRowsDeleted))" class="border-start mx-1" style="height: 1.5rem;" />
            <BulkEditModal v-show="options.bulkRecordEditEnabled && canUpdateSelectedRecords && !showingDeletedRecords" :module="recordListModule" :namespace="namespace" :query="bulkQuery" allow-add-field @save="onBulkUpdate()" />
            <template v-if="canDeleteSelectedRecords && !areAllRowsDeleted">
              <c-input-confirm show-icon :tooltip="$t('recordList.tooltip.deleteSelected')" :button-style="{ width: '2rem', height: '2rem' }" @confirmed="handleDeleteSelectedRecords()" />
            </template>
            <template v-if="canRestoreSelectedRecords && areAllRowsDeleted">
              <c-input-confirm show-icon :icon="['fas', 'trash-restore']" :tooltip="$t('recordList.tooltip.restoreSelected')" variant="outline-warning" variant-ok="warning" @confirmed="handleRestoreSelectedRecords()" />
            </template>
          </div>
        </div>
      </div>
    </template>

    <template #default>
      <div v-if="recordListModule" class="d-flex position-relative h-100 rl-root" :class="[rlDisplayClass, { 'overflow-hidden': showListLoader || !items.length, 'rl-compact': options.compactRows, 'rl-align-numbers': options.alignNumbers }]">
        <button v-if="!block.options?.hideBrainButton" class="brain-button position-absolute d-flex align-items-center justify-content-center d-print-none" @click="promptAiChat" :title="$t('ai.askAboutRecord')">
          <font-awesome-icon :icon="['fas', 'brain']" />
        </button>

        <div class="rl-table-wrap table-responsive flex-grow-1">
        <table data-test-id="table-record-list" class="table record-list-table mh-100 h-100 mb-0 table-hover" :class="{ 'table-sm': options.compactRows }">
          <thead :class="{ 'sticky-top': options.stickyHeader !== false }">
            <tr :class="showingDeletedRecords ? 'table-warning' : ''">
              <th v-if="options.draggable && inlineEditing" style="width: 0%"></th>
              <th v-if="options.selectable" style="width: 0%;" class="d-print-none rl-check-col">
                <input type="checkbox" class="form-check-input rl-row-check ms-1" :disabled="disableSelectAll" :checked="areAllRowsSelected && !disableSelectAll" @change="handleSelectAllOnPage({ isChecked: $event.target.checked })" />
              </th>
              <th v-if="options.showRowSignal" class="rl-signal-col" style="width: 0%"></th>
              <th v-if="isFederated" style="width: 0%"></th>
              <th v-for="(field, fieldIndex) in fields" :key="field.key" :colspan="fieldIndex === (fields.length - 1) ? 2 : 1" :style="{ 'padding-right': fieldIndex === (fields.length - 1) ? '15px' : '' }" :class="fieldCellClass(field)">
                <div class="d-flex align-items-center">
                  <div :class="{ required: field.required }" class="d-flex align-self-center text-nowrap">{{ field.label }}</div>
                  <button v-if="field.sortable" class="btn btn-outline-extra-light d-flex align-items-center text-secondary d-print-none border-0 px-1 ms-1" data-bs-toggle="tooltip" :title="$t('recordList.sort.tooltip')" @click="handleSort(field)">
                    <span class="fa-layers d-print-none">
                      <font-awesome-icon :icon="['fas', 'angle-up']" class="mb-1" :class="{ 'text-primary': isSortedBy(field, 'ASC') }" />
                      <font-awesome-icon :icon="['fas', 'angle-down']" class="mt-1" :class="{ 'text-primary': isSortedBy(field, 'DESC') }" />
                    </span>
                  </button>
                  <RecordListFilter v-if="!options.hideFiltering && field.filterable" :target="uniqueID" :selected-field="field.moduleField" :namespace="namespace" :module="recordListModule" variant="outline-extra-light" :record-list-filter="recordListFilter" :allow-filter-preset-save="options.customFilterPresets" class="d-print-none ms-1" @filter="onFilter" @filter-preset="onSaveFilterPreset" />
                </div>
              </th>
            </tr>
          </thead>
          <draggable v-if="items.length && !showListLoader && !resizing && !isGrouped" item-key="id" v-model="items" :disabled="!inlineEditing || !options.draggable" group="items" tag="tbody" handle=".handle">
            <template #item="{ element, index }">
              <tr :key="`${index}${element.r.recordID}`" :class="rowClass(element)" @click="handleRowClick(element)" @mouseenter="showRowValueTooltip($event, element)" @mousemove="moveRowValueTooltip($event)" @mouseleave="hideRowValueTooltip()">
                <td v-if="options.draggable && inlineEditing" class="pe-0" @click.stop>
                  <font-awesome-icon :icon="['fas', 'bars']" class="handle text-secondary mt-2" style="padding-top: 0.2rem;" />
                </td>
                <td v-if="options.selectable" class="pe-0 d-print-none rl-check-col" @click.stop>
                  <input type="checkbox" class="form-check-input rl-row-check ms-1" :class="{ 'mt-2': inlineEditing }" :checked="selected.includes(element.id)" @change="onSelectRow($event.target.checked, element)" />
                </td>
                <td v-if="options.showRowSignal" class="rl-signal-col pe-0">
                  <span class="rl-signal" :class="signalClass(element)" :title="signalTitle(element)" />
                </td>
                <td v-if="isFederated" class="align-middle ps-0">
                  <span v-if="Object.keys(element.r.labels || {}).includes('federation')" class="badge bg-primary align-text-top">F</span>
                </td>
                <td v-for="field in fields" :key="field.key" class="record-value" :class="fieldCellClass(field)">
                  <FieldEditor v-if="field.moduleField.canUpdateRecordValue && field.editable && isFieldEditable(field.moduleField)" :field="field.moduleField" value-only :record="element.r" :module="module" :namespace="namespace" :errors="recordErrors(element, field)" class="mb-0" style="min-width: 250px;" @click.stop @change="onInlineFieldChange(element)" />
                  <div v-else-if="field.moduleField.canReadRecordValue && !field.edit" class="d-flex flex-column mb-0 gap-1" style="min-width: 10rem;">
                    <div class="d-flex mb-0 gap-1 align-items-center">
                      <FieldViewer :field="field.moduleField" value-only :record="element.r" :module="module" :namespace="namespace" :extra-options="options" include-styles />
                      <div v-if="showInlineActions(field)" class="d-flex flex-nowrap align-items-start gap-1 inline-actions">
                        <button v-if="showInlineEdit(field)" class="btn btn-outline-extra-light btn-sm text-secondary border-0" data-bs-toggle="tooltip" :title="$t('recordList.inlineEdit.button.title')" @click.stop="editInlineField(element.r, field.key)">
                          <font-awesome-icon :icon="['fas', 'pen']" />
                        </button>
                        <button v-if="showInlineFilter(field)" class="btn btn-outline-extra-light btn-sm text-secondary border-0" data-bs-toggle="tooltip" :title="$t('recordList.filterByValue')" @click.stop="filterByValue(element.r, field)">
                          <font-awesome-icon :icon="['fas', 'filter']" />
                        </button>
                      </div>
                    </div>
                    <div v-if="isSparklineField(field)" class="rl-sparkline" :title="String(recordFieldRaw(element.r, field.key) ?? '')">
                      <span class="rl-sparkline-bar" :style="{ width: sparklinePct(element) + '%' }" :class="signalClass(element)" />
                    </div>
                  </div>
                  <i v-else class="text-primary">{{ $t('field.noPermission') }}</i>
                </td>
                <td class="actions px-2" :class="{ 'actions-visible': inlineEditing && !editing && showSaveAction(element) }" @click.stop>
                  <div class="d-flex align-items-center justify-content-end gap-1" :class="{ 'mt-2': inlineEditing }">
                    <div v-if="processingInlineRecords[element.id] === 'save'" class="spinner-border spinner-border-sm text-primary" />
                    <button v-else-if="inlineEditing && !editing && showSaveAction(element)" class="btn btn-outline-extra-light d-flex align-items-center justify-content-center border-0" style="width: 2rem; height: 2rem;" data-bs-toggle="tooltip" :title="$t('recordList.tooltip.saveChanges')" :disabled="!!processingInlineRecords[element.id]" @click.stop="handleSaveInline(element, index)">
                      <font-awesome-icon :icon="['fas', 'check']" class="text-primary" />
                    </button>
                    <div v-if="processingInlineRecords[element.id] === 'deny'" class="spinner-border spinner-border-sm text-secondary" />
                    <button v-else-if="inlineEditing && !editing && showSaveAction(element)" class="btn btn-outline-extra-light d-flex align-items-center justify-content-center border-0" style="width: 2rem; height: 2rem;" data-bs-toggle="tooltip" :title="$t('recordList.tooltip.discardChanges')" :disabled="!!processingInlineRecords[element.id]" @click.stop="handleDenyInline(element, index)">
                      <font-awesome-icon :icon="['fas', 'times']" class="text-secondary" />
                    </button>
                    <div v-if="inlineEditing && !editing && showSaveAction(element) && areActionsVisible(element.r)" class="border-start mx-1" style="height: 1.5rem;" />
                    <div v-if="areActionsVisible(element.r)" class="dropdown dropstart">
                      <button class="btn btn-outline-extra-light d-flex align-items-center justify-content-center border-0 dropdown-toggle" style="width: 2rem; height: 2rem;" data-bs-toggle="dropdown" aria-expanded="false">
                        <font-awesome-icon :icon="['fas', 'ellipsis-v']" class="text-primary" />
                      </button>
                      <ul class="dropdown-menu m-0">
                        <template v-if="inlineEditing && editing">
                          <li v-if="isCloneRecordActionVisible">
                            <button class="dropdown-item" @click="handleCloneInline(element.r)">
                              <font-awesome-icon :icon="['far', 'clone']" class="text-primary" /> {{ $t('recordList.record.tooltip.clone') }}
                            </button>
                          </li>
                          <li v-if="isInlineRestoreActionVisible(element.r)" class="w-100">
                            <c-input-confirm :text="$t('recordList.record.tooltip.restore')" :icon="['fas', 'trash-restore']" show-icon borderless variant="link" variant-ok="warning" size="md" button-class="dropdown-item" icon-class="text-warning" class="w-100" @confirmed="handleRestoreInline(element, index)" />
                          </li>
                          <li v-else-if="isInlineDeleteActionVisible(element.r)">
                            <button class="dropdown-item" @click.prevent="handleDeleteInline(element, index)">
                              <font-awesome-icon :icon="['far', 'trash-alt']" class="text-danger" /> {{ $t('recordList.record.tooltip.delete') }}
                            </button>
                          </li>
                        </template>
                        <template v-else>
                          <li v-if="isViewRecordActionVisible(element.r)">
                            <router-link class="dropdown-item" :to="viewRecordRoute(element.r.recordID)">
                              <font-awesome-icon :icon="['far', 'file-alt']" class="text-primary" /> {{ $t('recordList.record.tooltip.view') }}
                            </router-link>
                          </li>
                          <li v-if="isEditRecordActionVisible(element.r)">
                            <router-link class="dropdown-item" :to="editRecordRoute(element.r.recordID)">
                              <font-awesome-icon :icon="['far', 'edit']" class="text-primary" /> {{ $t('recordList.record.tooltip.edit') }}
                            </router-link>
                          </li>
                          <li v-if="isCloneRecordActionVisible">
                            <button class="dropdown-item" @click="handleCloneRecordAction(element.r.recordID, element.r.values)">
                              <font-awesome-icon :icon="['far', 'clone']" class="text-primary" /> {{ $t('recordList.record.tooltip.clone') }}
                            </button>
                          </li>
                          <li v-if="isReminderActionVisible">
                            <button class="dropdown-item" @click="createReminder(element.r)">
                              <font-awesome-icon :icon="['far', 'bell']" class="text-primary" /> {{ $t('recordList.record.tooltip.reminder') }}
                            </button>
                          </li>
                          <li v-if="isRecordPermissionButtonVisible(element.r)">
                            <c-permissions-button :resource="`corteza::compose:record/${element.r.namespaceID}/${element.r.moduleID}/${element.r.recordID}`" :target="element.r.recordID" :title="element.r.recordID" :button-label="$t('recordList.record.tooltip.permissions')" class="dropdown-item" />
                          </li>
                          <li v-if="isDeleteActionVisible(element.r)" class="w-100">
                            <c-input-confirm :text="$t('recordList.record.tooltip.delete')" show-icon borderless variant="link" size="md" button-class="dropdown-item" icon-class="text-danger" class="w-100" @confirmed="handleDeleteSelectedRecords(element.r.recordID)" />
                          </li>
                          <li v-else-if="isRestoreActionVisible(element.r)" class="w-100">
                            <c-input-confirm :text="$t('recordList.record.tooltip.restore')" :icon="['fas', 'trash-restore']" show-icon borderless variant="link" variant-ok="warning" size="md" button-class="dropdown-item" icon-class="text-warning" class="w-100" @confirmed="handleRestoreSelectedRecords(element.r.recordID)" />
                          </li>
                        </template>
                      </ul>
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </draggable>
          <tbody v-else-if="items.length && !showListLoader && !resizing && isGrouped">
            <template v-for="group in groupedItems" :key="group.key">
              <tr class="rl-group-header">
                <td :colspan="tableColSpan">
                  <span class="rl-group-label">{{ group.label }}</span>
                  <span class="rl-group-count text-muted ms-2">{{ group.items.length }}</span>
                </td>
              </tr>
              <tr v-for="(element, index) in group.items" :key="`${group.key}-${index}${element.r.recordID}`" :class="rowClass(element)" @click="handleRowClick(element)" @mouseenter="showRowValueTooltip($event, element)" @mousemove="moveRowValueTooltip($event)" @mouseleave="hideRowValueTooltip()">
                <td v-if="options.selectable" class="pe-0 d-print-none rl-check-col" @click.stop>
                  <input type="checkbox" class="form-check-input rl-row-check ms-1" :checked="selected.includes(element.id)" @change="onSelectRow($event.target.checked, element)" />
                </td>
                <td v-if="options.showRowSignal" class="rl-signal-col pe-0">
                  <span class="rl-signal" :class="signalClass(element)" :title="signalTitle(element)" />
                </td>
                <td v-if="isFederated" class="align-middle ps-0">
                  <span v-if="Object.keys(element.r.labels || {}).includes('federation')" class="badge bg-primary align-text-top">F</span>
                </td>
                <td v-for="field in fields" :key="field.key" class="record-value" :class="fieldCellClass(field)">
                  <div v-if="field.moduleField.canReadRecordValue" class="d-flex flex-column mb-0 gap-1" style="min-width: 10rem;">
                    <FieldViewer :field="field.moduleField" value-only :record="element.r" :module="module" :namespace="namespace" :extra-options="options" include-styles />
                    <div v-if="isSparklineField(field)" class="rl-sparkline">
                      <span class="rl-sparkline-bar" :style="{ width: sparklinePct(element) + '%' }" :class="signalClass(element)" />
                    </div>
                  </div>
                </td>
                <td class="actions px-2" @click.stop />
              </tr>
            </template>
          </tbody>
          <tbody v-else-if="showListLoader || showListEmpty">
            <tr>
              <td :colspan="tableColSpan" class="border-0">
                <div class="rl-empty d-print-none">
                  <div v-if="showListLoader" class="spinner-border text-primary" role="status">
                    <span class="visually-hidden">Loading…</span>
                  </div>
                  <template v-else>
                    <div class="rl-empty-icon">
                      <font-awesome-icon :icon="['fas', 'box-archive']" />
                    </div>
                    <p class="rl-empty-title mb-1">{{ $t('recordList.empty.title') }}</p>
                    <p class="rl-empty-text mb-0 text-muted">{{ $t('recordList.noRecords') }}</p>
                  </template>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        </div>

        <div class="rl-cards-wrap flex-grow-1">
          <div v-if="showListLoader" class="rl-empty">
            <div class="spinner-border text-primary" role="status">
              <span class="visually-hidden">Loading…</span>
            </div>
          </div>
          <div v-else-if="showListEmpty" class="rl-empty">
            <div class="rl-empty-icon"><font-awesome-icon :icon="['fas', 'box-archive']" /></div>
            <p class="rl-empty-title mb-1">{{ $t('recordList.empty.title') }}</p>
            <p class="rl-empty-text mb-0 text-muted">{{ $t('recordList.noRecords') }}</p>
          </div>
          <template v-else>
            <template v-for="group in groupedItems" :key="'card-'+group.key">
              <div v-if="isGrouped" class="rl-card-group-title">{{ group.label }} <span class="text-muted">({{ group.items.length }})</span></div>
              <div class="rl-cards">
                <button
                  v-for="(element, index) in group.items"
                  :key="'c'+index+element.r.recordID"
                  type="button"
                  class="rl-card"
                  :class="rowClass(element)"
                  @click="handleRowClick(element)"
                  @mouseenter="showRowValueTooltip($event, element)"
                  @mousemove="moveRowValueTooltip($event)"
                  @mouseleave="hideRowValueTooltip()"
                >
                  <div class="rl-card-head">
                    <span v-if="options.showRowSignal" class="rl-signal" :class="signalClass(element)" />
                    <span class="rl-card-title text-truncate">{{ cardTitle(element) }}</span>
                  </div>
                  <div class="rl-card-body">
                    <div v-for="field in cardFields" :key="field.key" class="rl-card-row">
                      <span class="rl-card-label">{{ field.label }}</span>
                      <FieldViewer :field="field.moduleField" value-only :record="element.r" :module="module" :namespace="namespace" :extra-options="options" include-styles class="rl-card-value" />
                    </div>
                    <div v-if="options.sparklineField" class="rl-sparkline mt-2">
                      <span class="rl-sparkline-bar" :style="{ width: sparklinePct(element) + '%' }" :class="signalClass(element)" />
                    </div>
                  </div>
                </button>
              </div>
            </template>
          </template>
        </div>
      </div>
      <div v-else-if="options.moduleID && options.moduleID !== NoID" class="rl-empty w-100">
        <div class="spinner-border text-primary" role="status">
          <span class="visually-hidden">Loading…</span>
        </div>
      </div>
      <label v-else class="text-primary p-3">{{ $t('recordList.noModule') }}</label>
    </template>

    <template v-if="recordListModule && showFooter" #footer>
      <div v-if="listSummaries.length || options.customSummaries" class="d-flex flex-wrap align-items-center">
        <div v-for="(summary, index) in listSummaries" :key="index" class="d-flex flex-wrap align-items-center border-end border-bottom p-1">
          <div class="d-flex align-items-center px-3 py-2 mb-0" :class="{ 'custom-summary': !!summary.custom }" @click="openCustomSummaryModal(summary)">
            {{ summary.label }}:
            <label v-if="!isProcessing" class="ms-2 mb-0">{{ summary.value }}</label>
            <div v-else class="spinner-border spinner-border-sm text-secondary ms-1" />
          </div>
        </div>
        <div v-if="options.customSummaries" class="d-flex align-items-center flex-fill border-bottom">
          <button class="btn btn-outline-extra-light text-secondary border-0 py-2 m-1" data-bs-toggle="tooltip" :title="$t('recordList.summaries.customSummaries.add.tooltip')" @click="openCustomSummaryModal()">
            <font-awesome-icon :icon="['fas', 'plus']" /> {{ $t('recordList.summaries.customSummaries.add.label') }}
          </button>
        </div>
      </div>

      <div v-if="showPagination" class="record-list-footer d-flex align-items-center flex-wrap justify-content-between px-3 py-2 gap-1">
        <div class="d-flex align-items-center flex-wrap gap-3">
          <div v-if="options.showTotalCount" class="text-nowrap text-truncate">
            <span v-if="pagination.count > recordsPerPage" data-test-id="pagination-range">{{ $t('recordList.pagination.showing', getPagination) }}</span>
            <span v-else data-test-id="pagination-single-number">{{ $t(`recordList.pagination.single_${pagination.count === 1 ? 'one' : 'other'}`, getPagination) }}</span>
          </div>
          <div v-if="options.showRecordPerPageOption" class="d-flex align-items-center gap-1 text-nowrap">
            <span>{{ $t('recordList.pagination.recordsPerPage') }}</span>
            <select v-model="recordsPerPage" class="form-select form-control form-select-sm" @change="handlePerPageChange">
              <option v-for="opt in perPageOptions" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
            </select>
          </div>
        </div>
        <div v-if="showPageNavigation" class="d-flex align-items-center justify-content-end">
          <nav v-if="options.fullPageNavigation" aria-label="Record list pagination">
            <ul class="pagination pagination-sm m-0 d-print-none">
              <li class="page-item" :class="{ disabled: getPagination.page <= 1 }">
                <button class="page-link" @click="goToPage(1)" :disabled="getPagination.page <= 1"><font-awesome-icon :icon="['fas', 'angle-double-left']" /></button>
              </li>
              <li class="page-item" :class="{ disabled: !hasPrevPage || isProcessing }">
                <button class="page-link" @click="goToPage(getPagination.page - 1)" :disabled="!hasPrevPage || isProcessing"><font-awesome-icon :icon="['fas', 'angle-left']" /></button>
              </li>
              <li v-for="p in pagination.pages" :key="p.page" class="page-item" :class="{ active: p.page === getPagination.page }">
                <button class="page-link" @click="goToPage(p.page)">{{ p.label || p.page }}</button>
              </li>
              <li class="page-item" :class="{ disabled: !hasNextPage || isProcessing }">
                <button class="page-link" @click="goToPage(getPagination.page + 1)" :disabled="!hasNextPage || isProcessing"><font-awesome-icon :icon="['fas', 'angle-right']" /></button>
              </li>
              <li class="page-item" :class="{ disabled: getPagination.page >= getPagination.count / getPagination.perPage }">
                <button class="page-link" @click="goToPage(Math.ceil(getPagination.count / getPagination.perPage))"><font-awesome-icon :icon="['fas', 'angle-double-right']" /></button>
              </li>
            </ul>
          </nav>
          <div v-else class="btn-group gap-1">
            <button class="btn btn-outline-extra-light d-flex align-items-center justify-content-center text-dark border-0 p-1" :disabled="!hasPrevPage || isProcessing" data-test-id="first-page" @click="goToPage(1)">
              <font-awesome-icon :icon="['fas', 'angle-double-left']" />
            </button>
            <button class="btn btn-outline-extra-light d-flex align-items-center justify-content-center text-dark border-0 p-1" :disabled="!hasPrevPage || isProcessing" data-test-id="previous-page" @click="goToPage(getPagination.page - 1)">
              <font-awesome-icon :icon="['fas', 'angle-left']" class="me-1" /> {{ $t('recordList.pagination.prev') }}
            </button>
            <button class="btn btn-outline-extra-light d-flex align-items-center justify-content-center text-dark border-0 p-1" :disabled="!hasNextPage || isProcessing" data-test-id="next-page" @click="goToPage(getPagination.page + 1)">
              {{ $t('recordList.pagination.next') }} <font-awesome-icon :icon="['fas', 'angle-right']" class="ms-1" />
            </button>
          </div>
        </div>
      </div>

      <BulkEditModal v-if="options.inlineRecordEditEnabled" :namespace="namespace" :module="recordListModule" :selected-fields="inlineEdit.fields" :initial-record="inlineEdit.record" :query="inlineEdit.query" :modal-title="$t('recordList.inlineEdit.modal.title')" open-on-select :allow-add-field="options.inlineRecordEditAllowAddField" @save="onInlineEdit()" @close="onInlineEditClose()" />
      <CustomFilterPreset v-if="options.customFilterPresets" :visible="showCustomPresetFilterModal" @save="setStorageRecordListFilterPreset" @close="showCustomPresetFilterModal = false" />
      <CustomSummary v-if="options.customSummaries" :visible="showCustomSummariesModal" :module="recordListModule" :summary="customSummary" :summary-index="customSummaryIndex" @save="onCustomSummarySave" @delete="onCustomSummaryDelete" @close="onCustomSummaryClose" />
    </template>
  </Wrap>
  <Teleport to="body">
    <div
      v-if="rowTooltip.visible && rowTooltip.record"
      class="rl-row-tooltip"
      :style="rowTooltip.style"
      role="tooltip"
    >
      <div v-if="rowTooltip.title" class="rl-row-tooltip-title">{{ rowTooltip.title }}</div>
      <div
        v-for="(line, idx) in rowTooltip.lines"
        :key="idx"
        class="rl-row-tooltip-row"
      >
        <span
          class="rl-row-tooltip-marker"
          :style="{ backgroundColor: line.color }"
        />
        <span class="rl-row-tooltip-name">{{ line.label }}</span>
        <span class="rl-row-tooltip-value">
          <FieldViewer
            v-if="line.moduleField?.canReadRecordValue !== false"
            :field="line.moduleField"
            value-only
            :record="rowTooltip.record"
            :module="recordListModule"
            :namespace="namespace"
            :extra-options="options"
            include-styles
          />
          <span v-else>—</span>
        </span>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
defineOptions({ inheritAttrs: false, i18nOptions: { namespaces: 'block' } })
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount, nextTick, inject } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useStore } from '../../store'
import { NoID, compose, validator } from 'corteza-lib/js/dist'
import { components, url } from 'corteza-lib/vue/dist'
import axios from 'axios'
import ColumnPicker from 'corteza-webapp-compose/src/components/Admin/Module/Records/ColumnPicker'
import RecordListFilter from 'corteza-webapp-compose/src/components/Common/RecordListFilter'
import FieldEditor from 'corteza-webapp-compose/src/components/ModuleFields/Editor'
import FieldViewer from 'corteza-webapp-compose/src/components/ModuleFields/Viewer'
import CustomFilterPreset from 'corteza-webapp-compose/src/components/PageBlocks/RecordList/CustomFilterPreset'
import CustomSummary from 'corteza-webapp-compose/src/components/PageBlocks/RecordList/CustomSummary'
import AutomationButtons from 'corteza-webapp-compose/src/components/PageBlocks/Shared/AutomationButtons.vue'
import BulkEditModal from 'corteza-webapp-compose/src/components/Public/Record/BulkEdit'
import ExporterModal from 'corteza-webapp-compose/src/components/Public/Record/Exporter'
import ImporterModal from 'corteza-webapp-compose/src/components/Public/Record/Importer'
import { getItem, removeItem, setItem } from 'corteza-webapp-compose/src/lib/local-storage'
import { evalPrefilterOrSkip, formatActiveFilterOperator, isBetweenOperator, isFieldInFilter, queryToFilter, convertRecordListFilter, getFieldFilter } from 'corteza-webapp-compose/src/lib/record-filter'
import draggable from 'vuedraggable'
import Wrap from './Wrap/index.js'
import { usePageBlockBase } from './usePageBlockBase'

const { CInputSearch } = components
const { t: _$t } = useI18n({ useScope: 'global' })
const $t = (...args) => {
  try {
    const result = _$t(...args)
    if (result === null || result === undefined) {
      return typeof args[1] === 'object' && args[1]?.default ? args[1].default : args[0]
    }
    return result
  } catch (e) {
    return typeof args[1] === 'object' && args[1]?.default ? args[1].default : args[0]
  }
}
const store = useStore()
const $auth = inject('$auth')
const $ComposeAPI = inject('$ComposeAPI')
const $router = useRouter()
const $route = useRoute()

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

const emit = defineEmits(['errors', 'save-fields'])

const { processing, isProcessing, options, refreshBlock } = usePageBlockBase(props, emit)

const inlineErrors = ref(new validator.Validated())
const dirtyInlineRecords = ref({})
const processingDirtyRecords = ref('')
const processingInlineRecords = ref({})
const uniqueID = ref(undefined)
const prefilter = ref(undefined)
const recordListFilter = ref([])
const query = ref(null)
const filter = reactive({ query: '', sort: '', limit: 10, pageCursor: '', prevPage: '', nextPage: '' })
const pagination = reactive({ pages: [], page: 1, count: 0 })
const selected = ref([])
const inlineEdit = reactive({ fields: [], recordIDs: [], initialRecord: {} })
const sortBy = ref(undefined)
const sortDirection = ref(undefined)
const summaries = ref([])
const customSummaries = ref([])
const showCustomSummariesModal = ref(false)
const customSummaryIndex = ref(-1)
const customSummary = ref({ metric: '', field: '', label: '' })
let ctr = 0
const items = ref([])
const hasLoadedOnce = ref(false)
const showingDeletedRecords = ref(false)
const customPresetFilters = ref([])
const currentCustomPresetFilter = ref(undefined)
const showCustomPresetFilterModal = ref(false)
const selectedAllRecords = ref(false)
const abortableRequests = ref([])
const recordsPerPage = ref(undefined)
const customConfiguredFields = ref([])
let processingTimeout = undefined
const cancelled = ref(false)
const stayOnPage = ref(undefined)

const getModuleByID = computed(() => store.module.getByID)
const pages = computed(() => store.page.set)
const recordListModule = computed(() => options.value.moduleID ? getModuleByID.value(options.value.moduleID) : undefined)
const isFederated = computed(() => Object.keys(recordListModule.value?.labels || {}).includes('federation'))
const showPagination = computed(() => showPageNavigation.value || options.value.showTotalCount || options.value.showRecordPerPageOption)
const showFooter = computed(() => showPagination.value || options.value.customSummaries)
const perPageOptions = computed(() => {
  const defaultText = options.value.perPage === 0 ? $t('label.all') : String(options.value.perPage)
  return [{ text: defaultText, value: options.value.perPage }, { text: '25', value: 25 }, { text: '50', value: 50 }, { text: '100', value: 100 }]
    .filter((v, i) => i === 0 || v.value !== options.value.perPage)
    .sort((a, b) => { if (a.value === 0) return 1; if (b.value === 0) return -1; return a.value - b.value })
})
const getPagination = computed(() => {
  const { page = 1, count = 0 } = pagination
  const pp = recordsPerPage.value
  return { from: ((page - 1) * pp) + 1, to: pp > 0 ? Math.min(page * pp, count) : count, page, perPage: pp, count }
})
const needHeaderBlock = computed(() => recordListModule.value?.canCreateRecord || (options.value.allowExport && !inlineEditing.value) || filterPresets.value.length || !options.value.hideConfigureFieldsButton || !options.value.hideSearch)
const hasPrevPage = computed(() => !!filter.prevPage)
const hasNextPage = computed(() => !!filter.nextPage)
const editing = computed(() => props.mode === 'editor')
const showPageNavigation = computed(() => !options.value.hidePaging)
const isRowClickable = computed(() => !inlineEditing.value && options.value.recordDisplayOption !== 'doNothing')
const showPerPageSelector = computed(() => options.value.showRecordPerPageOption)
const disableSelectAll = computed(() => options.value.hidePaging ? !items.value.length : items.value.length === 0)
const inlineEditing = computed(() => !!options.value.editable)
const areAllRowsSelected = computed(() => selected.value.length === items.value.length)
const areAllRowsDeleted = computed(() => {
  const selItems = items.value.filter(({ id }) => selected.value.includes(id))
  return !!selected.value.length && !selItems.find(({ r }) => !r.deletedAt)
})
const recordPageID = computed(() => {
  let modID = recordListModule.value?.moduleID
  if (!modID) return undefined
  const rdtmid = options.value.recordDisplayTargetModuleID
  if (rdtmid !== undefined && rdtmid !== '0') modID = rdtmid
  return pages.value.find(p => p.moduleID === modID)?.pageID
})
const allFields = computed(() => recordListModule.value ? [...recordListModule.value.fields, ...recordListModule.value.systemFields()] : [])
const fields = computed(() => {
  let flds = []
  const editableFields = !inlineEditing.value ? [] : options.value.editFields.map(({ name }) => name)
  if (!options.value.hideConfigureFieldsButton && customConfiguredFields.value.length > 0) {
    flds = recordListModule.value.filterFields(customConfiguredFields.value)
  } else if (options.value.fields.length > 0) {
    flds = recordListModule.value.filterFields(options.value.fields)
  } else {
    flds = [...recordListModule.value.fields.slice(0, 5), ...recordListModule.value.systemFields()]
  }
  const configured = flds.map(mf => ({
    key: mf.name,
    label: mf.isSystem ? $t(`system.${mf.name}`) : mf.label || mf.name,
    moduleField: mf,
    sortable: !options.value.hideSorting && !(options.value.editable && editing.value) && !mf.isMulti && mf.isSortable,
    filterable: mf.isFilterable,
    tdClass: 'record-value',
    editable: !!editableFields.find(f => mf.name === f),
    canEdit: isFieldEditable(mf),
    required: inlineEditing.value && mf.isRequired,
  }))
  return [...configured]
})
const cardFields = computed(() => fields.value.slice(0, 4))
const isGrouped = computed(() => !!options.value.groupByField && !inlineEditing.value)
const rlDisplayClass = computed(() => {
  const mode = options.value.displayMode || 'table'
  if (mode === 'cards') return 'rl-display-cards'
  if (mode === 'responsive') return 'rl-display-responsive'
  return 'rl-display-table'
})
const tableColSpan = computed(() => {
  let n = fields.value.length + 1 // actions
  if (options.value.selectable) n++
  if (options.value.showRowSignal) n++
  if (isFederated.value) n++
  if (options.value.draggable && inlineEditing.value) n++
  return n
})
const groupedItems = computed(() => {
  const list = items.value || []
  const gbf = options.value.groupByField
  if (!gbf || inlineEditing.value) {
    return [{ key: '', label: '', items: list }]
  }
  const map = new Map()
  for (const el of list) {
    const raw = recordFieldRaw(el.r, gbf)
    const key = raw == null || raw === '' ? '__empty__' : String(raw)
    if (!map.has(key)) map.set(key, [])
    map.get(key).push(el)
  }
  return [...map.entries()].map(([key, groupItems]) => ({
    key,
    label: key === '__empty__' ? '—' : key,
    items: groupItems,
  }))
})
const canDeleteSelectedRecords = computed(() => items.value.filter(({ id, r }) => selected.value.includes(id) && r.canDeleteRecord).length > 0)
const canUpdateSelectedRecords = computed(() => items.value.filter(({ id, r }) => selected.value.includes(id) && r.canUpdateRecord).length > 0)
const canRestoreSelectedRecords = computed(() => items.value.filter(({ id, r }) => selected.value.includes(id) && r.canUndeleteRecord).length > 0)
const isCloneRecordActionVisible = computed(() => !options.value.hideRecordCloneButton && recordListModule.value?.canCreateRecord && (options.value.rowCreateUrl || recordPageID.value || inlineEditing.value))
const isReminderActionVisible = computed(() => !options.value.hideRecordReminderButton)
const filterPresets = computed(() => [...options.value.filterPresets.filter(({ name, roles }) => name && isUserRoleMember(roles)), ...customPresetFilters.value])
const authUserRoles = computed(() => $auth?.user?.roles || [])
const selectedRecordsDisplayText = computed(() => {
  const count = selectedAllRecords.value ? (options.value.showTotalCount ? pagination.count : undefined) : selected.value.length
  const total = items.value.length
  const key = selectedAllRecords.value ? 'selectedFromAllPages' : 'selected'
  return $t(`recordList.${key}`, { count, total })
})
const bulkQuery = computed(() => {
  if (selectedAllRecords.value) return filter.query
  return selected.value.map(r => `recordID='${r}'`).join(' OR ')
})
const isOnRecordPage = computed(() => props.page?.moduleID !== NoID)

// ${variables.x} in the prefilter, sourced from the page's session-only
// variable values (see PageBlocks/Variables).
const pageVariables = computed(() => store.pageVariables.getValuesForPage(props.page.pageID))

const groupRecordListFilter = computed(() => {
  return recordListFilter.value.map(group => {
    group.filter = convertRecordListFilter(group.filter.map(f => createDefaultFilter(f, f.value, f.operator)))
    return group
  }).filter(({ filter }) => filter.length)
})

const groupedByConnector = computed(() => {
  const result = []
  let currentAndGroup = []
  groupRecordListFilter.value.forEach((group, idx) => {
    const condition = group.groupCondition || 'OR'
    if (idx === 0 || condition === 'AND') {
      currentAndGroup.push({ ...group, originalIndex: idx })
    } else {
      if (currentAndGroup.length) result.push({ groups: currentAndGroup, connector: 'OR' })
      currentAndGroup = [{ ...group, originalIndex: idx }]
    }
  })
  if (currentAndGroup.length) result.push({ groups: currentAndGroup, connector: null })
  return result
})

const listSummaries = computed(() => {
  return [...options.value.summaries.filter(s => s.metric && s.field && isUserRoleMember(s.roles)), ...customSummaries.value.filter(s => s.metric && s.field).map(s => ({ ...s, custom: true }))]
    .map(s => {
      const name = `${s.metric} ${s.field}`
      const { value } = summaries.value[name] || {}
      return { custom: s.custom, name, label: s.label, field: s.field, metric: s.metric, value }
    })
})

const dirtyRecordsCount = computed(() => items.value.filter(item => showSaveAction(item)).length)
const showListLoader = computed(() => {
  if (!options.value.moduleID || options.value.moduleID === NoID) return false
  return isProcessing.value || !hasLoadedOnce.value
})
const showListEmpty = computed(() => hasLoadedOnce.value && !isProcessing.value && !items.value.length)

watch(() => options.value.moduleID, () => {
  hasLoadedOnce.value = false
})

watch(recordListModule, (mod, prev) => {
  if (mod && (!prev || prev.moduleID !== mod.moduleID) && !hasLoadedOnce.value) {
    prepRecordList()
    refresh(true)
  }
})

watch(() => options.value, () => {
  if (!props.loadingRecord) { prepRecordList(); refresh(true) }
}, { deep: true, immediate: true })

watch(() => [props.record?.recordID, props.loadingRecord], () => {
  if (props.loadingRecord) return
  createEvents()
  getCustomSummaries()
  getStorageRecordListFilter()
  getStorageRecordListFilterPreset()
  getStorageRecordListConfiguredFields()
  prepRecordList()
  refresh(true)
}, { immediate: true })

onMounted(() => {
  if (!inlineEditing.value) refreshBlock(refresh, false, true)
})

onBeforeUnmount(() => {
  hideRowValueTooltip()
  abortRequests()
  destroyEvents()
  setDefaultValues()
})

function createEvents() {
  const { pageID = NoID } = props.page
  const { recordID = NoID } = props.record || {}
  if (uniqueID.value) destroyEvents()
  uniqueID.value = [pageID, recordID, props.block.blockID, props.magnified].map(v => v || NoID).join('-')
  window.addEventListener(`record-line:collect:${uniqueID.value}`, resolveRecords)
  window.addEventListener(`page-block:validate:${uniqueID.value}`, validatePageBlock)
  window.addEventListener(`drill-down-recordList:${uniqueID.value}`, setDrillDownFilter)
  window.addEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.addEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.addEventListener('page-variable-change', refetchOnPageVariableChange)
  window.addEventListener('refetch-records', refreshAndResetPagination)
}

function refetchOnPrefilterValueChange({ detail: { fieldName } } = {}) {
  if (isFieldInFilter(fieldName, options.value.prefilter)) { prepRecordList(); refresh() }
}

function refetchOnPageVariableChange({ detail: { pageID, fieldName } } = {}) {
  if (pageID !== props.page.pageID) return
  if (isFieldInFilter(`variables.${fieldName}`, options.value.prefilter)) { prepRecordList(); refresh() }
}

function refreshOnRelatedRecordsUpdate({ detail: { moduleID, excludeUniqueID } = {} } = {}) {
  if (excludeUniqueID === uniqueID.value) return
  if (recordListModule.value?.moduleID === moduleID) { refresh(true); return }
  const recordFields = fields.value.filter(f => f.moduleField.kind === 'Record')
  if (recordFields.some(r => r.moduleField.options.moduleID === moduleID)) refresh(false)
}

function onFilter(filter = []) {
  filter.forEach(f => { f.name = $t('recordList.customFilter') })
  recordListFilter.value = filter
  setStorageRecordListFilter()
  refresh(true)
}

function handlePerPageChange() { filter.limit = recordsPerPage.value; refresh(true) }

function handleSearch(searchQuery) { query.value = searchQuery ? searchQuery.trim() : null; refresh(true) }

function handleAiSearch(searchQuery) { query.value = searchQuery ? searchQuery.trim() : null; promptAiChat() }

function onSaveFilterPreset(filter = []) {
  currentCustomPresetFilter.value = { filter }
  showCustomPresetFilterModal.value = true
}

function resetFilter() { onFilter() }

function onUpdateFields(fields = []) {
  options.value.fields = [...fields]
  customConfiguredFields.value = fields.map(f => f.isSystem ? f.name : f.fieldID).filter(f => !!f)
  setStorageRecordListConfiguredFields()
  emit('save-fields', options.value.fields)
}

function setStorageRecordListConfiguredFields() {
  try { setItem(`record-list-configured-columns-${uniqueID.value}`, customConfiguredFields.value) } catch (e) { console.warn($t('notification.record-list.corrupted-configured-fields')) }
}

function getStorageRecordListConfiguredFields() {
  try { customConfiguredFields.value = getItem(`record-list-configured-columns-${uniqueID.value}`) } catch (e) { console.warn($t('notification.record-list.corrupted-configured-fields')); removeItem(`record-list-configured-columns-${uniqueID.value}`) }
}

function onSelectRow(selectedFlag, item) {
  if (selectedFlag) {
    if (!selected.value.includes(item.id)) selected.value.push(item.id)
  } else {
    const i = selected.value.indexOf(item.id)
    if (i >= 0) selected.value.splice(i, 1)
    selectedAllRecords.value = false
  }
}

function isSortedBy({ key }, dir) {
  const { sort = '' } = filter
  const sortedFields = sort.includes(',') ? sort.split(',') : [sort]
  return sortedFields.map(v => v.trim()).some(value => {
    let valueDir = 'ASC'
    if (value.includes(' ')) { value = value.split(' ')[0]; valueDir = 'DESC' }
    return valueDir === dir && value === key
  })
}

function handleShowDeleted() { showingDeletedRecords.value = !showingDeletedRecords.value; selectedAllRecords.value = false; refresh(true) }

function recordErrors(item, field) {
  const id = `${item.id}:${field.key}`
  const errors = new validator.Validated()
  if (props.errors) errors.push(...props.errors.filterByMeta('id', item.id).filterByMeta('field', field.key).get())
  if (inlineErrors.value) errors.push(...inlineErrors.value.filterByMeta('id', item.id).filterByMeta('field', field.key).get())
  if (errors.set.length > 0) emit('errors', { errors, id })
  else emit('errors', { errors: undefined, id })
  return errors
}

function wrapRecord(r, id) {
  if (r.id) { id = r.id; r = r.r }
  return { r, id: id || (r.recordID !== NoID ? r.recordID : `${uniqueID.value}:${ctr++}`) }
}

function addInlineRecord() {
  const r = new compose.Record(recordListModule.value, {})
  if (options.value.refField) {
    const refField = recordListModule.value.fields.find(f => f.name === options.value.refField)
    if (refField?.isMulti) r.values[options.value.refField] = [(props.record || {}).recordID]
    else r.values[options.value.refField] = (props.record || {}).recordID
  }
  items.value.unshift(wrapRecord(r))
}

function resolveRecords(resolve) {
  ctr = 0
  items.value = items.value.map(wrapRecord)
  resolve({ items: items.value, module: recordListModule.value, refField: options.value.refField, positionField: options.value.positionField, idPrefix: uniqueID.value })
}

function validatePageBlock(resolve) {
  if (!options.value.editable) { resolve({ valid: true }); return }
  const req = new Set(recordListModule.value.fields.filter(({ isRequired }) => isRequired).map(({ name }) => name))
  for (const f of options.value.editFields) req.delete(f.name)
  resolve({ valid: !req.size })
  req.clear()
}

function handleDeleteInline(item, i) {
  if (item.r.recordID !== NoID) {
    const r = new compose.Record(recordListModule.value, { ...item.r, deletedAt: new Date() })
    items.value.splice(i, 1, wrapRecord(r, item.id))
  } else items.value.splice(i, 1)
}

function handleRestoreInline(item, i) {
  const r = new compose.Record(recordListModule.value, { ...item.r, deletedAt: undefined })
  items.value.splice(i, 1, wrapRecord(r, item.id))
}

function handleCloneInline(r) {
  r = new compose.Record(r.module, { ...r.values })
  items.value.splice(0, 0, wrapRecord(r))
}

function onInlineFieldChange(item) { dirtyInlineRecords.value[item.id] = true }

function showSaveAction(item) {
  if (!recordListModule.value) return false
  const r = item.r
  if (r.deletedAt) return true
  const { canCreateRecord } = recordListModule.value
  const { canUpdateRecord } = r
  if (r.recordID === NoID) return canCreateRecord
  if (!dirtyInlineRecords.value[item.id]) return false
  return canUpdateRecord
}

async function handleSaveInline(item, index) {
  if (!recordListModule.value) return
  processingInlineRecords.value[item.id] = 'save'
  const isNew = item.r.recordID === NoID
  let action = 'update'
  if (item.r.deletedAt) action = 'delete'
  else if (isNew) action = 'create'
  inlineErrors.value = inlineErrors.value.filter(e => e.meta.id !== item.id)
  try {
    if (item.r.deletedAt) {
      if (isNew) { items.value.splice(index, 1); return }
      await $ComposeAPI.recordDelete(item.r)
      items.value.splice(index, 1)
      const eUID = Object.keys(dirtyInlineRecords.value).length > 0 ? uniqueID.value : undefined
      window.dispatchEvent(new CustomEvent('module-records-updated', { detail: { moduleID: recordListModule.value.moduleID, excludeUniqueID: eUID } }))
      return
    }
    const v = new compose.RecordValidator(recordListModule.value)
    const fields = recordListModule.value.fields.filter(({ canReadRecordValue, canUpdateRecordValue }) => canReadRecordValue && canUpdateRecordValue).map(({ name }) => name)
    const err = v.run(item.r, ...fields)
    if (!err.valid()) {
      const fieldNames = new Set(err.set.map(e => recordListModule.value.fields.find(f => f.name === e.meta.field)?.label || e.meta.field))
      err.get().forEach(e => { e.meta.id = item.id })
      inlineErrors.value.push(...err.get())
      return
    }
    let saved
    if (isNew) saved = await $ComposeAPI.recordCreate(item.r)
    else saved = await $ComposeAPI.recordUpdate(item.r)
    delete dirtyInlineRecords.value[item.id]
    const newRecord = new compose.Record(recordListModule.value, saved)
    items.value.splice(index, 1, wrapRecord(newRecord))
    const eUID = Object.keys(dirtyInlineRecords.value).length > 0 ? uniqueID.value : undefined
    window.dispatchEvent(new CustomEvent('module-records-updated', { detail: { moduleID: recordListModule.value.moduleID, excludeUniqueID: eUID } }))
  } catch (e) {
    const { details } = e
    if (details && Array.isArray(details)) inlineErrors.value.push(...details.filter(d => !d.kind?.includes('warning')).map(d => ({ ...d, meta: { ...d.meta, id: item.id } })))
  } finally { delete processingInlineRecords.value[item.id] }
}

async function handleDenyInline(item, index) {
  if (!recordListModule.value) return
  processingInlineRecords.value[item.id] = 'deny'
  delete dirtyInlineRecords.value[item.id]
  inlineErrors.value = inlineErrors.value.filter(e => e.meta.id !== item.id)
  const isNew = item.r.recordID === NoID
  if (isNew) { items.value.splice(index, 1); delete processingInlineRecords.value[item.id] } else {
    try { const freshRecord = await $ComposeAPI.recordRead(item.r); items.value.splice(index, 1, wrapRecord(new compose.Record(recordListModule.value, freshRecord))) }
    finally { delete processingInlineRecords.value[item.id] }
  }
}

async function handleSaveDirtyRecords() {
  if (!recordListModule.value) return
  const itemsToSave = items.value.filter(item => {
    if (!showSaveAction(item)) return false
    if (selected.value.length > 0) return selected.value.includes(item.id)
    return true
  })
  if (!itemsToSave.length) return
  processingDirtyRecords.value = 'save'
  let hasError = false
  const updatableFields = recordListModule.value.fields.filter(({ canReadRecordValue, canUpdateRecordValue }) => canReadRecordValue && canUpdateRecordValue).map(({ name }) => name)
  for (const item of itemsToSave) {
    const isNew = item.r.recordID === NoID
    let action = 'update'
    if (item.r.deletedAt) action = 'delete'
    else if (isNew) action = 'create'
    try {
      const index = items.value.findIndex(i => i.id === item.id)
      if (index === -1) continue
      inlineErrors.value = inlineErrors.value.filter(e => e.meta.id !== item.id)
      if (item.r.deletedAt) {
        if (isNew) items.value.splice(index, 1)
        else { await $ComposeAPI.recordDelete(item.r); items.value.splice(index, 1) }
        delete dirtyInlineRecords.value[item.id]
        continue
      }
      const validator = new compose.RecordValidator(recordListModule.value)
      const err = validator.run(item.r, ...updatableFields)
      if (!err.valid()) {
        err.get().forEach(e => { e.meta.id = item.id })
        inlineErrors.value.push(...err.get())
        throw new Error($t('notification.record.validationErrors', { fields: [...new Set(err.set.map(e => recordListModule.value.fields.find(f => f.name === e.meta.field)?.label || e.meta.field))].join(', ') }))
      }
      let saved
      if (isNew) saved = await $ComposeAPI.recordCreate(item.r)
      else saved = await $ComposeAPI.recordUpdate(item.r)
      delete dirtyInlineRecords.value[item.id]
      items.value.splice(index, 1, wrapRecord(new compose.Record(recordListModule.value, saved)))
    } catch (e) {
      hasError = true
      const { details } = e
      if (details && Array.isArray(details)) inlineErrors.value.push(...details.filter(d => !d.kind?.includes('warning')).map(d => ({ ...d, meta: { ...d.meta, id: item.id } })))
    }
  }
  processingDirtyRecords.value = ''
  if (!hasError) { selected.value = []; const eUID = Object.keys(dirtyInlineRecords.value).length > 0 ? uniqueID.value : undefined; window.dispatchEvent(new CustomEvent('module-records-updated', { detail: { moduleID: recordListModule.value.moduleID, excludeUniqueID: eUID } })) }
}

async function handleDenyDirtyRecords() {
  if (!recordListModule.value) return
  const itemsToDeny = items.value.filter(item => {
    if (!showSaveAction(item)) return false
    if (selected.value.length > 0) return selected.value.includes(item.id)
    return true
  })
  if (!itemsToDeny.length) return
  processingDirtyRecords.value = 'deny'
  let hasError = false
  for (const item of itemsToDeny) {
    try {
      const index = items.value.findIndex(i => i.id === item.id)
      if (index === -1) continue
      const isNew = item.r.recordID === NoID
      delete dirtyInlineRecords.value[item.id]
      inlineErrors.value = inlineErrors.value.filter(e => e.meta.id !== item.id)
      if (isNew) items.value.splice(index, 1)
      else { const freshRecord = await $ComposeAPI.recordRead(item.r); items.value.splice(index, 1, wrapRecord(new compose.Record(recordListModule.value, freshRecord))) }
    } catch (e) { hasError = true }
  }
  processingDirtyRecords.value = ''
  if (!hasError) selected.value = []
}

function moduleOmitsSystemField (mod, name) {
  if (typeof mod?.systemFieldOmitted === 'function') return mod.systemFieldOmitted(name)
  const enc = mod?.config?.dal?.systemFieldEncoding?.[name]
  return !!(enc && typeof enc === 'object' && enc.omit)
}

function sanitizeRecordListSort (presort, mod) {
  let sort = (presort || '').trim()
  if (!sort) {
    sort = moduleOmitsSystemField(mod, 'createdAt') ? '' : 'createdAt DESC'
  }
  const omitted = ['createdAt', 'updatedAt', 'deletedAt'].filter(f => moduleOmitsSystemField(mod, f))
  if (!omitted.length || !sort) return sort
  return sort.split(',').map(s => s.trim()).filter(part => {
    const col = part.split(/\s+/)[0]
    return col && !omitted.includes(col)
  }).join(', ')
}

function prepRecordList() {
  const { moduleID, presort, prefilter: pf, editable, refField, positionField, perPage } = options.value
  if (!moduleID || !recordListModule.value) throw new Error($t('record.moduleOrPageNotSet'))
  recordsPerPage.value = perPage
  if (isOnRecordPage.value && options.value.linkToParent) {
    options.value.linkToParent = false
    if (!options.value.refField) options.value.refField = (recordListModule.value.fields.find(f => f.kind === 'Record' && f.options.moduleID === props.page.moduleID) || {}).name
  }
  if (!props.record) {
    if ((pf || '').includes('${record')) throw new Error($t('record.invalidRecordVar'))
    if ((pf || '').includes('${ownerID}')) throw new Error($t('record.invalidOwnerVar'))
  }
  let sort = sanitizeRecordListSort(presort, recordListModule.value)
  if (editable && positionField) sort = positionField
  const filterArr = []
  if (pf) {
    const { skip, filter } = evalPrefilterOrSkip(pf, {
      record: props.record, user: $auth?.user || {},
      recordID: (props.record || {}).recordID || NoID,
      ownerID: (props.record || {}).ownedBy || NoID,
      userID: ($auth?.user || {}).userID || NoID,
      loadingRecord: !!props.loadingRecord,
      variables: pageVariables.value,
    })
    filterArr.push(`(${skip ? 'false' : filter})`)
  }
  if (refField) {
    if (!props.record) throw new Error($t('record.invalidRecordVar'))
    const refFieldObj = recordListModule.value.fields.find(f => f.name === refField)
    filterArr.push(getFieldFilter(refField, 'Record', props.record.recordID, refFieldObj?.isMulti ? 'IN' : '='))
  }
  prefilter.value = filterArr.join(' AND ')
  filter.limit = recordsPerPage.value
  filter.sort = sort
}

function createReminder(record) {
  const { recordID, values = {} } = record
  const { name, isMulti } = (options.value.fields || []).find(({ name }) => !!values[name]) || {}
  const title = isMulti ? values[name].join(', ') : values[name]
  const resource = `compose:record:${recordID}`
  const payload = { title, link: { name: 'page.record', label: 'Record page', params: { slug: props.namespace.slug || props.namespace.namespaceID, pageID: recordPageID.value, recordID } } }
  window.dispatchEvent(new CustomEvent('reminder.create', { detail: { payload, resource } }))
  window.dispatchEvent(new CustomEvent('rightPanel.toggle', { detail: true }))
}

function onExport(e) {
  processing.value = true
  const { namespaceID, moduleID } = filter
  const { filter: expFilter, filterRaw, timezone, resolveRefs } = e
  let download = { ...e, namespaceID, moduleID, filename: `${props.namespace.slug || namespaceID} - ${recordListModule.value.name}` }
  if (filterRaw.rangeType === 'range') download.filename += ` - ${filterRaw.date.start} - ${filterRaw.date.end}`
  else download.filename += ` - ${filterRaw.rangeType}`
  if (timezone) download.filename += ` - ${timezone.label}`
  download.filename = encodeURIComponent(download.filename.replace(/\./g, '-'))
  const exportUrl = url.Make({
    url: `${$ComposeAPI.baseURL}${$ComposeAPI.recordExportEndpoint(download)}`,
    query: { fields: e.fields, multiValueDelimiter: e.multiValueDelimiter, filter: selectedAllRecords.value ? bulkQuery.value : expFilter, jwt: $auth.accessToken, timezone: timezone?.tzCode, resolveRefs },
  })
  window.open(exportUrl)
  processing.value = false
}

function recordFieldRaw (record, name) {
  if (!record || !name) return undefined
  if (Object.prototype.hasOwnProperty.call(record, name) && record[name] !== undefined && !(record.values && Object.prototype.hasOwnProperty.call(record.values, name))) {
    // system-ish fields occasionally live on the record root
  }
  if (record.values && Object.prototype.hasOwnProperty.call(record.values, name)) {
    const v = record.values[name]
    return Array.isArray(v) ? v[0] : v
  }
  const v = record[name]
  return Array.isArray(v) ? v[0] : v
}

function severityTone (val) {
  if (val != null && val !== '' && !Number.isNaN(Number(val)) && String(val).trim() !== '' && !/[a-z]/i.test(String(val))) {
    const n = Number(val)
    if (n >= 60) return 'danger'
    if (n >= 30) return 'warning'
    if (n > 0) return 'info'
    return 'success'
  }
  const s = String(val ?? '').trim().toLowerCase().replace(/_/g, ' ')
  if (['critical', 'open', 'escalated'].includes(s)) return 'danger'
  if (['understock', 'high', 'in progress', 'warning'].includes(s)) return 'warning'
  if (['ok', 'resolved', 'closed', 'low', 'success', 'norma', 'норма'].includes(s)) return 'success'
  if (['overstock', 'info'].includes(s)) return 'info'
  return ''
}

function rowHighlightValue (element) {
  const field = options.value.rowHighlightField || options.value.signalField
  if (!field) return undefined
  return recordFieldRaw(element.r, field)
}

function rowClass (element) {
  const cls = {}
  if (isRowClickable.value) cls.pointer = true
  if (inlineEditing.value && (dirtyInlineRecords.value[element.id] || element.r.deletedAt)) cls['table-warning'] = true
  const tone = severityTone(rowHighlightValue(element))
  if (tone) cls[`rl-row-${tone}`] = true
  return cls
}

function signalClass (element) {
  const field = options.value.signalField || options.value.rowHighlightField
  const tone = severityTone(field ? recordFieldRaw(element.r, field) : rowHighlightValue(element))
  return tone ? `rl-signal-${tone}` : 'rl-signal-muted'
}

function signalTitle (element) {
  const field = options.value.signalField || options.value.rowHighlightField
  if (!field) return ''
  return String(recordFieldRaw(element.r, field) ?? '')
}

function fieldCellClass (field) {
  if (!options.value.alignNumbers) return undefined
  const kind = field.moduleField?.kind
  if (kind === 'Number') return 'text-end rl-num'
  return undefined
}

function isSparklineField (field) {
  return !!options.value.sparklineField && field.key === options.value.sparklineField
}

function sparklinePct (element) {
  const field = options.value.sparklineField
  if (!field) return 0
  const v = Number(recordFieldRaw(element.r, field))
  const max = Number(options.value.sparklineMax) || 30
  if (!Number.isFinite(v) || max <= 0) return 0
  return Math.max(0, Math.min(100, (v / max) * 100))
}

function cardTitle (element) {
  const preferred = ['slice_label', 'product_name', 'body', 'name', 'title']
  for (const name of preferred) {
    const hit = fields.value.find(f => f.key === name)
    if (!hit) continue
    const v = recordFieldRaw(element.r, name)
    if (v != null && v !== '') return String(v)
  }
  const first = fields.value[0]
  if (!first) return element.r.recordID
  const v = recordFieldRaw(element.r, first.key)
  return v != null && v !== '' ? String(v) : element.r.recordID
}

const rowTooltip = reactive({
  visible: false,
  title: '',
  record: null,
  lines: [],
  style: {},
})

const TOOLTIP_MARKER_COLORS = [
  '#5470c6', '#91cc75', '#fac858', '#ee6666', '#73c0de',
  '#3ba272', '#fc8452', '#9a60b4', '#ea7ccc',
]

function tooltipMarkerColor (value, index) {
  const tone = severityTone(value)
  if (tone === 'danger') return '#ee6666'
  if (tone === 'warning') return '#fac858'
  if (tone === 'success') return '#91cc75'
  if (tone === 'info') return '#73c0de'
  return TOOLTIP_MARKER_COLORS[index % TOOLTIP_MARKER_COLORS.length]
}

function buildRowTooltipLines (element) {
  return fields.value.map((f, index) => ({
    label: f.label,
    moduleField: f.moduleField,
    color: tooltipMarkerColor(recordFieldRaw(element.r, f.key), index),
  }))
}

function positionRowTooltip (clientX, clientY) {
  const pad = 12
  const width = 300
  const approxH = Math.min(360, 36 + rowTooltip.lines.length * 24)
  let left = clientX + 16
  let top = clientY + 16
  if (left + width + pad > window.innerWidth) left = Math.max(pad, clientX - width - 16)
  if (top + approxH + pad > window.innerHeight) top = Math.max(pad, clientY - approxH - 16)
  rowTooltip.style = {
    top: `${top}px`,
    left: `${left}px`,
    minWidth: `${Math.min(width, 240)}px`,
    maxWidth: '420px',
  }
}

function showRowValueTooltip (event, element) {
  if (options.value.rowValueTooltip === false) return
  if (inlineEditing.value) return
  const tag = (event.target?.tagName || '').toLowerCase()
  if (['input', 'button', 'a', 'select', 'textarea'].includes(tag)) return
  if (event.target?.closest?.('.actions, .inline-actions, .dropdown, .rl-check-col')) return

  const lines = buildRowTooltipLines(element)
  if (!lines.length) return
  rowTooltip.title = cardTitle(element)
  rowTooltip.record = element.r
  rowTooltip.lines = lines
  rowTooltip.visible = true
  positionRowTooltip(event.clientX, event.clientY)
}

function moveRowValueTooltip (event) {
  if (!rowTooltip.visible) return
  positionRowTooltip(event.clientX, event.clientY)
}

function hideRowValueTooltip () {
  rowTooltip.visible = false
  rowTooltip.title = ''
  rowTooltip.record = null
  rowTooltip.lines = []
}

function handleRowClick(item) {
  let { recordID } = item.r
  const { values = {} } = item.r
  if (options.value.recordDisplayListField) recordID = options.value.recordDisplayTargetModuleField + '=' + values[options.value.recordDisplayListField]
  if (options.value.recordDisplayOption === 'doNothing') return
  const pageID = recordPageID.value
  if (inlineEditing.value || (!pageID && !options.value.rowViewUrl)) return
  if (options.value.enableRecordPageNavigation) store.ui.loadPaginationRecords({ filter: { ...filter, limit: 50 } })
  if (options.value.recordDisplayOption === 'modal' || props.mode === 'modal') {
    window.dispatchEvent(new CustomEvent('show-record-modal', { detail: { recordID, recordPageID: pageID, edit: options.value.openRecordInEditMode } }))
    return
  }
  const name = options.value.openRecordInEditMode ? options.value.rowEditUrl || 'page.record.edit' : options.value.rowViewUrl || 'page.record'
  const route = { name, params: { pageID, recordID }, query: null }
  if (options.value.recordDisplayOption === 'newTab') window.open($router.resolve(route).href)
  else $router.push(route)
}

function handleSort({ key, sortable }) {
  if (!sortable) return
  if (sortBy.value !== key) { filter.sort = `${key}`; sortDirection.value = 'ASC' }
  else { filter.sort = sortDirection.value === 'ASC' ? `${key} DESC` : `${key}`; sortDirection.value = sortDirection.value === 'ASC' ? 'DESC' : 'ASC' }
  sortBy.value = key
  refresh(true)
}

async function goToPage(page) {
  if (page >= 1) {
    const idx = page - 1
    filter.pageCursor = (pagination.pages[idx] || {}).cursor || ''
    pagination.page = page
  }
  return refresh()
}

function handleSelectAllOnPage({ isChecked }) {
  selected.value = isChecked ? items.value.map(({ id }) => id) : []
  if (!isChecked) selectedAllRecords.value = false
}

function selectAllRecords() { selectedAllRecords.value = !selectedAllRecords.value; handleSelectAllOnPage({ isChecked: selectedAllRecords.value }) }

function handleRestoreSelectedRecords(recordID) {
  if (inlineEditing.value && editing.value) {
    const sel = new Set(selected.value)
    items.value.forEach((item, index) => { if (sel.has(item.id)) handleRestoreInline(item, index) })
    sel.clear()
  } else {
    processing.value = true
    const query = recordID ? `recordID = ${recordID}` : bulkQuery.value
    const { moduleID, namespaceID } = filter
    $ComposeAPI.recordBulkUndelete({ moduleID, namespaceID, query }).then(() => { refresh(true) }).finally(() => { setTimeout(() => { processing.value = false; selectedAllRecords.value = false }, 300) })
  }
}

function handleDeleteSelectedRecords(recordID) {
  if (inlineEditing.value && editing.value) {
    const sel = new Set(selected.value)
    for (let i = 0; i < items.value.length; i++) { if (sel.has(items.value[i].id)) handleDeleteInline(items.value[i], i) }
    sel.clear()
  } else {
    processing.value = true
    const query = recordID ? `recordID = ${recordID}` : bulkQuery.value
    const { moduleID, namespaceID } = filter
    $ComposeAPI.recordBulkDelete({ moduleID, namespaceID, query }).then(() => refresh(true)).finally(() => { setTimeout(() => { processing.value = false; selectedAllRecords.value = false }, 300) })
  }
}

async function refresh(resetPagination = false, checkSelected = false) {
  if (checkSelected && (selected.value.length || inlineEdit.recordIDs.length)) return
  processing.value = true
  await nextTick()
  try {
    return await pullRecords(resetPagination)
  } catch (e) {
    if (recordListModule.value) {
      hasLoadedOnce.value = true
    }
    processing.value = false
  }
}

async function pullRecords(resetPagination = false) {
  if (!recordListModule.value || recordListModule.value.moduleID !== options.value.moduleID) throw new Error($t('record.moduleMismatch'))
  abortRequests()
  processing.value = true
  selected.value = []
  let searchFields = []
  if (options.value.searchableFields.length > 0) searchFields = recordListModule.value.filterFields(options.value.searchableFields)
  else searchFields = fields.value.map(({ moduleField }) => moduleField)
  const queryStr = queryToFilter(query.value, prefilter.value, searchFields, groupRecordListFilter.value)
  const { moduleID, namespaceID } = recordListModule.value
  let paginationOptions = {}
  let summariesStr = []
  if (resetPagination) {
    filter.pageCursor = undefined
    const { fullPageNavigation = false, showTotalCount = false } = options.value
    paginationOptions = { incPageNavigation: fullPageNavigation, incTotal: showTotalCount }
    summariesStr = JSON.stringify(listSummaries.value.map(s => ({ name: s.metric, field: s.field })))
  } else if (filter.pageCursor) filter.sort = ''
  showingDeletedRecords.value ? filter.deleted = 2 : filter.deleted = 0
  const { response, cancel } = $ComposeAPI.recordListCancellable({ ...filter, moduleID, namespaceID, query: queryStr, ...paginationOptions, summaries: summariesStr })
  abortableRequests.value.push(cancel)
  return Promise.all([response(), new Promise(resolve => setTimeout(resolve, 300))])
    .then(([{ set, filter: respFilter, summaries: summs = {} }]) => {
      const records = set.map(r => new compose.Record(r, recordListModule.value))
      store.record.updateRecords(records)
      Object.assign(filter, respFilter)
      filter.nextPage = respFilter.nextPage
      filter.prevPage = respFilter.prevPage
      if (resetPagination) {
        summaries.value = summs
        let count = pagination.count || 0
        if (paginationOptions.incTotal) { count = respFilter.total || 0; filter.incTotal = false }
        if (paginationOptions.incPageNavigation) {
          const pages = respFilter.pageNavigation || []
          pagination.pages = pages
          if (!paginationOptions.incTotal) count = pages.length > 1 ? ((pages.length - 1) * recordsPerPage.value) + (pages[pages.length - 1]?.items || 0) : records.length
          filter.incPageNavigation = false
        }
        pagination.count = count
        pagination.page = 1
      }
      if (stayOnPage.value) { const goToPageNumber = stayOnPage.value; stayOnPage.value = undefined; return goToPage(goToPageNumber) }
      const flds = fields.value.filter(f => f.moduleField).map(f => f.moduleField)
      return Promise.all([
        fetchUsers(flds, records),
        fetchRecords(namespaceID, flds, records),
      ]).then(() => {
        dirtyInlineRecords.value = {}
        inlineErrors.value = new validator.Validated()
        processingInlineRecords.value = {}
        items.value = records.map(r => wrapRecord(r))
        hasLoadedOnce.value = true
        processing.value = false
      })
    }).catch(e => {
      if (axios.isCancel(e)) return
      hasLoadedOnce.value = true
      processing.value = false
    }).finally(() => { cancelled.value = false })
}

function getStorageRecordListFilter() {
  try {
    const currentFilters = getItem(`record-list-filters-${uniqueID.value}`)
    if (!Array.isArray(currentFilters)) { console.warn($t('notification.record-list.incorrect-filter-structure', { filterID: uniqueID.value })); removeItem(`record-list-filters-${uniqueID.value}`) }
    else recordListFilter.value = currentFilters
  } catch (e) { console.warn($t('notification.record-list.corrupted-filter')); removeItem(`record-list-filters-${uniqueID.value}`) }
}

function getCustomSummaries() {
  try { customSummaries.value = getItem(`record-list-custom-summaries-${uniqueID.value}`) } catch (e) { console.warn($t('notification.record-list.corrupted-summaries')) }
}

function getStorageRecordListFilterPreset() {
  try { customPresetFilters.value = getItem(`record-list-preset-${uniqueID.value}`) || [] } catch (e) { console.warn($t('notification.record-list.corrupted-filter')); removeItem(`record-list-filters-${uniqueID.value}`) }
}

function setStorageRecordListFilter() { try { setItem(`record-list-filters-${uniqueID.value}`, recordListFilter.value) } catch (e) { console.warn($t('notification.record-list.corrupted-filter')) } }

function setStorageCustomSummaries() { try { setItem(`record-list-custom-summaries-${uniqueID.value}`, customSummaries.value) } catch (e) { console.warn($t('notification.record-list.corrupted-summaries')) } }

function setStorageRecordListFilterPreset({ name } = {}) {
  showCustomPresetFilterModal.value = false
  const currentListFilters = [...customPresetFilters.value]
  if (name) currentListFilters.push({ ...currentCustomPresetFilter.value, name })
  customPresetFilters.value = currentListFilters
  try { setItem(`record-list-preset-${uniqueID.value}`, currentListFilters) } catch (e) { console.warn($t('notification.record-list.corrupted-filter')) }
}

function removeRecordListFilterPreset(name) {
  customPresetFilters.value = customPresetFilters.value.filter(f => f.name !== name)
  setStorageRecordListFilterPreset()
}

function onImportSuccessful() { window.dispatchEvent(new CustomEvent('module-records-updated', { detail: { moduleID: recordListModule.value?.moduleID } })) }

function createDefaultFilter(field = {}, value = undefined, operator = undefined) {
  if (!field.resourceID) field = allFields.value.find(({ name }) => name === field.name) || field
  if (field) { field = new compose.ModuleFieldMaker(field); field.isMulti = false }
  let record = new compose.Record(recordListModule.value)
  if (isBetweenOperator(operator)) {
    record = [new compose.Record(recordListModule.value), new compose.Record(recordListModule.value)]
    if (field.isSystem) { record[0][field.name] = value.start; record[1][field.name] = value.end }
    else { record[0].values[field.name] = value.start; record[1].values[field.name] = value.end }
  } else {
    if (field.isSystem) record[field.name] = value
    else record.values[field.name] = value
  }
  return { name: field.name, operator: operator || (field.isMulti ? 'IN' : '='), value, kind: field.kind, label: field.label || field.name, field, record }
}

function setDrillDownFilter({ detail: { prefilter: drillDownFilter, name, value: fieldValue } } = {}) {
  if (!drillDownFilter) return
  let rlf = recordListFilter.value
  if (!rlf.length) rlf = [{ filter: [createDefaultFilter({ name }, fieldValue, '=')] }]
  else { const { filter } = rlf[0]; if (!filter.length || (filter.length && !filter[0].name)) { rlf[0].filter = []; rlf[0].filter.push(createDefaultFilter({ name }, fieldValue)) } else rlf[0].filter.push(createDefaultFilter({ name }, fieldValue)) }
  onFilter(rlf)
}

function isInlineRestoreActionVisible({ deletedAt }) { return !options.value.hideRecordDeleteButton && !!deletedAt }
function isInlineDeleteActionVisible({ recordID, canDeleteRecord, deletedAt }) { return !options.value.hideRecordDeleteButton && !deletedAt && (canDeleteRecord || recordID === NoID) }
function isViewRecordActionVisible({ canReadRecord }) { return !options.value.hideRecordViewButton && canReadRecord && (options.value.rowViewUrl || recordPageID.value) }
function isEditRecordActionVisible({ canUpdateRecord }) { return !options.value.hideRecordEditButton && canUpdateRecord && (options.value.rowEditUrl || recordPageID.value) }
function isRecordPermissionButtonVisible({ canGrant }) { return canGrant && !options.value.hideRecordPermissionsButton }
function isDeleteActionVisible({ deletedAt, canDeleteRecord }) { return !options.value.hideRecordDeleteButton && !deletedAt && canDeleteRecord }
function isRestoreActionVisible({ canUndeleteRecord }) { return !options.value.hideRecordDeleteButton && canUndeleteRecord }

function areActionsVisible(record) {
  if (inlineEditing.value && editing.value) return [isCloneRecordActionVisible.value, isInlineRestoreActionVisible(record), isInlineDeleteActionVisible(record)].some(v => v)
  return [isCloneRecordActionVisible.value, isReminderActionVisible.value, isViewRecordActionVisible(record), isEditRecordActionVisible(record), isRecordPermissionButtonVisible(record), isDeleteActionVisible(record), isRestoreActionVisible(record)].some(v => v)
}

function onBulkUpdate() { selectedAllRecords.value = false }

function editInlineField(record, field) {
  inlineEdit.fields = [field]
  inlineEdit.record = record.clone()
  inlineEdit.query = `recordID = ${record.recordID}`
}

function filterByValue(record, { moduleField: field }) {
  const value = field.isSystem ? record[field.name] : record.values[field.name]
  const operator = field.isMulti ? 'IN' : '='
  const setFilter = (fld, val) => {
    if (!recordListFilter.value.length) recordListFilter.value = [{ filter: [createDefaultFilter(fld, val, fld.isMulti ? 'IN' : operator)] }]
    else { const { filter } = recordListFilter.value[0]; if (!filter.length || (filter.length && !filter[0].name)) { recordListFilter.value[0].filter = []; recordListFilter.value[0].filter.push(createDefaultFilter(fld, val, operator)) } else if (!recordListFilter.value[0].filter.some(f => f.name === fld.name && f.value === val)) recordListFilter.value[0].filter.push(createDefaultFilter(fld, val, operator)) }
  }
  if (field.isMulti) value.forEach(v => setFilter(field, v))
  else setFilter(field, value)
  pullRecords(true)
}

function showInlineActions(field) { return showInlineEdit(field) || showInlineFilter(field) }

function showInlineEdit(field) {
  const isfieldInlineEditable = () => {
    if (Array.isArray(options.value.inlineEditFields) && options.value.inlineEditFields.length === 0) return true
    return options.value.inlineEditFields.some(fieldID => fieldID === field.moduleField.fieldID || fieldID === field.moduleField.name)
  }
  return options.value.inlineRecordEditEnabled && field.canEdit && !showingDeletedRecords.value && isfieldInlineEditable()
}

function showInlineFilter(field) { return options.value.inlineValueFiltering && !options.value.hideFiltering && field.filterable }
function onInlineEditClose() { inlineEdit.fields = []; inlineEdit.record = {}; inlineEdit.query = '' }
function onInlineEdit() { onInlineEditClose() }

function isFieldEditable(field) {
  if (!field) return false
  const { canCreateOwnedRecord } = recordListModule.value || {}
  const { createdAt, canManageOwnerOnRecord } = props.record || {}
  const { name, canUpdateRecordValue, isSystem, expressions = {} } = field
  if (!canUpdateRecordValue) return false
  if (isSystem) return name === 'ownedBy' ? (createdAt ? canManageOwnerOnRecord : canCreateOwnedRecord) : false
  return !expressions.value
}

function updateFilter(filterArr = [], name) {
  filterArr = filterArr.map(f => ({ ...f, name }))
  recordListFilter.value = recordListFilter.value.concat(filterArr)
  refresh(true)
}

function removeFilter(groupIndex, filterIndex) {
  if (recordListFilter.value[groupIndex]) {
    recordListFilter.value[groupIndex].filter = (recordListFilter.value[groupIndex].filter || []).filter((_, index) => index !== filterIndex)
    if (!recordListFilter.value[groupIndex].filter.length) {
      if (groupIndex === 0 && recordListFilter.value[1]) recordListFilter.value[1].groupCondition = 'OR'
      recordListFilter.value.splice(groupIndex, 1)
    }
  }
  const hasAnyFilters = recordListFilter.value.some(group => group.filter && group.filter.length > 0)
  if (!hasAnyFilters) { onFilter(); return }
  setStorageRecordListFilter()
  refresh(true)
}

function isUserRoleMember(roles) { return !roles.length || roles.some(roleID => authUserRoles.value.includes(roleID)) }

function openCustomSummaryModal(summary) {
  const { custom, metric, field } = summary || {}
  if (summary && !custom) return
  customSummaryIndex.value = customSummaries.value.findIndex(s => s.field === field && s.metric === metric)
  customSummary.value = customSummaryIndex.value === -1 ? { custom: true, label: '', field: '', metric: '' } : { ...customSummaries.value[customSummaryIndex.value] }
  showCustomSummariesModal.value = true
}

function onCustomSummarySave(summary) {
  if (customSummaryIndex.value === -1) customSummaries.value.push(summary)
  else customSummaries.value[customSummaryIndex.value] = summary
  onCustomSummaryClose()
  setStorageCustomSummaries()
  pullRecords(true)
}

function onCustomSummaryDelete() { customSummaries.value.splice(customSummaryIndex.value, 1); onCustomSummaryClose(); setStorageCustomSummaries(); pullRecords(true) }
function onCustomSummaryClose() { customSummaryIndex.value = -1; customSummary.value = {}; showCustomSummariesModal.value = false }

function setDefaultValues() {
  uniqueID.value = undefined; processing.value = false; hasLoadedOnce.value = false; prefilter.value = undefined; recordListFilter.value = []; query.value = null
  Object.assign(filter, { query: '', sort: '', limit: 10, pageCursor: '', prevPage: '', nextPage: '' })
  Object.assign(pagination, { pages: [], page: 1, count: 0 })
  selected.value = []; inlineEdit.recordIDs = []; inlineEdit.fields = []; inlineEdit.initialRecord = {}
  sortBy.value = undefined; sortDirection.value = undefined; ctr = 0; items.value = []
  showingDeletedRecords.value = false; customPresetFilters.value = []; currentCustomPresetFilter.value = undefined
  showCustomPresetFilterModal.value = false; selectedAllRecords.value = false; abortableRequests.value = []
  summaries.value = []; customSummaries.value = []; customSummaryIndex.value = -1; customSummary.value = {}
  showCustomSummariesModal.value = false; cancelled.value = false; stayOnPage.value = undefined
  if (processingTimeout) { clearTimeout(processingTimeout); processingTimeout = undefined }
}

function abortRequests() {
  if (processingTimeout) clearTimeout(processingTimeout)
  abortableRequests.value.forEach(cancel => cancel())
}

function refreshAndResetPagination({ detail: { stayOnPage: sop = true } = {} } = {}) {
  if (sop) stayOnPage.value = pagination.page
  refresh(true)
}

function destroyEvents() {
  window.removeEventListener(`record-line:collect:${uniqueID.value}`, resolveRecords)
  window.removeEventListener(`page-block:validate:${uniqueID.value}`, validatePageBlock)
  window.removeEventListener(`drill-down-recordList:${uniqueID.value}`, setDrillDownFilter)
  window.removeEventListener('module-records-updated', refreshOnRelatedRecordsUpdate)
  window.removeEventListener('record-field-change', refetchOnPrefilterValueChange)
  window.removeEventListener('page-variable-change', refetchOnPageVariableChange)
  window.removeEventListener('refetch-records', refreshAndResetPagination)
  if (processingTimeout) clearTimeout(processingTimeout)
}

function handleAddRecord() {
  const refRecord = options.value.refField && props.record?.recordID !== NoID ? props.record : undefined
  const pageID = recordPageID.value
  if (!(pageID || options.value.rowCreateUrl)) return
  const route = { name: options.value.rowCreateUrl || 'page.record.create', params: { pageID, refRecord }, query: null, edit: true }
  if (props.mode === 'modal' || options.value.addRecordDisplayOption === 'modal') {
    window.dispatchEvent(new CustomEvent('show-record-modal', { detail: { recordID: NoID, recordPageID: recordPageID.value, refRecord, edit: true } }))
  } else if (options.value.addRecordDisplayOption === 'newTab') { window.open($router.resolve(route).href) }
  else { $router.push(route) }
}

function viewRecordRoute(recordID) {
  if (props.mode === 'modal') return { name: $route.name, params: $route.params, query: { ...$route.query, recordPageID: recordPageID.value, recordID }, edit: false }
  return { name: options.value.rowViewUrl || 'page.record', params: { pageID: recordPageID.value, recordID }, query: null, edit: false }
}

function editRecordRoute(recordID) {
  if (props.mode === 'modal') return { name: $route.name, params: $route.params, query: { ...$route.query, recordPageID: recordPageID.value, recordID }, edit: true }
  return { name: options.value.rowEditUrl || 'page.record.edit', params: { pageID: recordPageID.value, recordID }, query: null, edit: true }
}

function handleCloneRecordAction(recordID, values) {
  if (props.mode === 'modal') { window.dispatchEvent(new CustomEvent('show-record-modal', { detail: { recordID, recordPageID: recordPageID.value, values, edit: true } })); return }
  $router.push({ name: options.value.rowCreateUrl || 'page.record.create', params: { pageID: recordPageID.value, values }, query: null, edit: true })
}

function toCSV(rows, headers) {
  const esc = (v) => { const s = v === null || v === undefined ? '' : String(v); return '"' + s.replace(/"/g, '""') + '"' }
  const head = headers.map(h => esc(h.label)).join(',')
  const body = rows.map(r => headers.map(h => esc(r[h.key])).join(',')).join('\r\n')
  return head + '\r\n' + body + '\r\n'
}

function promptAiChat() {
  const page = props.page; const block = props.block; const namespace = props.namespace
  let prompt = block.prompt || page.config?.prompt || namespace.prompt || ''
  if (!prompt.length) { const locale = navigator.language || 'en-US'; prompt = locale.startsWith('ru') ? 'Что показывает этот список записей? О чём он говорит?' : 'What does this record list show?' }
  prompt += '\r\n*' + page.title + '*\r\n*' + block.title + '*\r\n'
  if (recordListModule.value) prompt += 'Module: ' + recordListModule.value.name + '\r\n'
  if (query.value) prompt += 'Search: ' + query.value + '\r\n'
  if (prefilter.value) prompt += 'Filter: ' + prefilter.value + '\r\n'
  const files = []
  if (items.value.length) {
    const headers = fields.value.map(f => ({ key: f.key, label: f.label }))
    const firstRows = items.value.slice(0, 25)
    const csv = toCSV(firstRows.map(item => {
      const row = {}; for (const h of headers) row[h.key] = item.r.values[h.key]; return row
    }), headers)
    files.push({ name: 'record_list_data.csv', content: csv, type: 'text/csv' })
  }
  window.dispatchEvent(new CustomEvent('show-chat-modal', { detail: { namespace: page.namespaceID, module: page.moduleID, prompt, files } }))
}

function fetchUsers(fields, records) {
  const list = [...new Set(records.map(r => fields.filter(c => c.kind === 'User').map(f => f.isSystem ? [r[f.name]] : f.isMulti ? r.values[f.name] : [r.values[f.name]])).flat(Infinity))].filter(uID => uID !== NoID)
  if (list.length) return store.user.resolveUsers(list)
}

function fetchRecords(namespaceID, fields, records) {
  const moduleRecords = {}
  fields.filter(c => c.kind === 'Record').forEach(f => {
    const { moduleID: fmid } = f.options || {}
    if (!fmid) return
    if (!moduleRecords[fmid]) moduleRecords[fmid] = new Set()
    records.forEach(r => { const ids = f.isSystem ? [r[f.name]] : f.isMulti ? r.values[f.name] : [r.values[f.name]]; ids.forEach(id => { if (id) moduleRecords[fmid].add(id) }) })
  })
  return Promise.all(Object.entries(moduleRecords).map(([fmid, ids]) => {
    ids = [...ids]
    return ids.length ? store.record.resolveRecords({ namespaceID, moduleID: fmid, recordIDs: ids }) : Promise.resolve([])
  }))
}
</script>
<style lang="scss" scoped>
.brain-button {     z-index: 1041; top: 0.5rem; right: 0.5rem; width: 32px; height: 32px; background: #fff; border-radius: 50%; box-shadow: 0 1px 4px rgba(0,0,0,0.15); border: 1px solid var(--bs-border-color, #dee2e6); color: var(--secondary); font-size: 16px; cursor: pointer; transition: box-shadow 0.2s, color 0.2s; }
.brain-button:hover { opacity: 1; box-shadow: 0 2px 8px rgba(0,0,0,0.2); }
.handle { cursor: grab; }
.pointer { cursor: pointer; }
th.required::after { content: "*"; display: inline-block; color: var(--primary); vertical-align: sub; margin-left: 2px; width: 10px; height: 16px; overflow: hidden; }
tr:hover td.actions { opacity: 1; &:not(.actions-visible) { background-color: var(--light); } }
.inline-actions { margin-top: -2px; opacity: 0; transition: opacity 0.25s; }
tr:hover .inline-actions { opacity: 1; button:hover { color: var(--primary) !important; } }
.custom-summary { cursor: pointer !important; border-radius: 0.25rem; > label { cursor: pointer !important; } &:hover { background-color: var(--extra-light); } }

.rl-root {
  width: 100%;
  min-height: 0;
}

.rl-display-table .rl-cards-wrap,
.rl-display-cards .rl-table-wrap {
  display: none !important;
}

.rl-display-responsive {
  .rl-cards-wrap { display: none !important; }
  @media (max-width: 767.98px) {
    .rl-table-wrap { display: none !important; }
    .rl-cards-wrap { display: block !important; }
  }
}

.table-responsive {
  border-radius: 0.5rem;
  border: 1px solid var(--bs-border-color, #dee2e6);
  width: 100%;
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

  td.actions {
    padding-top: 8px;
    right: 0;
    opacity: 0;
    position: sticky;
    transition: opacity 0.2s;
    width: 1%;
    background: linear-gradient(90deg, rgba(255,255,255,0) 0%, var(--bs-white, #fff) 20%);

    &.actions-visible {
      opacity: 1;
    }
  }

  tr:hover td.actions {
    opacity: 1;
    z-index: 1;
  }
}

.rl-compact .record-list-table {
  thead th, tbody td {
    padding-top: 0.4rem;
    padding-bottom: 0.4rem;
    font-size: 0.8125rem;
  }
}

.rl-align-numbers .rl-num {
  font-variant-numeric: tabular-nums;
}

.rl-check-col {
  vertical-align: middle !important;
}

.rl-row-check.form-check-input {
  margin-top: 0;
  vertical-align: middle;
  position: static;
  float: none;

  &.mt-2 {
    margin-top: 0.5rem;
  }
}

.rl-signal-col {
  width: 0.75rem !important;
  padding-left: 0.75rem !important;
  padding-right: 0.25rem !important;
}

.rl-signal {
  display: inline-block;
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 50%;
  background: var(--bs-secondary, #6c757d);
  box-shadow: 0 0 0 3px rgba(108, 117, 125, 0.15);
}

.rl-signal-danger { background: #e74a3b; box-shadow: 0 0 0 3px rgba(231, 74, 59, 0.18); }
.rl-signal-warning { background: #f6c23e; box-shadow: 0 0 0 3px rgba(246, 194, 62, 0.2); }
.rl-signal-success { background: #1cc88a; box-shadow: 0 0 0 3px rgba(28, 200, 138, 0.18); }
.rl-signal-info { background: #36b9cc; box-shadow: 0 0 0 3px rgba(54, 185, 204, 0.18); }
.rl-signal-muted { background: #adb5bd; box-shadow: none; }

.rl-row-danger { background-color: rgba(231, 74, 59, 0.06) !important; }
.rl-row-warning { background-color: rgba(246, 194, 62, 0.08) !important; }
.rl-row-success { background-color: rgba(28, 200, 138, 0.05) !important; }
.rl-row-info { background-color: rgba(54, 185, 204, 0.06) !important; }

.rl-group-header td {
  background: #f1f3f5 !important;
  font-weight: 600;
  font-size: 0.8125rem;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--bs-secondary-color, #6c757d);
  padding-top: 0.5rem !important;
  padding-bottom: 0.5rem !important;
}

.rl-sparkline {
  height: 4px;
  width: 100%;
  max-width: 8rem;
  background: rgba(0, 0, 0, 0.06);
  border-radius: 999px;
  overflow: hidden;
}

.rl-sparkline-bar {
  display: block;
  height: 100%;
  border-radius: 999px;
  background: var(--bs-primary, #4e73df);
  &.rl-signal-danger { background: #e74a3b; }
  &.rl-signal-warning { background: #f6c23e; }
  &.rl-signal-success { background: #1cc88a; }
  &.rl-signal-info { background: #36b9cc; }
  &.rl-signal-muted { background: #adb5bd; }
}

.rl-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 3rem 1.5rem;
  min-height: 12rem;
}

.rl-empty-icon {
  width: 3rem;
  height: 3rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f8f9fa;
  color: #adb5bd;
  font-size: 1.25rem;
  margin-bottom: 0.75rem;
}

.rl-empty-title {
  font-weight: 600;
  font-size: 1rem;
}

.rl-empty-text {
  font-size: 0.875rem;
  max-width: 22rem;
}

.rl-cards-wrap {
  width: 100%;
  overflow: auto;
  padding: 0.75rem;
}

.rl-card-group-title {
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--bs-secondary-color, #6c757d);
  margin: 0.5rem 0.25rem;
}

.rl-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(16rem, 1fr));
  gap: 0.75rem;
}

.rl-card {
  text-align: left;
  background: #fff;
  border: 1px solid var(--bs-border-color, #dee2e6);
  border-radius: 0.75rem;
  padding: 0.85rem 1rem;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
  transition: box-shadow 0.15s, border-color 0.15s;
  cursor: pointer;

  &:hover {
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.08);
    border-color: rgba(var(--bs-primary-rgb, 13 110 253), 0.35);
  }
}

.rl-card-head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.65rem;
}

.rl-card-title {
  font-weight: 600;
  font-size: 0.95rem;
}

.rl-card-row {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  font-size: 0.8125rem;
  padding: 0.2rem 0;
}

.rl-card-label {
  color: var(--bs-secondary-color, #6c757d);
  flex-shrink: 0;
}

.rl-card-value {
  text-align: right;
  min-width: 0;
}
</style>
<style lang="scss">
.rl-row-tooltip {
  position: fixed;
  z-index: 1090;
  pointer-events: none;
  background-color: #fff;
  color: #666;
  border: 1px solid #eee;
  border-radius: 4px;
  padding: 10px;
  box-shadow: 1px 2px 10px rgba(0, 0, 0, 0.2);
  font-size: 14px;
  line-height: 1.5;
  max-height: 60vh;
  overflow: hidden;
  font-family: sans-serif;
}

.rl-row-tooltip-title {
  margin-bottom: 4px;
  font-weight: 600;
  color: #333;
}

.rl-row-tooltip-row {
  display: flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
}

.rl-row-tooltip-marker {
  display: inline-block;
  flex-shrink: 0;
  width: 10px;
  height: 10px;
  border-radius: 10px;
  margin-right: 4px;
}

.rl-row-tooltip-name {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #666;
}

.rl-row-tooltip-value {
  margin-left: 20px;
  flex-shrink: 0;
  max-width: 60%;
  text-align: right;
  color: #666;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;

  // Match FieldViewer output to compact tooltip chrome
  :deep(.value) {
    margin: 0;
  }

  :deep(.badge) {
    font-size: 12px;
    vertical-align: middle;
  }

  :deep(a) {
    color: #5470c6;
  }
}

.record-list-table { .actions { padding-top: 8px; position: sticky; right: -1px; opacity: 0; transition: opacity 0.25s; width: 1%; font-family: var(--font-regular) !important; z-index: 3; &.actions-visible { opacity: 1; } } tbody tr td:nth-last-child(2) { padding-right: 5rem; } }
.record-list-footer { font-family: var(--font-medium); }
.active-filter { white-space: nowrap; font-family: var(--font-normal); .field-label { font-family: var(--font-medium); } &-text { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; vertical-align: middle; margin: 0; } &-item { vertical-align: middle; margin: 0; } &-close-btn { vertical-align: middle; opacity: 0.5; svg { height: 0.8rem; } &:hover { opacity: 1; } } }
</style>
