import { Component } from 'vue';
interface Handler {
    value(props: object): unknown;
    update(props: object, value: unknown): void;
}
interface Props {
    [_: string]: unknown;
}
interface Control extends Handler {
    component: Component;
    props?: Props;
}
interface Specs {
    handler: string | Handler;
    props?: Props;
}
export declare function generic(component: Component, { handler, props }: Specs): Control;
export declare function input(label: string, handler: string | Handler): Control;
export declare function textarea(label: string, handler: string | Handler): Control;
export declare function checkbox(label: string, handler: string | Handler): Control;
export declare function select(label: string, handler: string | Handler, options: object): Control;
export {};
