import { expect } from 'chai'
import { ModuleField } from './base'

describe('check module field casting', () => {
  it('simple assignment', () => {
    const f = new ModuleField({
      name: 'fname',
      kind: 'number',
    })

    expect(f.name).to.equal('fname')
  })

  it('applies readonly option', () => {
    const f = new ModuleField({
      name: 'status',
      kind: 'String',
      options: { readonly: true },
    })
    expect(f.isReadonly).to.equal(true)
    expect(f.options.readonly).to.equal(true)
  })
})
