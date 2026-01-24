export interface PasswordValidationResult {
  valid: boolean
  error?: string
}

export const validatePassword = (password: string, t: (key: string) => string): PasswordValidationResult => {
  if (password.length < 8) {
    return { valid: false, error: t('common.validation.passwordLength') }
  }
  
  if (!/[A-Z]/.test(password)) {
    return { valid: false, error: t('common.validation.passwordUppercase') }
  }
  
  if (!/[a-z]/.test(password)) {
    return { valid: false, error: t('common.validation.passwordLowercase') }
  }
  
  if (!/[0-9]/.test(password)) {
    return { valid: false, error: t('common.validation.passwordNumber') }
  }
  
  return { valid: true }
}
