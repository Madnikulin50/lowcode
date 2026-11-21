import { expect } from 'chai'
import { isUnknownReportCount, isUnknownTotal, TotalUnknown } from './unknown-total'

describe('unknown-total', () => {
  it('detects the TotalUnknown sentinel', () => {
    expect(TotalUnknown).to.equal(-1)
    expect(isUnknownTotal(-1)).to.equal(true)
    expect(isUnknownTotal('-1')).to.equal(true)
    expect(isUnknownTotal(0)).to.equal(false)
    expect(isUnknownTotal(12)).to.equal(false)
    expect(isUnknownTotal(-2)).to.equal(false)
    expect(isUnknownTotal(null)).to.equal(false)
    expect(isUnknownTotal(undefined)).to.equal(false)
    expect(isUnknownTotal('')).to.equal(false)
  })

  it('treats a count-only -1 report row as unknown', () => {
    expect(isUnknownReportCount({ count: -1 })).to.equal(true)
    expect(isUnknownReportCount({ count: '-1' })).to.equal(true)
    expect(isUnknownReportCount({ count: 0 })).to.equal(false)
    expect(isUnknownReportCount({ count: 42 })).to.equal(false)
  })

  it('does not treat a real aggregate of -1 as unknown', () => {
    expect(isUnknownReportCount({ rp: -1 })).to.equal(false)
    expect(isUnknownReportCount({ rp: -1, count: -1 })).to.equal(false)
    expect(isUnknownReportCount(null)).to.equal(false)
    expect(isUnknownReportCount(-1)).to.equal(false)
  })
})
