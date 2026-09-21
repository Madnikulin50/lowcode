import { resolve } from 'path'

/**
 * corteza-lib/vue/dist is aliased into client3/lib/vue, so Node resolution
 * finds that tree's own node_modules/{pinia,vue,fortawesome,...}. Rollup then
 * emits two copies — two Pinia runtimes, two Font Awesome `library`s — and:
 *   TypeError: Cannot read properties of undefined (reading '_s')
 *   "Could not find one or more icon(s)" for icons added in lib/vue (sun/moon)
 *
 * Pin every webapp to the copies in its own node_modules.
 */
export function vueRuntimeSingletons (appRoot) {
  const aliases = {
    pinia: resolve(appRoot, 'node_modules/pinia'),
    vue: resolve(appRoot, 'node_modules/vue'),
    'vue-i18n': resolve(appRoot, 'node_modules/vue-i18n'),
    'vue-router': resolve(appRoot, 'node_modules/vue-router'),
    '@fortawesome/fontawesome-svg-core': resolve(appRoot, 'node_modules/@fortawesome/fontawesome-svg-core'),
    '@fortawesome/free-solid-svg-icons': resolve(appRoot, 'node_modules/@fortawesome/free-solid-svg-icons'),
    '@fortawesome/free-regular-svg-icons': resolve(appRoot, 'node_modules/@fortawesome/free-regular-svg-icons'),
    '@fortawesome/vue-fontawesome': resolve(appRoot, 'node_modules/@fortawesome/vue-fontawesome'),
  }
  return {
    aliases,
    aliasEntries: Object.entries(aliases).map(([find, replacement]) => ({ find, replacement })),
    dedupe: Object.keys(aliases),
  }
}

function isIgnoredAnnotation (log) {
  if (log?.code !== 'INVALID_ANNOTATION') return false
  const hay = [log.id, log.loc?.file, log.message].filter(Boolean).join('\n')
  return hay.includes('bootstrap-vue-next') || hay.includes('@__NO_SIDE_EFFECTS__')
}

/**
 * Vite owns Rollup's logger, so a plugin `onwarn` is often ignored. Put this
 * on the user `build` config (see withQuietRollup) and also rewrite the
 * vendor file so Rollup never sees the misplaced annotation.
 *
 * bootstrap-vue-next (and its bundled vueuse helpers) put
 * `@__NO_SIDE_EFFECTS__` inside JSDoc. Rollup only accepts that tag
 * immediately before a function.
 */
export function withQuietRollup (build = {}) {
  const prevOnLog = build.rollupOptions?.onLog
  const prevOnWarn = build.rollupOptions?.onwarn
  return {
    ...build,
    rollupOptions: {
      ...build.rollupOptions,
      onLog (level, log, handler) {
        if (isIgnoredAnnotation(log)) return
        if (prevOnLog) return prevOnLog(level, log, handler)
        handler(level, log)
      },
      onwarn (warning, defaultHandler) {
        if (isIgnoredAnnotation(warning)) return
        if (prevOnWarn) return prevOnWarn(warning, defaultHandler)
        defaultHandler(warning)
      },
    },
  }
}

export function quietVendorAnnotations () {
  return {
    name: 'quiet-vendor-annotations',
    enforce: 'pre',
    transform (code, id) {
      if (!id.includes('bootstrap-vue-next')) return
      if (!code.includes('@__NO_SIDE_EFFECTS__')) return
      // Drop the tag from JSDoc (`* @__NO_SIDE_EFFECTS__`). Leave real
      // `/* @__NO_SIDE_EFFECTS__ */` annotations in place.
      const fixed = code.replace(/^[ \t]*\*[ \t]*@__NO_SIDE_EFFECTS__[ \t]*$/gm, ' *')
      if (fixed !== code) return { code: fixed, map: null }
    },
  }
}
