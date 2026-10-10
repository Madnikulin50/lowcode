import { defineStore } from 'pinia'
import { system } from 'corteza-lib/js/dist'
import {
  loadAllDraftsFromStorage,
  saveDraftToStorage,
  removeDraftFromStorage,
  clearAllDraftsFromStorage,
} from './storage'

import { parseDraftResource, mergeDrafts } from './sync'

const { Revision } = system

// The compose API client, when drafts should also be kept on the server (set by
// init). Without it - or while the server is out of reach - drafts live in the
// browser only, as they always did.
interface DraftsAPI {
  recordDraftList: (a?: Record<string, unknown>) => Promise<{ set?: Array<{ changeID: string; payload: unknown }> }>
  recordDraftSave: (a: Record<string, unknown>) => Promise<unknown>
  recordDraftDelete: (a: Record<string, unknown>) => Promise<unknown>
  recordDraftClear: () => Promise<unknown>
}

let backend: DraftsAPI | undefined

// edits come in fast (the record form saves a draft every couple of seconds):
// the server gets the last state of a draft once it has been quiet for a moment
const pushDelayMs = 1500
const pending = new Map<string, ReturnType<typeof setTimeout>>()

function hasDraftsAPI (api: unknown): api is DraftsAPI {
  const a = api as Partial<DraftsAPI> | undefined
  return !!a && typeof a.recordDraftList === 'function' && typeof a.recordDraftSave === 'function'
}

// pushDraft sends the draft to the server; false when it could not be sent
async function pushDraft (revision: system.Revision): Promise<boolean> {
  const ref = parseDraftResource(revision.resource)
  if (!backend || !ref) return false

  try {
    await backend.recordDraftSave({ changeID: String(revision.changeID), ...ref, payload: revision })
    return true
  } catch (e) {
    console.warn('Failed to keep the draft on the server:', revision.changeID, e)
    return false
  }
}

export interface DraftEntry {
  revision: system.Revision
  source: 'local' | 'backend'
}

export const useDraftsStore = defineStore('drafts', {
  state: () => ({
    drafts: {} as { [key: string]: DraftEntry },
    loading: false,
    visible: false,
  }),

  getters: {
    getDraft: (state) => (changeID: string): DraftEntry | undefined => {
      return state.drafts[changeID]
    },
    hasDraft: (state) => (changeID: string): boolean => {
      return !!state.drafts[changeID]
    },
    getAllDrafts: (state): DraftEntry[] => {
      return Object.values(state.drafts)
    },
    getAllDraftsMap: (state): { [key: string]: DraftEntry } => {
      return state.drafts
    },
    getDraftsByResourceType: (state) => (resourceType: string): DraftEntry[] => {
      return Object.values(state.drafts).filter(
        entry => entry.revision.resource.startsWith(resourceType),
      )
    },
    getDraftsBySource: (state) => (source: 'local' | 'backend'): DraftEntry[] => {
      return Object.values(state.drafts).filter(entry => entry.source === source)
    },
    getDraftsForRecord: (state) => (recordID: string): DraftEntry[] => {
      return Object.values(state.drafts).filter(
        entry => entry.revision.resource.endsWith(`/${recordID}`),
      )
    },
    isLoading: (state): boolean => state.loading,
  },

  actions: {
    async init ({ resourceType, composeAPI }: { resourceType?: string; composeAPI?: unknown } = {}) {
      backend = hasDraftsAPI(composeAPI) ? composeAPI : undefined
      await this.loadAllDrafts({ resourceType })
    },

    async loadAllDrafts ({ resourceType }: { resourceType?: string } = {}) {
      this.loading = true
      try {
        await this.loadLocalDrafts()
        await this.syncWithBackend()
      } finally {
        this.loading = false
      }
    },

    // Brings in the drafts kept on the server and sends up the local ones the
    // server does not have (or has an older copy of). A server that cannot be
    // reached leaves the local drafts as they are.
    async syncWithBackend () {
      if (!backend) return

      let remote: system.Revision[]
      try {
        const { set } = await backend.recordDraftList()
        remote = (set || []).map(item => {
          const r = new Revision(item.payload as ConstructorParameters<typeof Revision>[0])
          if (!r.changeID) r.changeID = item.changeID
          return r
        })
      } catch (e) {
        console.warn('Failed to load drafts from the server:', e)
        return
      }

      const local = Object.values(this.drafts)
        .filter(entry => entry.source === 'local')
        .map(entry => entry.revision)

      const { merged, toUpload } = mergeDrafts(local, remote)

      const next = { ...this.drafts }
      for (const [changeID, { draft, source }] of Object.entries(merged)) {
        next[changeID] = { revision: draft as system.Revision, source }
        // keep a copy in the browser too, so the draft survives losing the connection
        saveDraftToStorage(changeID, draft as system.Revision)
      }
      this.drafts = next

      for (const revision of toUpload) {
        if (await pushDraft(revision)) {
          const changeID = String(revision.changeID)
          this.drafts = { ...this.drafts, [changeID]: { revision, source: 'backend' } }
        }
      }
    },

    loadLocalDrafts () {
      const localDrafts = loadAllDraftsFromStorage()
      localDrafts.forEach((revisionData, changeID) => {
        const revision = new Revision(revisionData)
        this.drafts = {
          ...this.drafts,
          [String(revision.changeID)]: {
            revision,
            source: 'local' as const,
          },
        }
      })
    },

    saveDraft ({ revision }: { revision: system.Revision }) {
      const changeID = String(revision.changeID)
      saveDraftToStorage(changeID, revision)
      this.drafts = { ...this.drafts, [changeID]: { revision, source: 'local' } }

      if (backend) {
        clearTimeout(pending.get(changeID))
        pending.set(changeID, setTimeout(async () => {
          pending.delete(changeID)
          // the draft may have been removed or edited again while waiting
          const current = this.drafts[changeID]
          if (!current) return

          if (await pushDraft(current.revision) && this.drafts[changeID]) {
            this.drafts = { ...this.drafts, [changeID]: { revision: this.drafts[changeID].revision, source: 'backend' } }
          }
        }, pushDelayMs))
      }
    },

    async removeDraft ({ changeID }: { changeID: string }) {
      clearTimeout(pending.get(changeID))
      pending.delete(changeID)

      removeDraftFromStorage(changeID)
      const d = { ...this.drafts }
      delete d[changeID]
      this.drafts = d

      if (backend) {
        try {
          await backend.recordDraftDelete({ changeID })
        } catch (e) {
          console.warn('Failed to delete the draft on the server:', changeID, e)
        }
      }
    },

    async clearDrafts () {
      pending.forEach(clearTimeout)
      pending.clear()

      clearAllDraftsFromStorage()
      this.drafts = {}

      if (backend) {
        try {
          await backend.recordDraftClear()
        } catch (e) {
          console.warn('Failed to delete the drafts on the server:', e)
        }
      }
    },

    toggleVisibility () {
      this.visible = !this.visible
    },
  },
})

export { getDraftFromStorage } from './storage'
