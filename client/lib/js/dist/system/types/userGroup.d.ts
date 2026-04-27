interface PartialUserGroup extends Partial<Omit<UserGroup, 'createdAt' | 'updatedAt' | 'deletedAt' | 'suspendedAt'>> {
    createdAt?: string | number | Date;
    updatedAt?: string | number | Date;
    deletedAt?: string | number | Date;
    suspendedAt?: string | number | Date;
}
interface Meta {
    description: string;
    short: string;
}
interface Config {
    path: {
        selfID: string;
        name: string;
    }[];
}
export declare class UserGroup {
    userGroupID: string;
    handle: string;
    isRoot: boolean;
    config: Config;
    meta: Meta;
    labels: object;
    canGrant: boolean;
    canUpdateUserGroup: boolean;
    canDeleteUserGroup: boolean;
    canManageMembersOnUserGroup: boolean;
    createdAt?: Date;
    updatedAt?: Date;
    deletedAt?: Date;
    suspendedAt?: Date;
    roles?: Array<string>;
    constructor(u?: PartialUserGroup);
    apply(u?: PartialUserGroup): void;
    /**
     * Returns resource ID
     */
    get resourceID(): string;
    /**
     * Resource type
     */
    get resourceType(): string;
    get fts(): string;
    clone(): UserGroup;
    properties(): string[];
}
export {};
