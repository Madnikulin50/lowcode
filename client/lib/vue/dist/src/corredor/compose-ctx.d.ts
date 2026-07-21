import { compose, apiClients, corredor } from '../../../../lib/js/dist';
import ComposeUIHelper from './compose-ui';
interface Vue {
    $SystemAPI: apiClients.System;
    $ComposeAPI: apiClients.Compose;
    $store: {
        getters: {
            [_: string]: Array<compose.Page>;
        };
    };
    $emit: unknown;
    $router: {
        push: unknown;
    };
}
/**
 * Extends corredor exec context with compose UI helper
 */
export default class ComposeCtx extends corredor.Ctx {
    protected emitter: unknown;
    protected routePusher: unknown;
    protected pages: Array<compose.Page>;
    protected composeUI: ComposeUIHelper;
    protected vue: Vue;
    constructor(args: corredor.BaseArgs, vue: Vue);
    /**
     * Clones context and uses new arguments
     */
    withArgs(args: corredor.BaseArgs): ComposeCtx;
    get ComposeUI(): ComposeUIHelper;
}
export {};
