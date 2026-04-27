import { apiClients, automation } from '../../../../../lib/js/dist';
import { StoreOptions } from 'vuex';
interface Options {
    api: apiClients.Automation;
    ws: WebSocket;
    watchInterval: number;
    webapp: string;
}
interface State {
    loading: boolean;
    prompts: Array<automation.Prompt>;
    /**
     * Is prompt component active (modal open)?
     *   prompt = modal is open, show this/current prompt
     *   true   = modal is open, show list of pending prompts
     *   false  = modal is closed
     */
    active: automation.Prompt | boolean;
}
export default function ({ api, webapp }: Options): StoreOptions<State>;
export {};
