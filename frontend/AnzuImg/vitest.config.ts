import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./test/setup.ts'],
    include: ['**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}'],
    exclude: ['**/node_modules/**', '**/dist/**', '**/.nuxt/**'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        '.nuxt/',
        'dist/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/test/**'
      ]
    }
  },
  resolve: {
    alias: {
      '#imports': new URL('./test/nuxt-imports.ts', import.meta.url).pathname,
      '~': new URL('./app', import.meta.url).pathname,
      '@': new URL('./app', import.meta.url).pathname
    }
  }
})
