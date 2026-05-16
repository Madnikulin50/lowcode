import { ModuleField, Registry, Options, defaultOptions } from './base'
import { Apply } from '../../../cast'

const kind = 'Expression'

interface ExpressionOptions extends Options {

}

const defaults = (): Readonly<ExpressionOptions> => Object.freeze({
  ...defaultOptions(),
})

export class ModuleFieldExpression extends ModuleField {
  readonly kind = kind

  options: ExpressionOptions = { ...defaults() }

  constructor (i?: Partial<ModuleFieldExpression>) {
    super(i)

    this.applyOptions(i?.options)
  }

  applyOptions (o?: Partial<ExpressionOptions>): void {
    if (!o) return
    super.applyOptions(o)
  }
}

Registry.set(kind, ModuleFieldExpression)
