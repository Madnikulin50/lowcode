import { expect } from 'chai'
import { PageBlockCalendar } from './page-block'

describe('calendar', () => {
  describe('check namespace casting', () => {
    it('simple assignment', () => {
      const cal = new PageBlockCalendar({
        title: 'My Calendar',
        options: {
          defaultView: 'month',
          feeds: [
            { endField: 'EndDateTime', moduleID: '69055788747849745', startField: 'ActivityDate', titleField: 'Subject' },
            { endField: null, moduleID: '69055789049839633', startField: 'ActivityDate', titleField: 'Subject' },
          ],
          header: { views: ['timeGridWeek', 'dayGridMonth', 'timeGridDay', 'dayGridMonth', 'month'] },
        },
        style: { variants: { bodyBg: 'white', border: 'dark', headerBg: 'white', headerText: 'dark' } },
        kind: 'Calendar',
        xywh: [0, 0, 6, 14],
      },
      )

      expect(cal.getHeader()).to.have.property('center').equal('title')
      expect(cal.getHeader()).to.have.property('right').equal('dayGridMonth,timeGridWeek,timeGridDay')
    })
  })

  describe('record feed field names', () => {
    it('treats missing startField as invalid', () => {
      const feed = PageBlockCalendar.makeFeed({
        resource: 'compose:record',
        options: { moduleID: '69055788747849745' },
      })
      expect(feed.startField).to.equal('')
      expect(feed.isValid()).to.equal(false)
    })

    it('does not keep the literal "undefined" as a field name', () => {
      const feed = PageBlockCalendar.makeFeed({
        resource: 'compose:record',
        startField: 'undefined',
        endField: undefined,
        titleField: { name: 'Subject' },
        options: { moduleID: '69055788747849745' },
      })
      expect(feed.startField).to.equal('')
      expect(feed.titleField).to.equal('Subject')
      expect(feed.isValid()).to.equal(false)
    })

    it('lifts startField nested under options', () => {
      const feed = PageBlockCalendar.makeFeed({
        resource: 'compose:record',
        options: { moduleID: '69055788747849745', startField: 'ActivityDate' },
      })
      expect(feed.startField).to.equal('ActivityDate')
      expect(feed.isValid()).to.equal(true)
    })

    it('skips recordList when startField is missing', async () => {
      let called = false
      const api = {
        recordList () {
          called = true
          return Promise.resolve({ set: [] })
        },
      }
      const feed = PageBlockCalendar.makeFeed({
        resource: 'compose:record',
        options: { moduleID: '1' },
      })
      const events = await PageBlockCalendar.RecordFeed(
        api,
        {},
        { namespaceID: '1' },
        feed,
        { start: new Date(), end: new Date() },
      )
      expect(called).to.equal(false)
      expect(events).to.eql([])
    })

    it('skips recordList for a raw feed whose startField is undefined', async () => {
      let query = ''
      const api = {
        recordList (params: { query?: string }) {
          query = params.query || ''
          return Promise.resolve({ set: [] })
        },
      }
      const events = await PageBlockCalendar.RecordFeed(
        api,
        {},
        { namespaceID: '1' },
        { startField: undefined, endField: undefined, titleField: undefined, options: { color: '', prefilter: '' }, allDay: false },
        { start: new Date(), end: new Date() },
      )
      expect(query).to.equal('')
      expect(events).to.eql([])
    })

    it('queries with the real start field name', async () => {
      let query = ''
      const api = {
        recordList (params: { query?: string }) {
          query = params.query || ''
          return Promise.resolve({ set: [] })
        },
      }
      const feed = PageBlockCalendar.makeFeed({
        resource: 'compose:record',
        startField: 'ActivityDate',
        options: { moduleID: '1' },
      })
      await PageBlockCalendar.RecordFeed(
        api,
        {},
        { namespaceID: '1' },
        feed,
        { start: new Date('2026-01-01'), end: new Date('2026-02-01') },
      )
      expect(query).to.include('date(ActivityDate)')
      expect(query).to.not.include('undefined')
    })
  })
})
