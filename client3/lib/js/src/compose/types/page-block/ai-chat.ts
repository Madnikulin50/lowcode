import { PageBlock, PageBlockInput, Registry } from './base'
import { Apply } from '../../../cast'

const kind = 'AiChat'

interface Options {
  startPrompt: string;
  model: string;
}

const defaults: Readonly<Options> = Object.freeze({
  startPrompt: '',
  model: '',
})

export class PageBlockAiChat extends PageBlock {
  readonly kind = kind

  options: Options = { ...defaults }

  constructor (i?: PageBlockInput) {
    super(i)
    this.applyOptions(i?.options as Partial<Options> & { prompt?: string })
  }

  applyOptions (o?: Partial<Options> & { prompt?: string }): void {
    if (!o) return
    Apply(this.options, o, String, 'startPrompt', 'model')
    // Legacy key: configurator used to bind `options.startPrompt` while the
    // type stored `options.prompt`. PageBlockMaker re-inits defaults and
    // would drop an unknown startPrompt unless we copy it here.
    if (!this.options.startPrompt && o.prompt) {
      this.options.startPrompt = String(o.prompt)
    }
  }
}

Registry.set(kind, PageBlockAiChat)
