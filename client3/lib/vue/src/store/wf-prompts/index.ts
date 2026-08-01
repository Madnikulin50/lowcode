import { defineStore } from 'pinia'
import { apiClients, automation } from 'corteza-lib/js/dist'
import { promptDefinitions } from '../../components/prompts'

function loadPrompts (api: apiClients.Automation): Promise<Array<automation.Prompt>> {
  return api.sessionListPrompts().then(({ set } = {}) => {
    if (!Array.isArray(set)) {
      return []
    }
    return set.map((p: automation.Prompt) => new automation.Prompt(p))
  })
}

async function resumeState (api: apiClients.Automation, { sessionID, stateID }: automation.Prompt, input: automation.Vars): Promise<unknown> {
  return api.sessionResumeState({ sessionID, stateID, input })
}

async function cancelState (api: apiClients.Automation, { sessionID, stateID }: automation.Prompt): Promise<unknown> {
  return api.sessionCancel({ sessionID, stateID })
}

function onlyFresh (existing: Array<automation.Prompt>, fresh: Array<automation.Prompt>): Array<automation.Prompt> {
  const index = existing.map(({ stateID }) => stateID)
  return fresh.filter(({ stateID = undefined }) => stateID && !index.includes(stateID))
}

export const useWfPromptsStore = defineStore('wfPrompts', {
  state: () => ({
    loading: false,
    active: false as automation.Prompt | boolean,
    prompts: [] as Array<automation.Prompt>,
  }),

  getters: {
    all: (state) => state.prompts,
    isLoading: (state) => state.loading,
    isActive: (state) => state.active !== false,
    current: (state) => {
      if (typeof state.active === 'boolean') {
        return undefined
      } else {
        return state.active
      }
    },
  },

  actions: {
    async activate (m?: automation.Prompt | true) {
      this.active = m ?? true
    },

    async deactivate () {
      this.active = false
    },

    async update () {
      const api = (window as any).__automationAPI as apiClients.Automation | undefined
      if (!api) return
      return loadPrompts(api).then(pp => {
        if (pp.length === 0) {
          this.prompts = []
          return
        }
        const fresh = onlyFresh(this.prompts, pp)
        if (fresh.length > 0) {
          this.prompts.push(...fresh)
        }
      })
    },

    new (prompt: automation.Prompt) {
      this.prompts.push(prompt)
    },

    async resume ({ prompt, input }: { prompt: automation.Prompt; input: automation.Vars }) {
      this.loading = true
      const api = (window as any).__automationAPI as apiClients.Automation | undefined
      if (!api) return
      return resumeState(api, prompt, input)
        .then(() => { this.remove(prompt) })
        .catch(() => { this.remove(prompt) })
        .finally(() => { this.loading = false })
    },

    async cancel (prompt: automation.Prompt) {
      this.loading = true
      const api = (window as any).__automationAPI as apiClients.Automation | undefined
      if (!api) return
      return cancelState(api, prompt)
        .then(() => { this.remove(prompt) })
        .finally(() => { this.loading = false })
    },

    clear (prompt: automation.Prompt) {
      this.remove(prompt)
    },

    remove (prompt: automation.Prompt) {
      this.prompts = this.prompts.filter(({ stateID }) => stateID !== prompt.stateID)
      if (typeof this.active === 'object' && this.active.stateID === prompt.stateID) {
        this.active = this.prompts.length > 0
      }
    },
  },
})
