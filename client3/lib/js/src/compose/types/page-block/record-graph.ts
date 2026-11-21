import { PageBlock, PageBlockInput, Registry } from './base'
import { Apply, CortezaID, NoID } from '../../../cast'

const kind = 'RecordGraph'

interface Options {
  // how many levels around the record are loaded when the block appears
  // (0: the record alone, the rest is opened by clicking nodes)
  depth: number;
  // upper bound for nodes on the graph; more records are left out
  maxNodes: number;
  // records of one module loaded for a node at a time
  perModuleLimit: number;
  // modules whose records are not shown
  excludeModules: string[];
  // follow the Record fields of a record to the records it refers to
  showParents: boolean;
  // follow the Record fields of other modules to the records that refer to it
  showChildren: boolean;
  showLegend: boolean;
  // how the records of a module are named on the graph: a template with the
  // fields in double curly braces, "{{position_number}} · {{work_name}}"
  // and, for the modules that have one, the Select field whose colour rings the node
  labels: Array<{ moduleID: string; template: string; statusField: string }>;
  // how the nodes are laid out: a force simulation, rings around the record, or
  // layers (what the record refers to above, what refers to it below)
  layout: 'force' | 'radial' | 'layers';
  // from how many records of one module a node stands for them all (0: never)
  clusterFrom: number;
  showToolbar: boolean;
  // what "open the record" does for a selected node, as in other blocks:
  // in the same tab, a new tab or a modal; doNothing turns opening off
  displayOption: 'sameTab' | 'newTab' | 'modal' | 'doNothing';
}

export const displayOptions = Object.freeze(['sameTab', 'newTab', 'modal', 'doNothing'])
export const layouts = Object.freeze(['force', 'radial', 'layers'])

export const limits = Object.freeze({
  depth: { min: 0, max: 2 },
  maxNodes: { min: 10, max: 500 },
  perModuleLimit: { min: 1, max: 100 },
  clusterFrom: { min: 0, max: 50 },
})

const defaults: Readonly<Options> = Object.freeze({
  depth: 1,
  maxNodes: 150,
  perModuleLimit: 20,
  excludeModules: [],
  showParents: true,
  showChildren: true,
  showLegend: true,
  labels: [],
  displayOption: 'sameTab',
  layout: 'force',
  clusterFrom: 8,
  showToolbar: true,
})

const maxTemplate = 200

function clamp (n: unknown, { min, max }: { min: number; max: number }, fallback: number): number {
  const v = Math.trunc(Number(n))
  if (!Number.isFinite(v)) return fallback
  return Math.min(max, Math.max(min, v))
}

export class PageBlockRecordGraph extends PageBlock {
  readonly kind = kind

  options: Options = { ...defaults, excludeModules: [], labels: [] }

  constructor (i?: PageBlockInput) {
    super(i)
    this.applyOptions(i?.options as Partial<Options>)
  }

  applyOptions (o?: Partial<Options>): void {
    if (!o) return

    if (o.depth !== undefined) this.options.depth = clamp(o.depth, limits.depth, defaults.depth)
    if (o.maxNodes !== undefined) this.options.maxNodes = clamp(o.maxNodes, limits.maxNodes, defaults.maxNodes)
    if (o.perModuleLimit !== undefined) this.options.perModuleLimit = clamp(o.perModuleLimit, limits.perModuleLimit, defaults.perModuleLimit)

    if (Array.isArray(o.excludeModules)) {
      this.options.excludeModules = o.excludeModules.map(id => CortezaID(id)).filter(id => id !== NoID)
    }

    if (Array.isArray(o.labels)) {
      // one entry per module, the last one given wins; an entry with neither a
      // template nor a status field says nothing and is dropped
      const byModule = new Map<string, { moduleID: string; template: string; statusField: string }>()
      for (const l of o.labels) {
        const moduleID = CortezaID(l && l.moduleID)
        const template = String((l && l.template) || '').trim().slice(0, maxTemplate)
        const statusField = String((l && l.statusField) || '').trim().slice(0, 64)
        if (moduleID !== NoID && (template || statusField)) byModule.set(moduleID, { moduleID, template, statusField })
      }
      this.options.labels = [...byModule.values()]
    }

    if (o.layout !== undefined) {
      this.options.layout = (layouts as readonly string[]).includes(o.layout) ? o.layout : defaults.layout
    }
    if (o.clusterFrom !== undefined) this.options.clusterFrom = clamp(o.clusterFrom, limits.clusterFrom, defaults.clusterFrom)

    if (o.displayOption !== undefined) {
      this.options.displayOption = (displayOptions as readonly string[]).includes(o.displayOption) ? o.displayOption : defaults.displayOption
    }

    Apply(this.options, o, Boolean, 'showParents', 'showChildren', 'showLegend', 'showToolbar')
    // Boolean(undefined) would switch these off when a payload leaves them out
    for (const k of ['showParents', 'showChildren', 'showLegend', 'showToolbar'] as const) {
      if (o[k] === undefined) this.options[k] = defaults[k]
    }
  }
}

Registry.set(kind, PageBlockRecordGraph)
