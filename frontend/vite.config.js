import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath } from 'url' // 1. Import this built-in Node module

// 2. Re-create __dirname functionality using modern ESM syntax
const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

import path from 'path'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})