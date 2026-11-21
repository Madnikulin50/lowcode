<template>
  <c-form-table-wrapper :labels="{ addButton: t('label.add') }" @add-item="addParam">
    <table v-if="groups.length" class="table table-borderless table-sm">
      <thead>
        <tr>
          <th class="w-25">{{ t('datasources.name') }}</th>
          <th class="w-25">{{ t('datasources.label') }}</th>
          <th class="w-50">{{ t('datasources.expression') }}</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(group, index) in groups" :key="index">
          <td><input v-model="group.name" class="form-control form-control-sm" :placeholder="t('datasources.new.name')" /></td>
          <td><input v-model="group.label" class="form-control form-control-sm" :placeholder="t('datasources.new.label')" /></td>
          <td><input v-model="group.def.raw" class="form-control form-control-sm" :placeholder="t('datasources.expression')" /></td>
          <td class="align-middle"><c-input-confirm show-icon @confirmed="deleteGroup(index)" /></td>
        </tr>
      </tbody>
    </table>
  </c-form-table-wrapper>
</template>
<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({ groupBy: { type: Array, required: true } })
const emit = defineEmits(['update:groupBy'])

const { t } = useI18n()

const groups = computed({
  get: () => props.groupBy || [],
  set: (val) => emit('update.groupBy', val),
})

function addParam() {
  groups.value = [...groups.value, { name: '', label: '', def: { raw: '' } }]
}

function deleteGroup(index) {
  const arr = [...groups.value]
  arr.splice(index, 1)
  groups.value = arr
}
</script>