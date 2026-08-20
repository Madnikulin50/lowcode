import { computed, inject } from 'vue'
import { useRoute } from 'vue-router'
import { compose, validator } from 'corteza-lib/js/dist'
import { useUiStore } from '../../../store/ui'
import { themeColor } from 'corteza-webapp-compose/src/lib/color.js'

export function useEditorBase(props, emit) {
  const route = useRoute()
  const uiStore = useUiStore()
  const $Settings = inject('$Settings')

  const themeSettings = computed(() => $Settings?.get ? $Settings.get('ui.studio.themes', []) : [])

  function getColor (val) {
    return themeColor(val, themeSettings.value)
  }

  const formGroupStyleClasses = computed(() => ({
    required: isRequired.value,
    small: false,
    'value-only': props.valueOnly,
    row: !!props.horizontal,
    'g-2': !!props.horizontal,
    'align-items-start': !!props.horizontal,
  }))

  const labelColClass = computed(() => props.horizontal ? 'col-md-6 col-xl-5' : '')
  const contentColClass = computed(() => props.horizontal ? 'col-md-6 col-xl-7' : '')

  const isRequired = computed(() => {
    return props.field.isRequired || uiStore.isFieldRequiredByLayout(props.field.name || props.field.fieldID)
  })

  const state = computed(() => {
    if (!props.errors.valid()) return null
    return props.errors.valid() === true ? null : false
  })

  const value = computed({
    get () {
      if (props.field.isSystem) return props.record[props.field.name]
      return props.record.values[props.field.name]
    },
    set (val) {
      if (props.field.isSystem) {
        props.record[props.field.name] = val
      } else {
        props.record.values[props.field.name] = val
      }
      emit('change', val)
    },
  })

  const showPopover = computed({
    get: () => false,
    set: (v) => emit('update.preventPopoverClose', v),
  })

  const label = computed(() => props.field.label || props.field.name)

  const description = computed(() => {
    if (props.valueOnly) return ''
    const { view, edit } = props.field.options.description
    return edit || view
  })

  const hint = computed(() => {
    if (props.valueOnly) return ''
    const { view, edit } = props.field.options.hint
    return edit || view
  })

  const inModal = computed(() => {
    const { recordPageID, magnifiedBlockID } = route.query
    return !!recordPageID || !!magnifiedBlockID
  })

  function getFieldCypressId (field) {
    return `field-${field.toLowerCase().split(' ').join('-')}`
  }

  function setMultiValue (value, index) {
    const arr = Array.isArray(value) ? [...value] : []
    arr[index] = value
    if (props.field.isSystem) {
      props.record[props.field.name] = arr
    } else {
      props.record.values[props.field.name] = arr
    }
    emit('change', value)
  }

  return { formGroupStyleClasses, labelColClass, contentColClass, isRequired, state, value, showPopover, label, description, hint, inModal, getFieldCypressId, setMultiValue, getColor }
}
