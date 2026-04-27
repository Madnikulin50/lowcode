import { Component } from 'vue';
import { automation } from '../../../../../../lib/js/dist';
interface Handler {
    (this: Component, input: automation.Vars): void | Promise<void>;
}
interface PromptDefinition {
    component?: Component;
    /**
     *
     */
    handler?: Handler;
    /**
     * Passive prompt, will not be listed
     *
     * Also, when displaying toasts, we'll display all
     * passive toasts first and then, at the and one single
     * non-passive toast
     */
    passive?: boolean;
}
declare const definitions: Record<string, PromptDefinition>;
export default definitions;
