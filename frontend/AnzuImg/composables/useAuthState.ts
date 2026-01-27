import { useState } from '#imports'

export const useAuthState = () => {
  const initialized = useState<boolean | null>('anzuimg.auth.initialized', () => null)
  const authenticated = useState<boolean>('anzuimg.auth.authenticated', () => false)
  const lastValidatedAt = useState<number | null>('anzuimg.auth.lastValidatedAt', () => null)
  const CACHE_TTL = 5 * 60 * 1000
  const setInitialized = (v: boolean) => {
    initialized.value = v
  }

  const setAuthenticated = (v: boolean) => {
    authenticated.value = v
    lastValidatedAt.value = Date.now()
  }

  const resetAuth = () => {
    authenticated.value = false
    lastValidatedAt.value = null
  }

  const isAuthValid = (): boolean => {
    if (!authenticated.value || !lastValidatedAt.value) return false
    return Date.now() - lastValidatedAt.value < CACHE_TTL
  }

  return {
    initialized,
    authenticated,
    lastValidatedAt,
    setInitialized,
    setAuthenticated,
    resetAuth,
    isAuthValid,
  }
}

