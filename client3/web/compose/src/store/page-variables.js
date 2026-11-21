import { defineStore } from 'pinia'

// Session-only runtime values for page variables (PageBlocks/Variables).
// Not persisted (no localStorage, no API) — resets on reload, by design.
export const usePageVariablesStore = defineStore('pageVariables', {
  state: () => ({
    // { [pageID]: { [variableName]: rawValue } }
    valuesByPage: {},
  }),

  getters: {
    getValue: (state) => (pageID, name) => {
      const bag = state.valuesByPage[pageID]
      return bag ? bag[name] : undefined
    },

    getValuesForPage: (state) => (pageID) => state.valuesByPage[pageID] || {},
  },

  actions: {
    setValue (pageID, name, value) {
      if (!this.valuesByPage[pageID]) {
        this.valuesByPage[pageID] = {}
      }
      this.valuesByPage[pageID][name] = value
    },

    // Returns true if a default was actually written (i.e. no value was set
    // yet for this page+name), so callers can decide whether sibling blocks
    // need to be notified.
    setDefaultIfUnset (pageID, name, defaultValue) {
      if (!this.valuesByPage[pageID]) {
        this.valuesByPage[pageID] = {}
      }
      if (this.valuesByPage[pageID][name] === undefined) {
        this.valuesByPage[pageID][name] = defaultValue
        return true
      }
      return false
    },

    clearPage (pageID) {
      delete this.valuesByPage[pageID]
    },
  },
})
