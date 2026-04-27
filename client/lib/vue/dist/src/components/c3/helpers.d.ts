export interface CatalogueItem {
    group?: Array<string>;
    name?: string;
}
export interface Catalogue {
    [_: string]: CatalogueItem;
}
/**
 * Extract subgroups from the given groups
 * @param groups
 * @param path Array of groups that represent the path
 */
export declare function extractGroups(groups: Array<Array<string>>, ...path: Array<string>): Array<string>;
/**
 * Extract subgroups from the catalogue that match the given path
 *
 * @param cat Catalogue
 * @param path Array of groups that represent the path
 * @constructor
 */
export declare function ExtractSubgroups(cat: Catalogue, ...path: Array<string>): Array<string>;
/**
 * Returns all components that match the given path
 *
 * @param cat Catalogue
 * @param path Array of groups that represent the path

 */
export declare function ExtractComponents(cat: Catalogue, ...path: Array<string>): Array<CatalogueItem>;
