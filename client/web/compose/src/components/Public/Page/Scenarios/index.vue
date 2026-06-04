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
      :label="$t('scenarios:datasource')"
      label-class="text-primary"
    >
      <b-form-select
        v-model="currentDatasourceName"
        :options="datasourceOptions"
      />
    </b-form-group>

    <b-form-group
      v-if="currentDatasourceName && scenario.filters[currentDatasourceName]"
      :label="$t('scenarios:prefilter')"
      label-class="text-primary"
    >
      <prefilter
        :filter="scenario.filters[currentDatasourceName]"
        :columns="columns"
      />
    </b-form-group>
  </div>
</template>

<script>
import Prefilter from 'corteza-webapp-compose/src/components/Common/Prefilter'

export default {
  components: {
    Prefilter,
  },

  props: {
    currentIndex: {
      type: Number,
      default: () => -1,
    },

    scenario: {
      type: Object,
      required: true,
      default: () => ({}),
    },

    module: {
      type: Object,
      required: false,
      default: () => {},
    },
  },

  data () {
    return {
      currentDatasourceName: '',
      columns: [],
    }
  },

  computed: {
    datasourceOptions () {
      const options = [{ value: '', text: this.$t('general:label.none') }]
      return options
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
