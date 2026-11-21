const handlers = new Map()

export default {
  $emit(event, ...args) {
    const h = handlers.get(event)
    if (h) h.forEach(fn => fn(...args))
  },
  $on(event, fn) {
    if (!handlers.has(event)) handlers.set(event, new Set())
    handlers.get(event).add(fn)
  },
  $off(event, fn) {
    const h = handlers.get(event)
    if (h) h.delete(fn)
  },
}
