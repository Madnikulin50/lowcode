import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath } from 'node:url'
import { resolve } from 'node:path'

const root = fileURLToPath(new URL('.', import.meta.url))
const srcDir = resolve(root, '../src')

export default defineConfig({
  root,
  plugins: [
    // Mirrors lib/vue/rollup.config.js's internal-alias plugin, which
    // resolves this self-reference only at library-build time. This
    // throwaway harness isn't going through that build, so stand in for it.
    {
      name: 'corteza-vue-self-alias',
      resolveId (source) {
        if (source.startsWith('~corteza-vue/')) {
          return resolve(srcDir, source.slice('~corteza-vue/'.length))
        }
        return null
      },
    },
    vue(),
  ],
  css: { preprocessorOptions: { scss: { additionalData: '' } } },
  server: { port: 5199 },
})
