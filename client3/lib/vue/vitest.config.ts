import { defineConfig } from 'vitest/config'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

// Unit-test harness for the shared Vue library. The library itself is bundled
// by rollup, where `corteza-lib/js/dist` stays external and is resolved by the
// consuming webapp's alias. Under vitest there is no such webapp, so we point
// that specifier at the sibling lib/js build (its ESM output).
const here = dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  test: {
    // The suite predates the vitest move (mocha-era `describe`/`it` globals).
    globals: true,
    environment: 'node',
    include: ['src/**/*.test.ts'],
  },
  resolve: {
    alias: [
      { find: /^corteza-lib\/js\/dist/, replacement: resolve(here, '../js/dist') },
      { find: /^corteza-lib\/vue\/dist/, replacement: resolve(here, 'dist') },
    ],
  },
})
