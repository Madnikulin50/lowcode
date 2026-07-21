<template>
  <div class="card shadow-sm mt-3" data-test-id="card-filter-list">
    <div class="card-header">
      <h4 class="m-0">
        {{ $t('filters.title') }}
      </h4>
    </div>

    <ul class="nav nav-tabs card-header-tabs" data-test-id="filter-steps" role="tablist">
      <li
        v-for="(step, index) in steps"
        :key="index"
        class="nav-item"
        role="presentation"
      >
        <button
          :id="steps[index]"
          class="nav-link"
          :class="{ active: selectedTab === index }"
          :data-bs-toggle="'tab'"
          :data-bs-target="'#tab-' + index"
          type="button"
          role="tab"
          @click="onActivateTab(index)"
        >
          {{ $t(`filters.step_title.${step}`) }}
        </button>
      </li>
    </ul>

    <div class="tab-content border-0 p-0">
      <div
        v-for="(step, index) in steps"
        :key="index"
        :id="'tab-' + index"
        class="tab-pane"
        :class="{ show: selectedTab === index, active: selectedTab === index }"
        role="tabpanel"
      >
        <c-filters-dropdown
          :available-filters="getAvailableFiltersByStep"
          :filters="getSelectedFiltersByStep"
          class="p-3"
          @addFilter="onAddFilter"
        />

        <c-filters-table
          :filters="getSelectedFiltersByStep"
          :step="index"
          :fetching="fetching"
          @filterSelect="onFilterSelect"
          @removeFilter="onRemoveFilter"
          @sortFilters="onSortFilters"
        />
      </div>
    </div>

    <div class="card-footer border-top d-flex flex-wrap flex-fill-child gap-1">
      <c-button-submit
        :disabled="disabled"
        :processing="processing"
        :success="success"
        :text="$t('label.submit')"
        class="ms-auto"
        @submit="$emit('submit')"
      />
    </div>

    <c-filter-modal
      :visible="!!selectedFilter"
      :filter="selectedFilter"
      @submit="onSubmit"
      @reset="onReset"
    />
  </div>
</template>

<script setup>
import { ref, computed, getCurrentInstance } from 'vue'
import CFilterModal from 'corteza-webapp-admin/src/components/Apigw/CFilterModal'
import CFiltersTable from 'corteza-webapp-admin/src/components/Apigw/CFiltersTable'
import CFiltersDropdown from 'corteza-webapp-admin/src/components/Apigw/CFiltersDropdown'

const { proxy } = getCurrentInstance()

const mapKindToStep = {
  prefilter: 0,
  processer: 1,
  postfilter: 2,
}

const props = defineProps({
  fetching: {
    type: Boolean,
    value: false,
  },
  processing: {
    type: Boolean,
    value: false,
  },
  success: {
    type: Boolean,
    value: false,
  },
  filters: {
    type: Array,
    required: true,
  },
  availableFilters: {
    type: Array,
    required: true,
  },
  steps: {
    type: Array,
    required: true,
  },
})

const emit = defineEmits(['update:filters', 'submit'])

const selectedFilter = ref(null)
const selectedTab = ref(0)

const disabled = computed(() => {
  return !(props.filters.some(({ updated, created, deleted }) => updated || created || deleted))
})

const getSelectedFilter = computed(() => {
  return selectedFilter.value ? selectedFilter.value : null
})

const getAvailableFiltersByStep = computed(() => {
  return (props.availableFilters || []).filter(({ kind }) => {
    return mapKindToStep[kind] === selectedTab.value
  })
})

const getSelectedFiltersByStep = computed(() => {
  return (props.filters || []).filter(({ kind, deleted }) => {
    return mapKindToStep[kind] === selectedTab.value && !deleted
  }).sort((a, b) => a.weight - b.weight)
})

const disabledRemoveButton = computed(() => {
  return !props.filters.some(({ options }) => (options || { checked: false }).checked)
})

function onAddFilter (filter) {
  const i = props.filters.findIndex(({ ref, deleted }) => ref === filter.ref && !deleted)

  if (i < 0) {
    selectedFilter.value = filter
  } else {
    selectedFilter.value = props.filters[i]
  }
}

function onSubmit (filter) {
  const i = props.filters.findIndex(({ ref, deleted }) => ref === filter.ref && !deleted)

  if (i < 0) {
    filter.weight = getSelectedFiltersByStep.value.length
    props.filters.push(filter)
  } else {
    props.filters.splice(props.filters.findIndex(({ ref, deleted }) => ref === filter.ref && !deleted), 1, filter)
  }

  emit('update.filters', props.filters)
}

function onReset () {
  selectedFilter.value = undefined
}

function onSortFilters (sortedFilters) {
  props.filters.forEach(filter => {
    const i = sortedFilters.findIndex(({ ref, deleted }) => ref === filter.ref && !deleted)

    if (i >= 0) {
      filter.weight = sortedFilters.findIndex(({ ref, deleted }) => ref === filter.ref && !deleted)
      filter.updated = true
    }
  })
  props.filters.sort((a, b) => a.weight - b.weight)
}

function onRemoveFilter (filter) {
  if (filter.filterID) {
    props.filters.splice(props.filters.findIndex(({ filterID, deleted }) => filterID === filter.filterID && !deleted), 1, { ...filter, deleted: true })
  } else {
    props.filters.splice(props.filters.findIndex(({ ref, deleted }) => ref === filter.ref && !deleted), 1)
  }

  emit('update.filters', props.filters)
}

function onFilterSelect (filter = {}) {
  selectedFilter.value = { ...filter }
}

function onActivateTab (index) {
  selectedTab.value = index
}
</script>
