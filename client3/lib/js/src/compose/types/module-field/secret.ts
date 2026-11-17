import { ModuleField, Registry, Options, defaultOptions } from './base'
import { Apply } from '../../../cast'

const kind = 'Secret'

interface SecretOptions extends Options {
  multiDelimiter: string;
}

const defaults = (): Readonly<SecretOptions> => Object.freeze({
  ...defaultOptions(),
  multiDelimiter: '\n',
})

// Secret — серверно-шифруемое строковое поле: хранится в БД шифротекстом
// (AES-256-GCM, ключ платформы из SECRET_FIELD_KEY), API отдаёт маску,
// полное значение — только по ?exposeSecrets=1 с правом обновления записи.
export class ModuleFieldSecret extends ModuleField {
  readonly kind = kind

  options: SecretOptions = { ...defaults() }

  constructor (i?: Partial<ModuleFieldSecret>) {
    super(i)
    this.applyOptions(i?.options)
  }

  applyOptions (o?: Partial<SecretOptions>): void {
    if (!o) return
    super.applyOptions(o)
    Apply(this.options, o, String, 'multiDelimiter')
  }
}

Registry.set(kind, ModuleFieldSecret)
