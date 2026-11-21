<template>
  <c-form-table-wrapper :labels="{ addButton: t('label.add') }" @add-item="addColumn">
    <table v-if="columns.length" class="table table-borderless table-sm">
      <thead>
        <tr>
          <th class="w-25">{{ t('datasources.name') }}</th>
          <th class="w-25">{{ t('datasources.label') }}</th>
          <th class="w-50">{{ t('datasources.expression') }}</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(col, index) in columns" :key="index">
          <td><input v-model="col.name" class="form-control form-control-sm" :placeholder="t('datasources.new.name')" /></td>
          <td><input v-model="col.label" class="form-control form-control-sm" :placeholder="t('datasources.new.label')" /></td>
          <td><input v-model="col.def.raw" class="form-control form-control-sm" :placeholder="t('datasources.expression')" /></td>
          <td class="align-middle"><c-input-confirm show-icon @confirmed="deleteColumn(index)" /></td>
        </tr>
      </tbody>
    </table>
  </c-form-table-wrapper>
</template>
<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({ aggregate: { type: Array, required: true } })
const emit = defineEmits(['update:aggregate'])

const { t } = useI18n()

const columns = computed({
  get: () => props.aggregate || [],
  set: (val) => emit('update.aggregate', val),
})

function addColumn() {
  columns.value = [...columns.value, { name: '', label: '', def: { raw: '' } }]
}

function deleteColumn(index) {
  const arr = [...columns.value]
  arr.splice(index, 1)
  columns.value = arr
}
</script>