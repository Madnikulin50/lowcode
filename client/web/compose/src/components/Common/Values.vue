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
        <b-form-input
          v-model="arg.symbol"
          :placeholder="$t('builder:symbol')"
        />
      </td>
      <td>
        <b-form-input
          v-model="arg.value"
          :placeholder="$t('builder:value')"
        />
      </td>
      <td
        class="fit text-center align-middle pl-2 pr-0"
      >
        <c-input-confirm
          show-icon
          @confirmed="deleteValue(argIndex)"
        />
      </td>
    </tr>
  </c-form-table-wrapper>
</template>

<script>
export default {

  props: {
    values: {
      type: Object,
      default: () => {
        return { values: [] }
      },
    },
  },

  data () {
    return {
      render: true,
    }
  },

  methods: {
    addValue () {
      this.values.list.push({ symbol: '', value: '', type: 'String' })
      this.reRender()
    },

    deleteValue: function (argIndex) {
      this.values.list.splice(argIndex, 1)
      this.reRender()
    },

    reRender () {
      this.render = false
      this.$nextTick().then(() => {
        this.render = true
      })
    },
  },
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
