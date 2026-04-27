import { apiClients, system } from '../../../../../lib/js/dist';
import { StoreOptions } from 'vuex';
interface Options {
    api: apiClients.System;
    ws: WebSocket;
    watchInterval: number;
    webapp: string;
}
interface State {
    notifications: Array<system.Notification>;
    visible: boolean;
    pageCursor: string | null;
    muted: boolean;
}
export default function ({ api }: Options): StoreOptions<State>;
export {};
