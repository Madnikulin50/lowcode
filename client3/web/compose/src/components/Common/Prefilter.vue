<template>
  <c-form-table-wrapper hide-add-button>
    <table
      v-if="render && filter.ref"
      class="table table-sm table-borderless"
    >
      <template
        v-for="(group, groupIndex) in filter.args"
        :key="groupIndex"
      >
        <tr
          v-for="(arg, argIndex) in group.args[0].args"
          :key="`${groupIndex}-${argIndex}`"
        >
          <td
            class="fit text-center align-middle ps-0"
            style="white-space: nowrap; width: 1%;"
          >
            <select
              v-if="argIndex === 1"
              v-model="group.args[0].ref"
              class="form-select form-control form-select-sm w-auto"
              @change="reRender()"
            >
              <option
                v-for="cond in conditions"
                :key="cond.value"
                :value="cond.value"
              >
                {{ cond.text }}
              </option>
            </select>

            <span
              v-else
              class="px-3"
              style="min-width: 60px;"
            >
              {{ argIndex === 0 ? 'Where' : `${group.args[0].ref[0].toUpperCase() + group.args[0].ref.slice(1).toLowerCase()}` }}
            </span>
          </td>

          <template v-if="Object.keys(arg).includes('raw')">
            <td>
              <div class="input-group input-group-sm">
                <button
                  type="button"
                  class="btn btn-outline-primary d-flex justify-content-center align-items-center"
                  @click="toggleMode(groupIndex, argIndex)"
                >
                  <font-awesome-icon
                    :icon="['fas', 'filter']"
                    size="sm"
                  />
                </button>

                <input
                  v-model="arg.raw"
                  class="form-control form-control-sm"
                  :placeholder="$t('builder.filter-expression')"
                />
              </div>
            </td>
          </template>

          <template v-else-if="group">
            <td>
              <div class="input-group input-group-sm">
                <button
                  type="button"
                  class="btn btn-primary d-flex justify-content-center align-items-center"
                  @click="toggleMode(groupIndex, argIndex)"
                >
                  <font-awesome-icon
                    :icon="['fas', 'filter']"
                    size="sm"
                  />
                </button>

                <template v-if="group.args[0].args[argIndex].args[0].args[0].value && getColumnData(group.args[0].args[argIndex].args[0].args[1]).multivalue">
                  <column-selector
                    :columns="columns"
                    :value="group.args[0].args[argIndex].args[0].args[1].symbol"
                    @input="setType(groupIndex, argIndex, $event, group.args[0].args[argIndex].args[0].args[0].value['@value'])"
                  />

                  <select
                    v-model="group.args[0].args[argIndex].args[0].ref"
                    class="form-select form-control form-select-sm"
                    style="max-width: 120px; border-left: 0;"
                    @change="reRender"
                  >
                    <option
                      v-for="op in getOperators(getColumnData(group.args[0].args[argIndex].args[0].args[1]))"
                      :key="op.value"
                      :value="op.value"
                    >
                      {{ op.text }}
                    </option>
                  </select>

                  <input
                    v-model="group.args[0].args[argIndex].args[0].args[0].value['@value']"
                    class="form-control form-control-sm"
                    :placeholder="$t('builder.value')"
                  />
                </template>

                <template v-else>
                  <column-selector
                    :value="group.args[0].args[argIndex].args[0].args[0].symbol"
                    :columns="columns"
                    @input="setType(groupIndex, argIndex, $event, group.args[0].args[argIndex].args[0].args[1].value['@value'])"
                  />

                  <select
                    v-model="group.args[0].args[argIndex].args[0].ref"
                    class="form-select form-control form-select-sm"
                    style="max-width: 120px; border-left: 0;"
                    @change="reRender"
                  >
                    <option
                      v-for="op in getOperators(getColumnData(group.args[0].args[argIndex].args[0].args[0]))"
                      :key="op.value"
                      :value="op.value"
                    >
                      {{ op.text }}
                    </option>
                  </select>

                  <input
                    v-model="group.args[0].args[argIndex].args[0].args[1].value['@value']"
                    class="form-control form-control-sm"
                    :placeholder="$t('builder.value')"
                  />
                </template>
              </div>
            </td>
          </template>

          <td
            class="fit text-center align-middle ps-2 pe-0"
            style="white-space: nowrap; width: 1%;"
          >
            <c-input-confirm
              show-icon
              @confirmed="deleteFilter(groupIndex, argIndex)"
            />
          </td>
        </tr>

        <tr>
          <td
            class="fit align-middle ps-0"
            :class="{ 'text-center': group.args[0].args && group.args[0].args.length }"
          >
            <button
              type="button"
              class="btn btn-primary btn-sm mt-1"
              @click="addFilter(groupIndex)"
            >
              <font-awesome-icon
                :icon="['fas', 'plus']"
                size="sm"
                class="me-1"
              />
              {{ $t('label.add') }}
            </button>
          </td>
        </tr>

        <tr
          v-if="group.args[0].args && group.args[0].args.length"
        >
          <td
            colspan="100%"
            class="p-0 filter-border text-center"
            :class="{ 'pb-1': groupIndex < filter.args.length - 1 }"
          >
            <select
              v-if="groupIndex < filter.args.length - 1"
              v-model="filter.ref"
              class="form-select form-control form-select-sm w-auto"
              @change="reRender()"
            >
              <option
                v-for="cond in conditions"
                :key="cond.value"
                :value="cond.value"
              >
                {{ cond.text }}
              </option>
            </select>

            <button
              v-else
              type="button"
              class="btn btn-outline-primary btn-add-group bg-white py-2 px-3"
              @click="addGroup()"
            >
              <font-awesome-icon
                :icon="['fas', 'plus']"
                class="h6 mb-0"
              />
            </button>
          </td>
        </tr>
      </template>
    </table>

    <div v-else>
      <button
        type="button"
        class="btn btn-primary btn-sm"
        @click="initFilter()"
      >
        <font-awesome-icon
          :icon="['fas', 'plus']"
          size="sm"
          class="me-1"
        />
        {{ $t('label.add') }}
      </button>
    </div>
  </c-form-table-wrapper>
</template>

<script setup>
import { ref, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import ColumnSelector from 'corteza-webapp-compose/src/components/Common/ColumnSelector.vue'

const { t } = useI18n()

const props = defineProps({
  filter: {
    type: Object,
    required: true,
  },
  columns: {
    type: Array,
    default: () => [],
  },
})

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

function initFilter () {
  props.filter.ref = 'and'
  props.filter.args = []
  addGroup()
}

function addGroup () {
  if (props.filter.args) {
    props.filter.args.push({
      ref: 'group',
      args: [{
        ref: 'or',
        args: [{
          ref: 'group',
          args: [{
            ref: 'eq',
            args: [
              { symbol: '' },
              { value: { '@type': '', '@value': '' } },
            ],
          }],
        }],
      }],
    })
  }
  reRender()
}

function addFilter (groupIndex) {
  if (!props.filter.args[groupIndex].args[0].args) {
    props.filter.args[groupIndex].args[0].args = []
  }
  props.filter.args[groupIndex].args[0].args.push({
    ref: 'group',
    args: [{
      ref: 'eq',
      args: [
        { symbol: '' },
        { value: { '@type': '', '@value': '' } },
      ],
    }],
  })
  reRender()
}

function deleteFilter (groupIndex, argIndex) {
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

function toggleMode (groupIndex, argIndex) {
  const { args } = props.filter.args[groupIndex].args[0]
  if (args[argIndex]) {
    if (Object.keys(args[argIndex]).includes('raw')) {
      args[argIndex].args = [{
        ref: 'eq',
        args: [
          { symbol: '' },
          { value: { '@type': '', '@value': '' } },
        ],
      }]
      delete args[argIndex].raw
    } else {
      args[argIndex].raw = ''
      delete args[argIndex].args
    }
    reRender()
  }
}

function setType (groupIndex, argIndex, symbol, value) {
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

function reRender () {
  render.value = false
  nextTick().then(() => {
    render.value = true
  })
}

function getColumnData (group) {
  return props.columns.find(({ name }) => name === group.symbol)
}

function getOperators (column) {
  return column ? operators.filter(value => value.isMulti === column.multivalue) : operators
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
