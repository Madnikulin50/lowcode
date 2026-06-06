<template>
  <div
    v-if="scenario"
  >
    <b-form-group
      :label="$t('scenarios:label')"
      label-class="text-primary"
    >
      <b-form-input
        v-model="scenario.label"
        :placeholder="$t('scenarios:scenario-name')"
      />
    </b-form-group>
    <b-form-group
      :label="$t('scenarios:values')"
      label-class="text-primary"
    >
      <values
        :values.sync="values"
      />
    </b-form-group>
  </div>
</template>

<script>
import Values from 'corteza-webapp-compose/src/components/Common/Values'

export default {
  components: {
    Values,
  },

  props: {
    currentIndex: {
      type: Number,
      default: () => -1,
    },

    scenario: {
      type: Object,
      required: true,
      default: () => ({ values: { list: [] } }),
    },

    module: {
      type: Object,
      required: false,
      default: () => {},
    },
  },

  data () {

  },

  computed: {
    values () {
      return this.scenario.values
    },
  },

  watch: {
    currentIndex: {
      immediate: true,
      handler () {
        // Select first defined filter on switch
        const { filters = {} } = this.scenario
        const definedFilters = Object.keys(filters)

        this.currentDatasourceName = definedFilters.length ? definedFilters[0] : ''
      },
    },

    currentDatasourceName: {
      immediate: true,
      handler (name) {
        if (name && !this.scenario.filters[name]) {
          this.scenario.filters[name] = {}
        }

        this.getSourceColumns()
      },
    },
  },

  methods: {
    async getSourceColumns () {
      this.columns = []
    },
  },
}
</script>

<style>

</style>
