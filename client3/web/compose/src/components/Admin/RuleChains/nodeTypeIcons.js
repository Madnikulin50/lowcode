// Icon per rule chain node type, shown in the node header in RuleChainGraph.
// Keys match the `type` string the server's /admin/rulechain/nodes catalog
// (compose/rest/rulechain_node_types.go + rulechain_agent_nodes.go) uses.
export const NODE_TYPE_ICONS = {
  condition: 'question',
  crud: 'database',
  'crud.upsert': 'database',
  foreach: 'sync',
  detach: 'rss',
  mail: 'envelope',
  http: 'globe',
  ai: 'brain',
  script: 'code',
  gonec: 'code',
  workflow: 'sitemap',
  fork: 'code-branch',
  'document.extract': 'file-lines',
  'score.matrix': 'table',
  'score.weighted': 'scale-balanced',
  'risk.band': 'triangle-exclamation',
  // Remote agent / component nodes (cmdb, backup, generic service call)
  'service.call': 'plug',
  'cmdb/scan': 'server',
  'backup/run': 'server',
  'backup/restore': 'server',
  'backup/prune': 'server',
  'backup/due': 'server',
}

// Fallback for node types not in the map above — e.g. live-fetched types
// reported by a remote agent's /meta endpoint (cmdb/backup agent URLs).
export const DEFAULT_NODE_ICON = 'cog'

export function nodeTypeIcon (type) {
  return NODE_TYPE_ICONS[type] || DEFAULT_NODE_ICON
}
