import { PageBlock, PageBlockInput, Registry } from './base'
import { Apply } from '../../../cast'

const kind = 'Document'

interface Options {
  documentID: string;
}

const defaults: Readonly<Options> = Object.freeze({
  documentID: '',
})

export class PageBlockDocument extends PageBlock {
  readonly kind = kind

  options: Options = { ...defaults }

  constructor (i?: PageBlockInput) {
    super(i)
    this.applyOptions(i?.options as Partial<Options>)
  }

  applyOptions (o?: Partial<Options>): void {
    if (!o) return

    Apply(this.options, o, String, 'documentID')
  }
}

Registry.set(kind, PageBlockDocument)
