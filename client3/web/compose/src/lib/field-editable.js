export function isFieldReadonly (field) {
  if (!field) return false
  return !!(field.isReadonly || field.options?.readonly)
}

export function isUserWritableField (field) {
  if (!field || isFieldReadonly(field)) return false
  if (field.canUpdateRecordValue === false) return false
  if (field.expressions?.value) return false
  return true
}
