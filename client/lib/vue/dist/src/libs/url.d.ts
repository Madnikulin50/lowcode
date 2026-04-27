/**
 * Little URL helper
 *
 * We need it to handle relative URLs, especially ones w/o schema
 */
export declare function Make({ url, query, hash, ref, config }: {
    url?: string | undefined;
    query?: {} | undefined;
    hash?: string | undefined;
    ref?: string | undefined;
    config?: {} | undefined;
}): string;
