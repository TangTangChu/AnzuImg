import { vi } from 'vitest'


export const useCookie = vi.fn(() => ({ value: null }))
export const navigateTo = vi.fn()
export const useRouter = vi.fn(() => ({ push: vi.fn() }))
export const useRequestHeaders = vi.fn(() => ({}))
export const useRuntimeConfig = vi.fn(() => ({ public: { apiPrefix: '/korori' } }))

export const $fetch = vi.fn()

