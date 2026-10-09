import { ModuleField, Registry, Options, defaultOptions } from './base'
import { Apply } from '../../../cast'

const kind = 'String'

interface JSONVariant {
  value: string;
  variant: string;
}

interface StringOptions extends Options {
  multiLine: boolean;
  useRichTextEditor: boolean;
  multiDelimiter: string;
  displayType: 'text' | 'json' | 'ports' | '';
  jsonLayout: 'chips' | 'table' | 'kv' | 'pretty' | '';
  jsonTemplate: string;
  jsonFields: string;
  jsonVariantField: string;
  jsonVariants: JSONVariant[];
}

const defaults = (): Readonly<StringOptions> => Object.freeze({
  ...defaultOptions(),
  multiLine: false,
  useRichTextEditor: false,
  multiDelimiter: '\n',
  displayType: '',
  jsonLayout: '',
  jsonTemplate: '',
  jsonFields: '',
  jsonVariantField: '',
  jsonVariants: [],
})

export class ModuleFieldString extends ModuleField {
  readonly kind = kind

  options: StringOptions = { ...defaults() }

  constructor (i?: Partial<ModuleFieldString>) {
    super(i)

    this.applyOptions(i?.options)
  }

  applyOptions (o?: Partial<StringOptions>): void {
    if (!o) return
    super.applyOptions(o)

    Apply(this.options, o, String, 'multiDelimiter', 'displayType', 'jsonLayout', 'jsonTemplate', 'jsonFields', 'jsonVariantField')
    Apply(this.options, o, Boolean, 'multiLine', 'useRichTextEditor')

    if (o.jsonVariants) {
      this.options.jsonVariants = Array.isArray(o.jsonVariants) ? o.jsonVariants : []
    }
  }
}

Registry.set(kind, ModuleFieldString)
