<script>
import { compose, NoID, validator } from 'corteza-lib/js/dist'
import Wrap from './Wrap'

export default {
  i18nOptions: {
    namespaces: 'block',
  },

  components: {
    Wrap,
  },

  props: {
    blockIndex: {
      type: Number,
      default: -1,
    },

    namespace: {
      type: compose.Namespace,
      required: true,
    },

    page: {
      type: compose.Page,
      required: true,
    },

    blocks: {
      type: Array,
      default: () => [],
    },

    block: {
      type: compose.PageBlock,
      required: true,
    },

    module: {
      type: compose.Module,
      required: false,
      default: undefined,
    },

    record: {
      type: compose.Record,
      required: false,
      default: undefined,
    },

    mode: {
      type: String,
      required: false,
      default: '',
    },

    editable: {
      type: Boolean,
      required: false,
      default: false,
    },

    resizing: {
      type: Boolean,
      required: false,
      default: false,
    },

    magnified: {
      type: Boolean,
      required: false,
      default: false,
    },

    unsavedBlocks: {
      type: Set,
      default: () => new Set(),
    },

    loadingRecord: {
      type: Boolean,
      required: false,
      default: false,
    },

    errors: {
      type: validator.Validated,
      required: false,
      default: () => new validator.Validated(),
    },
  },

  data () {
    return {
      processing: false,
      refreshInterval: null,
      key: 0,
    }
  },

  computed: {
    options: {
      get () {
        return this.block.options
      },
      set (options) {
        this.block.options = options
      },
    },

    isProcessing () {
      return this.processing || this.loadingRecord
    },

    autoRefreshEnabled () {
      return this.options.refreshRate >= 5 && ['page', 'page.record'].includes(this.$route.name)
    },

    // detect when a page block is opened in a modal through magnification or record open type
    inModal () {
      const { recordPageID, magnifiedBlockID } = this.$route.query

      return !!recordPageID || !!magnifiedBlockID
    },

    isRecordPage () {
      return this.page && this.page.moduleID !== NoID
    },

    errorID () {
      const { recordID = NoID } = this.record || {}
      return recordID === NoID ? 'parent:0' : recordID
    },

    themeSettings () {
      return this.$Settings.get('ui.studio.themes', [])
    },
  },

  beforeDestroy () {
    this.setBaseDefaultValues()
  },

  methods: {
    /**
     * Returns errors, filtered for a specific field
     * @param name
     * @returns {validator.Validated} filtered validation results
     */
    fieldErrors (name) {
      if (!this.errors) {
        this.$emit('errors', { errors: undefined, id: `${this.errorID}:${name}` })
        return new validator.Validated()
      }

      const errors = this.errors.filterByMeta('field', name).filterByMeta('id', this.errorID)

      if (errors.set.length > 0) {
        this.$emit('errors', { errors, id: `${this.errorID}:${name}` })
      } else {
        this.$emit('errors', { errors: undefined, id: `${this.errorID}:${name}` })
      }

      return errors
    },

    genStyle (s = {}, options = { forLabel: false, addStyle: {} }) {
      const d = {
        fill: options.forLabel ? (s.labelColor || s.color) : s.color,
        backgroundColor: s.backgroundColor,
        fontSize: s.fontSize ? s.fontSize + 'px' : undefined,
        color: options.forLabel ? (s.labelColor || s.color) : s.color,
      }
      for (const v of Object.keys(options.addStyle)) {
        if (d[v] === undefined) {
          d[v] = options.addStyle[v]
        }
      }
      for (const v of Object.keys(d)) {
        if (d[v] === undefined) {
          delete d[v]
        }
      }

      return d
    },

    getColor (value) {
      if (value[0] === '#') {
        return value
      }
      const themes = this.themeSettings
        .filter((theme) => theme.id !== 'general') // remove general theme
        .map((theme) => {
          return {
            id: theme.id,
            values: JSON.parse(theme.values),
          }
        })

      return themes[0].values[value] || value
    },

    /**
     *
     * @param {*} refreshFunction
     * If reloading data does not refresh the page block
     * You should attach :key="key" to it and increment it in the refreshFunction
     * @param params
     */
    refreshBlock (refreshFunction, ...params) {
      if (!this.autoRefreshEnabled || this.refreshInterval) return

      this.refreshInterval = setInterval(() => {
        refreshFunction(...params)
      }, this.options.refreshRate * 1000)
    },

    setBaseDefaultValues () {
      if (this.refreshInterval) {
        clearInterval(this.refreshInterval)
        this.refreshInterval = null
      }

      this.processing = false
      this.key = 0
    },
  },
}
</script>
