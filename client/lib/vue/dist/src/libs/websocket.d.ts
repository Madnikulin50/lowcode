/**
 * Default websocket configuration
 */
import { Vue } from 'vue/types/vue';
export declare const config: {
    format: string;
    reconnection: boolean;
    reconnectionAttempts: number;
    reconnectionDelay: number;
    connectManually: boolean;
};
/**
 * Extract websocket endpoint from window props (set via config.js)
 */
export declare function endpoint(): string;
/**
 * Binds auth and websocket events so that we can pass current access token
 *  - when ws connection opens
 *  - when auth token in fetched/renewed
 *
 *  @todo get rid of ts-ignore lines
 */
export declare function init(vue: Vue): void;
