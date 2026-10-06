import { getFieldFilter } from './record-filter'

// Helpers of the RelatedRecords page block: a block that lists, in one place,
// the records of other modules that refer to the record of the page.

const modeledOutside = ['datasource', 'connector']

function isOwnModule (module) {
  return !modeledOutside.includes(module?.config?.type)
}

// Every field, in any of the modules, that is a Record reference to moduleID -
// the candidates for a relation. The module itself is included (a tree of
// records refers to its own module). A module with several such fields gives
// one candidate for each of them, they are different relations.
export function findReferencingFields (modules = [], moduleID) {
  if (!moduleID) return []

  const out = []
  for (const module of modules) {
    if (!isOwnModule(module)) continue
    for (const field of module.fields || []) {
      if (field.kind === 'Record' && field.options && String(field.options.moduleID) === String(moduleID)) {
        out.push({ module, field })
      }
    }
  }

  return out.sort((a, b) => {
    const byModule = (a.module.name || a.module.handle || '').localeCompare(b.module.name || b.module.handle || '')
    return byModule || (a.field.label || a.field.name).localeCompare(b.field.label || b.field.name)
  })
}

// Two relations to the same module through different fields are different
export function relationKey (relation) {
  return `${relation.moduleID}:${relation.refField}`
}

export function hasRelation (relations = [], moduleID, refField) {
  return relations.some(r => r.moduleID === moduleID && r.refField === refField)
}

// The query that selects the child records of recordID
export function childrenQuery (field, recordID) {
  return getFieldFilter(field.name, 'Record', recordID, field.isMulti ? 'IN' : '=')
}

// The relations that can be shown: the module exists and still has the field.
// Others (a module or a field that was deleted after the block was set up) are
// left out rather than breaking the block.
export function usableRelations (relations = [], findModule) {
  const out = []
  for (const relation of relations) {
    const module = findModule(relation.moduleID)
    const field = module && (module.fields || []).find(f => f.name === relation.refField && f.kind === 'Record')
    if (module && field) out.push({ relation, module, field })
  }
  return out
}

// The block that renders one relation as a record list
export function relationListBlock (relation, extra = {}) {
  return {
    kind: 'RecordList',
    title: '',
    options: {
      moduleID: relation.moduleID,
      refField: relation.refField,
      fields: relation.fields || [],
      perPage: relation.perPage,
      presort: relation.presort,
      hideAddButton: relation.hideAddButton,
      hideHeader: true,
      ...extra,
    },
  }
}

// How many child records a section is asked for to count them. The server does
// not COUNT a query on a Record field (the values are JSON, a COUNT scans them)
// and answers total -1 - unless the page is not full, then the page is the
// answer. So the count is read from a page that is big enough for any usual
// section, and a section with more shows "N+".
export const COUNT_PAGE_SIZE = 100

// The number of records a list answer stands for: { count, more }.
// more is true when the real number is higher than count.
export function interpretCount (set = [], filter = {}, pageSize = COUNT_PAGE_SIZE) {
  const total = Number(filter.total)
  if (Number.isFinite(total) && total >= 0) return { count: total, more: false }

  const count = set.length
  return { count, more: count >= pageSize || !!filter.nextPage }
}

export function countLabel ({ count, more } = {}) {
  return more ? `${count}+` : String(count)
}
