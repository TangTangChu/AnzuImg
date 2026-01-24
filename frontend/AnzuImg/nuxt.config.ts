// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";

const backendUrl = process.env.BACKEND_URL || 'http://localhost:8080';

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  modules: ['@nuxt/image', '@nuxtjs/i18n'],
  css: ['~/assets/css/main.css'],
  vite: {
    plugins: [
      tailwindcss(),
    ],
  },
  i18n: {
    strategy: 'no_prefix',
    locales: [
      { code: 'zh', iso: 'zh-CN', name: '简体中文', file: 'zh.json' },
      { code: 'en', iso: 'en-US', name: 'English', file: 'en.json' },
      { code: 'ja', iso: 'ja-JP', name: '日本語', file: 'ja.json' },
      { code: 'ko', iso: 'ko-KR', name: '한국어', file: 'ko.json' },
    ],
    defaultLocale: 'zh',
  },
  routeRules: {
    '/': { redirect: '/gallery' },
    '/api/**': { proxy: `${backendUrl}/api/**` },
    '/health': { proxy: `${backendUrl}/health` },
    '/i/**': { proxy: `${backendUrl}/i/**` },
  },
  app: {
    head: {
      title: 'AnzuIMG',
      link: [
        { rel: 'icon', href: '/favicon.svg' },
      ],
      htmlAttrs: {
        lang: 'zh-CN'
      },
    }
  },
})
