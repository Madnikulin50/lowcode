import { PageBlock, PageBlockInput, Registry } from './base'
import { Apply, CortezaID, NoID } from '../../../cast'

const kind = 'Anomaly'

interface Options {
  moduleID: string;
  severity: string;
  status: string;
  showSummary: boolean;
  showList: boolean;
  limit: number;
  refreshRate: number;
}

const defaults: Readonly<Options> = Object.freeze({
  moduleID: NoID,
  severity: '',
  status: '',
  showSummary: true,
  showList: true,
  limit: 10,
  refreshRate: 0,
})

export class PageBlockAnomaly extends PageBlock {
  readonly kind = kind

  options: Options = { ...defaults }

  constructor (i?: PageBlockInput) {
    super(i)
    this.applyOptions(i?.options as Partial<Options>)
  }

  applyOptions (o?: Partial<Options>): void {
    if (!o) return

    Apply(this.options, o, CortezaID, 'moduleID')
    Apply(this.options, o, String, 'severity', 'status')
    Apply(this.options, o, Number, 'limit', 'refreshRate')
    Apply(this.options, o, Boolean, 'showSummary', 'showList')
  }
}

Registry.set(kind, PageBlockAnomaly)
