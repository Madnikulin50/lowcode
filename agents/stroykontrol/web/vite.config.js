import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

// Relative base ("./") — this app can be served either at its own origin
// (agents/stroykontrol directly) or mounted under a sub-path via a reverse
// proxy (see compose/create_apigw_proxy.mjs). Root-absolute asset paths
// would break the moment this is accessed through any prefix.
export default defineConfig({
  plugins: [vue()],
  base: './',
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 8093,
    proxy: {
      // Points at the real stroykontrol-web backend (main.go's --listen
      // default), not at the Vite dev server itself.
      '/api': {
        target: process.env.AGENT_URL || 'http://localhost:8092',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
