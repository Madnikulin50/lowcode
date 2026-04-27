interface Meta {
    name?: string;
    description?: string;
    visual?: Record<string, unknown>;
    subWorkflow?: boolean;
}
interface PartialWorkflow extends Partial<Omit<Workflow, 'createdAt' | 'updatedAt' | 'deletedAt' | 'suspendedAt'>> {
    meta?: Partial<Meta>;
    createdAt?: string | number | Date;
    updatedAt?: string | number | Date;
    deletedAt?: string | number | Date;
}
export declare class Workflow {
    workflowID: string;
    handle: string;
    enabled: boolean;
    trace: boolean;
    keepSessions: number;
    labels: Record<string, string>;
    meta: Meta;
    scope?: Record<string, unknown>;
    steps?: unknown[];
    paths?: unknown[];
    issues?: unknown[];
    runAs: string;
    ownedBy: string;
    createdBy: string;
    updatedBy: string;
    deletedBy: string;
    createdAt?: Date;
    updatedAt?: Date;
    deletedAt?: Date;
    canGrant: boolean;
    canUpdateWorkflow: boolean;
    canDeleteWorkflow: boolean;
    canExecuteWorkflow: boolean;
    constructor(w?: PartialWorkflow);
    apply(w?: PartialWorkflow): void;
    /**
     * Returns resource ID
     */
    get resourceID(): string;
    /**
     * Resource type
     */
    get resourceType(): string;
}
export {};
