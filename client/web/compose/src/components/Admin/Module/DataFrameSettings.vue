<template>
  <div
    v-if="module"
  >
    <configurator
      v-if="module"
      :items="items"
      :current-index="datasources.currentIndex"
      draggable
      @select="setCurrentDatasource"
      @add="openDatasourceSelector()"
      @delete="deleteCurrentDataSource"
    >
      <template #label="{ item: { step } }">
        <span
          class="d-inline-block text-truncate"
        >
          {{ datasourceLabel(step, datasources.currentIndex) }}
        </span>
      </template>
      <template #configurator>
        <component
          :is="getDatasourceComponent(items[datasources.currentIndex])"
          v-if="currentDatasourceStep"
          :index="datasources.currentIndex"
          :datasources="items"
          :step.sync="currentDatasourceStep"
          :creating="items[datasources.currentIndex].meta.creating"
        />
      </template>
    </configurator>

    <b-modal
      v-model="datasources.showSelector"
      size="lg"
      scrollable
      hide-footer
      :title="$t('add.datasource')"
      body-class="px-0 py-3"
      no-fade
    >
      <selector
        :items="datasources.types"
        display-mode="text"
        @select="addDatasource"
      />
    </b-modal>
  </div>
</template>

<script>
import { compose, NoID, reporter } from 'corteza-lib/js/dist'
import datasources from 'corteza-webapp-compose/src/components/Admin/Module/Datasources/loader'
import Configurator from 'corteza-webapp-compose/src/components/Common/Configurator'
import Selector from 'corteza-webapp-compose/src/components/Common/Selector'
import * as displayElementThumbnails from 'corteza-webapp-compose/src/assets/DisplayElements'
import { cloneDeep } from 'lodash'

const PrimaryConnType = 'corteza::system:primary-dal-connection'

export default {
  name: 'data-source-settings',
  i18nOptions: {
    namespaces: 'module',
    keyPrefix: 'edit.config.datasource',
  },

  components: {
    Selector,
    Configurator,
  },

  props: {
    module: {
      type: compose.Module,
      required: true,
    },
  },

  data () {
    return {
      processing: false,
      connections: [],

      displayElements: {
        showSelector: false,

        currentIndex: undefined,

        types: [
          {
            label: this.$t('display-elements.types.text'),
            kind: 'Text',
            value: displayElementThumbnails.Text,
          },
          {
            label: this.$t('display-elements.types.metric'),
            kind: 'Metric',
            value: displayElementThumbnails.Metric,
          },
          {
            label: this.$t('display-elements.types.table'),
            kind: 'Table',
            value: displayElementThumbnails.Table,
          },
          {
            label: this.$t('display-elements.types.chart'),
            kind: 'Chart',
            value: displayElementThumbnails.Chart,
          },
        ],
      },

      datasources: {
        showSelector: false,
        showConfigurator: false,

        processing: false,
        currentIndex: undefined,

        types: [
          {
            label: this.$t('types.load.label'),
            kind: 'Load',
            value: this.$t('types.load.data-from-resource'),
          },
          {
            label: this.$t('types.link.label'),
            kind: 'Link',
            value: this.$t('types.link.load-datasources'),
          },
          {
            label: this.$t('types.join.label'),
            kind: 'Join',
            value: this.$t('types.join.load-datasources'),
          },
          {
            label: this.$t('types.aggregate.label'),
            kind: 'Aggregate',
            value: this.$t('types.aggregate.load-datasource'),
          },
        ],
      },

      scenarios: {
        showConfigurator: false,

        currentIndex: undefined,

        selected: undefined,
      },

      editor: undefined,
    }
  },

  computed: {
    items: {
      get () {
        const dataSource = this.module.config.dataSource ?? {}
        return dataSource.items ?? []
      },

      set (value) {
        this.module.config.dataSource = this.module.config.dataSource ?? {}
        this.module.config.dataSource = { items: value }
      },
    },
    currentDatasourceStep: {
      get () {
        return this.datasources.currentIndex !== undefined ? this.items[this.datasources.currentIndex].step : undefined
      },

      set (step) {
        if (this.datasources.currentIndex !== undefined) {
          this.items[this.datasources.currentIndex].step = step
        }
      },
    },
  },

  watch: {
  },

  mounted () {
    this.datasources.showConfigurator = true
    const dataSource = this.module.config.dataSource ?? {}
    this.items = cloneDeep(dataSource.items ?? []).map(ds => {
      ds.meta.creating = false
      return ds
    })
    this.datasources.currentIndex = this.items.length ? 0 : undefined
  },

  beforeDestroy () {
    this.setDefaultValues()
  },

  methods: {
    async fetchConnections () {
      this.processing = true
      return this.$SystemAPI.dalConnectionList()
        .then(({ set = [] }) => {
          this.connections = set

          const { connectionID } = this.module.config.dal || {}
          if (!connectionID || connectionID === NoID) {
            const primaryConnectionID = (this.connections.find(c => c.type === PrimaryConnType) || { connectionID: NoID }).connectionID
            this.module.config.dal.connectionID = primaryConnectionID
          }
        })
        .catch(this.toastErrorHandler(this.$t('connections.fetch-failed')))
        .finally(() => {
          this.processing = false
        })
    },

    getOptionKey ({ connectionID }) {
      return connectionID
    },

    setDefaultValues () {
      this.processing = false
      this.connections = []
      this.moduleFields = []
      this.moduleFieldEncoding = []
      this.selectedGroup = ''
      this.systemFields = []
      this.systemFieldEncoding = []
      this.optionsGroups = []
    },
    getDatasourceComponent ({ step }) {
      let datasource

      if (step) {
        for (const s in step) {
          datasource = datasources(s)
          if (datasource) {
            break
          }
        }
      }

      return datasource
    },

    datasourceLabel (datasource, currentIndex) {
      for (const v of Object.values(datasource)) {
        if (v && v.name) {
          return v.name
        }
      }

      return `${this.$t('datasources:source')} ${currentIndex}`
    },

    openDatasourceSelector () {
      this.datasources.showSelector = true
      this.datasources.currentIndex = this.items.length ? 0 : undefined
    },

    setCurrentDatasource (index) {
      this.datasources.currentIndex = index
    },

    deleteCurrentDataSource () {
      this.items.splice(this.datasources.currentIndex, 1)
      this.datasources.currentIndex = this.items.length ? 0 : undefined
    },

    addDatasource (kind = '') {
      if (kind) {
        let step

        switch (kind) {
          case 'Aggregate':
            step = reporter.StepFactory({
              aggregate: {
                name: 'Aggregate',
                keys: [],
                columns: [],
                filter: {},
                sort: '',
              },
            })
            break

          case 'Link':
            step = reporter.StepFactory({
              link: {
                name: 'Link',
                foreignColumn: '',
                foreignSource: '',
                localColumn: '',
                localSource: '',
              },
            })
            break

          case 'Join':
            step = reporter.StepFactory({
              join: {
                name: 'Join',
                foreignColumn: '',
                foreignSource: '',
                localColumn: '',
                localSource: '',
              },
            })
            break

          default:
            step = reporter.StepFactory({
              load: {
                name: 'Load',
                source: 'composeRecords',
                definition: {},
                filter: {},
                sort: '',
              },
            })
        }
        const items = this.items
        items.push({
          step,
          meta: {},
        })
        this.items = items
      }

      // Select newly added datasource in configurator
      this.datasources.currentIndex = this.items.length - 1

      // Close selector, open configurator
      this.datasources.showSelector = false
      this.datasources.showConfigurator = true
    },
  },
}
</script>
