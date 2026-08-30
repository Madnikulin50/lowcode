import { PageBlock, Registry } from './base'
export { PageBlockAiChat } from './ai-chat'
export { PageBlockAutomation } from './automation'
export { PageBlockChart } from './chart'
export { PageBlockContent } from './content'
export { PageBlockFile } from './file'
export { PageBlockIFrame } from './iframe'
export { PageBlockRecord } from './record'
export { PageBlockRecordList } from './record-list'
export { PageBlockRecordRevisions } from './record-revisions'
export { PageBlockRecordOrganizer } from './record-organizer'
export { PageBlockSocialFeed } from './social-feed'
export { PageBlockCalendar } from './calendar'
export { PageBlockMetric } from './metric'
export { PageBlockComment } from './comment'
export { PageBlockReport } from './report'
export { PageBlockProgress } from './progress'
export { PageBlockNavigation } from './navigation'
export { PageBlockTab } from './tabs'
export { PageBlockGeometry } from './geometry'
export { PageBlockRuleChain } from './rule-chain'
export { PageBlockVariables } from './variables'

function retainBlockDocs (block: PageBlock, raw?: { options?: Record<string, unknown> }): void {
  const src = raw?.options
  if (!src || typeof src !== 'object') return
  if (!block.options) (block as { options: Record<string, unknown> }).options = {}
  const opts = block.options as Record<string, unknown>
  if (typeof src.help === 'string') opts.help = src.help
  if (typeof src.hideHelpButton === 'boolean') opts.hideHelpButton = src.hideHelpButton
  if (typeof src.hideBrainButton === 'boolean') opts.hideBrainButton = src.hideBrainButton
}

export function PageBlockMaker<T extends PageBlock> (i: { kind: string; options?: Record<string, unknown> }): T {
  const PageBlockTemp = Registry.get(i.kind)
  if (PageBlockTemp === undefined) {
    throw new Error(`unknown block kind '${i.kind}'`)
  }

  if (i instanceof PageBlock) {
    // Get rid of the references
    i = JSON.parse(JSON.stringify(i))
  }

  const block = new PageBlockTemp(i) as T
  // Subclass field `options = { ...defaults }` wipes keys that Apply() does not know
  // (help, hideHelpButton). Put them back from the raw payload.
  retainBlockDocs(block, i)
  return block
}

export {
  Registry as PageBlockRegistry,
  PageBlock,
}
