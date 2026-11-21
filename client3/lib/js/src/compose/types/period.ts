// Shared QL date-window filter generation for period-over-period comparison
// features. Used by both PageBlockMetric (page-block/metric.ts) and Chart
// (chart/chart.ts) — lives at this level (sibling to both page-block/ and
// chart/) rather than inside either, so neither has to import from the
// other's subtree. No Vue/DOM dependency — independently testable.
//
// Prefers this system's own native period-comparison QL functions
// (this_month/prev_month/prev_month_truncated, and the week/quarter/year
// equivalents — server/pkg/gvalfnc/time.go, translated per-driver e.g.
// server/store/adapters/rdbms/drivers/postgres/ql.go) over hand-rolled
// DATE_FORMAT/DATE_SUB SQL: they're simpler, already used elsewhere in this
// app (see the "Сводка" page's hand-built "Выручка месяц-к-месяцу" block),
// and prev_X_truncated compares only the *same partial stretch* of the
// previous period — correct when the current period isn't finished yet,
// which the generic approach can't do.
//
// Native "prev_X_truncated" functions only exist for week/month/quarter
// (server/store/adapters/rdbms/drivers/postgres/ql.go); year-over-year
// comparison at those granularities has no working native equivalent
// (this_month_prev_year_truncated is registered but unimplemented for any
// SQL driver) — that combination falls back to generic DATE_FORMAT/YEAR/
// DATE_SUB SQL, evaluated live against NOW() at query time either way.

import moment from 'moment'

export type PeriodGranularity = 'day' | 'week' | 'month' | 'quarter' | 'year'
export type PeriodCompareMode = 'previous-period' | 'year-over-year'

export interface BuildPeriodFiltersInput {
  field: string;
  granularity: PeriodGranularity;
  mode: PeriodCompareMode;
}

export interface PeriodFilters {
  currentFilter: string;
  previousFilter: string;
}

interface GranularityDef {
  /** Native "this <bucket>" predicate function, when one exists for this granularity. */
  nativeThis?: (f: string) => string;
  /** Native "previous <bucket>, same partial stretch" predicate function. */
  nativePrevTruncated?: (f: string) => string;
  /** Generic fallback: QL expression matching field f against the same bucket as reference date expr `ref`. */
  bucketExpr: (f: string, ref: string) => string;
  /** DATE_SUB(...) interval literal for one bucket of this granularity (generic fallback only). */
  interval: string;
}

const GRANULARITY: Record<PeriodGranularity, GranularityDef> = {
  day: {
    // No native this_day/prev_day function — day-level comparison has no
    // partial-period ambiguity to truncate anyway.
    bucketExpr: (f, ref) => `DATE(${f}) = DATE(${ref})`,
    interval: '1 DAY',
  },
  week: {
    nativeThis: (f) => `this_week(${f})`,
    nativePrevTruncated: (f) => `prev_week_truncated(${f})`,
    bucketExpr: (f, ref) => `YEAR(${f}) = YEAR(${ref}) AND WEEK(${f}) = WEEK(${ref})`,
    interval: '1 WEEK',
  },
  month: {
    nativeThis: (f) => `this_month(${f})`,
    nativePrevTruncated: (f) => `prev_month_truncated(${f})`,
    bucketExpr: (f, ref) => `DATE_FORMAT(${f}, '%Y-%m') = DATE_FORMAT(${ref}, '%Y-%m')`,
    interval: '1 MONTH',
  },
  quarter: {
    nativeThis: (f) => `this_quarter(${f})`,
    nativePrevTruncated: (f) => `prev_quarter_truncated(${f})`,
    bucketExpr: (f, ref) => `YEAR(${f}) = YEAR(${ref}) AND QUARTER(${f}) = QUARTER(${ref})`,
    interval: '3 MONTH',
  },
  year: {
    nativeThis: (f) => `this_year(${f})`,
    // No prev_year_truncated needed for the year-vs-year case: prev_year(f)
    // already means "any time last year", which is what previous-period
    // AND year-over-year both mean at year granularity.
    nativePrevTruncated: (f) => `prev_year(${f})`,
    bucketExpr: (f, ref) => `YEAR(${f}) = YEAR(${ref})`,
    interval: '1 YEAR',
  },
}

/**
 * Builds a "current period" and "previous period" QL filter pair for a
 * DateTime field, evaluated live against NOW() at query time.
 *
 * mode: 'previous-period' compares against the same partial stretch of the
 * immediately preceding bucket (month -> previous calendar month, apples-
 * to-apples truncated to today's day-of-month; quarter/week analogous).
 * 'year-over-year' compares against the same bucket one year ago (month ->
 * same month last year). For granularity 'year' the two modes coincide.
 */
export function buildPeriodFilters ({ field, granularity, mode }: BuildPeriodFiltersInput): PeriodFilters {
  const g = GRANULARITY[granularity] || GRANULARITY.month
  const now = 'NOW()'

  const currentFilter = g.nativeThis ? g.nativeThis(field) : g.bucketExpr(field, now)

  if (mode === 'previous-period' || granularity === 'year') {
    const previousFilter = g.nativePrevTruncated
      ? g.nativePrevTruncated(field)
      : g.bucketExpr(field, `DATE_SUB(${now}, INTERVAL ${g.interval})`)
    return { currentFilter, previousFilter }
  }

  // year-over-year at sub-year granularity: no native truncated equivalent
  // implemented anywhere yet, fall back to generic SQL.
  const previousFilter = g.bucketExpr(field, `DATE_SUB(${now}, INTERVAL 1 YEAR)`)
  return { currentFilter, previousFilter }
}

/**
 * Escapes a value for a QL string literal (single-quote delimited).
 * Mirrors corteza-webapp-compose/src/lib/record-filter.js's escapeQlString —
 * duplicated here (not imported) because lib/js is a lower-level package
 * that web/compose depends on, not the other way around.
 */
function escapeQlString (str: string): string {
  return String(str)
    .replace(/\\/g, '\\\\')
    .replace(/'/g, "\\'")
}

/** Builds a `field = 'value'` QL equality filter, safely escaped. */
export function buildEqualityFilter (field: string, value: string): string {
  return `${field} = '${escapeQlString(value)}'`
}

/** ANDs two QL filter fragments, omitting either side if empty. */
export function andFilters (base: string | undefined | null, extra: string): string {
  const b = (base || '').trim()
  if (!b) return extra
  return `(${b}) AND (${extra})`
}

/**
 * Maps an absolute date value to a position *relative to its own period*,
 * so a "current period" row and a "previous period" row that fall on the
 * same relative position line up on one shared chart x-axis despite having
 * different absolute dates (e.g. Aug 15 and Jul 15 both become day 15).
 *
 * Returned as a zero-padded string so lexicographic and numeric sort agree
 * (matters for e.g. day-of-quarter reaching 3 digits).
 */
export function relativeBucketLabel (value: unknown, granularity: PeriodGranularity): string {
  const m = moment(value as any)
  if (!m.isValid()) return String(value ?? '')

  switch (granularity) {
    case 'week':
      // ISO day of week, 1 (Mon) .. 7 (Sun) — stable regardless of locale.
      return String(m.isoWeekday())
    case 'month':
      return String(m.date()).padStart(2, '0')
    case 'quarter': {
      const startOfQuarter = m.clone().startOf('quarter')
      return String(m.diff(startOfQuarter, 'days') + 1).padStart(3, '0')
    }
    case 'year':
      return String(m.month() + 1).padStart(2, '0')
    case 'day':
    default:
      return m.format('YYYY-MM-DD')
  }
}
