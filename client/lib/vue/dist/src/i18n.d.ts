import { VueConstructor } from 'vue';
import VueI18Next from '@panter/vue-i18next';
interface Options {
    app: string;
    resources: object;
    lng: string;
    fallbackLng: string | false;
    /**
     * Namespace(s) to preload
     */
    ns: string | Array<string>;
    /**
     * Where too look for keys that are not found in the default namespace,
     */
    fallbackNS: string | false;
    /**
     * What namespace to use when not explicitly defined
     * When empty, default is set to the value of (the first item in) ns
     */
    defaultNS: string;
    baseURL: string;
    pseudo: boolean;
}
/**
 * Initializes i18n options, registers plugin on a given Vue instance and returns the options
 *
 * To be used as:
 * import { i18n } from 'corteza-lib/vue/dist'
 * new Vue({
 *   i18n: i18n(Vue, {
 *     app: 'corteza-webapp-....'
 *     namespaces: [ .... ]
 *    }),
*   })
 *
 * The most convenient way to use it:
 * i18n(Vue, 'app name', 'namespace...', 'additional namespace...')
 */
declare const _default: (Vue: VueConstructor, app: string | Partial<Options>, ...namespaces: Array<string>) => VueI18Next;
export default _default;
