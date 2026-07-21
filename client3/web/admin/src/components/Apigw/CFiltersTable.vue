<template>
  <div>
    <table class="table table-hover mb-0">
      <thead class="table-light">
        <tr>
          <th scope="col" />
          <th scope="col">{{ $t('filters.list.filters') }}</th>
          <th scope="col">{{ $t('filters.list.status') }}</th>
          <th scope="col" />
        </tr>
      </thead>

      <draggable
            item-key="id"
        v-if="!fetching"
        v-model="sortableFilters"
        :options="{ handle: '.handle' }"
        tag="tbody"
      >
        <tr
          v-for="(filter, index) in sortableFilters"
          :key="index"
          class="pointer"
          @click="onRowClick(filter, index)"
        >
          <td
            class="handle align-middle grab"
            style="width: 1%"
            @click.stop
          >
            <font-awesome-icon
              :icon="['fas', 'bars']"
              class="text-secondary"
            />
          </td>
          <td class="align-middle">
            {{ filter.label }}
          </td>
          <td class="align-middle">
            {{ $t(`filters.${filter.enabled ? 'enabled' : 'disabled'}`) }}
          </td>
          <td class="text-end align-middle">
            <c-input-confirm
              show-icon
              class="ms-1"
              @confirmed="onRemoveFilter(filter)"
              @click.stop
            />
          </td>
        </tr>
      </draggable>
    </table>

    <div class="d-flex flex-column align-items-center justify-content-center h-100 overflow-hidden">
      <div
        v-if="fetching"
        class="spinner-border my-4"
        role="status"
      >
        <span class="visually-hidden">{{ $t('label.loading') }}</span>
      </div>

      <p
        v-else-if="!sortableFilters.length"
        data-test-id="no-filters"
        class="my-4"
      >
        {{ $t('filters.list.noFilters') }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import draggable from 'vuedraggable'

const props = defineProps({
  filters: {
    type: Array,
    required: true,
  },
  step: {
    type: Number,
    default: () => 0,
  },
  fetching: {
    type: Boolean,
    value: false,
  },
})

const emit = defineEmits(['sortFilters', 'filterSelect', 'removeFilter'])
const selectedRow = computed(() => 0)
const selectedFilter = computed(() => ({}))

const sortableFilters = computed({
  get () {
    return props.filters
  },

  set (v) {
    emit('sortFilters', v)
  },
})

function onAddFilter (filter) {
  if (!props.filters.find(f => f.ref === filter.ref)) {
    props.filters.push({ ...filter })
  }

  if (props.filters.length === 1) {
    emit('filterSelect', filter)
  }
}

function onRemoveFilter (filter) {
  emit('removeFilter', filter)
}

function onRowClick (filter, index) {
  emit('filterSelect', filter)
}
</script>
