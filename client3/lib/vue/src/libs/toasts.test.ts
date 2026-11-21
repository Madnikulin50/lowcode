import { expect } from 'chai'
import { toasts, pushToast, removeToast, toastSuccess, toastDanger, toastErrorHandler } from './toasts'

// The toast store is a plain module-level reactive array shared by every
// webapp; the migration off bootstrap-vue rewrote this file, so pin its
// observable contract (ids, variant class, notification-prefix stripping).
describe('libs/toasts', () => {
  afterEach(() => {
    toasts.splice(0, toasts.length)
  })

  it('pushToast appends to the shared list', () => {
    const t = pushToast('hello', { variant: 'danger' })
    expect(toasts).to.have.length(1)
    expect(t.payload.notes).to.equal('hello')
    expect(t.options?.toastClass).to.equal('text-bg-danger')
  })

  it('pushToast defaults to the success variant', () => {
    const t = pushToast('ok')
    expect(t.options?.toastClass).to.equal('text-bg-success')
  })

  it('gives every toast a unique id', () => {
    const a = pushToast('a')
    const b = pushToast('b')
    expect(a.id).to.not.equal(b.id)
  })

  it('removeToast drops only the matching item', () => {
    const a = pushToast('a')
    const b = pushToast('b')
    removeToast(a.id)
    expect(toasts).to.have.length(1)
    expect(toasts[0].id).to.equal(b.id)
  })

  it('removeToast is a no-op for an unknown id', () => {
    pushToast('a')
    removeToast(9999)
    expect(toasts).to.have.length(1)
  })

  it('toastSuccess / toastDanger set the variant', () => {
    toastSuccess('s')
    toastDanger('d')
    expect(toasts[0].options?.toastClass).to.equal('text-bg-success')
    expect(toasts[1].options?.toastClass).to.equal('text-bg-danger')
  })

  it('toastErrorHandler strips the notification. prefix and capitalises', () => {
    const handler = toastErrorHandler({ title: 'boom' })
    handler({ message: 'notification.some.key' } as Error)
    expect(toasts[0].payload.notes).to.equal('Some.key')
    expect(toasts[0].payload.title).to.equal('Boom')
  })

  it('toastErrorHandler falls back to the title when there is no message', () => {
    const handler = toastErrorHandler('boom')
    handler({ message: '' } as Error)
    expect(toasts[0].payload.notes).to.equal('boom')
  })
})
