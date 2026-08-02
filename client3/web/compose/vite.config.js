import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve, dirname } from 'path'
import { execSync } from 'child_process'
import { readFileSync, writeFileSync } from 'fs'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  base: process.env.VITE_BASE_PATH || './',
  plugins: [
    {
      name: 'html-base-inject',
      apply: 'build',
      transformIndexHtml (html) {
        return html.replace('<base href="/" />', `<base href="${process.env.VITE_BASE_PATH || '/'}" />`)
      },
      closeBundle () {
        if (!process.env.VITE_BASE_PATH) return
        const cfg = resolve(__dirname, 'dist/config.js')
        let content = readFileSync(cfg, 'utf-8')
        content = content.replace("window.CortezaAPI = 'http://localhost:3333'", "window.CortezaAPI = '/api'")
        content = content.replace("window.CortezaAuth = 'http://localhost:3333/auth'", "window.CortezaAuth = '/auth'")
        writeFileSync(cfg, content)
      },
    },
    vue(),
    {
      name: 'fix-vue-default-import',
      transform (code, id) {
        if (/^\s*import Vue\s+from\s*['"]vue['"]/.test(code)) {
          console.log('FIXING default import from vue in:', id)
          return code.replace(/^\s*import Vue\s+from\s*['"]vue['"]/, `import * as Vue from 'vue'`)
        }
        if (/^\s*export\s*\{\s*default\s*\}\s*from\s*['"]vue['"]/.test(code)) {
          console.log('FIXING re-export default from vue in:', id)
          return code.replace(/^\s*export\s*\{\s*default\s*\}\s*from\s*['"]vue['"]/, `export * from 'vue'`)
        }
      },
    },
    {
      name: 'fix-vue-grid-layout',
      transform (code, id) {
        if (id.includes('vue-grid-layout') && id.endsWith('.common.js')) {
          const fixed = code.replace(/\)\["default"\]\s*;/, ');')
          if (fixed !== code) {
            console.log('FIXING vue-grid-layout CJS export in:', id)
            return fixed
          }
        }
      },
    },
  ],
  optimizeDeps: {
    include: [
      'portal-vue', 'echarts/core', 'echarts/renderers', 'echarts/charts', 'echarts/components', 'vue-echarts',
      '@fullcalendar/vue3', '@fullcalendar/core', '@fullcalendar/daygrid', '@fullcalendar/timegrid', '@fullcalendar/list', '@fullcalendar/bootstrap',
      'markdown-it', 'html2pdf.js', 'docx', 'vue-tweet-embed',
    ],
    exclude: ['vue-demi', 'vue-grid-layout'],
  },
  resolve: {
    alias: [
      { find: '@', replacement: resolve(__dirname, 'src') },
      { find: 'corteza-webapp-compose/src/stores', replacement: resolve(__dirname, 'src/store') },
      { find: 'corteza-webapp-compose', replacement: resolve(__dirname, '.') },

      { find: 'corteza-lib/vue/dist', replacement: resolve(__dirname, '../../lib/vue/dist') },
      { find: 'corteza-lib/js/dist', replacement: resolve(__dirname, '../../lib/js/dist') },
    ],
    modules: [
      resolve(__dirname, 'node_modules'),
      resolve(__dirname, '../../lib/vue/node_modules'),
      resolve(__dirname, '../../node_modules'),
      'node_modules',
    ],
    extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json', '.vue'],
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: '',
      },
    },
  },
  server: {
    port: 8080,
    proxy: {
      '/api': {
        target: process.env.VITE_API_URL || 'http://localhost:3333',
        changeOrigin: true,
      },
      '/auth': {
        target: process.env.VITE_API_URL || 'http://localhost:3333',
        changeOrigin: true,
      },
      '/system': {
        target: process.env.VITE_API_URL || 'http://localhost:3333',
        changeOrigin: true,
      },
      '/automation': {
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
  define: {
    WEBAPP: JSON.stringify('Compose'),
    VERSION: JSON.stringify(process.env.BUILD_VERSION || ('' + execSync('git describe --always --tags')).trim()),
    BUILD_TIME: JSON.stringify(new Date().toISOString()),
    __WEBAPP__: JSON.stringify('compose'),
    __FLAVOUR__: JSON.stringify('Namespaces'),
    __APP_LABEL__: JSON.stringify('Lowcooode Compose'),
    'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development'),
  },
})
