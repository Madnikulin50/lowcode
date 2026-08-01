import { PageBlock, PageBlockInput, Registry } from './base'
import { Apply } from '../../../cast'

const kind = 'AiChat'

interface Options {
  prompt: string;
  model: string;
}

const defaults: Readonly<Options> = Object.freeze({
  prompt: '',
  model: 'deepseek-v2',
})

export class PageBlockAiChat extends PageBlock {
  readonly kind = kind

  options: Options = { ...defaults }

  constructor (i?: PageBlockInput) {
    super(i)
    this.applyOptions(i?.options as Partial<Options>)
  }

  applyOptions (o?: Partial<Options>): void {
    if (!o) return
    Apply(this.options, o, String, 'prompt', 'model')
  }
}

Registry.set(kind, PageBlockAiChat)
