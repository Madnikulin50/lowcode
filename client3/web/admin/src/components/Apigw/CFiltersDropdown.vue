<template>
  <div class="dropdown" data-test-id="dropdown-add-filter">
    <button
      class="btn btn-primary dropdown-toggle"
      type="button"
      data-bs-toggle="dropdown"
      aria-expanded="false"
    >
      {{ $t('filters.addFilter') }}
    </button>
    <ul class="dropdown-menu">
      <template v-if="filterList.length">
        <li
          v-for="(filter, index) in filterList"
          :key="index"
        >
          <button
            :data-test-id="filterDropdownCypressId(filter.label)"
            class="dropdown-item"
            :disabled="filter.disabled"
            @click="onAddFilter(filter)"
          >
            {{ filter.label }}
          </button>
        </li>
      </template>
      <li v-else>
        <button
          class="dropdown-item"
          disabled
        >
          <span
            data-test-id="filter-list-empty"
            class="text-danger"
          >
            {{ $t('filters.filterListEmpty') }}
          </span>
        </button>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t: $t } = useI18n()

const props = defineProps({
  availableFilters: {
    type: Array,
    required: true,
  },
  filters: {
    type: Array,
    required: true,
  },
})

const emit = defineEmits(['addFilter'])

const filterList = computed(() => {
  return props.availableFilters.map(f => {
    return { ...f, disabled: !!(props.filters || []).some(filter => filter.ref === f.ref) }
  })
})

function onAddFilter (filter) {
  const add = { ...filter, created: true, params: [] }
  const { params = [] } = filter

  for (const p of params) {
    add.params.push({ ...p, options: { ...p.options } })
  }

  emit('addFilter', add)
}

function filterDropdownCypressId (filter) {
  return filter.toLowerCase().split(' ').join('-')
}
</script>
