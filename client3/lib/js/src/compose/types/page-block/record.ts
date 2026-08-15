import { PageBlock, PageBlockInput, Registry } from './base'
import { Apply } from '../../../cast'

const kind = 'Record'

interface FieldCondition {
  field: string;
  condition: string;
  clearOnHide?: boolean;
}

export type RecordFieldRole = 'default' | 'title' | 'subtitle' | 'badge' | 'meta' | 'body'

export interface RecordSection {
  title: string;
  fields: string[];
}

interface Options {
  fields: unknown[];
  fieldConditions: FieldCondition[];
  clearConditionalFieldsOnHide: boolean;
  recordSelectorShowAddRecordButton: boolean;
  magnifyOption: string;
  recordSelectorDisplayOption: string;
  recordSelectorAddRecordDisplayOption: string;
  referenceField?: string;
  referenceModuleID?: string;
  inlineRecordEditEnabled: boolean;
  horizontalFieldLayoutEnabled: boolean;
  recordFieldLayoutOption: string;
  inlineRecordEditAllowAddField: boolean;
  viewStyle: object;
  /** Field display density */
  density: 'comfortable' | 'compact';
  /** Hide fields with empty values */
  hideEmptyFields: boolean;
  /** Show placeholder when value is empty (if not hidden) */
  showEmptyPlaceholder: boolean;
  /** Role per field name: title | subtitle | badge | meta | body | default */
  fieldRoles: { [fieldName: string]: RecordFieldRole };
  /** Named field groups for the record body */
  sections: RecordSection[];
}

const defaults: Readonly<Options> = Object.freeze({
  fields: [],
  fieldConditions: [],
  clearConditionalFieldsOnHide: false,
  recordSelectorShowAddRecordButton: false,
  magnifyOption: '',
  recordSelectorDisplayOption: 'sameTab',
  recordSelectorAddRecordDisplayOption: 'sameTab',
  referenceField: '',
  referenceModuleID: undefined,
  inlineRecordEditEnabled: false,
  inlineRecordEditAllowAddField: false,
  horizontalFieldLayoutEnabled: false,
  recordFieldLayoutOption: 'default',
  viewStyle: {},
  density: 'comfortable',
  hideEmptyFields: false,
  showEmptyPlaceholder: true,
  fieldRoles: {},
  sections: [],
})

export class PageBlockRecord extends PageBlock {
  readonly kind = kind

  options: Options = {
    ...defaults,
    fieldRoles: {},
    sections: [],
    viewStyle: {},
  }

  constructor (i?: PageBlockInput) {
    super(i)
    this.applyOptions(i?.options as Partial<Options>)
  }

  applyOptions (o?: Partial<Options>): void {
    if (!o) return

    Apply(this.options, o, String, 'magnifyOption', 'recordSelectorDisplayOption', 'recordSelectorAddRecordDisplayOption', 'referenceField', 'referenceModuleID', 'recordFieldLayoutOption', 'density')
    Apply(this.options, o, Boolean, 'recordSelectorShowAddRecordButton', 'inlineRecordEditEnabled', 'horizontalFieldLayoutEnabled', 'inlineRecordEditAllowAddField', 'clearConditionalFieldsOnHide', 'hideEmptyFields', 'showEmptyPlaceholder')

    if (o.fields) {
      this.options.fields = o.fields
    }

    if (o.fieldConditions) {
      this.options.fieldConditions = o.fieldConditions
    }

    if (o.viewStyle) {
      this.options.viewStyle = o.viewStyle
    }

    if (o.fieldRoles && typeof o.fieldRoles === 'object') {
      this.options.fieldRoles = { ...o.fieldRoles }
    }

    if (Array.isArray(o.sections)) {
      this.options.sections = o.sections.map(s => ({
        title: s?.title || '',
        fields: Array.isArray(s?.fields) ? [...s.fields] : [],
      }))
    }
  }
}

Registry.set(kind, PageBlockRecord)
