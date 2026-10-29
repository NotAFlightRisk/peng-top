import { writeFileSync } from 'node:fs'
import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// go:embed needs dist to exist in a fresh clone, so the build puts the placeholder back
const keepDist = { name: 'keep-dist', closeBundle: () => writeFileSync('dist/.gitkeep', '') }

export default defineConfig({
  base: './',
  plugins: [svelte(), keepDist],
  build: { outDir: 'dist', emptyOutDir: true, assetsInlineLimit: 0 },
  server: { proxy: { '/api': 'http://127.0.0.1:7340' } },
})
