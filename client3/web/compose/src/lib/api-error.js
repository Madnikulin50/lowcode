// Shared helpers for turning raw API errors into something a user can read.
//
// Go's context.DeadlineExceeded bubbles up from the server as a bare
// "context deadline exceeded" string (no user-friendly wrapping), whenever a
// request (record report, record list, ...) is cancelled by a server-side
// timeout. PageBlocks that render API errors directly (Chart, Metric) show
// this raw string as-is unless it's caught and swapped for a translated
// message first.

const TIMEOUT_PATTERN = /context deadline exceeded/i

/**
 * @param {unknown} err
 * @returns {boolean}
 */
export function isTimeoutError (err) {
  if (err == null) return false
  const msg = typeof err === 'string'
    ? err
    : (err.message || err.response?.data?.error?.message || err.response?.data?.error || '')
  return TIMEOUT_PATTERN.test(String(msg))
}

/**
 * Returns a message safe to render to the user: the translated timeout
 * message for a timed-out request, or the error's own message otherwise.
 *
 * @param {unknown} err
 * @param {(key: string) => string} t
 * @returns {string}
 */
export function friendlyApiErrorMessage (err, t) {
  if (isTimeoutError(err)) return t('error.timeout')
  if (typeof err === 'string') return err
  return err?.message || err?.response?.data?.error?.message || err?.response?.data?.error || String(err)
}
