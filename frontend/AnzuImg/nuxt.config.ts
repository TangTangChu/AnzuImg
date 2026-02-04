// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";

const backendUrl = process.env.BACKEND_URL || 'http://localhost:9211';
const apiPrefixRaw = process.env.API_PREFIX ?? '/korori';
const appBaseUrlRaw = process.env.APP_BASE_URL || '/';

const normalizeBaseUrl = (base: string): string => {
  const trimmed = (base || '').trim();
  if (trimmed === '' || trimmed === '/') return '/';
  const withLeading = trimmed.startsWith('/') ? trimmed : `/${trimmed}`;
  const withTrailing = withLeading.endsWith('/') ? withLeading : `${withLeading}/`;
  return withTrailing.replace(/\/\/+/, '/');
};

const normalizePrefix = (prefix: string): string => {
  const trimmed = (prefix || '').trim();
  if (trimmed === '' || trimmed === '/') return '';
  const withLeading = trimmed.startsWith('/') ? trimmed : `/${trimmed}`;
  return withLeading.replace(/\/+$/, '');
};

const apiPrefix = normalizePrefix(apiPrefixRaw);
const appBaseURL = normalizeBaseUrl(appBaseUrlRaw);

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
    [`${apiPrefix}/api/**`]: { proxy: `${backendUrl}/api/**` },
    [`${apiPrefix}/health`]: { proxy: `${backendUrl}/health` },
    '/health': { proxy: `${backendUrl}/health` },
    '/i/**': { proxy: `${backendUrl}/i/**` },
  },
  runtimeConfig: {
    public: {
      apiPrefix,
      apiUseAbsoluteUrl: true,
    },
  },
  app: {
    baseURL: appBaseURL,
    head: {
      title: 'AnzuIMG',
      link: [
        { rel: 'icon', href: `${appBaseURL}favicon.svg` },
      ],
      htmlAttrs: {
        lang: 'zh-CN'
      },
    }
  },
})
