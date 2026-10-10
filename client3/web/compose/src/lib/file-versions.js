// Helpers for the file version dialog of File fields.

// versionLabel: "v3" for a versioned file, '' for a file with no versions
export function versionLabel (att) {
  const n = Number(att && att.version)
  return n > 0 ? `v${n}` : ''
}

// replaceValue puts `to` where `from` stood in the field value (a single ID or
// a list of IDs for a multi-value field); other values keep their place.
// Nothing changes when `from` is not in the value.
export function replaceValue (value, from, to) {
  const f = String(from)
  const t = String(to)

  if (Array.isArray(value)) {
    return value.map(v => String(v) === f ? t : v)
  }

  return String(value) === f ? t : value
}

// canMakeCurrent: a version can be made current when it is not the one the
// record already points to
export function canMakeCurrent (current, versionAttachmentID) {
  return !(current || []).map(String).includes(String(versionAttachmentID))
}
