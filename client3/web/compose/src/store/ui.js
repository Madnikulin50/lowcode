import { defineStore } from 'pinia'
import { getComposeAPI } from './api'

export const useUiStore = defineStore('ui', {
  state: () => ({
    loading: 0,
    pending: false,
    recordPaginationIDs: [],
    recordPaginationUsable: false,

    previousPages: [],
    previousPage: null,
    modalPreviousPages: [],

    namespaceSlug: '',
    pageHandle: '',
    layoutHandle: '',

    modalPageHandle: '',
    modalLayoutHandle: '',

    layoutRequiredFields: [],
  }),

  getters: {
    getNextAndPrevRecord: ({ recordPaginationIDs }) => (recordID) => {
      const recordIndex = recordPaginationIDs.indexOf(recordID)
      const prev = recordIndex >= 0 ? recordPaginationIDs[recordIndex - 1] : undefined
      const next = recordIndex >= 0 ? recordPaginationIDs[recordIndex + 1] : undefined

      return { prev, next }
    },

    isFieldRequiredByLayout: (state) => (fieldID) => state.layoutRequiredFields.includes(fieldID),
  },

  actions: {
    incLoader () {
      this.loading++
    },

    decLoader () {
      this.loading--
    },

    async loadPaginationRecords ({ filter } = {}) {
      this.pending = true
      this.recordPaginationUsable = true

      const { prevPage, pageCursor, nextPage } = filter

      const cursors = new Set([prevPage, pageCursor, nextPage])

      const ComposeAPI = getComposeAPI()
      return Promise.all([...cursors].map(pageCursor => {
        return ComposeAPI.recordList({ ...filter, pageCursor })
          .then(({ set }) => {
            return set.map(({ recordID }) => recordID)
          })
      })).then(([...records]) => {
        this.recordPaginationIDs = [...new Set(records.flatMap(r => r))]
      }).finally(() => {
        this.pending = false
      })
    },

    clearRecordPagination () {
      this.recordPaginationIDs = []
    },

    setRecordPaginationUsable (value) {
      this.recordPaginationUsable = value
    },

    setPreviousPages (value) {
      this.previousPages = value
    },

    pushPreviousPages (value) {
      this.previousPages.push(value)
    },

    popPreviousPages () {
      const previousPage = this.previousPages.slice(-1)[0]
      this.previousPages.pop()
      return new Promise((resolve) => resolve(previousPage))
    },

    setPreviousPage (value) {
      const shouldNotSavePage = value.name !== 'admin.pages.builder' &&
            !value.query.layoutID && value.name !== 'admin.modules.create' &&
              value.name !== 'admin.charts.create' &&
                value.name !== 'namespace.create'

      if (value && value.name && shouldNotSavePage) {
        this.previousPage = value
      }
    },

    pushModalPreviousPage (value) {
      const previousPage = this.modalPreviousPages[this.modalPreviousPages.length - 1] || {}
      if (previousPage.recordID === value.recordID && previousPage.recordPageID === value.recordPageID) {
        return
      }

      this.modalPreviousPages.push(value)
    },

    clearModalPreviousPage () {
      this.modalPreviousPages = []
    },

    popModalPreviousPage () {
      const previousPage = this.modalPreviousPages[this.modalPreviousPages.length - 2] || {}
      this.modalPreviousPages.pop()
      return new Promise((resolve) => resolve(previousPage))
    },

    setNamespaceSlug (value) {
      this.namespaceSlug = value || ''
    },

    setPageHandle (value) {
      this.pageHandle = value || ''
    },

    setLayoutHandle (value) {
      this.layoutHandle = value || ''
    },

    setModalPageHandle (value) {
      this.modalPageHandle = value || ''
    },

    setModalLayoutHandle (value) {
      this.modalLayoutHandle = value || ''
    },

    setLayoutRequiredFields (fields) {
      this.layoutRequiredFields = fields || []
    },

    clearLayoutRequiredFields () {
      this.layoutRequiredFields = []
    },
  },
})
