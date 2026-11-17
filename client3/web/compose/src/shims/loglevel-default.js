// loglevel's CJS build (lib/loglevel.js) has no ESM default export.
// Vite serves it raw when a file under client3/lib/vue imports it, and
// `import log from 'loglevel'` then fails. Callers only need a logger.
const levels = { TRACE: 0, DEBUG: 1, INFO: 2, WARN: 3, ERROR: 4, SILENT: 5 }

function write (method) {
  return (...args) => {
    const out = console[method] || console.log
    out(...args)
  }
}

function makeLogger (name) {
  return {
    name,
    trace: write('debug'),
    debug: write('debug'),
    log: write('log'),
    info: write('info'),
    warn: write('warn'),
    error: write('error'),
    setLevel () {},
    getLevel () { return levels.WARN },
    setDefaultLevel () {},
    enableAll () {},
    disableAll () {},
    methodFactory () { return () => {} },
    getLogger (child) { return makeLogger(child || name) },
  }
}

const log = makeLogger('log')
log.levels = levels

export default log
export const getLogger = (name) => log.getLogger(name)
export { levels }
