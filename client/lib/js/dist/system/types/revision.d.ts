export type RevisionStatus = '' | 'draft';
export type RevisionOperation = 'created' | 'updated' | 'soft-deleted' | 'undeleted' | 'hard-deleted';
export interface RevisionChange {
    key: string;
    old: unknown[];
    new: unknown[];
}
interface PartialRevision extends Partial<Omit<Revision, 'timestamp' | 'deletedAt'>> {
    timestamp?: string | number | Date;
    deletedAt?: string | number | Date;
}
export declare class Revision {
    changeID: string;
    timestamp?: Date;
    resource: string;
    revision: number;
    operation: RevisionOperation;
    status: RevisionStatus;
    userID: string;
    changes: RevisionChange[];
    comment: string;
    deletedAt?: Date;
    deletedBy: string;
    record?: unknown;
    constructor(r?: PartialRevision);
    apply(r?: PartialRevision): void;
    get resourceIdentifier(): string;
    get isDraft(): boolean;
    clone(): Revision;
}
export {};
