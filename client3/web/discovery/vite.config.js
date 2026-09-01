import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { execSync } from 'child_process'
import { readFileSync, writeFileSync, existsSync } from 'fs'
import { fileURLToPath } from 'url'
import { dirname } from 'path'
import { vueRuntimeSingletons } from '../vite.singletons.js'
import { dxfOpentypePlugin, dxfOpentypeAliases } from '../vite.dxf-opentype.js'

const __dirname = dirname(fileURLToPath(import.meta.url))
const singletons = vueRuntimeSingletons(__dirname)

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
    dxfOpentypePlugin(),
  ],
  define: {
    WEBAPP: JSON.stringify('Discovery'),
    VERSION: JSON.stringify(process.env.BUILD_VERSION || ('' + execSync('git describe --always --tags')).trim()),
    BUILD_TIME: JSON.stringify(new Date().toISOString()),
  },
  resolve: {
    alias: [
      ...singletons.aliasEntries,
      { find: '@', replacement: resolve(__dirname, 'src') },
      { find: 'corteza-webapp-discovery', replacement: resolve(__dirname, '.') },
      { find: 'corteza-lib/vue/dist', replacement: resolve(__dirname, '../../lib/vue/dist') },
      { find: 'corteza-lib/js/dist', replacement: resolve(__dirname, '../../lib/js/dist') },
      ...dxfOpentypeAliases(__dirname),
    ],
    dedupe: singletons.dedupe,
    modules: [
      resolve(__dirname, 'node_modules'),
      resolve(__dirname, '../../lib/vue/node_modules'),
      resolve(__dirname, '../../lib/js/node_modules'),
      resolve(__dirname, '../../node_modules'),
      'node_modules',
    ],
    extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json', '.vue'],
    mainFields: ['browser', 'main'],
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
  optimizeDeps: {
    include: [
      'corteza-lib/vue/dist',
      'corteza-lib/js/dist',
      'vue-simple-markdown',
    ],
    exclude: ['dxf-viewer', 'opentype.js'],
  },
})