import { PageBlock, PageBlockInput } from './base';
interface Options {
    moduleID: string;
    filter: string;
    titleField: string;
    contentField: string;
    replyField: string;
    referenceField: string;
    sortDirection: string;
    refreshRate: number;
    showRefresh: boolean;
    magnifyOption: string;
    attachmentField: string;
    reactionsField: string;
}
export declare class PageBlockComment extends PageBlock {
    readonly kind = "Comment";
    options: Options;
    constructor(i?: PageBlockInput);
    applyOptions(o?: Partial<Options>): void;
}
export {};
