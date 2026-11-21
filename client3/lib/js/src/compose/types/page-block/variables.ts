import { PageBlock, PageBlockInput, Registry } from './base'
import lodash from 'lodash'
const { merge } = lodash
import { Apply } from '../../../cast'
const kind = 'Variables'

export interface PageVariableDef {
  name: string;
  label: string;
  /** A ModuleFieldRegistry kind: String/Number/Bool/DateTime/Select/Record/User/... */
  kind: string;
  isRequired: boolean;
  isMulti: boolean;
  /** Kind-specific field options (Select.options, Record.moduleID, ...) */
  options: Record<string, any>;
  /** Raw scalar/array value, not the ModuleField {name,value}[] default-value shape */
  defaultValue: any;
}

const defaultVariable: Readonly<PageVariableDef> = Object.freeze({
  name: '',
  label: '',
  kind: 'String',
  isRequired: false,
  isMulti: false,
  options: {},
  defaultValue: '',
})

interface Options {
  variables: PageVariableDef[];
  density: 'comfortable' | 'compact';
  horizontalFieldLayoutEnabled: boolean;
  showVariableLabels: boolean;
}

const defaults: Readonly<Options> = Object.freeze({
  variables: [],
  density: 'comfortable',
  horizontalFieldLayoutEnabled: false,
  showVariableLabels: false,
})

export class PageBlockVariables extends PageBlock {
  readonly kind = kind

  options: Options = {
    ...defaults,
    variables: [],
  }

  constructor (i?: PageBlockInput) {
    super(i)
    this.applyOptions(i?.options as Partial<Options>)
  }

  applyOptions (o?: Partial<Options>): void {
    if (!o) return
    Apply(this.options, o, String, 'density')
    Apply(this.options, o, Boolean, 'horizontalFieldLayoutEnabled', 'showVariableLabels')

    if (Array.isArray(o.variables)) {
      this.options.variables = o.variables.map(v => merge({}, defaultVariable, v))
    }
  }
}

Registry.set(kind, PageBlockVariables)
