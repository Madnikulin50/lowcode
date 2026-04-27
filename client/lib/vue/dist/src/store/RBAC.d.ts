import { StoreOptions } from 'vuex';
interface Rule {
    resource: string;
    operation: string;
    allow: boolean;
}
interface State {
    loaded: boolean;
    rules: Array<Rule>;
}
interface Fetcher {
    permissionsEffective: () => Promise<Array<Rule>>;
}
export default function (...apis: Array<Fetcher>): StoreOptions<State>;
export {};
