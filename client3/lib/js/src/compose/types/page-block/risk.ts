import { PageBlock, PageBlockInput, Registry } from './base'
import { Apply } from '../../../cast'

const kind = 'Risk'

interface Options {
  // 'record' shows the current record's latest assessment (record cards);
  // 'portfolio' shows a ranked list + level distribution across every
  // subject under the binding (reports/dashboards). One block kind, two
  // views — see server/pkg/riskengine's architecture note.
  mode: string;
  bindingID: string;
  showAttribution: boolean;
  showExplanation: boolean;
  topN: number;
}

const defaults: Readonly<Options> = Object.freeze({
  mode: 'record',
  bindingID: '',
  showAttribution: true,
  showExplanation: true,
  topN: 10,
})

export class PageBlockRisk extends PageBlock {
  readonly kind = kind

  options: Options = { ...defaults }

  constructor (i?: PageBlockInput) {
    super(i)
    this.applyOptions(i?.options as Partial<Options>)
  }

  applyOptions (o?: Partial<Options>): void {
    if (!o) return
    Apply(this.options, o, String, 'mode', 'bindingID')
    Apply(this.options, o, Boolean, 'showAttribution', 'showExplanation')
    Apply(this.options, o, Number, 'topN')
  }
}

Registry.set(kind, PageBlockRisk)
