import { ref, computed, onBeforeUnmount } from 'vue'
import { compose, NoID, validator } from 'corteza-lib/js/dist'
import { useRoute } from 'vue-router'
import { useSettings } from 'corteza-lib/vue/dist'

export function usePageBlockBase (props, emit) {
  const route = useRoute()
  const { $Settings } = useSettings()

  const processing = ref(false)
  const refreshInterval = ref(null)
  const key = ref(0)

  const options = computed({
    get: () => props.block.options || {},
    set: (opts) => { props.block.options = opts },
  })

  const isProcessing = computed(() => processing.value || props.loadingRecord)

  const autoRefreshEnabled = computed(() => {
    return options.value.refreshRate >= 5 && ['page', 'page.record'].includes(route.name)
  })

  const inModal = computed(() => {
    const { recordPageID, magnifiedBlockID } = route.query
    return !!recordPageID || !!magnifiedBlockID
  })

  const isRecordPage = computed(() => props.page && props.page.moduleID !== NoID)

  const errorID = computed(() => {
    const { recordID = NoID } = props.record || {}
    return recordID === NoID ? 'parent:0' : recordID
  })

  const themeSettings = computed(() => $Settings.get('ui.studio.themes', []))

  function browserLocale () {
    return navigator.language || navigator.languages?.[0] || 'en-US'
  }

  function emitErrors (payload) {
    if (typeof emit === 'function') emit('errors', payload)
  }

  function fieldErrors (name) {
    const id = `${errorID.value}:${name}`
    if (!props.errors || typeof props.errors.filterByMeta !== 'function') {
      emitErrors({ errors: undefined, id })
      return new validator.Validated()
    }
    const errors = props.errors.filterByMeta('field', name).filterByMeta('id', errorID.value)
    emitErrors({ errors: errors.set.length > 0 ? errors : undefined, id })
    return errors
  }

  function genStyle (s = {}, options = { forLabel: false, addStyle: {} }) {
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
  }

  function getColor (value) {
    if (!value) return undefined
    if (value[0] === '#') return value
    const themes = themeSettings.value
      .filter((theme) => theme.id !== 'general')
      .map((theme) => ({ id: theme.id, values: typeof theme.values === 'string' ? JSON.parse(theme.values) : theme.values }))
    return themes[0]?.values?.[value] || value
  }

  function refreshBlock (refreshFunction, ...params) {
    if (!autoRefreshEnabled.value || refreshInterval.value) return
    refreshInterval.value = setInterval(() => {
      refreshFunction(...params)
    }, options.value.refreshRate * 1000)
  }

  function setBaseDefaultValues () {
    if (refreshInterval.value) {
      clearInterval(refreshInterval.value)
      refreshInterval.value = null
    }
    processing.value = false
    key.value = 0
  }

  onBeforeUnmount(() => {
    setBaseDefaultValues()
  })

  return {
    processing,
    refreshInterval,
    key,
    options,
    isProcessing,
    autoRefreshEnabled,
    inModal,
    isRecordPage,
    errorID,
    themeSettings,
    browserLocale,
    fieldErrors,
    genStyle,
    getColor,
    refreshBlock,
    setBaseDefaultValues,
  }
}
