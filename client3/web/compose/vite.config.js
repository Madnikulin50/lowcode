import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve, dirname } from 'path'
import { execSync } from 'child_process'
import { readFileSync, writeFileSync } from 'fs'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  base: process.env.VITE_BASE_PATH || './',
  build: { assetsDir: 'webapp-assets' },
  assetsInclude: ['**/*.wasm'],
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
    {
      name: 'fix-dxf-viewer-opentype',
      enforce: 'pre',
      transform (code, id) {
        if (!id.includes('dxf-viewer')) return
        if (!code.includes('from "opentype.js"') && !code.includes("from 'opentype.js'")) return
        const fixed = code
          .replace(/import opentype from ["']opentype\.js["']/, 'import { parse as opentypeParse } from "opentype.js"; const opentype = { parse: opentypeParse }')
        if (fixed !== code) {
          console.log('FIXING opentype.js default import in:', id)
          return fixed
        }
      },
    },
  ],
  optimizeDeps: {
    include: [
      'portal-vue', 'echarts/core', 'echarts/renderers', 'echarts/charts', 'echarts/components', 'vue-echarts',
      '@fullcalendar/vue3', '@fullcalendar/core', '@fullcalendar/daygrid', '@fullcalendar/timegrid', '@fullcalendar/list', '@fullcalendar/bootstrap',
      'markdown-it', 'html2pdf.js', 'docx', 'vue-tweet-embed', 'mxgraph', 'v-jsoneditor',
      'docx-preview', 'xlsx', 'pptx-preview', 'jszip', 'three',
    ],
    exclude: ['vue-demi', 'vue-grid-layout', 'web-ifc', '@thatopen/components', '@thatopen/fragments', 'dxf-viewer', 'opentype.js'],
    esbuildOptions: {
      plugins: [
        {
          name: 'opentype-default-export',
          setup (build) {
            build.onLoad({ filter: /[\\/]dxf-viewer[\\/].*DxfWorker\.js$/ }, (args) => {
              const contents = readFileSync(args.path, 'utf8').replace(
                /import opentype from ["']opentype\.js["']/,
                'import { parse as opentypeParse } from "opentype.js"; const opentype = { parse: opentypeParse }',
              )
              return { contents, loader: 'js' }
            })
          },
        },
      ],
    },
  },
  resolve: {
    alias: [
      { find: '@', replacement: resolve(__dirname, 'src') },
      { find: 'corteza-webapp-compose/src/stores', replacement: resolve(__dirname, 'src/store') },
      { find: 'corteza-webapp-compose', replacement: resolve(__dirname, '.') },

      { find: 'corteza-lib/vue/src', replacement: resolve(__dirname, '../../lib/vue/src') },
      { find: '~corteza-vue', replacement: resolve(__dirname, '../../lib/vue/src') },
      { find: 'corteza-lib/vue/dist/WorkflowEditor', replacement: resolve(__dirname, '../../lib/vue/src/components/workflow/WorkflowEditor.vue') },
      { find: 'corteza-lib/vue/dist', replacement: resolve(__dirname, '../../lib/vue/dist') },
      { find: 'corteza-lib/js/dist', replacement: resolve(__dirname, '../../lib/js/dist') },
      { find: 'mxgraph', replacement: resolve(__dirname, 'node_modules/mxgraph/javascript/dist/build.js') },
      { find: 'v-jsoneditor', replacement: resolve(__dirname, 'node_modules/v-jsoneditor/dist/v-jsoneditor.min.js') },
      { find: 'file-saver', replacement: resolve(__dirname, 'node_modules/file-saver/dist/FileSaver.min.js') },
      // dxf-viewer does `import opentype from "opentype.js"`, but opentype.mjs
      // only has named exports. Match the bare specifier only so the shim can
      // import the real file via opentype.js/dist/opentype.mjs.
      { find: /^opentype\.js$/, replacement: resolve(__dirname, 'src/shims/opentype-default.js') },
    ],
    // corteza-lib/vue/dist is aliased to a file physically living under
    // client3/lib/vue, so plain Node resolution finds its own
    // node_modules/pinia there instead of this app's copy. Even with
    // matching versions, Rollup treats them as two distinct modules and
    // bundles pinia twice, each with its own (uninitialized) activePinia
    // instance — causing "Cannot read properties of undefined (reading
    // '_s')" on mount. Force a single shared instance.
    //
    // Same story for @tiptap/* + prosemirror-*: CRichTextInput lives under
    // client3/lib/vue/src too, so it resolves its own copy of
    // prosemirror-state instead of this app's. Two Plugin instances with
    // the same key but different module identity make prosemirror's
    // Configuration throw "Adding different instances of a keyed plugin"
    // when the editor is reconfigured/remounted (e.g. Builder.vue's
    // editBlock). Force a single shared instance here too.
    dedupe: [
      'pinia', 'vue', 'vue-router', 'vue-i18n', 'three',
      '@tiptap/core', '@tiptap/vue-3', '@tiptap/suggestion', '@tiptap/starter-kit',
      '@tiptap/extension-list', '@tiptap/extension-mention', '@tiptap/extension-emoji',
      '@tiptap/extension-table', '@tiptap/extension-text-align', '@tiptap/extension-text-style',
      '@tiptap/extension-placeholder', '@tiptap/extension-image',
      'prosemirror-state', 'prosemirror-view', 'prosemirror-model',
      'prosemirror-transform', 'prosemirror-keymap',
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
      '/compose': {
        target: process.env.VITE_API_URL || 'http://localhost:3333',
        changeOrigin: true,
        bypass (req) {
          // Serve Vite's SPA entry, not /index.html (public/index.html is a
          // Vue CLI leftover and would leak <%= BASE_URL %> into the page).
          if (req.url?.startsWith('/compose/auth/callback')) return '/'
          const accept = req.headers.accept || ''
          if (accept.includes('text/html')) return '/'
        },
      },
      // Negative lookahead: /auth/callback is the SPA OAuth return URL.
      // Vite 6 bypass:false is a hard 404, not a skip.
      '^/auth(?!/callback)': {
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
      '/connector': {
        target: process.env.VITE_API_URL || 'http://localhost:3333',
        changeOrigin: true,
      },
      '/datasource': {
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
