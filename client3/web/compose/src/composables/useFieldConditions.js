import { ref, toValue, watch, onMounted, onBeforeUnmount, inject } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { debounce } from 'lodash'
import { composables } from 'corteza-lib/vue/dist'

/**
 * Conditional field visibility for Record page blocks.
 *
 * Evaluates the block's `options.fieldConditions` rules (configured in the
 * block configurator, "Conditional fields" section) against the current
 * record via the system expression endpoint. A field whose condition is
 * falsy is hidden; optional per-rule `clearOnHide` (or the global
 * `clearConditionalFieldsOnHide`) resets its value when it becomes hidden.
 *
 * @param {object} sources
 * @param {object|import('vue').MaybeRefOrGetter<object>} sources.block  page block (options carry fieldConditions)
 * @param {object|import('vue').MaybeRefOrGetter<object>} [sources.record]  record being displayed/edited
 * @param {object|import('vue').MaybeRefOrGetter<object>} [sources.module]  record's module
 * @param {import('vue').MaybeRefOrGetter<boolean>} [sources.active]  when false, evaluation is skipped
 */
export function useFieldConditions ({ block, record, module, active = true } = {}) {
  const { t: $t } = useI18n({ useScope: 'global' })
  const { toastErrorHandler } = composables.useToast()
  const route = useRoute()
  const $auth = inject('$auth')
  const $SystemAPI = inject('$SystemAPI', typeof window !== 'undefined' ? window.__systemAPI : undefined)

  const hidden = ref(new Set())
  const evaluating = ref(false)
  let evaluationSeq = 0

  const conditions = () => {
    const list = toValue(block)?.options?.fieldConditions
    return Array.isArray(list) ? list : []
  }

  const enabled = () => {
    if (toValue(active) === false) return false
    if (route.name === 'admin.pages.builder') return false
    return conditions().some(({ field, condition }) => field && (condition || '').trim() !== '')
  }

  function screenVariables () {
    const width = window.innerWidth
    const breakpoint = width < 576 ? 'xs' : width < 768 ? 'sm' : width < 992 ? 'md' : width < 1200 ? 'lg' : width < 1400 ? 'xl' : 'xxl'
    return { width, height: window.innerHeight, userAgent: navigator.userAgent, breakpoint }
  }

  function prepareData () {
    const expressions = {}
    const rec = toValue(record)
    const variables = {
      user: $auth?.user,
      record: rec ? rec.serialize() : {},
      screen: screenVariables(),
    }

    conditions().forEach(({ field, condition }) => {
      if (field && (condition || '').trim() !== '') {
        expressions[field] = condition
      }
    })

    return { expressions, variables }
  }

  function clearNewlyHidden (previousHidden, nextHidden) {
    const newlyHidden = [...nextHidden].filter(key => !previousHidden.has(key))
    if (!newlyHidden.length) return

    const clearAllOnHide = !!toValue(block)?.options?.clearConditionalFieldsOnHide
    const rec = toValue(record)
    const mod = toValue(module)
    if (!rec || !mod) return

    conditions().forEach(({ field, clearOnHide }) => {
      if (!(clearAllOnHide || clearOnHide)) return
      if (!field || !newlyHidden.includes(String(field))) return

      const moduleField = (mod.fields || []).find(f => String(f.fieldID) === String(field) || f.name === field)
      if (!moduleField) return

      if (moduleField.isMulti) {
        rec.values[moduleField.name] = []
      } else {
        rec.values[moduleField.name] = undefined
      }
    })
  }

  async function runEvaluation () {
    if (!enabled()) return
    const { expressions, variables } = prepareData()
    if (!Object.keys(expressions).length) return

    const seq = ++evaluationSeq
    evaluating.value = true
    try {
      const res = await $SystemAPI.expressionEvaluate({ variables, expressions })
      if (seq !== evaluationSeq) return

      const previousHidden = hidden.value
      const nextHidden = new Set()
      Object.keys(res).forEach(key => {
        if (!res[key]) nextHidden.add(String(key))
      })
      hidden.value = nextHidden

      clearNewlyHidden(previousHidden, nextHidden)
    } catch (e) {
      if (seq === evaluationSeq) {
        toastErrorHandler($t('notification.record.fieldConditions.failed'))(e)
      }
    } finally {
      if (seq === evaluationSeq) evaluating.value = false
    }
  }

  const evaluate = debounce(runEvaluation, 300)

  function onRecordFieldChange () {
    evaluate()
  }

  onMounted(() => {
    window.addEventListener('record-field-change', onRecordFieldChange)
    evaluate()
  })

  onBeforeUnmount(() => {
    evaluate.cancel()
    window.removeEventListener('record-field-change', onRecordFieldChange)
  })

  watch(() => toValue(record)?.recordID, () => evaluate())
  watch(() => toValue(record)?.values, () => evaluate(), { deep: true })

  function isHidden (field) {
    const id = field?.fieldID && String(field.fieldID) !== '0' ? String(field.fieldID) : undefined
    return (!!id && hidden.value.has(id)) || (!!field?.name && hidden.value.has(field.name))
  }

  /**
   * RBAC read permission + conditional visibility.
   */
  function canDisplay (field) {
    if (field?.canReadRecordValue === false) return false
    return !isHidden(field)
  }

  return { hidden, evaluating, canDisplay, isHidden, evaluate, evaluateNow: runEvaluation }
}
