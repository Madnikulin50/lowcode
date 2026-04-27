import { automation } from '../../../../../lib/js/dist';
export declare function pVal<T = unknown>(vars: automation.Vars, k: string, def?: T): T | undefined;
export declare function pType(vars: automation.Vars, k: string, def?: string): string | undefined;
