import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { compose, validator } from 'corteza-lib/js/dist'
import { useUiStore } from '../../../store/ui'

export function useEditorBase(props, emit) {
  const route = useRoute()
  const uiStore = useUIStore()

  const formGroupStyleClasses = computed(() => ({
    required: isRequired.value,
    small: false,
    'value-only': props.valueOnly,
  }))

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

  return { formGroupStyleClasses, isRequired, state, value, showPopover, label, description, hint, inModal, getFieldCypressId, setMultiValue }
}
