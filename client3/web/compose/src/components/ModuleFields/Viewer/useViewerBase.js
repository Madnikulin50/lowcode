import { computed, inject } from 'vue'
import { useRoute } from 'vue-router'
import { themeColor } from 'corteza-webapp-compose/src/lib/color.js'

export function useViewerBase(p) {
  const r = useRoute()
  const s = inject('$Settings')
  const v = computed(() => {
    if (p.field?.isSystem) return p.record?.[p.field?.name]
    return p.record ? p.record.values?.[p.field?.name] : undefined
  })
  const f = computed(() => {
    if (p.field?.isMulti) {
      const val = v.value
      if (Array.isArray(val)) return val.join(p.field?.options?.multiDelimiter)
      return val == null ? '' : String(val)
    }
    return v.value
  })
  const c = computed(() => {
    const cls = []
    const { fieldID } = p.field || {}
    const { textStyles = {} } = p.extraOptions || {}
    if (p.field?.isMulti) cls.push('multiline')
    else if (p.includeStyles) {
      if (!textStyles.wrappedFields || !textStyles.wrappedFields.includes(fieldID)) cls.push('text-nowrap')
    }
    return cls
  })
  const o = computed(() => p.field?.options)
  const im = computed(() => {
    const { recordPageID, magnifiedBlockID } = r.query
    return !!recordPageID || !!magnifiedBlockID
  })
  const ts = computed(() => s?.get ? s.get('ui.studio.themes', []) : [])
  function gc (val) {
    return themeColor(val, ts.value)
  }
  return { value: v, formatted: f, classes: c, options: o, inModal: im, themeSettings: ts, getColor: gc }
}
