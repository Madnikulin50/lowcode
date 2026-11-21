import { expect } from 'chai'
import { PageBlockMaker } from './index'
import { PageBlockRecordGraph } from './record-graph'

describe('record graph block', () => {
  it('has sensible defaults', () => {
    const b = new PageBlockRecordGraph()
    expect(b.kind).to.equal('RecordGraph')
    expect(b.options).to.deep.include({
      depth: 1, maxNodes: 150, perModuleLimit: 20, showParents: true, showChildren: true, showLegend: true,
    })
    expect(b.options.excludeModules).to.deep.equal([])
  })

  it('is made by the block maker', () => {
    const b = PageBlockMaker<PageBlockRecordGraph>({
      kind: 'RecordGraph',
      options: { depth: 2, maxNodes: 80, excludeModules: ['11', '22', '0'], showChildren: false },
    })
    expect(b).to.be.instanceOf(PageBlockRecordGraph)
    expect(b.options.depth).to.equal(2)
    expect(b.options.maxNodes).to.equal(80)
    expect(b.options.excludeModules).to.deep.equal(['11', '22'])
    expect(b.options.showChildren).to.equal(false)
    expect(b.options.showParents).to.equal(true)
  })

  it('keeps the limits within bounds', () => {
    const b = new PageBlockRecordGraph({ options: { depth: 9, maxNodes: 1, perModuleLimit: 5000 } } as never)
    expect(b.options.depth).to.equal(2)
    expect(b.options.maxNodes).to.equal(10)
    expect(b.options.perModuleLimit).to.equal(100)

    const c = new PageBlockRecordGraph({ options: { depth: 'x', maxNodes: null } } as never)
    expect(c.options.depth).to.equal(1)
    expect(c.options.maxNodes).to.equal(10)
  })

  it('knows what opening a record does, and refuses anything else', () => {
    expect(new PageBlockRecordGraph().options.displayOption).to.equal('sameTab')
    for (const o of ['sameTab', 'newTab', 'modal', 'doNothing']) {
      expect(new PageBlockRecordGraph({ options: { displayOption: o } } as never).options.displayOption).to.equal(o)
    }
    expect(new PageBlockRecordGraph({ options: { displayOption: 'popup' } } as never).options.displayOption).to.equal('sameTab')
  })

  it('keeps one template per module and drops the blank ones', () => {
    const b = new PageBlockRecordGraph({
      options: {
        labels: [
          { moduleID: '5', template: ' {{a}} · {{b}} ' },
          { moduleID: '6', template: '   ' },
          { moduleID: '0', template: '{{x}}' },
          { moduleID: '5', template: '{{c}}' },
          { moduleID: '7', template: 'x'.repeat(500) },
        ],
      },
    } as never)
    expect(b.options.labels).to.have.lengthOf(2)
    expect(b.options.labels[0]).to.deep.equal({ moduleID: '5', template: '{{c}}', statusField: '' })
    expect(b.options.labels[1].template).to.have.lengthOf(200)
    expect(new PageBlockRecordGraph().options.labels).to.deep.equal([])
  })

  it('keeps the status field of a module, with or without a template', () => {
    const b = new PageBlockRecordGraph({
      options: { labels: [{ moduleID: '5', statusField: 'severity' }, { moduleID: '6', template: '{{a}}', statusField: ' status ' }, { moduleID: '7' }] },
    } as never)
    expect(b.options.labels).to.deep.equal([
      { moduleID: '5', template: '', statusField: 'severity' },
      { moduleID: '6', template: '{{a}}', statusField: 'status' },
    ])
  })

  it('knows its layout, cluster threshold and toolbar', () => {
    const d = new PageBlockRecordGraph().options
    expect(d.layout).to.equal('force')
    expect(d.clusterFrom).to.equal(8)
    expect(d.showToolbar).to.equal(true)

    for (const l of ['force', 'radial', 'layers']) {
      expect(new PageBlockRecordGraph({ options: { layout: l } } as never).options.layout).to.equal(l)
    }
    expect(new PageBlockRecordGraph({ options: { layout: 'spiral' } } as never).options.layout).to.equal('force')
    expect(new PageBlockRecordGraph({ options: { clusterFrom: 0 } } as never).options.clusterFrom).to.equal(0)
    expect(new PageBlockRecordGraph({ options: { clusterFrom: 999 } } as never).options.clusterFrom).to.equal(50)
    expect(new PageBlockRecordGraph({ options: { showToolbar: false } } as never).options.showToolbar).to.equal(false)
  })

  it('does not share state between blocks', () => {
    const a = new PageBlockRecordGraph()
    const c = new PageBlockRecordGraph()
    a.options.excludeModules.push('1')
    expect(c.options.excludeModules).to.have.lengthOf(0)
  })
})
