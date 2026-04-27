import { PluginFunction } from 'vue';
interface Options {
    baseURL?: string;
    accessTokenFn?: () => string | undefined;
}
/**
 * Generic Corteza API plugin
 *
 * Install a specific plugin:
 * Vue.use(plugins.CortezaAPI('compose'))
 *
 * @constructor
 */
export default function (service: string, opt?: Options): PluginFunction<Options>;
export {};
