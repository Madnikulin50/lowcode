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
    {
      name: 'fix-dxf-viewer-opentype',
      enforce: 'pre',
      transform (code, id) {
        if (!id.includes('dxf-viewer')) return
        if (!code.includes('from "opentype.js"') && !code.includes("from 'opentype.js'")) return
        const fixed = code
          .replace(/import opentype from ["']opentype\.js["']/, 'import { parse as opentypeParse } from "opentype.js"; const opentype = { parse: opentypeParse }')
        if (fixed !== code) {
          return fixed
        }
      },
    },
  ],
  define: {
    WEBAPP: JSON.stringify('Admin'),
    VERSION: JSON.stringify(process.env.BUILD_VERSION || ('' + execSync('git describe --always --tags')).trim()),
    BUILD_TIME: JSON.stringify(new Date().toISOString()),
  },
  resolve: {
    alias: [
      { find: '@', replacement: resolve(__dirname, 'src') },
      { find: 'corteza-webapp-admin', replacement: resolve(__dirname, '.') },
      { find: 'corteza-lib/vue/dist', replacement: resolve(__dirname, '../../lib/vue/dist') },
      { find: 'corteza-lib/js/dist', replacement: resolve(__dirname, '../../lib/js/dist') },
      // dxf-viewer does `import opentype from "opentype.js"`, but opentype.mjs
      // only has named exports. Match the bare specifier only so the shim can
      // import the real file via opentype.js/dist/opentype.mjs.
      { find: /^opentype\.js$/, replacement: resolve(__dirname, 'src/shims/opentype-default.js') },
      { find: 'opentype.js/dist/opentype.mjs', replacement: resolve(__dirname, '../../lib/vue/node_modules/opentype.js/dist/opentype.mjs') },
    ],
    modules: [
      resolve(__dirname, 'node_modules'),
      resolve(__dirname, '../../lib/vue/node_modules'),
      resolve(__dirname, '../../lib/js/node_modules'),
      resolve(__dirname, '../../node_modules'),
      'node_modules',
    ],
    extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json', '.vue'],
  },
  optimizeDeps: {
    exclude: ['dxf-viewer', 'opentype.js'],
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
