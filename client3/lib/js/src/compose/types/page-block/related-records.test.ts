import { expect } from 'chai'
import { PageBlockMaker } from './index'
import { PageBlockRelatedRecords } from './related-records'

describe('related records block', () => {
  it('has sensible defaults', () => {
    const b = new PageBlockRelatedRecords()
    expect(b.kind).to.equal('RelatedRecords')
    expect(b.options.relations).to.deep.equal([])
    expect(b.options.showCounts).to.equal(true)
    expect(b.options.hideEmptySections).to.equal(false)
  })

  it('is made by the block maker and keeps its relations', () => {
    const b = PageBlockMaker<PageBlockRelatedRecords>({
      kind: 'RelatedRecords',
      options: {
        hideEmptySections: true,
        relations: [{ moduleID: '123', refField: 'object', title: 'Позиции', perPage: 25, fields: [{ name: 'name' }] }],
      },
    })
    expect(b).to.be.instanceOf(PageBlockRelatedRecords)
    expect(b.options.hideEmptySections).to.equal(true)
    expect(b.options.relations).to.have.lengthOf(1)
    expect(b.options.relations[0]).to.include({ moduleID: '123', refField: 'object', title: 'Позиции', perPage: 25 })
  })

  it('fills in what a relation leaves out', () => {
    const b = new PageBlockRelatedRecords({ options: { relations: [{ moduleID: '5', refField: 'parent' }] } } as never)
    expect(b.options.relations[0]).to.deep.include({
      perPage: 10, hideAddButton: false, expanded: false, title: '', presort: 'createdAt DESC',
    })
    expect(b.options.relations[0].fields).to.deep.equal([])
  })

  it('does not share state between blocks', () => {
    const a = new PageBlockRelatedRecords()
    const c = new PageBlockRelatedRecords()
    a.options.relations.push({ moduleID: '1' } as never)
    expect(c.options.relations).to.have.lengthOf(0)
  })

  it('repairs a bad page size', () => {
    const b = new PageBlockRelatedRecords({ options: { relations: [{ moduleID: '5', refField: 'x', perPage: 0 }] } } as never)
    expect(b.options.relations[0].perPage).to.equal(10)
  })
})
