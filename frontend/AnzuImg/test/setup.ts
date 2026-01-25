import { vi } from 'vitest'
import { $fetch } from '#imports'

    ; (globalThis as any).$fetch = $fetch

vi.mock('@simplewebauthn/browser', () => ({
    startAuthentication: vi.fn(),
    startRegistration: vi.fn(),
}))

