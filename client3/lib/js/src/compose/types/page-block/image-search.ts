import { PageBlock, PageBlockInput, Registry } from './base'
import { Apply } from '../../../cast'

const kind = 'ImageSearch'

interface Options {
  // Record field names (handles) concatenated into the search query, e.g.
  // ["store_name", "address"] for "МСК-01 г. Москва, ул. Арбат, д. 54/2".
  fields: string[];
  // Literal text appended after the field values — a domain hint ("фасад
  // магазина"), or a fallback when fields is empty.
  extraQuery: string;
  limit: number;
}

const defaults: Readonly<Options> = Object.freeze({
  fields: [],
  extraQuery: '',
  limit: 6,
})

export class PageBlockImageSearch extends PageBlock {
  readonly kind = kind

  options: Options = { ...defaults }

  constructor (i?: PageBlockInput) {
    super(i)
    this.applyOptions(i?.options as Partial<Options>)
  }

  applyOptions (o?: Partial<Options>): void {
    if (!o) return
    Apply(this.options, o, String, 'extraQuery')
    Apply(this.options, o, Number, 'limit')
    if (Array.isArray(o.fields)) {
      this.options.fields = o.fields.map(String)
    }
  }
}

Registry.set(kind, PageBlockImageSearch)
