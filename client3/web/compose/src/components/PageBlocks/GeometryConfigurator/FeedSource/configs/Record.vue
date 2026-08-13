<template>
  <div class="row">
    <template v-if="feed.options">
      <div class="col-12">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('geometry.recordFeed.moduleLabel') }}</label>
          <div class="input-group">
            <c-input-select
              v-model="feed.options.moduleID"
              :options="modules"
              :reduce="o => o.moduleID"
              :placeholder="$t('calendar.recordFeed.modulePlaceholder')"
              default-value="0"
              label="name"
              class="flex-grow-1"
              @input="onModuleChange"
            />
          </div>
        </div>
      </div>

      <template v-if="module">
        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('geometry.recordFeed.geometryFieldLabel') }}</label>
            <c-input-select
              v-model="feed.geometryField"
              :options="geometryFields"
              :get-option-key="getOptionGeometryAndTitleFieldKey"
              :get-option-label="getOptionGeometryAndTitleFieldLabel"
              :reduce="o => o.name"
              :placeholder="$t('geometry.recordFeed.geometryFieldPlaceholder')"
            />
          </div>
        </div>

        <div class="col-12 col-lg-6">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('geometry.recordFeed.titleLabel') }}</label>
            <c-input-select
              v-model="feed.titleField"
              :options="titleFields"
              :get-option-key="getOptionGeometryAndTitleFieldKey"
              :get-option-label="getOptionGeometryAndTitleFieldLabel"
              :reduce="o => o.name"
              :placeholder="$t('geometry.recordFeed.titlePlaceholder')"
            />
          </div>
        </div>

        <div class="col-12">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('calendar.recordFeed.prefilterLabel') }}</label>
            <c-input-expression
              v-model="feed.options.prefilter"
              min-height="3.688rem"
              :suggestion-params="recordAutoCompleteParams"
              :placeholder="$t('calendar.recordFeed.prefilterPlaceholder')"
            />
            <i18next
              path="interpolationFootnote"
              tag="small"
              class="text-muted"
            >
              <code>${record.values.fieldName}</code>
              <code>${recordID}</code>
              <code>${ownerID}</code>
              <span><code>${userID}</code>, <code>${user.name}</code></span>
            </i18next>
          </div>
        </div>

        <div class="col-12 col-lg-4">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('geometry.recordFeed.displayMarker') }}</label>
            <c-input-checkbox
              v-model="feed.displayMarker"
              switch
              :labels="checkboxLabel"
            />
          </div>
        </div>

        <div class="col-12 col-lg-4">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('geometry.recordFeed.displayPolygon') }}</label>
            <c-input-checkbox
              v-model="feed.displayPolygon"
              switch
              :labels="checkboxLabel"
            />
          </div>
        </div>

        <div class="col-12 col-lg-4">
          <div class="mb-3">
            <label class="form-label text-primary">{{ $t('geometry.recordFeed.colorLabel') }}</label>
            <c-input-color-picker
              v-model="feed.options.color"
              :translations="{
                modalTitle: $t('geometry.recordFeed.colorPicker'),
                light: $t('themes.labels.light'),
                dark: $t('themes.labels.dark'),
                cancelBtnLabel: $t('label.cancel'),
                saveBtnLabel: $t('label.saveAndClose')
              }"
              :theme-settings="themeSettings"
            />
          </div>
        </div>
      </template>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, inject } from 'vue'
import { components } from 'corteza-lib/vue/dist'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const { CInputColorPicker, CInputExpression } = components

const props = defineProps({
  feed: { type: Object, required: true, default: () => ({}) },
  modules: { type: Array, required: false, default: () => [] },
})

const $Settings = inject('$Settings')
const $auth = inject('$auth')
const checkboxLabel = ref({ on: t('label.yes'), off: t('label.no') })

const module = computed(() => {
  if (!(props.feed.options || {}).moduleID) return
  return props.modules.find(({ moduleID }) => moduleID === props.feed.options.moduleID)
})

const recordAutoCompleteParams = computed(() => processRecordAutoCompleteParams({ module: module.value, operators: true }))

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

const titleFields = computed(() => {
  if (!module.value) return []
  return module.value.fields
    .filter(f => ['DateTime', 'Select', 'Number', 'Bool', 'String', 'Record', 'User'].includes(f.kind) && f.label)
    .sort((a, b) => a.label.localeCompare(b.label))
})

const geometryFields = computed(() => {
  if (!module.value) return []
  return [
    ...module.value.fields,
    ...module.value.systemFields().map(sf => { sf.label = `system.${sf.name}`; return sf }),
  ].filter(f => f.kind === 'Geometry')
    .sort((a, b) => a.label.localeCompare(b.label))
})

const themeSettings = computed(() => $Settings.get('ui.studio.themes', []))

function onModuleChange () {
  props.feed.geometryField = ''
  props.feed.titleField = ''
}

function getOptionGeometryAndTitleFieldKey ({ name }) { return name }
function getOptionGeometryAndTitleFieldLabel ({ name, label }) { return label || name }
</script>
