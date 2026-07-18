import { compose } from '@cortezaproject/corteza-js';
interface ComposeUIContext {
    $namespace?: compose.Namespace;
    $module?: compose.Module;
    $record?: compose.Record;
    pages: Array<compose.Page>;
    emitter: (...a: any) => any;
    routePusher: (...a: any) => any;
}
/**
 * ComposeUIHelper provides helpers for accessing Compose's UI
 *
 */
export default class ComposeUIHelper {
    readonly $namespace?: compose.Namespace;
    readonly $module?: compose.Module;
    readonly $record?: compose.Record;
    readonly pages?: Array<compose.Page>;
    readonly emitter: (...a: any) => any;
    readonly routePusher: (...a: any) => any;
    /**
     *
     * @param {Namespace} ctx.$namespace - Current namespace
     * @param {Module} ctx.$module - Current module
     * @param {Record} ctx.$record - Current record
     * @param {Page[]} ctx.pages - Array of Page objects
     * @param {Function} ctx.emitter - Event emitter (vm.$emit)
     * @param {Function} ctx.routePusher - Route pusher (vm.$route.push)
     */
    constructor(ctx: ComposeUIContext);
    /**
     * Open record viewer page
     *
     * It searches for page that matches record's module and redirects
     * user to the view mode on that page
     *
     * @example
     * // Edit current record
     * ComposeUI.gotoRecordViewer($record)
     *
     * // Edit current record ($record can be omitted)
     * ComposeUI.gotoRecordViewer()
     *
     * @param {Record} record
     */
    gotoRecordViewer(record?: compose.Record | undefined): void;
    /**
     * Open record editor page
     *
     * It searches for page that matches record's module and redirects
     * user to the edit mode on that page.
     *
     * @example
     * // Edit current record
     * ComposeUI.gotoRecordEditor($record)
     *
     * // Edit current record ($record can be omitted)
     * ComposeUI.gotoRecordEditor()
     *
     * @param {Record} record
     */
    gotoRecordEditor(record?: compose.Record | undefined): void;
    /**
     * Open record page
     *
     * @private
     * @param {string} name
     * @param {Record} record
     * @param {string} record.recordID
     * @param {string} record.moduleID
     */
    gotoRecordPage(name: string, record?: compose.Record | undefined): void;
    /**
     * Returns record page
     *
     */
    private getRecordPage;
    /**
     * Go to a specific route
     *
     * @private
     * @param {string} name
     * @param {Object} params for $router.push
     */
    goto(name: string, params: object): void;
    /**
     * Show a success alert
     *
     * @example
     * ComposeUI.success('Change was successful')
     *
     * @param message
     */
    success(message: string): void;
    /**
     * Show a warning alert
     *
     * @example
     * ComposeUI.warning('Could not save your changes')
     *
     * @param message
     */
    warning(message: string): void;
}
export {};
