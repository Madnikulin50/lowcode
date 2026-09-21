/**
 * Normalize a Font Awesome icon from stored config or picker values.
 * Accepts `['fas', 'file-alt']`, `'fas file-alt'`, `'fas fa-file-alt'`, or `'file-alt'`.
 */
export function parseFaIcon (src, fallback = ['fas', 'file-alt']) {
  if (Array.isArray(src)) {
    if (src.length >= 2 && src[1]) {
      const prefix = src[0] === 'fa' ? 'fas' : src[0]
      return [prefix, String(src[1]).replace(/^fa-/, '')]
    }
    return fallback
  }
  if (!src || typeof src !== 'string') return fallback
  const parts = src.trim().split(/\s+/).filter(Boolean)
  if (!parts.length) return fallback
  if (parts.length >= 2) {
    const prefix = parts[0] === 'fa' ? 'fas' : parts[0]
    const name = parts.slice(1).join(' ').replace(/^fa-/, '')
    return name ? [prefix, name] : fallback
  }
  return ['fas', parts[0].replace(/^fa-/, '')]
}
