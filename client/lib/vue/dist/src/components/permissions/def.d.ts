export declare const modalOpenEventName = "c-permissions-modal-open";
interface ResourceParts {
    component: string;
    resourceType?: string;
    references: Array<string>;
    i18nPrefix: string;
}
/**
 * Splitting strings:
 *  - corteza::compose:moduleField/42/21/12
 *  - corteza::compose/42/21/12
 */
export declare function split(input: string): ResourceParts;
export {};
