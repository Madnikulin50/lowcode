<template>
  <div>
    <h5 class="mb-3">{{ $t('recordList.record.generalLabel') }}</h5>
    <div class="row">
      <div class="col-12">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('module') }}</label>
          <input v-if="module" v-model="module.name" type="text" class="form-control" readonly />
        </div>
      </div>
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('record.inlineEdit.enabled') }}</label>
          <c-input-checkbox v-model="inlineRecordEditEnabled" switch :labels="checkboxLabel" />
        </div>
      </div>
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('record.inlineEdit.allowAddField') }}</label>
          <c-input-checkbox v-model="options.inlineRecordEditAllowAddField" switch :labels="checkboxLabel" />
        </div>
      </div>
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('record.horizontalFormLayout') }}</label>
          <c-input-checkbox v-model="options.horizontalFieldLayoutEnabled" switch :disabled="options.recordFieldLayoutOption === 'noWrap'" :labels="checkboxLabel" />
        </div>
      </div>
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('record.fieldsLayoutMode.label') }}</label>
          <c-input-select v-model="options.recordFieldLayoutOption" :options="recordFieldLayoutOptions" :reduce="option => option.value" :get-option-key="option => option.label" @input="handleRecordFieldLayout" />
        </div>
      </div>
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('record.referenceRecordField') }}</label>
          <small class="text-muted d-block mb-1">{{ $t('record.referenceRecordFieldDescription') }}</small>
          <c-input-select v-model="options.referenceField" :options="recordSelectorFields" :get-option-label="f => f.label || f.name" :get-option-key="f => f.fieldID !== NoID ? f.fieldID : f.name" :placeholder="$t('record.referenceRecordFieldPlaceholder')" :reduce="f => f.fieldID !== NoID ? f.fieldID : f.name" @input="updateReferenceModule($event, [])" />
        </div>
      </div>
    </div>

    <hr />
    <h5 class="mb-3">{{ $t('record.appearance.label') }}</h5>
    <div class="row">
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('record.appearance.density') }}</label>
          <select v-model="options.density" class="form-select form-control">
            <option value="comfortable">{{ $t('record.appearance.densityOptions.comfortable') }}</option>
            <option value="compact">{{ $t('record.appearance.densityOptions.compact') }}</option>
          </select>
        </div>
      </div>
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('record.appearance.hideEmptyFields') }}</label>
          <c-input-checkbox v-model="options.hideEmptyFields" switch :labels="checkboxLabel" />
        </div>
      </div>
      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('record.appearance.showEmptyPlaceholder') }}</label>
          <small class="text-muted d-block mb-1">{{ $t('record.appearance.showEmptyPlaceholderDescription') }}</small>
          <c-input-checkbox v-model="options.showEmptyPlaceholder" switch :labels="checkboxLabel" :disabled="options.hideEmptyFields" />
        </div>
      </div>
    </div>
    <div class="row">
      <div class="col-12 col-lg-12">
        <BlockStyle class="mt-2" :value-color="true" :label-color="true" :font-size="true" :options="viewStyle">
          <h5>{{ $t('metric.editStyle.valueLabel') }}</h5>
        </BlockStyle>
      </div>
    </div>

    <hr v-if="module" />

    <div v-if="module" class="">
      <h5 class="mb-3">{{ $t('module.general.fields') }}</h5>
      <div class="row">
        <div class="col-12">
          <FieldPicker :module="fieldModule" v-model:fields="options.fields" style="height: 52vh;" />
        </div>
      </div>

      <div v-if="configuredFields.length" class="mt-4">
        <h5 class="mb-2">{{ $t('record.appearance.fieldRoles') }}</h5>
        <small class="text-muted d-block mb-2">{{ $t('record.appearance.fieldRolesDescription') }}</small>
        <table class="table table-sm align-middle">
          <thead>
            <tr>
              <th class="text-primary">{{ $t('record.appearance.field') }}</th>
              <th class="text-primary" style="width: 14rem;">{{ $t('record.appearance.role') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="field in configuredFields" :key="field.name">
              <td>
                <span class="fw-medium">{{ field.label || field.name }}</span>
                <small class="text-muted ms-1">({{ field.name }})</small>
              </td>
              <td>
                <select
                  class="form-select form-control form-select-sm"
                  :value="fieldRole(field.name)"
                  @change="setFieldRole(field.name, $event.target.value)"
                >
                  <option
                    v-for="opt in fieldRoleOptions"
                    :key="opt.value"
                    :value="opt.value"
                  >
                    {{ opt.text }}
                  </option>
                </select>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="mt-4">
        <h5 class="d-flex align-items-center mb-2">
          {{ $t('record.appearance.sections') }}
        </h5>
        <small class="text-muted d-block mb-2">{{ $t('record.appearance.sectionsDescription') }}</small>
        <c-form-table-wrapper :labels="{ addButton: $t('record.appearance.addSection') }" class="my-2" @add-item="addSection">
          <table v-if="options.sections?.length" class="table table-sm table-borderless align-middle">
            <thead>
              <tr>
                <th class="text-primary">{{ $t('record.appearance.sectionTitle') }}</th>
                <th class="text-primary">{{ $t('record.appearance.sectionFields') }}</th>
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
                    :placeholder="$t('record.appearance.sectionTitlePlaceholder')"
                  >
                </td>
                <td style="min-width: 16rem;">
                  <c-input-select
                    v-model="section.fields"
                    :options="configuredFields"
                    multiple
                    :close-on-select="false"
                    :get-option-label="f => f.label || f.name"
                    :get-option-key="f => f.name"
                    :reduce="f => f.name"
                    :placeholder="$t('record.appearance.sectionFieldsPlaceholder')"
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

      <div v-if="isRecordFieldUsedConfigured" class="row mt-3">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('record.recordSelectorDisplayOptions') }}</label>
            <select v-model="options.recordSelectorDisplayOption" class="form-select form-control">
              <option v-for="opt in recordDisplayOptions" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
            </select>
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('record.recordSelectorCanAddRecord') }}</label>
            <c-input-checkbox v-model="options.recordSelectorShowAddRecordButton" switch :labels="checkboxLabel" />
          </div>
        </div>
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('record.recordSelectorAddRecordDisplayOption') }}</label>
            <select v-model="options.recordSelectorAddRecordDisplayOption" class="form-select form-control" :disabled="!options.recordSelectorShowAddRecordButton">
              <option v-for="opt in recordDisplayOptions" :key="opt.value" :value="opt.value">{{ opt.text }}</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <hr />

    <div class="">
      <h5 class="d-flex align-items-center mb-2">
        {{ $t('record.fieldConditions.label') }}
        <c-hint :tooltip="$t('record.fieldConditions.tooltip.performance')" icon-class="text-warning" />
        <a :href="visibilityDocumentationURL" target="_blank" class="btn btn-link p-0 ms-auto">{{ $t('label.examples') }}</a>
      </h5>

      <div class="row mt-3">
        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('record.fieldConditions.clearAllOnHide') }}</label>
            <small class="text-muted d-block mb-1">{{ $t('record.fieldConditions.clearAllOnHideDescription') }}</small>
            <c-input-checkbox v-model="options.clearConditionalFieldsOnHide" switch :labels="checkboxLabel" />
          </div>
        </div>
      </div>

      <c-form-table-wrapper :labels="{ addButton: $t('label.add') }" class="my-3" @add-item="addRule">
        <table v-if="block.options.fieldConditions.length > 0" class="table table-sm table-borderless">
          <thead>
            <tr>
              <th class="text-primary">{{ $t('record.fieldConditions.field') }}</th>
              <th class="text-primary">{{ $t('record.fieldConditions.condition') }}</th>
              <th class="text-primary text-center" style="width: 150px;">
                <div class="d-flex align-items-center justify-content-center">
                  {{ $t('record.fieldConditions.clearOnHide') }}
                  <c-hint :tooltip="$t('record.fieldConditions.clearOnHideTooltip')" class="ms-1" />
                </div>
              </th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(condition, i) in block.options.fieldConditions" :key="i">
              <td style="width: 33%; min-width: 250px;">
                <c-input-select v-model="condition.field" :options="block.options.fields" :placeholder="$t('record.fieldConditions.selectPlaceholder')" :selectable="option => isSelectable(option)" :get-option-label="option => option.label || option.name" :get-option-key="option => option.fieldID !== NoID ? option.fieldID : option.name" :reduce="option => option.isSystem ? option.name : option.fieldID" />
              </td>
              <td class="align-middle" style="min-width: 300px;">
                <div class="input-group">
                  <span class="input-group-text bg-extra-light">ƒ</span>
                  <c-input-expression v-model="condition.condition" auto-complete :placeholder="$t('record.fieldConditions.placeholder')" :suggestion-params="visibilityAutoCompleteParams" class="flex-grow-1" />
                </div>
              </td>
              <td style="width: 20px;">
                <div class="d-flex align-items-center justify-content-center">
                  <c-input-checkbox v-model="condition.clearOnHide" switch />
                </div>
              </td>
              <td class="text-end" style="width: 4rem;">
                <c-input-confirm show-icon @confirmed="deleteRule(i)" />
              </td>
            </tr>
          </tbody>
        </table>
      </c-form-table-wrapper>

      <small class="text-muted d-block">
        <code>record.values.fieldName</code>
        <code>user.(userID/email...)</code>
        <code>user.userID == record.createdBy</code>
        <code>record.values.fieldName == "value"</code>
        <code>record.ownedBy == user.userID</code>
        <code>screen.width &lt; 1024</code>
      </small>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'block' } })
import { ref, computed, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { useStore } from '../../store'
import BlockStyle from './BlockStyle'
import FieldPicker from 'corteza-webapp-compose/src/components/Common/FieldPicker'
import { NoID, compose } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'

const { CInputExpression } = components
const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  block: { type: Object, required: true },
  module: { type: Object, required: false },
  namespace: { type: Object, required: true },
})

const store = useStore()
const $auth = inject('$auth')

const options = computed(() => {
  const o = props.block.options
  if (!o.fieldRoles || typeof o.fieldRoles !== 'object') o.fieldRoles = {}
  if (!Array.isArray(o.sections)) o.sections = []
  if (!o.density) o.density = 'comfortable'
  if (o.hideEmptyFields === undefined) o.hideEmptyFields = false
  if (o.showEmptyPlaceholder === undefined) o.showEmptyPlaceholder = true
  return o
})
const referenceModule = ref(undefined)

const checkboxLabel = computed(() => ({ on: $t('label.yes'), off: $t('label.no') }))

const visibilityDocumentationURL = computed(() => {
  const year = '2024'; const month = '12'
  return `https://docs.cortezaproject.org/corteza-docs/${year}.${month}/integrator-guide/compose-configuration/page-layouts.html#visibility-condition`
})

const recordDisplayOptions = computed(() => [
  { value: 'sameTab', text: $t('record.openInSameTab') },
  { value: 'newTab', text: $t('record.openInNewTab') },
  { value: 'modal', text: $t('record.openInModal') },
])

const recordFieldLayoutOptions = computed(() => [
  { value: 'default', label: $t('record.fieldsLayoutMode.default') },
  { value: 'noWrap', label: $t('record.fieldsLayoutMode.noWrap') },
  { value: 'wrap', label: $t('record.fieldsLayoutMode.wrap') },
])

const fieldRoleOptions = computed(() => [
  { value: 'default', text: $t('record.appearance.roleOptions.default') },
  { value: 'title', text: $t('record.appearance.roleOptions.title') },
  { value: 'subtitle', text: $t('record.appearance.roleOptions.subtitle') },
  { value: 'badge', text: $t('record.appearance.roleOptions.badge') },
  { value: 'meta', text: $t('record.appearance.roleOptions.meta') },
  { value: 'body', text: $t('record.appearance.roleOptions.body') },
])

const recordSelectorFields = computed(() => props.module?.fields.filter(f => f.kind === 'Record' && !f.isMulti) || [])

const fieldModule = computed(() => (options.value.referenceField && referenceModule.value) ? referenceModule.value : props.module)

const configuredFields = computed(() => {
  const mod = fieldModule.value
  if (!mod) return []
  const selected = options.value.fields
  if (selected?.length) {
    return mod.filterFields(selected)
  }
  return mod.fields || []
})

const inlineRecordEditEnabled = computed({
  get: () => !!options.value.inlineRecordEditEnabled,
  set: (v) => { options.value.inlineRecordEditEnabled = v },
})

const viewStyle = computed({
  get: () => options.value.viewStyle || { fontSize: 20 },
  set: (v) => { options.value.viewStyle = v },
})

const isRecordFieldUsedConfigured = computed(() => {
  if (!options.value.fields?.length) return props.module?.fields.some(f => f.kind === 'Record')
  return options.value.fields.some(f => f.kind === 'Record')
})

const visibilityAutoCompleteParams = computed(() => processVisibilityAutoCompleteParams({ module: fieldModule.value }))

function processVisibilityAutoCompleteParams ({ module: mod } = {}) {
  const { fields = [] } = mod || {}
  const moduleFields = fields.map(({ name }) => name)
  const userProperties = ($auth?.user?.properties?.()) || []

  return [
    { value: 'user', properties: userProperties },
    { value: 'screen', properties: ['width', 'height', 'userAgent', 'breakpoint'] },
    ...moduleFields,
  ]
}

function fieldRole (name) {
  return options.value.fieldRoles?.[name] || 'default'
}

function setFieldRole (name, role) {
  if (!options.value.fieldRoles) options.value.fieldRoles = {}
  if (!role || role === 'default') {
    delete options.value.fieldRoles[name]
  } else {
    // Only one title / subtitle
    if (role === 'title' || role === 'subtitle') {
      for (const [k, v] of Object.entries(options.value.fieldRoles)) {
        if (v === role && k !== name) delete options.value.fieldRoles[k]
      }
    }
    options.value.fieldRoles[name] = role
  }
}

function addSection () {
  if (!Array.isArray(options.value.sections)) options.value.sections = []
  options.value.sections.push({ title: '', fields: [] })
}

function removeSection (i) {
  options.value.sections.splice(i, 1)
}

function addRule() {
  options.value.fieldConditions.push({ field: undefined, condition: '', clearOnHide: false })
}

function deleteRule(i) {
  options.value.fieldConditions.splice(i, 1)
}

function isSelectable(option) {
  return !props.block.options.fieldConditions.find(({ field }) => field === option.fieldID || field === option.name) && !option.isRequired
}

function updateReferenceModule(fieldID, fields) {
  if (!fieldID) {
    options.value.fields = []
    options.value.referenceModuleID = undefined
    return
  }
  const field = recordSelectorFields.value.find(f => f.fieldID === fieldID || f.name === fieldID)
  const moduleID = field?.options?.moduleID
  if (moduleID) {
    store.module.findByID({ namespace: props.namespace.namespaceID, moduleID })
      .then(mod => {
        options.value.fields = fields
        options.value.referenceModuleID = mod.moduleID
        referenceModule.value = new compose.Module({ ...mod })
      })
  }
}

function handleRecordFieldLayout(v) {
  if (v === 'noWrap') options.value.horizontalFieldLayoutEnabled = false
}
</script>
