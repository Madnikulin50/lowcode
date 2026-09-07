import { resolve } from 'path'
import { fileURLToPath } from 'url'

const shim = fileURLToPath(new URL('./shims/opentype-default.js', import.meta.url))

/**
 * dxf-viewer does `import opentype from "opentype.js"`, but opentype.mjs
 * only has named exports. Rewrite the default import and pin the specifier
 * to a shim that re-exports parse/load as default.
 */
export function dxfOpentypePlugin () {
  return {
    name: 'fix-dxf-viewer-opentype',
    enforce: 'pre',
    transform (code, id) {
      if (!id.includes('dxf-viewer')) return
      if (!code.includes('from "opentype.js"') && !code.includes("from 'opentype.js'")) return
      const fixed = code.replace(
        /import opentype from ["']opentype\.js["']/,
        'import { parse as opentypeParse } from "opentype.js"; const opentype = { parse: opentypeParse }',
      )
      if (fixed !== code) return fixed
    },
  }
}

export function dxfOpentypeAliases (appRoot) {
  return [
    { find: /^opentype\.js$/, replacement: shim },
    {
      find: 'opentype.js/dist/opentype.mjs',
      replacement: resolve(appRoot, '../../lib/vue/node_modules/opentype.js/dist/opentype.mjs'),
    },
  ]
}
