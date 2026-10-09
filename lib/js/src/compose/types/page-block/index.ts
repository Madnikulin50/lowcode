import { PageBlock, Registry } from './base'
import { PageBlockRuleChain } from './rule-chain'
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
export { PageBlockRuleChain }

export function PageBlockMaker<T extends PageBlock> (i: { kind: string }): T {
  // Computed lookup keeps RuleChain in the bundle; Registry.set() is otherwise
  // dropped when nothing instantiates compose.PageBlockRuleChain by name.
  const extra = { RuleChain: PageBlockRuleChain }
  const PageBlockTemp = Registry.get(i.kind) ?? extra[i.kind as keyof typeof extra]
  if (PageBlockTemp === undefined) {
    throw new Error(`unknown block kind '${i.kind}'`)
  }

  if (i instanceof PageBlock) {
    // Get rid of the references
    i = JSON.parse(JSON.stringify(i))
  }

  return new PageBlockTemp(i) as T
}

export {
  Registry as PageBlockRegistry,
  PageBlock,
}
