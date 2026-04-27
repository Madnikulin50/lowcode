import { PageBlock, PageBlockInput } from './base';
interface FieldCondition {
    field: string;
    condition: string;
    clearOnHide?: boolean;
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
}
export declare class PageBlockRecord extends PageBlock {
    readonly kind = "Record";
    options: Options;
    constructor(i?: PageBlockInput);
    applyOptions(o?: Partial<Options>): void;
}
export {};
