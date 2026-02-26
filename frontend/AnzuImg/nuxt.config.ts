// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";

const backendUrl = process.env.ANZUIMG_FRONTEND_BACKEND_URL || 'http://backend:8080';
const apiPrefixRaw = process.env.ANZUIMG_FRONTEND_API_PREFIX ?? '';
const appBaseUrlRaw = process.env.ANZUIMG_FRONTEND_APP_BASE_URL || '/';

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
const withApiPrefix = (path: string): string => {
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  return apiPrefix ? `${apiPrefix}${normalizedPath}` : normalizedPath;
};

const backendApiBase = backendUrl.replace(/\/+$/, '');

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
    [`${withApiPrefix('/api/**')}`]: { proxy: `${backendApiBase}/api/**` },
    [`${withApiPrefix('/health')}`]: { proxy: `${backendApiBase}/health` },
    '/health': { proxy: `${backendApiBase}/health` },
    '/i/**': { proxy: `${backendApiBase}/i/**` },
  },
  nitro: {
    devProxy: {
      [`${withApiPrefix('/api')}`]: {
        target: `${backendApiBase}/api`,
        changeOrigin: true,
      },
      [`${withApiPrefix('/health')}`]: {
        target: `${backendApiBase}/health`,
        changeOrigin: true,
      },
      '/health': {
        target: `${backendApiBase}/health`,
        changeOrigin: true,
      },
      '/i': {
        target: `${backendApiBase}/i`,
        changeOrigin: true,
      },
    },
  },
  runtimeConfig: {
    public: {
      apiPrefix,
      apiUseAbsoluteUrl: false,
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
