import { getter } from './util'

export default function (field, translations, currentLanguage, resource) {
  const get = getter(translations, currentLanguage, resource)

  field.label = get('label')
  field.options.description.view = get('meta.description.view') || undefined
  field.options.description.edit = get('meta.description.edit') || undefined
  field.options.hint.view = get('meta.hint.view') || undefined
  field.options.hint.edit = get('meta.hint.edit') || undefined
  field.options.prefix = get('meta.prefix') || undefined
  field.options.suffix = get('meta.suffix') || undefined

  if (field.expressions && Array.isArray(field.expressions.validators)) {
    for (const vld of field.expressions.validators) {
      vld.error = get(`expression.validator.${vld.validatorID}.error`)
    }
  }
}
