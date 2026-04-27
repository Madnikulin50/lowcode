import { AxiosInstance } from 'axios';
import { system } from '../../../../lib/js/dist';
import { PluginFunction } from 'vue';
declare const accessToken: unique symbol;
declare const user: unique symbol;
type eventListenerSignature = <K extends keyof WindowEventMap>(type: K, listener: (this: Window, ev: WindowEventMap[K]) => any, options?: boolean | AddEventListenerOptions) => void;
interface AuthInfo {
    accessTokenFn: () => string | undefined;
    user: system.User;
}
interface PluginOpts {
    cortezaAuthURL: string;
    callbackURL: string;
}
interface AuthCtor {
    app: string;
    /**
     * when true, use console as a logger, no-op otherwise.
     */
    verbose: boolean;
    /**
     * where the authe backend is
     */
    cortezaAuthURL: string;
    /**
     * URL we'll be listening to for callbacksa
     */
    callbackURL: string;
    /**
     * used for redirection
     */
    location: Location;
    /**
     * used for storing
     */
    sessionStorage: Storage;
    /**
     * used for event listeners
     */
    registerEventListener: eventListenerSignature;
    /**
     * Static string with entry-point URL stored at app init
     * so that there is no risk of changes when Vue router gets it's hands on it
     */
    entrypointURL: string;
    /**
     * multiply factor for token expiration
     * this will tell internal refresh system how much
     * before the token expiration we'll refresh the access toke
     *
     * keep in mind that access token is exchanged on every app load
     */
    refreshFactor: number;
}
interface Logger {
    debug(...data: unknown[]): void;
    info(...data: unknown[]): void;
    error(...data: unknown[]): void;
}
export declare class Auth {
    /**
     * Access token is only stored here (in-memory)!
     * we do not want to keep it in the local store
     */
    private [accessToken]?;
    /**
     * Access token is only stored here (in-memory)!
     * we do not want to keep it in the local store
     */
    private [user]?;
    /**
     * Name of the app that is using the auth plugin
     */
    readonly app: string;
    readonly refreshFactor: number;
    readonly verbose: boolean;
    readonly cortezaAuthURL: string;
    readonly callbackURL: string;
    readonly location: Location;
    readonly sessionStorage: Storage;
    readonly registerEventListener: eventListenerSignature;
    /**
     * Application entrypoint URL
     */
    readonly entrypointURL: string;
    /**
     * Keeps track of timeout callback in case we re-run it before it timesout
     * @private
     */
    private refreshTimeout?;
    private expiresIn;
    private listenersBound;
    private $emit?;
    constructor({ app, verbose, cortezaAuthURL, callbackURL, entrypointURL, location, sessionStorage, refreshFactor, registerEventListener }: AuthCtor);
    vue(vue: Vue): Auth;
    get axios(): AxiosInstance;
    /**
     * wrapper for console (when in debug mode) or a simple no-op obj
     */
    get log(): Logger;
    /**
     * Returns function that returns current access token
     */
    get accessTokenFn(): () => string | undefined;
    /**
     * Handles initial authentication check
     *
     * handle function should be called immediately when application is created
     * it checks whether app was requested on an URL with /auth/callback at the end
     * if there is an error or code passed and handles that request appropriately:
     *
     *  .../auth/callback?code=... exchanged authorization code for access token
     *  .../auth/callback?error=... renders an error that we got from the oauth2 provider
     *
     * If handle was called without /auth/callback or without params mentioned above:
     *   if user is not authorized, redirect to the configured path to start oauth2 flow
     *   if user is authorized, continue with execution
     */
    handle(req?: URL): Promise<AuthInfo | null>;
    /**
     * Flagging session storage
     *
     * Challenge with sessions and browser tabs:
     * Each browser tab & window interacts with an isolated session. When user clicks on a link to and wants to open
     * it in a new window or a tab, or when tab is duplicated, session contents are copied!
     *
     * Consequences of that are that two (or more) tabs end up with the same session
     * and the same refresh token, and we need to detect if we're dealing with refresh token in an old or new  session.
     *
     * With this function we start the final state and flag the session. This way we'll
     * know, after the redirection to the final location, if this is a final stage or not and if the refresh token
     * belongs to this session or not.
     */
    handleStateManagement(): boolean;
    bindListeners(): void;
    cleanFlags(): void;
    /**
     * Called when refresh token is re-fetched.
     *
     * Cleanup aux items in the session store
     */
    completeFinalState(): void;
    /**
     * Exchanges the auth parameters for access & refresh token.
     *
     * If the parameters are correct and exchange is successful, the refresh token
     * gets stored in localStorage for further use, the access token and current user get stored
     * in-memory.
     *
     * Function will throw null when user is unauthenticated
     */
    handleCallbackRoute(state: string | null, code: string): Promise<AuthInfo | null>;
    /**
     * Checks current auth state; is access token loaded OR do we have a refresh token we can use
     *
     * check uses system API client verify given/current JWT
     *
     * If JWT is valid, it is stored into local storage alongside
     * loaded user.
     *
     * We're explicitly passing systemAPI to minimize plugin initialization complexity
     *
     * Function will throw null when user is unauthenticated
     */
    handleState(): Promise<AuthInfo | null>;
    logout(): void;
    /**
     * Starts new authentication flow
     *
     * It generates simple rand state to harden security and to
     * keep track of before-flow-start location of the user
     */
    startAuthenticationFlow(): void;
    getRedirect(url: string): string;
    isCallback(url: string): boolean;
    /**
     * protects against too many tries when we try to auto-fix the "state does not match" error
     * by restarting the aut flow.
     */
    private incFlowCounter;
    startAutoLogout(): Promise<number>;
    stopAutoLogout(): Promise<AuthInfo | null>;
    /**
     * Exchanges authorization code for access and refresh tokens
     */
    private exchangeCode;
    /**
     * Exchanges refresh token for new access and new refresh token
     *
     * After successful token exchange, we call response processing function
     * to update internals & stored values
     *
     * @param refreshToken
     */
    private exchangeRefresh;
    /**
     * Processes fetched token and stores it
     *
     * Access token is stored only to instance of this object
     * Refresh token is stored only to local store
     *
     * @param oa2tkn OAuth2 token response
     * @private
     */
    private procTokenResponse;
    /**
     * oauth2token exchanges authorization code or refresh token for (new) access token
     *
     * @param payload
     * @private
     */
    private oauth2token;
    private pruneStore;
    get accessToken(): string | undefined;
    get user(): system.User | undefined;
}
export default function (): PluginFunction<PluginOpts>;
export {};
