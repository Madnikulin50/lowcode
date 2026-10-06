import { PageBlock, PageBlockInput, Registry } from './base'
import { Apply, CortezaID, NoID } from '../../../cast'

const kind = 'RelatedRecords'

// One list of child records: records of moduleID whose refField points at the
// record the page shows.
export interface Relation {
  moduleID: string;
  // name of the Record field in moduleID that refers to the page's module
  refField: string;
  // section title; empty means the module's name
  title: string;
  // columns, as in a RecordList block; empty means all the module's fields
  fields: unknown[];
  perPage: number;
  presort: string;
  hideAddButton: boolean;
  // open the section when the page loads
  expanded: boolean;
}

interface Options {
  relations: Relation[];
  // do not show sections that have no records
  hideEmptySections: boolean;
  // show how many records each section has in its header
  showCounts: boolean;
}

const relationDefaults: Readonly<Relation> = Object.freeze({
  moduleID: NoID,
  refField: '',
  title: '',
  fields: [],
  perPage: 10,
  presort: 'createdAt DESC',
  hideAddButton: false,
  expanded: false,
})

const defaults: Readonly<Options> = Object.freeze({
  relations: [],
  hideEmptySections: false,
  showCounts: true,
})

export function makeRelation (r?: Partial<Relation>): Relation {
  const out: Relation = { ...relationDefaults, fields: [] }
  if (!r) return out

  Apply(out, r, CortezaID, 'moduleID')
  Apply(out, r, String, 'refField', 'title', 'presort')
  Apply(out, r, Number, 'perPage')
  Apply(out, r, Boolean, 'hideAddButton', 'expanded')
  if (Array.isArray(r.fields)) out.fields = [...r.fields]
  if (!(out.perPage > 0)) out.perPage = relationDefaults.perPage
  return out
}

export class PageBlockRelatedRecords extends PageBlock {
  readonly kind = kind

  options: Options = { ...defaults, relations: [] }

  constructor (i?: PageBlockInput) {
    super(i)
    this.applyOptions(i?.options as Partial<Options>)
  }

  applyOptions (o?: Partial<Options>): void {
    if (!o) return

    Apply(this.options, o, Boolean, 'hideEmptySections')
    if (o.showCounts !== undefined) this.options.showCounts = !!o.showCounts
    if (Array.isArray(o.relations)) {
      this.options.relations = o.relations.map(r => makeRelation(r))
    }
  }
}

Registry.set(kind, PageBlockRelatedRecords)
