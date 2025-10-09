/// <reference types="vitest/config" />
import path from 'path'
import { defineConfig } from 'vite'

export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: './resources/js/vitest.setup.ts',
    include: ['resources/js/**/*.test.{ts,tsx}', 'resources/js/**/__tests__/*.{ts,tsx}'],
  },
  esbuild: {
    jsx: 'automatic',
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'resources/js'),
    },
  },
})