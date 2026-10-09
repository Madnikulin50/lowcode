<template>
  <select v-if="filter.ref" class="form-select form-select-sm w-auto mb-2">
    <option v-for="cond in conditions" :key="cond.value" :value="cond.value" :selected="filter.ref === cond.value">
      {{ cond.text }}
    </option>
  </select>
  <button v-else class="btn btn-primary btn-sm" @click="initFilter">
    <font-awesome-icon :icon="['fas', 'plus']" size="sm" class="me-1" />
    {{ t('label.add') }}
  </button>
  <div v-if="render && filter.ref" class="table-responsive">
    <table class="table table-borderless table-sm">
      <tbody>
        <template v-for="(group, groupIndex) in filter.args" :key="groupIndex">
          <tr v-for="(arg, argIndex) in group.args[0].args" :key="`${groupIndex}-${argIndex}`">
            <td class="text-nowrap text-center align-middle ps-0">
              <select
                v-if="argIndex === 1"
                v-model="group.args[0].ref"
                class="form-select form-select-sm w-auto"
                @change="reRender"
              >
                <option v-for="cond in conditions" :key="cond.value" :value="cond.value">{{ cond.text }}</option>
              </select>
              <span v-else class="px-3" style="min-width:60px">
                {{ argIndex === 0 ? 'Where' : `${group.args[0].ref[0].toUpperCase()}${group.args[0].ref.slice(1).toLowerCase()}` }}
              </span>
            </td>
            <td>
              <div class="input-group input-group-sm">
                <button class="btn btn-outline-primary" @click="toggleMode(groupIndex, argIndex)">
                  <font-awesome-icon :icon="['fas', 'filter']" size="sm" />
                </button>
                <ColumnSelector
                  v-if="arg.args?.[0]?.args?.[0]?.value && getColumnData(arg.args[0].args[1])?.multivalue"
                  :value="arg.args[0].args[1].symbol"
                  :columns="columns"
                  @input="setType(groupIndex, argIndex, $event, arg.args[0].args[0].value['@value'])"
                />
                <select
                  v-model="group.args[0].args[argIndex].args[0].ref"
                  class="form-select form-select-sm border-start-0"
                  style="max-width:120px"
                >
                  <option v-for="op in getOperators(getColumnData(arg.args[0].args[1]))" :key="op.value" :value="op.value">{{ op.text }}</option>
                </select>
                <input
                  v-model="arg.args[0].args[0].value['@value']"
                  class="form-control form-control-sm"
                  :placeholder="t('builder.value')"
                />
              </div>
            </td>
            <td class="text-nowrap text-center align-middle ps-2 pe-0">
              <c-input-confirm show-icon @confirmed="deleteFilter(groupIndex, argIndex)" />
            </td>
          </tr>
          <tr>
            <td class="text-nowrap align-middle ps-0" :class="{ 'text-center': group.args[0].args?.length }">
              <button class="btn btn-primary btn-sm mt-1" @click="addFilter(groupIndex)">
                <font-awesome-icon :icon="['fas', 'plus']" size="sm" class="me-1" />
                {{ t('label.add') }}
              </button>
            </td>
          </tr>
          <tr v-if="group.args[0].args?.length">
            <td colspan="100%" class="p-0 text-center position-relative" style="background-image:linear-gradient(to left,lightgray,lightgray);background-repeat:no-repeat;background-size:100% 1px;background-position:center">
              <select
                v-if="groupIndex < filter.args.length - 1"
                v-model="filter.ref"
                class="form-select form-select-sm w-auto"
                @change="reRender"
              >
                <option v-for="cond in conditions" :key="cond.value" :value="cond.value">{{ cond.text }}</option>
              </select>
            </td>
          </tr>
        </template>
      </tbody>
    </table>
  </div>
  <c-form-table-wrapper v-else hide-add-button />
</template>
<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import ColumnSelector from '../Common/ColumnSelector.vue'

const props = defineProps({
  filter: { type: Object, required: true },
  columns: { type: Array, default: () => [] },
})

const { t } = useI18n()
const render = ref(true)

const conditions = [
  { value: 'and', text: t('label.and') },
  { value: 'or', text: t('label.or') },
]
const operators = [
  { value: 'eq', text: t('builder.filter.operators.equal'), isMulti: false },
  { value: 'ne', text: t('builder.filter.operators.notEqual'), isMulti: false },
  { value: 'lt', text: t('builder.filter.operators.lessThan'), isMulti: false },
  { value: 'le', text: t('builder.filter.operators.lessThanEqualTo'), isMulti: false },
  { value: 'gt', text: t('builder.filter.operators.greaterThan'), isMulti: false },
  { value: 'ge', text: t('builder.filter.operators.greaterThanEqualTo'), isMulti: false },
  { value: 'in', text: t('builder.filter.operators.contains'), isMulti: true },
  { value: 'nin', text: t('builder.filter.operators.notContains'), isMulti: true },
]

function initFilter() {
  props.filter.ref = 'and'
  props.filter.args = []
  addGroup()
}

function addGroup() {
  if (props.filter.args) {
    props.filter.args.push({
      ref: 'group',
      args: [{ ref: 'or', args: [{ ref: 'group', args: [{ ref: 'eq', args: [{ symbol: '' }, { value: { '@type': '', '@value': '' } }] }] }] }],
    })
  }
  reRender()
}

function addFilter(groupIndex) {
  if (!props.filter.args[groupIndex].args[0].args) props.filter.args[groupIndex].args[0].args = []
  props.filter.args[groupIndex].args[0].args.push({ ref: 'group', args: [{ ref: 'eq', args: [{ symbol: '' }, { value: { '@type': '', '@value': '' } }] }] })
  reRender()
}

function deleteFilter(groupIndex, argIndex) {
  const { args } = props.filter.args[groupIndex].args[0]
  if (args) {
    if (props.filter.args.length === 1 && args.length === 1) {
      delete props.filter.ref
      delete props.filter.args
    } else if (args.length === 1) {
      props.filter.args.splice(groupIndex, 1)
    } else {
      args.splice(argIndex, 1)
    }
  }
  reRender()
}

function toggleMode(groupIndex, argIndex) {
  const { args } = props.filter.args[groupIndex].args[0]
  if (args[argIndex]) {
    if ('raw' in args[argIndex]) {
      args[argIndex].args = [{ ref: 'eq', args: [{ symbol: '' }, { value: { '@type': '', '@value': '' } }] }]
      delete args[argIndex].raw
    } else {
      args[argIndex].raw = ''
      delete args[argIndex].args
    }
    reRender()
  }
}

function setType(groupIndex, argIndex, symbol, value) {
  if (!props.filter.args[groupIndex].args[0].args[argIndex]) return
  const { kind = '', multivalue } = props.columns.find(({ name }) => name === symbol) || {}
  if (multivalue) {
    props.filter.args[groupIndex].args[0].args[argIndex].args[0].args[0] = { value: { '@type': kind, '@value': value } }
    props.filter.args[groupIndex].args[0].args[argIndex].args[0].args[1] = { symbol }
    props.filter.args[groupIndex].args[0].args[argIndex].args[0].ref = 'in'
  } else {
    props.filter.args[groupIndex].args[0].args[argIndex].args[0].args[0] = { symbol }
    props.filter.args[groupIndex].args[0].args[argIndex].args[0].args[1] = { value: { '@type': kind, '@value': value } }
    props.filter.args[groupIndex].args[0].args[argIndex].args[0].ref = 'eq'
  }
  reRender()
}

function reRender() {
  render.value = false
  setTimeout(() => { render.value = true }, 0)
}

function getColumnData(group) {
  return props.columns.find(({ name }) => name === group?.symbol)
}

function getOperators(column) {
  return column ? operators.filter(value => value.isMulti === column.multivalue) : operators
}
</script>
<style lang="scss" scoped>
.table td.fit, .table th.fit { white-space: nowrap; width: 1%; }
.btn-add-group:hover, .btn-add-group:active { background-color: var(--primary) !important; color: var(--white) !important; }
.filter-border { background-image: linear-gradient(to left, lightgray, lightgray); background-repeat: no-repeat; background-size: 100% 1px; background-position: center; }
</style>