import { defineStore } from 'pinia'
import { system } from '@cortezaproject/corteza-js'
import {
  loadAllDraftsFromStorage,
  saveDraftToStorage,
  removeDraftFromStorage,
  clearAllDraftsFromStorage,
} from './storage'

const { Revision } = system

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
    async init ({ resourceType }: { resourceType?: string } = {}) {
      await this.loadAllDrafts({ resourceType })
    },

    async loadAllDrafts ({ resourceType }: { resourceType?: string } = {}) {
      this.loading = true
      try {
        await this.loadLocalDrafts()
      } finally {
        this.loading = false
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
    },

    async removeDraft ({ changeID }: { changeID: string }) {
      removeDraftFromStorage(changeID)
      const d = { ...this.drafts }
      delete d[changeID]
      this.drafts = d
    },

    async clearDrafts () {
      clearAllDraftsFromStorage()
      this.drafts = {}
    },

    toggleVisibility () {
      this.visible = !this.visible
    },
  },
})

export { getDraftFromStorage } from './storage'
