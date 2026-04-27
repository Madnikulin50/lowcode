import { system } from '../../../../../lib/js/dist';
export interface DraftEntry {
    revision: system.Revision;
    source: 'local' | 'backend';
}
interface DraftsState {
    drafts: {
        [key: string]: DraftEntry;
    };
    loading: boolean;
    visible: boolean;
}
declare const types: {
    SET_DRAFT: string;
    REMOVE_DRAFT: string;
    SET_LOADING: string;
    setVisible: string;
    CLEAR_DRAFTS: string;
};
export default function (): {
    namespaced: boolean;
    state: () => DraftsState;
    getters: {
        getDraft: (state: DraftsState) => (changeID: string) => DraftEntry | undefined;
        hasDraft: (state: DraftsState) => (changeID: string) => boolean;
        getAllDrafts: (state: DraftsState) => DraftEntry[];
        getAllDraftsMap: (state: DraftsState) => {
            [key: string]: DraftEntry;
        };
        getDraftsByResourceType: (state: DraftsState) => (resourceType: string) => DraftEntry[];
        getDraftsBySource: (state: DraftsState) => (source: "local" | "backend") => DraftEntry[];
        getDraftsForRecord: (state: DraftsState) => (recordID: string) => DraftEntry[];
        isLoading: (state: DraftsState) => boolean;
        visible: (state: DraftsState) => boolean;
    };
    actions: {
        init({ dispatch }: {
            dispatch: (action: string, payload?: any) => Promise<void>;
        }, { resourceType }: {
            resourceType?: string;
        }): Promise<void>;
        loadAllDrafts({ commit, dispatch }: {
            commit: (mutation: string, payload?: any) => void;
            dispatch: (action: string, payload?: any) => Promise<void>;
        }, { resourceType }?: {
            resourceType?: string;
        }): Promise<void>;
        loadLocalDrafts({ commit }: {
            commit: (mutation: string, payload?: any) => void;
        }): void;
        saveDraft({ commit }: {
            commit: (mutation: string, payload?: any) => void;
        }, { revision }: {
            revision: system.Revision;
        }): void;
        removeDraft({ commit }: {
            commit: (mutation: string, payload?: any) => void;
        }, { changeID }: {
            changeID: string;
        }): Promise<void>;
        clearDrafts({ commit }: {
            commit: (mutation: string, payload?: any) => void;
        }): Promise<void>;
        toggleVisibility({ commit, state }: {
            commit: (mutation: string, payload?: any) => void;
            state: DraftsState;
        }): void;
    };
    mutations: {
        [types.SET_DRAFT]: (state: DraftsState, { changeID, entry }: {
            changeID: string;
            entry: DraftEntry;
        }) => void;
        [types.REMOVE_DRAFT]: (state: DraftsState, changeID: string) => void;
        [types.CLEAR_DRAFTS]: (state: DraftsState) => void;
        [types.SET_LOADING]: (state: DraftsState, loading: boolean) => void;
        [types.setVisible]: (state: DraftsState, visible: boolean) => void;
    };
};
export { getDraftFromStorage } from './storage';
