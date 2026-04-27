import { AxiosInstance, AxiosRequestConfig } from 'axios';
interface KV {
    [header: string]: unknown;
}
interface Headers {
    [header: string]: string;
}
interface Ctor {
    baseURL?: string;
    accessTokenFn?: () => string | undefined;
    headers?: Headers;
}
export default class Discovery {
    protected baseURL?: string;
    protected accessTokenFn?: () => (string | undefined);
    protected headers: Headers;
    constructor({ baseURL, headers, accessTokenFn }: Ctor);
    setAccessTokenFn(fn: () => string | undefined): Discovery;
    setHeaders(headers?: Headers): Discovery;
    setHeader(name: string, value: string | undefined): Discovery;
    api(): AxiosInstance;
    query(a: KV, extra?: AxiosRequestConfig): Promise<KV>;
    queryCancellable(a: KV, extra?: AxiosRequestConfig): {
        response: () => Promise<KV>;
        cancel: () => void;
    };
}
export {};
