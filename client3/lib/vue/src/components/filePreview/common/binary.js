export const OFFICE_MAX_BYTES = 30 * 1024 * 1024
export const CAD_MAX_BYTES = 50 * 1024 * 1024
export const BIM_MAX_BYTES = 80 * 1024 * 1024

export function originalSize (meta) {
  const size = meta?.original?.size
  return typeof size === 'number' && size >= 0 ? size : undefined
}

export function assertPreviewSize (meta, maxBytes, labels = {}) {
  const size = originalSize(meta)
  if (size != null && size > maxBytes) {
    const err = new Error(labels.tooLarge || 'File is too large to preview in the browser. Please download it.')
    err.code = 'tooLarge'
    throw err
  }
}

export async function fetchBinary (src) {
  if (src instanceof ArrayBuffer) {
    return src
  }
  if (src instanceof Blob) {
    return src.arrayBuffer()
  }
  if (typeof src !== 'string' || !src) {
    const err = new Error('src.notValid')
    err.code = 'src'
    throw err
  }
  const res = await fetch(src)
  if (!res.ok) {
    const err = new Error(`Failed to load file (${res.status})`)
    err.code = 'fetch'
    throw err
  }
  return res.arrayBuffer()
}
