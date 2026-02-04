import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useAuth } from './useAuth'
import { $fetch, navigateTo, useCookie } from '#imports'
import { startAuthentication } from '@simplewebauthn/browser'

describe('useAuth', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(useCookie).mockReturnValue({ value: null } as any)
  })

  describe('login', () => {
    it('should login successfully with valid password', async () => {
      const mockToken = 'test-token-123'
      const fetchMock = vi.mocked($fetch)
      fetchMock.mockResolvedValueOnce({ token: mockToken })

      const { login, token } = useAuth()

      const result = await login('valid-password')

      expect(result).toBe(true)
      expect(fetchMock).toHaveBeenCalledWith('/kotori/api/v1/auth/login', {
        method: 'POST',
        body: { password: 'valid-password' }
      })
      expect(token.value).toBeNull()
    })

    it('should return false when login fails', async () => {
      const fetchMock = vi.mocked($fetch)
      fetchMock.mockRejectedValueOnce(new Error('Login failed'))

      const { login, token } = useAuth()

      const result = await login('invalid-password')

      expect(result).toBe(false)
      expect(token.value).toBeNull()
    })
  })

  describe('checkInit', () => {
    it('should return true when system is initialized', async () => {
      const fetchMock = vi.mocked($fetch)
      fetchMock.mockResolvedValueOnce({ initialized: true })

      const { checkInit } = useAuth()

      const result = await checkInit()

      expect(result).toBe(true)
      expect(fetchMock).toHaveBeenCalledWith('/kotori/api/v1/auth/status')
    })

    it('should return false when system is not initialized', async () => {
      const fetchMock = vi.mocked($fetch)
      fetchMock.mockResolvedValueOnce({ initialized: false })

      const { checkInit } = useAuth()

      const result = await checkInit()

      expect(result).toBe(false)
    })

    it('should return false when backend is unavailable', async () => {
      const fetchMock = vi.mocked($fetch)
      fetchMock.mockRejectedValueOnce(new Error('Network error'))

      const { checkInit } = useAuth()

      const result = await checkInit()

      expect(result).toBe(false)
    })
  })

  describe('setup', () => {
    it('should setup system successfully', async () => {
      const fetchMock = vi.mocked($fetch)
      fetchMock.mockResolvedValueOnce({})

      const { setup } = useAuth()

      const result = await setup('new-password-123')

      expect(result).toBe(true)
      expect(fetchMock).toHaveBeenCalledWith('/kotori/api/v1/auth/setup', {
        method: 'POST',
        body: { password: 'new-password-123', setup_token: undefined },
      })
    })

    it('should include X-Setup-Token header when setup token is provided', async () => {
      const fetchMock = vi.mocked($fetch)
      fetchMock.mockResolvedValueOnce({})

      const { setup } = useAuth()

      const result = await setup('new-password-123', 'setup-token-abc')

      expect(result).toBe(true)
      expect(fetchMock).toHaveBeenCalledWith('/kotori/api/v1/auth/setup', {
        method: 'POST',
        body: { password: 'new-password-123', setup_token: 'setup-token-abc' },
      })
    })

    it('should return false when setup fails', async () => {
      const fetchMock = vi.mocked($fetch)
      fetchMock.mockRejectedValueOnce(new Error('Setup failed'))

      const { setup } = useAuth()

      const result = await setup('new-password-123')

      expect(result).toBe(false)
    })
  })

  describe('logout', () => {
    it('should clear token and navigate to login', () => {
      const navMock = vi.mocked(navigateTo)

      const { logout, token } = useAuth()

      token.value = 'existing-token'

      logout()

      expect(token.value).toBeNull()
      expect(navMock).toHaveBeenCalledWith('/login')
    })
  })

  describe('loginWithPasskey', () => {
    it('should login successfully with passkey', async () => {
      const mockToken = 'passkey-token-123'
      const mockAssertion = { challenge: 'test-challenge' }
      const mockAuthResp = { id: 'credential-id' }

      const fetchMock = vi.mocked($fetch)
      const startAuthMock = vi.mocked(startAuthentication)

      startAuthMock.mockResolvedValueOnce(mockAuthResp as any)
      fetchMock
        .mockResolvedValueOnce({
          assertion: { publicKey: mockAssertion },
          session_id: 'test-session-id'
        })
        .mockResolvedValueOnce({ token: mockToken })

      const { loginWithPasskey, token } = useAuth()

      const result = await loginWithPasskey()

      expect(result).toBe(true)
      expect(fetchMock).toHaveBeenCalledTimes(2)
      expect(fetchMock).toHaveBeenNthCalledWith(1, '/kotori/api/v1/auth/passkey/login/begin')
      expect(fetchMock).toHaveBeenNthCalledWith(2, '/kotori/api/v1/auth/passkey/login/finish', {
        method: 'POST',
        body: mockAuthResp,
        headers: {
          'X-Session-ID': 'test-session-id'
        }
      })
      expect(startAuthMock).toHaveBeenCalledWith({ optionsJSON: mockAssertion })
      expect(token.value).toBeNull()
    })

    it('should return false when passkey login fails', async () => {
      const fetchMock = vi.mocked($fetch)
      fetchMock.mockRejectedValueOnce(new Error('Passkey login failed'))

      const { loginWithPasskey, token } = useAuth()

      const result = await loginWithPasskey()

      expect(result).toBe(false)
      expect(token.value).toBeNull()
    })
  })
})
