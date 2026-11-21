import { PageBlock, PageBlockInput, Registry } from './base'
import { Apply } from '../../../cast'

const kind = 'AiChat'

interface Options {
  prompt: string;
  model: string;
  // temperature is the default sampling temperature offered to users in the
  // chat (they can still adjust it via the in-chat slider). null/undefined
  // keeps the model/provider default.
  temperature: number | null;
}

const defaults: Readonly<Options> = Object.freeze({
  prompt: '',
  model: 'deepseek-v2',
  temperature: null,
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
    // o.temperature is typed number|null, but a form field bound to it can
    // still hand us '' (cleared input) at runtime — check for that too.
    if (o.temperature === null || o.temperature === undefined || (o.temperature as unknown) === '') {
      this.options.temperature = null
    } else {
      const n = Number(o.temperature)
      this.options.temperature = Number.isFinite(n) ? n : null
    }
  }
}

Registry.set(kind, PageBlockAiChat)
