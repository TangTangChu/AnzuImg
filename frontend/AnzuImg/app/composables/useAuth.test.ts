import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useAuth } from './useAuth'

describe('useAuth', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    const useCookie = vi.mocked(require('#imports').useCookie)
    useCookie.mockReturnValue({ value: null })
  })

  describe('login', () => {
    it('should login successfully with valid password', async () => {
      const mockToken = 'test-token-123'
      const $fetch = vi.mocked((global as any).$fetch)
      $fetch.mockResolvedValueOnce({ token: mockToken })

      const { login, token } = useAuth()
      
      const result = await login('valid-password')
      
      expect(result).toBe(true)
      expect($fetch).toHaveBeenCalledWith('/api/v1/auth/login', {
        method: 'POST',
        body: { password: 'valid-password' }
      })
      expect(token.value).toBe(mockToken)
    })

    it('should return false when login fails', async () => {
      const $fetch = vi.mocked((global as any).$fetch)
      $fetch.mockRejectedValueOnce(new Error('Login failed'))

      const { login, token } = useAuth()
      
      const result = await login('invalid-password')
      
      expect(result).toBe(false)
      expect(token.value).toBeNull()
    })
  })

  describe('checkInit', () => {
    it('should return true when system is initialized', async () => {
      const $fetch = vi.mocked((global as any).$fetch)
      $fetch.mockResolvedValueOnce({ initialized: true })

      const { checkInit } = useAuth()
      
      const result = await checkInit()
      
      expect(result).toBe(true)
      expect($fetch).toHaveBeenCalledWith('/api/v1/auth/status')
    })

    it('should return false when system is not initialized', async () => {
      const $fetch = vi.mocked((global as any).$fetch)
      $fetch.mockResolvedValueOnce({ initialized: false })

      const { checkInit } = useAuth()
      
      const result = await checkInit()
      
      expect(result).toBe(false)
    })

    it('should return false when backend is unavailable', async () => {
      const $fetch = vi.mocked((global as any).$fetch)
      $fetch.mockRejectedValueOnce(new Error('Network error'))

      const { checkInit } = useAuth()
      
      const result = await checkInit()
      
      expect(result).toBe(false)
    })
  })

  describe('setup', () => {
    it('should setup system successfully', async () => {
      const $fetch = vi.mocked((global as any).$fetch)
      $fetch.mockResolvedValueOnce({})

      const { setup } = useAuth()
      
      const result = await setup('new-password-123')
      
      expect(result).toBe(true)
      expect($fetch).toHaveBeenCalledWith('/api/v1/auth/setup', {
        method: 'POST',
        body: { password: 'new-password-123' }
      })
    })

    it('should return false when setup fails', async () => {
      const $fetch = vi.mocked((global as any).$fetch)
      $fetch.mockRejectedValueOnce(new Error('Setup failed'))

      const { setup } = useAuth()
      
      const result = await setup('new-password-123')
      
      expect(result).toBe(false)
    })
  })

  describe('logout', () => {
    it('should clear token and navigate to login', () => {
      const mockPush = vi.fn()
      const useRouter = vi.mocked(require('#imports').useRouter)
      useRouter.mockReturnValueOnce({ push: mockPush })

      const { logout, token } = useAuth()
      
      // Set a token first
      token.value = 'existing-token'
      
      logout()
      
      expect(token.value).toBeNull()
      expect(mockPush).toHaveBeenCalledWith('/login')
    })
  })

  describe('loginWithPasskey', () => {
    it('should login successfully with passkey', async () => {
      const mockToken = 'passkey-token-123'
      const mockAssertion = { challenge: 'test-challenge' }
      const mockAuthResp = { id: 'credential-id' }
      
      const $fetch = vi.mocked((global as any).$fetch)
      const startAuthentication = vi.mocked(require('@simplewebauthn/browser').startAuthentication)
      
      startAuthentication.mockResolvedValueOnce(mockAuthResp)
      $fetch
        .mockResolvedValueOnce({ 
          assertion: mockAssertion,
          session_id: 'test-session-id' 
        })
        .mockResolvedValueOnce({ token: mockToken })

      const { loginWithPasskey, token } = useAuth()
      
      const result = await loginWithPasskey()
      
      expect(result).toBe(true)
      expect($fetch).toHaveBeenCalledTimes(2)
      expect($fetch).toHaveBeenNthCalledWith(1, '/api/v1/auth/passkey/login/begin')
      expect($fetch).toHaveBeenNthCalledWith(2, '/api/v1/auth/passkey/login/finish', {
        method: 'POST',
        body: mockAuthResp,
        headers: {
          'X-Session-ID': 'test-session-id'
        }
      })
      expect(startAuthentication).toHaveBeenCalledWith(mockAssertion)
      expect(token.value).toBe(mockToken)
    })

    it('should return false when passkey login fails', async () => {
      const $fetch = vi.mocked((global as any).$fetch)
      $fetch.mockRejectedValueOnce(new Error('Passkey login failed'))

      const { loginWithPasskey, token } = useAuth()
      
      const result = await loginWithPasskey()
      
      expect(result).toBe(false)
      expect(token.value).toBeNull()
    })
  })
})
