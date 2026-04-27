import { PluginFunction } from 'vue';
import { eventbus } from '../../../../lib/js/dist';
interface UIProp {
    name: string;
    value: string;
}
interface Trigger {
    resourceTypes: string[];
    eventTypes: string[];
    uiProps: UIProp[];
    constraints: object[];
    weight?: number;
}
interface Script {
    name: string;
    label: string;
    description?: string;
    errors?: string[];
    triggers: Trigger[];
}
export declare class Button {
    readonly label: string;
    readonly description?: string;
    readonly script: string;
    readonly resourceType: string;
    readonly weight: number;
    readonly variant?: string;
    readonly page?: string;
    readonly slot?: string;
    readonly constraints: eventbus.ConstraintMatcher[];
    constructor(s: Script, t: Trigger);
}
/**
 * Consumes scripts that can be triggered manually and converts it to list of buttons
 *
 * These buttons can put manually to various compose page block or
 * positioned automatically on designated pages & slots
 */
export declare class UIHooks {
    readonly app: string;
    readonly verbose = false;
    protected set: Button[];
    constructor(opt: string | Partial<UIHooks>);
    /**
     * Takes one or more scripts and converts them to buttons
     *
     * With every script added it removes ALL
     * buttons that use the same script
     */
    Register(...scripts: Script[]): void;
    /**
     * Remove all buttons that match a script
     * @param name
     * @constructor
     */
    Unregister({ name }: Script): void;
    /**
     * Searches for buttons that match the requirements
     *
     * This is used in 2 kinds of places:
     *  - currated list of buttons in compose blocks where admin can
     *    picks, reorder, name and style scripts by hand
     *  - different slots on pages where scripts are automatically placed
     *
     * @param resourceType
     * @param page
     * @param slot
     * @constructor
     */
    Find(resourceType: string | string[], page?: string, slot?: string): Button[];
    FindByScript(script: string): Button | undefined;
}
export default function (): PluginFunction<Partial<UIHooks>>;
export {};
