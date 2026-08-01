import { PageBlock, PageBlockInput, Registry } from './base'
import { Apply } from '../../../cast'

const kind = 'RuleChain'

interface Options {
  chainID: string;
  label: string;
  variant: string;
  size: string;
  icon: string;
  context: Record<string, unknown>;
}

const defaults: Readonly<Options> = Object.freeze({
  chainID: '',
  label: '',
  variant: 'primary',
  size: '',
  icon: 'play',
  context: {},
})

export class PageBlockRuleChain extends PageBlock {
  readonly kind = kind

  options: Options = { ...defaults }

  constructor (i?: PageBlockInput) {
    super(i)
    this.applyOptions(i?.options as Partial<Options>)
  }

  applyOptions (o?: Partial<Options>): void {
    if (!o) return
    Apply(this.options, o, String, 'chainID', 'label', 'variant', 'size', 'icon')
    if (o.context) {
      this.options.context = { ...this.options.context, ...o.context }
    }
  }
}

Registry.set(kind, PageBlockRuleChain)
