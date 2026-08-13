import { computed, inject } from 'vue'
import { useRoute } from 'vue-router'

export function useViewerBase(p) {
  const r = useRoute()
  const s = inject('$Settings')
  const v = computed(() => {
    if (p.field?.isSystem) return p.record?.[p.field?.name]
    return p.record ? p.record.values?.[p.field?.name] : undefined
  })
  const f = computed(() => {
    if (p.field?.isMulti) return v.value ? v.value.join(p.field?.options?.multiDelimiter) : ''
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
    if (val[0] === '#') return val
    const themes = ts.value.filter(t => t.id !== 'general').map(t => ({ id: t.id, values: JSON.parse(t.values) }))
    return themes[0]?.values[val] || val
  }
  return { value: v, formatted: f, classes: c, options: o, inModal: im, themeSettings: ts, getColor: gc }
}
