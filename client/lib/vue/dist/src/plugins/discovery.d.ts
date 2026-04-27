import { PluginFunction } from 'vue';
interface Options {
    baseURL?: string;
    accessTokenFn?: () => string | undefined;
}
/**
 * Corteza Discovery API plugin
 *
 * Install:
 * Vue.use(plugins.DiscoveryAPI())
 *
 * @constructor
 */
export default function (opt?: Options): PluginFunction<Options>;
export {};
