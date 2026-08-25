import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { execSync } from 'child_process'
import { readFileSync, writeFileSync, existsSync } from 'fs'
import { fileURLToPath } from 'url'
import { dirname } from 'path'

const __dirname = dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  base: './',
  build: { assetsDir: 'webapp-assets' },
  plugins: [    {
      name: 'html-base-inject',
      apply: 'build',
      transformIndexHtml (html) {
        return html.replace('<base href="/" />', `<base href="${process.env.VITE_BASE_PATH || '/'}" />`)
      },
      closeBundle () {
        if (!process.env.VITE_BASE_PATH) return
        
        const cfg = resolve(__dirname, 'dist/config.js')
        if (!existsSync(cfg)) return
        let content = readFileSync(cfg, 'utf-8')
        content = content.replace("window.CortezaAPI = 'http://localhost:3333'", "window.CortezaAPI = '/api'")
        content = content.replace("window.CortezaAuth = 'http://localhost:3333/auth'", "window.CortezaAuth = '/auth'")
        writeFileSync(cfg, content)
      },
    },
    vue(),
  ],
  define: {
    WEBAPP: JSON.stringify('Admin'),
    VERSION: JSON.stringify(process.env.BUILD_VERSION || ('' + execSync('git describe --always --tags')).trim()),
    BUILD_TIME: JSON.stringify(new Date().toISOString()),
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      'corteza-webapp-admin': resolve(__dirname, '.'),
      'corteza-lib/vue/dist': resolve(__dirname, '../../lib/vue/dist'),
      'corteza-lib/js/dist': resolve(__dirname, '../../lib/js/dist'),
    },
    // corteza-lib/vue/dist is aliased to a file physically living under
    // client3/lib/vue, so plain Node resolution finds its own
    // node_modules/pinia there instead of this app's copy. Even when the
    // versions match, Rollup treats them as two distinct modules and bundles
    // pinia twice, giving each a separate (uninitialized) activePinia
    // instance — that's the "Cannot read properties of undefined (reading
    // '_s')" crash on mount. Force a single shared instance.
    dedupe: ['pinia', 'vue', 'vue-router', 'vue-i18n'],
    modules: [
      resolve(__dirname, 'node_modules'),
      resolve(__dirname, '../../lib/vue/node_modules'),
      resolve(__dirname, '../../lib/js/node_modules'),
      resolve(__dirname, '../../node_modules'),
      'node_modules',
    ],
    extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json', '.vue'],
  },
  server: {
    proxy: {
      '/api': {
        target: process.env.VITE_API_URL || 'http://localhost:3333',
        changeOrigin: true,
      },
      '/auth': {
        target: process.env.VITE_API_URL || 'http://localhost:3333',
        changeOrigin: true,
      },
      '/custom.css': {
        target: process.env.VITE_API_URL || 'http://localhost:3333',
        changeOrigin: true,
      },
      '/code-snippets.js': {
        target: process.env.VITE_API_URL || 'http://localhost:3333',
        changeOrigin: true,
      },
    },
  },
})
