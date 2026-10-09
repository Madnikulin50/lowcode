import { feedResources } from './resources'
import { Apply, NoID } from '../../../../cast'
import { IsOf } from '../../../../guards'

interface FeedOptions {
  moduleID: string;
  color: string;
  prefilter: string;
}

interface LegacyFeed {
  moduleID?: string;
  startField?: string;
  endField?: string;
  titleField?: string;
  allDay?: boolean;
}

export type FeedInput = Partial<Feed> | Feed | LegacyFeed

const defOptions = {
  moduleID: NoID,
  color: '#4e73df',
  prefilter: '',
}

/**
 * Normalize a calendar feed field name for QL.
 * Vue 3 selects may store the whole option object, `undefined`, or the
 * literal string "undefined"; none of those are valid attribute names.
 */
export function asFeedFieldName (value: unknown): string {
  if (value && typeof value === 'object' && !Array.isArray(value) && 'name' in (value as { name?: unknown })) {
    return asFeedFieldName((value as { name: unknown }).name)
  }
  if (typeof value !== 'string') {
    return ''
  }
  const name = value.trim()
  if (!name || name === 'undefined' || name === 'null') {
    return ''
  }
  return name
}

/**
 * Feed class represents an event feed for the given calendar
 */
export default class Feed {
  public resource = 'compose:record'
  public startField = ''
  public endField = ''
  public titleField = ''
  public options: FeedOptions = { ...defOptions }

  public allDay = false

  constructor (i?: FeedInput) {
    this.apply(i)
  }

  apply (i?: FeedInput): void {
    if (!i) return

    if (!IsOf<Feed>(i, 'resource') && IsOf<LegacyFeed>(i, 'moduleID')) {
      i = Feed.fromLegacy(i)
    }

    if (IsOf<Feed>(i, 'resource')) {
      Apply(this, i, String, 'resource')
      Apply(this, i, asFeedFieldName, 'startField', 'endField', 'titleField')
      Apply(this, i, Boolean, 'allDay')

      if (i.options) {
        this.options = { ...this.options, ...i.options }
        if (!this.options.color) {
          this.options.color = defOptions.color
        }
        if (this.options.prefilter == null) {
          this.options.prefilter = ''
        }
        // Field names sometimes land under options (Vue 3 / legacy JSON).
        const nested = i.options as Partial<Feed>
        if (!this.startField) this.startField = asFeedFieldName(nested.startField)
        if (!this.endField) this.endField = asFeedFieldName(nested.endField)
        if (!this.titleField) this.titleField = asFeedFieldName(nested.titleField)
      }
    }
  }

  isValid (): boolean {
    return this.options.moduleID !== NoID && !!asFeedFieldName(this.startField)
  }

  static fromLegacy (legacy: LegacyFeed): Partial<Feed> {
    const p: Partial<Feed> = {
      // legacy does not have resource,
      // we've used it with records only
      resource: feedResources.record,

      ...legacy,
    }

    if (legacy.moduleID) {
      if (!p.options) {
        p.options = { ...defOptions }
      }

      // module was moved under options
      p.options.moduleID = legacy.moduleID
    }

    return p
  }
}
