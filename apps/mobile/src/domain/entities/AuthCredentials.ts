import { UserRole } from './AuthUser'

export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterCredentials {
  email: string
  password: string
  name: string
  role: UserRole
}

export interface ValidationResult {
  valid: boolean
  errors: string[]
}

export function validateEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

export function validatePassword(password: string): boolean {
  return password.length >= 8
}

export function validateLoginCredentials(
  credentials: LoginCredentials
): ValidationResult {
  const errors: string[] = []

  if (!credentials.email || credentials.email.trim() === '') {
    errors.push('Email is required')
  } else if (!validateEmail(credentials.email)) {
    errors.push('Email format is invalid')
  }

  if (!credentials.password || credentials.password.trim() === '') {
    errors.push('Password is required')
  }

  return { valid: errors.length === 0, errors }
}

export function validateRegisterCredentials(
  credentials: RegisterCredentials
): ValidationResult {
  const errors: string[] = []

  if (!credentials.email || credentials.email.trim() === '') {
    errors.push('Email is required')
  } else if (!validateEmail(credentials.email)) {
    errors.push('Email format is invalid')
  }

  if (!credentials.password || credentials.password.trim() === '') {
    errors.push('Password is required')
  } else if (!validatePassword(credentials.password)) {
    errors.push('Password must be at least 8 characters')
  }

  if (!credentials.name || credentials.name.trim() === '') {
    errors.push('Name is required')
  }

  if (!credentials.role) {
    errors.push('Role is required')
  }

  return { valid: errors.length === 0, errors }
}
