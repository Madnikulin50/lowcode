<template>
  <c-form-table-wrapper hide-add-button>
    <div class="mb-3 m-0">
      <label class="form-label text-primary">{{ $t('recordList.record.prefilterCommand') }}</label>
      <template v-if="textInput">
        <c-input-expression v-model="options.prefilter" min-height="3.688rem" :suggestion-params="recordAutoCompleteParams" />
        <small class="text-muted d-block mt-1">
          <code>${record.values.fieldName}</code>
          <code>${recordID}</code>
          <code>${ownerID}</code>
          <code>${userID}</code>, <code>${user.name}</code>
        </small>
        <div class="d-flex align-items-center justify-content-end mt-1">
          <button class="btn btn-link btn-sm text-decoration-none" @click="toggleFilterInputType">{{ $t('recordList.prefilter.toggleInputType') }}</button>
        </div>
      </template>
      <template v-else>
        <FilterToolbox v-model="filterGroup" :module="module" :namespace="namespace" reset-filter-on-created start-empty />
        <div class="d-flex align-items-center justify-content-end mt-1 gap-1">
          <button class="btn btn-outline-secondary btn-sm" @click="toggleFilterInputType">{{ $t('label.cancel') }}</button>
          <button class="btn btn-primary btn-sm ms-1" @click="saveFilter">{{ $t('label.save') }}</button>
        </div>
      </template>
    </div>
  </c-form-table-wrapper>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { compose } from 'corteza-lib/js/dist'
import { components } from 'corteza-lib/vue/dist'
import { convertRecordListFilter, queryToFilter } from 'corteza-webapp-compose/src/lib/record-filter.js'
import FilterToolbox from 'corteza-webapp-compose/src/components/Common/FilterToolbox.vue'

const { CInputExpression } = components
const { t: $t } = useI18n({ useScope: 'global' })

const props = defineProps({
  options: { type: Object, required: true },
  namespace: { type: compose.Namespace, required: true },
  module: { type: compose.Module, required: true },
  record: { type: [Object, null], required: false, default: null },
})

const textInput = ref(true)
const filterGroup = ref([])

const recordAutoCompleteParams = computed(() => {
  if (typeof window.__composeAPI?.processRecordAutoCompleteParams === 'function') {
    return window.__composeAPI.processRecordAutoCompleteParams({ operators: true })
  }
  return {}
})

function saveFilter(filter) {
  if (filter && filter[0] && !filter[0].filter[0].name) return
  props.options.prefilter = parseFilter()
  toggleFilterInputType()
}

function toggleFilterInputType() {
  textInput.value = !textInput.value
  filterGroup.value = []
}

function parseFilter() {
  return queryToFilter('', '', [], filterGroup.value.map(group => {
    group.filter = convertRecordListFilter(group.filter)
    return group
  }).filter(({ filter }) => filter?.length))
}
</script>
