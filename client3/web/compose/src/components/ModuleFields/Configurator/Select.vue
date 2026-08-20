<template>
  <div class="row g-0">
    <div class="col">
      <div class="mb-3">
        <label class="form-label text-primary mr-2">{{ t('kind.select.optionType.label') }}</label>
        <div class="btn-group" data-bs-toggle="buttons">
          <label
            v-for="opt in selectOptions"
            :key="opt.value"
            class="btn btn-outline-primary btn-sm"
            :class="{ active: f.options.selectType === opt.value }"
          >
            <input
              type="radio"
              class="btn-check"
              :value="opt.value"
              :checked="f.options.selectType === opt.value"
              autocomplete="off"
              @change="onSelectTypeChange(opt.value)"
            />
            {{ opt.text }}
          </label>
        </div>
      </div>

      <div v-if="shouldAllowDuplicates" class="form-check mb-3">
        <input
          id="allowDuplicates"
          v-model="f.options.isUniqueMultiValue"
          :true-value="false"
          :false-value="true"
          type="checkbox"
          class="form-check-input"
        />
        <label class="form-check-label" for="allowDuplicates">{{ t('kind.select.allow-duplicates') }}</label>
      </div>

      <div class="mb-3">
        <label class="form-label text-primary mr-2">{{ t('kind.select.displayType.label') }}</label>
        <div class="btn-group" data-bs-toggle="buttons">
          <label
            v-for="opt in displayOptions"
            :key="opt.value"
            class="btn btn-outline-primary btn-sm"
            :class="{ active: f.options.displayType === opt.value }"
          >
            <input
              type="radio"
              class="btn-check"
              :value="opt.value"
              :checked="f.options.displayType === opt.value"
              autocomplete="off"
              @change="f.options.displayType = opt.value"
            />
            {{ opt.text }}
          </label>
        </div>
      </div>

      <div v-if="f.options.displayType === 'badge'" class="form-check mb-3">
        <input
          id="badgeGradient"
          v-model="f.options.badgeGradient"
          type="checkbox"
          class="form-check-input"
        />
        <label class="form-check-label" for="badgeGradient">{{ t('kind.select.options.badgeGradient') }}</label>
      </div>

      <div class="mb-3">
        <label class="form-label text-primary">{{ t('kind.select.optionsLabel') }}</label>
        <c-form-table-wrapper
          :labels="{ addButton: t('label.add') }"
          @add-item="handleAddOption"
        >
          <table class="table table-sm" borderless="borderless">
            <thead>
              <tr>
                <th v-if="f.options.options.length > 0"></th>
                <th class="text-primary" style="min-width: 200px;">{{ t('kind.select.options.value') }}</th>
                <th class="text-primary">{{ t('kind.select.options.label') }}</th>
                <th v-if="f.options.displayType === 'badge'" class="text-primary">{{ t('kind.select.options.style.textColor') }}</th>
                <th v-if="f.options.displayType === 'badge'" class="text-primary">{{ t('kind.select.options.style.backgroundColor') }}</th>
                <th></th>
              </tr>
            </thead>
            <draggable
            item-key="id"
              v-model="f.options.options"
              group="sort"
              handle=".grab"
              tag="tbody"
            >
              <template #item="{ element, index }">
                <tr :key="index">
                  <td class="align-middle text-center">
                    <font-awesome-icon :icon="['fas', 'bars']" class="grab text-secondary" />
                  </td>
                  <td style="min-width: 200px;">
                    <input
                      v-model.trim="f.options.options[index].value"
                      type="text"
                      class="form-control form-control-sm"
                      :placeholder="t('kind.select.options.value')"
                      :class="{ 'is-invalid': !f.options.options[index].value }"
                    />
                  </td>
                  <td style="min-width: 200px;">
                    <div class="input-group input-group-sm">
                      <input
                        v-model.trim="f.options.options[index].text"
                        type="text"
                        class="form-control form-control-sm"
                        :placeholder="t('kind.select.options.label')"
                        :class="{ 'is-invalid': !f.options.options[index].text }"
                      />
                      <FieldSelectTranslator
                        v-if="field"
                        :field="field"
                        :module="module"
                        :highlight-key="`meta.options.${element.value}.text`"
                        size="sm"
                        :disabled="isNew || element.new"
                      />
                    </div>
                  </td>
                  <td v-if="f.options.displayType === 'badge'" style="min-width: 120px;">
                    <CInputColorPicker
                      :model-value="resolveColor(f.options.options[index].style.textColor)"
                      :default-value="defaultTextColor"
                      :theme-settings="themeSettings"
                      :translations="{
                        modalTitle: t('kind.select.options.style.textColor'),
                        cancelBtnLabel: t('label.cancel'),
                        saveBtnLabel: t('label.saveAndClose')
                      }"
                      @update:model-value="setOptionStyle(index, 'textColor', $event)"
                    />
                  </td>
                  <td v-if="f.options.displayType === 'badge'" style="min-width: 130px;">
                    <CInputColorPicker
                      :model-value="resolveColor(f.options.options[index].style.backgroundColor)"
                      :default-value="defaultBackgroundColor"
                      :theme-settings="themeSettings"
                      :translations="{
                        modalTitle: t('kind.select.options.style.backgroundColor'),
                        defaultBtnLabel: t('label.default'),
                        cancelBtnLabel: t('label.cancel'),
                        saveBtnLabel: t('label.saveAndClose')
                      }"
                      @update:model-value="setOptionStyle(index, 'backgroundColor', $event)"
                    />
                  </td>
                  <td class="align-middle text-end">
                    <c-input-confirm
                      show-icon
                      @confirmed="f.options.options.splice(index, 1)"
                    />
                  </td>
                </tr>
              </template>
            </draggable>
          </table>
        </c-form-table-wrapper>
      </div>
    </div>
  </div>
</template>

<script setup>
defineOptions({ i18nOptions: { namespaces: 'field' } })
import { ref, computed, inject, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useConfiguratorBase } from './base'
import Draggable from 'vuedraggable'
import { NoID } from 'corteza-lib/js/dist'
import FieldSelectTranslator from 'corteza-webapp-compose/src/components/Admin/Module/FieldSelectTranslator'
import { components } from 'corteza-lib/vue/dist'
import { themeColor } from 'corteza-webapp-compose/src/lib/color.js'
const { CInputColorPicker } = components

const props = defineProps({
  namespace: { type: Object, required: true },
  module: { type: Object, required: true },
  field: { type: Object, required: true },
  hasRecords: { type: Boolean, default: false },
})

const emit = defineEmits(['update:field'])

const { t } = useI18n({ useScope: 'global', messages: {} })
const { f, isNew, hasData } = useConfiguratorBase(props, emit)

const $settings = inject('$Settings')

const newOption = ref({ value: undefined, text: undefined, new: true })

const selectTypes = ref([
  { text: t('kind.select.optionType.default'), value: 'default', allowDuplicates: true },
  { text: t('kind.select.optionType.multiple'), value: 'multiple', onlyMulti: true },
  { text: t('kind.select.optionType.each'), value: 'each', allowDuplicates: true, onlyMulti: true },
  { value: 'list' },
])

const newEmpty = computed(() => !newOption.value.text || !newOption.value.value)

const newOptState = computed(() => {
  if (f.value.options.options.find(({ text, value }) => text === newOption.value.text || value === newOption.value.value)) {
    return false
  }
  return null
})

const selectOptions = computed(() => {
  const result = selectTypes.value.map((o) => {
    if (o.value === 'list') {
      o.text = t(`kind.select.optionType.${f.value.isMulti ? 'checkbox' : 'radio'}`)
    }
    return o
  })
  if (f.value.isMulti) return result
  return result.filter(({ onlyMulti }) => !onlyMulti)
})

const displayOptions = computed(() => [
  { text: t('kind.select.displayType.text'), value: 'text' },
  { text: t('kind.select.displayType.badge'), value: 'badge' },
])

const shouldAllowDuplicates = computed(() => {
  if (!f.value.isMulti) return false
  const { allowDuplicates } = selectTypes.value.find(({ value }) => value === f.value.options.selectType) || {}
  return !!allowDuplicates
})

const themeSettings = computed(() => {
  return $settings?.get ? $settings.get('ui.studio.themes', []) : []
})

const defaultTextColor = computed(() => {
  return getComputedStyle(document.documentElement).getPropertyValue('--dark')
})

const defaultBackgroundColor = computed(() => {
  return getComputedStyle(document.documentElement).getPropertyValue('--extra-light')
})

function resolveColor (val) {
  return themeColor(val, themeSettings.value)
}

function setOptionStyle (index, key, value) {
  const opt = f.value.options.options[index]
  if (!opt) return
  opt.style = opt.style || {}
  opt.style[key] = value
}

if (!f.value) {
  f.value.options = { options: [] }
} else if (!f.value.options.options) {
  f.value.options.options = []
}

onBeforeUnmount(() => {
  setDefaultValues()
})

function handleAddOption () {
  const option = f.value.createSelectOption()
  option.new = true
  f.value.options.options.push(option)
}

function onSelectTypeChange (value) {
  f.value.options.selectType = value
  updateIsUniqueMultiValue(value)
}

function updateIsUniqueMultiValue (value) {
  const { allowDuplicates = false } = selectTypes.value.find(({ value: v }) => v === value) || {}
  if (!allowDuplicates) {
    f.value.options.isUniqueMultiValue = true
  }
}

function setDefaultValues () {
  newOption.value = {}
  selectTypes.value = []
}
</script>
