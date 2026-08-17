import { Apply, NoID } from '../../../../cast'
import { IsOf } from '../../../../guards'
import { asFeedFieldName } from '../calendar/feed'

interface FeedOptions {
  color: string;
  prefilter: string;
  moduleID: string;
}

export type FeedInput = Partial<Feed> | Feed

const defOptions = {
  moduleID: NoID,
  color: '#4e73df',
  prefilter: '',
}

/**
 * Feed class represents an event feed for the given calendar
 */
export default class Feed {
  public resource = 'compose:record'
  public titleField = ''
  public geometryField = ''
  public displayMarker = true
  public displayPolygon = false
  public options: FeedOptions = { ...defOptions }

  constructor (i?: FeedInput) {
    this.apply(i)
  }

  apply (i?: FeedInput): void {
    if (!i) return

    if (IsOf<Feed>(i, 'resource')) {
      Apply(this, i, String, 'resource')
      Apply(this, i, Boolean, 'displayMarker', 'displayPolygon')
      this.titleField = asFeedFieldName(i.titleField)
      this.geometryField = asFeedFieldName(i.geometryField)

      if (i.options) {
        this.options = { ...this.options, ...i.options }
      }
    }
  }

  isValid (): boolean {
    return this.options.moduleID !== NoID && !!asFeedFieldName(this.geometryField)
  }
}
