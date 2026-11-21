<template>
  <c-form-table-wrapper
    :labels="{
      addButton: $t('label.add')
    }"
    @add-item="addColumn"
  >
    <table
      v-if="columns.length"
      class="table table-sm table-borderless"
    >
      <thead>
        <tr>
          <th class="w-25">
            {{ $t('datasources.name') }}
          </th>

          <th class="w-25">
            {{ $t('datasources.label') }}
          </th>

          <th class="w-50">
            {{ $t('datasources.expression') }}
          </th>
        </tr>
      </thead>

      <tbody>
        <tr
          v-for="(column, index) in columns"
          :key="index"
        >
          <td>
            <input
              v-model="column.name"
              class="form-control form-control-sm"
              :placeholder="$t('datasources.new.name')"
            >
          </td>

          <td>
            <input
              v-model="column.label"
              class="form-control form-control-sm"
              :placeholder="$t('datasources.new.label')"
            >
          </td>

          <td>
            <input
              v-model="column.def.raw"
              class="form-control form-control-sm"
              :placeholder="$t('datasources.expression')"
            >
          </td>

          <td class="align-middle">
            <c-input-confirm
              show-icon
              @confirmed="deleteColumn(index)"
            />
          </td>
        </tr>
      </tbody>
    </table>
  </c-form-table-wrapper>
</template>

<script setup lang="js">
import { computed } from 'vue'

const props = defineProps({
  aggregate: {
    type: Array,
    required: true,
  },
})

const emit = defineEmits(['update:aggregate'])

const columns = computed({
  get () {
    return props.aggregate || []
  },
  set (aggregate) {
    emit('update.aggregate', aggregate)
  },
})

function addColumn () {
  columns.value.push({
    name: '',
    label: '',
    def: { raw: '' },
  })
}

function deleteColumn (index) {
  columns.value.splice(index, 1)
}
</script>
