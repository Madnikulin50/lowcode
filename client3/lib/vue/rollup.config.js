import vue from 'rollup-plugin-vue'
import resolve from '@rollup/plugin-node-resolve'
import commonjs from '@rollup/plugin-commonjs'
import typescript from 'rollup-plugin-typescript2'
import babel from '@rollup/plugin-babel'
import json from '@rollup/plugin-json'
import styles from 'rollup-plugin-styles'

import { readFileSync } from 'fs'
import { createRequire } from 'module'
import { resolve as resolvePath } from 'path'
import { fileURLToPath } from 'url'

const require = createRequire(import.meta.url)
const pkg = JSON.parse(readFileSync(new URL('./package.json', import.meta.url)))
const ts = require('typescript')

export default {
  input: 'src/index.ts',
  output: [
    {
      file: 'dist/index.cjs',
      format: 'cjs',
      sourcemap: true,
      inlineDynamicImports: true,
    },
    {
      file: 'dist/index.js',
      format: 'es',
      sourcemap: true,
      inlineDynamicImports: true,
    },
  ],

  external: [
    /^corteza-lib\/js\/dist/,
    /^corteza-lib\/vue\/dist/,
    /^@fortawesome\//,
    /^@tiptap\//,
    /^echarts/,
    /^brace\//,
    /^ace-builds/,
    'bootstrap-vue-next',
    'vue3-ace-editor',
    'vue', 'vue-router', 'vue-i18n', 'pinia', 'vue-demi',
    'vue3-dropzone', 'vue3-grid-layout', 'vue-grid-layout', 'vue3-resize', 'portal-vue', 'vue-color',
    'vuedraggable', 'sortablejs',
    'vue-select',
    'mxgraph', 'v-jsoneditor',
    'axios', 'i18next', 'i18next-http-backend', 'i18next-browser-languagedetector', 'i18next-multiload-backend-adapter', 'i18next-pseudo',
    'lodash', 'moment', 'numeral', 'qs',
    'bootstrap',
    'mime', 'pino',
    'docx-preview', 'xlsx', 'pptx-preview', 'jszip',
    'dxf-viewer', 'three',
    'web-ifc', '@thatopen/components', '@thatopen/fragments', 'camera-controls',
    'vue-echarts',
    'leaflet', '@vue-leaflet/vue-leaflet',
    'fs', 'path',
  ],

  plugins: [
    {
      name: 'internal-alias',
      resolveId (source) {
        if (source.startsWith('~corteza-vue/')) {
          const rel = source.slice('~corteza-vue/'.length)
          return resolvePath(fileURLToPath(new URL('.', import.meta.url)), 'src', rel)
        }
        return null
      },
    },
    resolve({
      browser: true,
      preferBuiltins: false,
    }),
    commonjs({
      include: /node_modules/,
      defaultIsModuleExports: false,
    }),
    typescript({
      typescript: ts,
      tsconfig: './tsconfig.json',
      sourceMap: true,
      check: false,
    }),
    vue({
      preprocessStyles: true,
    }),
    babel({
      exclude: /node_modules\/(?!pdfjs-dist).*/,
      babelHelpers: 'bundled',
      presets: [
        ['@babel/preset-env'],
      ],
      plugins: [
        '@babel/plugin-transform-class-properties',
        '@babel/plugin-transform-private-methods',
        '@babel/plugin-transform-private-property-in-object',
      ],
    }),
    json(),
    styles({
      exclude: ['node_modules/vue2-dropzone/**'],
      url: false,
    }),
  ],

  watch: {
    exclude: ['node_modules/**'],
  },
}
