/**
 * Same sentinel as server pkg/filter.TotalUnknown: COUNT timed out
 * (record/report empty-metrics path, or list incTotal).
 */
export const TotalUnknown = -1

/**
 * True when a count/total is the unknown sentinel, not a real metric.
 * JSON may yield a number or a string depending on the client.
 */
export function isUnknownTotal (n: unknown): boolean {
  return n === TotalUnknown || n === '-1'
}

/**
 * True when a record/report row is the COUNT-timeout payload `{ count: -1 }`.
 * Rows with `rp` are real aggregates (MAX/MIN/…) and must not be treated as unknown
 * even if that aggregate happens to be -1.
 */
export function isUnknownReportCount (row: unknown): boolean {
  if (!row || typeof row !== 'object') {
    return false
  }
  const r = row as { count?: unknown, rp?: unknown }
  if (r.rp !== undefined && r.rp !== null) {
    return false
  }
  return isUnknownTotal(r.count)
}
