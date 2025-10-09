/// <reference types="vitest/config" />
import path from 'path'
import { defineConfig } from 'vite'

export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: './resources/js/vitest.setup.ts',
    include: ['resources/js/**/*.test.{ts,tsx}', 'resources/js/**/__tests__/*.{ts,tsx}'],
    coverage: {
      provider: 'v8', 
      reporter: ['text'],
      exclude: [
        'public/**',
        '**/*.d.ts',
        '**/**config.*',
        '**/app.tsx',
        '**/ssr.tsx',
        '**/wrapper.tsx',
        'resources/js/actions/**',
        'resources/js/layouts/**',
        'resources/js/routes/**',
        'resources/js/types/**',
        'resources/js/components/ui/**',
        'vendor/**',
      ],
    },
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