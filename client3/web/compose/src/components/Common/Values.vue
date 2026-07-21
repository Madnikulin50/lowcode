<template>
  <c-form-table-wrapper
    :labels="{ addButton: $t('values.addValue') }"
    @add-item="addValue"
  >
    <tr
      v-for="(arg, argIndex) in values.list"
      :key="`$${argIndex}`"
    >
      <td>
        <input
          v-model="arg.symbol"
          class="form-control form-control-sm"
          :placeholder="$t('builder.symbol')"
        />
      </td>
      <td>
        <input
          v-model="arg.value"
          class="form-control form-control-sm"
          :placeholder="$t('builder.value')"
        />
      </td>
      <td
        class="fit text-center align-middle ps-2 pe-0"
        style="white-space: nowrap; width: 1%;"
      >
        <c-input-confirm
          show-icon
          @confirmed="deleteValue(argIndex)"
        />
      </td>
    </tr>
  </c-form-table-wrapper>
</template>

<script setup>
import { ref, nextTick } from 'vue'

const props = defineProps({
  values: {
    type: Object,
    default: () => {
      return { values: [] }
    },
  },
})

const render = ref(true)

function addValue () {
  props.values.list.push({ symbol: '', value: '', type: 'String' })
  reRender()
}

function deleteValue (argIndex) {
  props.values.list.splice(argIndex, 1)
  reRender()
}

function reRender () {
  render.value = false
  nextTick().then(() => {
    render.value = true
  })
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

<style lang="scss">
.prefilter .column-selector {
  .vs__dropdown-toggle {
    border-right: 0;
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
  }
}
</style>
