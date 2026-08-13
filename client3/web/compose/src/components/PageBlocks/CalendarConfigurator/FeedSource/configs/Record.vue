<template>
  <div class="row">
    <div class="col-12 col-lg-6">
      <div class="mb-3">
        <label class="form-label text-primary">{{ $t('calendar.recordFeed.moduleLabel') }}</label>
        <div class="input-group">
          <c-input-select
            v-model="feed.options.moduleID"
            :options="modules"
            :reduce="o => o.moduleID"
            default-value="0"
            :placeholder="$t('calendar.recordFeed.modulePlaceholder')"
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
          <label class="form-label text-primary">{{ $t('calendar.recordFeed.titleLabel') }}</label>
          <c-input-select
            v-model="feed.titleField"
            :options="titleFields"
            :get-option-key="getOptionEventFieldKey"
            :get-option-label="getOptionEventFieldLabel"
            :reduce="o => o.name"
            :placeholder="$t('calendar.recordFeed.titlePlaceholder')"
          />
        </div>
      </div>

      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('calendar.recordFeed.eventStartFieldLabel') }}</label>
          <c-input-select
            v-model="feed.startField"
            :options="dateFields"
            :get-option-key="getOptionEventFieldKey"
            :get-option-label="getOptionEventFieldLabel"
            :reduce="o => o.name"
            :placeholder="$t('calendar.recordFeed.eventStartFieldPlaceholder')"
          />
        </div>
      </div>

      <div class="col-12 col-lg-6">
        <div class="mb-3">
          <label class="form-label text-primary">{{ $t('calendar.recordFeed.eventEndFieldLabel') }}</label>
          <c-input-select
            v-model="feed.endField"
            :options="dateFields"
            :get-option-key="getOptionEventFieldKey"
            :get-option-label="getOptionEventFieldLabel"
            :reduce="o => o.name"
            :disabled="feed.allDay"
            :placeholder="$t('calendar.recordFeed.eventEndFieldPlaceholder')"
          />
          <div class="form-check mt-1">
            <input
              v-model="feed.allDay"
              :true-value="true"
              :false-value="false"
              type="checkbox"
              class="form-check-input"
            />
            <label class="form-check-label">{{ $t('calendar.recordFeed.eventAllDay') }}</label>
          </div>
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
    </template>
  </div>
</template>

<script setup>
import { computed, inject } from 'vue'
import { components } from 'corteza-lib/vue/dist'
import { compose, NoID } from 'corteza-lib/js/dist'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const { CInputExpression } = components
const $auth = inject('$auth')

const props = defineProps({
  feed: { type: Object, required: true, default: () => ({}) },
  modules: { type: Array, required: false, default: () => [] },
  record: { type: compose.Record, required: false, default: undefined },
  page: { type: compose.Page, required: true },
})

const module = computed(() => {
  if (!(props.feed.options || {}).moduleID) return
  return props.modules.find(({ moduleID }) => moduleID === props.feed.options.moduleID)
})

const titleFields = computed(() => {
  if (!module.value) return []
  return [...module.value.fields]
    .filter(f => ['String', 'Email', 'Url'].includes(f.kind))
    .sort((a, b) => a.label.localeCompare(b.label))
})

const dateFields = computed(() => {
  if (!module.value) return []
  const moduleFields = module.value.fields.slice().sort((a, b) => a.label.localeCompare(b.label))
  return [
    ...moduleFields,
    ...module.value.systemFields().map(sf => {
      sf.label = `system.${sf.name}`
      return sf
    }),
  ].filter(f => f.kind === 'DateTime' && !f.isMulti)
})

const isRecordPage = computed(() => props.page && props.page.moduleID !== NoID)

const recordAutoCompleteParams = computed(() => processRecordAutoCompleteParams({ module: module.value, operators: true }))

function processRecordAutoCompleteParams ({ module: mod, operators = false } = {}) {
  const { fields = [] } = mod || {}
  const moduleFields = fields.map(({ name }) => name)
  const userProperties = ($auth?.user?.properties?.()) || []

  const recordSuggestions = isRecordPage.value && props.record
    ? [
        ...(['ownerID', 'recordID'].map(value => ({ interpolate: true, value }))),
        {
          interpolate: true,
          value: 'record',
          properties: [
            { value: 'values', properties: Object.keys(props.record.values) || [] },
            ...(props.record.properties || []),
          ],
        },
      ]
    : []

  return [
    ...recordSuggestions,
    ...(operators ? ['AND', 'OR'] : []),
    { interpolate: true, value: 'userID' },
    { interpolate: true, value: 'user', properties: userProperties },
    ...moduleFields,
  ]
}

function onModuleChange () {
  props.feed.titleField = ''
  props.feed.startField = ''
  props.feed.endField = ''
}

function getOptionEventFieldKey ({ name }) {
  return name
}

function getOptionEventFieldLabel ({ name, label }) {
  return label || name
}
</script>
