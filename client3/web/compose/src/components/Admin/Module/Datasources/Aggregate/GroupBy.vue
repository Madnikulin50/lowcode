<template>
  <c-form-table-wrapper
    :labels="{
      addButton: $t('label.add')
    }"
    @add-item="addParam"
  >
    <table
      v-if="groups.length"
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
          v-for="(group, index) in groups"
          :key="index"
        >
          <td>
            <input
              v-model="group.name"
              class="form-control form-control-sm"
              :placeholder="$t('datasources.new.name')"
            >
          </td>

          <td>
            <input
              v-model="group.label"
              class="form-control form-control-sm"
              :placeholder="$t('datasources.new.label')"
            >
          </td>

          <td>
            <input
              v-model="group.def.raw"
              class="form-control form-control-sm"
              :placeholder="$t('datasources.expression')"
            >
          </td>

          <td class="align-middle">
            <c-input-confirm
              show-icon
              @confirmed="deleteGroup(index)"
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
  groupBy: {
    type: Array,
    required: true,
  },
})

const emit = defineEmits(['update:groupBy'])

const groups = computed({
  get () {
    return props.groupBy || []
  },
  set (groupBy) {
    emit('update.groupBy', groupBy)
  },
})

function addParam () {
  groups.value.push({
    name: '',
    label: '',
    def: { raw: '' },
  })
}

function deleteGroup (index) {
  groups.value.splice(index, 1)
}
</script>

<style lang="scss" scoped>
.table td.fit,
.table th.fit {
  white-space: nowrap;
  width: 1%;
}

.btn-add-group {
  &:hover, &:active {
    background-color: var(--primary) !important;
    color: var(--white) !important;
  }
}

.filter-border {
  background-image: linear-gradient(to left, lightgray, lightgray);
  background-repeat: no-repeat;
  background-size: 100% 1px;
  background-position: center;
}
</style>
